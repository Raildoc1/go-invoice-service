package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Create(connectionString string) (*pgxpool.Pool, error) {
	if err := runMigrations(connectionString); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}
	pool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to create a connection pool: %w", err)
	}
	return pool, nil
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}
