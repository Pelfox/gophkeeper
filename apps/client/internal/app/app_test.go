package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Pelfox/gophkeeper/apps/client/internal/api"
	"github.com/Pelfox/gophkeeper/apps/client/internal/config"
	"github.com/Pelfox/gophkeeper/apps/client/internal/crypto"
	"github.com/Pelfox/gophkeeper/shared/protocol"
	"github.com/google/uuid"
)

type unsupportedClient struct{}

func newTestApp(t *testing.T, client api.ClientWithResponsesInterface) *App {
	t.Helper()

	configRoot := t.TempDir()
	t.Setenv("HOME", configRoot)
	t.Setenv("XDG_CONFIG_HOME", configRoot)

	return &App{
		Config: &config.AppConfig{},
		Client: client,
	}
}

func testVault() protocol.ProtocolVault {
	now := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)

	return protocol.ProtocolVault{
		ID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:      "personal",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func testVaultItem() protocol.ProtocolVaultItem {
	now := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	vault := testVault()

	return protocol.ProtocolVaultItem{
		ID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		VaultID:   vault.ID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func testHTTPResponse(statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Header:     make(http.Header),
	}
}

func testKeyring(t *testing.T, vaultPassword string) (
	[]byte,
	protocol.GetVaultKeyringResponse,
) {
	t.Helper()

	result, err := crypto.CreateMasterKey([]byte(vaultPassword))
	if err != nil {
		t.Fatalf("CreateMasterKey returned error: %v", err)
	}

	vault := testVault()
	return result.MasterKey, protocol.GetVaultKeyringResponse{
		ProtocolKeyring: protocol.ProtocolKeyring{
			VaultID:               vault.ID,
			EncryptionSalt:        result.EncryptionSalt,
			EncryptionNonce:       result.EncryptionNonce,
			EncryptedMasterKey:    result.EncryptedMasterKey,
			EncryptionTimeCost:    result.EncryptionParameters.TimeCost,
			EncryptionMemoryCost:  result.EncryptionParameters.MemoryCost,
			EncryptionParallelism: uint32(result.EncryptionParameters.Parallelism),
			EncryptionKeySize:     result.EncryptionParameters.KeySize,
		},
	}
}

func (unsupportedClient) unsupportedMethod() error {
	return errors.New("unexpected API method call")
}

func (c unsupportedClient) PostAuthLoginWithBodyWithResponse(
	context.Context,
	string,
	io.Reader,
	...api.RequestEditorFn,
) (*api.PostAuthLoginResponse, error) {
	return nil, c.unsupportedMethod()
}

func (c unsupportedClient) PostAuthLoginWithResponse(
	context.Context,
	api.PostAuthLoginJSONRequestBody,
	...api.RequestEditorFn,
) (*api.PostAuthLoginResponse, error) {
	return nil, c.unsupportedMethod()
}

func (c unsupportedClient) PostAuthRegisterWithBodyWithResponse(
	context.Context,
	string,
	io.Reader,
	...api.RequestEditorFn,
) (*api.PostAuthRegisterResponse, error) {
	return nil, c.unsupportedMethod()
}

func (c unsupportedClient) PostAuthRegisterWithResponse(
	context.Context,
	api.PostAuthRegisterJSONRequestBody,
	...api.RequestEditorFn,
) (*api.PostAuthRegisterResponse, error) {
	return nil, c.unsupportedMethod()
}

func (c unsupportedClient) GetVaultsWithResponse(
	context.Context,
	...api.RequestEditorFn,
) (*api.GetVaultsResponse, error) {
	return nil, c.unsupportedMethod()
}

func (c unsupportedClient) PostVaultsWithBodyWithResponse(
	context.Context,
	string,
	io.Reader,
	...api.RequestEditorFn,
) (*api.PostVaultsResponse, error) {
	return nil, c.unsupportedMethod()
}

func (c unsupportedClient) PostVaultsWithResponse(
	context.Context,
	api.PostVaultsJSONRequestBody,
	...api.RequestEditorFn,
) (*api.PostVaultsResponse, error) {
	return nil, c.unsupportedMethod()
}

func (c unsupportedClient) DeleteVaultsIdWithResponse(
	context.Context,
	string,
	...api.RequestEditorFn,
) (*api.DeleteVaultsIdResponse, error) {
	return nil, c.unsupportedMethod()
}

func (c unsupportedClient) PatchVaultsIdWithBodyWithResponse(
	context.Context,
	string,
	string,
	io.Reader,
	...api.RequestEditorFn,
) (*api.PatchVaultsIdResponse, error) {
	return nil, c.unsupportedMethod()
}

func (c unsupportedClient) PatchVaultsIdWithResponse(
	context.Context,
	string,
	api.PatchVaultsIdJSONRequestBody,
	...api.RequestEditorFn,
) (*api.PatchVaultsIdResponse, error) {
	return nil, c.unsupportedMethod()
}

func (c unsupportedClient) GetVaultsIdItemsWithResponse(
	context.Context,
	string,
	...api.RequestEditorFn,
) (*api.GetVaultsIdItemsResponse, error) {
	return nil, c.unsupportedMethod()
}

func (c unsupportedClient) PostVaultsIdItemsWithBodyWithResponse(
	context.Context,
	string,
	string,
	io.Reader,
	...api.RequestEditorFn,
) (*api.PostVaultsIdItemsResponse, error) {
	return nil, c.unsupportedMethod()
}

func (c unsupportedClient) PostVaultsIdItemsWithResponse(
	context.Context,
	string,
	api.PostVaultsIdItemsJSONRequestBody,
	...api.RequestEditorFn,
) (*api.PostVaultsIdItemsResponse, error) {
	return nil, c.unsupportedMethod()
}

func (c unsupportedClient) DeleteVaultsIdItemsItemIdWithResponse(
	context.Context,
	string,
	string,
	...api.RequestEditorFn,
) (*api.DeleteVaultsIdItemsItemIdResponse, error) {
	return nil, c.unsupportedMethod()
}

func (c unsupportedClient) GetVaultsIdItemsItemIdWithResponse(
	context.Context,
	string,
	string,
	...api.RequestEditorFn,
) (*api.GetVaultsIdItemsItemIdResponse, error) {
	return nil, c.unsupportedMethod()
}

func (c unsupportedClient) PatchVaultsIdItemsItemIdWithBodyWithResponse(
	context.Context,
	string,
	string,
	string,
	io.Reader,
	...api.RequestEditorFn,
) (*api.PatchVaultsIdItemsItemIdResponse, error) {
	return nil, c.unsupportedMethod()
}

func (c unsupportedClient) PatchVaultsIdItemsItemIdWithResponse(
	context.Context,
	string,
	string,
	api.PatchVaultsIdItemsItemIdJSONRequestBody,
	...api.RequestEditorFn,
) (*api.PatchVaultsIdItemsItemIdResponse, error) {
	return nil, c.unsupportedMethod()
}

func (c unsupportedClient) GetVaultsIdKeyringWithResponse(
	context.Context,
	string,
	...api.RequestEditorFn,
) (*api.GetVaultsIdKeyringResponse, error) {
	return nil, c.unsupportedMethod()
}
