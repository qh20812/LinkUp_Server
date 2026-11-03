package payment_test

import (
	"net/url"
	"testing"

	"linkup/config"
	"linkup/services"
	"linkup/utils"
)

func TestVNPaySignKnownVector(t *testing.T) {
	params := map[string]string{
		"vnp_Amount":  "1000000",
		"vnp_Command": "pay",
		"vnp_TmnCode": "TEST",
	}
	const want = "d608fdfd10c0157b2a971e196cb80d448062f4c6b00fd5530911451df60af1a3b15a0e39d054ea356250943a3992ad69662efcec077a4c9fb679937a8c0d5979"

	if got := utils.VNPaySign(params, "SECRET"); got != want {
		t.Fatalf("VNPaySign = %s, want %s", got, want)
	}
}

func TestVNPayParamsToQuerySortsAndEncodes(t *testing.T) {
	params := map[string]string{
		"b": "2 3",
		"a": "1",
	}
	const want = "a=1&b=2+3"
	if got := utils.VNPayParamsToQuery(params); got != want {
		t.Fatalf("VNPayParamsToQuery = %q, want %q", got, want)
	}
}

func TestVNPayVerify(t *testing.T) {
	params := map[string]string{"vnp_Amount": "100", "vnp_TxnRef": "LK1"}
	signature := utils.VNPaySign(params, "SECRET")

	if !utils.VNPayVerify(params, "SECRET", signature) {
		t.Fatal("expected valid signature to pass")
	}
	if utils.VNPayVerify(params, "SECRET", "") {
		t.Fatal("empty signature must not pass")
	}

	params["vnp_Amount"] = "101"
	if utils.VNPayVerify(params, "SECRET", signature) {
		t.Fatal("tampered params must not pass")
	}
}

func TestVNPayAmount(t *testing.T) {
	got, err := utils.VNPayAmount(990000)
	if err != nil {
		t.Fatalf("VNPayAmount returned error: %v", err)
	}
	if got != 99000000 {
		t.Fatalf("VNPayAmount = %d, want 99000000", got)
	}

	if _, err := utils.VNPayAmount(0); err == nil {
		t.Fatal("zero amount must be rejected")
	}
}

func TestVNPayParamsFromValuesDropsSignature(t *testing.T) {
	values := url.Values{
		"vnp_TxnRef":         {"LK1"},
		"vnp_SecureHash":     {"abc"},
		"vnp_SecureHashType": {"HmacSHA512"},
	}

	params := utils.VNPayParamsFromValues(values)
	if len(params) != 1 || params["vnp_TxnRef"] != "LK1" {
		t.Fatalf("unexpected params: %#v", params)
	}
}

func TestVNPayClientIP(t *testing.T) {
	cases := map[string]string{
		"203.0.113.7":       "203.0.113.7",
		"2001:db8::1":       "127.0.0.1",
		"":                  "127.0.0.1",
		"not-an-ip-address": "127.0.0.1",
	}
	for input, want := range cases {
		if got := utils.VNPayClientIP(input); got != want {
			t.Errorf("VNPayClientIP(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestProcessVNPayCallbackRejectsBadSignature(t *testing.T) {
	svc := services.NewPaymentService(nil, nil, config.Env{
		VNPayTMNCode:    "TEST",
		VNPayHashSecret: "SECRET",
		VNPayPaymentURL: "https://example.test/pay",
		VNPayReturnURL:  "https://example.test/return",
	})

	values := url.Values{
		"vnp_TmnCode":      {"TEST"},
		"vnp_TxnRef":       {"LK1"},
		"vnp_Amount":       {"100"},
		"vnp_ResponseCode": {"00"},
		"vnp_SecureHash":   {"deadbeef"},
	}
	result := svc.ProcessVNPayCallback(values)
	if result.Valid {
		t.Fatal("callback with a bad signature must not be valid")
	}
}

func TestProcessVNPayCallbackRejectsWhenNotConfigured(t *testing.T) {
	svc := services.NewPaymentService(nil, nil, config.Env{})
	result := svc.ProcessVNPayCallback(url.Values{"vnp_TxnRef": {"LK1"}})
	if result.Valid {
		t.Fatal("callback must be rejected when VNPay is not configured")
	}
}
