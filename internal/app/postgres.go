package app

import (
	"context"
	"log"

	"github.com/4udiwe/commnets-feed/config"
	database "github.com/4udiwe/commnets-feed/internal/database/migrations"
	"github.com/4udiwe/commnets-feed/pkg/postgres"
	"github.com/sirupsen/logrus"
)

func (app *App) initPostgres() {
	if app.cfg.Storage.Type != config.StoragePostgres {
		return
	}

	if app.cfg.Postgres.URL == "" {
		log.Fatalf("app - initPostgres - POSTGRES_URL is required for storage type 'postgres'")
	}

	logrus.Info("Connecting to PostgreSQL...")

	db, err := postgres.New(app.cfg.Postgres.URL, postgres.ConnAttempts(5))
	if err != nil {
		log.Fatalf("app - initPostgres - failed: %v", err)
	}

	app.postgres = db

	if err := database.RunMigrations(context.Background(), app.postgres.Pool); err != nil {
		logrus.Fatalf("app - Start - Migrations failed: %v", err)
	}
}
