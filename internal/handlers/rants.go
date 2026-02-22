package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/services"
)

type RantHandler struct {
	svc *services.RantService
}

func NewRantHandler(svc *services.RantService) *RantHandler {
	return &RantHandler{svc: svc}
}

func (h *RantHandler) List(c *gin.Context) {
	role := c.GetString(string(models.ContextKeyRole))
	if role != "child" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to perform this action."})
		return
	}
	userID := c.GetString(string(models.ContextKeyUserID))
	rants, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"rants": rants})
}

func (h *RantHandler) Create(c *gin.Context) {
	role := c.GetString(string(models.ContextKeyRole))
	if role != "child" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to perform this action."})
		return
	}
	userID := c.GetString(string(models.ContextKeyUserID))

	var req struct {
		Title    *string `json:"title"`
		Content  string  `json:"content" binding:"required"`
		Password *string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rant, err := h.svc.Create(c.Request.Context(), services.CreateRantInput{
		UserID:   userID,
		Title:    req.Title,
		Content:  req.Content,
		Password: req.Password,
	})
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusCreated, rant)
}

func (h *RantHandler) Get(c *gin.Context) {
	role := c.GetString(string(models.ContextKeyRole))
	if role != "child" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to perform this action."})
		return
	}
	userID := c.GetString(string(models.ContextKeyUserID))
	id := c.Param("id")
	password := c.GetHeader("X-Rant-Password")

	rant, err := h.svc.Get(c.Request.Context(), id, userID, password)
	if err == services.ErrRantWrongPassword {
		c.JSON(http.StatusForbidden, gin.H{"error": "The provided rant password is incorrect."})
		return
	}
	if err == services.ErrRantLocked {
		// No password supplied — return metadata only
		c.JSON(http.StatusOK, gin.H{
			"id":         rant.ID,
			"title":      rant.Title,
			"is_locked":  true,
			"created_at": rant.CreatedAt,
			"message":    "This rant is locked. Provide X-Rant-Password header.",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found."})
		return
	}
	c.JSON(http.StatusOK, rant)
}

func (h *RantHandler) Update(c *gin.Context) {
	role := c.GetString(string(models.ContextKeyRole))
	if role != "child" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to perform this action."})
		return
	}
	userID := c.GetString(string(models.ContextKeyUserID))
	id := c.Param("id")

	var req struct {
		Title    *string `json:"title"`
		Content  string  `json:"content" binding:"required"`
		Password *string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rant, err := h.svc.Update(c.Request.Context(), id, userID, services.UpdateRantInput{
		Title:    req.Title,
		Content:  req.Content,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found."})
		return
	}
	c.JSON(http.StatusOK, rant)
}

func (h *RantHandler) Delete(c *gin.Context) {
	role := c.GetString(string(models.ContextKeyRole))
	if role != "child" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to perform this action."})
		return
	}
	userID := c.GetString(string(models.ContextKeyUserID))
	id := c.Param("id")

	if err := h.svc.Delete(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
