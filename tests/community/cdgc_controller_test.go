package community_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"linkup/controllers"
	"linkup/services"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newCommunityController() *controllers.CommunityController {
	return controllers.NewCommunityController(&services.CommunityService{}, nil)
}

func TestCreateCommunity_AuthGuard(t *testing.T) {
	ctrl := newCommunityController()
	unauthRouter := gin.New()
	unauthRouter.Use(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
	})
	unauthRouter.POST("/api/communities", ctrl.CreateCommunity)

	req := httptest.NewRequest(http.MethodPost, "/api/communities", nil)
	w := httptest.NewRecorder()
	unauthRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusUnauthorized, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("Không tìm thấy thông tin chứng thực người dùng")) {
		t.Fatalf("body = %q, want to contain auth error", w.Body.String())
	}
}

func TestCreateCommunity_InvalidBody(t *testing.T) {
	ctrl := newCommunityController()
	authRouter := gin.New()
	authRouter.Use(func(c *gin.Context) {
		c.Set("userID", "user-1")
		c.Next()
	})
	authRouter.POST("/api/communities", ctrl.CreateCommunity)

	req := httptest.NewRequest(http.MethodPost, "/api/communities",
		bytes.NewReader([]byte(`not a valid multipart form`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	authRouter.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("Dữ liệu đầu vào không hợp lệ")) {
		t.Fatalf("body = %q, want to contain parse error", w.Body.String())
	}
}

func TestRequestJoin_AuthGuard(t *testing.T) {
	ctrl := newCommunityController()
	unauthRouter := gin.New()
	unauthRouter.Use(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
	})
	unauthRouter.POST("/api/communities/:communityID/join", ctrl.RequestJoin)

	req := httptest.NewRequest(http.MethodPost, "/api/communities/community-1/join", nil)
	w := httptest.NewRecorder()
	unauthRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusUnauthorized, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("Không tìm thấy thông tin chứng thực người dùng")) {
		t.Fatalf("body = %q, want to contain auth error", w.Body.String())
	}
}

func TestRequestJoin_MissingCommunityID(t *testing.T) {
	ctrl := newCommunityController()
	authRouter := gin.New()
	authRouter.Use(func(c *gin.Context) {
		c.Set("userID", "user-1")
		c.Next()
	})
	authRouter.POST("/api/communities/:communityID/join", ctrl.RequestJoin)

	req := httptest.NewRequest(http.MethodPost, "/api/communities//join", nil)
	w := httptest.NewRecorder()
	authRouter.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("communityID là bắt buộc")) {
		t.Fatalf("body = %q, want to contain communityID error", w.Body.String())
	}
}

func TestApproveJoinRequest_AuthGuard(t *testing.T) {
	ctrl := newCommunityController()
	unauthRouter := gin.New()
	unauthRouter.Use(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
	})
	unauthRouter.PUT("/api/communities/:communityID/join-requests/:requestID/approve", ctrl.ApproveJoinRequest)

	req := httptest.NewRequest(http.MethodPut,
		"/api/communities/community-1/join-requests/request-1/approve", nil)
	w := httptest.NewRecorder()
	unauthRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusUnauthorized, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("Không tìm thấy thông tin chứng thực người dùng")) {
		t.Fatalf("body = %q, want to contain auth error", w.Body.String())
	}
}
