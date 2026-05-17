package services

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Pelfox/gophkeeper/apps/server/internal/models"
	"github.com/Pelfox/gophkeeper/apps/server/internal/repositories"
	"github.com/google/uuid"
)

func testVaultItemInput() VaultItemInput {
	return VaultItemInput{
		KeySalt:    []byte{1, 2, 3},
		ItemNonce:  []byte{4, 5, 6},
		Ciphertext: []byte{7, 8, 9},
	}
}

func testVaultItemModel(vaultID uuid.UUID, itemID uuid.UUID) *models.VaultItem {
	return &models.VaultItem{
		ID:         itemID,
		VaultID:    vaultID,
		KeySalt:    []byte{1, 2, 3},
		ItemNonce:  []byte{4, 5, 6},
		Ciphertext: []byte{7, 8, 9},
		CreatedAt:  fixedTime(),
		UpdatedAt:  fixedTime().Add(time.Minute),
	}
}

func assertVaultItemResult(t *testing.T, result *VaultItemResult, expected *models.VaultItem) {
	t.Helper()

	if result.ID != expected.ID || result.VaultID != expected.VaultID {
		t.Fatalf("unexpected result IDs: got %#v want %#v", result, expected)
	}
	if !bytes.Equal(result.KeySalt, expected.KeySalt) ||
		!bytes.Equal(result.ItemNonce, expected.ItemNonce) ||
		!bytes.Equal(result.Ciphertext, expected.Ciphertext) {
		t.Fatalf("unexpected result bytes: got %#v want %#v", result, expected)
	}
	if !result.CreatedAt.Equal(expected.CreatedAt) ||
		!result.UpdatedAt.Equal(expected.UpdatedAt) {
		t.Fatalf(
			"unexpected timestamps: got %s/%s want %s/%s",
			result.CreatedAt,
			result.UpdatedAt,
			expected.CreatedAt,
			expected.UpdatedAt,
		)
	}
}

// TestVaultItemsServiceCreateSuccess verifies that encrypted vault item
// material is forwarded to the repository and mapped back to a service result.
func TestVaultItemsServiceCreateSuccess(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()
	itemID := uuid.New()

	input := testVaultItemInput()
	repoItem := testVaultItemModel(vaultID, itemID)
	repo := &fakeVaultItemsRepository{createItem: repoItem}
	service := NewVaultItemsService(repo, testLogger())

	result, err := service.Create(context.Background(), userID, vaultID, input)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if repo.createCalls != 1 || repo.createUser != userID {
		t.Fatalf(
			"unexpected create call: calls=%d user=%s",
			repo.createCalls,
			repo.createUser,
		)
	}
	if repo.createInput.VaultID != vaultID ||
		!bytes.Equal(repo.createInput.KeySalt, input.KeySalt) ||
		!bytes.Equal(repo.createInput.ItemNonce, input.ItemNonce) ||
		!bytes.Equal(repo.createInput.Ciphertext, input.Ciphertext) {
		t.Fatalf("unexpected create input: %#v", repo.createInput)
	}
	assertVaultItemResult(t, result, repoItem)
}

// TestVaultItemsServiceCreateErrors verifies vault item creation error
// translation.
func TestVaultItemsServiceCreateErrors(t *testing.T) {
	tests := []struct {
		name          string
		repoErr       error
		expectedError error
	}{
		{
			name:          "vault not found",
			repoErr:       repositories.ErrVaultNotFound,
			expectedError: ErrVaultNotFound,
		},
		{
			name:          "create failed",
			repoErr:       errors.New("insert failed"),
			expectedError: ErrVaultItemCreationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeVaultItemsRepository{createErr: tt.repoErr}
			service := NewVaultItemsService(repo, testLogger())

			result, err := service.Create(
				context.Background(),
				uuid.New(),
				uuid.New(),
				testVaultItemInput(),
			)
			if !errors.Is(err, tt.expectedError) {
				t.Fatalf(
					"unexpected error: got %v want %v",
					err,
					tt.expectedError,
				)
			}
			if result != nil {
				t.Fatalf("expected nil result, got %#v", result)
			}
		})
	}
}

// TestVaultItemsServiceUpdateSuccess verifies that encrypted vault item updates
// are forwarded and mapped back to a service result.
func TestVaultItemsServiceUpdateSuccess(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()
	itemID := uuid.New()

	input := testVaultItemInput()
	repoItem := testVaultItemModel(vaultID, itemID)
	repo := &fakeVaultItemsRepository{updateValue: repoItem}
	service := NewVaultItemsService(repo, testLogger())

	result, err := service.Update(
		context.Background(),
		userID,
		vaultID,
		itemID,
		input,
	)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if repo.updateCalls != 1 ||
		repo.updateUser != userID ||
		repo.updateVault != vaultID ||
		repo.updateItem != itemID {
		t.Fatalf(
			"unexpected update call: calls=%d user=%s vault=%s item=%s",
			repo.updateCalls,
			repo.updateUser,
			repo.updateVault,
			repo.updateItem,
		)
	}
	if !bytes.Equal(repo.updateInput.KeySalt, input.KeySalt) ||
		!bytes.Equal(repo.updateInput.ItemNonce, input.ItemNonce) ||
		!bytes.Equal(repo.updateInput.Ciphertext, input.Ciphertext) {
		t.Fatalf("unexpected update input: %#v", repo.updateInput)
	}
	assertVaultItemResult(t, result, repoItem)
}

// TestVaultItemsServiceUpdateErrors verifies vault item update error
// translation.
func TestVaultItemsServiceUpdateErrors(t *testing.T) {
	tests := []struct {
		name          string
		repoErr       error
		expectedError error
	}{
		{
			name:          "item not found",
			repoErr:       repositories.ErrVaultItemNotFound,
			expectedError: ErrVaultItemNotFound,
		},
		{
			name:          "update failed",
			repoErr:       errors.New("update failed"),
			expectedError: ErrVaultItemUpdateFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeVaultItemsRepository{updateErr: tt.repoErr}
			service := NewVaultItemsService(repo, testLogger())

			result, err := service.Update(
				context.Background(),
				uuid.New(),
				uuid.New(),
				uuid.New(),
				testVaultItemInput(),
			)
			if !errors.Is(err, tt.expectedError) {
				t.Fatalf(
					"unexpected error: got %v want %v",
					err,
					tt.expectedError,
				)
			}
			if result != nil {
				t.Fatalf("expected nil result, got %#v", result)
			}
		})
	}
}

// TestVaultItemsServiceGetByIDSuccess verifies that a single encrypted vault
// item is retrieved by user, vault and item IDs.
func TestVaultItemsServiceGetByIDSuccess(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()
	itemID := uuid.New()

	repoItem := testVaultItemModel(vaultID, itemID)
	repo := &fakeVaultItemsRepository{getByIDValue: repoItem}
	service := NewVaultItemsService(repo, testLogger())

	result, err := service.GetByID(context.Background(), userID, vaultID, itemID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if repo.getByIDCalls != 1 ||
		repo.getByIDUser != userID ||
		repo.getByIDVault != vaultID ||
		repo.getByIDItem != itemID {
		t.Fatalf(
			"unexpected get call: calls=%d user=%s vault=%s item=%s",
			repo.getByIDCalls,
			repo.getByIDUser,
			repo.getByIDVault,
			repo.getByIDItem,
		)
	}
	assertVaultItemResult(t, result, repoItem)
}

// TestVaultItemsServiceGetByIDErrors verifies vault item lookup error
// translation.
func TestVaultItemsServiceGetByIDErrors(t *testing.T) {
	tests := []struct {
		name          string
		repoErr       error
		expectedError error
	}{
		{
			name:          "item not found",
			repoErr:       repositories.ErrVaultItemNotFound,
			expectedError: ErrVaultItemNotFound,
		},
		{
			name:          "query failed",
			repoErr:       errors.New("select failed"),
			expectedError: ErrVaultItemsQueryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeVaultItemsRepository{getByIDErr: tt.repoErr}
			service := NewVaultItemsService(repo, testLogger())

			result, err := service.GetByID(
				context.Background(),
				uuid.New(),
				uuid.New(),
				uuid.New(),
			)
			if !errors.Is(err, tt.expectedError) {
				t.Fatalf(
					"unexpected error: got %v want %v",
					err,
					tt.expectedError,
				)
			}
			if result != nil {
				t.Fatalf("expected nil result, got %#v", result)
			}
		})
	}
}

// TestVaultItemsServiceGetForVaultSuccess verifies that all encrypted vault
// items for a vault are mapped to service results.
func TestVaultItemsServiceGetForVaultSuccess(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()

	rawItems := []models.VaultItem{
		*testVaultItemModel(vaultID, uuid.New()),
		{
			ID:         uuid.New(),
			VaultID:    vaultID,
			KeySalt:    []byte{10},
			ItemNonce:  []byte{11},
			Ciphertext: []byte{12},
			CreatedAt:  fixedTime().Add(2 * time.Minute),
			UpdatedAt:  fixedTime().Add(3 * time.Minute),
		},
	}
	repo := &fakeVaultItemsRepository{getForVaultItems: rawItems}
	service := NewVaultItemsService(repo, testLogger())

	result, err := service.GetForVault(context.Background(), userID, vaultID)
	if err != nil {
		t.Fatalf("GetForVault returned error: %v", err)
	}
	if repo.getForVaultCalls != 1 ||
		repo.getForVaultUser != userID ||
		repo.getForVaultVault != vaultID {
		t.Fatalf(
			"unexpected list call: calls=%d user=%s vault=%s",
			repo.getForVaultCalls,
			repo.getForVaultUser,
			repo.getForVaultVault,
		)
	}
	if len(result) != len(rawItems) {
		t.Fatalf(
			"unexpected item count: got %d want %d",
			len(result),
			len(rawItems),
		)
	}
	for i := range result {
		assertVaultItemResult(t, &result[i], &rawItems[i])
	}
}

// TestVaultItemsServiceGetForVaultError verifies vault item list error
// translation.
func TestVaultItemsServiceGetForVaultError(t *testing.T) {
	repo := &fakeVaultItemsRepository{
		getForVaultErr: errors.New("select failed"),
	}
	service := NewVaultItemsService(repo, testLogger())

	result, err := service.GetForVault(
		context.Background(),
		uuid.New(),
		uuid.New(),
	)
	if !errors.Is(err, ErrVaultItemsQueryFailed) {
		t.Fatalf(
			"unexpected error: got %v want %v",
			err,
			ErrVaultItemsQueryFailed,
		)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
}

// TestVaultItemsServiceDeleteSuccess verifies that vault item deletion is
// delegated with user, vault and item IDs.
func TestVaultItemsServiceDeleteSuccess(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()
	itemID := uuid.New()

	repo := &fakeVaultItemsRepository{}
	service := NewVaultItemsService(repo, testLogger())

	err := service.Delete(context.Background(), userID, vaultID, itemID)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if repo.deleteCalls != 1 ||
		repo.deleteUser != userID ||
		repo.deleteVault != vaultID ||
		repo.deleteItem != itemID {
		t.Fatalf(
			"unexpected delete call: calls=%d user=%s vault=%s item=%s",
			repo.deleteCalls,
			repo.deleteUser,
			repo.deleteVault,
			repo.deleteItem,
		)
	}
}

// TestVaultItemsServiceDeleteErrors verifies vault item delete error
// translation.
func TestVaultItemsServiceDeleteErrors(t *testing.T) {
	tests := []struct {
		name          string
		repoErr       error
		expectedError error
	}{
		{
			name:          "item not found",
			repoErr:       repositories.ErrVaultItemNotFound,
			expectedError: ErrVaultItemNotFound,
		},
		{
			name:          "delete failed",
			repoErr:       errors.New("delete failed"),
			expectedError: ErrVaultItemDeletionFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeVaultItemsRepository{deleteErr: tt.repoErr}
			service := NewVaultItemsService(repo, testLogger())

			err := service.Delete(
				context.Background(),
				uuid.New(),
				uuid.New(),
				uuid.New(),
			)
			if !errors.Is(err, tt.expectedError) {
				t.Fatalf(
					"unexpected error: got %v want %v",
					err,
					tt.expectedError,
				)
			}
		})
	}
}
