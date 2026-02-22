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
		c.JSON(http.StatusBadRequest, gin.H{"error": "The request body is invalid. Please verify required fields and value formats."})
		return
	}
	if req.Size <= 0 || req.Size > maxUploadBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is too large. Maximum allowed size is 5 MB."})
		return
	}
	if !isAllowedImageContentType(req.ContentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported content type. Allowed types are image/jpeg, image/png, image/webp, and image/gif."})
		return
	}

	objectKey, err := h.r2.BuildAvatarObjectKey(familyID, userID, req.ContentType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported content type. Allowed types are image/jpeg, image/png, image/webp, and image/gif."})
		return
	}

	uploadURL, err := h.r2.CreatePresignedUpload(c.Request.Context(), objectKey, req.ContentType)
	if err != nil {
		respondInternalErrorWithMessage(c, "Unable to generate avatar upload URL at the moment.", err)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "The request body is invalid. Please verify required fields and value formats."})
		return
	}

	expectedPrefix := fmt.Sprintf("families/%s/avatars/%s/", familyID, userID)
	if !strings.HasPrefix(req.ObjectKey, expectedPrefix) {
		c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found."})
		return
	}

	meta, err := h.r2.GetObjectMetadata(c.Request.Context(), req.ObjectKey)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found."})
			return
		}
		respondInternalErrorWithMessage(c, "Unable to verify uploaded avatar at the moment.", err)
		return
	}
	if meta.SizeBytes <= 0 || meta.SizeBytes > maxUploadBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is too large. Maximum allowed size is 5 MB."})
		return
	}
	if !isAllowedImageContentType(meta.ContentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported content type. Allowed types are image/jpeg, image/png, image/webp, and image/gif."})
		return
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication is required or your session is invalid."})
		return
	}
	if err := h.userRepo.UpdateAvatar(c.Request.Context(), uid, familyID, req.ObjectKey); err != nil {
		respondInternalErrorWithMessage(c, "Unable to update profile image at the moment.", err)
		return
	}

	downloadURL, err := h.r2.CreatePresignedDownload(c.Request.Context(), req.ObjectKey)
	if err != nil {
		respondInternalErrorWithMessage(c, "Unable to generate avatar download URL at the moment.", err)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "The request body is invalid. Please verify required fields and value formats."})
		return
	}
	if req.Size <= 0 || req.Size > maxUploadBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is too large. Maximum allowed size is 5 MB."})
		return
	}
	if !isAllowedImageContentType(req.ContentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported content type. Allowed types are image/jpeg, image/png, image/webp, and image/gif."})
		return
	}

	objectKey, err := h.r2.BuildLogoObjectKey(familyID, req.ContentType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported content type. Allowed types are image/jpeg, image/png, image/webp, and image/gif."})
		return
	}

	uploadURL, err := h.r2.CreatePresignedUpload(c.Request.Context(), objectKey, req.ContentType)
	if err != nil {
		respondInternalErrorWithMessage(c, "Unable to generate family logo upload URL at the moment.", err)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "The request body is invalid. Please verify required fields and value formats."})
		return
	}

	expectedPrefix := fmt.Sprintf("families/%s/logos/", familyID)
	if !strings.HasPrefix(req.ObjectKey, expectedPrefix) {
		c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found."})
		return
	}

	meta, err := h.r2.GetObjectMetadata(c.Request.Context(), req.ObjectKey)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found."})
			return
		}
		respondInternalErrorWithMessage(c, "Unable to verify uploaded family logo at the moment.", err)
		return
	}
	if meta.SizeBytes <= 0 || meta.SizeBytes > maxUploadBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is too large. Maximum allowed size is 5 MB."})
		return
	}
	if !isAllowedImageContentType(meta.ContentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported content type. Allowed types are image/jpeg, image/png, image/webp, and image/gif."})
		return
	}

	if err := h.familyRepo.UpdateLogoURL(c.Request.Context(), familyID, req.ObjectKey); err != nil {
		respondInternalErrorWithMessage(c, "Unable to update family logo at the moment.", err)
		return
	}

	downloadURL, err := h.r2.CreatePresignedDownload(c.Request.Context(), req.ObjectKey)
	if err != nil {
		respondInternalErrorWithMessage(c, "Unable to generate family logo download URL at the moment.", err)
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
