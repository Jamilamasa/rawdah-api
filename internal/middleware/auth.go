package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rawdah/rawdah-api/internal/auth"
	"github.com/rawdah/rawdah-api/internal/config"
	"github.com/rawdah/rawdah-api/internal/models"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}
		claims, err := auth.ValidateAccessToken(tokenStr, cfg.JWTAccessSecret)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}
		c.Set(string(models.ContextKeyUserID), claims.UserID.String())
		c.Set(string(models.ContextKeyFamilyID), claims.FamilyID.String())
		c.Set(string(models.ContextKeyRole), claims.Role)
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	}

	// WebSocket connections commonly pass token in query params.
	if c.Request.URL.Path == "/ws" {
		return strings.TrimSpace(c.Query("token"))
	}

	return ""
}
