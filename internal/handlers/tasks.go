package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/repository"
	"github.com/rawdah/rawdah-api/internal/services"
)

type TaskHandler struct {
	svc *services.TaskService
}

func NewTaskHandler(svc *services.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func (h *TaskHandler) List(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))
	role := c.GetString(string(models.ContextKeyRole))

	filter := repository.TaskFilter{
		Status: c.Query("status"),
	}

	// Children only see their own tasks
	if role == "child" {
		filter.AssignedTo = userID
	} else if assignedTo := c.Query("assigned_to"); assignedTo != "" {
		filter.AssignedTo = assignedTo
	}

	tasks, err := h.svc.ListTasks(c.Request.Context(), familyID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func (h *TaskHandler) Create(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	creatorID := c.GetString(string(models.ContextKeyUserID))

	var req struct {
		Title       string     `json:"title"       binding:"required"`
		Description *string    `json:"description"`
		AssignedTo  string     `json:"assigned_to" binding:"required"`
		RewardID    *string    `json:"reward_id"`
		DueDate     *time.Time `json:"due_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	fid, err := uuid.Parse(familyID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	cid, err := uuid.Parse(creatorID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	aid, err := uuid.Parse(req.AssignedTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assigned_to"})
		return
	}

	var rewardID *uuid.UUID
	if req.RewardID != nil {
		rid, err := uuid.Parse(*req.RewardID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reward_id"})
			return
		}
		rewardID = &rid
	}

	task, err := h.svc.CreateTask(c.Request.Context(), services.CreateTaskInput{
		FamilyID:    fid,
		Title:       req.Title,
		Description: req.Description,
		AssignedTo:  aid,
		CreatedBy:   cid,
		RewardID:    rewardID,
		DueDate:     req.DueDate,
	})
	if err != nil {
		if err == services.ErrInvalidAssignee {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if err == services.ErrInvalidTaskData {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid task data"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) Get(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	role := c.GetString(string(models.ContextKeyRole))
	userID := c.GetString(string(models.ContextKeyUserID))
	id := c.Param("id")

	task, err := h.svc.GetTask(c.Request.Context(), id, familyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if role == "child" && task.AssignedTo.String() != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) Update(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	id := c.Param("id")

	var req struct {
		Title       string     `json:"title"       binding:"required"`
		Description *string    `json:"description"`
		RewardID    *string    `json:"reward_id"`
		DueDate     *time.Time `json:"due_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	var rewardID *uuid.UUID
	if req.RewardID != nil {
		rid, err := uuid.Parse(*req.RewardID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reward_id"})
			return
		}
		rewardID = &rid
	}

	task, err := h.svc.UpdateTask(c.Request.Context(), id, familyID, req.Title, req.Description, rewardID, req.DueDate)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) Delete(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	id := c.Param("id")

	if err := h.svc.DeleteTask(c.Request.Context(), id, familyID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *TaskHandler) Start(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))
	id := c.Param("id")

	task, err := h.svc.StartTask(c.Request.Context(), id, familyID, userID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unable to start task"})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) Complete(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))
	id := c.Param("id")

	task, err := h.svc.CompleteTask(c.Request.Context(), id, familyID, userID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unable to complete task"})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) RequestReward(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))
	id := c.Param("id")

	task, err := h.svc.RequestReward(c.Request.Context(), id, familyID, userID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unable to request reward"})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) ApproveReward(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	id := c.Param("id")

	task, err := h.svc.ApproveReward(c.Request.Context(), id, familyID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unable to approve reward"})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) DeclineReward(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	id := c.Param("id")

	task, err := h.svc.DeclineReward(c.Request.Context(), id, familyID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unable to decline reward"})
		return
	}
	c.JSON(http.StatusOK, task)
}
