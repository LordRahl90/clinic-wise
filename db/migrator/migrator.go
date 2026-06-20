package migrator

import (
	"database/sql"

	"clinic-wise/db/migrations"

	"github.com/pressly/goose/v3"
)

func MigrateUp(db *sql.DB) error {
	goose.SetBaseFS(migrations.Migrations)
	if err := goose.SetDialect("mysql"); err != nil {
		return err
	}

	if err := goose.Up(db, "."); err != nil {
		return err
	}

	return nil
}

func MigrateDown(db *sql.DB) error {
	goose.SetBaseFS(migrations.Migrations)
	if err := goose.SetDialect("mysql"); err != nil {
		return err
	}

	if err := goose.Down(db, ""); err != nil {
		return err
	}

	return nil
}
