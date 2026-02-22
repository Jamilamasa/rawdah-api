package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/services"
)

type NotificationHandler struct {
	svc *services.NotificationService
}

func NewNotificationHandler(svc *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID := c.GetString(string(models.ContextKeyUserID))
	notifs, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"notifications": notifs})
}

func (h *NotificationHandler) ReadAll(c *gin.Context) {
	userID := c.GetString(string(models.ContextKeyUserID))
	if err := h.svc.ReadAll(c.Request.Context(), userID); err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "all notifications marked as read"})
}

func (h *NotificationHandler) ReadOne(c *gin.Context) {
	userID := c.GetString(string(models.ContextKeyUserID))
	id := c.Param("id")

	if err := h.svc.ReadOne(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
}
