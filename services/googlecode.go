package services

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// googleTokenEndpoint cho phép test ghi đè bằng httptest server.
var googleTokenEndpoint = "https://oauth2.googleapis.com/token"

var googleExchangeHTTPClient = &http.Client{Timeout: 10 * time.Second}

// ExchangeGoogleAuthCode đổi authorization code của flow auth-code
// (popup, redirect_uri=postmessage) lấy ID token của Google.
// Chỉ dùng stdlib, không thêm dependency. Lỗi trả về chung chung,
// không lộ chi tiết cho client.
func ExchangeGoogleAuthCode(ctx context.Context, code, clientID, clientSecret string) (string, error) {
	if strings.TrimSpace(code) == "" || strings.TrimSpace(clientID) == "" || strings.TrimSpace(clientSecret) == "" {
		return "", ErrInvalidGoogleToken
	}

	form := url.Values{
		"code":          {code},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {"postmessage"},
		"grant_type":    {"authorization_code"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googleTokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", ErrInvalidGoogleToken
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := googleExchangeHTTPClient.Do(req)
	if err != nil {
		return "", ErrInvalidGoogleToken
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", ErrInvalidGoogleToken
	}

	var out struct {
		IDToken string `json:"id_token"`
	}
	if err := json.Unmarshal(body, &out); err != nil || out.IDToken == "" {
		return "", ErrInvalidGoogleToken
	}
	return out.IDToken, nil
}
