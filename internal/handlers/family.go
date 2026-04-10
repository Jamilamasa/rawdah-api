package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/services"
)

type FamilyHandler struct {
	svc    *services.FamilyService
	signer mediaURLSigner
}

func NewFamilyHandler(svc *services.FamilyService, signer mediaURLSigner) *FamilyHandler {
	return &FamilyHandler{svc: svc, signer: signer}
}

func (h *FamilyHandler) Get(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	family, err := h.svc.GetFamily(c.Request.Context(), familyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found."})
		return
	}
	applySignedMediaToFamily(c.Request.Context(), h.signer, family)
	c.JSON(http.StatusOK, family)
}

func (h *FamilyHandler) Update(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	var req struct {
		Name    string  `json:"name"     binding:"required"`
		LogoURL *string `json:"logo_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The request body is invalid. Please verify required fields and value formats."})
		return
	}

	family, err := h.svc.UpdateFamily(c.Request.Context(), familyID, req.Name, req.LogoURL)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	applySignedMediaToFamily(c.Request.Context(), h.signer, family)
	c.JSON(http.StatusOK, family)
}

func (h *FamilyHandler) ListMembers(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	members, err := h.svc.ListMembers(c.Request.Context(), familyID)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	applySignedMediaToUsers(c.Request.Context(), h.signer, members)
	c.JSON(http.StatusOK, gin.H{"members": members})
}

func (h *FamilyHandler) CreateMember(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	creatorID := c.GetString(string(models.ContextKeyUserID))

	var req struct {
		Role             string     `json:"role"             binding:"required"`
		Name             string     `json:"name"             binding:"required"`
		Username         *string    `json:"username"`
		Email            *string    `json:"email"`
		Password         string     `json:"password"         binding:"required"`
		ChildAge         *int       `json:"child_age"`
		DateOfBirth      *time.Time `json:"date_of_birth"`
		GameLimitMinutes int        `json:"game_limit_minutes"`
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
	cid, err := uuid.Parse(creatorID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication is required or your session is invalid."})
		return
	}

	member, err := h.svc.CreateMember(c.Request.Context(), services.CreateMemberInput{
		FamilyID:         fid,
		Role:             req.Role,
		Name:             req.Name,
		Username:         req.Username,
		Email:            req.Email,
		Password:         req.Password,
		ChildAge:         req.ChildAge,
		DateOfBirth:      req.DateOfBirth,
		GameLimitMinutes: req.GameLimitMinutes,
		CreatedBy:        cid,
	})
	if err != nil {
		switch err {
		case services.ErrPasswordTooShort:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case services.ErrInvalidRole:
			c.JSON(http.StatusBadRequest, gin.H{"error": "The provided role is invalid."})
		case services.ErrInvalidMemberData:
			c.JSON(http.StatusBadRequest, gin.H{"error": "The member data is invalid."})
		default:
			respondInternalError(c, err)
		}
		return
	}
	applySignedMediaToUser(c.Request.Context(), h.signer, member)
	c.JSON(http.StatusCreated, member)
}

func (h *FamilyHandler) GetMember(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	memberID := c.Param("id")

	member, err := h.svc.GetMember(c.Request.Context(), memberID, familyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found."})
		return
	}
	applySignedMediaToUser(c.Request.Context(), h.signer, member)
	c.JSON(http.StatusOK, member)
}

func (h *FamilyHandler) UpdateMember(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	memberID := c.Param("id")

	var req struct {
		Name             *string    `json:"name"`
		Theme            *string    `json:"theme"`
		GameLimitMinutes *int       `json:"game_limit_minutes"`
		ChildAge         *int       `json:"child_age"`
		DateOfBirth      *time.Time `json:"date_of_birth"`
		Password         *string    `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The request body is invalid. Please verify required fields and value formats."})
		return
	}

	updates := map[string]interface{}{
		"name":               req.Name,
		"theme":              req.Theme,
		"game_limit_minutes": req.GameLimitMinutes,
		"child_age":          req.ChildAge,
		"date_of_birth":      req.DateOfBirth,
		"password":           req.Password,
	}

	member, err := h.svc.UpdateMember(c.Request.Context(), memberID, familyID, updates)
	if err != nil {
		if err == services.ErrInvalidMemberData {
			c.JSON(http.StatusBadRequest, gin.H{"error": "The member data is invalid."})
			return
		}
		if err == services.ErrPasswordTooShort {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found."})
		return
	}
	applySignedMediaToUser(c.Request.Context(), h.signer, member)
	c.JSON(http.StatusOK, member)
}

func (h *FamilyHandler) DeactivateMember(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	memberID := c.Param("id")

	if err := h.svc.DeactivateMember(c.Request.Context(), memberID, familyID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member deactivated"})
}

func (h *FamilyHandler) RantCount(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	childID := c.Param("id")

	count, err := h.svc.GetRantCount(c.Request.Context(), childID, familyID)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (h *FamilyHandler) ListAccessControl(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))

	list, err := h.svc.ListAccessControl(c.Request.Context(), familyID)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_control": list})
}

func (h *FamilyHandler) SetAccessControl(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	grantorIDStr := c.GetString(string(models.ContextKeyUserID))
	granteeID := c.Param("grantee_id")

	var req struct {
		Permissions []string `json:"permissions" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The request body is invalid. Please verify required fields and value formats."})
		return
	}

	grantorID, err := uuid.Parse(grantorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Grantor ID format is invalid."})
		return
	}

	fac, err := h.svc.SetPermissions(c.Request.Context(), granteeID, familyID, grantorID, req.Permissions)
	if err != nil {
		if err == services.ErrMemberNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found."})
			return
		}
		if err == services.ErrInvalidPermissions {
			c.JSON(http.StatusBadRequest, gin.H{"error": "One or more permissions are invalid."})
			return
		}
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, fac)
}

func (h *FamilyHandler) RevokeAccessControl(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	granteeID := c.Param("grantee_id")

	if err := h.svc.RevokePermissions(c.Request.Context(), granteeID, familyID); err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "access revoked"})
}
