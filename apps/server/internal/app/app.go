package app

import (
	"context"
	"fmt"

	"github.com/Pelfox/gophkeeper/apps/server/internal/config"
	"github.com/Pelfox/gophkeeper/apps/server/internal/controllers"
	"github.com/Pelfox/gophkeeper/apps/server/internal/database"
	"github.com/Pelfox/gophkeeper/apps/server/internal/middlewares"
	"github.com/Pelfox/gophkeeper/apps/server/internal/repositories"
	"github.com/Pelfox/gophkeeper/apps/server/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// StartApp configures the server application, configures routing and starts
// the HTTP server itself.
func StartApp(cfg *config.AppConfig, logger zerolog.Logger) error {
	ctx := context.Background()
	router := gin.Default()

	pool, err := database.NewPostgres(ctx, cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Creating all required repositories.
	usersRepository := repositories.NewUsersRepository(pool)
	sessionsRepository := repositories.NewSessionsRepository(pool)
	vaultsRepository := repositories.NewVaultsRepository(pool)
	keyringsRepository := repositories.NewKeyringsRepository(pool)
	vaultItemsRepository := repositories.NewVaultItemsRepository(pool)

	// Creating all required services.
	authService := services.NewAuthService(
		usersRepository,
		sessionsRepository,
		cfg.JWTSecret,
		logger,
	)
	vaultsService := services.NewVaultsService(
		usersRepository,
		vaultsRepository,
		keyringsRepository,
		logger,
	)
	vaultItemsService := services.NewVaultItemsService(
		vaultItemsRepository,
		logger,
	)

	// Grouping routes and controllers.
	authGroup := router.Group("/auth")
	controllers.NewAuthController(authService).RegisterRoutes(authGroup)

	// Grouping all routes and controllers that require auth.
	authorizedRoutesGroup := router.Group("")
	authorizedRoutesGroup.Use(middlewares.AuthMiddleware(authService))

	vaultsGroup := authorizedRoutesGroup.Group("/vaults")
	controllers.NewVaultsController(vaultsService, vaultItemsService).
		RegisterRoutes(vaultsGroup)

	return router.Run(cfg.ListenAddr)
}
