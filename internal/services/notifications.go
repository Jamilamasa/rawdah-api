package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/push"
	"github.com/rawdah/rawdah-api/internal/repository"
	"github.com/rawdah/rawdah-api/internal/ws"
)

type NotificationService struct {
	notifRepo  *repository.NotificationRepo
	pushSender *push.PushSender
	hub        *ws.Hub
}

func NewNotificationService(notifRepo *repository.NotificationRepo, pushSender *push.PushSender, hub *ws.Hub) *NotificationService {
	return &NotificationService{
		notifRepo:  notifRepo,
		pushSender: pushSender,
		hub:        hub,
	}
}

type CreateNotificationInput struct {
	UserID uuid.UUID
	Type   string
	Title  string
	Body   string
	Data   *string
}

func (s *NotificationService) CreateNotification(ctx context.Context, input CreateNotificationInput) (*models.Notification, error) {
	notif := &models.Notification{
		UserID: input.UserID,
		Type:   input.Type,
		Title:  input.Title,
		Body:   input.Body,
		Data:   input.Data,
	}

	if err := s.notifRepo.Create(ctx, notif); err != nil {
		return nil, err
	}

	// Emit WS event
	s.hub.Broadcast(ws.BroadcastMsg{
		Room:  fmt.Sprintf("user:%s", input.UserID.String()),
		Event: ws.WSEvent{Type: "notification.new", Payload: notif},
	})

	// Async push notification
	go func() {
		s.pushSender.SendToUser(context.Background(), input.UserID.String(), push.PushPayload{
			Title: input.Title,
			Body:  input.Body,
		})
	}()

	return notif, nil
}

func (s *NotificationService) List(ctx context.Context, userID string) ([]*models.Notification, error) {
	return s.notifRepo.List(ctx, userID)
}

func (s *NotificationService) ReadOne(ctx context.Context, id, userID string) error {
	return s.notifRepo.ReadOne(ctx, id, userID)
}

func (s *NotificationService) ReadAll(ctx context.Context, userID string) error {
	return s.notifRepo.ReadAll(ctx, userID)
}
