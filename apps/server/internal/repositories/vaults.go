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
	// ErrVaultNotFound is returned when no vault is found for the given
	// conditions.
	ErrVaultNotFound = errors.New("vault not found")
)

// CreateVaultInput contains the data needed to create a vault.
type CreateVaultInput struct {
	// OwnerID is owner user's ID.
	OwnerID uuid.UUID
	// Name is vault's visible name.
	Name string
}

// UpdateVaultInput contains the data needed to update a vault.
type UpdateVaultInput struct {
	// Name is new vault's visible name.
	Name *string
}

// VaultsRepository describes vault-related operations.
type VaultsRepository interface {
	// Create stores a new vault.
	Create(ctx context.Context, input CreateVaultInput) (*models.Vault, error)
	// GetForUser returns all vaults that are linked to provided user's ID.
	GetForUser(ctx context.Context, userID uuid.UUID) ([]models.Vault, error)
	// Update updates vault with the given ID with new data.
	Update(
		ctx context.Context,
		id uuid.UUID,
		input UpdateVaultInput,
	) (*models.Vault, error)
	// Delete deletes vault with the given ID.
	Delete(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error
}

type vaultsRepository struct {
	pool *pgxpool.Pool
	sq   squirrel.StatementBuilderType
}

// NewUsersRepository creates a vaults repository backed by PostgreSQL.
func NewVaultsRepository(pool *pgxpool.Pool) VaultsRepository {
	return &vaultsRepository{
		pool: pool,
		sq:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *vaultsRepository) Create(
	ctx context.Context,
	input CreateVaultInput,
) (*models.Vault, error) {
	query, args, err := r.sq.Insert("vaults").
		Columns("owner_id", "name").
		Values(input.OwnerID, input.Name).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	vault := models.Vault{
		OwnerID: input.OwnerID,
		Name:    input.Name,
	}

	err = r.pool.QueryRow(ctx, query, args...).
		Scan(&vault.ID, &vault.CreatedAt, &vault.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &vault, nil
}

func (r *vaultsRepository) GetForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]models.Vault, error) {
	query, args, err := r.sq.Select("id", "name", "created_at", "updated_at").
		From("vaults").
		Where(squirrel.Eq{"owner_id": userID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	vaults := make([]models.Vault, 0)
	for rows.Next() {
		vault := models.Vault{OwnerID: userID}
		err = rows.Scan(&vault.ID, &vault.Name, &vault.CreatedAt, &vault.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan the row: %w", err)
		}
		vaults = append(vaults, vault)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	return vaults, nil
}

func (r *vaultsRepository) Update(
	ctx context.Context,
	id uuid.UUID,
	input UpdateVaultInput,
) (*models.Vault, error) {
	builder := r.sq.Update("vaults")

	if input.Name != nil {
		builder = builder.Set("name", *input.Name)
	}

	query, args, err := builder.Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING owner_id, name, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	vault := models.Vault{ID: id}

	err = r.pool.QueryRow(ctx, query, args...).
		Scan(&vault.OwnerID, &vault.Name, &vault.CreatedAt, &vault.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVaultNotFound
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &vault, nil
}

func (r *vaultsRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
	ownerID uuid.UUID,
) error {
	query, args, err := r.sq.Delete("vaults").
		Where(squirrel.Eq{"id": id, "owner_id": ownerID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	cmd, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return ErrVaultNotFound
	}

	return nil
}
