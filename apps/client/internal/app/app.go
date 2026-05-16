package app

import (
	"github.com/Pelfox/gophkeeper/apps/client/internal/api"
	"github.com/Pelfox/gophkeeper/apps/client/internal/config"
)

// App describes the current state of the application.
type App struct {
	// Config holds the current state of application's configuration.
	Config *config.AppConfig
	// Client holds an instance of the API client that is used to call server.
	Client api.ClientWithResponsesInterface
}
