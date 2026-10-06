package db

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"os"
	"time"

	"linkup/config"

	"github.com/go-sql-driver/mysql"
)

// tlsConfigKey is the name registered via mysql.RegisterTLSConfig when a CA
// certificate is provided (managed providers like Aiven mandate TLS).
const tlsConfigKey = "aiven"

// tlsParamForEnv registers a verifying TLS config from DB_CA_CERT_PATH and
// returns the DSN suffix ("&tls=aiven"). Empty path = legacy behavior
// (no TLS suffix), so existing local setups keep working unchanged.
func tlsParamForEnv(env config.Env) (string, error) {
	if env.DBCACertPath == "" {
		return "", nil
	}
	pemBytes, err := os.ReadFile(env.DBCACertPath)
	if err != nil {
		return "", fmt.Errorf("read db ca cert: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pemBytes) {
		return "", fmt.Errorf("parse db ca cert: no valid PEM certificates")
	}
	if err := mysql.RegisterTLSConfig(tlsConfigKey, &tls.Config{
		RootCAs:    pool,
		ServerName: env.DBHost,
		MinVersion: tls.VersionTLS12,
	}); err != nil {
		return "", fmt.Errorf("register db tls config: %w", err)
	}
	return "&tls=" + tlsConfigKey, nil
}

func ConnectDb(env config.Env) (*sql.DB, error) {

	tlsParam, err := tlsParamForEnv(env)
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci%s",
		env.DBUser,
		env.DBPassword,
		env.DBHost,
		env.DBPort,
		env.DBName,
		tlsParam,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql failed: %w", err)
	}

	return db, nil
}

