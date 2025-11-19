package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	LogLevel   string
	Database   DatabaseConfig
	Redis      RedisConfig
	ClickHouse ClickHouseConfig
	DNS        DNSConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type ClickHouseConfig struct {
	Addr string
	DB   string
}

type DNSConfig struct {
	Port string
}

func Load() (*Config, error) {
	_ = godotenv.Load() // Ignore error if .env not found

	return &Config{
		Port:     getEnv("PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "user"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "goflare"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		ClickHouse: ClickHouseConfig{
			Addr: getEnv("CLICKHOUSE_ADDR", "localhost:9000"),
			DB:   getEnv("CLICKHOUSE_DB", "goflare_analytics"),
		},
		DNS: DNSConfig{
			Port: getEnv("DNS_PORT", "8053"), // Default to 8053 for non-root dev
		},
	}, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}
