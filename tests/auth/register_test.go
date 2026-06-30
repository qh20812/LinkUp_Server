package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"linkup/config"
	"linkup/controllers"
	"linkup/db"
	"linkup/repository"
	"linkup/routes"
	"linkup/services"
	"linkup/validations"
	"linkup/ws"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// validation unit tests
func TestValidateRegisterInput(t *testing.T) {
	v := validations.NewAuthValidation()

	tests := []struct {
		name        string
		displayName string
		email       string
		password    string
		wantErr     string
	}{
		{"valid input", "Test User", "test@example.com", "Password123!", ""},
		{"empty display name", "", "test@example.com", "Password123!", "tên hiển thị không được để trống"},
		{"display name too short", "A", "test@example.com", "Password123!", "tên hiển thị phải có ít nhất 3 ký tự"},
		{"display name too long", string(make([]rune, 56)), "test@example.com", "Password123!", "tên hiển thị không được vượt quá 55 ký tự"},
		{"empty email", "Test User", "", "Password123!", "email không được để trống"},
		{"invalid email", "Test User", "not-an-email", "Password123!", "định dạng email không hợp lệ"},
		{"empty password", "Test User", "test@example.com", "", "mật khẩu không được để trống"},
		{"password too short", "Test User", "test@example.com", "Ab1!", "mật khẩu phải có ít nhất 8 ký tự"},
		{"password too long", "Test User", "test@example.com", "Aa1!" + string(make([]byte, 130)), "mật khẩu không được vượt quá 128 ký tự"},
		{"password missing uppercase", "Test User", "test@example.com", "password123!", "mật khẩu phải chứa ít nhất một chữ cái in hoa"},
		{"password missing lowercase", "Test User", "test@example.com", "PASSWORD123!", "mật khẩu phải chứa ít nhất một chữ cái in thường"},
		{"password missing digit", "Test User", "test@example.com", "Password!!!", "mật khẩu phải chứa ít nhất một chữ số"},
		{"password missing special char", "Test User", "test@example.com", "Password123", "mật khẩu phải chứa ít nhất một ký tự đặc biệt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateRegisterInput(tt.displayName, tt.email, tt.password)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.wantErr)
				} else if err.Error() != tt.wantErr {
					t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
				}
			}
		})
	}
}

// HTTP handler tests — only validation error paths (no DB needed)
func TestRegisterHandler_ValidationErrors(t *testing.T) {
	validation := validations.NewAuthValidation()
	ctrl := controllers.NewAuthController(nil, validation)

	router := gin.New()
	router.POST("/api/auth/register", ctrl.Register)

	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
		wantErr    string
	}{
		{
			name:       "missing all fields",
			body:       map[string]interface{}{},
			wantStatus: http.StatusBadRequest,
			wantErr:    "tên hiển thị không được để trống",
		},
		{
			name: "invalid email",
			body: map[string]interface{}{
				"display_name": "Test User",
				"email":        "not-an-email",
				"password":     "Password123!",
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    "định dạng email không hợp lệ",
		},
		{
			name: "weak password",
			body: map[string]interface{}{
				"display_name": "Test User",
				"email":        "test@example.com",
				"password":     "short",
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    "mật khẩu phải có ít nhất 8 ký tự",
		},
		{
			name: "invalid json body",
			body: nil,
			wantStatus: http.StatusBadRequest,
			wantErr:    "dữ liệu đầu vào không hợp lệ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body == nil {
				req = httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte(`{invalid`)))
			} else {
				bodyBytes, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(bodyBytes))
			}
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			var resp map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}
			if resp["error"] != tt.wantErr {
				t.Errorf("error = %q, want %q", resp["error"], tt.wantErr)
			}
		})
	}
}

// Integration test — requires TEST_DSN env var pointing to a MySQL database
func TestRegisterHandler_Success(t *testing.T) {
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		t.Skip("TEST_DSN not set; skipping integration test")
	}

	if err := config.LoadEnv(); err != nil {
		t.Fatalf("load env: %v", err)
	}
	env := config.GetEnv()

	// override DSN for test
	env.DBHost = ""
	env.DBPort = 0
	env.DBUser = ""
	env.DBPassword = ""
	env.DBName = ""

	database, err := db.ConnectDb(env)
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	defer database.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{Conn: database}), &gorm.Config{})
	if err != nil {
		t.Fatalf("init gorm: %v", err)
	}

	authRepo := repository.NewAuthRepository(gormDB)
	profileRepo := repository.NewProfileRepository(gormDB)
	authService := services.NewAuthService(authRepo, profileRepo, env)
	authValidation := validations.NewAuthValidation()
	authController := controllers.NewAuthController(authService, authValidation)

	hub := ws.NewHub()
	go hub.Run()

	router := gin.New()
	routes.RegisterAuthRoutes(router, authController, env)

	body := map[string]string{
		"display_name": "Integration Test",
		"email":        "integtest@example.com",
		"password":     "Password123!",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d; body = %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if _, ok := resp["user"]; !ok {
		t.Error("response missing 'user' field")
	}
	if _, ok := resp["tokens"]; !ok {
		t.Error("response missing 'tokens' field")
	}
}
