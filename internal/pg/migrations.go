package pg

import (
	"fmt"

	"github.com/GlebRadaev/gofermart/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func RunMigrations(pool *pgxpool.Pool) error {
	goose.SetBaseFS(migrations.Migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}
	db := stdlib.OpenDBFromPool(pool)
	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	if err := db.Close(); err != nil {
		return fmt.Errorf("failed to close db: %w", err)
	}
	return nil
}
