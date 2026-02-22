package main

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/rawdah/rawdah-api/internal/config"
	"github.com/rawdah/rawdah-api/internal/migrate"
	"github.com/rawdah/rawdah-api/internal/repository"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	db, err := repository.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := migrate.Run(ctx, db); err != nil {
		log.Fatal().Err(err).Msg("migration failed")
	}

	log.Info().Msg("migrations applied")
}
