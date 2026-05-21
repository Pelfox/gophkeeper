package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// TestVaultsRepositoryCreate verifies that Create stores a vault for its
// owner and returns generated metadata.
func TestVaultsRepositoryCreate(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultsRepository(pool)

	ownerID := createTestUser(t, ctx, pool, "vault-owner@example.com")

	vault, err := repo.Create(ctx, CreateVaultInput{
		OwnerID: ownerID,
		Name:    "Personal",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if vault.ID == uuid.Nil ||
		vault.OwnerID != ownerID ||
		vault.Name != "Personal" {
		t.Fatalf("unexpected vault: %#v", vault)
	}
}

// TestVaultsRepositoryCreateUnknownOwner verifies that Create preserves the
// foreign-key constraint on owner_id.
func TestVaultsRepositoryCreateUnknownOwner(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultsRepository(pool)

	_, err := repo.Create(ctx, CreateVaultInput{
		OwnerID: uuid.New(),
		Name:    "Personal",
	})
	assertPgErrorCode(t, err, "23503")
}

// TestVaultsRepositoryGetForUser verifies that GetForUser returns vaults owned
// by the supplied user.
func TestVaultsRepositoryGetForUser(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultsRepository(pool)

	ownerID := createTestUser(t, ctx, pool, "vault-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, ownerID, "Personal")

	vaults, err := repo.GetForUser(ctx, ownerID)
	if err != nil {
		t.Fatalf("GetForUser returned error: %v", err)
	}
	if len(vaults) != 1 || vaults[0].ID != vaultID {
		t.Fatalf("unexpected vaults: %#v", vaults)
	}
}

// TestVaultsRepositoryGetForUserEmpty verifies that GetForUser returns an
// empty slice when the user owns no vaults.
func TestVaultsRepositoryGetForUserEmpty(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultsRepository(pool)
	ownerID := createTestUser(t, ctx, pool, "vault-owner@example.com")

	vaults, err := repo.GetForUser(ctx, ownerID)
	if err != nil {
		t.Fatalf("GetForUser returned error: %v", err)
	}
	if len(vaults) != 0 {
		t.Fatalf("expected no vaults, got %#v", vaults)
	}
}

// TestVaultsRepositoryUpdate verifies that Update persists vault fields when
// the owner matches.
func TestVaultsRepositoryUpdate(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultsRepository(pool)

	ownerID := createTestUser(t, ctx, pool, "vault-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, ownerID, "Personal")
	renamed := "Work"

	vault, err := repo.Update(ctx, vaultID, ownerID, UpdateVaultInput{
		Name: &renamed,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if vault.ID != vaultID ||
		vault.OwnerID != ownerID ||
		vault.Name != renamed {
		t.Fatalf("unexpected vault: %#v", vault)
	}
}

// TestVaultsRepositoryUpdateNoFields verifies that Update with no mutable
// fields still refreshes metadata and returns the existing vault.
func TestVaultsRepositoryUpdateNoFields(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultsRepository(pool)

	ownerID := createTestUser(t, ctx, pool, "vault-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, ownerID, "Personal")

	vault, err := repo.Update(ctx, vaultID, ownerID, UpdateVaultInput{})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if vault.ID != vaultID || vault.OwnerID != ownerID || vault.Name != "Personal" {
		t.Fatalf("unexpected vault: %#v", vault)
	}
}

// TestVaultsRepositoryUpdateWrongOwner verifies that Update is scoped by
// owner and hides vaults owned by another user.
func TestVaultsRepositoryUpdateWrongOwner(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultsRepository(pool)

	ownerID := createTestUser(t, ctx, pool, "vault-owner@example.com")
	otherOwnerID := createTestUser(t, ctx, pool, "other-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, ownerID, "Personal")
	renamed := "Work"

	_, err := repo.Update(
		ctx,
		vaultID,
		otherOwnerID,
		UpdateVaultInput{Name: &renamed},
	)
	if !errors.Is(err, ErrVaultNotFound) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrVaultNotFound)
	}
}

// TestVaultsRepositoryDelete verifies that Delete removes a vault when the
// owner matches.
func TestVaultsRepositoryDelete(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultsRepository(pool)

	ownerID := createTestUser(t, ctx, pool, "vault-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, ownerID, "Personal")

	if err := repo.Delete(ctx, vaultID, ownerID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
}

// TestVaultsRepositoryDeleteWrongOwner verifies that Delete is scoped by owner
// and hides vaults owned by another user.
func TestVaultsRepositoryDeleteWrongOwner(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultsRepository(pool)

	ownerID := createTestUser(t, ctx, pool, "vault-owner@example.com")
	otherOwnerID := createTestUser(t, ctx, pool, "other-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, ownerID, "Personal")

	if err := repo.Delete(ctx, vaultID, otherOwnerID); !errors.Is(err, ErrVaultNotFound) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrVaultNotFound)
	}
}

// TestVaultsRepositoryDeleteNotFound verifies that Delete returns
// ErrVaultNotFound for an unknown vault.
func TestVaultsRepositoryDeleteNotFound(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultsRepository(pool)

	ownerID := createTestUser(t, ctx, pool, "vault-owner@example.com")

	if err := repo.Delete(ctx, uuid.New(), ownerID); !errors.Is(err, ErrVaultNotFound) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrVaultNotFound)
	}
}

// TestVaultsRepositoryDeleteCascades verifies that deleting a vault removes
// dependent keyrings and items through database cascades.
func TestVaultsRepositoryDeleteCascades(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewVaultsRepository(pool)
	keyrings := NewKeyringsRepository(pool)
	items := NewVaultItemsRepository(pool)

	ownerID := createTestUser(t, ctx, pool, "vault-owner@example.com")
	vaultID := createTestVault(t, ctx, pool, ownerID, "Personal")
	if _, err := keyrings.Create(ctx, testKeyringInput(ownerID, vaultID)); err != nil {
		t.Fatalf("failed to create test keyring: %v", err)
	}
	item := createTestVaultItem(t, ctx, items, ownerID, vaultID)

	if err := repo.Delete(ctx, vaultID, ownerID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if _, err := keyrings.Get(ctx, ownerID, vaultID); !errors.Is(err, ErrKeyringNotFound) {
		t.Fatalf("unexpected keyring error: got %v want %v", err, ErrKeyringNotFound)
	}
	if _, err := items.GetByID(ctx, ownerID, vaultID, item.ID); !errors.Is(err, ErrVaultItemNotFound) {
		t.Fatalf("unexpected item error: got %v want %v", err, ErrVaultItemNotFound)
	}
}
