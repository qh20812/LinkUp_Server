package auth_test

import (
	"context"
	"testing"

	"linkup/services"
)

func TestGoogleVerifier_DisabledWhenNoClientIDs(t *testing.T) {
	v, err := services.NewGoogleIDTokenVerifier(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != nil {
		t.Fatal("expected nil verifier when GOOGLE_CLIENT_IDS is empty")
	}
}

func TestGoogleAuth_SentinelExported(t *testing.T) {
	if services.ErrInvalidGoogleToken == nil {
		t.Fatal("ErrInvalidGoogleToken must be defined")
	}
}