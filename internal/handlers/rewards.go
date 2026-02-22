package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/services"
)

type RewardHandler struct {
	svc *services.RewardService
}

func NewRewardHandler(svc *services.RewardService) *RewardHandler {
	return &RewardHandler{svc: svc}
}

func (h *RewardHandler) List(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	rewards, err := h.svc.List(c.Request.Context(), familyID)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"rewards": rewards})
}

func (h *RewardHandler) Create(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))

	var req struct {
		Title       string  `json:"title"       binding:"required"`
		Description *string `json:"description"`
		Value       float64 `json:"value"`
		Type        string  `json:"type"`
		Icon        *string `json:"icon"`
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
	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication is required or your session is invalid."})
		return
	}

	reward, err := h.svc.Create(c.Request.Context(), services.CreateRewardInput{
		FamilyID:    fid,
		Title:       req.Title,
		Description: req.Description,
		Value:       req.Value,
		Type:        req.Type,
		Icon:        req.Icon,
		CreatedBy:   uid,
	})
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusCreated, reward)
}

func (h *RewardHandler) Update(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	id := c.Param("id")

	var req struct {
		Title       string  `json:"title"       binding:"required"`
		Description *string `json:"description"`
		Value       float64 `json:"value"`
		Type        string  `json:"type"`
		Icon        *string `json:"icon"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The request body is invalid. Please verify required fields and value formats."})
		return
	}

	reward, err := h.svc.Update(c.Request.Context(), id, familyID, services.UpdateRewardInput{
		Title:       req.Title,
		Description: req.Description,
		Value:       req.Value,
		Type:        req.Type,
		Icon:        req.Icon,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found."})
		return
	}
	c.JSON(http.StatusOK, reward)
}

func (h *RewardHandler) Delete(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	id := c.Param("id")

	if err := h.svc.Delete(c.Request.Context(), id, familyID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
