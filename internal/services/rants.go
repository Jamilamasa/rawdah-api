package services

import (
	"context"
	"errors"

	"github.com/rawdah/rawdah-api/internal/auth"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/repository"
)

var ErrRantLocked = errors.New("rant is locked")

type RantService struct {
	rantRepo *repository.RantRepo
}

func NewRantService(rantRepo *repository.RantRepo) *RantService {
	return &RantService{rantRepo: rantRepo}
}

func (s *RantService) List(ctx context.Context, userID string) ([]*models.Rant, error) {
	return s.rantRepo.List(ctx, userID)
}

type CreateRantInput struct {
	UserID   string
	Title    *string
	Content  string
	Password *string
}

func (s *RantService) Create(ctx context.Context, input CreateRantInput) (*models.Rant, error) {
	userUUID, err := parseUUID(input.UserID)
	if err != nil {
		return nil, err
	}

	var passwordHash *string
	if input.Password != nil && *input.Password != "" {
		hash, err := auth.HashPassword(*input.Password)
		if err != nil {
			return nil, err
		}
		passwordHash = &hash
	}

	rant := &models.Rant{
		UserID:       userUUID,
		Title:        input.Title,
		Content:      input.Content,
		PasswordHash: passwordHash,
	}

	if err := s.rantRepo.Create(ctx, rant); err != nil {
		return nil, err
	}
	rant.IsLocked = rant.PasswordHash != nil
	return rant, nil
}

func (s *RantService) Get(ctx context.Context, id, userID, password string) (*models.Rant, error) {
	rant, err := s.rantRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if rant.IsLocked {
		if password == "" {
			// Return without content
			rant.Content = ""
			return rant, ErrRantLocked
		}
		if err := auth.CheckPassword(password, *rant.PasswordHash); err != nil {
			rant.Content = ""
			return rant, ErrRantLocked
		}
	}

	return rant, nil
}

type UpdateRantInput struct {
	Title    *string
	Content  string
	Password *string
}

func (s *RantService) Update(ctx context.Context, id, userID string, input UpdateRantInput) (*models.Rant, error) {
	var passwordHash *string
	if input.Password != nil && *input.Password != "" {
		hash, err := auth.HashPassword(*input.Password)
		if err != nil {
			return nil, err
		}
		passwordHash = &hash
	} else if input.Password != nil && *input.Password == "" {
		// Empty password means remove the lock; passwordHash stays nil
	}

	return s.rantRepo.Update(ctx, id, userID, input.Title, input.Content, passwordHash)
}

func (s *RantService) Delete(ctx context.Context, id, userID string) error {
	return s.rantRepo.Delete(ctx, id, userID)
}
