package utils

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"math"
	"net"
	"net/url"
	"sort"
	"strings"
	"time"
)

// VNPayLocation is the timezone VNPay expects for vnp_CreateDate/vnp_ExpireDate.
var VNPayLocation = time.FixedZone("ICT", 7*60*60)

// VNPayTime formats a timestamp the way VNPay expects (yyyyMMddHHmmss, GMT+7).
func VNPayTime(t time.Time) string {
	return t.In(VNPayLocation).Format("20060102150405")
}

// VNPayParamsToQuery renders params in VNPay's canonical form: keys sorted
// ascending and both key and value URL-encoded.
func VNPayParamsToQuery(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var b strings.Builder
	for i, key := range keys {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(url.QueryEscape(key))
		b.WriteByte('=')
		b.WriteString(url.QueryEscape(params[key]))
	}
	return b.String()
}

// VNPaySign returns the HMAC-SHA512 signature of the canonical query string.
func VNPaySign(params map[string]string, secret string) string {
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write([]byte(VNPayParamsToQuery(params)))
	return hex.EncodeToString(mac.Sum(nil))
}

// VNPayVerify checks a vnp_SecureHash against the canonical query string.
func VNPayVerify(params map[string]string, secret, secureHash string) bool {
	if secureHash == "" {
		return false
	}
	expected := VNPaySign(params, secret)
	return hmac.Equal([]byte(expected), []byte(strings.ToLower(strings.TrimSpace(secureHash))))
}

// VNPayAmount converts a VND amount to the integer VNPay expects (VND × 100).
func VNPayAmount(amountVND float64) (int64, error) {
	if amountVND <= 0 || math.IsNaN(amountVND) || math.IsInf(amountVND, 0) {
		return 0, fmt.Errorf("số tiền không hợp lệ: %v", amountVND)
	}
	return int64(math.Round(amountVND)) * 100, nil
}

// VNPayParamsFromValues copies a callback query into a plain map, dropping the
// signature fields (they are not part of the signed payload).
func VNPayParamsFromValues(values url.Values) map[string]string {
	out := make(map[string]string, len(values))
	for key, list := range values {
		if key == "vnp_SecureHash" || key == "vnp_SecureHashType" {
			continue
		}
		if len(list) > 0 {
			out[key] = list[0]
		}
	}
	return out
}

// VNPayClientIP returns an IPv4 address VNPay accepts. IPv6/unknown addresses
// fall back to loopback because VNPay rejects non-IPv4 vnp_IpAddr values.
func VNPayClientIP(ip string) string {
	if parsed := net.ParseIP(strings.TrimSpace(ip)); parsed != nil {
		if v4 := parsed.To4(); v4 != nil {
			return v4.String()
		}
	}
	return "127.0.0.1"
}
