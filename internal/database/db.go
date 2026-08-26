package database

import (
	"database/sql"
	"log/slog"

	_ "github.com/lib/pq"
)

func Connect(connStr string) (*sql.DB, error) {
    conn, err := sql.Open("postgres", connStr)
    if err != nil {
        slog.Error("failed to open database connection", "error", err)
        return nil, err
    }
    if err := conn.Ping(); err != nil {
        slog.Error("database ping failed", "error", err)
        return nil, err
    }
    slog.Info("database connection established")
    return conn, nil
}