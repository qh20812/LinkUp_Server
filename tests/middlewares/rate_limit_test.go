package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"linkup/middlewares"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRateLimiter_AllowsWithinLimit(t *testing.T) {
	l := middlewares.NewRateLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !l.Allow("ip-1") {
			t.Fatalf("hit %d rejected, want allowed", i)
		}
	}
	if l.Allow("ip-1") {
		t.Error("4th hit allowed, want rejected")
	}
	if !l.Allow("ip-2") {
		t.Error("different key rejected, want per-key isolation")
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	l := middlewares.NewRateLimiter(1, 50*time.Millisecond)
	if !l.Allow("ip-1") {
		t.Fatal("first hit rejected")
	}
	if l.Allow("ip-1") {
		t.Fatal("immediate second hit allowed")
	}
	time.Sleep(60 * time.Millisecond)
	if !l.Allow("ip-1") {
		t.Error("hit after window rejected, want allowed")
	}
}

func TestRateLimitMiddleware_429(t *testing.T) {
	router := gin.New()
	router.Use(middlewares.RateLimit(2, time.Minute))
	router.POST("/api/communities", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/communities", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("request %d: status = %d, want 201", i, w.Code)
		}
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/communities", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", w.Code)
	}
	if body := w.Body.String(); !strings.Contains(body, "community.CREATE_RATE_LIMITED") {
		t.Fatalf("body = %q, want rate-limit code", body)
	}
}
