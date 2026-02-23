package services

import (
	"context"
	"errors"
	"fmt"
	"html"
	"strings"

	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/mailer"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/repository"
	"github.com/rawdah/rawdah-api/internal/ws"
)

type RequestService struct {
	requestRepo *repository.RequestRepo
	familyRepo  *repository.FamilyRepo
	mailer      *mailer.Mailer
	hub         *ws.Hub
}

var (
	ErrInvalidRequestData   = errors.New("invalid request data")
	ErrInvalidRequestTarget = errors.New("invalid request target")
)

func NewRequestService(requestRepo *repository.RequestRepo, familyRepo *repository.FamilyRepo, m *mailer.Mailer, hub *ws.Hub) *RequestService {
	return &RequestService{
		requestRepo: requestRepo,
		familyRepo:  familyRepo,
		mailer:      m,
		hub:         hub,
	}
}

type CreateRequestInput struct {
	FamilyID    uuid.UUID
	RequesterID uuid.UUID
	TargetID    *uuid.UUID
	Title       string
	Description *string
}

func (s *RequestService) Create(ctx context.Context, input CreateRequestInput) (*models.Request, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" || len(title) > 160 {
		return nil, ErrInvalidRequestData
	}
	if input.TargetID != nil {
		target, err := s.familyRepo.GetMemberByID(ctx, input.TargetID.String(), input.FamilyID.String())
		if err != nil || !target.IsActive {
			return nil, ErrInvalidRequestTarget
		}
	}

	req := &models.Request{
		FamilyID:    input.FamilyID,
		RequesterID: input.RequesterID,
		TargetID:    input.TargetID,
		Title:       title,
		Description: input.Description,
		Status:      "pending",
	}
	if err := s.requestRepo.Create(ctx, req); err != nil {
		return nil, err
	}

	// Notify family members in realtime.
	s.hub.Broadcast(ws.BroadcastMsg{
		Room:  fmt.Sprintf("family:%s", input.FamilyID.String()),
		Event: ws.WSEvent{Type: "request.new", Payload: req},
	})

	// Async email notification to adults who can act on the request.
	go s.sendCreateEmails(req)

	return req, nil
}

func (s *RequestService) List(ctx context.Context, familyID string) ([]*models.Request, error) {
	return s.requestRepo.List(ctx, familyID)
}

func (s *RequestService) GetByID(ctx context.Context, id, familyID string) (*models.Request, error) {
	return s.requestRepo.GetByID(ctx, id, familyID)
}

type RespondInput struct {
	ID          string
	FamilyID    string
	Status      string
	Message     *string
	RespondedBy uuid.UUID
}

func (s *RequestService) Respond(ctx context.Context, input RespondInput) (*models.Request, error) {
	if input.Status != "approved" && input.Status != "declined" {
		return nil, fmt.Errorf("invalid status: must be approved or declined")
	}

	req, err := s.requestRepo.Respond(ctx, input.ID, input.FamilyID, input.Status, input.Message, input.RespondedBy)
	if err != nil {
		return nil, err
	}

	// Notify requester
	s.hub.Broadcast(ws.BroadcastMsg{
		Room:  fmt.Sprintf("user:%s", req.RequesterID.String()),
		Event: ws.WSEvent{Type: "request.responded", Payload: req},
	})

	// Async email
	go func() {
		member, err := s.familyRepo.GetMemberByID(context.Background(), req.RequesterID.String(), input.FamilyID)
		if err != nil || member.Email == nil {
			return
		}
		statusWord := "approved"
		if input.Status == "declined" {
			statusWord = "declined"
		}
		subject := fmt.Sprintf("Your request has been %s", statusWord)
		body := fmt.Sprintf(`<p>Your request "<strong>%s</strong>" has been <strong>%s</strong>.</p>`, req.Title, statusWord)
		if input.Message != nil && *input.Message != "" {
			body += fmt.Sprintf(`<p>Message: %s</p>`, *input.Message)
		}
		html := mailer.BuildEmail("Request Update", body, "View Requests", "https://kids.rawdah.app/requests", "")
		_ = s.mailer.Send(mailer.BrevoContact{Name: member.Name, Email: *member.Email}, subject, html)
	}()

	return req, nil
}

func (s *RequestService) sendCreateEmails(req *models.Request) {
	if req == nil || s.mailer == nil {
		return
	}

	familyID := req.FamilyID.String()

	requesterName := "A child"
	if requester, err := s.familyRepo.GetMemberByID(context.Background(), req.RequesterID.String(), familyID); err == nil {
		if trimmed := strings.TrimSpace(requester.Name); trimmed != "" {
			requesterName = trimmed
		}
	}

	recipients := make([]*models.User, 0, 4)
	if req.TargetID != nil {
		target, err := s.familyRepo.GetMemberByID(context.Background(), req.TargetID.String(), familyID)
		if err == nil {
			recipients = append(recipients, target)
		}
	} else {
		members, err := s.familyRepo.ListMembers(context.Background(), familyID)
		if err == nil {
			recipients = append(recipients, members...)
		}
	}

	if len(recipients) == 0 {
		return
	}

	safeTitle := html.EscapeString(req.Title)
	safeRequesterName := html.EscapeString(requesterName)
	subject := fmt.Sprintf("New request from %s", requesterName)
	baseBody := fmt.Sprintf(`<p><strong>%s</strong> sent a new request: "<strong>%s</strong>".</p>`, safeRequesterName, safeTitle)
	if req.Description != nil {
		if description := strings.TrimSpace(*req.Description); description != "" {
			baseBody += fmt.Sprintf("<p>Details: %s</p>", html.EscapeString(description))
		}
	}

	seen := make(map[string]struct{}, len(recipients))
	for _, member := range recipients {
		if member == nil || !member.IsActive || member.Role == "child" || member.Email == nil {
			continue
		}
		email := strings.TrimSpace(*member.Email)
		if email == "" {
			continue
		}
		if _, exists := seen[email]; exists {
			continue
		}
		seen[email] = struct{}{}

		htmlBody := mailer.BuildEmail("New Request", baseBody, "Review Request", "https://app.rawdah.app/requests", "")
		_ = s.mailer.Send(mailer.BrevoContact{Name: member.Name, Email: email}, subject, htmlBody)
	}
}
