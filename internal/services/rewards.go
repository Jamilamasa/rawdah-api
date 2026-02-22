package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/repository"
)

type RewardService struct {
	rewardRepo *repository.RewardRepo
}

func NewRewardService(rewardRepo *repository.RewardRepo) *RewardService {
	return &RewardService{rewardRepo: rewardRepo}
}

func (s *RewardService) List(ctx context.Context, familyID string) ([]*models.Reward, error) {
	return s.rewardRepo.List(ctx, familyID)
}

type CreateRewardInput struct {
	FamilyID    uuid.UUID
	Title       string
	Description *string
	Value       float64
	Type        string
	Icon        *string
	CreatedBy   uuid.UUID
}

func (s *RewardService) Create(ctx context.Context, input CreateRewardInput) (*models.Reward, error) {
	if input.Type == "" {
		input.Type = "virtual"
	}
	reward := &models.Reward{
		FamilyID:    input.FamilyID,
		Title:       input.Title,
		Description: input.Description,
		Value:       input.Value,
		Type:        input.Type,
		Icon:        input.Icon,
		CreatedBy:   input.CreatedBy,
	}
	if err := s.rewardRepo.Create(ctx, reward); err != nil {
		return nil, err
	}
	return reward, nil
}

func (s *RewardService) GetByID(ctx context.Context, id, familyID string) (*models.Reward, error) {
	return s.rewardRepo.GetByID(ctx, id, familyID)
}

type UpdateRewardInput struct {
	Title       string
	Description *string
	Value       float64
	Type        string
	Icon        *string
}

func (s *RewardService) Update(ctx context.Context, id, familyID string, input UpdateRewardInput) (*models.Reward, error) {
	return s.rewardRepo.Update(ctx, id, familyID, input.Title, input.Description, input.Value, input.Type, input.Icon)
}

func (s *RewardService) Delete(ctx context.Context, id, familyID string) error {
	return s.rewardRepo.Delete(ctx, id, familyID)
}
