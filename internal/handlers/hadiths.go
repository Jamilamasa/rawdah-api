package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/repository"
)

type HadithHandler struct {
	repo *repository.HadithRepo
}

func NewHadithHandler(repo *repository.HadithRepo) *HadithHandler {
	return &HadithHandler{repo: repo}
}

func (h *HadithHandler) List(c *gin.Context) {
	difficulty := c.Query("difficulty")
	hadiths, err := h.repo.List(c.Request.Context(), difficulty)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"hadiths": hadiths})
}

func (h *HadithHandler) Random(c *gin.Context) {
	difficulty := c.Query("difficulty")
	hadith, err := h.repo.Random(c.Request.Context(), difficulty)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no hadiths found"})
		return
	}
	c.JSON(http.StatusOK, hadith)
}

// Learned returns all hadiths the authenticated child has completed a quiz for.
func (h *HadithHandler) Learned(c *gin.Context) {
	userID := c.GetString(string(models.ContextKeyUserID))
	hadiths, err := h.repo.Learned(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"hadiths": hadiths, "count": len(hadiths)})
}

func (h *HadithHandler) Get(c *gin.Context) {
	id := c.Param("id")
	hadith, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, hadith)
}
