package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	APICookie  string
	ServerPort string
	RedisURL   string
	CacheTTL   time.Duration
}

// Load reads .env file and environment variables
func Load() (*Config, error) {
	_ = godotenv.Load()

	ttl, err := time.ParseDuration(getEnvDefault("CACHE_TTL", "5m"))
	if err != nil {
		ttl = 5 * time.Minute
	}

	cfg := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     getEnvDefault("DB_PORT", "5432"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBSSLMode:  getEnvDefault("DB_SSLMODE", "disable"),
		APICookie:  os.Getenv("API_COOKIE"),
		ServerPort: getEnvDefault("SERVER_PORT", "8080"),
		RedisURL:   os.Getenv("REDIS_URL"),
		CacheTTL:   ttl,
	}

	if cfg.DBHost == "" || cfg.DBUser == "" || cfg.DBName == "" {
		return nil, fmt.Errorf("missing required DB environment variables")
	}

	return cfg, nil
}

func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
