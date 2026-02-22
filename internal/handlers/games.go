package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/services"
)

type GameHandler struct {
	svc *services.GameService
}

func NewGameHandler(svc *services.GameService) *GameHandler {
	return &GameHandler{svc: svc}
}

func (h *GameHandler) ListAvailable(c *gin.Context) {
	games := h.svc.ListAvailable()
	c.JSON(http.StatusOK, gin.H{"games": games})
}

func (h *GameHandler) StartSession(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))

	var req struct {
		GameName string `json:"game_name" binding:"required"`
		GameType string `json:"game_type" binding:"required"`
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
	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	session, err := h.svc.StartSession(c.Request.Context(), services.StartSessionInput{
		UserID:   uid,
		FamilyID: fid,
		GameName: req.GameName,
		GameType: req.GameType,
	})
	if err != nil {
		if err == services.ErrGameLimitExceeded {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "daily game limit exceeded"})
			return
		}
		if err == services.ErrInvalidGame {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, session)
}

func (h *GameHandler) EndSession(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))
	id := c.Param("id")

	session, err := h.svc.EndSession(c.Request.Context(), id, userID, familyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *GameHandler) ListSessions(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))
	role := c.GetString(string(models.ContextKeyRole))

	filterUserID := ""
	if role == "child" {
		filterUserID = userID
	} else if uid := c.Query("user_id"); uid != "" {
		filterUserID = uid
	}

	sessions, err := h.svc.ListSessions(c.Request.Context(), familyID, filterUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}
