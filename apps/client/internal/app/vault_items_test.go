package app

import (
	"context"
	"errors"
	"testing"

	"github.com/Pelfox/gophkeeper/apps/client/internal/api"
	"github.com/Pelfox/gophkeeper/shared/protocol"
	"github.com/google/uuid"
)

type vaultItemClientMock struct {
	unsupportedClient

	getKeyring func(
		context.Context,
		string,
	) (*api.GetVaultsIdKeyringResponse, error)
	create func(
		context.Context,
		string,
		protocol.CreateVaultItemRequest,
	) (*api.PostVaultsIdItemsResponse, error)
	update func(
		context.Context,
		string,
		string,
		protocol.UpdateVaultItemRequest,
	) (*api.PatchVaultsIdItemsItemIdResponse, error)
	delete func(
		context.Context,
		string,
		string,
	) (*api.DeleteVaultsIdItemsItemIdResponse, error)
	list func(context.Context, string) (*api.GetVaultsIdItemsResponse, error)
}

func (m *vaultItemClientMock) GetVaultsIdKeyringWithResponse(
	ctx context.Context,
	id string,
	_ ...api.RequestEditorFn,
) (*api.GetVaultsIdKeyringResponse, error) {
	if m.getKeyring == nil {
		return nil, m.unsupportedMethod()
	}

	return m.getKeyring(ctx, id)
}

func (m *vaultItemClientMock) PostVaultsIdItemsWithResponse(
	ctx context.Context,
	id string,
	body api.PostVaultsIdItemsJSONRequestBody,
	_ ...api.RequestEditorFn,
) (*api.PostVaultsIdItemsResponse, error) {
	if m.create == nil {
		return nil, m.unsupportedMethod()
	}

	return m.create(ctx, id, body)
}

func (m *vaultItemClientMock) PatchVaultsIdItemsItemIdWithResponse(
	ctx context.Context,
	id string,
	itemID string,
	body api.PatchVaultsIdItemsItemIdJSONRequestBody,
	_ ...api.RequestEditorFn,
) (*api.PatchVaultsIdItemsItemIdResponse, error) {
	if m.update == nil {
		return nil, m.unsupportedMethod()
	}

	return m.update(ctx, id, itemID, body)
}

func (m *vaultItemClientMock) DeleteVaultsIdItemsItemIdWithResponse(
	ctx context.Context,
	id string,
	itemID string,
	_ ...api.RequestEditorFn,
) (*api.DeleteVaultsIdItemsItemIdResponse, error) {
	if m.delete == nil {
		return nil, m.unsupportedMethod()
	}

	return m.delete(ctx, id, itemID)
}

func (m *vaultItemClientMock) GetVaultsIdItemsWithResponse(
	ctx context.Context,
	id string,
	_ ...api.RequestEditorFn,
) (*api.GetVaultsIdItemsResponse, error) {
	if m.list == nil {
		return nil, m.unsupportedMethod()
	}

	return m.list(ctx, id)
}

// TestPlaintextVaultItemTypeValid verifies the supported plaintext item types.
func TestPlaintextVaultItemTypeValid(t *testing.T) {
	tests := []struct {
		name     string
		itemType PlaintextVaultItemType
		want     bool
	}{
		{name: "password", itemType: PlaintextVaultItemTypePassword, want: true},
		{name: "text note", itemType: PlaintextVaultItemTypeTextNote, want: true},
		{name: "binary file", itemType: PlaintextVaultItemTypeBinaryFile, want: true},
		{name: "bank card", itemType: PlaintextVaultItemTypeBankCard, want: true},
		{name: "unknown", itemType: PlaintextVaultItemType("unknown"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.itemType.Valid(); got != tt.want {
				t.Fatalf("Valid() = %t, want %t", got, tt.want)
			}
		})
	}
}

// TestCreateVaultItemRejectsInvalidType verifies that invalid item types are
// rejected before the client performs API calls.
func TestCreateVaultItemRejectsInvalidType(t *testing.T) {
	keyringRequested := false

	client := &vaultItemClientMock{
		getKeyring: func(
			context.Context,
			string,
		) (*api.GetVaultsIdKeyringResponse, error) {
			keyringRequested = true
			return nil, nil
		},
	}
	app := newTestApp(t, client)

	_, err := app.CreateVaultItem(
		context.Background(),
		testVault(),
		"password",
		PlaintextVaultItem{Type: PlaintextVaultItemType("unknown")},
	)
	if !errors.Is(err, ErrInvalidVaultItemType) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrInvalidVaultItemType)
	}
	if keyringRequested {
		t.Fatalf("expected keyring request not to be made")
	}
}

// TestCreateVaultItemEncryptsPlaintext verifies that item creation unlocks the
// vault and sends encrypted item bytes to the server.
func TestCreateVaultItemEncryptsPlaintext(t *testing.T) {
	vault := testVault()

	_, keyring := testKeyring(t, "vault-password")
	item := testVaultItem()

	var keyringVaultID string
	var itemVaultID string
	var request protocol.CreateVaultItemRequest

	client := &vaultItemClientMock{
		getKeyring: func(
			_ context.Context,
			vaultID string,
		) (*api.GetVaultsIdKeyringResponse, error) {
			keyringVaultID = vaultID

			return &api.GetVaultsIdKeyringResponse{
				HTTPResponse: testHTTPResponse(200),
				JSON200:      &keyring,
			}, nil
		},
		create: func(
			_ context.Context,
			vaultID string,
			createRequest protocol.CreateVaultItemRequest,
		) (*api.PostVaultsIdItemsResponse, error) {
			itemVaultID = vaultID
			request = createRequest

			return &api.PostVaultsIdItemsResponse{
				HTTPResponse: testHTTPResponse(201),
				JSON201: &protocol.CreateVaultItemResponse{
					ProtocolVaultItem: item,
				},
			}, nil
		},
	}
	app := newTestApp(t, client)

	resp, err := app.CreateVaultItem(
		context.Background(),
		vault,
		"vault-password",
		PlaintextVaultItem{
			Type: PlaintextVaultItemTypeTextNote,
			Name: "note",
			Payload: TextNotePayload{
				Text: "secret",
			},
		},
	)
	if err != nil {
		t.Fatalf("CreateVaultItem returned error: %v", err)
	}
	if resp.ID != item.ID {
		t.Fatalf("unexpected item ID: got %s want %s", resp.ID, item.ID)
	}
	if keyringVaultID != vault.ID.String() {
		t.Fatalf("unexpected keyring vault ID: %q", keyringVaultID)
	}
	if itemVaultID != vault.ID.String() {
		t.Fatalf("unexpected item vault ID: %q", itemVaultID)
	}
	if len(request.KeySalt) == 0 {
		t.Fatal("expected item key salt to be sent")
	}
	if len(request.ItemNonce) == 0 {
		t.Fatal("expected item nonce to be sent")
	}
	if len(request.Ciphertext) == 0 {
		t.Fatal("expected item ciphertext to be sent")
	}
}

// TestUpdateVaultItemEncryptsPlaintext verifies that item update sends the
// selected vault ID, item ID and newly encrypted item bytes.
func TestUpdateVaultItemEncryptsPlaintext(t *testing.T) {
	vault := testVault()

	_, keyring := testKeyring(t, "vault-password")
	item := VaultItem{ID: testVaultItem().ID, VaultID: vault.ID}

	var vaultID string
	var itemID string
	var request protocol.UpdateVaultItemRequest

	client := &vaultItemClientMock{
		getKeyring: func(
			context.Context,
			string,
		) (*api.GetVaultsIdKeyringResponse, error) {
			return &api.GetVaultsIdKeyringResponse{
				HTTPResponse: testHTTPResponse(200),
				JSON200:      &keyring,
			}, nil
		},
		update: func(
			_ context.Context,
			updateVaultID string,
			updateItemID string,
			updateRequest protocol.UpdateVaultItemRequest,
		) (*api.PatchVaultsIdItemsItemIdResponse, error) {
			vaultID = updateVaultID
			itemID = updateItemID
			request = updateRequest

			return &api.PatchVaultsIdItemsItemIdResponse{
				HTTPResponse: testHTTPResponse(200),
				JSON200: &protocol.UpdateVaultItemResponse{
					ProtocolVaultItem: testVaultItem(),
				},
			}, nil
		},
	}
	app := newTestApp(t, client)

	_, err := app.UpdateVaultItem(
		context.Background(),
		vault,
		"vault-password",
		item,
		PlaintextVaultItem{
			Type:    PlaintextVaultItemTypePassword,
			Name:    "login",
			Payload: PasswordPayload{Password: "secret"},
		},
	)
	if err != nil {
		t.Fatalf("UpdateVaultItem returned error: %v", err)
	}
	if vaultID != vault.ID.String() {
		t.Fatalf("unexpected vault ID: got %q", vaultID)
	}
	if itemID != item.ID.String() {
		t.Fatalf("unexpected item ID: got %q", itemID)
	}
	if len(request.Ciphertext) == 0 {
		t.Fatal("expected item ciphertext to be sent")
	}
}

// TestDeleteVaultItem verifies that successful item deletion accepts HTTP 204.
func TestDeleteVaultItem(t *testing.T) {
	vault := testVault()

	item := VaultItem{ID: testVaultItem().ID}
	var vaultID string
	var itemID string

	client := &vaultItemClientMock{
		delete: func(
			_ context.Context,
			deleteVaultID string,
			deleteItemID string,
		) (*api.DeleteVaultsIdItemsItemIdResponse, error) {
			vaultID = deleteVaultID
			itemID = deleteItemID

			return &api.DeleteVaultsIdItemsItemIdResponse{
				HTTPResponse: testHTTPResponse(204),
			}, nil
		},
	}
	app := newTestApp(t, client)

	err := app.DeleteVaultItem(context.Background(), vault, item)
	if err != nil {
		t.Fatalf("DeleteVaultItem returned error: %v", err)
	}
	if vaultID != vault.ID.String() {
		t.Fatalf("unexpected vault ID: got %q", vaultID)
	}
	if itemID != item.ID.String() {
		t.Fatalf("unexpected item ID: got %q", itemID)
	}
}

// TestListVaultItemsDecryptsSupportedPayloads verifies that list decrypts and
// maps every supported plaintext payload type.
func TestListVaultItemsDecryptsSupportedPayloads(t *testing.T) {
	vault := testVault()
	masterKey, keyring := testKeyring(t, "vault-password")

	encryptedItems := []protocol.ProtocolVaultItem{
		encryptedProtocolItem(t, masterKey, PlaintextVaultItem{
			Type: PlaintextVaultItemTypePassword,
			Name: "login",
			Payload: PasswordPayload{
				Password: "secret",
			},
		}),
		encryptedProtocolItem(t, masterKey, PlaintextVaultItem{
			Type: PlaintextVaultItemTypeTextNote,
			Name: "note",
			Payload: TextNotePayload{
				Text: "secret note",
			},
		}),
		encryptedProtocolItem(t, masterKey, PlaintextVaultItem{
			Type: PlaintextVaultItemTypeBinaryFile,
			Name: "file",
			Payload: BinaryFilePayload{
				Path: "secret.txt",
				Data: []byte("secret bytes"),
			},
		}),
		encryptedProtocolItem(t, masterKey, PlaintextVaultItem{
			Type: PlaintextVaultItemTypeBankCard,
			Name: "card",
			Payload: BankCardPayload{
				Number:         "4111111111111111",
				CVV:            "123",
				ExpirationDate: "12/30",
				HolderName:     "User Name",
			},
		}),
	}

	client := &vaultItemClientMock{
		getKeyring: func(
			context.Context,
			string,
		) (*api.GetVaultsIdKeyringResponse, error) {
			return &api.GetVaultsIdKeyringResponse{
				HTTPResponse: testHTTPResponse(200),
				JSON200:      &keyring,
			}, nil
		},
		list: func(
			context.Context,
			string,
		) (*api.GetVaultsIdItemsResponse, error) {
			return &api.GetVaultsIdItemsResponse{
				HTTPResponse: testHTTPResponse(200),
				JSON200:      &encryptedItems,
			}, nil
		},
	}
	app := newTestApp(t, client)

	items, err := app.ListVaultItems(
		context.Background(),
		vault,
		"vault-password",
	)
	if err != nil {
		t.Fatalf("ListVaultItems returned error: %v", err)
	}
	if len(items) != len(encryptedItems) {
		t.Fatalf("unexpected item count: got %d want %d", len(items), 4)
	}
	if _, ok := items[0].Plaintext.Payload.(PasswordPayload); !ok {
		t.Fatalf("unexpected password payload: %#v", items[0].Plaintext.Payload)
	}
	if _, ok := items[1].Plaintext.Payload.(TextNotePayload); !ok {
		t.Fatalf("unexpected text note payload: %#v", items[1].Plaintext.Payload)
	}
	if _, ok := items[2].Plaintext.Payload.(BinaryFilePayload); !ok {
		t.Fatalf("unexpected binary file payload: %#v", items[2].Plaintext.Payload)
	}
	if _, ok := items[3].Plaintext.Payload.(BankCardPayload); !ok {
		t.Fatalf("unexpected bank card payload: %#v", items[3].Plaintext.Payload)
	}
}

// TestListVaultItemsWrongPassword verifies that list fails when keyring
// decryption rejects the provided vault password.
func TestListVaultItemsWrongPassword(t *testing.T) {
	vault := testVault()
	_, keyring := testKeyring(t, "correct-password")

	client := &vaultItemClientMock{
		getKeyring: func(
			context.Context,
			string,
		) (*api.GetVaultsIdKeyringResponse, error) {
			return &api.GetVaultsIdKeyringResponse{
				HTTPResponse: testHTTPResponse(200),
				JSON200:      &keyring,
			}, nil
		},
	}
	app := newTestApp(t, client)

	_, err := app.ListVaultItems(context.Background(), vault, "wrong-password")
	if err == nil {
		t.Fatal("expected ListVaultItems to fail")
	}
}

func encryptedProtocolItem(
	t *testing.T,
	masterKey []byte,
	plaintext PlaintextVaultItem,
) protocol.ProtocolVaultItem {
	t.Helper()

	encrypted, err := encryptPlaintextVaultItem(masterKey, plaintext)
	if err != nil {
		t.Fatalf("encryptPlaintextVaultItem returned error: %v", err)
	}

	item := testVaultItem()
	item.ID = uuid.New()
	item.KeySalt = encrypted.KeySalt
	item.ItemNonce = encrypted.ItemNonce
	item.Ciphertext = encrypted.Ciphertext

	return item
}
