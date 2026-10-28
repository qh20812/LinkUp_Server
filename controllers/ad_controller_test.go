package controllers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"linkup/dto"
	"linkup/models"
	"linkup/services"
)

// fakeAdService chỉ override 2 method cần test; phần còn lại của
// services.AdService được nhúng để không phải implement cả interface.
type fakeAdService struct {
	services.AdService

	createCalled bool
	createInput  dto.CreateAdInput
	createFiles  int
	createUser   string
	createAsAdmin bool
	createErr    error

	updateCalled  bool
	updateID      string
	updateUser    string
	updateInput   dto.UpdateAdInput
	updateErr     error

	ad *models.Ad
}

func (f *fakeAdService) CreateAdWithMedia(_ *gin.Context, input dto.CreateAdInput, files []*multipart.FileHeader, partnerID string, platformAdmin bool) (*models.Ad, error) {
	f.createCalled = true
	f.createInput = input
	f.createFiles = len(files)
	f.createUser = partnerID
	f.createAsAdmin = platformAdmin
	if f.createErr != nil {
		return nil, f.createErr
	}
	if f.ad != nil {
		return f.ad, nil
	}
	return &models.Ad{ID: "ad-1", Title: input.Title, Status: models.AdStatusActive}, nil
}

func (f *fakeAdService) UpdateAd(id, partnerID string, input dto.UpdateAdInput) (*models.Ad, error) {
	f.updateCalled = true
	f.updateID = id
	f.updateUser = partnerID
	f.updateInput = input
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return &models.Ad{ID: id, Title: input.Title, Status: models.AdStatusActive}, nil
}

func newAdMultipartContext(t *testing.T, method, path string, fields map[string]string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write field %q: %v", key, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, &buf)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	c.Set("userID", "user-1")
	return c, w
}

func newAdJSONContext(t *testing.T, method, path, body string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userID", "user-1")
	return c, w
}

func newAdParamContext(t *testing.T, method, path, adID string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	c, w := newAdJSONContext(t, method, path, "")
	c.Params = gin.Params{{Key: "id", Value: adID}}
	return c, w
}

func TestParseAdTime(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantNil bool
		wantErr bool
		want    time.Time
	}{
		{name: "rỗng", raw: "", wantNil: true},
		{name: "khoảng trắng", raw: "   ", wantNil: true},
		{name: "RFC3339", raw: "2026-10-27T23:30:00Z", want: time.Date(2026, 10, 27, 23, 30, 0, 0, time.UTC)},
		{name: "datetime-local", raw: "2026-10-27T23:30", want: time.Date(2026, 10, 27, 23, 30, 0, 0, time.Local)},
		{name: "date-only", raw: "2026-10-27", want: time.Date(2026, 10, 27, 0, 0, 0, 0, time.Local)},
		{name: "sai định dạng", raw: "27/10/2026", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseAdTime(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected %v, got nil", tc.want)
			}
			if !got.Equal(tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, *got)
			}
		})
	}
}

func TestCreateAd_AcceptsDataJSONField(t *testing.T) {
	svc := &fakeAdService{}
	ctrl := NewAdController(svc, nil)

	payload := `{"title":"Khuyến mãi","content":"Nội dung","target_url":"https://linkup.vn","budget":100000,` +
		`"started_at":"2026-10-27T23:30","target_age_min":18,"target_age_max":40}`
	c, w := newAdMultipartContext(t, http.MethodPost, "/api/ads-management", map[string]string{"data": payload})

	ctrl.CreateAd(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%s)", w.Code, w.Body.String())
	}
	if !svc.createCalled {
		t.Fatalf("service was not called")
	}
	if svc.createInput.Title != "Khuyến mãi" || svc.createInput.Budget != 100000 {
		t.Fatalf("unexpected input: %+v", svc.createInput)
	}
	if svc.createUser != "user-1" {
		t.Fatalf("expected partnerID user-1, got %q", svc.createUser)
	}
	if svc.createInput.Format != "image" {
		t.Fatalf("expected default format image, got %q", svc.createInput.Format)
	}
	if svc.createInput.StartedAt == nil {
		t.Fatalf("started_at bị mất khi parse")
	}
	if want := time.Date(2026, 10, 27, 23, 30, 0, 0, time.Local); !svc.createInput.StartedAt.Equal(want) {
		t.Fatalf("expected started_at %v, got %v", want, *svc.createInput.StartedAt)
	}

	body := decodeBody(t, w)
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("missing data object: %v", body)
	}
	// Không còn bước duyệt pending: ad trả về phải là active.
	if data["status"] != string(models.AdStatusActive) {
		t.Fatalf("expected status active, got %v", data["status"])
	}
	// Không có userRole (partner) => không được coi là admin.
	if svc.createAsAdmin {
		t.Fatalf("partner không được tạo với quyền admin")
	}
}

func TestCreateAd_PlatformAdminBypassesSubscription(t *testing.T) {
	svc := &fakeAdService{}
	ctrl := NewAdController(svc, nil)

	c, w := newAdMultipartContext(t, http.MethodPost, "/api/ads-management", map[string]string{
		"data": `{"title":"Ad của admin","content":"Nội dung","target_url":"https://linkup.vn","budget":100}`,
	})
	c.Set("userRole", string(models.RoleAdmin))

	ctrl.CreateAd(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%s)", w.Code, w.Body.String())
	}
	if !svc.createAsAdmin {
		t.Fatalf("ADMIN phải được tạo ad không cần gói")
	}
}

func TestCreateAd_AcceptsFlatFormFields(t *testing.T) {
	svc := &fakeAdService{}
	ctrl := NewAdController(svc, nil)

	c, w := newAdMultipartContext(t, http.MethodPost, "/api/ads-management", map[string]string{
		"title":           "Quảng cáo cũ",
		"content":         "Nội dung",
		"target_url":      "https://linkup.vn",
		"budget":          "500",
		"format":          "video",
		"target_locations": `["Hà Nội"]`,
	})

	ctrl.CreateAd(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%s)", w.Code, w.Body.String())
	}
	if svc.createInput.Format != "video" {
		t.Fatalf("expected format video, got %q", svc.createInput.Format)
	}
	if len(svc.createInput.TargetLocations) != 1 || svc.createInput.TargetLocations[0] != "Hà Nội" {
		t.Fatalf("unexpected target_locations: %v", svc.createInput.TargetLocations)
	}
}

func TestCreateAd_AcceptsJSONBody(t *testing.T) {
	svc := &fakeAdService{}
	ctrl := NewAdController(svc, nil)

	body := `{"title":"JSON","content":"Nội dung","target_url":"https://linkup.vn","budget":10,` +
		`"expires_at":"2026-12-31"}`
	c, w := newAdJSONContext(t, http.MethodPost, "/api/ads-management", body)

	ctrl.CreateAd(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%s)", w.Code, w.Body.String())
	}
	if svc.createInput.ExpiresAt == nil {
		t.Fatalf("expires_at bị mất khi parse")
	}
	if svc.createFiles != 0 {
		t.Fatalf("expected no media files, got %d", svc.createFiles)
	}
}

func TestCreateAd_MissingTitleRejects(t *testing.T) {
	svc := &fakeAdService{}
	ctrl := NewAdController(svc, nil)

	c, w := newAdMultipartContext(t, http.MethodPost, "/api/ads-management", map[string]string{
		"data": `{"content":"Nội dung","target_url":"https://linkup.vn","budget":100}`,
	})

	ctrl.CreateAd(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (%s)", w.Code, w.Body.String())
	}
	if svc.createCalled {
		t.Fatalf("service must not be called khi dữ liệu không hợp lệ")
	}
	if body := decodeBody(t, w); !strings.Contains(toStr(body["message"]), "tiêu đề") {
		t.Fatalf("expected tiêu đề error, got %v", body["message"])
	}
}

func TestCreateAd_InvalidStartedAtRejects(t *testing.T) {
	svc := &fakeAdService{}
	ctrl := NewAdController(svc, nil)

	c, w := newAdMultipartContext(t, http.MethodPost, "/api/ads-management", map[string]string{
		"data": `{"title":"A","content":"B","target_url":"https://linkup.vn","budget":100,"started_at":"27/10/2026"}`,
	})

	ctrl.CreateAd(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (%s)", w.Code, w.Body.String())
	}
	if body := decodeBody(t, w); !strings.Contains(toStr(body["message"]), "không hợp lệ") {
		t.Fatalf("expected thời gian error, got %v", body["message"])
	}
}

func TestUpdateAd_AcceptsJSONBody(t *testing.T) {
	svc := &fakeAdService{}
	ctrl := NewAdController(svc, nil)

	c, w := newAdParamContext(t, http.MethodPatch, "/api/ads-management/ad-9", "ad-9")
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/ads-management/ad-9",
		strings.NewReader(`{"title":"Mới","budget":200,"started_at":"2026-11-01T08:00"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	ctrl.UpdateAd(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
	if !svc.updateCalled || svc.updateID != "ad-9" || svc.updateInput.Budget != 200 {
		t.Fatalf("unexpected update call: %+v", svc)
	}
	if svc.updateInput.StartedAt == nil {
		t.Fatalf("started_at bị mất khi parse")
	}
}

func TestUpdateAd_AcceptsDataFieldMultipart(t *testing.T) {
	svc := &fakeAdService{}
	ctrl := NewAdController(svc, nil)

	c, w := newAdMultipartContext(t, http.MethodPatch, "/api/ads-management/ad-9", map[string]string{
		"data": `{"title":"Mới","content":"Nội dung","target_url":"https://linkup.vn","budget":200}`,
	})
	c.Params = gin.Params{{Key: "id", Value: "ad-9"}}

	ctrl.UpdateAd(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
	if svc.updateUser != "user-1" {
		t.Fatalf("expected partnerID user-1, got %q", svc.updateUser)
	}
}

func TestUpdateAd_NegativeBudgetRejects(t *testing.T) {
	svc := &fakeAdService{}
	ctrl := NewAdController(svc, nil)

	c, w := newAdParamContext(t, http.MethodPatch, "/api/ads-management/ad-9", "ad-9")
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/ads-management/ad-9",
		strings.NewReader(`{"budget":-1}`))
	c.Request.Header.Set("Content-Type", "application/json")

	ctrl.UpdateAd(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (%s)", w.Code, w.Body.String())
	}
	if svc.updateCalled {
		t.Fatalf("service must not be called khi budget âm")
	}
}

func toStr(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	raw, _ := json.Marshal(v)
	return string(raw)
}
