package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/services"
)

type DashboardHandler struct {
	svc *services.DashboardService
}

func NewDashboardHandler(svc *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func parseDays(c *gin.Context, defaultDays int) int {
	daysStr := c.Query("days")
	if daysStr == "" {
		return defaultDays
	}
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		return defaultDays
	}
	return days
}

func (h *DashboardHandler) Summary(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	summary, err := h.svc.Summary(c.Request.Context(), familyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *DashboardHandler) TaskCompletion(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	days := parseDays(c, 30)

	data, err := h.svc.TaskCompletion(c.Request.Context(), familyID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "days": days})
}

func (h *DashboardHandler) GameTime(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	days := parseDays(c, 30)

	data, err := h.svc.GameTime(c.Request.Context(), familyID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "days": days})
}

func (h *DashboardHandler) QuizScores(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	days := parseDays(c, 30)

	data, err := h.svc.QuizScores(c.Request.Context(), familyID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "days": days})
}

func (h *DashboardHandler) LearnProgress(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	data, err := h.svc.LearnProgress(c.Request.Context(), familyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}
