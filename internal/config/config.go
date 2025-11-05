// internal/config/config.go
package config

import (
	"fmt"
	"os"
)

type DBConfig struct {
	Host string
	Port string
	User string
	Pass string
	Name string
	SSL  string
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=%s TimeZone=UTC",
		c.User, c.Pass, c.Host, c.Port, c.Name, c.SSL,
	)
}

type AppConfig struct {
	Env       string
	DB        DBConfig
	JWTSecret string
}

func Load() AppConfig {
	return AppConfig{
		Env:       getenv("APP_ENV", "development"),
		JWTSecret: getenv("JWT_SECRET", "default-secret-change-in-production"),
		DB: DBConfig{
			Host: getenv("DB_HOST", "localhost"),
			Port: getenv("DB_PORT", "5432"),
			User: getenv("DB_USER", "postgres"),
			Pass: getenv("DB_PASS", "postgres"),
			Name: getenv("DB_NAME", "yourapp"),
			SSL:  getenv("DB_SSLMODE", "disable"),
		},
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
