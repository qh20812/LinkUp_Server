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
	Port         string
	DBHost       string
	DBPort       int
	DBUser       string
	DBPassword   string
	DBSSL        bool
	DBName       string
	JWTSecret    string
	JWTExpiresIn int
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
		// DBCACertPath: getString("DB_CA_CERT_PATH", ""),
		DBName:       getRequiredString("DB_NAME"),
		JWTSecret:    getRequiredString("JWT_SECRET"),
		JWTExpiresIn: jwtExpiresIn,
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
