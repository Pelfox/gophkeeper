package cmd

import (
	"fmt"

	"github.com/Pelfox/gophkeeper/apps/client/internal/app"
	"github.com/Pelfox/gophkeeper/apps/client/internal/terminal"
	"github.com/spf13/cobra"
)

func NewRegisterCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "register",
		Short: "Registers a new account for you.",
		RunE: func(cmd *cobra.Command, args []string) error {
			email, err := terminal.RequestUserInput("Enter your email: ")
			if err != nil {
				return fmt.Errorf("failed to get user's email: %w", err)
			}

			password, err := terminal.RequestHiddenUserInput("Enter your password: ")
			if err != nil {
				return fmt.Errorf("failed to get user's password: %w", err)
			}

			registerResponse, err := application.Register(cmd.Context(), email, password)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully registered as %s.\n", registerResponse.User.Email)
			return nil
		},
	}
}
