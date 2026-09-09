package services

import "testing"

func TestGoogleAudienceAllowed(t *testing.T) {
	ids := []string{"client-a", "client-b"}
	if !googleAudienceAllowed("client-a", ids) {
		t.Error("expected client-a to be allowed")
	}
	if googleAudienceAllowed("evil", ids) {
		t.Error("expected evil to be rejected")
	}
	if googleAudienceAllowed("", ids) {
		t.Error("expected empty aud to be rejected")
	}
	if googleAudienceAllowed("client-a", nil) {
		t.Error("expected nil client list to reject all")
	}
}