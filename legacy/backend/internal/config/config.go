package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"modelmarket/pkg/logger"
)

// Config 应用全局配置
type Config struct {
	ServerPort       string
	GinMode          string
	DBDriver         string
	DBDSN            string
	JWTSecret        string
	JWTExpiresHours  int
	EncryptionKey    string
	AdminUsername    string
	AdminPassword    string
	CacheTTLSeconds  int
	VendorAPIMode    string // mock | real
	LogLevel         string
}

var Global *Config

// Load 读取 .env 并构造配置
func Load() *Config {
	// 尝试加载 .env 文件，没有也不报错
	if err := godotenv.Load(); err != nil {
		logger.L().Info(".env file not found, falling back to environment variables")
	}

	cfg := &Config{
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		GinMode:          getEnv("GIN_MODE", "debug"),
		DBDriver:         getEnv("DB_DRIVER", "sqlite"),
		DBDSN:            getEnv("DB_DSN", "modelmarket.db"),
		JWTSecret:        getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTExpiresHours:  getEnvInt("JWT_EXPIRES_HOURS", 24),
		EncryptionKey:    getEnv("ENCRYPTION_KEY", ""),
		AdminUsername:    getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:    getEnv("ADMIN_PASSWORD", "admin123"),
		CacheTTLSeconds:  getEnvInt("CACHE_TTL_SECONDS", 3600),
		VendorAPIMode:    getEnv("VENDOR_API_MODE", "mock"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
	}

	Global = cfg
	logger.L().Infof("Config loaded: driver=%s mode=%s vendorMode=%s cacheTTL=%ds",
		cfg.DBDriver, cfg.GinMode, cfg.VendorAPIMode, cfg.CacheTTLSeconds)
	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
