package services

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/auth"
	"github.com/rawdah/rawdah-api/internal/mailer"
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

const (
	minChildAge           = 5
	maxChildAge           = 18
	defaultAdultPlatform  = "https://app.rawdah.app"
	defaultKidsPlatform   = "https://kids.rawdah.app"
	loginEmailSendTimeout = 10 * time.Second
)

type FamilyService struct {
	familyRepo       *repository.FamilyRepo
	xpRepo           *repository.XPRepo
	mailer           *mailer.Mailer
	adultPlatformURL string
	kidsPlatformURL  string
}

func NewFamilyService(
	familyRepo *repository.FamilyRepo,
	xpRepo *repository.XPRepo,
	m *mailer.Mailer,
	adultPlatformURL string,
	kidsPlatformURL string,
) *FamilyService {
	adultPlatformURL = strings.TrimSpace(adultPlatformURL)
	if adultPlatformURL == "" {
		adultPlatformURL = defaultAdultPlatform
	}

	kidsPlatformURL = strings.TrimSpace(kidsPlatformURL)
	if kidsPlatformURL == "" {
		kidsPlatformURL = defaultKidsPlatform
	}

	return &FamilyService{
		familyRepo:       familyRepo,
		xpRepo:           xpRepo,
		mailer:           m,
		adultPlatformURL: adultPlatformURL,
		kidsPlatformURL:  kidsPlatformURL,
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
	ChildAge         *int
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

		if input.ChildAge != nil {
			dob, err := dateOfBirthFromAge(*input.ChildAge)
			if err != nil {
				return nil, ErrInvalidMemberData
			}
			input.DateOfBirth = &dob
		}

		if input.DateOfBirth == nil || !dateOfBirthInChildRange(*input.DateOfBirth) {
			return nil, ErrInvalidMemberData
		}
	}

	if input.Email != nil {
		e := strings.TrimSpace(strings.ToLower(*input.Email))
		if e == "" {
			input.Email = nil
		} else {
			if _, err := mail.ParseAddress(e); err != nil {
				return nil, ErrInvalidMemberData
			}
			input.Email = &e
		}
	}

	if input.Role != "child" && input.Email == nil {
		return nil, ErrInvalidMemberData
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

	// Notify active parent accounts when a new family member is added.
	go s.notifyParentsMemberAdded(input.FamilyID, input.CreatedBy, user)
	go s.notifyMemberLoginDetails(input.FamilyID, user, input.Password)

	return user, nil
}

func (s *FamilyService) GetMember(ctx context.Context, memberID, familyID string) (*models.User, error) {
	return s.familyRepo.GetMemberByID(ctx, memberID, familyID)
}

func (s *FamilyService) UpdateMember(ctx context.Context, memberID, familyID string, updates map[string]interface{}) (*models.User, error) {
	member, err := s.familyRepo.GetMemberByID(ctx, memberID, familyID)
	if err != nil {
		return nil, ErrMemberNotFound
	}

	if passwordRaw, ok := updates["password"]; ok && passwordRaw != nil {
		passwordPtr, ok := passwordRaw.(*string)
		if !ok || passwordPtr == nil {
			return nil, ErrInvalidMemberData
		}
		password := strings.TrimSpace(*passwordPtr)
		if password == "" {
			return nil, ErrInvalidMemberData
		}

		minLen := 8
		if member.Role == "child" {
			minLen = 6
		}
		if len(password) < minLen {
			return nil, ErrPasswordTooShort
		}

		passwordHash, err := auth.HashPassword(password)
		if err != nil {
			return nil, err
		}
		updates["password_hash"] = passwordHash
	}
	delete(updates, "password")

	if childAgeRaw, ok := updates["child_age"]; ok && childAgeRaw != nil {
		childAgePtr, ok := childAgeRaw.(*int)
		if !ok || childAgePtr == nil || member.Role != "child" {
			return nil, ErrInvalidMemberData
		}

		dob, err := dateOfBirthFromAge(*childAgePtr)
		if err != nil {
			return nil, ErrInvalidMemberData
		}
		updates["date_of_birth"] = &dob
	}
	delete(updates, "child_age")

	if dobRaw, ok := updates["date_of_birth"]; ok && dobRaw != nil && member.Role == "child" {
		dobPtr, ok := dobRaw.(*time.Time)
		if !ok || dobPtr == nil || !dateOfBirthInChildRange(*dobPtr) {
			return nil, ErrInvalidMemberData
		}
	}

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

func (s *FamilyService) notifyParentsMemberAdded(familyID, createdBy uuid.UUID, member *models.User) {
	if s.mailer == nil || member == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), loginEmailSendTimeout)
	defer cancel()

	family, err := s.familyRepo.GetFamilyByID(ctx, familyID.String(), familyID.String())
	if err != nil {
		return
	}
	members, err := s.familyRepo.ListMembers(ctx, familyID.String())
	if err != nil {
		return
	}

	creatorName := "A parent"
	if creator, err := s.familyRepo.GetMemberByID(ctx, createdBy.String(), familyID.String()); err == nil {
		if trimmed := strings.TrimSpace(creator.Name); trimmed != "" {
			creatorName = trimmed
		}
	}

	memberName := html.EscapeString(member.Name)
	roleLabel := html.EscapeString(memberRoleDisplayName(member.Role))
	addedBy := html.EscapeString(creatorName)
	subject := fmt.Sprintf("New family member added to %s", family.Name)
	body := fmt.Sprintf(
		"<p><strong>%s</strong> was added as <strong>%s</strong> in your family.</p><p>Added by: <strong>%s</strong></p>",
		memberName, roleLabel, addedBy,
	)
	emailHTML := mailer.BuildEmail(
		"Family Updated",
		body,
		"View Family",
		platformPath(s.adultPlatformURL, "/family"),
		family.Name,
	)

	for _, m := range members {
		if m == nil || !m.IsActive || m.Role != "parent" || m.Email == nil {
			continue
		}
		if m.ID == member.ID {
			continue
		}

		email := strings.TrimSpace(*m.Email)
		if email == "" {
			continue
		}
		if _, err := mail.ParseAddress(email); err != nil {
			continue
		}
		_ = s.mailer.Send(mailer.BrevoContact{Name: m.Name, Email: email}, subject, emailHTML)
	}
}

func (s *FamilyService) notifyMemberLoginDetails(familyID uuid.UUID, member *models.User, password string) {
	if s.mailer == nil || member == nil || member.Email == nil {
		return
	}

	email := strings.TrimSpace(*member.Email)
	if email == "" {
		return
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return
	}

	loginValue := email
	if member.Username != nil {
		username := strings.TrimSpace(*member.Username)
		if username != "" {
			loginValue = username
		}
	}

	platformURL := s.platformURLForRole(member.Role)
	platformLabel := "Adults platform"
	if member.Role == "child" {
		platformLabel = "Kids platform"
	}

	loginLabel := "Username / email"
	if member.Role == "child" {
		loginLabel = "Username"
	}

	familyName := "your family"
	ctx, cancel := context.WithTimeout(context.Background(), loginEmailSendTimeout)
	defer cancel()

	if family, err := s.familyRepo.GetFamilyByID(ctx, familyID.String(), familyID.String()); err == nil {
		if trimmed := strings.TrimSpace(family.Name); trimmed != "" {
			familyName = trimmed
		}
	}

	escapedFamilyName := html.EscapeString(familyName)
	escapedLoginLabel := html.EscapeString(loginLabel)
	escapedLoginValue := html.EscapeString(loginValue)
	escapedPassword := html.EscapeString(password)
	escapedPlatformLabel := html.EscapeString(platformLabel)
	escapedPlatformURL := html.EscapeString(platformURL)

	subject := "Your Rawdah login details"
	body := fmt.Sprintf(
		"<p>Your account has been created for <strong>%s</strong>.</p>"+
			"<p>Use these details to sign in:</p>"+
			"<ul>"+
			"<li><strong>Family name:</strong> %s</li>"+
			"<li><strong>%s:</strong> %s</li>"+
			"<li><strong>Password:</strong> %s</li>"+
			"<li><strong>%s:</strong> <a href=\"%s\">%s</a></li>"+
			"</ul>",
		escapedFamilyName,
		escapedFamilyName,
		escapedLoginLabel,
		escapedLoginValue,
		escapedPassword,
		escapedPlatformLabel,
		escapedPlatformURL,
		escapedPlatformURL,
	)
	emailHTML := mailer.BuildEmail("Welcome to Rawdah", body, "Open Platform", platformURL, familyName)
	_ = s.mailer.Send(mailer.BrevoContact{Name: member.Name, Email: email}, subject, emailHTML)
}

func (s *FamilyService) platformURLForRole(role string) string {
	if role == "child" {
		return s.kidsPlatformURL
	}
	return s.adultPlatformURL
}

func platformPath(baseURL, path string) string {
	return strings.TrimRight(baseURL, "/") + path
}

func memberRoleDisplayName(role string) string {
	switch role {
	case "parent":
		return "Parent"
	case "adult_relative":
		return "Adult Relative"
	case "child":
		return "Child"
	default:
		return "Family Member"
	}
}

func dateOfBirthFromAge(age int) (time.Time, error) {
	if age < minChildAge || age > maxChildAge {
		return time.Time{}, ErrInvalidMemberData
	}

	now := time.Now().UTC()
	dob := now.AddDate(-age, 0, 0)
	return time.Date(dob.Year(), dob.Month(), dob.Day(), 0, 0, 0, 0, time.UTC), nil
}

func dateOfBirthInChildRange(dob time.Time) bool {
	age := childAgeFromDateOfBirth(&dob)
	return age >= minChildAge && age <= maxChildAge
}
