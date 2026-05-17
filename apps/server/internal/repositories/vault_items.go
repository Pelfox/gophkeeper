package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/Pelfox/gophkeeper/apps/server/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrVaultItemNotFound is returned when no vault item was found for the
	// given conditions.
	ErrVaultItemNotFound = errors.New("vault item not found")
)

// CreateVaultItemInput describes the parameters needed to create a new vault
// item.
type CreateVaultItemInput struct {
	// VaultID is an ID of the vault that holds this item.
	VaultID uuid.UUID
	// KeySalt is the salt used to derive the item encryption key.
	KeySalt []byte
	// ItemNonce is the nonce used to encrypt the item.
	ItemNonce []byte
	// Ciphertext contains the encrypted item data.
	Ciphertext []byte
}

// UpdateVaultItemInput describes the parameters needed to update vault item.
type UpdateVaultItemInput struct {
	// KeySalt is the new salt used to derive the item encryption key.
	KeySalt []byte
	// ItemNonce is the new nonce used to encrypt the item.
	ItemNonce []byte
	// Ciphertext contains the new encrypted item data.
	Ciphertext []byte
}

// VaultItemsRepository describes vault item related operations.
type VaultItemsRepository interface {
	// Create creates a new saves a new vault item.
	Create(
		ctx context.Context,
		userID uuid.UUID,
		input CreateVaultItemInput,
	) (*models.VaultItem, error)
	// GetByID returns a single vault item available to the given user.
	GetByID(
		ctx context.Context,
		userID uuid.UUID,
		vaultID uuid.UUID,
		itemID uuid.UUID,
	) (*models.VaultItem, error)
	// GetForVault returns all items associated with the given vault.
	GetForVault(
		ctx context.Context,
		userID uuid.UUID,
		vaultID uuid.UUID,
	) ([]models.VaultItem, error)
	// Update updates the given vault item with new data.
	Update(
		ctx context.Context,
		userID uuid.UUID,
		vaultID uuid.UUID,
		itemID uuid.UUID,
		input UpdateVaultItemInput,
	) (*models.VaultItem, error)
	// Delete deletes existing vault item.
	Delete(ctx context.Context, userID uuid.UUID, vaultID uuid.UUID, itemID uuid.UUID) error
}

type vaultItemsRepository struct {
	pool *pgxpool.Pool
	sq   squirrel.StatementBuilderType
}

// NewVaultItemsRepository creates a new vault items repository, backed by
// Postgres.
func NewVaultItemsRepository(pool *pgxpool.Pool) VaultItemsRepository {
	return &vaultItemsRepository{
		pool: pool,
		sq:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *vaultItemsRepository) Create(
	ctx context.Context,
	userID uuid.UUID,
	input CreateVaultItemInput,
) (*models.VaultItem, error) {
	selectBuilder := r.sq.
		Select("id").
		Column(squirrel.Expr("?", input.KeySalt)).
		Column(squirrel.Expr("?", input.ItemNonce)).
		Column(squirrel.Expr("?", input.Ciphertext)).
		From("vaults").
		Where(squirrel.Eq{"id": input.VaultID, "owner_id": userID})

	query, args, err := r.sq.Insert("vault_items").
		Columns("vault_id", "key_salt", "item_nonce", "ciphertext").
		Select(selectBuilder).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	vaultItem := models.VaultItem{
		VaultID:    input.VaultID,
		KeySalt:    input.KeySalt,
		ItemNonce:  input.ItemNonce,
		Ciphertext: input.Ciphertext,
	}

	err = r.pool.QueryRow(ctx, query, args...).
		Scan(&vaultItem.ID, &vaultItem.CreatedAt, &vaultItem.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVaultNotFound
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &vaultItem, nil
}

func (r *vaultItemsRepository) GetByID(
	ctx context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
	itemID uuid.UUID,
) (*models.VaultItem, error) {
	query, args, err := r.sq.
		Select(
			"vi.vault_id",
			"vi.key_salt",
			"vi.item_nonce",
			"vi.ciphertext",
			"vi.created_at",
			"vi.updated_at",
		).
		From("vault_items vi").
		Join("vaults v ON v.id = vi.vault_id").
		Where(squirrel.Eq{"vi.id": itemID, "vi.vault_id": vaultID, "v.owner_id": userID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	vaultItem := models.VaultItem{ID: itemID}
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&vaultItem.VaultID,
		&vaultItem.KeySalt,
		&vaultItem.ItemNonce,
		&vaultItem.Ciphertext,
		&vaultItem.CreatedAt,
		&vaultItem.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVaultItemNotFound
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &vaultItem, nil
}

func (r *vaultItemsRepository) GetForVault(
	ctx context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
) ([]models.VaultItem, error) {
	query, args, err := r.sq.
		Select(
			"vi.id",
			"vi.key_salt",
			"vi.item_nonce",
			"vi.ciphertext",
			"vi.created_at",
			"vi.updated_at",
		).
		From("vault_items vi").
		Join("vaults v ON v.id = vi.vault_id").
		Where(squirrel.Eq{"vi.vault_id": vaultID, "v.owner_id": userID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	vaultItems := make([]models.VaultItem, 0)
	for rows.Next() {
		vaultItem := models.VaultItem{VaultID: vaultID}
		err := rows.Scan(
			&vaultItem.ID,
			&vaultItem.KeySalt,
			&vaultItem.ItemNonce,
			&vaultItem.Ciphertext,
			&vaultItem.CreatedAt,
			&vaultItem.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		vaultItems = append(vaultItems, vaultItem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan rows: %w", err)
	}

	return vaultItems, nil
}

func (r *vaultItemsRepository) Update(
	ctx context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
	itemID uuid.UUID,
	input UpdateVaultItemInput,
) (*models.VaultItem, error) {
	query, args, err := r.sq.Update("vault_items").
		Set("key_salt", input.KeySalt).
		Set("item_nonce", input.ItemNonce).
		Set("ciphertext", input.Ciphertext).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": itemID, "vault_id": vaultID}).
		Where("vault_id IN (SELECT id FROM vaults WHERE owner_id = ?)", userID).
		Suffix("RETURNING vault_id, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	vaultItem := models.VaultItem{
		ID:         itemID,
		KeySalt:    input.KeySalt,
		ItemNonce:  input.ItemNonce,
		Ciphertext: input.Ciphertext,
	}

	err = r.pool.QueryRow(ctx, query, args...).
		Scan(&vaultItem.VaultID, &vaultItem.CreatedAt, &vaultItem.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVaultItemNotFound
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &vaultItem, nil
}

func (r *vaultItemsRepository) Delete(
	ctx context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
	itemID uuid.UUID,
) error {
	query, args, err := r.sq.Delete("vault_items").
		Where(squirrel.Eq{"id": itemID, "vault_id": vaultID}).
		Where("vault_id IN (SELECT id FROM vaults WHERE owner_id = ?)", userID).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	cmd, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return ErrVaultItemNotFound
	}

	return nil
}
