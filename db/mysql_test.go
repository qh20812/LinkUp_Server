package db

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"linkup/config"
)

// writeTestCA generates a self-signed CA, writes it as PEM, and returns the path.
// No network involved.
func writeTestCA(t *testing.T) string {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate test CA key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "linkup-test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create test CA cert: %v", err)
	}
	path := filepath.Join(t.TempDir(), "ca.pem")
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	if err := os.WriteFile(path, pemBytes, 0600); err != nil {
		t.Fatalf("write test CA cert: %v", err)
	}
	return path
}

func TestTlsParamForEnvEmpty(t *testing.T) {
	param, err := tlsParamForEnv(config.Env{})
	if err != nil {
		t.Fatalf("expected no error for empty CA path, got: %v", err)
	}
	if param != "" {
		t.Fatalf("expected empty TLS param for legacy behavior, got %q", param)
	}
}

func TestTlsParamForEnvMissingFile(t *testing.T) {
	env := config.Env{
		DBHost:       "mysql-test.aivencloud.com",
		DBCACertPath: filepath.Join(t.TempDir(), "does-not-exist.pem"),
	}
	if _, err := tlsParamForEnv(env); err == nil {
		t.Fatal("expected error for missing CA file, got nil")
	}
}

func TestTlsParamForEnvBadPEM(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ca.pem")
	if err := os.WriteFile(path, []byte("not-a-certificate"), 0600); err != nil {
		t.Fatalf("write bad PEM: %v", err)
	}
	env := config.Env{DBHost: "mysql-test.aivencloud.com", DBCACertPath: path}
	if _, err := tlsParamForEnv(env); err == nil {
		t.Fatal("expected error for invalid PEM, got nil")
	}
}

func TestTlsParamForEnvValidCA(t *testing.T) {
	env := config.Env{
		DBHost:       "mysql-test.aivencloud.com",
		DBCACertPath: writeTestCA(t),
	}
	param, err := tlsParamForEnv(env)
	if err != nil {
		t.Fatalf("expected no error for valid CA, got: %v", err)
	}
	if param != "&tls="+tlsConfigKey {
		t.Fatalf("unexpected TLS param: %q", param)
	}
	// Re-registering (e.g. reconnects, seed + app sharing the process)
	// must stay idempotent.
	if _, err := tlsParamForEnv(env); err != nil {
		t.Fatalf("expected re-registration to succeed, got: %v", err)
	}
}
