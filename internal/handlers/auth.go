package handlers

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
		Email      string `json:"email"       binding:"required,email"`
		Password   string `json:"password"    binding:"required,min=8"`
	}
	if !bindJSONWithValidation(c, &req) {
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required."})
		return
	}

	if !slugRegex.MatchString(req.Slug) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug may only contain lowercase letters, numbers, and hyphens."})
		return
	}

	tokens, err := h.svc.Signup(c.Request.Context(), services.SignupInput{
		FamilyName: req.FamilyName,
		Slug:       req.Slug,
		Name:       name,
		Email:      req.Email,
		Password:   req.Password,
	})
	if err != nil {
		switch err {
		case services.ErrSlugTaken:
			c.JSON(http.StatusConflict, gin.H{"error": "This family slug is already in use."})
		case services.ErrEmailTaken:
			c.JSON(http.StatusConflict, gin.H{"error": "An account with this email already exists."})
		default:
			respondInternalError(c, err)
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
	if !bindJSONWithValidation(c, &req) {
		return
	}

	tokens, err := h.svc.Signin(c.Request.Context(), services.SigninInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "The provided credentials are incorrect."})
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
	if !bindJSONWithValidation(c, &req) {
		return
	}

	tokens, err := h.svc.ChildSignin(c.Request.Context(), services.ChildSigninInput{
		FamilySlug: req.FamilySlug,
		Username:   req.Username,
		Password:   req.Password,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "The provided credentials are incorrect."})
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token is missing. Please sign in again."})
		return
	}

	tokens, err := h.svc.Refresh(c.Request.Context(), rawToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token is invalid or expired. Please sign in again."})
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
		c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found."})
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
	if !bindJSONWithValidation(c, &req) {
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), userID, familyID, req.CurrentPassword, req.NewPassword); err != nil {
		if err == services.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "The current password is incorrect."})
			return
		}
		if err == services.ErrPasswordTooShort {
			c.JSON(http.StatusBadRequest, gin.H{"error": "The new password does not meet minimum length requirements."})
			return
		}
		respondInternalError(c, err)
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

func bindJSONWithValidation(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		writeValidationError(c, err)
		return false
	}
	return true
}

func writeValidationError(c *gin.Context, err error) {
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		details := make([]map[string]string, 0, len(verrs))
		for _, fe := range verrs {
			details = append(details, map[string]string{
				"field": fe.Field(),
				"error": validationMessage(fe),
				"tag":   fe.Tag(),
				"param": fe.Param(),
			})
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Request validation failed.",
			"details": details,
		})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "The request body is invalid. Please verify required fields and value formats."})
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		if fe.Param() != "" {
			return "must have at least " + fe.Param() + " characters"
		}
		return "does not meet minimum length"
	default:
		if fe.Param() != "" {
			return "failed validation: " + fe.Tag() + "=" + fe.Param()
		}
		return "failed validation: " + fe.Tag()
	}
}
