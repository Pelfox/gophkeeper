package main

import (
	"fmt"
	"os"

	"github.com/Pelfox/gophkeeper/apps/client/cmd"
	"github.com/Pelfox/gophkeeper/apps/client/internal/api"
	"github.com/Pelfox/gophkeeper/apps/client/internal/app"
	"github.com/Pelfox/gophkeeper/apps/client/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "gophkeeper",
	Short:        "Secrets manager, written in Go.",
	SilenceUsage: true,
}

func main() {
	application := &app.App{}
	serverURL := "http://localhost:3000"

	rootCmd.PersistentFlags().StringVar(
		&serverURL,
		"server",
		"http://localhost:3000",
		"URL of the server.",
	)

	// After parsing persistent flags, initialize API client.
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		client, err := api.NewClientWithResponses(serverURL)
		if err != nil {
			return fmt.Errorf("failed to initialize API client: %w", err)
		}

		application.Config = cfg
		application.Client = client
		return nil
	}

	// Add all commands of the CLI.
	rootCmd.AddCommand(cmd.NewLoginCmd(application))
	rootCmd.AddCommand(cmd.NewRegisterCmd(application))

	// Finally, run the CLI.
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
