package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/repository"
	"github.com/rawdah/rawdah-api/internal/storage"
)

type UploadHandler struct {
	r2         *storage.R2Client
	userRepo   *repository.UserRepo
	familyRepo *repository.FamilyRepo
}

const maxUploadBytes = 5 * 1024 * 1024

func NewUploadHandler(r2 *storage.R2Client, userRepo *repository.UserRepo, familyRepo *repository.FamilyRepo) *UploadHandler {
	return &UploadHandler{r2: r2, userRepo: userRepo, familyRepo: familyRepo}
}

func (h *UploadHandler) Avatar(c *gin.Context) {
	userID := c.GetString(string(models.ContextKeyUserID))
	familyID := c.GetString(string(models.ContextKeyFamilyID))

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar file required"})
		return
	}
	defer file.Close()

	// Max 5MB
	if header.Size > maxUploadBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large, max 5MB"})
		return
	}

	fileBytes, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}
	if len(fileBytes) > maxUploadBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large, max 5MB"})
		return
	}

	contentType := http.DetectContentType(fileBytes)
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only JPEG, PNG, and WebP images are allowed"})
		return
	}

	url, err := h.r2.UploadAvatar(c.Request.Context(), familyID, userID, fileBytes, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		return
	}

	// Update user's avatar URL
	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	_ = h.userRepo.UpdateAvatar(c.Request.Context(), uid, familyID, url)

	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *UploadHandler) Logo(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))

	file, header, err := c.Request.FormFile("logo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "logo file required"})
		return
	}
	defer file.Close()

	// Max 5MB
	if header.Size > maxUploadBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large, max 5MB"})
		return
	}

	fileBytes, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}
	if len(fileBytes) > maxUploadBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large, max 5MB"})
		return
	}

	contentType := http.DetectContentType(fileBytes)
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only JPEG, PNG, and WebP images are allowed"})
		return
	}

	url, err := h.r2.UploadLogo(c.Request.Context(), familyID, fileBytes, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		return
	}

	// Update family logo URL in DB
	_ = h.familyRepo.UpdateLogoURL(c.Request.Context(), familyID, url)

	c.JSON(http.StatusOK, gin.H{"url": url})
}
