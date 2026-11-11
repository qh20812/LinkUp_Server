package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newAuthDeviceContext(userAgent, deviceName string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	c.Request.RemoteAddr = "127.0.0.1:1234"
	if userAgent != "" {
		c.Request.Header.Set("User-Agent", userAgent)
	}
	if deviceName != "" {
		c.Request.Header.Set("X-Device-Name", deviceName)
	}
	return c
}

func TestClientDeviceInfoPrefersDeviceHeader(t *testing.T) {
	c := newAuthDeviceContext("okhttp/4.12.0", "  Pixel 7 (Android 14)\n ")
	name, _, ua := clientDeviceInfo(c)
	if name != "Pixel 7 (Android 14)" {
		t.Fatalf("device name = %q, want %q", name, "Pixel 7 (Android 14)")
	}
	if ua != "okhttp/4.12.0" {
		t.Fatalf("user agent = %q", ua)
	}
}

func TestClientDeviceInfoFallsBackToMobileUserAgent(t *testing.T) {
	c := newAuthDeviceContext("okhttp/4.12.0", "")
	name, _, _ := clientDeviceInfo(c)
	if name != "Mobile" {
		t.Fatalf("device name = %q, want Mobile", name)
	}
}

func TestClientDeviceInfoUnknownIsNotWeb(t *testing.T) {
	c := newAuthDeviceContext("CustomApp/1.0", "")
	name, _, _ := clientDeviceInfo(c)
	if name == "Web" {
		t.Fatal("unknown native client must not be labelled Web")
	}
	if name != "Thiết bị không xác định" {
		t.Fatalf("device name = %q", name)
	}
}

func TestSanitizeDeviceName(t *testing.T) {
	long := strings.Repeat("a", maxDeviceNameRunes+20)
	got := sanitizeDeviceName("  " + long + "\x00  ")
	if len([]rune(got)) != maxDeviceNameRunes {
		t.Fatalf("runes = %d, want %d", len([]rune(got)), maxDeviceNameRunes)
	}
	if strings.ContainsRune(got, '\x00') {
		t.Fatal("control characters must be removed")
	}
}
