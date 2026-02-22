package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/repository"
	"github.com/rawdah/rawdah-api/internal/ws"
)

type LessonService struct {
	lessonRepo *repository.LessonRepo
	quizRepo   *repository.QuizRepo
	quranRepo  *repository.QuranRepo
	familyRepo *repository.FamilyRepo
	xpSvc      *XPService
	hub        *ws.Hub
}

func NewLessonService(lessonRepo *repository.LessonRepo, quizRepo *repository.QuizRepo, quranRepo *repository.QuranRepo, familyRepo *repository.FamilyRepo, xpSvc *XPService, hub *ws.Hub) *LessonService {
	return &LessonService{
		lessonRepo: lessonRepo,
		quizRepo:   quizRepo,
		quranRepo:  quranRepo,
		familyRepo: familyRepo,
		xpSvc:      xpSvc,
		hub:        hub,
	}
}

var (
	ErrInvalidLessonAssignee = errors.New("invalid lesson assignee")
	ErrInvalidLearnData      = errors.New("invalid learn content")
)

type AssignLessonInput struct {
	FamilyID   uuid.UUID
	VerseID    uuid.UUID
	AssignedTo uuid.UUID
	AssignedBy uuid.UUID
	RewardID   *uuid.UUID
}

func (s *LessonService) AssignLesson(ctx context.Context, input AssignLessonInput) (*models.QuranLesson, error) {
	assignee, err := s.familyRepo.GetMemberByID(ctx, input.AssignedTo.String(), input.FamilyID.String())
	if err != nil || !assignee.IsActive || assignee.Role != "child" {
		return nil, ErrInvalidLessonAssignee
	}

	lesson := &models.QuranLesson{
		FamilyID:   input.FamilyID,
		VerseID:    input.VerseID,
		AssignedTo: input.AssignedTo,
		AssignedBy: input.AssignedBy,
		RewardID:   input.RewardID,
		Status:     "pending",
	}
	if err := s.lessonRepo.CreateLesson(ctx, lesson); err != nil {
		return nil, err
	}

	s.hub.Broadcast(ws.BroadcastMsg{
		Room:  fmt.Sprintf("user:%s", input.AssignedTo.String()),
		Event: ws.WSEvent{Type: "lesson.assigned", Payload: lesson},
	})

	return lesson, nil
}

func (s *LessonService) GetLesson(ctx context.Context, id, familyID string) (*models.QuranLesson, error) {
	return s.lessonRepo.GetLesson(ctx, id, familyID)
}

func (s *LessonService) ListLessons(ctx context.Context, familyID string) ([]*models.QuranLesson, error) {
	return s.lessonRepo.ListLessons(ctx, familyID)
}

func (s *LessonService) ListMyLessons(ctx context.Context, userID, familyID string) ([]*models.QuranLesson, error) {
	return s.lessonRepo.ListMyLessons(ctx, userID, familyID)
}

func (s *LessonService) CompleteLesson(ctx context.Context, id, userID, familyID string) (*models.QuranLesson, error) {
	lessonUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid lesson id")
	}

	if err := s.lessonRepo.CompleteLesson(ctx, lessonUUID, userID, familyID); err != nil {
		return nil, err
	}

	lesson, err := s.lessonRepo.GetLesson(ctx, id, familyID)
	if err != nil {
		return nil, err
	}

	// Award 30 XP
	go func() {
		_ = s.xpSvc.AwardXP(context.Background(), userID, familyID, 30, "quran_lesson", lessonUUID)
	}()

	s.hub.Broadcast(ws.BroadcastMsg{
		Room:  fmt.Sprintf("family:%s", familyID),
		Event: ws.WSEvent{Type: "lesson.completed", Payload: lesson},
	})

	return lesson, nil
}

// Learn Content

type CreateLearnContentInput struct {
	FamilyID    uuid.UUID
	AssignedTo  *uuid.UUID
	Title       string
	ContentType string
	Content     string
	RewardID    *uuid.UUID
	CreatedBy   uuid.UUID
}

func (s *LessonService) CreateLearnContent(ctx context.Context, input CreateLearnContentInput) (*models.LearnContent, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" || len(title) > 160 {
		return nil, ErrInvalidLearnData
	}
	contentText := strings.TrimSpace(input.Content)
	if contentText == "" {
		return nil, ErrInvalidLearnData
	}

	switch input.ContentType {
	case "text", "link", "video", "pdf":
	default:
		return nil, ErrInvalidLearnData
	}

	if input.AssignedTo != nil {
		assignee, err := s.familyRepo.GetMemberByID(ctx, input.AssignedTo.String(), input.FamilyID.String())
		if err != nil || !assignee.IsActive || assignee.Role != "child" {
			return nil, ErrInvalidLessonAssignee
		}
	}

	content := &models.LearnContent{
		FamilyID:    input.FamilyID,
		AssignedTo:  input.AssignedTo,
		Title:       title,
		ContentType: input.ContentType,
		Content:     contentText,
		RewardID:    input.RewardID,
		CreatedBy:   input.CreatedBy,
	}
	if err := s.lessonRepo.CreateLearnContent(ctx, content); err != nil {
		return nil, err
	}
	return content, nil
}

func (s *LessonService) GetLearnContent(ctx context.Context, familyID string) ([]*models.LearnContent, error) {
	return s.lessonRepo.GetLearnContent(ctx, familyID)
}

func (s *LessonService) GetMyLearnContent(ctx context.Context, userID, familyID string) ([]*models.LearnContent, error) {
	return s.lessonRepo.GetMyLearnContent(ctx, userID, familyID)
}

func (s *LessonService) CompleteLearnContent(ctx context.Context, contentID, userID, familyID string) error {
	cID, err := uuid.Parse(contentID)
	if err != nil {
		return fmt.Errorf("invalid content id")
	}
	return s.lessonRepo.CompleteLearnContent(ctx, cID, userID, familyID)
}
