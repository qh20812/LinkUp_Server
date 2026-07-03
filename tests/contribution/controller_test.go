package contribution_test

import (
	"bytes"
	"encoding/json"
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

func TestContributionController_InputGuards(t *testing.T) {
	ctrl := controllers.NewContributionController(&services.ContributionService{})
	unauthRouter := gin.New()
	unauthRouter.GET("/api/communities/:communityID/policy", ctrl.GetPolicy)
	unauthRouter.PUT("/api/communities/:communityID/policy", ctrl.UpdatePolicy)
	unauthRouter.POST("/api/communities/:communityID/challenges", ctrl.CreateChallenge)
	unauthRouter.POST("/api/communities/:communityID/challenges/:challengeID/join", ctrl.JoinChallenge)

	authRouter := gin.New()
	authRouter.Use(func(c *gin.Context) {
		c.Set("userID", "user-1")
		c.Next()
	})
	authRouter.PUT("/api/communities/:communityID/policy", ctrl.UpdatePolicy)
	authRouter.POST("/api/communities/:communityID/challenges", ctrl.CreateChallenge)

	tests := []struct {
		name       string
		router     *gin.Engine
		method     string
		url        string
		body       any
		wantStatus int
		wantErr    string
	}{
		{
			name:       "unauthorized policy",
			router:     unauthRouter,
			method:     http.MethodGet,
			url:        "/api/communities/community-1/policy",
			wantStatus: http.StatusUnauthorized,
			wantErr:    "Không tìm thấy thông tin chứng thực người dùng",
		},
		{
			name:       "unauthorized join challenge",
			router:     unauthRouter,
			method:     http.MethodPost,
			url:        "/api/communities/community-1/challenges/challenge-1/join",
			wantStatus: http.StatusUnauthorized,
			wantErr:    "Không tìm thấy thông tin chứng thực người dùng",
		},
		{
			name:       "invalid policy json after auth",
			router:     authRouter,
			method:     http.MethodPut,
			url:        "/api/communities/community-1/policy",
			body:       []byte(`{invalid`),
			wantStatus: http.StatusBadRequest,
			wantErr:    "Dữ liệu đầu vào không hợp lệ",
		},
		{
			name:       "invalid challenge json after auth",
			router:     authRouter,
			method:     http.MethodPost,
			url:        "/api/communities/community-1/challenges",
			body:       []byte(`{invalid`),
			wantStatus: http.StatusBadRequest,
			wantErr:    "Dữ liệu đầu vào không hợp lệ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if bodyBytes, ok := tt.body.([]byte); ok {
				req = httptest.NewRequest(tt.method, tt.url, bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
			} else if tt.body != nil {
				bodyBytes, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(tt.method, tt.url, bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.url, nil)
			}

			w := httptest.NewRecorder()
			tt.router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantErr != "" && !bytes.Contains(w.Body.Bytes(), []byte(tt.wantErr)) {
				t.Fatalf("body = %q, want to contain %q", w.Body.String(), tt.wantErr)
			}
		})
	}
}
