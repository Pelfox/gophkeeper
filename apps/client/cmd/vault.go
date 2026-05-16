package cmd

import (
	"fmt"
	"os"

	"github.com/Pelfox/gophkeeper/apps/client/internal/app"
	"github.com/Pelfox/gophkeeper/apps/client/internal/terminal"
	"github.com/aquasecurity/table"
	box "github.com/nyaosorg/go-box/v3"
	"github.com/spf13/cobra"
)

func NewRootVaultCmd(application *app.App) *cobra.Command {
	vaultCmd := &cobra.Command{
		Use:   "vault",
		Short: "Manages your vaults.",
	}

	vaultCmd.AddCommand(newVaultNewCmd(application))
	vaultCmd.AddCommand(newVaultListCmd(application))
	vaultCmd.AddCommand(newVaultUpdateCmd(application))
	vaultCmd.AddCommand(newVaultDeleteCmd(application))

	return vaultCmd
}

func newVaultNewCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "new",
		Short: "Creates a new vault.",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := terminal.RequestUserInput("Enter vault name: ")
			if err != nil {
				return fmt.Errorf("failed to get vault name: %w", err)
			}

			if name == "" {
				return fmt.Errorf("vault name cannot be empty")
			}

			vaultPassword, err := terminal.RequestHiddenUserInput("Enter vault password: ")
			if err != nil {
				return fmt.Errorf("failed to get vault password: %w", err)
			}

			if vaultPassword == "" {
				return fmt.Errorf("vault password cannot be empty")
			}

			confirmedVaultPassword, err := terminal.RequestHiddenUserInput("Confirm vault password: ")
			if err != nil {
				return fmt.Errorf("failed to confirm vault password: %w", err)
			}

			if vaultPassword != confirmedVaultPassword {
				return fmt.Errorf("vault passwords do not match")
			}

			vault, err := application.CreateVault(cmd.Context(), name, vaultPassword)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully created vault %q.\n", vault.Name)
			return nil
		},
	}
}

func newVaultListCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lists available vaults.",
		RunE: func(cmd *cobra.Command, args []string) error {
			vaults, err := application.ListVaults(cmd.Context())
			if err != nil {
				return err
			}

			if len(vaults) == 0 {
				fmt.Fprintln(os.Stdout, "No vaults found.")
				return nil
			}

			vaultsTable := table.New(os.Stdout)
			vaultsTable.SetHeaders("ID", "Name", "Created", "Updated")

			for _, vault := range vaults {
				vaultsTable.AddRow(
					vault.ID.String(),
					vault.Name,
					vault.CreatedAt.Format("2006-01-02 15:04:05"),
					vault.UpdatedAt.Format("2006-01-02 15:04:05"),
				)
			}

			vaultsTable.Render()
			return nil
		},
	}
}

func newVaultUpdateCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Updates a vault.",
		RunE: func(cmd *cobra.Command, args []string) error {
			vaults, err := application.ListVaults(cmd.Context())
			if err != nil {
				return err
			}

			if len(vaults) == 0 {
				fmt.Fprintln(os.Stdout, "No vaults found.")
				return nil
			}

			options := make([]string, len(vaults))
			for i, vault := range vaults {
				options[i] = vault.Name
			}

			fmt.Fprintln(os.Stdout, "Select vault to update:")
			selected, err := box.SelectIndex(options, false, os.Stdout)
			if err != nil {
				return fmt.Errorf("failed to select vault: %w", err)
			}

			if len(selected) == 0 {
				return fmt.Errorf("no vault selected")
			}
			fmt.Fprintln(os.Stdout)

			selectedVault := vaults[selected[0]]
			newName, err := terminal.RequestUserInput("Enter new vault name: ")
			if err != nil {
				return fmt.Errorf("failed to get vault name: %w", err)
			}

			if newName == "" {
				return fmt.Errorf("vault name cannot be empty")
			}

			updatedVault, err := application.UpdateVault(cmd.Context(), selectedVault, newName)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully updated vault's name to %q.\n", updatedVault.Name)
			return nil
		},
	}
}

func newVaultDeleteCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Deletes a vault.",
		RunE: func(cmd *cobra.Command, args []string) error {
			vaults, err := application.ListVaults(cmd.Context())
			if err != nil {
				return err
			}

			if len(vaults) == 0 {
				fmt.Fprintln(os.Stdout, "No vaults found.")
				return nil
			}

			options := make([]string, len(vaults))
			for i, vault := range vaults {
				options[i] = vault.Name
			}

			fmt.Fprintln(os.Stdout, "Select vault to delete:")
			selected, err := box.SelectIndex(options, false, os.Stdout)
			if err != nil {
				return fmt.Errorf("failed to select vault: %w", err)
			}

			if len(selected) == 0 {
				return fmt.Errorf("no vault selected")
			}
			fmt.Fprintln(os.Stdout)

			selectedVault := vaults[selected[0]]
			if err := application.DeleteVault(cmd.Context(), selectedVault); err != nil {
				return err
			}

			fmt.Printf("Successfully deleted vault %q.\n", selectedVault.Name)
			return nil
		},
	}
}
