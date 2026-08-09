package services

import (
	"context"
	"errors"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

var ErrInvalidGoogleToken = errors.New("xác thực google thất bại")

// GoogleClaims là dữ liệu đã xác minh trích từ ID token của Google.
type GoogleClaims struct {
	GoogleID      string
	Email         string
	Name          string
	Picture       string
	EmailVerified bool
}

// GoogleIDTokenVerifier cho phép test bằng fake, tránh phụ thuộc network.
type GoogleIDTokenVerifier interface {
	Verify(ctx context.Context, idToken string) (*GoogleClaims, error)
}

type googleIDTokenVerifier struct {
	verifier  *oidc.IDTokenVerifier
	clientIDs []string
}

// NewGoogleIDTokenVerifier trả nil khi không có client id (feature tắt).
func NewGoogleIDTokenVerifier(ctx context.Context, clientIDs []string) (GoogleIDTokenVerifier, error) {
	if len(clientIDs) == 0 {
		return nil, nil
	}
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, err
	}
	return &googleIDTokenVerifier{
		verifier:  provider.Verifier(&oidc.Config{SkipClientIDCheck: true}),
		clientIDs: clientIDs,
	}, nil
}

func (g *googleIDTokenVerifier) Verify(ctx context.Context, idToken string) (*GoogleClaims, error) {
	raw, err := g.verifier.Verify(ctx, idToken) // kiểm tra chữ ký, issuer, exp
	if err != nil {
		return nil, ErrInvalidGoogleToken
	}

	var claims map[string]interface{}
	if err := raw.Claims(&claims); err != nil {
		return nil, ErrInvalidGoogleToken
	}

	if verified, _ := claims["email_verified"].(bool); !verified {
		return nil, ErrInvalidGoogleToken
	}

	if !googleAudienceAllowed(audFromClaims(claims), g.clientIDs) {
		return nil, ErrInvalidGoogleToken
	}

	email, _ := claims["email"].(string)
	if strings.TrimSpace(email) == "" {
		return nil, ErrInvalidGoogleToken
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return nil, ErrInvalidGoogleToken
	}

	name, _ := claims["name"].(string)
	picture, _ := claims["picture"].(string)

	return &GoogleClaims{
		GoogleID:      sub,
		Email:         strings.ToLower(strings.TrimSpace(email)),
		Name:          name,
		Picture:       picture,
		EmailVerified: true,
	}, nil
}

// audFromClaims xử lý cả aud dạng string lẫn array.
func audFromClaims(claims map[string]interface{}) string {
	switch a := claims["aud"].(type) {
	case string:
		return a
	case []string:
		if len(a) > 0 {
			return a[0]
		}
	case []interface{}:
		if len(a) > 0 {
			s, _ := a[0].(string)
			return s
		}
	}
	return ""
}

func googleAudienceAllowed(aud string, clientIDs []string) bool {
	for _, id := range clientIDs {
		if id == aud {
			return true
		}
	}
	return false
}
