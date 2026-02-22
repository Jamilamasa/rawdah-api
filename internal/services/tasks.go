package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/mailer"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/repository"
	"github.com/rawdah/rawdah-api/internal/ws"
)

type TaskService struct {
	taskRepo   *repository.TaskRepo
	familyRepo *repository.FamilyRepo
	notifRepo  *repository.NotificationRepo
	xpSvc      *XPService
	mailer     *mailer.Mailer
	hub        *ws.Hub
}

var (
	ErrInvalidAssignee = errors.New("invalid assignee")
	ErrInvalidTaskData = errors.New("invalid task data")
)

func NewTaskService(taskRepo *repository.TaskRepo, familyRepo *repository.FamilyRepo, notifRepo *repository.NotificationRepo, xpSvc *XPService, m *mailer.Mailer, hub *ws.Hub) *TaskService {
	return &TaskService{
		taskRepo:   taskRepo,
		familyRepo: familyRepo,
		notifRepo:  notifRepo,
		xpSvc:      xpSvc,
		mailer:     m,
		hub:        hub,
	}
}

type CreateTaskInput struct {
	FamilyID    uuid.UUID
	Title       string
	Description *string
	AssignedTo  uuid.UUID
	CreatedBy   uuid.UUID
	RewardID    *uuid.UUID
	DueDate     *time.Time
}

func (s *TaskService) ListTasks(ctx context.Context, familyID string, filter repository.TaskFilter) ([]*models.Task, error) {
	return s.taskRepo.List(ctx, familyID, filter)
}

func (s *TaskService) CreateTask(ctx context.Context, input CreateTaskInput) (*models.Task, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" || len(title) > 160 {
		return nil, ErrInvalidTaskData
	}

	assignee, err := s.familyRepo.GetMemberByID(ctx, input.AssignedTo.String(), input.FamilyID.String())
	if err != nil || !assignee.IsActive || assignee.Role != "child" {
		return nil, ErrInvalidAssignee
	}

	task := &models.Task{
		FamilyID:    input.FamilyID,
		Title:       title,
		Description: input.Description,
		AssignedTo:  input.AssignedTo,
		CreatedBy:   input.CreatedBy,
		RewardID:    input.RewardID,
		Status:      "pending",
		DueDate:     input.DueDate,
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	// Emit WS event
	s.hub.Broadcast(ws.BroadcastMsg{
		Room: fmt.Sprintf("user:%s", input.AssignedTo.String()),
		Event: ws.WSEvent{
			Type:    "task.assigned",
			Payload: task,
		},
	})

	// Create notification
	notif := &models.Notification{
		UserID: input.AssignedTo,
		Type:   "task_assigned",
		Title:  "New Task Assigned",
		Body:   fmt.Sprintf("You have been assigned: %s", task.Title),
	}
	_ = s.notifRepo.Create(ctx, notif)

	// Async email notification
	go func() {
		member, err := s.familyRepo.GetMemberByID(context.Background(), input.AssignedTo.String(), input.FamilyID.String())
		if err != nil || member.Email == nil {
			return
		}
		subject := "New Task Assigned"
		body := fmt.Sprintf("<p>You have been assigned a new task: <strong>%s</strong></p>", task.Title)
		if task.DueDate != nil {
			body += fmt.Sprintf("<p>Due: %s</p>", task.DueDate.Format("January 2, 2006"))
		}
		html := mailer.BuildEmail("New Task", body, "View Task", "https://app.rawdah.app/tasks", "")
		_ = s.mailer.Send(mailer.BrevoContact{Name: member.Name, Email: *member.Email}, subject, html)
	}()

	return task, nil
}

func (s *TaskService) GetTask(ctx context.Context, id, familyID string) (*models.Task, error) {
	return s.taskRepo.GetByID(ctx, id, familyID)
}

func (s *TaskService) UpdateTask(ctx context.Context, id, familyID, title string, description *string, rewardID *uuid.UUID, dueDate *time.Time) (*models.Task, error) {
	return s.taskRepo.Update(ctx, id, familyID, title, description, rewardID, dueDate)
}

func (s *TaskService) DeleteTask(ctx context.Context, id, familyID string) error {
	return s.taskRepo.Delete(ctx, id, familyID)
}

func (s *TaskService) StartTask(ctx context.Context, id, familyID, userID string) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, id, familyID)
	if err != nil {
		return nil, err
	}
	if task.AssignedTo.String() != userID {
		return nil, fmt.Errorf("not authorized")
	}
	if task.Status != "pending" {
		return nil, fmt.Errorf("task cannot be started from status: %s", task.Status)
	}
	updated, err := s.taskRepo.UpdateStatus(ctx, id, familyID, "in_progress")
	if err != nil {
		return nil, err
	}

	s.hub.Broadcast(ws.BroadcastMsg{
		Room:  fmt.Sprintf("family:%s", familyID),
		Event: ws.WSEvent{Type: "task.status_updated", Payload: updated},
	})
	return updated, nil
}

func (s *TaskService) CompleteTask(ctx context.Context, id, familyID, userID string) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, id, familyID)
	if err != nil {
		return nil, err
	}
	if task.AssignedTo.String() != userID {
		return nil, fmt.Errorf("not authorized")
	}
	if task.Status != "in_progress" {
		return nil, fmt.Errorf("task must be in_progress to complete")
	}

	updated, err := s.taskRepo.UpdateStatus(ctx, id, familyID, "completed")
	if err != nil {
		return nil, err
	}

	// Award XP
	go func() {
		_ = s.xpSvc.AwardXP(context.Background(), userID, familyID, 20, "task", task.ID)
	}()

	s.hub.Broadcast(ws.BroadcastMsg{
		Room:  fmt.Sprintf("family:%s", familyID),
		Event: ws.WSEvent{Type: "task.status_updated", Payload: updated},
	})
	return updated, nil
}

func (s *TaskService) RequestReward(ctx context.Context, id, familyID, userID string) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, id, familyID)
	if err != nil {
		return nil, err
	}
	if task.AssignedTo.String() != userID {
		return nil, fmt.Errorf("not authorized")
	}
	if task.Status != "completed" {
		return nil, fmt.Errorf("task must be completed before requesting reward")
	}
	if task.RewardID == nil {
		return nil, fmt.Errorf("task has no reward")
	}

	updated, err := s.taskRepo.UpdateStatus(ctx, id, familyID, "reward_requested")
	if err != nil {
		return nil, err
	}

	s.hub.Broadcast(ws.BroadcastMsg{
		Room:  fmt.Sprintf("family:%s", familyID),
		Event: ws.WSEvent{Type: "reward.requested", Payload: updated},
	})
	return updated, nil
}

func (s *TaskService) ApproveReward(ctx context.Context, id, familyID string) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, id, familyID)
	if err != nil {
		return nil, err
	}
	if task.Status != "reward_requested" {
		return nil, fmt.Errorf("task must be in reward_requested status")
	}

	updated, err := s.taskRepo.UpdateStatus(ctx, id, familyID, "reward_approved")
	if err != nil {
		return nil, err
	}

	s.hub.Broadcast(ws.BroadcastMsg{
		Room: fmt.Sprintf("user:%s", task.AssignedTo.String()),
		Event: ws.WSEvent{Type: "reward.responded", Payload: map[string]interface{}{
			"task":     updated,
			"decision": "approved",
		}},
	})
	return updated, nil
}

func (s *TaskService) DeclineReward(ctx context.Context, id, familyID string) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, id, familyID)
	if err != nil {
		return nil, err
	}
	if task.Status != "reward_requested" {
		return nil, fmt.Errorf("task must be in reward_requested status")
	}

	updated, err := s.taskRepo.UpdateStatus(ctx, id, familyID, "reward_declined")
	if err != nil {
		return nil, err
	}

	s.hub.Broadcast(ws.BroadcastMsg{
		Room: fmt.Sprintf("user:%s", task.AssignedTo.String()),
		Event: ws.WSEvent{Type: "reward.responded", Payload: map[string]interface{}{
			"task":     updated,
			"decision": "declined",
		}},
	})
	return updated, nil
}
