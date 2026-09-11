package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	Port               = "8080"
	CORSAllowedOrigins = "*"
	RequestTimeout     = 15 * time.Second
	MaxUploadBytes     = 25 << 20
	S3Region           = "auto"
)

type Config struct {
	D1AccountID  string
	D1DatabaseID string
	D1APIToken   string

	S3Bucket          string
	S3Endpoint        string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3PublicBaseURL   string
}

func Load() (*Config, error) {
	cfg := &Config{
		D1AccountID:  os.Getenv("D1_ACCOUNT_ID"),
		D1DatabaseID: os.Getenv("D1_DATABASE_ID"),
		D1APIToken:   os.Getenv("D1_API_TOKEN"),

		S3Bucket:          os.Getenv("S3_BUCKET"),
		S3Endpoint:        os.Getenv("S3_ENDPOINT"),
		S3AccessKeyID:     os.Getenv("S3_ACCESS_KEY_ID"),
		S3SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
		S3PublicBaseURL:   os.Getenv("S3_PUBLIC_BASE_URL"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	var missing []string

	missing = appendIfEmpty(missing, "D1_ACCOUNT_ID", c.D1AccountID)
	missing = appendIfEmpty(missing, "D1_DATABASE_ID", c.D1DatabaseID)
	missing = appendIfEmpty(missing, "D1_API_TOKEN", c.D1APIToken)
	missing = appendIfEmpty(missing, "S3_BUCKET", c.S3Bucket)
	missing = appendIfEmpty(missing, "S3_ENDPOINT", c.S3Endpoint)
	missing = appendIfEmpty(missing, "S3_ACCESS_KEY_ID", c.S3AccessKeyID)
	missing = appendIfEmpty(missing, "S3_SECRET_ACCESS_KEY", c.S3SecretAccessKey)
	missing = appendIfEmpty(missing, "S3_PUBLIC_BASE_URL", c.S3PublicBaseURL)

	if len(missing) > 0 {
		return fmt.Errorf("variáveis de ambiente obrigatórias ausentes: %s",
			strings.Join(missing, ", "))
	}
	return nil
}

func appendIfEmpty(dst []string, key, value string) []string {
	if strings.TrimSpace(value) == "" {
		return append(dst, key)
	}
	return dst
}
