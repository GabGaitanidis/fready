package database

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(connStr string) error {
	m, err := migrate.New("file://internal/database/migrations", connStr)
	if err != nil {
        log.Fatal(err)
		return err
    }
    if err := m.Up(); err != nil {
        if err.Error() == "no change" {
            log.Println("no change made by migration scripts")
			return nil
        } else {
            log.Fatal(err)
			return err
        }
    }
	return nil
}