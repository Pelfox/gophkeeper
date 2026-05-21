package repositories

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/Pelfox/gophkeeper/apps/server/internal/models"
	"github.com/google/uuid"
)

// TestVaultItemsRepositoryCreate verifies that Create stores an encrypted item
// in a vault owned by the supplied user.
func TestVaultItemsRepositoryCreate(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultItemsRepository(pool)

	userID := createTestUser(t, ctx, pool, "items-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, userID, "Personal")

	item, err := repo.Create(ctx, userID, testCreateVaultItemInput(vaultID))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if item.ID == uuid.Nil || item.VaultID != vaultID {
		t.Fatalf("unexpected vault item: %#v", item)
	}
}

// TestVaultItemsRepositoryCreateWrongOwner verifies that Create rejects writes
// to vaults owned by another user.
func TestVaultItemsRepositoryCreateWrongOwner(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultItemsRepository(pool)

	userID := createTestUser(t, ctx, pool, "items-owner@example.com")
	otherUserID := createTestUser(t, ctx, pool, "items-other@example.com")
	vaultID := createTestVault(t, ctx, pool, userID, "Personal")

	_, err := repo.Create(ctx, otherUserID, testCreateVaultItemInput(vaultID))
	if !errors.Is(err, ErrVaultNotFound) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrVaultNotFound)
	}
}

// TestVaultItemsRepositoryGetByID verifies that GetByID returns an item when
// the vault and owner match.
func TestVaultItemsRepositoryGetByID(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultItemsRepository(pool)

	userID := createTestUser(t, ctx, pool, "items-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, userID, "Personal")
	item := createTestVaultItem(t, ctx, repo, userID, vaultID)

	found, err := repo.GetByID(ctx, userID, vaultID, item.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if found.ID != item.ID || found.VaultID != vaultID {
		t.Fatalf("unexpected vault item: %#v", found)
	}
}

// TestVaultItemsRepositoryGetByIDWrongOwner verifies that GetByID hides items
// from users who don't own the vault.
func TestVaultItemsRepositoryGetByIDWrongOwner(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultItemsRepository(pool)

	userID := createTestUser(t, ctx, pool, "items-owner@example.com")
	otherUserID := createTestUser(t, ctx, pool, "items-other@example.com")
	vaultID := createTestVault(t, ctx, pool, userID, "Personal")
	item := createTestVaultItem(t, ctx, repo, userID, vaultID)

	_, err := repo.GetByID(ctx, otherUserID, vaultID, item.ID)
	if !errors.Is(err, ErrVaultItemNotFound) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrVaultItemNotFound)
	}
}

// TestVaultItemsRepositoryGetByIDNotFound verifies that GetByID returns
// ErrVaultItemNotFound for an unknown item.
func TestVaultItemsRepositoryGetByIDNotFound(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultItemsRepository(pool)

	userID := createTestUser(t, ctx, pool, "items-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, userID, "Personal")

	_, err := repo.GetByID(ctx, userID, vaultID, uuid.New())
	if !errors.Is(err, ErrVaultItemNotFound) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrVaultItemNotFound)
	}
}

// TestVaultItemsRepositoryGetForVault verifies that GetForVault returns items
// from a vault owned by the supplied user.
func TestVaultItemsRepositoryGetForVault(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultItemsRepository(pool)

	userID := createTestUser(t, ctx, pool, "items-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, userID, "Personal")
	item := createTestVaultItem(t, ctx, repo, userID, vaultID)

	items, err := repo.GetForVault(ctx, userID, vaultID)
	if err != nil {
		t.Fatalf("GetForVault returned error: %v", err)
	}
	if len(items) != 1 || items[0].ID != item.ID {
		t.Fatalf("unexpected vault items: %#v", items)
	}
}

// TestVaultItemsRepositoryGetForVaultEmpty verifies that GetForVault returns
// an empty slice when the vault has no items.
func TestVaultItemsRepositoryGetForVaultEmpty(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultItemsRepository(pool)

	userID := createTestUser(t, ctx, pool, "items-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, userID, "Personal")

	items, err := repo.GetForVault(ctx, userID, vaultID)
	if err != nil {
		t.Fatalf("GetForVault returned error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no vault items, got %#v", items)
	}
}

// TestVaultItemsRepositoryUpdate verifies that Update persists encrypted item
// material when the vault and owner match.
func TestVaultItemsRepositoryUpdate(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultItemsRepository(pool)

	userID := createTestUser(t, ctx, pool, "items-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, userID, "Personal")
	item := createTestVaultItem(t, ctx, repo, userID, vaultID)
	input := UpdateVaultItemInput{
		KeySalt:    []byte{7, 8},
		ItemNonce:  []byte{9, 10},
		Ciphertext: []byte{11, 12},
	}

	updated, err := repo.Update(ctx, userID, vaultID, item.ID, input)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.VaultID != vaultID ||
		!bytes.Equal(updated.KeySalt, input.KeySalt) ||
		!bytes.Equal(updated.ItemNonce, input.ItemNonce) ||
		!bytes.Equal(updated.Ciphertext, input.Ciphertext) {
		t.Fatalf("unexpected vault item: %#v", updated)
	}
}

// TestVaultItemsRepositoryUpdateWrongOwner verifies that Update hides items
// from users who don't own the vault.
func TestVaultItemsRepositoryUpdateWrongOwner(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultItemsRepository(pool)

	userID := createTestUser(t, ctx, pool, "items-owner@example.com")
	otherUserID := createTestUser(t, ctx, pool, "items-other@example.com")
	vaultID := createTestVault(t, ctx, pool, userID, "Personal")
	item := createTestVaultItem(t, ctx, repo, userID, vaultID)

	_, err := repo.Update(
		ctx,
		otherUserID,
		vaultID,
		item.ID,
		testUpdateVaultItemInput(),
	)
	if !errors.Is(err, ErrVaultItemNotFound) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrVaultItemNotFound)
	}
}

// TestVaultItemsRepositoryUpdateNotFound verifies that Update returns
// ErrVaultItemNotFound for an unknown item.
func TestVaultItemsRepositoryUpdateNotFound(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultItemsRepository(pool)

	userID := createTestUser(t, ctx, pool, "items-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, userID, "Personal")

	_, err := repo.Update(
		ctx,
		userID,
		vaultID,
		uuid.New(),
		testUpdateVaultItemInput(),
	)
	if !errors.Is(err, ErrVaultItemNotFound) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrVaultItemNotFound)
	}
}

// TestVaultItemsRepositoryDelete verifies that Delete removes an item when the
// vault and owner match.
func TestVaultItemsRepositoryDelete(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultItemsRepository(pool)

	userID := createTestUser(t, ctx, pool, "items-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, userID, "Personal")
	item := createTestVaultItem(t, ctx, repo, userID, vaultID)

	if err := repo.Delete(ctx, userID, vaultID, item.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
}

// TestVaultItemsRepositoryDeleteWrongOwner verifies that Delete hides items
// from users who don't own the vault.
func TestVaultItemsRepositoryDeleteWrongOwner(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultItemsRepository(pool)

	userID := createTestUser(t, ctx, pool, "items-owner@example.com")
	otherUserID := createTestUser(t, ctx, pool, "items-other@example.com")
	vaultID := createTestVault(t, ctx, pool, userID, "Personal")
	item := createTestVaultItem(t, ctx, repo, userID, vaultID)

	if err := repo.Delete(ctx, otherUserID, vaultID, item.ID); !errors.Is(err, ErrVaultItemNotFound) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrVaultItemNotFound)
	}
}

// TestVaultItemsRepositoryDeleteNotFound verifies that Delete returns
// ErrVaultItemNotFound for an unknown item.
func TestVaultItemsRepositoryDeleteNotFound(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultItemsRepository(pool)

	userID := createTestUser(t, ctx, pool, "items-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, userID, "Personal")

	if err := repo.Delete(ctx, userID, vaultID, uuid.New()); !errors.Is(err, ErrVaultItemNotFound) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrVaultItemNotFound)
	}
}

func createTestVaultItem(
	t *testing.T,
	ctx context.Context,
	repo VaultItemsRepository,
	userID uuid.UUID,
	vaultID uuid.UUID,
) *models.VaultItem {
	t.Helper()

	item, err := repo.Create(ctx, userID, testCreateVaultItemInput(vaultID))
	if err != nil {
		t.Fatalf("failed to create test vault item: %v", err)
	}
	return item
}

func testCreateVaultItemInput(vaultID uuid.UUID) CreateVaultItemInput {
	return CreateVaultItemInput{
		VaultID:    vaultID,
		KeySalt:    []byte{1, 2},
		ItemNonce:  []byte{3, 4},
		Ciphertext: []byte{5, 6},
	}
}

func testUpdateVaultItemInput() UpdateVaultItemInput {
	return UpdateVaultItemInput{
		KeySalt:    []byte{7, 8},
		ItemNonce:  []byte{9, 10},
		Ciphertext: []byte{11, 12},
	}
}
