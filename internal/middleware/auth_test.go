package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/auth"
	"github.com/rawdah/rawdah-api/internal/config"
)

func TestAuthMiddlewareAcceptsBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{JWTAccessSecret: "test-secret"}
	userID := uuid.New()
	familyID := uuid.New()
	token, err := auth.IssueAccessToken(userID, familyID, "parent", cfg.JWTAccessSecret, 10*time.Minute)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	r := gin.New()
	r.GET("/v1/protected", AuthMiddleware(cfg), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuthMiddlewareAcceptsWSTokenFromQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{JWTAccessSecret: "test-secret"}
	token, err := auth.IssueAccessToken(uuid.New(), uuid.New(), "child", cfg.JWTAccessSecret, 10*time.Minute)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	r := gin.New()
	r.GET("/ws", AuthMiddleware(cfg), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ws?token="+token, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuthMiddlewareRejectsQueryTokenOnNonWSRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{JWTAccessSecret: "test-secret"}
	token, err := auth.IssueAccessToken(uuid.New(), uuid.New(), "child", cfg.JWTAccessSecret, 10*time.Minute)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	r := gin.New()
	r.GET("/v1/protected", AuthMiddleware(cfg), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/protected?token="+token, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
