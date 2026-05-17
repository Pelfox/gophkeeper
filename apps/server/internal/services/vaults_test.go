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

// TestVaultsServiceCreateSuccess verifies that vault creation stores both the
// vault metadata and its encrypted keyring.
func TestVaultsServiceCreateSuccess(t *testing.T) {
	ownerID := uuid.New()
	vault := &models.Vault{
		ID:        uuid.New(),
		OwnerID:   ownerID,
		Name:      "Personal",
		CreatedAt: fixedTime(),
		UpdatedAt: fixedTime().Add(time.Minute),
	}

	vaults := &fakeVaultsRepository{createItem: vault}
	keyrings := &fakeKeyringsRepository{}
	service := NewVaultsService(
		&fakeUsersRepository{},
		vaults,
		keyrings,
		testLogger(),
	)
	input := CreateVaultInput{
		OwnerID:               ownerID,
		Name:                  vault.Name,
		EncryptionSalt:        []byte{1, 2, 3},
		EncryptionNonce:       []byte{4, 5, 6},
		EncryptedMasterKey:    []byte{7, 8, 9},
		EncryptionTimeCost:    1,
		EncryptionMemoryCost:  2,
		EncryptionParallelism: 3,
		EncryptionKeySize:     4,
	}

	result, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if vaults.createCalls != 1 {
		t.Fatalf(
			"expected vaults.Create to be called once, got %d",
			vaults.createCalls,
		)
	}
	if vaults.createInput.OwnerID != ownerID ||
		vaults.createInput.Name != vault.Name {
		t.Fatalf("unexpected vault create input: %#v", vaults.createInput)
	}
	if keyrings.createCalls != 1 {
		t.Fatalf(
			"expected keyrings.Create to be called once, got %d",
			keyrings.createCalls,
		)
	}
	if keyrings.createInput.UserID != ownerID ||
		keyrings.createInput.VaultID != vault.ID {
		t.Fatalf("unexpected keyring owner/vault: %#v", keyrings.createInput)
	}
	if !bytes.Equal(keyrings.createInput.EncryptionSalt, input.EncryptionSalt) ||
		!bytes.Equal(keyrings.createInput.EncryptionNonce, input.EncryptionNonce) ||
		!bytes.Equal(keyrings.createInput.EncryptedMasterKey, input.EncryptedMasterKey) {
		t.Fatalf(
			"keyring encryption bytes were not forwarded: %#v",
			keyrings.createInput,
		)
	}
	if keyrings.createInput.EncryptionTimeCost != input.EncryptionTimeCost ||
		keyrings.createInput.EncryptionMemoryCost != input.EncryptionMemoryCost ||
		keyrings.createInput.EncryptionParallelism != input.EncryptionParallelism ||
		keyrings.createInput.EncryptionKeySize != input.EncryptionKeySize {
		t.Fatalf(
			"keyring KDF parameters were not forwarded: %#v",
			keyrings.createInput,
		)
	}
	if result.ID != vault.ID || result.OwnerID != ownerID || result.Name != vault.Name {
		t.Fatalf("unexpected result: %#v", result)
	}
}

// TestVaultsServiceCreateErrors verifies that vault and keyring creation
// failures are translated to vault creation errors.
func TestVaultsServiceCreateErrors(t *testing.T) {
	tests := []struct {
		name          string
		vaults        *fakeVaultsRepository
		keyrings      *fakeKeyringsRepository
		expectedCalls int
	}{
		{
			name: "vault creation failed",
			vaults: &fakeVaultsRepository{
				createErr: errors.New("insert vault failed"),
			},
			keyrings:      &fakeKeyringsRepository{},
			expectedCalls: 0,
		},
		{
			name: "keyring creation failed",
			vaults: &fakeVaultsRepository{
				createItem: &models.Vault{
					ID:      uuid.New(),
					OwnerID: uuid.New(),
					Name:    "Personal",
				},
			},
			keyrings: &fakeKeyringsRepository{
				createErr: errors.New("insert keyring failed"),
			},
			expectedCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewVaultsService(
				&fakeUsersRepository{},
				tt.vaults,
				tt.keyrings,
				testLogger(),
			)

			result, err := service.Create(
				context.Background(),
				CreateVaultInput{
					OwnerID:               uuid.New(),
					Name:                  "Personal",
					EncryptionSalt:        []byte{1},
					EncryptionNonce:       []byte{2},
					EncryptedMasterKey:    []byte{3},
					EncryptionTimeCost:    1,
					EncryptionMemoryCost:  1,
					EncryptionParallelism: 1,
					EncryptionKeySize:     1,
				},
			)
			if !errors.Is(err, ErrVaultCreationFailed) {
				t.Fatalf(
					"unexpected error: got %v want %v",
					err,
					ErrVaultCreationFailed,
				)
			}
			if result != nil {
				t.Fatalf("expected nil result, got %#v", result)
			}
			if tt.keyrings.createCalls != tt.expectedCalls {
				t.Fatalf(
					"unexpected keyrings.Create calls: got %d want %d",
					tt.keyrings.createCalls,
					tt.expectedCalls,
				)
			}
		})
	}
}

// TestVaultsServiceGetForUser verifies that user vaults are retrieved and
// mapped to service results.
func TestVaultsServiceGetForUser(t *testing.T) {
	userID := uuid.New()
	rawVaults := []models.Vault{
		{
			ID:        uuid.New(),
			OwnerID:   userID,
			Name:      "Personal",
			CreatedAt: fixedTime(),
			UpdatedAt: fixedTime().Add(time.Minute),
		},
		{
			ID:        uuid.New(),
			OwnerID:   userID,
			Name:      "Work",
			CreatedAt: fixedTime().Add(2 * time.Minute),
			UpdatedAt: fixedTime().Add(3 * time.Minute),
		},
	}

	vaults := &fakeVaultsRepository{getForUserItems: rawVaults}
	service := NewVaultsService(
		&fakeUsersRepository{},
		vaults,
		&fakeKeyringsRepository{},
		testLogger(),
	)

	result, err := service.GetForUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetForUser returned error: %v", err)
	}
	if vaults.getForUserCalls != 1 || vaults.getForUserID != userID {
		t.Fatalf(
			"unexpected GetForUser call: calls=%d id=%s",
			vaults.getForUserCalls,
			vaults.getForUserID,
		)
	}
	if len(result) != len(rawVaults) {
		t.Fatalf(
			"unexpected result length: got %d want %d",
			len(result),
			len(rawVaults),
		)
	}
	for i := range result {
		if result[i].ID != rawVaults[i].ID ||
			result[i].OwnerID != rawVaults[i].OwnerID ||
			result[i].Name != rawVaults[i].Name {
			t.Fatalf("unexpected vault at index %d: %#v", i, result[i])
		}
	}
}

// TestVaultsServiceGetForUserError verifies that vault list repository errors
// are translated to service-level query errors.
func TestVaultsServiceGetForUserError(t *testing.T) {
	vaults := &fakeVaultsRepository{getForUserErr: errors.New("select failed")}
	service := NewVaultsService(
		&fakeUsersRepository{},
		vaults,
		&fakeKeyringsRepository{},
		testLogger(),
	)

	result, err := service.GetForUser(context.Background(), uuid.New())
	if !errors.Is(err, ErrVaultsQueryFailed) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrVaultsQueryFailed)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
}

// TestVaultsServiceGetKeyringSuccess verifies that vault keyring data is
// retrieved and mapped without losing encrypted material.
func TestVaultsServiceGetKeyringSuccess(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()
	keyring := &models.Keyring{
		UserID:                userID,
		VaultID:               vaultID,
		EncryptionSalt:        []byte{1, 2},
		EncryptionNonce:       []byte{3, 4},
		EncryptedMasterKey:    []byte{5, 6},
		EncryptionTimeCost:    1,
		EncryptionMemoryCost:  2,
		EncryptionParallelism: 3,
		EncryptionKeySize:     4,
		CreatedAt:             fixedTime(),
		UpdatedAt:             fixedTime().Add(time.Minute),
	}

	keyrings := &fakeKeyringsRepository{getItem: keyring}
	service := NewVaultsService(
		&fakeUsersRepository{},
		&fakeVaultsRepository{},
		keyrings,
		testLogger(),
	)

	result, err := service.GetKeyring(context.Background(), userID, vaultID)
	if err != nil {
		t.Fatalf("GetKeyring returned error: %v", err)
	}
	if keyrings.getCalls != 1 ||
		keyrings.getUser != userID ||
		keyrings.getVault != vaultID {
		t.Fatalf(
			"unexpected keyring lookup: calls=%d user=%s vault=%s",
			keyrings.getCalls,
			keyrings.getUser,
			keyrings.getVault,
		)
	}
	if result.VaultID != vaultID ||
		!bytes.Equal(result.EncryptionSalt, keyring.EncryptionSalt) ||
		!bytes.Equal(result.EncryptedMasterKey, keyring.EncryptedMasterKey) {
		t.Fatalf("unexpected result: %#v", result)
	}
}

// TestVaultsServiceGetKeyringErrors verifies keyring not found and query
// failure handling.
func TestVaultsServiceGetKeyringErrors(t *testing.T) {
	tests := []struct {
		name          string
		repoErr       error
		expectedError error
	}{
		{
			name:          "not found",
			repoErr:       repositories.ErrKeyringNotFound,
			expectedError: ErrVaultNotFound,
		},
		{
			name:          "query failed",
			repoErr:       errors.New("select failed"),
			expectedError: ErrVaultKeyringRetrievalFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyrings := &fakeKeyringsRepository{getErr: tt.repoErr}
			service := NewVaultsService(
				&fakeUsersRepository{},
				&fakeVaultsRepository{},
				keyrings,
				testLogger(),
			)

			result, err := service.GetKeyring(
				context.Background(),
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

// TestVaultsServiceUpdateSuccess verifies that vault updates are scoped by
// owner and mapped to service results.
func TestVaultsServiceUpdateSuccess(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()
	name := "Updated"

	vault := &models.Vault{
		ID:        vaultID,
		OwnerID:   userID,
		Name:      name,
		CreatedAt: fixedTime(),
		UpdatedAt: fixedTime().Add(time.Minute),
	}
	vaults := &fakeVaultsRepository{updateItem: vault}
	service := NewVaultsService(
		&fakeUsersRepository{},
		vaults,
		&fakeKeyringsRepository{},
		testLogger(),
	)

	result, err := service.Update(
		context.Background(),
		vaultID,
		userID,
		UpdateVaultInput{Name: &name},
	)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if vaults.updateCalls != 1 ||
		vaults.updateID != vaultID ||
		vaults.updateOwner != userID {
		t.Fatalf(
			"unexpected update call: calls=%d id=%s user=%s",
			vaults.updateCalls,
			vaults.updateID,
			vaults.updateOwner,
		)
	}
	if vaults.updateInput.Name == nil || *vaults.updateInput.Name != name {
		t.Fatalf("unexpected update input: %#v", vaults.updateInput)
	}
	if result.ID != vaultID || result.OwnerID != userID || result.Name != name {
		t.Fatalf("unexpected result: %#v", result)
	}
}

// TestVaultsServiceUpdateErrors verifies that update failures are translated to
// expected service-level errors.
func TestVaultsServiceUpdateErrors(t *testing.T) {
	tests := []struct {
		name          string
		repoErr       error
		expectedError error
	}{
		{
			name:          "not found",
			repoErr:       repositories.ErrVaultNotFound,
			expectedError: ErrVaultNotFound,
		},
		{
			name:          "update failed",
			repoErr:       errors.New("update failed"),
			expectedError: ErrVaultUpdateFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vaults := &fakeVaultsRepository{updateErr: tt.repoErr}
			service := NewVaultsService(
				&fakeUsersRepository{},
				vaults,
				&fakeKeyringsRepository{},
				testLogger(),
			)
			name := "Updated"

			result, err := service.Update(
				context.Background(),
				uuid.New(),
				uuid.New(),
				UpdateVaultInput{Name: &name},
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

// TestVaultsServiceDeleteSuccess verifies that vault deletion is delegated with
// the vault ID and owner ID.
func TestVaultsServiceDeleteSuccess(t *testing.T) {
	userID := uuid.New()
	vaultID := uuid.New()

	vaults := &fakeVaultsRepository{}
	service := NewVaultsService(
		&fakeUsersRepository{},
		vaults,
		&fakeKeyringsRepository{},
		testLogger(),
	)

	err := service.Delete(context.Background(), vaultID, userID)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if vaults.deleteCalls != 1 ||
		vaults.deleteID != vaultID ||
		vaults.deleteOwner != userID {
		t.Fatalf(
			"unexpected delete call: calls=%d id=%s user=%s",
			vaults.deleteCalls,
			vaults.deleteID,
			vaults.deleteOwner,
		)
	}
}

// TestVaultsServiceDeleteErrors verifies that delete failures are translated to
// expected service-level errors.
func TestVaultsServiceDeleteErrors(t *testing.T) {
	tests := []struct {
		name          string
		repoErr       error
		expectedError error
	}{
		{
			name:          "not found",
			repoErr:       repositories.ErrVaultNotFound,
			expectedError: ErrVaultNotFound,
		},
		{
			name:          "delete failed",
			repoErr:       errors.New("delete failed"),
			expectedError: ErrVaultDeletionFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vaults := &fakeVaultsRepository{deleteErr: tt.repoErr}
			service := NewVaultsService(
				&fakeUsersRepository{},
				vaults,
				&fakeKeyringsRepository{},
				testLogger(),
			)

			err := service.Delete(context.Background(), uuid.New(), uuid.New())
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
