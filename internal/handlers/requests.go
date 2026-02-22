package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/services"
)

type RequestHandler struct {
	svc *services.RequestService
}

func NewRequestHandler(svc *services.RequestService) *RequestHandler {
	return &RequestHandler{svc: svc}
}

func (h *RequestHandler) List(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	requests, err := h.svc.List(c.Request.Context(), familyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"requests": requests})
}

func (h *RequestHandler) Create(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	requesterID := c.GetString(string(models.ContextKeyUserID))

	var req struct {
		TargetID    *string `json:"target_id"`
		Title       string  `json:"title"       binding:"required"`
		Description *string `json:"description"`
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
	rid, err := uuid.Parse(requesterID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var targetID *uuid.UUID
	if req.TargetID != nil {
		tid, err := uuid.Parse(*req.TargetID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target_id"})
			return
		}
		targetID = &tid
	}

	request, err := h.svc.Create(c.Request.Context(), services.CreateRequestInput{
		FamilyID:    fid,
		RequesterID: rid,
		TargetID:    targetID,
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		if err == services.ErrInvalidRequestData {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid request data"})
			return
		}
		if err == services.ErrInvalidRequestTarget {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, request)
}

func (h *RequestHandler) Get(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	id := c.Param("id")

	request, err := h.svc.GetByID(c.Request.Context(), id, familyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, request)
}

func (h *RequestHandler) Respond(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	responderID := c.GetString(string(models.ContextKeyUserID))
	id := c.Param("id")

	var req struct {
		Status  string  `json:"status"  binding:"required"`
		Message *string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	respondedBy, err := uuid.Parse(responderID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	request, err := h.svc.Respond(c.Request.Context(), services.RespondInput{
		ID:          id,
		FamilyID:    familyID,
		Status:      req.Status,
		Message:     req.Message,
		RespondedBy: respondedBy,
	})
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unable to respond to request"})
		return
	}
	c.JSON(http.StatusOK, request)
}
