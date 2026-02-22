package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/mailer"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/repository"
	"github.com/rawdah/rawdah-api/internal/ws"
)

type MessageService struct {
	msgRepo    *repository.MessageRepo
	familyRepo *repository.FamilyRepo
	xpSvc      *XPService
	mailer     *mailer.Mailer
	hub        *ws.Hub
}

var (
	ErrInvalidRecipient = errors.New("invalid recipient")
	ErrInvalidMessage   = errors.New("invalid message")
)

func NewMessageService(msgRepo *repository.MessageRepo, familyRepo *repository.FamilyRepo, xpSvc *XPService, m *mailer.Mailer, hub *ws.Hub) *MessageService {
	return &MessageService{
		msgRepo:    msgRepo,
		familyRepo: familyRepo,
		xpSvc:      xpSvc,
		mailer:     m,
		hub:        hub,
	}
}

func (s *MessageService) Conversations(ctx context.Context, userID, familyID string) ([]*models.Message, error) {
	return s.msgRepo.Conversations(ctx, userID, familyID)
}

func (s *MessageService) GetThread(ctx context.Context, userID, otherUserID, familyID string) ([]*models.Message, error) {
	return s.msgRepo.GetThread(ctx, userID, otherUserID, familyID)
}

type SendMessageInput struct {
	FamilyID    uuid.UUID
	SenderID    uuid.UUID
	RecipientID uuid.UUID
	Content     string
}

func (s *MessageService) Send(ctx context.Context, input SendMessageInput) (*models.Message, error) {
	if input.SenderID == input.RecipientID {
		return nil, ErrInvalidRecipient
	}

	content := strings.TrimSpace(input.Content)
	if content == "" || len(content) > 2000 {
		return nil, ErrInvalidMessage
	}

	if _, err := s.familyRepo.GetMemberByID(ctx, input.RecipientID.String(), input.FamilyID.String()); err != nil {
		return nil, ErrInvalidRecipient
	}

	msg := &models.Message{
		FamilyID:    input.FamilyID,
		SenderID:    input.SenderID,
		RecipientID: input.RecipientID,
		Content:     content,
	}

	if err := s.msgRepo.Send(ctx, msg); err != nil {
		return nil, err
	}

	s.hub.Broadcast(ws.BroadcastMsg{
		Room:  fmt.Sprintf("user:%s", input.RecipientID.String()),
		Event: ws.WSEvent{Type: "message.new", Payload: msg},
	})

	// Check messenger badge eligibility.
	go func() {
		s.xpSvc.CheckAndAwardBadges(context.Background(), input.SenderID.String(), input.FamilyID.String())
	}()

	// Async email notification
	go func() {
		recipient, err := s.familyRepo.GetMemberByID(context.Background(), input.RecipientID.String(), input.FamilyID.String())
		if err != nil || recipient.Email == nil {
			return
		}
		sender, err := s.familyRepo.GetMemberByID(context.Background(), input.SenderID.String(), input.FamilyID.String())
		if err != nil {
			return
		}
		subject := fmt.Sprintf("New message from %s", sender.Name)
		body := fmt.Sprintf("<p>You have a new message from <strong>%s</strong>:</p><blockquote>%s</blockquote>", sender.Name, input.Content)
		html := mailer.BuildEmail("New Message", body, "View Messages", "https://app.rawdah.app/messages", "")
		_ = s.mailer.Send(mailer.BrevoContact{Name: recipient.Name, Email: *recipient.Email}, subject, html)
	}()

	return msg, nil
}

func (s *MessageService) MarkRead(ctx context.Context, id, recipientID, familyID string) error {
	return s.msgRepo.MarkRead(ctx, id, recipientID, familyID)
}
