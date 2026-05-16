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
	// ErrKeyringNotFound is returned when no keyring was found for the given
	// conditions.
	ErrKeyringNotFound = errors.New("keyring not found")
)

// CreateKeyringInput describes the input parameters to create a new keyring.
type CreateKeyringInput struct {
	// UserID is an ID of the user that owns this keyring.
	UserID uuid.UUID
	// VaultID is an ID of the vault to connect this keyring to.
	VaultID uuid.UUID
	// EncryptionSalt holds the raw bytes that were used for the encryption of
	// master key.
	EncryptionSalt []byte
	// EncryptionNonce holds the raw bytes that were used by the encryption
	// algorithm to encrypt master key.
	EncryptionNonce []byte
	// EncryptedMasterKey holds an actual encrypted master key. This value is
	// safe to be stored in the database.
	EncryptedMasterKey []byte
	// EncryptionTimeCost describes amount of passes of the given
	// EncryptionMemoryCost.
	EncryptionTimeCost uint32
	// EncryptionMemoryCost describes how much memory should be used.
	EncryptionMemoryCost uint32
	// EncryptionParallelism describes how much threads should be used.
	EncryptionParallelism uint32
	// EncryptionKeySize describes the size of the returned byte slice.
	EncryptionKeySize uint32
}

// KeyringsRepository describes keyring-related operations.
type KeyringsRepository interface {
	// Create creates and saves a new keyring into the database.
	Create(
		ctx context.Context,
		input CreateKeyringInput,
	) (*models.Keyring, error)
	// Get returns found keyring (if any) for given user ID and vault ID.
	Get(
		ctx context.Context,
		userID uuid.UUID,
		vaultID uuid.UUID,
	) (*models.Keyring, error)
	// Delete deletes given keyring (if any).
	Delete(ctx context.Context, userID uuid.UUID, vaultID uuid.UUID) error
}

type keyringsRepository struct {
	pool *pgxpool.Pool
	sq   squirrel.StatementBuilderType
}

// NewKeyringsRepository creates a new keyrings repository, backed by Postgres.
func NewKeyringsRepository(pool *pgxpool.Pool) KeyringsRepository {
	return &keyringsRepository{
		pool: pool,
		sq:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *keyringsRepository) Create(
	ctx context.Context,
	input CreateKeyringInput,
) (*models.Keyring, error) {
	query, args, err := r.sq.Insert("keyrings").
		Columns(
			"user_id",
			"vault_id",
			"encryption_salt",
			"encryption_nonce",
			"encrypted_master_key",
			"encryption_time_cost",
			"encryption_memory_cost",
			"encryption_parallelism",
			"encryption_key_size",
		).
		Values(
			input.UserID,
			input.VaultID,
			input.EncryptionSalt,
			input.EncryptionNonce,
			input.EncryptedMasterKey,
			input.EncryptionTimeCost,
			input.EncryptionMemoryCost,
			input.EncryptionParallelism,
			input.EncryptionKeySize,
		).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	keyring := models.Keyring{
		UserID:                input.UserID,
		VaultID:               input.VaultID,
		EncryptionSalt:        input.EncryptionSalt,
		EncryptionNonce:       input.EncryptionNonce,
		EncryptedMasterKey:    input.EncryptedMasterKey,
		EncryptionTimeCost:    input.EncryptionTimeCost,
		EncryptionMemoryCost:  input.EncryptionMemoryCost,
		EncryptionParallelism: input.EncryptionParallelism,
		EncryptionKeySize:     input.EncryptionKeySize,
	}

	err = r.pool.QueryRow(ctx, query, args...).
		Scan(&keyring.CreatedAt, &keyring.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &keyring, nil
}

func (r *keyringsRepository) Get(
	ctx context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
) (*models.Keyring, error) {
	query, args, err := r.sq.
		Select(
			"encryption_salt",
			"encryption_nonce",
			"encrypted_master_key",
			"encryption_time_cost",
			"encryption_memory_cost",
			"encryption_parallelism",
			"encryption_key_size",
			"created_at",
			"updated_at",
		).
		From("keyrings").
		Where(squirrel.Eq{"user_id": userID, "vault_id": vaultID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	keyring := models.Keyring{
		UserID:  userID,
		VaultID: vaultID,
	}

	err = r.pool.QueryRow(ctx, query, args...).
		Scan(
			&keyring.EncryptionSalt,
			&keyring.EncryptionNonce,
			&keyring.EncryptedMasterKey,
			&keyring.EncryptionTimeCost,
			&keyring.EncryptionMemoryCost,
			&keyring.EncryptionParallelism,
			&keyring.EncryptionKeySize,
			&keyring.CreatedAt,
			&keyring.UpdatedAt,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrKeyringNotFound
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &keyring, nil
}

func (r *keyringsRepository) Delete(
	ctx context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
) error {
	query, args, err := r.sq.Delete("keyrings").
		Where(squirrel.Eq{"user_id": userID, "vault_id": vaultID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	cmd, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return ErrKeyringNotFound
	}

	return nil
}
