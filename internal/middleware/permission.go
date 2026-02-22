package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/rawdah/rawdah-api/internal/models"
)

type PermissionRepo interface {
	GetPermissions(ctx context.Context, userID, familyID string) ([]string, error)
}

func PermissionGuard(repo PermissionRepo, perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString(string(models.ContextKeyRole))
		if role == "parent" {
			c.Next()
			return
		}
		if role == "child" {
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
			return
		}
		userID := c.GetString(string(models.ContextKeyUserID))
		familyID := c.GetString(string(models.ContextKeyFamilyID))
		perms, err := repo.GetPermissions(c.Request.Context(), userID, familyID)
		if err != nil || !hasPermission(perms, perm) {
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

// AdultPermissionGuard only checks permissions for adult relatives.
// Parents and children pass through and are expected to be handled by role middleware.
func AdultPermissionGuard(repo PermissionRepo, perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString(string(models.ContextKeyRole))
		if role != "adult_relative" {
			c.Next()
			return
		}

		userID := c.GetString(string(models.ContextKeyUserID))
		familyID := c.GetString(string(models.ContextKeyFamilyID))
		perms, err := repo.GetPermissions(c.Request.Context(), userID, familyID)
		if err != nil || !hasPermission(perms, perm) {
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
			return
		}

		c.Next()
	}
}

func hasPermission(perms []string, perm string) bool {
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}
