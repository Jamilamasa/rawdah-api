package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rawdah/rawdah-api/internal/models"
)

func RoleGuard(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString(string(models.ContextKeyRole))
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
	}
}
