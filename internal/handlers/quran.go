package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rawdah/rawdah-api/internal/repository"
)

type QuranHandler struct {
	repo *repository.QuranRepo
}

func NewQuranHandler(repo *repository.QuranRepo) *QuranHandler {
	return &QuranHandler{repo: repo}
}

func (h *QuranHandler) ListVerses(c *gin.Context) {
	topic := c.Query("topic")
	difficulty := c.Query("difficulty")

	verses, err := h.repo.ListVerses(c.Request.Context(), topic, difficulty)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"verses": verses})
}

func (h *QuranHandler) GetVerse(c *gin.Context) {
	id := c.Param("id")
	verse, err := h.repo.GetVerseByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found."})
		return
	}
	c.JSON(http.StatusOK, verse)
}
