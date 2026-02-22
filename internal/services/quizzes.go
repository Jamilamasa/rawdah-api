package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/ai"
	"github.com/rawdah/rawdah-api/internal/mailer"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/repository"
	"github.com/rawdah/rawdah-api/internal/ws"
)

type QuizService struct {
	quizRepo    *repository.QuizRepo
	hadithRepo  *repository.HadithRepo
	prophetRepo *repository.ProphetRepo
	quranRepo   *repository.QuranRepo
	familyRepo  *repository.FamilyRepo
	notifRepo   *repository.NotificationRepo
	xpSvc       *XPService
	aiClient    *ai.Client
	mailer      *mailer.Mailer
	hub         *ws.Hub
}

var ErrInvalidQuizAssignee = errors.New("invalid quiz assignee")

func NewQuizService(
	quizRepo *repository.QuizRepo,
	hadithRepo *repository.HadithRepo,
	prophetRepo *repository.ProphetRepo,
	quranRepo *repository.QuranRepo,
	familyRepo *repository.FamilyRepo,
	notifRepo *repository.NotificationRepo,
	xpSvc *XPService,
	aiClient *ai.Client,
	m *mailer.Mailer,
	hub *ws.Hub,
) *QuizService {
	return &QuizService{
		quizRepo:    quizRepo,
		hadithRepo:  hadithRepo,
		prophetRepo: prophetRepo,
		quranRepo:   quranRepo,
		familyRepo:  familyRepo,
		notifRepo:   notifRepo,
		xpSvc:       xpSvc,
		aiClient:    aiClient,
		mailer:      m,
		hub:         hub,
	}
}

type AssignHadithInput struct {
	FamilyID      uuid.UUID
	HadithID      uuid.UUID
	AssignedTo    uuid.UUID
	AssignedBy    uuid.UUID
	ChildAge      int
	MemorizeUntil *time.Time
}

func (s *QuizService) AssignHadith(ctx context.Context, input AssignHadithInput) (*models.HadithQuiz, error) {
	assignee, err := s.familyRepo.GetMemberByID(ctx, input.AssignedTo.String(), input.FamilyID.String())
	if err != nil || !assignee.IsActive || assignee.Role != "child" {
		return nil, ErrInvalidQuizAssignee
	}

	hadith, err := s.hadithRepo.GetByID(ctx, input.HadithID.String())
	if err != nil {
		return nil, fmt.Errorf("hadith not found")
	}

	prompt := ai.BuildHadithPrompt(*hadith, input.ChildAge)
	questions, err := s.aiClient.GenerateQuiz(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate quiz: %w", err)
	}

	quiz := &models.HadithQuiz{
		FamilyID:      input.FamilyID,
		HadithID:      input.HadithID,
		AssignedTo:    input.AssignedTo,
		AssignedBy:    input.AssignedBy,
		Questions:     questions,
		Status:        "pending",
		MemorizeUntil: input.MemorizeUntil,
	}

	if err := s.quizRepo.CreateHadithQuiz(ctx, quiz); err != nil {
		return nil, err
	}

	s.hub.Broadcast(ws.BroadcastMsg{
		Room:  fmt.Sprintf("user:%s", input.AssignedTo.String()),
		Event: ws.WSEvent{Type: "quiz.assigned", Payload: quiz},
	})

	notif := &models.Notification{
		UserID: input.AssignedTo,
		Type:   "quiz_assigned",
		Title:  "New Hadith Quiz",
		Body:   fmt.Sprintf("You have a new hadith quiz from %s", hadith.Source),
	}
	_ = s.notifRepo.Create(ctx, notif)

	go func() {
		member, err := s.familyRepo.GetMemberByID(context.Background(), input.AssignedTo.String(), input.FamilyID.String())
		if err != nil || member.Email == nil {
			return
		}
		subject := "New Hadith Quiz Assigned"
		body := fmt.Sprintf("<p>You have been assigned a new hadith quiz. Source: <strong>%s</strong></p>", hadith.Source)
		html := mailer.BuildEmail("New Quiz", body, "Start Quiz", "https://kids.rawdah.app/quizzes", "")
		_ = s.mailer.Send(mailer.BrevoContact{Name: member.Name, Email: *member.Email}, subject, html)
	}()

	return quiz, nil
}

type AssignProphetInput struct {
	FamilyID   uuid.UUID
	ProphetID  uuid.UUID
	AssignedTo uuid.UUID
	AssignedBy uuid.UUID
	ChildAge   int
}

func (s *QuizService) AssignProphet(ctx context.Context, input AssignProphetInput) (*models.ProphetQuiz, error) {
	assignee, err := s.familyRepo.GetMemberByID(ctx, input.AssignedTo.String(), input.FamilyID.String())
	if err != nil || !assignee.IsActive || assignee.Role != "child" {
		return nil, ErrInvalidQuizAssignee
	}

	prophet, err := s.prophetRepo.GetByID(ctx, input.ProphetID.String())
	if err != nil {
		return nil, fmt.Errorf("prophet not found")
	}

	prompt := ai.BuildProphetPrompt(*prophet, input.ChildAge)
	questions, err := s.aiClient.GenerateQuiz(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate quiz: %w", err)
	}

	quiz := &models.ProphetQuiz{
		FamilyID:   input.FamilyID,
		ProphetID:  input.ProphetID,
		AssignedTo: input.AssignedTo,
		AssignedBy: input.AssignedBy,
		Questions:  questions,
		Status:     "pending",
	}

	if err := s.quizRepo.CreateProphetQuiz(ctx, quiz); err != nil {
		return nil, err
	}

	s.hub.Broadcast(ws.BroadcastMsg{
		Room:  fmt.Sprintf("user:%s", input.AssignedTo.String()),
		Event: ws.WSEvent{Type: "quiz.assigned", Payload: quiz},
	})

	notif := &models.Notification{
		UserID: input.AssignedTo,
		Type:   "quiz_assigned",
		Title:  "New Prophet Quiz",
		Body:   fmt.Sprintf("You have a new quiz about Prophet %s", prophet.NameEn),
	}
	_ = s.notifRepo.Create(ctx, notif)

	go func() {
		member, err := s.familyRepo.GetMemberByID(context.Background(), input.AssignedTo.String(), input.FamilyID.String())
		if err != nil || member.Email == nil {
			return
		}
		subject := "New Prophet Quiz Assigned"
		body := fmt.Sprintf("<p>You have been assigned a quiz about Prophet <strong>%s</strong>.</p>", prophet.NameEn)
		html := mailer.BuildEmail("New Quiz", body, "Start Quiz", "https://kids.rawdah.app/quizzes", "")
		_ = s.mailer.Send(mailer.BrevoContact{Name: member.Name, Email: *member.Email}, subject, html)
	}()

	return quiz, nil
}

type AssignQuranInput struct {
	FamilyID   uuid.UUID
	VerseID    uuid.UUID
	LessonID   *uuid.UUID
	AssignedTo uuid.UUID
	AssignedBy uuid.UUID
	ChildAge   int
}

func (s *QuizService) AssignQuran(ctx context.Context, input AssignQuranInput) (*models.QuranQuiz, error) {
	assignee, err := s.familyRepo.GetMemberByID(ctx, input.AssignedTo.String(), input.FamilyID.String())
	if err != nil || !assignee.IsActive || assignee.Role != "child" {
		return nil, ErrInvalidQuizAssignee
	}

	verse, err := s.quranRepo.GetVerseByID(ctx, input.VerseID.String())
	if err != nil {
		return nil, fmt.Errorf("verse not found")
	}

	prompt := ai.BuildQuranPrompt(*verse, input.ChildAge)
	questions, err := s.aiClient.GenerateQuiz(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate quiz: %w", err)
	}

	quiz := &models.QuranQuiz{
		FamilyID:   input.FamilyID,
		VerseID:    input.VerseID,
		LessonID:   input.LessonID,
		AssignedTo: input.AssignedTo,
		Questions:  questions,
		Status:     "pending",
	}

	if err := s.quizRepo.CreateQuranQuiz(ctx, quiz); err != nil {
		return nil, err
	}

	s.hub.Broadcast(ws.BroadcastMsg{
		Room:  fmt.Sprintf("user:%s", input.AssignedTo.String()),
		Event: ws.WSEvent{Type: "quiz.assigned", Payload: quiz},
	})

	return quiz, nil
}

func (s *QuizService) StartHadithQuiz(ctx context.Context, id, familyID, userID string) error {
	quiz, err := s.quizRepo.GetHadithQuizByID(ctx, id, familyID)
	if err != nil {
		return fmt.Errorf("quiz not found")
	}
	if quiz.AssignedTo.String() != userID {
		return fmt.Errorf("not authorized")
	}
	return s.quizRepo.UpdateHadithQuizStatus(ctx, id, familyID, "in_progress")
}

func (s *QuizService) StartProphetQuiz(ctx context.Context, id, familyID, userID string) error {
	quiz, err := s.quizRepo.GetProphetQuizByID(ctx, id, familyID)
	if err != nil {
		return fmt.Errorf("quiz not found")
	}
	if quiz.AssignedTo.String() != userID {
		return fmt.Errorf("not authorized")
	}
	return s.quizRepo.UpdateProphetQuizStatus(ctx, id, familyID, "in_progress")
}

func (s *QuizService) StartQuranQuiz(ctx context.Context, id, familyID, userID string) error {
	quiz, err := s.quizRepo.GetQuranQuizByID(ctx, id, familyID)
	if err != nil {
		return fmt.Errorf("quiz not found")
	}
	if quiz.AssignedTo.String() != userID {
		return fmt.Errorf("not authorized")
	}
	return s.quizRepo.UpdateQuranQuizStatus(ctx, id, familyID, "in_progress")
}

type SubmitAnswersInput struct {
	Answers []models.QuizAnswer
	UserID  string
}

func (s *QuizService) SubmitHadithQuiz(ctx context.Context, id, familyID string, input SubmitAnswersInput) (*models.HadithQuiz, error) {
	quiz, err := s.quizRepo.GetHadithQuizByID(ctx, id, familyID)
	if err != nil {
		return nil, fmt.Errorf("quiz not found")
	}
	if quiz.AssignedTo.String() != input.UserID {
		return nil, fmt.Errorf("not authorized")
	}

	score := gradeAnswers(quiz.Questions, input.Answers)
	xp := calculateXP("hadith", score)

	if err := s.quizRepo.SubmitHadithQuiz(ctx, id, familyID, input.Answers, score, xp); err != nil {
		return nil, err
	}

	go func() {
		_ = s.xpSvc.AwardXP(context.Background(), input.UserID, familyID, xp, "hadith_quiz", quiz.ID)
	}()

	updated, _ := s.quizRepo.GetHadithQuizByID(ctx, id, familyID)

	s.hub.Broadcast(ws.BroadcastMsg{
		Room:  fmt.Sprintf("family:%s", familyID),
		Event: ws.WSEvent{Type: "quiz.completed", Payload: updated},
	})

	return updated, nil
}

func (s *QuizService) SubmitProphetQuiz(ctx context.Context, id, familyID string, input SubmitAnswersInput) (*models.ProphetQuiz, error) {
	quiz, err := s.quizRepo.GetProphetQuizByID(ctx, id, familyID)
	if err != nil {
		return nil, fmt.Errorf("quiz not found")
	}
	if quiz.AssignedTo.String() != input.UserID {
		return nil, fmt.Errorf("not authorized")
	}

	score := gradeAnswers(quiz.Questions, input.Answers)
	xp := calculateXP("prophet", score)

	if err := s.quizRepo.SubmitProphetQuiz(ctx, id, familyID, input.Answers, score, xp); err != nil {
		return nil, err
	}

	go func() {
		_ = s.xpSvc.AwardXP(context.Background(), input.UserID, familyID, xp, "prophet_quiz", quiz.ID)
	}()

	updated, _ := s.quizRepo.GetProphetQuizByID(ctx, id, familyID)

	s.hub.Broadcast(ws.BroadcastMsg{
		Room:  fmt.Sprintf("family:%s", familyID),
		Event: ws.WSEvent{Type: "quiz.completed", Payload: updated},
	})

	return updated, nil
}

func (s *QuizService) SubmitQuranQuiz(ctx context.Context, id, familyID string, input SubmitAnswersInput) (*models.QuranQuiz, error) {
	quiz, err := s.quizRepo.GetQuranQuizByID(ctx, id, familyID)
	if err != nil {
		return nil, fmt.Errorf("quiz not found")
	}
	if quiz.AssignedTo.String() != input.UserID {
		return nil, fmt.Errorf("not authorized")
	}

	score := gradeAnswers(quiz.Questions, input.Answers)
	xp := calculateXP("quran", score)

	if err := s.quizRepo.SubmitQuranQuiz(ctx, id, familyID, input.Answers, score, xp); err != nil {
		return nil, err
	}

	go func() {
		_ = s.xpSvc.AwardXP(context.Background(), input.UserID, familyID, xp, "quran_quiz", quiz.ID)
	}()

	updated, _ := s.quizRepo.GetQuranQuizByID(ctx, id, familyID)

	s.hub.Broadcast(ws.BroadcastMsg{
		Room:  fmt.Sprintf("family:%s", familyID),
		Event: ws.WSEvent{Type: "quiz.completed", Payload: updated},
	})

	return updated, nil
}

func (s *QuizService) GetHadithQuiz(ctx context.Context, id, familyID string) (*models.HadithQuiz, error) {
	return s.quizRepo.GetHadithQuizByID(ctx, id, familyID)
}

func (s *QuizService) GetProphetQuiz(ctx context.Context, id, familyID string) (*models.ProphetQuiz, error) {
	return s.quizRepo.GetProphetQuizByID(ctx, id, familyID)
}

func (s *QuizService) GetQuranQuiz(ctx context.Context, id, familyID string) (*models.QuranQuiz, error) {
	return s.quizRepo.GetQuranQuizByID(ctx, id, familyID)
}

func (s *QuizService) ListHadithQuizzes(ctx context.Context, familyID string) ([]*models.HadithQuiz, error) {
	return s.quizRepo.ListHadithQuizzesByFamily(ctx, familyID)
}

func (s *QuizService) ListProphetQuizzes(ctx context.Context, familyID string) ([]*models.ProphetQuiz, error) {
	return s.quizRepo.ListProphetQuizzesByFamily(ctx, familyID)
}

func (s *QuizService) ListQuranQuizzes(ctx context.Context, familyID string) ([]*models.QuranQuiz, error) {
	return s.quizRepo.ListQuranQuizzesByFamily(ctx, familyID)
}

func (s *QuizService) ListMyHadithQuizzes(ctx context.Context, userID, familyID string) ([]*models.HadithQuiz, error) {
	return s.quizRepo.ListMyHadithQuizzes(ctx, userID, familyID)
}

func (s *QuizService) ListMyProphetQuizzes(ctx context.Context, userID, familyID string) ([]*models.ProphetQuiz, error) {
	return s.quizRepo.ListMyProphetQuizzes(ctx, userID, familyID)
}

func (s *QuizService) ListMyQuranQuizzes(ctx context.Context, userID, familyID string) ([]*models.QuranQuiz, error) {
	return s.quizRepo.ListMyQuranQuizzes(ctx, userID, familyID)
}

func gradeAnswers(questions []models.QuizQuestion, answers []models.QuizAnswer) int {
	if len(questions) == 0 {
		return 0
	}
	// Build answer map
	answerMap := make(map[string]string)
	for _, a := range answers {
		answerMap[a.QuestionID] = a.SelectedAnswer
	}

	correct := 0
	for _, q := range questions {
		if answerMap[q.ID] == q.CorrectAnswer {
			correct++
		}
	}

	return (correct * 100) / len(questions)
}

func calculateXP(quizType string, score int) int {
	baseXP := map[string]int{
		"hadith":  50,
		"prophet": 40,
		"quran":   40,
	}
	bonusXP := map[string]int{
		"hadith":  25,
		"prophet": 20,
		"quran":   20,
	}

	base := baseXP[quizType]
	if base == 0 {
		base = 30
	}
	bonus := bonusXP[quizType]
	if bonus == 0 {
		bonus = 15
	}

	if score == 100 {
		return base + bonus
	}
	if score >= 70 {
		return base
	}
	if score >= 50 {
		return base / 2
	}
	return 0
}
