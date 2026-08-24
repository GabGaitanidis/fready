package config

import (
	"fmt"
	"os"
)

func RequireEnv(key string) (string, error) {
    val, ok := os.LookupEnv(key)
    if !ok || val == "" {
        return "", fmt.Errorf("missing required env var: %s", key)
    }
    return val, nil
}

func LoadDbEnvVars() (string, string, error) {
    user, erru := RequireEnv("DB_USER")
    password, errp := RequireEnv("DB_PASSWORD")

    if erru != nil || errp != nil {
        return "", "", fmt.Errorf("user or password env missing")
    }

    return user, password, nil
}