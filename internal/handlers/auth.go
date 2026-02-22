package handlers

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/services"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

type AuthHandler struct {
	svc    *services.AuthService
	signer mediaURLSigner
}

func NewAuthHandler(svc *services.AuthService, signer mediaURLSigner) *AuthHandler {
	return &AuthHandler{svc: svc, signer: signer}
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req struct {
		FamilyName string `json:"family_name" binding:"required"`
		Slug       string `json:"slug"        binding:"required"`
		Name       string `json:"name"`
		ParentName string `json:"parent_name"`
		Email      string `json:"email"       binding:"required,email"`
		Password   string `json:"password"    binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	parentName := req.Name
	if parentName == "" {
		parentName = req.ParentName
	}
	if parentName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	if !slugRegex.MatchString(req.Slug) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug must only contain lowercase letters, numbers, and hyphens"})
		return
	}

	tokens, err := h.svc.Signup(c.Request.Context(), services.SignupInput{
		FamilyName: req.FamilyName,
		Slug:       req.Slug,
		ParentName: parentName,
		Email:      req.Email,
		Password:   req.Password,
	})
	if err != nil {
		switch err {
		case services.ErrSlugTaken:
			c.JSON(http.StatusConflict, gin.H{"error": "slug already taken"})
		case services.ErrEmailTaken:
			c.JSON(http.StatusConflict, gin.H{"error": "email already taken"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	applySignedMediaToUser(c.Request.Context(), h.signer, tokens.User)
	applySignedMediaToFamily(c.Request.Context(), h.signer, tokens.Family)

	setRefreshCookie(c, tokens.RefreshToken)
	c.JSON(http.StatusCreated, gin.H{
		"access_token": tokens.AccessToken,
		"user":         tokens.User,
		"family":       tokens.Family,
	})
}

func (h *AuthHandler) Signin(c *gin.Context) {
	var req struct {
		Email    string `json:"email"    binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	tokens, err := h.svc.Signin(c.Request.Context(), services.SigninInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	applySignedMediaToUser(c.Request.Context(), h.signer, tokens.User)
	applySignedMediaToFamily(c.Request.Context(), h.signer, tokens.Family)

	setRefreshCookie(c, tokens.RefreshToken)
	c.JSON(http.StatusOK, gin.H{
		"access_token": tokens.AccessToken,
		"user":         tokens.User,
		"family":       tokens.Family,
	})
}

func (h *AuthHandler) ChildSignin(c *gin.Context) {
	var req struct {
		FamilySlug string `json:"family_slug" binding:"required"`
		Username   string `json:"username"    binding:"required"`
		Password   string `json:"password"    binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	tokens, err := h.svc.ChildSignin(c.Request.Context(), services.ChildSigninInput{
		FamilySlug: req.FamilySlug,
		Username:   req.Username,
		Password:   req.Password,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	applySignedMediaToUser(c.Request.Context(), h.signer, tokens.User)
	applySignedMediaToFamily(c.Request.Context(), h.signer, tokens.Family)

	// Children do not receive refresh tokens. Clear any existing cookie.
	clearRefreshCookie(c)
	c.JSON(http.StatusOK, gin.H{
		"access_token": tokens.AccessToken,
		"user":         tokens.User,
		"family":       tokens.Family,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	rawToken, err := c.Cookie("rawdah_refresh")
	if err != nil || rawToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	tokens, err := h.svc.Refresh(c.Request.Context(), rawToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	applySignedMediaToUser(c.Request.Context(), h.signer, tokens.User)
	applySignedMediaToFamily(c.Request.Context(), h.signer, tokens.Family)

	setRefreshCookie(c, tokens.RefreshToken)
	c.JSON(http.StatusOK, gin.H{
		"access_token": tokens.AccessToken,
		"user":         tokens.User,
		"family":       tokens.Family,
	})
}

func (h *AuthHandler) Signout(c *gin.Context) {
	rawToken, err := c.Cookie("rawdah_refresh")
	if err == nil && rawToken != "" {
		_ = h.svc.Signout(c.Request.Context(), rawToken)
	}
	clearRefreshCookie(c)
	c.JSON(http.StatusOK, gin.H{"message": "signed out"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString(string(models.ContextKeyUserID))
	familyID := c.GetString(string(models.ContextKeyFamilyID))

	user, family, err := h.svc.Me(c.Request.Context(), userID, familyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	applySignedMediaToUser(c.Request.Context(), h.signer, user)
	applySignedMediaToFamily(c.Request.Context(), h.signer, family)
	c.JSON(http.StatusOK, gin.H{"user": user, "family": family})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.GetString(string(models.ContextKeyUserID))
	familyID := c.GetString(string(models.ContextKeyFamilyID))

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password"     binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), userID, familyID, req.CurrentPassword, req.NewPassword); err != nil {
		if err == services.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "current password is incorrect"})
			return
		}
		if err == services.ErrPasswordTooShort {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new password is too short"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated"})
}

func setRefreshCookie(c *gin.Context, token string) {
	if token == "" {
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"rawdah_refresh",
		token,
		int(7*24*60*60), // 7 days
		"/",
		"",
		true, // secure
		true, // httpOnly
	)
}

func clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("rawdah_refresh", "", -1, "/", "", true, true)
}
