package cmd

import (
	"github.com/Pelfox/gophkeeper/apps/client/internal/app"
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
	vaultCmd.AddCommand(newVaultItemCmd(application))

	return vaultCmd
}
