package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"linkup/models"
	"linkup/services"
)

// ─── fake dữ liệu role ─────────────────────────────────────────────────────

func withPlatformRoles(t *testing.T, roles []models.RoleName, err error) {
	t.Helper()
	original := platformRoles
	platformRoles = func(_ *gorm.DB, _ string) ([]models.RoleName, error) {
		return roles, err
	}
	t.Cleanup(func() { platformRoles = original })
}

// ─── helper chạy middleware ───────────────────────────────────────────────

func runMiddleware(t *testing.T, mw gin.HandlerFunc, setUser bool) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if setUser {
		c.Set("userID", "user-1")
	}

	// Không có engine nên c.Next() không chạy handler nào; kết quả kiểm tra nằm
	// ở status/abort state mà middleware ghi lại.
	mw(c)

	return c, w
}

func TestRequireRoles_UserWithPartnerAndUserIsAllowed(t *testing.T) {
	// Lỗi thực tế: seed cấp cả PARTNER và USER nên middleware đọc một dòng có
	// thể nhận USER và trả 403 cho tài khoản đối tác.
	withPlatformRoles(t, []models.RoleName{models.RoleUser, models.RolePartner}, nil)

	_, w := runMiddleware(t, RequireRoles(nil, models.RoleSuperAdmin, models.RoleAdmin, models.RolePartner), true)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestRequireRoles_SetsMostPrivilegedAllowedRole(t *testing.T) {
	withPlatformRoles(t, []models.RoleName{models.RolePartner, models.RoleAdmin}, nil)

	c, w := runMiddleware(t, RequireRoles(nil, models.RoleSuperAdmin, models.RoleAdmin, models.RolePartner), true)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
	if got := c.GetString("userRole"); got != string(models.RoleAdmin) {
		t.Fatalf("expected userRole ADMIN, got %q", got)
	}
}

func TestRequireRoles_PartnerRoleSelectedWhenOnlyPartnerMatches(t *testing.T) {
	withPlatformRoles(t, []models.RoleName{models.RoleUser, models.RolePartner}, nil)

	c, _ := runMiddleware(t, RequireRoles(nil, models.RoleSuperAdmin, models.RoleAdmin, models.RolePartner), true)

	if got := c.GetString("userRole"); got != string(models.RolePartner) {
		t.Fatalf("expected userRole PARTNER, got %q", got)
	}
}

func TestRequireRoles_PlainUserIsDenied(t *testing.T) {
	withPlatformRoles(t, []models.RoleName{models.RoleUser}, nil)

	_, w := runMiddleware(t, RequireRoles(nil, models.RoleSuperAdmin, models.RoleAdmin, models.RolePartner), true)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestRequireRoles_NoRoleIsForbidden(t *testing.T) {
	withPlatformRoles(t, nil, nil)

	_, w := runMiddleware(t, RequireRoles(nil, models.RolePartner), true)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestRequireRoles_QueryErrorIsForbidden(t *testing.T) {
	withPlatformRoles(t, nil, errors.New("db down"))

	_, w := runMiddleware(t, RequireRoles(nil, models.RolePartner), true)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestRequireRoles_MissingUserIsUnauthorized(t *testing.T) {
	withPlatformRoles(t, []models.RoleName{models.RolePartner}, nil)

	_, w := runMiddleware(t, RequireRoles(nil, models.RolePartner), false)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d (%s)", w.Code, w.Body.String())
	}
}

// ─── effectiveRole / containsRole ─────────────────────────────────────────

func TestEffectiveRole(t *testing.T) {
	allowed := []models.RoleName{models.RoleSuperAdmin, models.RoleAdmin, models.RolePartner}

	tests := []struct {
		name   string
		roles  []models.RoleName
		want   models.RoleName
		wantOK bool
	}{
		{name: "USER + PARTNER chọn PARTNER", roles: []models.RoleName{models.RoleUser, models.RolePartner}, want: models.RolePartner, wantOK: true},
		{name: "PARTNER + ADMIN chọn ADMIN", roles: []models.RoleName{models.RolePartner, models.RoleAdmin}, want: models.RoleAdmin, wantOK: true},
		{name: "chỉ USER không khớp", roles: []models.RoleName{models.RoleUser}, wantOK: false},
		{name: "danh sách rỗng", roles: nil, wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := effectiveRole(tc.roles, allowed)
			if ok != tc.wantOK {
				t.Fatalf("expected ok=%v, got %v", tc.wantOK, ok)
			}
			if ok && got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestContainsRole(t *testing.T) {
	roles := []models.RoleName{models.RoleUser, models.RolePartner}
	if !containsRole(roles, models.RolePartner) {
		t.Fatalf("expected PARTNER to be found")
	}
	if containsRole(roles, models.RoleAdmin) {
		t.Fatalf("expected ADMIN to be missing")
	}
	if containsRole(nil, models.RoleUser) {
		t.Fatalf("expected empty list to contain nothing")
	}
}

// ─── CheckAdOwnership ─────────────────────────────────────────────────────

type fakeAdService struct {
	services.AdService
	ad  *models.Ad
	err error
}

func (f *fakeAdService) GetAdByID(_ string) (*models.Ad, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.ad, nil
}

func TestCheckAdOwnership_AdminBypassesOwnership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := CheckAdOwnership(nil, &fakeAdService{err: errors.New("must not be queried")})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/ads-management/ad-1", nil)
	c.Set("userID", "admin-1")
	c.Set("userRole", string(models.RoleAdmin))
	c.Params = gin.Params{{Key: "id", Value: "ad-1"}}

	mw(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestCheckAdOwnership_PartnerOwnsAd(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := CheckAdOwnership(nil, &fakeAdService{ad: &models.Ad{ID: "ad-1", PartnerID: "user-1"}})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/ads-management/ad-1", nil)
	c.Set("userID", "user-1")
	c.Set("userRole", string(models.RolePartner))
	c.Params = gin.Params{{Key: "id", Value: "ad-1"}}

	mw(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestCheckAdOwnership_PartnerCannotTouchForeignAd(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := CheckAdOwnership(nil, &fakeAdService{ad: &models.Ad{ID: "ad-1", PartnerID: "other-user"}})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/ads-management/ad-1", nil)
	c.Set("userID", "user-1")
	c.Set("userRole", string(models.RolePartner))
	c.Params = gin.Params{{Key: "id", Value: "ad-1"}}

	mw(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d (%s)", w.Code, w.Body.String())
	}
	if !c.IsAborted() {
		t.Fatalf("expected request to be aborted")
	}
}

func TestCheckAdOwnership_MissingAdID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := CheckAdOwnership(nil, &fakeAdService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/ads-management", nil)
	c.Set("userID", "user-1")
	c.Set("userRole", string(models.RolePartner))

	mw(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (%s)", w.Code, w.Body.String())
	}
}
