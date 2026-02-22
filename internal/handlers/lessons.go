package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/services"
)

type LessonHandler struct {
	svc *services.LessonService
}

func NewLessonHandler(svc *services.LessonService) *LessonHandler {
	return &LessonHandler{svc: svc}
}

// Quran Lessons

func (h *LessonHandler) ListLessons(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	lessons, err := h.svc.ListLessons(c.Request.Context(), familyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"lessons": lessons})
}

func (h *LessonHandler) CreateLesson(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	assignedBy := c.GetString(string(models.ContextKeyUserID))

	var req struct {
		VerseID    string  `json:"verse_id"    binding:"required"`
		AssignedTo string  `json:"assigned_to" binding:"required"`
		RewardID   *string `json:"reward_id"`
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
	vid, err := uuid.Parse(req.VerseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid verse_id"})
		return
	}
	aid, err := uuid.Parse(req.AssignedTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assigned_to"})
		return
	}
	abid, err := uuid.Parse(assignedBy)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
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

	lesson, err := h.svc.AssignLesson(c.Request.Context(), services.AssignLessonInput{
		FamilyID:   fid,
		VerseID:    vid,
		AssignedTo: aid,
		AssignedBy: abid,
		RewardID:   rewardID,
	})
	if err != nil {
		if err == services.ErrInvalidLessonAssignee {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, lesson)
}

func (h *LessonHandler) ListMyLessons(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))

	lessons, err := h.svc.ListMyLessons(c.Request.Context(), userID, familyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"lessons": lessons})
}

func (h *LessonHandler) GetLesson(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	id := c.Param("id")

	lesson, err := h.svc.GetLesson(c.Request.Context(), id, familyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, lesson)
}

func (h *LessonHandler) CompleteLesson(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))
	id := c.Param("id")

	lesson, err := h.svc.CompleteLesson(c.Request.Context(), id, userID, familyID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unable to complete lesson"})
		return
	}
	c.JSON(http.StatusOK, lesson)
}

// Learn Content

func (h *LessonHandler) ListLearnContent(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	content, err := h.svc.GetLearnContent(c.Request.Context(), familyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"content": content})
}

func (h *LessonHandler) CreateLearnContent(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	createdBy := c.GetString(string(models.ContextKeyUserID))

	var req struct {
		AssignedTo  *string `json:"assigned_to"`
		Title       string  `json:"title"        binding:"required"`
		ContentType string  `json:"content_type" binding:"required"`
		Content     string  `json:"content"      binding:"required"`
		RewardID    *string `json:"reward_id"`
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
	cbid, err := uuid.Parse(createdBy)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var assignedTo *uuid.UUID
	if req.AssignedTo != nil {
		aid, err := uuid.Parse(*req.AssignedTo)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assigned_to"})
			return
		}
		assignedTo = &aid
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

	content, err := h.svc.CreateLearnContent(c.Request.Context(), services.CreateLearnContentInput{
		FamilyID:    fid,
		AssignedTo:  assignedTo,
		Title:       req.Title,
		ContentType: req.ContentType,
		Content:     req.Content,
		RewardID:    rewardID,
		CreatedBy:   cbid,
	})
	if err != nil {
		if err == services.ErrInvalidLessonAssignee {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if err == services.ErrInvalidLearnData {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid learn content"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, content)
}

func (h *LessonHandler) ListMyLearnContent(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))

	content, err := h.svc.GetMyLearnContent(c.Request.Context(), userID, familyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"content": content})
}

func (h *LessonHandler) CompleteLearnContent(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))
	id := c.Param("id")

	if err := h.svc.CompleteLearnContent(c.Request.Context(), id, userID, familyID); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unable to complete content"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "content completed"})
}
