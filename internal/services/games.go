package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/repository"
	"github.com/rawdah/rawdah-api/internal/ws"
)

var ErrGameLimitExceeded = errors.New("daily game limit exceeded")
var ErrInvalidGame = errors.New("invalid game")

type AvailableGame struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Icon string `json:"icon"`
	Desc string `json:"description"`
}

var AvailableGames = []AvailableGame{
	{ID: "arabic_letters", Name: "Arabic Letters", Type: "islamic", Icon: "🔤", Desc: "Learn and practice Arabic alphabet"},
	{ID: "quran_word_match", Name: "Quran Word Match", Type: "islamic", Icon: "📖", Desc: "Match Quran words with their meanings"},
	{ID: "prophet_names", Name: "Prophet Names", Type: "islamic", Icon: "⭐", Desc: "Learn the names of prophets"},
	{ID: "islamic_puzzle", Name: "Islamic Puzzle", Type: "islamic", Icon: "🧩", Desc: "Fun Islamic themed puzzles"},
	{ID: "memory_match", Name: "Memory Match", Type: "general", Icon: "🃏", Desc: "Classic memory card matching game"},
	{ID: "math_challenge", Name: "Math Challenge", Type: "general", Icon: "🔢", Desc: "Fun math problems"},
}

type GameService struct {
	gameRepo *repository.GameRepo
	userRepo *repository.UserRepo
	xpSvc    *XPService
	hub      *ws.Hub
}

func NewGameService(gameRepo *repository.GameRepo, userRepo *repository.UserRepo, xpSvc *XPService, hub *ws.Hub) *GameService {
	return &GameService{
		gameRepo: gameRepo,
		userRepo: userRepo,
		xpSvc:    xpSvc,
		hub:      hub,
	}
}

func (s *GameService) ListAvailable() []AvailableGame {
	return AvailableGames
}

type StartSessionInput struct {
	UserID   uuid.UUID
	FamilyID uuid.UUID
	GameName string
	GameType string
}

func (s *GameService) StartSession(ctx context.Context, input StartSessionInput) (*models.GameSession, error) {
	if !isValidGame(input.GameName, input.GameType) {
		return nil, ErrInvalidGame
	}

	// Check daily limit
	user, err := s.userRepo.GetByID(ctx, input.UserID.String(), input.FamilyID.String())
	if err != nil {
		return nil, err
	}

	today := time.Now()
	totalSeconds, err := s.gameRepo.TotalDurationToday(ctx, input.UserID.String(), input.FamilyID.String(), today)
	if err != nil {
		return nil, err
	}

	limitSeconds := user.GameLimitMinutes * 60
	if totalSeconds >= limitSeconds {
		return nil, ErrGameLimitExceeded
	}

	session := &models.GameSession{
		UserID:    input.UserID,
		FamilyID:  input.FamilyID,
		GameName:  input.GameName,
		GameType:  input.GameType,
		StartedAt: time.Now(),
	}

	if err := s.gameRepo.StartSession(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *GameService) EndSession(ctx context.Context, sessionID, userID, familyID string) (*models.GameSession, error) {
	sessions, err := s.gameRepo.ListSessions(ctx, familyID, userID)
	if err != nil {
		return nil, err
	}

	var session *models.GameSession
	for _, s := range sessions {
		if s.ID.String() == sessionID {
			session = s
			break
		}
	}
	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	if session.EndedAt != nil {
		return session, nil
	}

	durationSeconds := int(time.Since(session.StartedAt).Seconds())
	if err := s.gameRepo.EndSession(ctx, sessionID, userID, familyID, durationSeconds); err != nil {
		return nil, err
	}

	// Award 5 XP
	go func() {
		_ = s.xpSvc.AwardXP(context.Background(), userID, familyID, 5, "game", session.ID)
	}()

	// Check if now over limit
	userObj, err := s.userRepo.GetByID(ctx, userID, familyID)
	if err == nil {
		totalSeconds, _ := s.gameRepo.TotalDurationToday(ctx, userID, familyID, time.Now())
		limitSeconds := userObj.GameLimitMinutes * 60
		if totalSeconds >= limitSeconds {
			s.hub.Broadcast(ws.BroadcastMsg{
				Room: fmt.Sprintf("user:%s", userID),
				Event: ws.WSEvent{Type: "game.limit_reached", Payload: map[string]interface{}{
					"user_id":       userID,
					"limit_minutes": userObj.GameLimitMinutes,
				}},
			})
		}
	}

	now := time.Now()
	session.EndedAt = &now
	session.DurationSeconds = &durationSeconds
	return session, nil
}

func (s *GameService) ListSessions(ctx context.Context, familyID, userID string) ([]*models.GameSession, error) {
	return s.gameRepo.ListSessions(ctx, familyID, userID)
}

func isValidGame(name, gameType string) bool {
	for _, g := range AvailableGames {
		if g.Name == name && g.Type == gameType {
			return true
		}
	}
	return false
}
