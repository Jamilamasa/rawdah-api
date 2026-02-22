package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
var ErrInvalidTopicQuizData = errors.New("invalid topic quiz data")

const (
	defaultTopicQuestionCount = 20
	minTopicQuestionCount     = 15
	maxTopicQuestionCount     = 40
)

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
	AssignedTo    uuid.UUID
	AssignedBy    uuid.UUID
	Difficulty    string
	MemorizeUntil *time.Time
}

func (s *QuizService) AssignHadith(ctx context.Context, input AssignHadithInput) (*models.HadithQuiz, error) {
	assignee, err := s.familyRepo.GetMemberByID(ctx, input.AssignedTo.String(), input.FamilyID.String())
	if err != nil || !assignee.IsActive || assignee.Role != "child" {
		return nil, ErrInvalidQuizAssignee
	}

	difficulty := input.Difficulty
	if difficulty == "" {
		difficulty = "easy"
	}

	childAge := childAgeFromDateOfBirth(assignee.DateOfBirth)
	prompt := ai.BuildHadithPrompt(childAge, difficulty)
	result, err := s.aiClient.GenerateHadithQuiz(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate quiz: %w", err)
	}

	// Persist the AI-selected hadith so the FK constraint on hadith_quizzes is satisfied.
	h := &models.Hadith{
		TextEn:     result.Hadith.TextEn,
		Source:     result.Hadith.Source,
		Difficulty: difficulty,
	}
	if result.Hadith.TextAr != "" {
		h.TextAr = &result.Hadith.TextAr
	}
	if result.Hadith.Topic != "" {
		h.Topic = &result.Hadith.Topic
	}
	hadith, err := s.hadithRepo.Insert(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("failed to store hadith: %w", err)
	}

	quiz := &models.HadithQuiz{
		FamilyID:      input.FamilyID,
		HadithID:      hadith.ID,
		AssignedTo:    input.AssignedTo,
		AssignedBy:    input.AssignedBy,
		Questions:     result.Questions,
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

type SelfAssignHadithInput struct {
	FamilyID   uuid.UUID
	UserID     uuid.UUID
	Difficulty string
}

// SelfAssignHadith lets a child initiate their own hadith quiz without parent involvement.
// The AI picks an authentic hadith and generates questions based on the child's profile.
func (s *QuizService) SelfAssignHadith(ctx context.Context, input SelfAssignHadithInput) (*models.HadithQuiz, error) {
	member, err := s.familyRepo.GetMemberByID(ctx, input.UserID.String(), input.FamilyID.String())
	if err != nil || !member.IsActive || member.Role != "child" {
		return nil, ErrInvalidQuizAssignee
	}

	return s.AssignHadith(ctx, AssignHadithInput{
		FamilyID:   input.FamilyID,
		AssignedTo: input.UserID,
		AssignedBy: input.UserID,
		Difficulty: input.Difficulty,
	})
}

type AssignProphetInput struct {
	FamilyID   uuid.UUID
	ProphetID  uuid.UUID
	AssignedTo uuid.UUID
	AssignedBy uuid.UUID
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

	childAge := childAgeFromDateOfBirth(assignee.DateOfBirth)
	prompt := ai.BuildProphetPrompt(*prophet, childAge)
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

	childAge := childAgeFromDateOfBirth(assignee.DateOfBirth)
	prompt := ai.BuildQuranPrompt(*verse, childAge)
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

type AssignTopicInput struct {
	FamilyID      uuid.UUID
	AssignedTo    uuid.UUID
	AssignedBy    uuid.UUID
	Category      string
	Topic         string
	QuestionCount int
}

func (s *QuizService) AssignTopic(ctx context.Context, input AssignTopicInput) (*models.TopicQuiz, error) {
	assignee, err := s.familyRepo.GetMemberByID(ctx, input.AssignedTo.String(), input.FamilyID.String())
	if err != nil || !assignee.IsActive || assignee.Role != "child" {
		return nil, ErrInvalidQuizAssignee
	}

	category := normalizeTopicCategory(input.Category)
	if category == "" {
		return nil, ErrInvalidTopicQuizData
	}

	topic := strings.TrimSpace(input.Topic)
	if topic == "" || len(topic) > 180 {
		return nil, ErrInvalidTopicQuizData
	}

	questionCount := input.QuestionCount
	if questionCount == 0 {
		questionCount = defaultTopicQuestionCount
	}
	if questionCount < minTopicQuestionCount || questionCount > maxTopicQuestionCount {
		return nil, ErrInvalidTopicQuizData
	}

	childAge := childAgeFromDateOfBirth(assignee.DateOfBirth)
	prompt := ai.BuildTopicPackPrompt(category, topic, childAge, questionCount)
	pack, err := s.aiClient.GenerateTopicPack(prompt, questionCount)
	if err != nil {
		return nil, fmt.Errorf("failed to generate topic quiz pack: %w", err)
	}

	quiz := &models.TopicQuiz{
		FamilyID:      input.FamilyID,
		AssignedTo:    input.AssignedTo,
		AssignedBy:    input.AssignedBy,
		Category:      category,
		Topic:         topic,
		LessonContent: pack.LessonContent,
		Flashcards:    pack.Flashcards,
		Questions:     pack.Questions,
		Status:        "pending",
	}

	if err := s.quizRepo.CreateTopicQuiz(ctx, quiz); err != nil {
		return nil, err
	}

	s.hub.Broadcast(ws.BroadcastMsg{
		Room:  fmt.Sprintf("user:%s", input.AssignedTo.String()),
		Event: ws.WSEvent{Type: "quiz.assigned", Payload: quiz},
	})

	notif := &models.Notification{
		UserID: input.AssignedTo,
		Type:   "quiz_assigned",
		Title:  "New Learning Quiz",
		Body:   fmt.Sprintf("You have a new %s learning quiz: %s", toTitleCase(category), topic),
	}
	_ = s.notifRepo.Create(ctx, notif)

	go func() {
		member, err := s.familyRepo.GetMemberByID(context.Background(), input.AssignedTo.String(), input.FamilyID.String())
		if err != nil || member.Email == nil {
			return
		}
		subject := "New Learning Quiz Assigned"
		body := fmt.Sprintf("<p>You have been assigned a new <strong>%s</strong> learning quiz on <strong>%s</strong>.</p>", toTitleCase(category), topic)
		html := mailer.BuildEmail("New Learning Quiz", body, "Start Learning", "https://kids.rawdah.app/quizzes", "")
		_ = s.mailer.Send(mailer.BrevoContact{Name: member.Name, Email: *member.Email}, subject, html)
	}()

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

func (s *QuizService) StartTopicQuiz(ctx context.Context, id, familyID, userID string) error {
	quiz, err := s.quizRepo.GetTopicQuizByID(ctx, id, familyID)
	if err != nil {
		return fmt.Errorf("quiz not found")
	}
	if quiz.AssignedTo.String() != userID {
		return fmt.Errorf("not authorized")
	}
	return s.quizRepo.UpdateTopicQuizStatus(ctx, id, familyID, "in_progress")
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

func (s *QuizService) SubmitTopicQuiz(ctx context.Context, id, familyID string, input SubmitAnswersInput) (*models.TopicQuiz, error) {
	quiz, err := s.quizRepo.GetTopicQuizByID(ctx, id, familyID)
	if err != nil {
		return nil, fmt.Errorf("quiz not found")
	}
	if quiz.AssignedTo.String() != input.UserID {
		return nil, fmt.Errorf("not authorized")
	}

	score := gradeAnswers(quiz.Questions, input.Answers)
	xp := calculateXP("topic", score)

	if err := s.quizRepo.SubmitTopicQuiz(ctx, id, familyID, input.Answers, score, xp); err != nil {
		return nil, err
	}

	go func() {
		_ = s.xpSvc.AwardXP(context.Background(), input.UserID, familyID, xp, "topic_quiz", quiz.ID)
	}()

	updated, _ := s.quizRepo.GetTopicQuizByID(ctx, id, familyID)

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

func (s *QuizService) GetTopicQuiz(ctx context.Context, id, familyID string) (*models.TopicQuiz, error) {
	return s.quizRepo.GetTopicQuizByID(ctx, id, familyID)
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

func (s *QuizService) ListTopicQuizzes(ctx context.Context, familyID string) ([]*models.TopicQuiz, error) {
	return s.quizRepo.ListTopicQuizzesByFamily(ctx, familyID)
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

func (s *QuizService) ListMyTopicQuizzes(ctx context.Context, userID, familyID string) ([]*models.TopicQuiz, error) {
	return s.quizRepo.ListMyTopicQuizzes(ctx, userID, familyID)
}

func childAgeFromDateOfBirth(dateOfBirth *time.Time) int {
	defaultAge := 10
	if dateOfBirth == nil {
		return defaultAge
	}

	now := time.Now().UTC()
	age := now.Year() - dateOfBirth.Year()
	if now.Month() < dateOfBirth.Month() || (now.Month() == dateOfBirth.Month() && now.Day() < dateOfBirth.Day()) {
		age--
	}

	if age < minChildAge || age > maxChildAge {
		return defaultAge
	}
	return age
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
	base := map[string]int{
		"hadith":  50,
		"prophet": 40,
		"quran":   40,
		"topic":   60,
	}[quizType]
	if base == 0 {
		base = 30
	}
	bonus := map[string]int{
		"hadith":  25,
		"prophet": 20,
		"quran":   20,
		"topic":   30,
	}[quizType]
	if bonus == 0 {
		bonus = 15
	}
	if score == 100 {
		return base + bonus
	}
	return base
}

func normalizeTopicCategory(category string) string {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "hadith":
		return "hadith"
	case "quran":
		return "quran"
	case "science":
		return "science"
	case "fun_facts", "fun-facts", "funfacts":
		return "fun_facts"
	case "any":
		return "general"
	case "custom":
		return "custom"
	case "general":
		return "general"
	default:
		return ""
	}
}

func toTitleCase(value string) string {
	parts := strings.Split(strings.ReplaceAll(value, "-", "_"), "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}
	return strings.Join(parts, " ")
}
