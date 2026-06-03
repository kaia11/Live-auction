package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv                  string
	AppPort                 string
	MySQLDSN                string
	RequirePersistentLedger bool
	RedisAddr               string
	RedisPassword           string
	RedisDB                 int
	JWTSecret               string
	WSAllowedOrigin         string
}

func (c Config) HTTPAddress() string {
	return ":" + c.AppPort
}

func Load() Config {
	return Config{
		AppEnv:                  getEnv("APP_ENV", "development"),
		AppPort:                 getEnv("APP_PORT", "8080"),
		MySQLDSN:                getEnv("MYSQL_DSN", ""),
		RequirePersistentLedger: getEnvAsBool("REQUIRE_PERSISTENT_LEDGER", true),
		RedisAddr:               getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword:           getEnv("REDIS_PASSWORD", ""),
		RedisDB:                 getEnvAsInt("REDIS_DB", 0),
		JWTSecret:               getEnv("JWT_SECRET", "replace_me"),
		WSAllowedOrigin:         getEnv("WS_ALLOWED_ORIGIN", "*"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvAsBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}
