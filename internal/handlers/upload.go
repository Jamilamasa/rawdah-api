package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

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

func NewUploadHandler(r2 *storage.R2Client, userRepo *repository.UserRepo, familyRepo *repository.FamilyRepo) *UploadHandler {
	return &UploadHandler{r2: r2, userRepo: userRepo, familyRepo: familyRepo}
}

const maxUploadBytes int64 = 5 * 1024 * 1024

type presignUploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type" binding:"required"`
	Size        int64  `json:"size" binding:"required"`
}

type confirmUploadRequest struct {
	ObjectKey string `json:"object_key" binding:"required"`
}

func (h *UploadHandler) PresignAvatar(c *gin.Context) {
	userID := c.GetString(string(models.ContextKeyUserID))
	familyID := c.GetString(string(models.ContextKeyFamilyID))

	var req presignUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.Size <= 0 || req.Size > maxUploadBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large, max 5MB"})
		return
	}
	if !isAllowedImageContentType(req.ContentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported content type"})
		return
	}

	objectKey, err := h.r2.BuildAvatarObjectKey(familyID, userID, req.ContentType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported content type"})
		return
	}

	uploadURL, err := h.r2.CreatePresignedUpload(c.Request.Context(), objectKey, req.ContentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate upload url"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"upload_url":  uploadURL,
		"object_key":  objectKey,
		"expires_in":  h.r2.PresignTTLSeconds(),
		"max_size_mb": 5,
	})
}

func (h *UploadHandler) ConfirmAvatar(c *gin.Context) {
	userID := c.GetString(string(models.ContextKeyUserID))
	familyID := c.GetString(string(models.ContextKeyFamilyID))

	var req confirmUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	expectedPrefix := fmt.Sprintf("families/%s/avatars/%s/", familyID, userID)
	if !strings.HasPrefix(req.ObjectKey, expectedPrefix) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	meta, err := h.r2.GetObjectMetadata(c.Request.Context(), req.ObjectKey)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify object"})
		return
	}
	if meta.SizeBytes <= 0 || meta.SizeBytes > maxUploadBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large, max 5MB"})
		return
	}
	if !isAllowedImageContentType(meta.ContentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported content type"})
		return
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if err := h.userRepo.UpdateAvatar(c.Request.Context(), uid, familyID, req.ObjectKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile image"})
		return
	}

	downloadURL, err := h.r2.CreatePresignedDownload(c.Request.Context(), req.ObjectKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate download url"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"object_key": req.ObjectKey,
		"url":        downloadURL,
		"expires_in": h.r2.PresignTTLSeconds(),
	})
}

func (h *UploadHandler) PresignLogo(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))

	var req presignUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.Size <= 0 || req.Size > maxUploadBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large, max 5MB"})
		return
	}
	if !isAllowedImageContentType(req.ContentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported content type"})
		return
	}

	objectKey, err := h.r2.BuildLogoObjectKey(familyID, req.ContentType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported content type"})
		return
	}

	uploadURL, err := h.r2.CreatePresignedUpload(c.Request.Context(), objectKey, req.ContentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate upload url"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"upload_url":  uploadURL,
		"object_key":  objectKey,
		"expires_in":  h.r2.PresignTTLSeconds(),
		"max_size_mb": 5,
	})
}

func (h *UploadHandler) ConfirmLogo(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))

	var req confirmUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	expectedPrefix := fmt.Sprintf("families/%s/logos/", familyID)
	if !strings.HasPrefix(req.ObjectKey, expectedPrefix) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	meta, err := h.r2.GetObjectMetadata(c.Request.Context(), req.ObjectKey)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify object"})
		return
	}
	if meta.SizeBytes <= 0 || meta.SizeBytes > maxUploadBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large, max 5MB"})
		return
	}
	if !isAllowedImageContentType(meta.ContentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported content type"})
		return
	}

	if err := h.familyRepo.UpdateLogoURL(c.Request.Context(), familyID, req.ObjectKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update family logo"})
		return
	}

	downloadURL, err := h.r2.CreatePresignedDownload(c.Request.Context(), req.ObjectKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate download url"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"object_key": req.ObjectKey,
		"url":        downloadURL,
		"expires_in": h.r2.PresignTTLSeconds(),
	})
}

func isAllowedImageContentType(contentType string) bool {
	base := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	switch base {
	case "image/jpeg", "image/jpg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}
