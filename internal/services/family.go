package services

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/auth"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/repository"
)

var (
	ErrPasswordTooShort   = errors.New("password too short")
	ErrInvalidRole        = errors.New("invalid role")
	ErrMemberNotFound     = errors.New("member not found")
	ErrInvalidMemberData  = errors.New("invalid member data")
	ErrInvalidPermissions = errors.New("invalid permissions")
)

var validPermissions = map[string]struct{}{
	"view_dashboard":   {},
	"assign_tasks":     {},
	"view_tasks":       {},
	"approve_rewards":  {},
	"view_messages":    {},
	"manage_quizzes":   {},
	"manage_learn":     {},
	"respond_requests": {},
}

type FamilyService struct {
	familyRepo *repository.FamilyRepo
	xpRepo     *repository.XPRepo
}

func NewFamilyService(familyRepo *repository.FamilyRepo, xpRepo *repository.XPRepo) *FamilyService {
	return &FamilyService{
		familyRepo: familyRepo,
		xpRepo:     xpRepo,
	}
}

func (s *FamilyService) GetFamily(ctx context.Context, familyID string) (*models.Family, error) {
	return s.familyRepo.GetFamilyByID(ctx, familyID, familyID)
}

func (s *FamilyService) UpdateFamily(ctx context.Context, familyID, name string, logoURL *string) (*models.Family, error) {
	return s.familyRepo.UpdateFamily(ctx, familyID, familyID, name, logoURL)
}

func (s *FamilyService) ListMembers(ctx context.Context, familyID string) ([]*models.User, error) {
	return s.familyRepo.ListMembers(ctx, familyID)
}

type CreateMemberInput struct {
	FamilyID         uuid.UUID
	Role             string
	Name             string
	Username         *string
	Email            *string
	Password         string
	DateOfBirth      *time.Time
	GameLimitMinutes int
	CreatedBy        uuid.UUID
}

func (s *FamilyService) CreateMember(ctx context.Context, input CreateMemberInput) (*models.User, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || len(input.Name) > 120 {
		return nil, ErrInvalidMemberData
	}

	// Validate role
	switch input.Role {
	case "parent", "child", "adult_relative":
	default:
		return nil, ErrInvalidRole
	}

	if input.Role == "child" {
		if input.Username == nil || strings.TrimSpace(*input.Username) == "" {
			return nil, ErrInvalidMemberData
		}
		u := strings.TrimSpace(*input.Username)
		input.Username = &u
	}

	if input.Role != "child" {
		if input.Email == nil || strings.TrimSpace(*input.Email) == "" {
			return nil, ErrInvalidMemberData
		}
		e := strings.TrimSpace(strings.ToLower(*input.Email))
		if _, err := mail.ParseAddress(e); err != nil {
			return nil, ErrInvalidMemberData
		}
		input.Email = &e
	}

	// Validate minimum password length
	minLen := 8
	if input.Role == "child" {
		minLen = 6
	}
	if len(input.Password) < minLen {
		return nil, ErrPasswordTooShort
	}

	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	gameLimitMinutes := input.GameLimitMinutes
	if gameLimitMinutes == 0 {
		gameLimitMinutes = 60
	}

	user := &models.User{
		FamilyID:         input.FamilyID,
		Role:             input.Role,
		Name:             input.Name,
		Username:         input.Username,
		Email:            input.Email,
		PasswordHash:     passwordHash,
		Theme:            "forest",
		DateOfBirth:      input.DateOfBirth,
		GameLimitMinutes: gameLimitMinutes,
		IsActive:         true,
		CreatedBy:        &input.CreatedBy,
	}

	if err := s.familyRepo.CreateMember(ctx, user); err != nil {
		return nil, err
	}

	// Initialize XP
	_, _ = s.xpRepo.GetOrCreateUserXP(ctx, user.ID.String(), input.FamilyID.String())

	return user, nil
}

func (s *FamilyService) GetMember(ctx context.Context, memberID, familyID string) (*models.User, error) {
	return s.familyRepo.GetMemberByID(ctx, memberID, familyID)
}

func (s *FamilyService) UpdateMember(ctx context.Context, memberID, familyID string, updates map[string]interface{}) (*models.User, error) {
	return s.familyRepo.UpdateMember(ctx, memberID, familyID, updates)
}

func (s *FamilyService) DeactivateMember(ctx context.Context, memberID, familyID string) error {
	return s.familyRepo.DeactivateMember(ctx, memberID, familyID)
}

func (s *FamilyService) GetPermissions(ctx context.Context, userID, familyID string) ([]string, error) {
	perms, err := s.familyRepo.GetPermissions(ctx, userID, familyID)
	if err != nil {
		return []string{}, nil
	}
	return perms, nil
}

func (s *FamilyService) SetPermissions(ctx context.Context, granteeID, familyID string, grantorID uuid.UUID, perms []string) (*models.FamilyAccessControl, error) {
	grantee, err := s.familyRepo.GetMemberByID(ctx, granteeID, familyID)
	if err != nil {
		return nil, ErrMemberNotFound
	}
	if grantee.Role != "adult_relative" {
		return nil, ErrInvalidPermissions
	}
	for _, p := range perms {
		if _, ok := validPermissions[p]; !ok {
			return nil, ErrInvalidPermissions
		}
	}

	return s.familyRepo.SetPermissions(ctx, granteeID, familyID, grantorID, perms)
}

func (s *FamilyService) RevokePermissions(ctx context.Context, granteeID, familyID string) error {
	return s.familyRepo.RevokePermissions(ctx, granteeID, familyID)
}

func (s *FamilyService) ListAccessControl(ctx context.Context, familyID string) ([]*models.FamilyAccessControl, error) {
	return s.familyRepo.ListAccessControl(ctx, familyID)
}

func (s *FamilyService) GetRantCount(ctx context.Context, childID, familyID string) (int, error) {
	return s.familyRepo.GetRantCount(ctx, childID, familyID)
}
