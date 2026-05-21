package repositories

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// TestKeyringsRepositoryCreate verifies that Create stores keyring material
// for a user-vault pair.
func TestKeyringsRepositoryCreate(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewKeyringsRepository(pool)

	userID := createTestUser(t, ctx, pool, "keyring-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, userID, "Personal")
	input := testKeyringInput(userID, vaultID)

	keyring, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if keyring.UserID != userID || keyring.VaultID != vaultID {
		t.Fatalf("unexpected keyring: %#v", keyring)
	}
}

// TestKeyringsRepositoryCreateDuplicate verifies that Create preserves the
// composite primary-key constraint.
func TestKeyringsRepositoryCreateDuplicate(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewKeyringsRepository(pool)
	userID := createTestUser(t, ctx, pool, "keyring-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, userID, "Personal")
	input := testKeyringInput(userID, vaultID)

	if _, err := repo.Create(ctx, input); err != nil {
		t.Fatalf("failed to create test keyring: %v", err)
	}
	_, err := repo.Create(ctx, input)
	assertPgErrorCode(t, err, "23505")
}

// TestKeyringsRepositoryCreateUnknownUser verifies that Create preserves the
// foreign-key constraint on user_id.
func TestKeyringsRepositoryCreateUnknownUser(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewKeyringsRepository(pool)
	ownerID := createTestUser(t, ctx, pool, "keyring-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, ownerID, "Personal")

	_, err := repo.Create(ctx, testKeyringInput(uuid.New(), vaultID))
	assertPgErrorCode(t, err, "23503")
}

// TestKeyringsRepositoryCreateUnknownVault verifies that Create preserves the
// foreign-key constraint on vault_id.
func TestKeyringsRepositoryCreateUnknownVault(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewKeyringsRepository(pool)
	userID := createTestUser(t, ctx, pool, "keyring-owner@example.com")

	_, err := repo.Create(ctx, testKeyringInput(userID, uuid.New()))
	assertPgErrorCode(t, err, "23503")
}

// TestKeyringsRepositoryGet verifies that Get returns the stored encrypted
// keyring material without data loss.
func TestKeyringsRepositoryGet(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewKeyringsRepository(pool)

	userID := createTestUser(t, ctx, pool, "keyring-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, userID, "Personal")
	input := testKeyringInput(userID, vaultID)

	if _, err := repo.Create(ctx, input); err != nil {
		t.Fatalf("failed to create test keyring: %v", err)
	}

	keyring, err := repo.Get(ctx, userID, vaultID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !bytes.Equal(keyring.EncryptionSalt, input.EncryptionSalt) ||
		!bytes.Equal(keyring.EncryptionNonce, input.EncryptionNonce) ||
		!bytes.Equal(keyring.EncryptedMasterKey, input.EncryptedMasterKey) ||
		keyring.EncryptionTimeCost != input.EncryptionTimeCost ||
		keyring.EncryptionMemoryCost != input.EncryptionMemoryCost ||
		keyring.EncryptionParallelism != input.EncryptionParallelism ||
		keyring.EncryptionKeySize != input.EncryptionKeySize {
		t.Fatalf("unexpected keyring: %#v", keyring)
	}
}

// TestKeyringsRepositoryGetNotFound verifies that Get returns
// ErrKeyringNotFound for an unknown user-vault pair.
func TestKeyringsRepositoryGetNotFound(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewKeyringsRepository(pool)

	_, err := repo.Get(ctx, uuid.New(), uuid.New())
	if !errors.Is(err, ErrKeyringNotFound) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrKeyringNotFound)
	}
}

// TestKeyringsRepositoryDelete verifies that Delete removes a keyring for a
// known user-vault pair.
func TestKeyringsRepositoryDelete(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewKeyringsRepository(pool)

	userID := createTestUser(t, ctx, pool, "keyring-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, userID, "Personal")

	if _, err := repo.Create(ctx, testKeyringInput(userID, vaultID)); err != nil {
		t.Fatalf("failed to create test keyring: %v", err)
	}

	if err := repo.Delete(ctx, userID, vaultID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
}

// TestKeyringsRepositoryDeleteNotFound verifies that Delete returns
// ErrKeyringNotFound for an unknown user-vault pair.
func TestKeyringsRepositoryDeleteNotFound(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewKeyringsRepository(pool)

	if err := repo.Delete(ctx, uuid.New(), uuid.New()); !errors.Is(err, ErrKeyringNotFound) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrKeyringNotFound)
	}
}

func testKeyringInput(userID uuid.UUID, vaultID uuid.UUID) CreateKeyringInput {
	return CreateKeyringInput{
		UserID:                userID,
		VaultID:               vaultID,
		EncryptionSalt:        []byte{1, 2, 3},
		EncryptionNonce:       []byte{4, 5, 6},
		EncryptedMasterKey:    []byte{7, 8, 9},
		EncryptionTimeCost:    1,
		EncryptionMemoryCost:  64 * 1024,
		EncryptionParallelism: 4,
		EncryptionKeySize:     32,
	}
}
