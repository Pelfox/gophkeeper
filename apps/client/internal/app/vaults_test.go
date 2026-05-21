package app

import (
	"context"
	"testing"

	"github.com/Pelfox/gophkeeper/apps/client/internal/api"
	"github.com/Pelfox/gophkeeper/shared/protocol"
)

type vaultClientMock struct {
	unsupportedClient

	list   func(context.Context) (*api.GetVaultsResponse, error)
	create func(
		context.Context,
		protocol.CreateVaultRequest,
	) (*api.PostVaultsResponse, error)
	update func(
		context.Context,
		string,
		protocol.UpdateVaultRequest,
	) (*api.PatchVaultsIdResponse, error)
	delete func(context.Context, string) (*api.DeleteVaultsIdResponse, error)
}

func (m *vaultClientMock) GetVaultsWithResponse(
	ctx context.Context,
	_ ...api.RequestEditorFn,
) (*api.GetVaultsResponse, error) {
	if m.list == nil {
		return nil, m.unsupportedMethod()
	}

	return m.list(ctx)
}

func (m *vaultClientMock) PostVaultsWithResponse(
	ctx context.Context,
	body api.PostVaultsJSONRequestBody,
	_ ...api.RequestEditorFn,
) (*api.PostVaultsResponse, error) {
	if m.create == nil {
		return nil, m.unsupportedMethod()
	}

	return m.create(ctx, body)
}

func (m *vaultClientMock) PatchVaultsIdWithResponse(
	ctx context.Context,
	id string,
	body api.PatchVaultsIdJSONRequestBody,
	_ ...api.RequestEditorFn,
) (*api.PatchVaultsIdResponse, error) {
	if m.update == nil {
		return nil, m.unsupportedMethod()
	}

	return m.update(ctx, id, body)
}

func (m *vaultClientMock) DeleteVaultsIdWithResponse(
	ctx context.Context,
	id string,
	_ ...api.RequestEditorFn,
) (*api.DeleteVaultsIdResponse, error) {
	if m.delete == nil {
		return nil, m.unsupportedMethod()
	}

	return m.delete(ctx, id)
}

// TestListVaults verifies that vault list responses are returned unchanged.
func TestListVaults(t *testing.T) {
	vaults := []protocol.ProtocolVault{testVault()}

	client := &vaultClientMock{
		list: func(context.Context) (*api.GetVaultsResponse, error) {
			return &api.GetVaultsResponse{
				HTTPResponse: testHTTPResponse(200),
				JSON200:      &vaults,
			}, nil
		},
	}
	app := newTestApp(t, client)

	got, err := app.ListVaults(context.Background())
	if err != nil {
		t.Fatalf("ListVaults returned error: %v", err)
	}
	if len(got) != 1 || got[0].ID != vaults[0].ID {
		t.Fatalf("unexpected vaults: got %#v want %#v", got, vaults)
	}
}

// TestCreateVaultSendsEncryptedKeyring verifies that vault creation sends
// encrypted key material, not the plaintext vault password.
func TestCreateVaultSendsEncryptedKeyring(t *testing.T) {
	vault := testVault()
	var request protocol.CreateVaultRequest

	client := &vaultClientMock{
		create: func(
			_ context.Context,
			createRequest protocol.CreateVaultRequest,
		) (*api.PostVaultsResponse, error) {
			request = createRequest

			return &api.PostVaultsResponse{
				HTTPResponse: testHTTPResponse(201),
				JSON201: &protocol.CreateVaultResponse{
					ProtocolVault: vault,
				},
			}, nil
		},
	}
	app := newTestApp(t, client)

	resp, err := app.CreateVault(context.Background(), "personal", "password")
	if err != nil {
		t.Fatalf("CreateVault returned error: %v", err)
	}
	if resp.ID != vault.ID {
		t.Fatalf("unexpected vault ID: got %s want %s", resp.ID, vault.ID)
	}
	if request.Name != "personal" {
		t.Fatalf("unexpected vault name: got %q", request.Name)
	}
	if len(request.EncryptionSalt) == 0 {
		t.Fatal("expected encryption salt to be sent")
	}
	if len(request.EncryptionNonce) == 0 {
		t.Fatal("expected encryption nonce to be sent")
	}
	if len(request.EncryptedMasterKey) == 0 {
		t.Fatal("expected encrypted master key to be sent")
	}
	if request.EncryptionKeySize == 0 {
		t.Fatal("expected encryption parameters to be sent")
	}
}

// TestUpdateVault verifies that vault update sends the selected vault ID and
// requested name.
func TestUpdateVault(t *testing.T) {
	vault := testVault()

	var vaultID string
	var request protocol.UpdateVaultRequest

	client := &vaultClientMock{
		update: func(
			_ context.Context,
			updateVaultID string,
			updateRequest protocol.UpdateVaultRequest,
		) (*api.PatchVaultsIdResponse, error) {
			vaultID = updateVaultID
			request = updateRequest

			return &api.PatchVaultsIdResponse{
				HTTPResponse: testHTTPResponse(200),
				JSON200: &protocol.UpdateVaultResponse{
					ProtocolVault: vault,
				},
			}, nil
		},
	}
	app := newTestApp(t, client)

	resp, err := app.UpdateVault(context.Background(), vault, "updated")
	if err != nil {
		t.Fatalf("UpdateVault returned error: %v", err)
	}
	if resp.ID != vault.ID {
		t.Fatalf("unexpected vault ID: got %s want %s", resp.ID, vault.ID)
	}
	if vaultID != vault.ID.String() {
		t.Fatalf("unexpected vault ID: got %q", vaultID)
	}
	if request.Name == nil || *request.Name != "updated" {
		t.Fatalf("unexpected vault name update: %#v", request)
	}
}

// TestDeleteVault verifies that successful vault deletion accepts HTTP 204.
func TestDeleteVault(t *testing.T) {
	vault := testVault()
	var vaultID string

	client := &vaultClientMock{
		delete: func(
			_ context.Context,
			deleteVaultID string,
		) (*api.DeleteVaultsIdResponse, error) {
			vaultID = deleteVaultID

			return &api.DeleteVaultsIdResponse{
				HTTPResponse: testHTTPResponse(204),
			}, nil
		},
	}
	app := newTestApp(t, client)

	err := app.DeleteVault(context.Background(), vault)
	if err != nil {
		t.Fatalf("DeleteVault returned error: %v", err)
	}
	if vaultID != vault.ID.String() {
		t.Fatalf("unexpected vault ID: got %q", vaultID)
	}
}

// TestDeleteVaultNotFound verifies that the app maps server not-found
// responses to a useful error.
func TestDeleteVaultNotFound(t *testing.T) {
	client := &vaultClientMock{
		delete: func(
			context.Context,
			string,
		) (*api.DeleteVaultsIdResponse, error) {
			return &api.DeleteVaultsIdResponse{
				HTTPResponse: testHTTPResponse(404),
				JSON404: &protocol.ProtocolError{
					Message: "not found",
				},
			}, nil
		},
	}
	app := newTestApp(t, client)

	err := app.DeleteVault(context.Background(), testVault())
	if err == nil {
		t.Fatal("expected DeleteVault to fail")
	}
}
