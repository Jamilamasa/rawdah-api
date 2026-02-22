package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rawdah/rawdah-api/internal/models"
)

type fakePermRepo struct {
	perms []string
	err   error
}

func (f fakePermRepo) GetPermissions(_ context.Context, _, _ string) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.perms, nil
}

func TestAdultPermissionGuard_AllowsParentAndChild(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, role := range []string{"parent", "child"} {
		t.Run(role, func(t *testing.T) {
			r := gin.New()
			r.GET("/x", func(c *gin.Context) {
				c.Set(string(models.ContextKeyRole), role)
				c.Next()
			}, AdultPermissionGuard(fakePermRepo{}, "view_tasks"), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/x", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", w.Code)
			}
		})
	}
}

func TestAdultPermissionGuard_EnforcesAdultPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	makeRouter := func(repo PermissionRepo) *gin.Engine {
		r := gin.New()
		r.GET("/x", func(c *gin.Context) {
			c.Set(string(models.ContextKeyRole), "adult_relative")
			c.Set(string(models.ContextKeyUserID), "user-1")
			c.Set(string(models.ContextKeyFamilyID), "family-1")
			c.Next()
		}, AdultPermissionGuard(repo, "view_tasks"), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		return r
	}

	t.Run("forbidden when missing permission", func(t *testing.T) {
		r := makeRouter(fakePermRepo{perms: []string{"assign_tasks"}})
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("forbidden on repo error", func(t *testing.T) {
		r := makeRouter(fakePermRepo{err: errors.New("db down")})
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("allowed when permission exists", func(t *testing.T) {
		r := makeRouter(fakePermRepo{perms: []string{"view_tasks"}})
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})
}
