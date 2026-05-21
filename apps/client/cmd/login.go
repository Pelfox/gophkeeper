package cmd

import (
	"fmt"

	"github.com/Pelfox/gophkeeper/apps/client/internal/app"
	"github.com/Pelfox/gophkeeper/apps/client/internal/terminal"
	"github.com/spf13/cobra"
)

func NewLoginCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Performs a login into your account.",
		RunE: func(cmd *cobra.Command, args []string) error {
			email, err := terminal.RequestUserInput("Enter your account's email: ")
			if err != nil {
				return fmt.Errorf("failed to get user's email: %w", err)
			}

			password, err := terminal.RequestHiddenUserInput("Enter your account's password: ")
			if err != nil {
				return fmt.Errorf("failed to get user's password: %w", err)
			}

			loginResponse, err := application.Login(cmd.Context(), email, password)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully logged in as %s.\n", loginResponse.User.Email)
			return nil
		},
	}
}
