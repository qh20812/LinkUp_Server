package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// MoMoSign returns the HMAC-SHA256 signature MoMo expects for a raw data string.
func MoMoSign(rawData, secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(rawData))
	return hex.EncodeToString(mac.Sum(nil))
}

// MoMoVerify checks a MoMo signature against the raw data string.
func MoMoVerify(rawData, secretKey, signature string) bool {
	if signature == "" {
		return false
	}
	expected := MoMoSign(rawData, secretKey)
	return hmac.Equal([]byte(expected), []byte(strings.ToLower(strings.TrimSpace(signature))))
}

// MoMoCreateRawData builds the exact raw string used to sign a create-payment
// request (requestType = captureWallet).
func MoMoCreateRawData(accessKey, amount, extraData, ipnURL, orderID, orderInfo, partnerCode, redirectURL, requestID, requestType string) string {
	return fmt.Sprintf(
		"accessKey=%s&amount=%s&extraData=%s&ipnUrl=%s&orderId=%s&orderInfo=%s&partnerCode=%s&redirectUrl=%s&requestId=%s&requestType=%s",
		accessKey, amount, extraData, ipnURL, orderID, orderInfo, partnerCode, redirectURL, requestID, requestType,
	)
}

// MoMoCallbackRawData builds the raw string used to sign IPN/redirect callbacks.
func MoMoCallbackRawData(accessKey, amount, extraData, message, orderID, orderInfo, orderType, partnerCode, payType, requestID, responseTime, resultCode, transID string) string {
	return fmt.Sprintf(
		"accessKey=%s&amount=%s&extraData=%s&message=%s&orderId=%s&orderInfo=%s&orderType=%s&partnerCode=%s&payType=%s&requestId=%s&responseTime=%s&resultCode=%s&transId=%s",
		accessKey, amount, extraData, message, orderID, orderInfo, orderType, partnerCode, payType, requestID, responseTime, resultCode, transID,
	)
}

// MoMoValuesFromJSON converts a MoMo IPN JSON object into string values.
// MoMo sends numeric fields (amount, resultCode, responseTime, transId) as
// JSON numbers, so decoding straight into map[string]string would fail.
func MoMoValuesFromJSON(body []byte) (map[string]string, error) {
	var raw map[string]any
	decoder := json.NewDecoder(bytes.NewReader(body))
	// UseNumber giữ nguyên chữ số của transId/responseTime — decode mặc định
	// sang float64 có thể làm tròn số lớn và làm sai chữ ký MoMo.
	decoder.UseNumber()
	if err := decoder.Decode(&raw); err != nil {
		return nil, err
	}

	out := make(map[string]string, len(raw))
	for key, value := range raw {
		switch v := value.(type) {
		case string:
			out[key] = v
		case json.Number:
			out[key] = v.String()
		case float64:
			out[key] = strconv.FormatInt(int64(v), 10)
		case bool:
			out[key] = strconv.FormatBool(v)
		case nil:
			out[key] = ""
		default:
			out[key] = fmt.Sprint(v)
		}
	}
	return out, nil
}

// ValuesToMap copies a url.Values into a plain map using the first value.
func ValuesToMap(values map[string][]string) map[string]string {
	out := make(map[string]string, len(values))
	for key, list := range values {
		if len(list) > 0 {
			out[key] = list[0]
		}
	}
	return out
}
