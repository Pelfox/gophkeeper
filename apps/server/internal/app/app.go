package app

import (
	"context"
	"fmt"

	"github.com/Pelfox/gophkeeper/apps/server/internal/config"
	"github.com/Pelfox/gophkeeper/apps/server/internal/database"
	"github.com/rs/zerolog"
)

// StartApp configures the server application, configures routing and starts
// the HTTP server itself.
func StartApp(cfg *config.AppConfig, logger zerolog.Logger) error {
	ctx := context.Background()

	_, err := database.NewPostgres(ctx, cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	return nil
}
