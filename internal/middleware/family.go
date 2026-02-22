package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rawdah/rawdah-api/internal/models"
)

func FamilyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		paramFamilyID := c.Param("family_id")
		if paramFamilyID == "" {
			c.Next()
			return
		}
		jwtFamilyID := c.GetString(string(models.ContextKeyFamilyID))
		if paramFamilyID != jwtFamilyID {
			c.AbortWithStatusJSON(404, gin.H{"error": "not found"})
			return
		}
		c.Next()
	}
}
