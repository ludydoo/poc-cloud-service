package store

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	pgxDriver "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Migrate(ctx context.Context, connString string) error {
	src, err := iofs.New(migrations, "migrations")
	if err != nil {
		return err
	}

	db, err := sql.Open("pgx", connString)
	if err != nil {
		return err
	}

	driver, err := pgxDriver.WithInstance(db, &pgxDriver.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil

}
