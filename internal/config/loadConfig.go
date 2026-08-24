package config

import (
	"fmt"
	"os"
)

type Config struct {
	DB    DBConfig
	App   AppConfig
}

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
	SSLMode  string
}


type AppConfig struct {
	Port string
	Env  string
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode)
}

func RequireEnv(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return "", fmt.Errorf("missing required env var: %s", key)
	}
	return val, nil
}

func GetEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func LoadConfig() (*Config, error) {
	dbUser, err := RequireEnv("DB_USER")
	if err != nil {
		return nil, err
	}

	dbPassword, err := RequireEnv("DB_PASSWORD")
	if err != nil {
		return nil, err
	}

	dbHost, err := RequireEnv("DB_HOST")
	if err != nil {
		return nil, err
	}

	dbName, err := RequireEnv("DB_NAME")
	if err != nil {
		return nil, err
	}

	return &Config{
		DB: DBConfig{
			User:     dbUser,
			Password: dbPassword,
			Host:     dbHost,
			Port:     GetEnv("DB_PORT", "5432"),
			Name:     dbName,
			SSLMode:  GetEnv("DB_SSLMODE", "disable"),
		},
		App: AppConfig{
			Port: GetEnv("PORT", "8080"),
			Env:  GetEnv("APP_ENV", "development"),
		},
	}, nil
}