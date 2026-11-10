package wallet_test

import (
	"net/http"
	"strings"
	"testing"

	errorsapp "linkup/errors"
	"linkup/services"
)

func TestGenerateTopupCode(t *testing.T) {
	allowed := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	seen := make(map[string]struct{}, 200)

	for i := 0; i < 200; i++ {
		code, err := services.GenerateTopupCode()
		if err != nil {
			t.Fatalf("GenerateTopupCode: %v", err)
		}
		if len(code) != 8 {
			t.Fatalf("code = %q, want 8 chars", code)
		}
		if !strings.HasPrefix(code, "LK") {
			t.Fatalf("code = %q, want prefix LK", code)
		}
		for _, ch := range code[2:] {
			if !strings.ContainsRune(allowed, ch) {
				t.Fatalf("code %q contains disallowed char %q", code, ch)
			}
		}
		if _, dup := seen[code]; dup {
			t.Fatalf("duplicate code generated in 200 tries: %q", code)
		}
		seen[code] = struct{}{}
	}
}

func TestValidateTopupAmount(t *testing.T) {
	const min, max = int64(10000), int64(50000000)

	tests := []struct {
		name    string
		amount  int64
		wantErr string
	}{
		{"valid minimum", min, ""},
		{"valid maximum", max, ""},
		{"valid middle", 200000, ""},
		{"zero", 0, errorsapp.ErrCodeWalletInvalidAmount},
		{"negative", -1000, errorsapp.ErrCodeWalletInvalidAmount},
		{"below minimum", min - 1, errorsapp.ErrCodeWalletAmountTooSmall},
		{"above maximum", max + 1, errorsapp.ErrCodeWalletAmountTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := services.ValidateTopupAmount(tt.amount, min, max)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("amount %d: unexpected error %v", tt.amount, err)
				}
				return
			}
			appErr, ok := errorsapp.IsAppError(err)
			if !ok {
				t.Fatalf("amount %d: err = %v, want AppError %s", tt.amount, err, tt.wantErr)
			}
			if appErr.Code != tt.wantErr {
				t.Fatalf("amount %d: code = %s, want %s", tt.amount, appErr.Code, tt.wantErr)
			}
		})
	}
}

func TestBuildVietQRURL(t *testing.T) {
	got := services.BuildVietQRURL(
		"https://img.vietqr.io/image/",
		"970407",
		"23712299999999",
		"HUYNH NGOC QUI",
		200000,
		"LK7F3Q2A",
	)

	for _, want := range []string{
		"https://img.vietqr.io/image/970407-23712299999999-compact2.png?",
		"amount=200000",
		"addInfo=LK7F3Q2A",
		"accountName=HUYNH%20NGOC%20QUI",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("url = %q, want it to contain %q", got, want)
		}
	}

	// Không truyền accountName/memo thì URL vẫn hợp lệ và không có tham số rỗng.
	minimal := services.BuildVietQRURL("https://img.vietqr.io/image", "970407", "123", "", 50000, "")
	if strings.Contains(minimal, "accountName=") || strings.Contains(minimal, "addInfo=") {
		t.Fatalf("minimal url = %q, should omit empty params", minimal)
	}
	if !strings.HasSuffix(minimal, "?amount=50000") {
		t.Fatalf("minimal url = %q, want only amount query", minimal)
	}
}

func TestWalletStatusCodeMapping(t *testing.T) {
	tests := []struct {
		code string
		want int
	}{
		{errorsapp.ErrCodeWalletNotConfigured, http.StatusBadRequest},
		{errorsapp.ErrCodeWalletInvalidAmount, http.StatusBadRequest},
		{errorsapp.ErrCodeWalletInsufficient, http.StatusConflict},
		{errorsapp.ErrCodeWalletTopupNotFound, http.StatusNotFound},
		{errorsapp.ErrCodeWalletTopupNotPending, http.StatusConflict},
		{errorsapp.ErrCodeWalletTopupRateLimited, http.StatusTooManyRequests},
		{errorsapp.ErrCodeWalletNotOwner, http.StatusForbidden},
	}

	for _, tt := range tests {
		if got := errorsapp.StatusCode(tt.code); got != tt.want {
			t.Errorf("StatusCode(%s) = %d, want %d", tt.code, got, tt.want)
		}
	}
}
