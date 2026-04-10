package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/auth"
	"github.com/rawdah/rawdah-api/internal/config"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already taken")
	ErrSlugTaken          = errors.New("slug already taken")
	ErrUserInactive       = errors.New("user is inactive")
)

type AuthService struct {
	db         *sqlx.DB
	authRepo   *repository.AuthRepo
	userRepo   *repository.UserRepo
	familyRepo *repository.FamilyRepo
	cfg        *config.Config
}

func NewAuthService(db *sqlx.DB, authRepo *repository.AuthRepo, userRepo *repository.UserRepo, familyRepo *repository.FamilyRepo, cfg *config.Config) *AuthService {
	return &AuthService{
		db:         db,
		authRepo:   authRepo,
		userRepo:   userRepo,
		familyRepo: familyRepo,
		cfg:        cfg,
	}
}

type SignupInput struct {
	FamilyName string
	Slug       string
	Name       string
	Email      string
	Password   string
}

type AuthTokens struct {
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	User         *models.User   `json:"user"`
	Family       *models.Family `json:"family"`
}

func (s *AuthService) Signup(ctx context.Context, input SignupInput) (*AuthTokens, error) {
	input.Slug = strings.TrimSpace(strings.ToLower(input.Slug))
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.FamilyName = strings.TrimSpace(input.FamilyName)
	input.Name = strings.TrimSpace(input.Name)

	// Check slug uniqueness
	existingFamily, err := s.familyRepo.GetFamilyBySlug(ctx, input.Slug)
	if err == nil && existingFamily != nil {
		return nil, ErrSlugTaken
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("check family slug availability: %w", err)
	}

	// Check email uniqueness
	existingUser, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailTaken
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("check email availability: %w", err)
	}

	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	family := &models.Family{
		Name: input.FamilyName,
		Slug: input.Slug,
		Plan: "free",
	}
	if err := s.familyRepo.CreateFamily(ctx, family); err != nil {
		if isUniqueViolation(err, "families_slug_key") {
			return nil, ErrSlugTaken
		}
		return nil, fmt.Errorf("create family: %w", err)
	}

	email := input.Email
	user := &models.User{
		FamilyID:         family.ID,
		Role:             "parent",
		Name:             input.Name,
		Email:            &email,
		PasswordHash:     passwordHash,
		Theme:            "forest",
		IsActive:         true,
		GameLimitMinutes: 60,
	}
	if err := s.familyRepo.CreateMember(ctx, user); err != nil {
		if isUniqueViolation(err, "idx_users_email") {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("create initial parent user: %w", err)
	}

	// Create user XP record
	xpRepo := repository.NewXPRepo(s.db)
	_, _ = xpRepo.GetOrCreateUserXP(ctx, user.ID.String(), family.ID.String())

	tokens, err := s.issueTokens(ctx, user, family, s.cfg.AccessTokenTTL, true)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

type SigninInput struct {
	Email    string
	Password string
}

func (s *AuthService) Signin(ctx context.Context, input SigninInput) (*AuthTokens, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))

	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrUserInactive
	}
	if err := auth.CheckPassword(input.Password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	family, err := s.familyRepo.GetFamilyByID(ctx, user.FamilyID.String(), user.FamilyID.String())
	if err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user, family, s.cfg.AccessTokenTTL, true)
}

type ChildSigninInput struct {
	FamilySlug string
	Username   string
	Password   string
}

func (s *AuthService) ChildSignin(ctx context.Context, input ChildSigninInput) (*AuthTokens, error) {
	input.FamilySlug = strings.TrimSpace(strings.ToLower(input.FamilySlug))
	input.Username = strings.TrimSpace(input.Username)

	family, err := s.familyRepo.GetFamilyBySlugOrName(ctx, input.FamilySlug)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.GetByUsernameAndFamily(ctx, input.Username, family.ID.String())
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrUserInactive
	}
	if err := auth.CheckPassword(input.Password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	return s.issueTokens(ctx, user, family, s.cfg.ChildTokenTTL, false)
}

func (s *AuthService) Refresh(ctx context.Context, rawToken string) (*AuthTokens, error) {
	tokenHash := hashToken(rawToken)

	rt, err := s.authRepo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if time.Now().After(rt.ExpiresAt) {
		_ = s.authRepo.DeleteRefreshToken(ctx, tokenHash)
		return nil, ErrInvalidCredentials
	}

	// Delete old token (rotation)
	_ = s.authRepo.DeleteRefreshToken(ctx, tokenHash)

	// Lookup user without family constraint for refresh
	userByID := &models.User{}
	if err := s.db.GetContext(ctx, userByID,
		`SELECT id, family_id, role, name, username, email, password_hash, avatar_url,
		        theme, date_of_birth, game_limit_minutes, is_active, created_by, last_login_at, created_at
		 FROM users WHERE id = $1`, rt.UserID); err != nil {
		return nil, ErrInvalidCredentials
	}
	user := userByID
	if !user.IsActive {
		return nil, ErrUserInactive
	}
	if user.Role == "child" {
		return nil, ErrInvalidCredentials
	}

	family, err := s.familyRepo.GetFamilyByID(ctx, user.FamilyID.String(), user.FamilyID.String())
	if err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user, family, s.cfg.AccessTokenTTL, true)
}

func (s *AuthService) Signout(ctx context.Context, rawToken string) error {
	tokenHash := hashToken(rawToken)
	return s.authRepo.DeleteRefreshToken(ctx, tokenHash)
}

func (s *AuthService) Me(ctx context.Context, userID, familyID string) (*models.User, *models.Family, error) {
	user, err := s.userRepo.GetByID(ctx, userID, familyID)
	if err != nil {
		return nil, nil, err
	}
	family, err := s.familyRepo.GetFamilyByID(ctx, familyID, familyID)
	if err != nil {
		return nil, nil, err
	}
	return user, family, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID, familyID, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID, familyID)
	if err != nil {
		return err
	}

	minLen := 8
	if user.Role == "child" {
		minLen = 6
	}
	if len(newPassword) < minLen {
		return ErrPasswordTooShort
	}

	if err := auth.CheckPassword(currentPassword, user.PasswordHash); err != nil {
		return ErrInvalidCredentials
	}
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	return s.userRepo.UpdatePassword(ctx, uid, hash)
}

func (s *AuthService) issueTokens(ctx context.Context, user *models.User, family *models.Family, accessTTL time.Duration, includeRefresh bool) (*AuthTokens, error) {
	accessToken, err := auth.IssueAccessToken(user.ID, family.ID, user.Role, s.cfg.JWTAccessSecret, accessTTL)
	if err != nil {
		return nil, err
	}

	rawRefresh := ""
	if includeRefresh {
		rawRefresh, err = auth.GenerateRefreshToken()
		if err != nil {
			return nil, err
		}

		tokenHash := hashToken(rawRefresh)
		expiresAt := time.Now().Add(s.cfg.RefreshTokenTTL)
		if err := s.authRepo.StoreRefreshToken(ctx, user.ID, tokenHash, expiresAt); err != nil {
			return nil, err
		}
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		User:         user,
		Family:       family,
	}, nil
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func isUniqueViolation(err error, constraints ...string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	if pgErr.Code != "23505" {
		return false
	}
	if len(constraints) == 0 {
		return true
	}
	for _, c := range constraints {
		if pgErr.ConstraintName == c {
			return true
		}
	}
	return false
}
