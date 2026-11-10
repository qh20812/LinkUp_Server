package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Env struct {
	Port       string
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBSSL      bool
	DBName     string
	// Optional path to a PEM-encoded CA certificate for verifying the
	// MySQL server (required by managed providers like Aiven). Empty =
	// legacy behavior (no TLS in DSN).
	DBCACertPath string
	JWTSecret    string
	JWTExpiresIn int

	CloudinaryEnv     string
	FrontendURL       string
	MobileFrontendURL string

	IceServerUrls  string
	TurnServerUrl  string
	TurnUsername   string
	TurnCredential string

	MongoURI    string
	MongoDBName string

	GoogleClientIDs    []string
	GoogleClientID     string
	GoogleClientSecret string

	// VNPay — cổng thanh toán gói quảng cáo. Tất cả optional: thiếu cấu hình
	// chỉ làm endpoint thanh toán trả lỗi, không chặn server khởi động.
	VNPayTMNCode           string
	VNPayHashSecret        string
	VNPayPaymentURL        string
	VNPayReturnURL         string
	VNPayFrontendReturnURL string

	// MoMo — cổng thanh toán thay thế. Cũng optional như VNPay.
	MoMoPartnerCode string
	MoMoAccessKey   string
	MoMoSecretKey   string
	MoMoEndpoint    string
	MoMoRedirectURL string
	MoMoIPNURL      string
	// Trang web hiển thị kết quả sau khi MoMo redirect trình duyệt về.
	MoMoFrontendReturnURL string

	// Ví trả trước — nạp tiền bằng chuyển khoản ngân hàng + admin duyệt.
	// Tất cả optional: thiếu cấu hình chỉ làm endpoint tạo yêu cầu nạp trả lỗi
	// wallet.NOT_CONFIGURED, không chặn server khởi động.
	BankBIN           string
	BankAccountNo     string
	BankAccountName   string
	BankShortName     string
	TopupQRBaseURL    string
	TopupExpiresMin   int
	TopupMinAmount    int64
	TopupMaxAmount    int64
	TopupDailyLimit   int
	TopupPendingLimit int

	// Bật/tắt từng phương thức thanh toán gói. Mặc định chỉ bật ví trả trước;
	// MoMo/VNPay giữ code nhưng tắt cho tới khi cấu hình merchant.
	PaymentWalletEnabled bool
	PaymentMoMoEnabled   bool
	PaymentVNPayEnabled  bool
}

var (
	env     Env
	loaded  bool
	loadErr error
	mu      sync.RWMutex
)

func LoadEnv() error {
	mu.Lock()
	defer mu.Unlock()

	if loaded {
		return loadErr
	}

	if err := loadDotEnv(); err != nil {
		loadErr = err
		return loadErr
	}

	dbPort, err := getInt("DB_PORT")
	if err != nil {
		loadErr = err
		return loadErr
	}

	dbSSL, err := getBool("DB_SSL")
	if err != nil {
		loadErr = err
		return loadErr
	}

	jwtExpiresIn, err := getInt("JWT_EXPIRES_IN")
	if err != nil {
		loadErr = err
		return loadErr
	}

	env = Env{
		Port:         getString("PORT", "8080"),
		DBHost:       getRequiredString("DB_HOST"),
		DBPort:       dbPort,
		DBUser:       getRequiredString("DB_USER"),
		DBPassword:   getRequiredString("DB_PASSWORD"),
		DBSSL:        dbSSL,
		DBCACertPath: getString("DB_CA_CERT_PATH", ""),
		DBName:       getRequiredString("DB_NAME"),
		JWTSecret:    getRequiredString("JWT_SECRET"),
		JWTExpiresIn: jwtExpiresIn,

		CloudinaryEnv:     getRequiredString("CLOUDINARY_URL"),
		FrontendURL:       getString("VERIFY_EMAIL_URL", getString("FRONTEND_RESET_URL", "http://localhost:3000")),
		MobileFrontendURL: getString("MOBILE_VERIFY_URL", "linkupmobile://verify-email"),

		IceServerUrls:  getString("ICE_SERVER_URLS", ""),
		TurnServerUrl:  getString("TURN_SERVER_URL", ""),
		TurnUsername:   getString("TURN_USERNAME", ""),
		TurnCredential: getString("TURN_CREDENTIAL", ""),

		MongoURI:    getRequiredString("MONGO_URI"),
		MongoDBName: getString("MONGO_DB_NAME", "linkup"),

		GoogleClientIDs:    parseCSV(getString("GOOGLE_CLIENT_IDS", "")),
		GoogleClientID:     getString("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getString("GOOGLE_CLIENT_SECRET", ""),

		VNPayTMNCode:           getString("VNP_TMN_CODE", ""),
		VNPayHashSecret:        getString("VNP_HASH_SECRET", ""),
		VNPayPaymentURL:        getString("VNP_PAYMENT_URL", ""),
		VNPayReturnURL:         getString("VNP_RETURN_URL", ""),
		VNPayFrontendReturnURL: getString("VNP_FRONTEND_RETURN_URL", ""),

		MoMoPartnerCode: getString("MOMO_PARTNER_CODE", ""),
		MoMoAccessKey:   getString("MOMO_ACCESS_KEY", ""),
		MoMoSecretKey:   getString("MOMO_SECRET_KEY", ""),
		MoMoEndpoint:    getString("MOMO_ENDPOINT", ""),
		MoMoRedirectURL: getString("MOMO_REDIRECT_URL", ""),
		MoMoIPNURL:      getString("MOMO_IPN_URL", ""),
		MoMoFrontendReturnURL: getString(
			"MOMO_FRONTEND_RETURN_URL",
			getString("VNP_FRONTEND_RETURN_URL", ""),
		),

		BankBIN:           getString("BANK_BIN", ""),
		BankAccountNo:     getString("BANK_ACCOUNT_NO", ""),
		BankAccountName:   getString("BANK_ACCOUNT_NAME", ""),
		BankShortName:     getString("BANK_SHORT_NAME", ""),
		TopupQRBaseURL:    getString("TOPUP_QR_BASE_URL", "https://img.vietqr.io/image"),
		TopupExpiresMin:   getIntDefault("TOPUP_EXPIRES_MINUTES", 30),
		TopupMinAmount:    getInt64Default("TOPUP_MIN_AMOUNT", 10000),
		TopupMaxAmount:    getInt64Default("TOPUP_MAX_AMOUNT", 50000000),
		TopupDailyLimit:   getIntDefault("TOPUP_DAILY_LIMIT", 10),
		TopupPendingLimit: getIntDefault("TOPUP_PENDING_LIMIT", 5),

		PaymentWalletEnabled: getBoolDefault("PAYMENT_WALLET_ENABLED", true),
		PaymentMoMoEnabled:   getBoolDefault("PAYMENT_MOMO_ENABLED", false),
		PaymentVNPayEnabled:  getBoolDefault("PAYMENT_VNPAY_ENABLED", false),
	}

	missing := validateRequired(env)
	if len(missing) > 0 {
		loadErr = fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
		return loadErr
	}

	loaded = true
	return nil
}

func GetEnv() Env {
	mu.RLock()
	defer mu.RUnlock()
	return env
}

func loadDotEnv() error {
	path := filepath.Clean(".env")
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("open .env: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid .env line %d", lineNo)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"`)
		value = strings.TrimRight(value, "\r")

		if key == "" {
			return fmt.Errorf("invalid .env line %d: empty key", lineNo)
		}

		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("set env %s: %w", key, err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read .env: %w", err)
	}

	return nil
}

func validateRequired(e Env) []string {
	missing := make([]string, 0)
	if e.DBHost == "" {
		missing = append(missing, "DB_HOST")
	}
	if e.DBUser == "" {
		missing = append(missing, "DB_USER")
	}
	if e.DBPassword == "" {
		missing = append(missing, "DB_PASSWORD")
	}
	if e.DBName == "" {
		missing = append(missing, "DB_NAME")
	}
	if e.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	// if e.DBSSL && e.DBCACertPath == "" {
	// 	missing = append(missing, "DB_CA_CERT_PATH")
	// }
	if !e.DBSSL {
		missing = append(missing, "DB_SSL")
	}
	if e.DBPort == 0 {
		missing = append(missing, "DB_PORT")
	}
	if e.JWTExpiresIn == 0 {
		missing = append(missing, "JWT_EXPIRES_IN")
	}
	return missing
}

func getString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getRequiredString(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func getInt(key string) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return 0, fmt.Errorf("missing required env var: %s", key)
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer env var %s: %w", key, err)
	}
	return parsed, nil
}

func getBool(key string) (bool, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return false, fmt.Errorf("missing required env var: %s", key)
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid bool env var %s: %w", key, err)
	}
	return parsed, nil
}

// getIntDefault đọc biến optional; thiếu hoặc không hợp lệ thì dùng fallback
// (dùng cho cấu hình ví/nạp tiền — không được phép chặn server khởi động).
func getIntDefault(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

// getInt64Default đọc biến optional kiểu int64 (số tiền VND).
func getInt64Default(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

// getBoolDefault đọc biến optional kiểu bool; thiếu/không hợp lệ dùng fallback.
func getBoolDefault(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func parseCSV(value string) []string {
	var out []string
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
