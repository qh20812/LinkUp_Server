package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func withFakeTokenEndpoint(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	old := googleTokenEndpoint
	googleTokenEndpoint = srv.URL
	t.Cleanup(func() { googleTokenEndpoint = old })
}

func TestExchangeGoogleAuthCode_Success(t *testing.T) {
	withFakeTokenEndpoint(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
			t.Errorf("unexpected content type %q", ct)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		for _, k := range []string{"code", "client_id", "client_secret", "redirect_uri", "grant_type"} {
			if r.PostFormValue(k) == "" {
				t.Errorf("missing form field %q", k)
			}
		}
		if got := r.PostFormValue("redirect_uri"); got != "postmessage" {
			t.Errorf("redirect_uri = %q, want postmessage", got)
		}
		if got := r.PostFormValue("grant_type"); got != "authorization_code" {
			t.Errorf("grant_type = %q, want authorization_code", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id_token":"fake-id-token","access_token":"x","expires_in":3599}`))
	})

	got, err := ExchangeGoogleAuthCode(context.Background(), "code123", "cid", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "fake-id-token" {
		t.Fatalf("got %q, want fake-id-token", got)
	}
}

func TestExchangeGoogleAuthCode_RejectsBadInput(t *testing.T) {
	for name, args := range map[string][3]string{
		"empty code":   {"", "cid", "secret"},
		"empty cid":    {"code", "", "secret"},
		"empty secret": {"code", "cid", ""},
	} {
		if _, err := ExchangeGoogleAuthCode(context.Background(), args[0], args[1], args[2]); err == nil {
			t.Errorf("%s: expected error, got nil", name)
		}
	}
}

func TestExchangeGoogleAuthCode_Non200Fails(t *testing.T) {
	withFakeTokenEndpoint(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_grant"}`))
	})
	if _, err := ExchangeGoogleAuthCode(context.Background(), "bad", "cid", "secret"); err == nil {
		t.Fatal("expected error on non-200, got nil")
	}
}

func TestExchangeGoogleAuthCode_MissingIDTokenFails(t *testing.T) {
	withFakeTokenEndpoint(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"x"}`))
	})
	if _, err := ExchangeGoogleAuthCode(context.Background(), "code", "cid", "secret"); err == nil {
		t.Fatal("expected error when id_token missing, got nil")
	}
}
