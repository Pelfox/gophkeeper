package main

import (
	"os"

	"github.com/Pelfox/gophkeeper/apps/server/internal/app"
	"github.com/Pelfox/gophkeeper/apps/server/internal/config"
	"github.com/rs/zerolog"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load configuration")
	}

	if err := app.StartApp(cfg, logger); err != nil {
		logger.Fatal().Err(err).Msg("failed to run application")
	}
}
