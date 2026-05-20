package repository

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"

	"synaply/internal/config"
)

type Postgres struct {
	db *sqlx.DB
}

func NewPostgres(cfg config.DB) (*Postgres, error) {
	db, err := sqlx.Connect("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("error connecting to postgres: %w", err)
	}

	db.SetMaxOpenConns(30)
	db.SetMaxIdleConns(30)
	db.SetConnMaxLifetime(5 * time.Minute)

	p := &Postgres{db: db}

	if err := p.runMigratopns(cfg.MigrationsPath); err != nil {
		return nil, fmt.Errorf("error running migrations: %w", err)
	}
	return p, nil
}

func (p *Postgres) Close() {
	err := p.db.Close()
	if err != nil {
		log.Fatalf("error closing postgres connection: %s", err)
	}
}

func (p *Postgres) runMigratopns(path string) error {
	driver, err := postgres.WithInstance(p.db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("error creating postgres driver: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(path, "postgres", driver)
	if err != nil {
		return fmt.Errorf("error creating migrate instance: %w", err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error running migrations: %w", err)
	}
	return nil
}
