package database

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func Connect(connStr string) (*sql.DB, error) {
	conn, err := sql.Open("postgres", connStr)
    
	if err != nil {
        return nil, err
    }
    if err := conn.Ping(); err != nil {
        return nil, err
    }

    errMig := runMigrations(connStr)

    if errMig != nil {
        return nil, errMig
    }
    return conn, nil
} 