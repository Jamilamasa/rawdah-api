package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/services"
)

type AssistantHandler struct {
	svc *services.AssistantService
}

func NewAssistantHandler(svc *services.AssistantService) *AssistantHandler {
	return &AssistantHandler{svc: svc}
}

func (h *AssistantHandler) Ask(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))
	role := c.GetString(string(models.ContextKeyRole))

	var req struct {
		Question string `json:"question" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The request body is invalid. Please verify required fields and value formats."})
		return
	}

	answer, err := h.svc.Ask(c.Request.Context(), services.AskAssistantInput{
		FamilyID: familyID,
		UserID:   userID,
		Role:     role,
		Question: req.Question,
	})
	if err != nil {
		if err == services.ErrInvalidAssistantQuestion {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Question is invalid. Please provide a clear question between 2 and 2000 characters."})
			return
		}
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"answer": answer})
}
