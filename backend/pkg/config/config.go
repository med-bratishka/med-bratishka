package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Auth     AuthConfig     `json:"auth"`
	S3       S3Config       `json:"s3"`
}

type ServerConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
	SSLMode  string `json:"ssl_mode"`
	CertLoc  string `json:"cert_loc"`
}

type AuthConfig struct {
	JWTSecret              string        `json:"jwt_secret"`
	AccessTTL              time.Duration `json:"-"`
	RefreshTTL             time.Duration `json:"-"`
	TwoFactorEncryptionKey string        `json:"two_factor_encryption_key"`
	TwoFactorIssuer        string        `json:"two_factor_issuer"`
	TwoFactorChallengeTTL  time.Duration `json:"-"`
	TrustedDeviceTTL       time.Duration `json:"-"`
}

type S3Config struct {
	Endpoint        string `json:"endpoint"`
	PublicURL       string `json:"public_url"`
	Region          string `json:"region"`
	AccessKey       string `json:"access_key"`
	SecretKey       string `json:"secret_key"`
	Bucket          string `json:"bucket"`
	UseSSL          bool   `json:"use_ssl"`
	MaxUploadSizeMB int64  `json:"max_upload_size_mb"`
}

type jsonConfig struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Auth     struct {
		JWTSecret              string `json:"jwt_secret"`
		AccessTTL              string `json:"access_ttl"`
		RefreshTTL             string `json:"refresh_ttl"`
		TwoFactorEncryptionKey string `json:"two_factor_encryption_key"`
		TwoFactorIssuer        string `json:"two_factor_issuer"`
		TwoFactorChallengeTTL  string `json:"two_factor_challenge_ttl"`
		TrustedDeviceTTL       string `json:"trusted_device_ttl"`
	} `json:"auth"`
	S3 S3Config `json:"s3"`
}

func LoadConfig() *Config {
	cfg := defaultConfig()

	if cfgPath := resolveConfigPath(); cfgPath != "" {
		if loaded, ok := loadFromJSON(cfgPath); ok {
			cfg = loaded
		}
	}

	applyEnvOverrides(cfg)
	return cfg
}

func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: "8080",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "postgres",
			DBName:   "medbratishka",
			SSLMode:  "disable",
			CertLoc:  "",
		},
		Auth: AuthConfig{
			JWTSecret:              "triss-merigoldd-milashka",
			AccessTTL:              15 * time.Minute,
			RefreshTTL:             7 * 24 * time.Hour,
			TwoFactorEncryptionKey: "change-me-2fa-secret-key",
			TwoFactorIssuer:        "Medbratishka",
			TwoFactorChallengeTTL:  5 * time.Minute,
			TrustedDeviceTTL:       30 * 24 * time.Hour,
		},
		S3: S3Config{
			Endpoint:        "",
			Region:          "us-east-1",
			AccessKey:       "",
			SecretKey:       "",
			Bucket:          "",
			UseSSL:          true,
			MaxUploadSizeMB: 15,
		},
	}
}

func resolveConfigPath() string {
	if cfgPath := os.Getenv("CONFIG_PATH"); cfgPath != "" {
		return cfgPath
	}
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}
	return filepath.Join("configs", env+".json")
}

func loadFromJSON(path string) (*Config, bool) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var jc jsonConfig
	if err := json.Unmarshal(payload, &jc); err != nil {
		return nil, false
	}

	cfg := defaultConfig()
	if jc.Server.Host != "" {
		cfg.Server.Host = jc.Server.Host
	}
	if jc.Server.Port != "" {
		cfg.Server.Port = jc.Server.Port
	}
	if jc.Database.Host != "" {
		cfg.Database.Host = jc.Database.Host
	}
	if jc.Database.Port != "" {
		cfg.Database.Port = jc.Database.Port
	}
	if jc.Database.User != "" {
		cfg.Database.User = jc.Database.User
	}
	if jc.Database.Password != "" {
		cfg.Database.Password = jc.Database.Password
	}
	if jc.Database.DBName != "" {
		cfg.Database.DBName = jc.Database.DBName
	}
	if jc.Database.SSLMode != "" {
		cfg.Database.SSLMode = jc.Database.SSLMode
	}
	if jc.Database.CertLoc != "" {
		cfg.Database.CertLoc = jc.Database.CertLoc
	}
	if jc.Auth.JWTSecret != "" {
		cfg.Auth.JWTSecret = jc.Auth.JWTSecret
	}
	cfg.Auth.AccessTTL = parseDurationString(jc.Auth.AccessTTL, cfg.Auth.AccessTTL)
	cfg.Auth.RefreshTTL = parseDurationString(jc.Auth.RefreshTTL, cfg.Auth.RefreshTTL)
	if jc.Auth.TwoFactorEncryptionKey != "" {
		cfg.Auth.TwoFactorEncryptionKey = jc.Auth.TwoFactorEncryptionKey
	}
	if jc.Auth.TwoFactorIssuer != "" {
		cfg.Auth.TwoFactorIssuer = jc.Auth.TwoFactorIssuer
	}
	cfg.Auth.TwoFactorChallengeTTL = parseDurationString(jc.Auth.TwoFactorChallengeTTL, cfg.Auth.TwoFactorChallengeTTL)
	cfg.Auth.TrustedDeviceTTL = parseDurationString(jc.Auth.TrustedDeviceTTL, cfg.Auth.TrustedDeviceTTL)
	if jc.S3.Endpoint != "" {
		cfg.S3.Endpoint = jc.S3.Endpoint
	}
	if jc.S3.PublicURL != "" {
		cfg.S3.PublicURL = jc.S3.PublicURL
	}
	if jc.S3.Region != "" {
		cfg.S3.Region = jc.S3.Region
	}
	if jc.S3.AccessKey != "" {
		cfg.S3.AccessKey = jc.S3.AccessKey
	}
	if jc.S3.SecretKey != "" {
		cfg.S3.SecretKey = jc.S3.SecretKey
	}
	if jc.S3.Bucket != "" {
		cfg.S3.Bucket = jc.S3.Bucket
	}
	cfg.S3.UseSSL = jc.S3.UseSSL
	if jc.S3.MaxUploadSizeMB != 0 {
		cfg.S3.MaxUploadSizeMB = jc.S3.MaxUploadSizeMB
	}
	return cfg, true
}

func (cfg *Config) Validate() error {
	if cfg.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret is required")
	}
	if cfg.Auth.TwoFactorEncryptionKey == "" {
		return fmt.Errorf("auth.two_factor_encryption_key or TWO_FACTOR_ENCRYPTION_KEY is required")
	}
	if cfg.Auth.AccessTTL <= 0 {
		return fmt.Errorf("auth.access_ttl must be positive")
	}
	if cfg.Auth.RefreshTTL <= 0 {
		return fmt.Errorf("auth.refresh_ttl must be positive")
	}
	if cfg.Auth.TwoFactorChallengeTTL <= 0 {
		return fmt.Errorf("auth.two_factor_challenge_ttl must be positive")
	}
	if cfg.Auth.TrustedDeviceTTL <= 0 {
		return fmt.Errorf("auth.trusted_device_ttl must be positive")
	}
	return nil
}

func applyEnvOverrides(cfg *Config) {
	cfg.Server.Host = getEnv("SERVER_HOST", cfg.Server.Host)
	cfg.Server.Port = getEnv("SERVER_PORT", cfg.Server.Port)

	cfg.Database.Host = getEnv("DB_HOST", cfg.Database.Host)
	cfg.Database.Port = getEnv("DB_PORT", cfg.Database.Port)
	cfg.Database.User = getEnv("DB_USER", cfg.Database.User)
	cfg.Database.Password = getEnv("DB_PASSWORD", cfg.Database.Password)
	cfg.Database.DBName = getEnv("DB_NAME", cfg.Database.DBName)
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", cfg.Database.SSLMode)
	cfg.Database.CertLoc = getEnv("DB_CERT_LOC", cfg.Database.CertLoc)

	cfg.Auth.JWTSecret = getEnv("JWT_SECRET", cfg.Auth.JWTSecret)
	cfg.Auth.AccessTTL = parseDurationEnv("ACCESS_TTL", cfg.Auth.AccessTTL)
	cfg.Auth.RefreshTTL = parseDurationEnv("REFRESH_TTL", cfg.Auth.RefreshTTL)
	cfg.Auth.TwoFactorEncryptionKey = getEnv("TWO_FACTOR_ENCRYPTION_KEY", cfg.Auth.TwoFactorEncryptionKey)
	cfg.Auth.TwoFactorIssuer = getEnv("TWO_FACTOR_ISSUER", cfg.Auth.TwoFactorIssuer)
	cfg.Auth.TwoFactorChallengeTTL = parseDurationEnv("TWO_FACTOR_CHALLENGE_TTL", cfg.Auth.TwoFactorChallengeTTL)
	cfg.Auth.TrustedDeviceTTL = parseDurationEnv("TRUSTED_DEVICE_TTL", cfg.Auth.TrustedDeviceTTL)

	cfg.S3.Endpoint = getEnv("S3_ENDPOINT", cfg.S3.Endpoint)
	cfg.S3.PublicURL = getEnv("S3_PUBLIC_URL", cfg.S3.PublicURL)
	cfg.S3.Region = getEnv("S3_REGION", cfg.S3.Region)
	cfg.S3.AccessKey = getEnv("S3_ACCESS_KEY", cfg.S3.AccessKey)
	cfg.S3.SecretKey = getEnv("S3_SECRET_KEY", cfg.S3.SecretKey)
	cfg.S3.Bucket = getEnv("S3_BUCKET", cfg.S3.Bucket)
	cfg.S3.UseSSL = parseBoolEnv("S3_USE_SSL", cfg.S3.UseSSL)
	cfg.S3.MaxUploadSizeMB = parseInt64Env("S3_MAX_UPLOAD_SIZE_MB", cfg.S3.MaxUploadSizeMB)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDurationString(value string, defaultValue time.Duration) time.Duration {
	if value == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return d
}

func parseDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func parseBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	if value == "1" || value == "true" || value == "TRUE" || value == "yes" || value == "YES" {
		return true
	}
	if value == "0" || value == "false" || value == "FALSE" || value == "no" || value == "NO" {
		return false
	}
	return defaultValue
}

func parseInt64Env(key string, defaultValue int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return parsed
}
