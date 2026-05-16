package services

import (
	"context"
	"errors"
	"time"

	"github.com/Pelfox/gophkeeper/apps/server/internal/repositories"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

var (
	// ErrVaultCreationFailed is returned when service was not able to create a
	// new vault.
	ErrVaultCreationFailed = errors.New("failed to create a new vault")
	// ErrVaultsQueryFailed is returned when service was not able to query
	// user's vaults.
	ErrVaultsQueryFailed = errors.New("failed to query user's vaults")
	// ErrVaultNotFound is returned when request vault was not found.
	ErrVaultNotFound = errors.New("vault was not found")
	// ErrVaultDeletionFailed is returned when service was not able to delete
	// the vault.
	ErrVaultDeletionFailed = errors.New("failed to delete the vault")
)

// VaultResult describes a service-layer vault definition.
type VaultResult struct {
	// ID is a unique identifier for the vault.
	ID uuid.UUID
	// OwnerID is a unique identifier of the vault's owner.
	OwnerID uuid.UUID
	// Name is the name of the vault.
	Name string
	// CreatedAt is a timestamp when this vault was created.
	CreatedAt time.Time
	// UpdatedAt is a timestamp when this vault was last updated.
	UpdatedAt time.Time
}

// CreateVaultInput describes the parameters to create a new vault.
type CreateVaultInput struct {
	// OwnerID is an ID of the user that owns this vault.
	OwnerID uuid.UUID
	// Name is human-friendly name of the vault.
	Name string
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

// VaultsService describe all vault-related operations in a service.
type VaultsService interface {
	// Create creates a new vault and returns it.
	Create(ctx context.Context, input CreateVaultInput) (*VaultResult, error)
	// GetForUser returns all vaults associated with the given user.
	GetForUser(ctx context.Context, userID uuid.UUID) ([]VaultResult, error)
	// Delete deletes vault with the given ID.
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type vaultsService struct {
	usersRepository    repositories.UsersRepository
	vaultsRepository   repositories.VaultsRepository
	keyringsRepository repositories.KeyringsRepository
	logger             zerolog.Logger
}

// NewVaultsService creates a new vaults service, backed by a vaults repository.
func NewVaultsService(
	usersRepository repositories.UsersRepository,
	vaultsRepository repositories.VaultsRepository,
	keyringsRepository repositories.KeyringsRepository,
	logger zerolog.Logger,
) VaultsService {
	return &vaultsService{
		usersRepository:    usersRepository,
		vaultsRepository:   vaultsRepository,
		keyringsRepository: keyringsRepository,
		logger:             logger.With().Str("service", "vaults").Logger(),
	}
}

func (s *vaultsService) Create(
	ctx context.Context,
	input CreateVaultInput,
) (*VaultResult, error) {
	vault, err := s.vaultsRepository.Create(ctx, repositories.CreateVaultInput{
		OwnerID: input.OwnerID,
		Name:    input.Name,
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create a new vault")
		return nil, ErrVaultCreationFailed
	}

	_, err = s.keyringsRepository.Create(ctx, repositories.CreateKeyringInput{
		UserID:                input.OwnerID,
		VaultID:               vault.ID,
		EncryptionSalt:        input.EncryptionSalt,
		EncryptionNonce:       input.EncryptionNonce,
		EncryptedMasterKey:    input.EncryptedMasterKey,
		EncryptionTimeCost:    input.EncryptionTimeCost,
		EncryptionMemoryCost:  input.EncryptionMemoryCost,
		EncryptionParallelism: input.EncryptionParallelism,
		EncryptionKeySize:     input.EncryptionKeySize,
	})
	if err != nil {
		s.logger.Error().Err(err).
			Str("vault_id", vault.ID.String()).
			Msg("failed to create vault's keyring")
		return nil, ErrVaultCreationFailed
	}

	result := VaultResult{
		ID:        vault.ID,
		OwnerID:   vault.OwnerID,
		Name:      vault.Name,
		CreatedAt: vault.CreatedAt,
		UpdatedAt: vault.UpdatedAt,
	}
	return &result, nil
}

func (s *vaultsService) GetForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]VaultResult, error) {
	rawVaults, err := s.vaultsRepository.GetForUser(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to query user's vaults")
		return nil, ErrVaultsQueryFailed
	}

	vaults := make([]VaultResult, len(rawVaults))
	for i, vault := range rawVaults {
		vaults[i] = VaultResult{
			ID:        vault.ID,
			OwnerID:   vault.OwnerID,
			Name:      vault.Name,
			CreatedAt: vault.CreatedAt,
			UpdatedAt: vault.UpdatedAt,
		}
	}

	return vaults, nil
}

func (s *vaultsService) Delete(
	ctx context.Context,
	id uuid.UUID,
	userID uuid.UUID,
) error {
	if err := s.vaultsRepository.Delete(ctx, id, userID); err != nil {
		if errors.Is(err, repositories.ErrVaultNotFound) {
			return ErrVaultNotFound
		}

		s.logger.Error().Err(err).
			Str("vault_id", id.String()).
			Msg("failed to delete the vault")
		return ErrVaultDeletionFailed
	}
	return nil
}
