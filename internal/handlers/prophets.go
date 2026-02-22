package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rawdah/rawdah-api/internal/repository"
)

type ProphetHandler struct {
	repo *repository.ProphetRepo
}

func NewProphetHandler(repo *repository.ProphetRepo) *ProphetHandler {
	return &ProphetHandler{repo: repo}
}

func (h *ProphetHandler) List(c *gin.Context) {
	prophets, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"prophets": prophets})
}

func (h *ProphetHandler) Get(c *gin.Context) {
	id := c.Param("id")
	prophet, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, prophet)
}
