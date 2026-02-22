package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/services"
)

type MessageHandler struct {
	svc *services.MessageService
}

func NewMessageHandler(svc *services.MessageService) *MessageHandler {
	return &MessageHandler{svc: svc}
}

func (h *MessageHandler) Conversations(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))

	convs, err := h.svc.Conversations(c.Request.Context(), userID, familyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"conversations": convs})
}

func (h *MessageHandler) Thread(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))
	otherUserID := c.Param("user_id")

	messages, err := h.svc.GetThread(c.Request.Context(), userID, otherUserID, familyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *MessageHandler) Send(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	senderID := c.GetString(string(models.ContextKeyUserID))

	var req struct {
		RecipientID string `json:"recipient_id" binding:"required"`
		Content     string `json:"content"      binding:"required"`
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
	sid, err := uuid.Parse(senderID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rid, err := uuid.Parse(req.RecipientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid recipient_id"})
		return
	}

	msg, err := h.svc.Send(c.Request.Context(), services.SendMessageInput{
		FamilyID:    fid,
		SenderID:    sid,
		RecipientID: rid,
		Content:     req.Content,
	})
	if err != nil {
		if err == services.ErrInvalidRecipient {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if err == services.ErrInvalidMessage {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid message"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, msg)
}

func (h *MessageHandler) MarkRead(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))
	id := c.Param("id")

	if err := h.svc.MarkRead(c.Request.Context(), id, userID, familyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
}
