package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type CloudinaryEnv struct {
	CloudName string
	APIKey    string
	APISecret string
}

type cloudinaryUsageResponse struct {
	Plan string `json:"plan"`
}

func LoadCloudinaryEnv() (CloudinaryEnv, error) {
	if err := loadDotEnv(); err != nil {
		return CloudinaryEnv{}, err
	}

	env, err := loadCloudinaryEnvFromURL()
	if err != nil {
		return CloudinaryEnv{}, err
	}

	if env.CloudName == "" {
		env.CloudName = strings.TrimSpace(os.Getenv("CLOUDINARY_CLOUD_NAME"))
	}
	if env.APIKey == "" {
		env.APIKey = strings.TrimSpace(os.Getenv("CLOUDINARY_API_KEY"))
	}
	if env.APISecret == "" {
		env.APISecret = strings.TrimSpace(os.Getenv("CLOUDINARY_API_SECRET"))
	}

	missing := make([]string, 0, 3)
	if env.CloudName == "" {
		missing = append(missing, "CLOUDINARY_CLOUD_NAME")
	}
	if env.APIKey == "" {
		missing = append(missing, "CLOUDINARY_API_KEY")
	}
	if env.APISecret == "" {
		missing = append(missing, "CLOUDINARY_API_SECRET")
	}

	if len(missing) > 0 {
		return CloudinaryEnv{}, fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}

	return env, nil
}

func loadCloudinaryEnvFromURL() (CloudinaryEnv, error) {
	raw := strings.TrimSpace(os.Getenv("CLOUDINARY_URL"))
	if raw == "" {
		return CloudinaryEnv{}, nil
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return CloudinaryEnv{}, fmt.Errorf("parse CLOUDINARY_URL: %w", err)
	}

	if parsed.Scheme != "cloudinary" {
		return CloudinaryEnv{}, fmt.Errorf("parse CLOUDINARY_URL: expected scheme cloudinary, got %q", parsed.Scheme)
	}

	user := ""
	secret := ""
	if parsed.User != nil {
		user = strings.TrimSpace(parsed.User.Username())
		secret, _ = parsed.User.Password()
		secret = strings.TrimSpace(secret)
	}

	cloudName := strings.TrimSpace(parsed.Host)
	if cloudName == "" {
		return CloudinaryEnv{}, fmt.Errorf("parse CLOUDINARY_URL: missing cloud name")
	}

	if user == "" || secret == "" {
		return CloudinaryEnv{}, fmt.Errorf("parse CLOUDINARY_URL: missing api key or api secret")
	}

	return CloudinaryEnv{
		CloudName: cloudName,
		APIKey:    user,
		APISecret: secret,
	}, nil
}

func CheckCloudinaryConnection(env CloudinaryEnv) (string, error) {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/usage", env.CloudName), nil)
	if err != nil {
		return "", fmt.Errorf("create cloudinary request: %w", err)
	}

	request.SetBasicAuth(env.APIKey, env.APISecret)

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("request cloudinary usage: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read cloudinary response: %w", err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		message := strings.TrimSpace(string(body))
		if message == "" {
			return "", fmt.Errorf("cloudinary usage returned %s", response.Status)
		}
		return "", fmt.Errorf("cloudinary usage returned %s: %s", response.Status, message)
	}

	if len(body) == 0 {
		return "", nil
	}

	var usage cloudinaryUsageResponse
	if err := json.Unmarshal(body, &usage); err != nil {
		return "", fmt.Errorf("parse cloudinary usage response: %w", err)
	}

	return usage.Plan, nil
}
