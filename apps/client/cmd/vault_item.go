package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Pelfox/gophkeeper/apps/client/internal/app"
	"github.com/Pelfox/gophkeeper/apps/client/internal/terminal"
	"github.com/aquasecurity/table"
	box "github.com/nyaosorg/go-box/v3"
	"github.com/spf13/cobra"
)

func newVaultItemCmd(application *app.App) *cobra.Command {
	itemCmd := &cobra.Command{
		Use:   "item",
		Short: "Manages vault items.",
	}

	itemCmd.AddCommand(newVaultItemCreateCmd(application))
	itemCmd.AddCommand(newVaultItemListCmd(application))
	itemCmd.AddCommand(newVaultItemViewCmd(application))
	itemCmd.AddCommand(newVaultItemUpdateCmd(application))
	itemCmd.AddCommand(newVaultItemDeleteCmd(application))

	return itemCmd
}

func newVaultItemUpdateCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Updates a vault item.",
		RunE: func(cmd *cobra.Command, args []string) error {
			selectedVault, ok, err := selectVault(cmd, application, "Select vault:")
			if err != nil || !ok {
				return err
			}

			vaultPassword, err := terminal.RequestHiddenUserInput("Enter vault password: ")
			if err != nil {
				return fmt.Errorf("failed to get vault password: %w", err)
			}

			if vaultPassword == "" {
				return fmt.Errorf("vault password cannot be empty")
			}

			vaultItems, err := application.ListVaultItems(cmd.Context(), selectedVault, vaultPassword)
			if err != nil {
				return err
			}

			selectedItem, ok, err := selectVaultItem(vaultItems)
			if err != nil || !ok {
				return err
			}

			plaintextItem, err := requestPlaintextVaultItem("Enter new item name: ")
			if err != nil {
				return err
			}

			updatedItem, err := application.UpdateVaultItem(
				cmd.Context(),
				selectedVault,
				vaultPassword,
				selectedItem,
				plaintextItem,
			)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully updated vault item %s (%q).\n", plaintextItem.Name, updatedItem.ID)
			return nil
		},
	}
}

func newVaultItemDeleteCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Deletes a vault item.",
		RunE: func(cmd *cobra.Command, args []string) error {
			selectedVault, ok, err := selectVault(cmd, application, "Select vault:")
			if err != nil || !ok {
				return err
			}

			vaultPassword, err := terminal.RequestHiddenUserInput("Enter vault password: ")
			if err != nil {
				return fmt.Errorf("failed to get vault password: %w", err)
			}

			if vaultPassword == "" {
				return fmt.Errorf("vault password cannot be empty")
			}

			vaultItems, err := application.ListVaultItems(cmd.Context(), selectedVault, vaultPassword)
			if err != nil {
				return err
			}

			selectedItem, ok, err := selectVaultItem(vaultItems)
			if err != nil || !ok {
				return err
			}

			if err := application.DeleteVaultItem(cmd.Context(), selectedVault, selectedItem); err != nil {
				return err
			}

			fmt.Printf("Successfully deleted vault item %s (%q).\n", selectedItem.Plaintext.Name, selectedItem.ID)
			return nil
		},
	}
}

func newVaultItemViewCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Views a vault item.",
		RunE: func(cmd *cobra.Command, args []string) error {
			selectedVault, ok, err := selectVault(cmd, application, "Select vault:")
			if err != nil || !ok {
				return err
			}

			vaultPassword, err := terminal.RequestHiddenUserInput("Enter vault password: ")
			if err != nil {
				return fmt.Errorf("failed to get vault password: %w", err)
			}

			if vaultPassword == "" {
				return fmt.Errorf("vault password cannot be empty")
			}

			vaultItems, err := application.ListVaultItems(cmd.Context(), selectedVault, vaultPassword)
			if err != nil {
				return err
			}

			selectedItem, ok, err := selectVaultItem(vaultItems)
			if err != nil || !ok {
				return err
			}

			return renderVaultItem(selectedItem)
		},
	}
}

func newVaultItemListCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lists vault items.",
		RunE: func(cmd *cobra.Command, args []string) error {
			selectedVault, ok, err := selectVault(cmd, application, "Select vault:")
			if err != nil || !ok {
				return err
			}

			vaultPassword, err := terminal.RequestHiddenUserInput("Enter vault password: ")
			if err != nil {
				return fmt.Errorf("failed to get vault password: %w", err)
			}

			if vaultPassword == "" {
				return fmt.Errorf("vault password cannot be empty")
			}

			vaultItems, err := application.ListVaultItems(cmd.Context(), selectedVault, vaultPassword)
			if err != nil {
				return err
			}

			if len(vaultItems) == 0 {
				fmt.Fprintln(os.Stdout, "No vault items found.")
				return nil
			}

			itemsTable := table.New(os.Stdout)
			itemsTable.SetHeaders("ID", "Name", "Type", "Details", "Updated")

			for _, vaultItem := range vaultItems {
				itemsTable.AddRow(
					vaultItem.ID.String(),
					vaultItem.Plaintext.Name,
					string(vaultItem.Plaintext.Type),
					vaultItemDetails(vaultItem.Plaintext),
					vaultItem.UpdatedAt.Format(dateTimeFormat),
				)
			}

			itemsTable.Render()
			return nil
		},
	}
}

func newVaultItemCreateCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Creates a new vault item.",
		RunE: func(cmd *cobra.Command, args []string) error {
			selectedVault, ok, err := selectVault(cmd, application, "Select vault:")
			if err != nil || !ok {
				return err
			}

			plaintextItem, err := requestPlaintextVaultItem("Enter item name: ")
			if err != nil {
				return err
			}

			vaultPassword, err := terminal.RequestHiddenUserInput("Enter vault password: ")
			if err != nil {
				return fmt.Errorf("failed to get vault password: %w", err)
			}

			if vaultPassword == "" {
				return fmt.Errorf("vault password cannot be empty")
			}

			vaultItem, err := application.CreateVaultItem(
				cmd.Context(),
				selectedVault,
				vaultPassword,
				plaintextItem,
			)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully created vault item %s (%q).\n", plaintextItem.Name, vaultItem.ID)
			return nil
		},
	}
}

func requestPlaintextVaultItem(namePrompt string) (app.PlaintextVaultItem, error) {
	itemType, err := terminal.RequestVaultItemType()
	if err != nil {
		return app.PlaintextVaultItem{}, err
	}

	name, err := terminal.RequestUserInput(namePrompt)
	if err != nil {
		return app.PlaintextVaultItem{}, fmt.Errorf("failed to get item name: %w", err)
	}

	if name == "" {
		return app.PlaintextVaultItem{}, fmt.Errorf("item name cannot be empty")
	}

	payload, err := terminal.RequestVaultItemPayload(itemType)
	if err != nil {
		return app.PlaintextVaultItem{}, err
	}

	return app.PlaintextVaultItem{
		Type:    itemType,
		Name:    name,
		Payload: payload,
	}, nil
}

func vaultItemDetails(item app.PlaintextVaultItem) string {
	switch payload := item.Payload.(type) {
	case app.PasswordPayload:
		if payload.Website == nil {
			return "website: -"
		}

		return fmt.Sprintf("website: %s", *payload.Website)
	case app.TextNotePayload:
		return fmt.Sprintf("text: %s", truncate(payload.Text, 48))
	case app.BinaryFilePayload:
		return fmt.Sprintf("file: %s (%d bytes)", payload.Path, len(payload.Data))
	case app.BankCardPayload:
		return fmt.Sprintf("card: %s, holder: %s", maskedBankCardNumber(payload.Number), payload.HolderName)
	default:
		return "-"
	}
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}

	if limit <= 3 {
		return value[:limit]
	}

	return value[:limit-3] + "..."
}

func maskedBankCardNumber(number string) string {
	if len(number) <= 4 {
		return number
	}

	return "**** " + number[len(number)-4:]
}

func selectVaultItem(vaultItems []app.VaultItem) (app.VaultItem, bool, error) {
	if len(vaultItems) == 0 {
		fmt.Fprintln(os.Stdout, "No vault items found.")
		return app.VaultItem{}, false, nil
	}

	options := make([]string, len(vaultItems))
	for i, vaultItem := range vaultItems {
		options[i] = fmt.Sprintf(
			"%s (%s, %s)",
			vaultItem.Plaintext.Name,
			vaultItem.Plaintext.Type,
			vaultItem.ID,
		)
	}

	fmt.Fprintln(os.Stdout, "Select vault item:")
	selected, err := box.SelectIndex(options, false, os.Stdout)
	if err != nil {
		return app.VaultItem{}, false, fmt.Errorf("failed to select vault item: %w", err)
	}

	if len(selected) == 0 {
		return app.VaultItem{}, false, fmt.Errorf("no vault item selected")
	}
	fmt.Fprintln(os.Stdout)

	return vaultItems[selected[0]], true, nil
}

func renderVaultItem(vaultItem app.VaultItem) error {
	fmt.Fprintf(os.Stdout, "ID: %s\n", vaultItem.ID)
	fmt.Fprintf(os.Stdout, "Name: %s\n", vaultItem.Plaintext.Name)
	fmt.Fprintf(os.Stdout, "Type: %s\n", vaultItem.Plaintext.Type)
	fmt.Fprintf(os.Stdout, "Created: %s\n", vaultItem.CreatedAt.Format(dateTimeFormat))
	fmt.Fprintf(os.Stdout, "Updated: %s\n", vaultItem.UpdatedAt.Format(dateTimeFormat))

	switch payload := vaultItem.Plaintext.Payload.(type) {
	case app.PasswordPayload:
		renderPasswordPayload(payload)
	case app.TextNotePayload:
		renderTextNotePayload(payload)
	case app.BinaryFilePayload:
		if err := renderBinaryFilePayload(payload); err != nil {
			return err
		}
	case app.BankCardPayload:
		renderBankCardPayload(payload)
	default:
		fmt.Fprintln(os.Stdout, "Payload: -")
	}

	return nil
}

func renderPasswordPayload(payload app.PasswordPayload) {
	if payload.Website != nil {
		fmt.Fprintf(os.Stdout, "Website: %s\n", *payload.Website)
	}

	fmt.Fprintf(os.Stdout, "Password: %s\n", payload.Password)
}

func renderTextNotePayload(payload app.TextNotePayload) {
	fmt.Fprintln(os.Stdout, "Text:")
	fmt.Fprintln(os.Stdout, payload.Text)
}

func renderBinaryFilePayload(payload app.BinaryFilePayload) error {
	path, err := saveBinaryPayloadToTemp(payload)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Path: %s\n", payload.Path)
	fmt.Fprintf(os.Stdout, "Size: %d bytes\n", len(payload.Data))
	fmt.Fprintf(os.Stdout, "Saved to: %s\n", path)
	fmt.Fprintln(os.Stdout, "Open the saved file from the path above.")
	return nil
}

func renderBankCardPayload(payload app.BankCardPayload) {
	fmt.Fprintf(os.Stdout, "Number: %s\n", payload.Number)
	fmt.Fprintf(os.Stdout, "CVV: %s\n", payload.CVV)
	fmt.Fprintf(os.Stdout, "Expiration date: %s\n", payload.ExpirationDate)
	fmt.Fprintf(os.Stdout, "Holder name: %s\n", payload.HolderName)
}

func saveBinaryPayloadToTemp(payload app.BinaryFilePayload) (string, error) {
	dir, err := os.MkdirTemp("", "gophkeeper-vault-item-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	name := filepath.Base(payload.Path)
	if name == "." || name == string(filepath.Separator) {
		name = "vault-item.bin"
	}

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, payload.Data, 0600); err != nil {
		return "", fmt.Errorf("failed to save binary file: %w", err)
	}

	return path, nil
}
