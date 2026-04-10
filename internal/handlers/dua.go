package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/services"
)

type DuaHandler struct {
	svc *services.DuaService
}

func NewDuaHandler(svc *services.DuaService) *DuaHandler {
	return &DuaHandler{svc: svc}
}

func (h *DuaHandler) Generate(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))

	var req struct {
		AskingFor    string `json:"asking_for"     binding:"required"`
		HeavyOnHeart string `json:"heavy_on_heart" binding:"required"`
		AfraidOf     string `json:"afraid_of"      binding:"required"`
		IfAnswered   string `json:"if_answered"    binding:"required"`
		OutputStyle  string `json:"output_style"   binding:"required"`
		Depth        string `json:"depth"          binding:"required"`
		Tone         string `json:"tone"           binding:"required"`
	}
	if !bindJSONWithValidation(c, &req) {
		return
	}

	result, err := h.svc.Generate(c.Request.Context(), services.GenerateDuaInput{
		FamilyID:     familyID,
		UserID:       userID,
		AskingFor:    req.AskingFor,
		HeavyOnHeart: req.HeavyOnHeart,
		AfraidOf:     req.AfraidOf,
		IfAnswered:   req.IfAnswered,
		OutputStyle:  req.OutputStyle,
		Depth:        req.Depth,
		Tone:         req.Tone,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidDuaInput):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrDuaProviderUnavailable):
			_ = c.Error(err)
			c.JSON(http.StatusBadGateway, gin.H{
				"error": "Dua service is temporarily unavailable. Please try again shortly.",
			})
		default:
			respondInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *DuaHandler) ListHistory(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))

	limit := 20
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be a positive integer."})
			return
		}
		limit = parsedLimit
	}

	history, err := h.svc.ListHistory(c.Request.Context(), services.ListDuaHistoryInput{
		FamilyID: familyID,
		UserID:   userID,
		Limit:    limit,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidDuaInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			respondInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"history": history})
}

func (h *DuaHandler) GetHistory(c *gin.Context) {
	id := c.Param("id")
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))

	entry, err := h.svc.GetHistory(c.Request.Context(), id, familyID, userID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidDuaInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrDuaHistoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "The requested dua history entry was not found."})
		default:
			respondInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, entry)
}
