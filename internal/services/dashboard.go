package services

import (
	"context"

	"github.com/rawdah/rawdah-api/internal/repository"
)

type DashboardService struct {
	dashRepo *repository.DashboardRepo
}

func NewDashboardService(dashRepo *repository.DashboardRepo) *DashboardService {
	return &DashboardService{dashRepo: dashRepo}
}

func (s *DashboardService) Summary(ctx context.Context, familyID string) (*repository.DashboardSummary, error) {
	return s.dashRepo.Summary(ctx, familyID)
}

func (s *DashboardService) TaskCompletion(ctx context.Context, familyID string, days int) ([]*repository.DailyTaskCompletion, error) {
	if days == 0 {
		days = 30
	}
	return s.dashRepo.TaskCompletion(ctx, familyID, days)
}

func (s *DashboardService) GameTime(ctx context.Context, familyID string, days int) ([]*repository.DailyGameTime, error) {
	if days == 0 {
		days = 30
	}
	return s.dashRepo.GameTime(ctx, familyID, days)
}

func (s *DashboardService) QuizScores(ctx context.Context, familyID string, days int) ([]*repository.QuizScoreEntry, error) {
	if days == 0 {
		days = 30
	}
	return s.dashRepo.QuizScores(ctx, familyID, days)
}

func (s *DashboardService) LearnProgress(ctx context.Context, familyID string) ([]*repository.LearnProgressEntry, error) {
	return s.dashRepo.LearnProgress(ctx, familyID)
}
