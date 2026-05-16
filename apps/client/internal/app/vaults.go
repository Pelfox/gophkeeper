package app

import (
	"context"
	"fmt"

	"github.com/Pelfox/gophkeeper/apps/client/internal/crypto"
	"github.com/Pelfox/gophkeeper/shared/protocol"
)

// ListVaults retrieves vaults available to the current user.
func (a *App) ListVaults(ctx context.Context) ([]protocol.ProtocolVault, error) {
	resp, err := a.Client.GetVaultsWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to perform vault list request: %w", err)
	}

	if resp.JSON200 != nil {
		return *resp.JSON200, nil
	}

	if resp.JSON401 != nil {
		return nil, fmt.Errorf("unauthorized: %s", resp.JSON401.Message)
	}

	if resp.JSON500 != nil {
		return nil, fmt.Errorf("server error: %s", resp.JSON500.Message)
	}

	return nil, fmt.Errorf("unexpected response from server: %s", resp.Status())
}

// CreateVault creates a new vault and stores only encrypted key material on the
// server.
func (a *App) CreateVault(
	ctx context.Context,
	name string,
	vaultPassword string,
) (*protocol.CreateVaultResponse, error) {
	masterKeyResult, err := crypto.CreateMasterKey([]byte(vaultPassword))
	if err != nil {
		return nil, fmt.Errorf("failed to create vault master key: %w", err)
	}

	resp, err := a.Client.PostVaultsWithResponse(ctx, protocol.CreateVaultRequest{
		Name:                  name,
		EncryptionSalt:        masterKeyResult.EncryptionSalt,
		EncryptionNonce:       masterKeyResult.EncryptionNonce,
		EncryptedMasterKey:    masterKeyResult.EncryptedMasterKey,
		EncryptionTimeCost:    masterKeyResult.EncryptionParameters.TimeCost,
		EncryptionMemoryCost:  masterKeyResult.EncryptionParameters.MemoryCost,
		EncryptionParallelism: uint32(masterKeyResult.EncryptionParameters.Parallelism),
		EncryptionKeySize:     masterKeyResult.EncryptionParameters.KeySize,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to perform vault creation request: %w", err)
	}

	if resp.JSON201 != nil {
		return resp.JSON201, nil
	}

	if resp.JSON400 != nil {
		return nil, fmt.Errorf("vault creation failed: %s", resp.JSON400.Message)
	}

	if resp.JSON401 != nil {
		return nil, fmt.Errorf("unauthorized: %s", resp.JSON401.Message)
	}

	if resp.JSON422 != nil {
		return nil, fmt.Errorf("invalid request: %s", resp.JSON422.Message)
	}

	if resp.JSON500 != nil {
		return nil, fmt.Errorf("server error: %s", resp.JSON500.Message)
	}

	return nil, fmt.Errorf("unexpected response from server: %s", resp.Status())
}

// UpdateVault updates metadata of an existing vault.
func (a *App) UpdateVault(
	ctx context.Context,
	vault protocol.ProtocolVault,
	name string,
) (*protocol.UpdateVaultResponse, error) {
	resp, err := a.Client.PatchVaultsIdWithResponse(ctx, vault.ID.String(), protocol.UpdateVaultRequest{
		Name: &name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to perform vault update request: %w", err)
	}

	if resp.JSON200 != nil {
		return resp.JSON200, nil
	}

	if resp.JSON400 != nil {
		return nil, fmt.Errorf("vault update failed: %s", resp.JSON400.Message)
	}

	if resp.JSON401 != nil {
		return nil, fmt.Errorf("unauthorized: %s", resp.JSON401.Message)
	}

	if resp.JSON404 != nil {
		return nil, fmt.Errorf("vault not found: %s", resp.JSON404.Message)
	}

	if resp.JSON422 != nil {
		return nil, fmt.Errorf("invalid request: %s", resp.JSON422.Message)
	}

	if resp.JSON500 != nil {
		return nil, fmt.Errorf("server error: %s", resp.JSON500.Message)
	}

	return nil, fmt.Errorf("unexpected response from server: %s", resp.Status())
}

// DeleteVault deletes an existing vault.
func (a *App) DeleteVault(ctx context.Context, vault protocol.ProtocolVault) error {
	resp, err := a.Client.DeleteVaultsIdWithResponse(ctx, vault.ID.String())
	if err != nil {
		return fmt.Errorf("failed to perform vault delete request: %w", err)
	}

	if resp.StatusCode() == 204 {
		return nil
	}

	if resp.JSON400 != nil {
		return fmt.Errorf("vault deletion failed: %s", resp.JSON400.Message)
	}

	if resp.JSON401 != nil {
		return fmt.Errorf("unauthorized: %s", resp.JSON401.Message)
	}

	if resp.JSON404 != nil {
		return fmt.Errorf("vault not found: %s", resp.JSON404.Message)
	}

	return fmt.Errorf("unexpected response from server: %s", resp.Status())
}
