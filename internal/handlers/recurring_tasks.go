package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/config"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/services"
)

type RecurringTaskHandler struct {
	svc *services.TaskService
	cfg *config.Config
}

func NewRecurringTaskHandler(svc *services.TaskService, cfg *config.Config) *RecurringTaskHandler {
	return &RecurringTaskHandler{svc: svc, cfg: cfg}
}

func (h *RecurringTaskHandler) List(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	tasks, err := h.svc.ListRecurringTasks(c.Request.Context(), familyID)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"recurring_tasks": tasks})
}

func (h *RecurringTaskHandler) Create(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	creatorID := c.GetString(string(models.ContextKeyUserID))

	var req struct {
		Title       string  `json:"title"       binding:"required"`
		Description *string `json:"description"`
		AssignedTo  string  `json:"assigned_to" binding:"required"`
		RewardID    *string `json:"reward_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The request body is invalid. Please verify required fields and value formats."})
		return
	}

	fid, err := uuid.Parse(familyID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication is required or your session is invalid."})
		return
	}
	cid, err := uuid.Parse(creatorID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication is required or your session is invalid."})
		return
	}
	aid, err := uuid.Parse(req.AssignedTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Assigned user ID format is invalid."})
		return
	}

	var rewardID *uuid.UUID
	if req.RewardID != nil {
		rid, err := uuid.Parse(*req.RewardID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Reward ID format is invalid."})
			return
		}
		rewardID = &rid
	}

	task, err := h.svc.CreateRecurringTask(c.Request.Context(), services.CreateRecurringTaskInput{
		FamilyID:    fid,
		Title:       req.Title,
		Description: req.Description,
		AssignedTo:  aid,
		CreatedBy:   cid,
		RewardID:    rewardID,
	})
	if err != nil {
		if err == services.ErrInvalidAssignee {
			c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found."})
			return
		}
		if err == services.ErrInvalidTaskData {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Task data is invalid for this operation."})
			return
		}
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusCreated, task)
}

func (h *RecurringTaskHandler) Delete(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	id := c.Param("id")

	if err := h.svc.DeleteRecurringTask(c.Request.Context(), id, familyID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *RecurringTaskHandler) TriggerWeekend(c *gin.Context) {
	secret := c.GetHeader("X-Cron-Secret")
	if h.cfg.CronSecret == "" || secret != h.cfg.CronSecret {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing cron secret."})
		return
	}

	triggered, errs, err := h.svc.TriggerWeekendTasks(c.Request.Context())
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"triggered": triggered, "errors": errs})
}
