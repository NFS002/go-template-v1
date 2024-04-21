package db

import (
	"database/sql"
	"errors"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func MigrateUp() error {
	postgresqlUrl, exists := os.LookupEnv("POSTGRESQL_URL")
	if !exists {
		return errors.New("environent variable 'POSTGRESQL_URL' is not set but required")
	}
	db, err := sql.Open("postgres", postgresqlUrl)
	if err != nil {
		return err
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"auth", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil

}
