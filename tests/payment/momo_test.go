package payment_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"linkup/config"
	"linkup/models"
	"linkup/services"
	"linkup/utils"
)

func momoTestEnv(endpoint string) config.Env {
	return config.Env{
		MoMoPartnerCode:     "PC",
		MoMoAccessKey:       "AK",
		MoMoSecretKey:       "SECRET",
		MoMoEndpoint:        endpoint,
		MoMoRedirectURL:     "https://example.test/return",
		MoMoIPNURL:          "https://example.test/ipn",
		PaymentMoMoEnabled:  true,
	}
}

func TestMoMoSignKnownVector(t *testing.T) {
	raw := "accessKey=AK&amount=150000&extraData=&ipnUrl=https://example.test/ipn&orderId=LK1&orderInfo=Thanh toan goi&partnerCode=PC&redirectUrl=https://example.test/return&requestId=RQ1&requestType=captureWallet"
	const want = "0e7bf46d59e019a437514f2e1bdb86c7290739b73bf1230df26c31e2b13ba0a2"

	if got := utils.MoMoSign(raw, "SECRET"); got != want {
		t.Fatalf("MoMoSign = %s, want %s", got, want)
	}
}

func TestMoMoValuesFromJSONConvertsNumbers(t *testing.T) {
	body := []byte(`{"orderId":"LK1","amount":990000,"resultCode":0,"responseTime":1761900000000}`)
	values, err := utils.MoMoValuesFromJSON(body)
	if err != nil {
		t.Fatalf("MoMoValuesFromJSON returned error: %v", err)
	}
	if values["amount"] != "990000" || values["resultCode"] != "0" || values["responseTime"] != "1761900000000" {
		t.Fatalf("unexpected values: %#v", values)
	}
}

func TestCreateMoMoPaymentBuildsSignedRequest(t *testing.T) {
	var gotPayload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotPayload)

		rawData := utils.MoMoCreateRawData(
			"AK",
			strconv.FormatInt(int64(gotPayload["amount"].(float64)), 10),
			"",
			"https://example.test/ipn",
			gotPayload["orderId"].(string),
			gotPayload["orderInfo"].(string),
			"PC",
			"https://example.test/return",
			gotPayload["requestId"].(string),
			"captureWallet",
		)
		if !utils.MoMoVerify(rawData, "SECRET", gotPayload["signature"].(string)) {
			t.Errorf("create request signature is invalid")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"resultCode":0,"message":"Success","payUrl":"https://test.momo.vn/pay/abc"}`))
	}))
	defer server.Close()

	pkgRepo := &fakePackageRepository{
		pkg: &models.AdPackage{ID: "pkg-1", Name: "Starter", PriceMonthly: 990000, MaxDurationDays: 15},
	}
	payRepo := &fakePaymentRepository{}
	svc := services.NewMoMoService(payRepo, pkgRepo, momoTestEnv(server.URL))

	res, err := svc.CreateMoMoPayment("user-1", "pkg-1")
	if err != nil {
		t.Fatalf("CreateMoMoPayment returned error: %v", err)
	}
	if res.PaymentURL != "https://test.momo.vn/pay/abc" {
		t.Errorf("PaymentURL = %q", res.PaymentURL)
	}
	if res.Amount != 990000 {
		t.Errorf("Amount = %d, want 990000", res.Amount)
	}
	if payRepo.created == nil {
		t.Fatal("transaction was not created")
	}
	if payRepo.created.Gateway != "momo" || payRepo.created.RequestID == "" {
		t.Errorf("unexpected transaction: %#v", payRepo.created)
	}
	if payRepo.created.TxnRef != res.TxnRef {
		t.Errorf("txn ref mismatch: %q vs %q", payRepo.created.TxnRef, res.TxnRef)
	}
}

func TestProcessMoMoCallbackActivatesSubscription(t *testing.T) {
	pkgRepo := &fakePackageRepository{
		pkg: &models.AdPackage{ID: "pkg-1", Name: "Starter", PriceMonthly: 990000, MaxDurationDays: 15},
	}
	payRepo := &fakePaymentRepository{
		txn: &models.PaymentTransaction{
			ID:        "txn-1",
			TxnRef:    "LK1",
			RequestID: "RQ1",
			UserID:    "user-1",
			PackageID: "pkg-1",
			Amount:    990000,
			Status:    models.PaymentStatusPending,
			Gateway:   "momo",
		},
	}
	svc := services.NewMoMoService(payRepo, pkgRepo, momoTestEnv("https://example.test/create"))

	values := map[string]string{
		"partnerCode":  "PC",
		"orderId":      "LK1",
		"requestId":    "RQ1",
		"amount":       "990000",
		"orderInfo":    "Thanh toan goi quang cao Starter",
		"orderType":    "momo_wallet",
		"payType":      "qr",
		"resultCode":   "0",
		"message":      "Successful.",
		"transId":      "123456",
		"responseTime": "1761900000000",
		"extraData":    "",
	}
	values["signature"] = utils.MoMoSign(utils.MoMoCallbackRawData(
		"AK", values["amount"], values["extraData"], values["message"], values["orderId"],
		values["orderInfo"], values["orderType"], values["partnerCode"], values["payType"],
		values["requestId"], values["responseTime"], values["resultCode"], values["transId"],
	), "SECRET")

	result := svc.ProcessMoMoCallback(values)
	if !result.Valid || !result.Success || result.AlreadyConfirmed {
		t.Fatalf("unexpected result: %#v", result)
	}
	if !payRepo.completed || payRepo.completedSub == nil {
		t.Fatal("subscription was not activated")
	}
	if payRepo.completedSub.UserID != "user-1" || payRepo.completedSub.PackageID != "pkg-1" {
		t.Errorf("unexpected subscription: %#v", payRepo.completedSub)
	}
}

func TestProcessMoMoCallbackRejectsInvalidSignatureAndAmount(t *testing.T) {
	pkgRepo := &fakePackageRepository{
		pkg: &models.AdPackage{ID: "pkg-1", Name: "Starter", PriceMonthly: 990000, MaxDurationDays: 15},
	}
	payRepo := &fakePaymentRepository{
		txn: &models.PaymentTransaction{
			ID:        "txn-1",
			TxnRef:    "LK1",
			RequestID: "RQ1",
			UserID:    "user-1",
			PackageID: "pkg-1",
			Amount:    990000,
			Status:    models.PaymentStatusPending,
			Gateway:   "momo",
		},
	}
	svc := services.NewMoMoService(payRepo, pkgRepo, momoTestEnv("https://example.test/create"))

	base := map[string]string{
		"partnerCode":  "PC",
		"orderId":      "LK1",
		"requestId":    "RQ1",
		"amount":       "990000",
		"orderInfo":    "Thanh toan goi quang cao Starter",
		"orderType":    "momo_wallet",
		"payType":      "qr",
		"resultCode":   "0",
		"message":      "Successful.",
		"transId":      "123456",
		"responseTime": "1761900000000",
		"extraData":    "",
	}
	base["signature"] = "deadbeef"
	if result := svc.ProcessMoMoCallback(base); result.Valid {
		t.Fatal("callback with a bad signature must not be valid")
	}

	badAmount := map[string]string{}
	for key, value := range base {
		badAmount[key] = value
	}
	badAmount["amount"] = "1000"
	badAmount["signature"] = utils.MoMoSign(utils.MoMoCallbackRawData(
		"AK", badAmount["amount"], badAmount["extraData"], badAmount["message"], badAmount["orderId"],
		badAmount["orderInfo"], badAmount["orderType"], badAmount["partnerCode"], badAmount["payType"],
		badAmount["requestId"], badAmount["responseTime"], badAmount["resultCode"], badAmount["transId"],
	), "SECRET")

	result := svc.ProcessMoMoCallback(badAmount)
	if !result.Valid || !result.AmountMismatch || result.Success {
		t.Fatalf("unexpected result: %#v", result)
	}
	if payRepo.completed {
		t.Fatal("subscription must not be activated on amount mismatch")
	}
}
