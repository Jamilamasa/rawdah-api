package handlers

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/repository"
)

type PushHandler struct {
	repo *repository.PushRepo
}

func NewPushHandler(repo *repository.PushRepo) *PushHandler {
	return &PushHandler{repo: repo}
}

func (h *PushHandler) Subscribe(c *gin.Context) {
	userID := c.GetString(string(models.ContextKeyUserID))

	var req struct {
		Endpoint string `json:"endpoint" binding:"required"`
		P256dh   string `json:"p256dh"   binding:"required"`
		Auth     string `json:"auth"     binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The request body is invalid. Please verify required fields and value formats."})
		return
	}
	if !isValidPushEndpoint(req.Endpoint) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The push endpoint URL is invalid."})
		return
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication is required or your session is invalid."})
		return
	}
	sub := &models.PushSubscription{
		UserID:   uid,
		Endpoint: req.Endpoint,
		P256dh:   req.P256dh,
		Auth:     req.Auth,
	}

	if err := h.repo.Subscribe(c.Request.Context(), sub); err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "subscribed"})
}

func (h *PushHandler) Unsubscribe(c *gin.Context) {
	userID := c.GetString(string(models.ContextKeyUserID))

	var req struct {
		Endpoint string `json:"endpoint" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The request body is invalid. Please verify required fields and value formats."})
		return
	}
	if !isValidPushEndpoint(req.Endpoint) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The push endpoint URL is invalid."})
		return
	}

	if err := h.repo.Unsubscribe(c.Request.Context(), req.Endpoint, userID); err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "unsubscribed"})
}

func isValidPushEndpoint(endpoint string) bool {
	u, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil {
		return false
	}
	return u.Scheme == "https" && u.Host != ""
}
