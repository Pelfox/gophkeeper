package controllers

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/Pelfox/gophkeeper/apps/server/internal/services"
	"github.com/Pelfox/gophkeeper/shared/protocol"
	"github.com/google/uuid"
)

// TestVaultsControllerCreate verifies that create maps protocol keyring
// material to the vault service and returns the created vault.
func TestVaultsControllerCreate(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()

	vaultsService := &fakeVaultsService{
		createResult: testVaultResult(userID, vaultID),
	}
	router := newVaultsRouter(
		vaultsService,
		&fakeVaultItemsService{},
		testSession(userID),
	)
	request := testCreateVaultRequest()

	recorder := performJSONRequest(
		router,
		http.MethodPost,
		"/vaults/",
		request,
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusCreated,
		)
	}
	if vaultsService.createCalls != 1 ||
		vaultsService.createInput.OwnerID != userID ||
		vaultsService.createInput.Name != request.Name ||
		!bytes.Equal(
			vaultsService.createInput.EncryptionSalt,
			request.EncryptionSalt,
		) {
		t.Fatalf("unexpected create input: %#v", vaultsService.createInput)
	}

	response := decodeResponse[protocol.CreateVaultResponse](t, recorder)
	if response.ID != vaultID || response.Name != "Personal" {
		t.Fatalf("unexpected response: %#v", response)
	}
}

// TestVaultsControllerList verifies that list returns vaults for the session
// user.
func TestVaultsControllerList(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()

	vaultsService := &fakeVaultsService{
		getForUserItems: []services.VaultResult{
			*testVaultResult(userID, vaultID),
		},
	}
	router := newVaultsRouter(
		vaultsService,
		&fakeVaultItemsService{},
		testSession(userID),
	)

	recorder := performJSONRequest(router, http.MethodGet, "/vaults/", nil)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusOK,
		)
	}
	if vaultsService.getForUserCalls != 1 ||
		vaultsService.getForUserID != userID {
		t.Fatalf("unexpected GetForUser call: %#v", vaultsService)
	}

	response := decodeResponse[protocol.ListVaultsResponse](t, recorder)
	if len(response) != 1 || response[0].ID != vaultID {
		t.Fatalf("unexpected response: %#v", response)
	}
}

// TestVaultsControllerGetKeyring verifies that getKeyring returns encrypted
// keyring material for a valid vault ID.
func TestVaultsControllerGetKeyring(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()

	vaultsService := &fakeVaultsService{
		getKeyringItem: testKeyringResult(vaultID),
	}
	router := newVaultsRouter(
		vaultsService,
		&fakeVaultItemsService{},
		testSession(userID),
	)

	recorder := performJSONRequest(
		router,
		http.MethodGet,
		"/vaults/"+vaultID.String()+"/keyring",
		nil,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusOK,
		)
	}
	if vaultsService.getKeyringCalls != 1 ||
		vaultsService.getKeyringUser != userID ||
		vaultsService.getKeyringVault != vaultID {
		t.Fatalf("unexpected GetKeyring call: %#v", vaultsService)
	}

	response := decodeResponse[protocol.GetVaultKeyringResponse](t, recorder)
	if response.VaultID != vaultID ||
		!bytes.Equal(response.EncryptedMasterKey, []byte{7, 8, 9}) {
		t.Fatalf("unexpected response: %#v", response)
	}
}

// TestVaultsControllerGetKeyringNotFound verifies that missing keyrings are
// mapped to vault-not-found responses.
func TestVaultsControllerGetKeyringNotFound(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()

	vaultsService := &fakeVaultsService{
		getKeyringErr: services.ErrVaultNotFound,
	}
	router := newVaultsRouter(
		vaultsService,
		&fakeVaultItemsService{},
		testSession(userID),
	)

	recorder := performJSONRequest(
		router,
		http.MethodGet,
		"/vaults/"+vaultID.String()+"/keyring",
		nil,
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusNotFound,
		)
	}

	response := decodeProtocolError(t, recorder)
	if response.Code != protocol.ProtocolErrorVaultNotFound {
		t.Fatalf(
			"unexpected error code: got %s want %s",
			response.Code,
			protocol.ProtocolErrorVaultNotFound,
		)
	}
}

// TestVaultsControllerUpdate verifies that update maps a valid request to the
// vault service and returns the updated vault.
func TestVaultsControllerUpdate(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()
	renamed := "Work"

	vaultsService := &fakeVaultsService{
		updateItem: testVaultResult(userID, vaultID),
	}
	router := newVaultsRouter(
		vaultsService,
		&fakeVaultItemsService{},
		testSession(userID),
	)

	recorder := performJSONRequest(
		router,
		http.MethodPatch,
		"/vaults/"+vaultID.String(),
		protocol.UpdateVaultRequest{Name: &renamed},
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusOK,
		)
	}
	if vaultsService.updateCalls != 1 ||
		vaultsService.updateVault != vaultID ||
		vaultsService.updateUser != userID ||
		vaultsService.updateInput.Name == nil ||
		*vaultsService.updateInput.Name != renamed {
		t.Fatalf("unexpected update input: %#v", vaultsService)
	}
}

// TestVaultsControllerUpdateNotFound verifies that update maps missing vaults
// to not-found responses.
func TestVaultsControllerUpdateNotFound(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()
	renamed := "Work"

	vaultsService := &fakeVaultsService{
		updateErr: services.ErrVaultNotFound,
	}
	router := newVaultsRouter(
		vaultsService,
		&fakeVaultItemsService{},
		testSession(userID),
	)

	recorder := performJSONRequest(
		router,
		http.MethodPatch,
		"/vaults/"+vaultID.String(),
		protocol.UpdateVaultRequest{Name: &renamed},
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusNotFound,
		)
	}

	response := decodeProtocolError(t, recorder)
	if response.Code != protocol.ProtocolErrorVaultNotFound {
		t.Fatalf(
			"unexpected error code: got %s want %s",
			response.Code,
			protocol.ProtocolErrorVaultNotFound,
		)
	}
}

// TestVaultsControllerDelete verifies that delete maps valid vault IDs to the
// vault service and returns no content.
func TestVaultsControllerDelete(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()
	vaultsService := &fakeVaultsService{}
	router := newVaultsRouter(
		vaultsService,
		&fakeVaultItemsService{},
		testSession(userID),
	)

	recorder := performJSONRequest(
		router,
		http.MethodDelete,
		"/vaults/"+vaultID.String(),
		nil,
	)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusNoContent,
		)
	}
	if vaultsService.deleteCalls != 1 ||
		vaultsService.deleteVault != vaultID ||
		vaultsService.deleteUser != userID {
		t.Fatalf("unexpected delete call: %#v", vaultsService)
	}
}

// TestVaultsControllerDeleteNotFound verifies that delete maps missing vaults
// to not-found responses.
func TestVaultsControllerDeleteNotFound(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()

	vaultsService := &fakeVaultsService{
		deleteErr: services.ErrVaultNotFound,
	}
	router := newVaultsRouter(
		vaultsService,
		&fakeVaultItemsService{},
		testSession(userID),
	)

	recorder := performJSONRequest(
		router,
		http.MethodDelete,
		"/vaults/"+vaultID.String(),
		nil,
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusNotFound,
		)
	}

	response := decodeProtocolError(t, recorder)
	if response.Code != protocol.ProtocolErrorVaultNotFound {
		t.Fatalf(
			"unexpected error code: got %s want %s",
			response.Code,
			protocol.ProtocolErrorVaultNotFound,
		)
	}
}

// TestVaultsControllerCreateItem verifies that createItem maps encrypted item
// material to the vault item service.
func TestVaultsControllerCreateItem(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()
	itemID := uuid.New()

	itemsService := &fakeVaultItemsService{
		createItem: testVaultItemResult(vaultID, itemID),
	}
	router := newVaultsRouter(
		&fakeVaultsService{},
		itemsService,
		testSession(userID),
	)
	request := testVaultItemRequest()

	recorder := performJSONRequest(
		router,
		http.MethodPost,
		"/vaults/"+vaultID.String()+"/items",
		request,
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusCreated,
		)
	}
	if itemsService.createCalls != 1 ||
		itemsService.createUser != userID ||
		itemsService.createVault != vaultID ||
		!bytes.Equal(itemsService.createInput.Ciphertext, request.Ciphertext) {
		t.Fatalf("unexpected create item call: %#v", itemsService)
	}

	response := decodeResponse[protocol.CreateVaultItemResponse](t, recorder)
	if response.ID != itemID || response.VaultID != vaultID {
		t.Fatalf("unexpected response: %#v", response)
	}
}

// TestVaultsControllerListItems verifies that listItems returns encrypted
// items for a vault.
func TestVaultsControllerListItems(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()
	itemID := uuid.New()

	itemsService := &fakeVaultItemsService{
		getForVaultItems: []services.VaultItemResult{
			*testVaultItemResult(vaultID, itemID),
		},
	}
	router := newVaultsRouter(
		&fakeVaultsService{},
		itemsService,
		testSession(userID),
	)

	recorder := performJSONRequest(
		router,
		http.MethodGet,
		"/vaults/"+vaultID.String()+"/items",
		nil,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusOK,
		)
	}
	if itemsService.getForVaultCalls != 1 ||
		itemsService.getForVaultUser != userID ||
		itemsService.getForVaultID != vaultID {
		t.Fatalf("unexpected list items call: %#v", itemsService)
	}

	response := decodeResponse[protocol.ListVaultItemsResponse](t, recorder)
	if len(response) != 1 || response[0].ID != itemID {
		t.Fatalf("unexpected response: %#v", response)
	}
}

// TestVaultsControllerGetItem verifies that getItem returns a single encrypted
// vault item.
func TestVaultsControllerGetItem(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()
	itemID := uuid.New()

	itemsService := &fakeVaultItemsService{
		getByIDItem: testVaultItemResult(vaultID, itemID),
	}
	router := newVaultsRouter(
		&fakeVaultsService{},
		itemsService,
		testSession(userID),
	)

	recorder := performJSONRequest(
		router,
		http.MethodGet,
		"/vaults/"+vaultID.String()+"/items/"+itemID.String(),
		nil,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusOK,
		)
	}
	if itemsService.getByIDCalls != 1 ||
		itemsService.getByIDUser != userID ||
		itemsService.getByIDVault != vaultID ||
		itemsService.getByIDID != itemID {
		t.Fatalf("unexpected get item call: %#v", itemsService)
	}
}

// TestVaultsControllerGetItemNotFound verifies that getItem maps missing items
// to not-found responses.
func TestVaultsControllerGetItemNotFound(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()
	itemID := uuid.New()

	itemsService := &fakeVaultItemsService{
		getByIDErr: services.ErrVaultItemNotFound,
	}
	router := newVaultsRouter(
		&fakeVaultsService{},
		itemsService,
		testSession(userID),
	)

	recorder := performJSONRequest(
		router,
		http.MethodGet,
		"/vaults/"+vaultID.String()+"/items/"+itemID.String(),
		nil,
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusNotFound,
		)
	}

	response := decodeProtocolError(t, recorder)
	if response.Code != protocol.ProtocolErrorVaultItemNotFound {
		t.Fatalf(
			"unexpected error code: got %s want %s",
			response.Code,
			protocol.ProtocolErrorVaultItemNotFound,
		)
	}
}

// TestVaultsControllerUpdateItem verifies that updateItem maps encrypted item
// material to the vault item service.
func TestVaultsControllerUpdateItem(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()
	itemID := uuid.New()

	itemsService := &fakeVaultItemsService{
		updateItem: testVaultItemResult(vaultID, itemID),
	}
	router := newVaultsRouter(
		&fakeVaultsService{},
		itemsService,
		testSession(userID),
	)
	request := testUpdateVaultItemRequest()

	recorder := performJSONRequest(
		router,
		http.MethodPatch,
		"/vaults/"+vaultID.String()+"/items/"+itemID.String(),
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusOK,
		)
	}
	if itemsService.updateCalls != 1 ||
		itemsService.updateUser != userID ||
		itemsService.updateVault != vaultID ||
		itemsService.updateID != itemID ||
		!bytes.Equal(itemsService.updateInput.Ciphertext, request.Ciphertext) {
		t.Fatalf("unexpected update item call: %#v", itemsService)
	}
}

// TestVaultsControllerDeleteItem verifies that deleteItem maps valid IDs to
// the vault item service and returns no content.
func TestVaultsControllerDeleteItem(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()
	itemID := uuid.New()
	itemsService := &fakeVaultItemsService{}
	router := newVaultsRouter(
		&fakeVaultsService{},
		itemsService,
		testSession(userID),
	)

	recorder := performJSONRequest(
		router,
		http.MethodDelete,
		"/vaults/"+vaultID.String()+"/items/"+itemID.String(),
		nil,
	)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusNoContent,
		)
	}
	if itemsService.deleteCalls != 1 ||
		itemsService.deleteUser != userID ||
		itemsService.deleteVault != vaultID ||
		itemsService.deleteID != itemID {
		t.Fatalf("unexpected delete item call: %#v", itemsService)
	}
}

func testVaultResult(
	ownerID uuid.UUID,
	vaultID uuid.UUID,
) *services.VaultResult {
	return &services.VaultResult{
		ID:        vaultID,
		OwnerID:   ownerID,
		Name:      "Personal",
		CreatedAt: fixedTime(),
		UpdatedAt: fixedTime().Add(time.Minute),
	}
}

func testKeyringResult(vaultID uuid.UUID) *services.KeyringResult {
	return &services.KeyringResult{
		VaultID:               vaultID,
		EncryptionSalt:        []byte{1, 2, 3},
		EncryptionNonce:       []byte{4, 5, 6},
		EncryptedMasterKey:    []byte{7, 8, 9},
		EncryptionTimeCost:    1,
		EncryptionMemoryCost:  64 * 1024,
		EncryptionParallelism: 4,
		EncryptionKeySize:     32,
		CreatedAt:             fixedTime(),
		UpdatedAt:             fixedTime().Add(time.Minute),
	}
}

func testVaultItemResult(
	vaultID uuid.UUID,
	itemID uuid.UUID,
) *services.VaultItemResult {
	return &services.VaultItemResult{
		ID:         itemID,
		VaultID:    vaultID,
		KeySalt:    []byte{1, 2},
		ItemNonce:  []byte{3, 4},
		Ciphertext: []byte{5, 6},
		CreatedAt:  fixedTime(),
		UpdatedAt:  fixedTime().Add(time.Minute),
	}
}

func testCreateVaultRequest() protocol.CreateVaultRequest {
	return protocol.CreateVaultRequest{
		Name:                  "Personal",
		EncryptionSalt:        []byte{1, 2, 3},
		EncryptionNonce:       []byte{4, 5, 6},
		EncryptedMasterKey:    []byte{7, 8, 9},
		EncryptionTimeCost:    1,
		EncryptionMemoryCost:  64 * 1024,
		EncryptionParallelism: 4,
		EncryptionKeySize:     32,
	}
}

func testVaultItemRequest() protocol.CreateVaultItemRequest {
	return protocol.CreateVaultItemRequest{
		KeySalt:    []byte{1, 2},
		ItemNonce:  []byte{3, 4},
		Ciphertext: []byte{5, 6},
	}
}

func testUpdateVaultItemRequest() protocol.UpdateVaultItemRequest {
	return protocol.UpdateVaultItemRequest{
		KeySalt:    []byte{7, 8},
		ItemNonce:  []byte{9, 10},
		Ciphertext: []byte{11, 12},
	}
}
