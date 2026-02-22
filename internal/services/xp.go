package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/repository"
)

type XPService struct {
	xpRepo     *repository.XPRepo
	quizRepo   *repository.QuizRepo
	lessonRepo *repository.LessonRepo
}

func NewXPService(xpRepo *repository.XPRepo, quizRepo *repository.QuizRepo, lessonRepo *repository.LessonRepo) *XPService {
	return &XPService{
		xpRepo:     xpRepo,
		quizRepo:   quizRepo,
		lessonRepo: lessonRepo,
	}
}

func (s *XPService) AwardXP(ctx context.Context, userID, familyID string, amount int, source string, sourceID uuid.UUID) error {
	if err := s.xpRepo.InsertXPEvent(ctx, userID, familyID, source, sourceID, amount); err != nil {
		return err
	}
	_, err := s.xpRepo.AddXP(ctx, userID, familyID, amount)
	if err != nil {
		return err
	}

	go s.CheckAndAwardBadges(context.Background(), userID, familyID)
	return nil
}

func (s *XPService) CheckAndAwardBadges(ctx context.Context, userID, familyID string) {
	checks := []struct {
		slug    string
		checkFn func() (bool, error)
	}{
		{
			slug: "first_task",
			checkFn: func() (bool, error) {
				count, err := s.xpRepo.CountCompletedTasks(ctx, userID, familyID)
				return count >= 1, err
			},
		},
		{
			slug: "first_quiz",
			checkFn: func() (bool, error) {
				count, err := s.quizRepo.CountAllCompletedQuizzes(ctx, userID, familyID)
				return count >= 1, err
			},
		},
		{
			slug: "hadith_scholar",
			checkFn: func() (bool, error) {
				count, err := s.quizRepo.CountCompletedHadithQuizzes(ctx, userID, familyID)
				return count >= 10, err
			},
		},
		{
			slug: "prophet_explorer",
			checkFn: func() (bool, error) {
				count, err := s.quizRepo.CountDistinctCompletedProphetQuizzes(ctx, userID, familyID)
				return count >= 10, err
			},
		},
		{
			slug: "quran_learner",
			checkFn: func() (bool, error) {
				count, err := s.lessonRepo.CountCompletedLessons(ctx, userID, familyID)
				return count >= 5, err
			},
		},
		{
			slug: "perfect_score",
			checkFn: func() (bool, error) {
				return s.quizRepo.HasPerfectScoreQuiz(ctx, userID, familyID)
			},
		},
		{
			slug: "game_player",
			checkFn: func() (bool, error) {
				count, err := s.xpRepo.CountGameSessions(ctx, userID, familyID)
				return count >= 1, err
			},
		},
		{
			slug: "messenger",
			checkFn: func() (bool, error) {
				count, err := s.xpRepo.CountMessagesSent(ctx, userID, familyID)
				return count >= 1, err
			},
		},
		{
			slug: "streak_7",
			checkFn: func() (bool, error) {
				return s.xpRepo.CheckStreak(ctx, userID, familyID, 7)
			},
		},
		{
			slug: "streak_30",
			checkFn: func() (bool, error) {
				return s.xpRepo.CheckStreak(ctx, userID, familyID, 30)
			},
		},
	}

	for _, check := range checks {
		badge, err := s.xpRepo.GetBadge(ctx, check.slug)
		if err != nil {
			continue
		}

		hasIt, err := s.xpRepo.HasBadge(ctx, userID, badge.ID)
		if err != nil || hasIt {
			continue
		}

		earned, err := check.checkFn()
		if err != nil || !earned {
			continue
		}

		_ = s.xpRepo.AwardBadge(ctx, userID, badge.ID)

		// Award XP for badge
		if badge.XPReward > 0 {
			_, _ = s.xpRepo.AddXP(ctx, userID, familyID, badge.XPReward)
		}
	}
}
