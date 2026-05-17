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
	// ErrVaultItemCreationFailed is returned when service was not able to
	// create a new vault item.
	ErrVaultItemCreationFailed = errors.New("failed to create vault item")
	// ErrVaultItemUpdateFailed is returned when service was not able to update
	// vault item.
	ErrVaultItemUpdateFailed = errors.New("failed to update vault item")
	// ErrVaultItemNotFound is returned when requested vault item was not found.
	ErrVaultItemNotFound = errors.New("vault item was not found")
	// ErrVaultItemsQueryFailed is returned when service was not able to query
	// vault items.
	ErrVaultItemsQueryFailed = errors.New("failed to query vault items")
	// ErrVaultItemDeletionFailed is returned when service was not able to
	// delete vault item.
	ErrVaultItemDeletionFailed = errors.New("failed to delete vault item")
)

// VaultItemResult describes the result of operation with vault item.
type VaultItemResult struct {
	// ID is vault's item id.
	ID uuid.UUID `json:"id"`
	// VaultID is parent vault's ID.
	VaultID uuid.UUID `json:"vault_id"`
	// KeySalt is the salt used to derive the item encryption key.
	KeySalt []byte
	// ItemNonce is the nonce used to encrypt the item.
	ItemNonce []byte
	// Ciphertext contains the encrypted item data.
	Ciphertext []byte
	// CreatedAt is the time the item was created.
	CreatedAt time.Time
	// UpdatedAt is the time the item was last updated.
	UpdatedAt time.Time
}

// VaultItemInput describes generic parameters needed for performing operations
// with vault item.
type VaultItemInput struct {
	// KeySalt is the salt used to derive the item encryption key.
	KeySalt []byte
	// ItemNonce is the nonce used to encrypt the item.
	ItemNonce []byte
	// Ciphertext contains the encrypted item data.
	Ciphertext []byte
}

// VaultItemsService describes all vault items related operations.
type VaultItemsService interface {
	// Create creates a new vault item.
	Create(
		ctx context.Context,
		userID uuid.UUID,
		vaultID uuid.UUID,
		input VaultItemInput,
	) (*VaultItemResult, error)
	// Update updates given vault item with a new one.
	Update(
		ctx context.Context,
		userID uuid.UUID,
		vaultID uuid.UUID,
		itemID uuid.UUID,
		input VaultItemInput,
	) (*VaultItemResult, error)
	// GetByID returns vault item with the given ID.
	GetByID(
		ctx context.Context,
		userID uuid.UUID,
		vaultID uuid.UUID,
		id uuid.UUID,
	) (*VaultItemResult, error)
	// GetForVault returns all vault items.
	GetForVault(
		ctx context.Context,
		userID uuid.UUID,
		id uuid.UUID,
	) ([]VaultItemResult, error)
	// Delete deletes a single vault item.
	Delete(ctx context.Context, userID uuid.UUID, vaultID uuid.UUID, id uuid.UUID) error
}

type vaultItemsService struct {
	vaultItemsRepository repositories.VaultItemsRepository
	logger               zerolog.Logger
}

// NewVaultItemsService creates a new vault items service, backed by a vault
// items repository.
func NewVaultItemsService(
	vaultItemsRepository repositories.VaultItemsRepository,
	logger zerolog.Logger,
) VaultItemsService {
	return &vaultItemsService{
		vaultItemsRepository: vaultItemsRepository,
		logger:               logger.With().Str("service", "vault_items").Logger(),
	}
}

func (s *vaultItemsService) Create(
	ctx context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
	input VaultItemInput,
) (*VaultItemResult, error) {
	vaultItem, err := s.vaultItemsRepository.Create(ctx, userID, repositories.CreateVaultItemInput{
		VaultID:    vaultID,
		KeySalt:    input.KeySalt,
		ItemNonce:  input.ItemNonce,
		Ciphertext: input.Ciphertext,
	})
	if err != nil {
		if errors.Is(err, repositories.ErrVaultNotFound) {
			return nil, ErrVaultNotFound
		}

		s.logger.Error().Err(err).
			Str("vault_id", vaultID.String()).
			Msg("failed to create vault item")
		return nil, ErrVaultItemCreationFailed
	}

	return &VaultItemResult{
		ID:         vaultItem.ID,
		VaultID:    vaultItem.VaultID,
		KeySalt:    vaultItem.KeySalt,
		ItemNonce:  vaultItem.ItemNonce,
		Ciphertext: vaultItem.Ciphertext,
		CreatedAt:  vaultItem.CreatedAt,
		UpdatedAt:  vaultItem.UpdatedAt,
	}, nil
}

func (s *vaultItemsService) Update(
	ctx context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
	itemID uuid.UUID,
	input VaultItemInput,
) (*VaultItemResult, error) {
	vaultItem, err := s.vaultItemsRepository.Update(ctx, userID, vaultID, itemID, repositories.UpdateVaultItemInput{
		KeySalt:    input.KeySalt,
		ItemNonce:  input.ItemNonce,
		Ciphertext: input.Ciphertext,
	})
	if err != nil {
		if errors.Is(err, repositories.ErrVaultItemNotFound) {
			return nil, ErrVaultItemNotFound
		}

		s.logger.Error().Err(err).
			Str("vault_item_id", itemID.String()).
			Msg("failed to update vault item")
		return nil, ErrVaultItemUpdateFailed
	}

	return &VaultItemResult{
		ID:         vaultItem.ID,
		VaultID:    vaultItem.VaultID,
		KeySalt:    vaultItem.KeySalt,
		ItemNonce:  vaultItem.ItemNonce,
		Ciphertext: vaultItem.Ciphertext,
		CreatedAt:  vaultItem.CreatedAt,
		UpdatedAt:  vaultItem.UpdatedAt,
	}, nil
}

func (s *vaultItemsService) GetByID(
	ctx context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
	id uuid.UUID,
) (*VaultItemResult, error) {
	vaultItem, err := s.vaultItemsRepository.GetByID(ctx, userID, vaultID, id)
	if err != nil {
		if errors.Is(err, repositories.ErrVaultItemNotFound) {
			return nil, ErrVaultItemNotFound
		}

		s.logger.Error().Err(err).
			Str("vault_item_id", id.String()).
			Msg("failed to get vault item")
		return nil, ErrVaultItemsQueryFailed
	}

	return &VaultItemResult{
		ID:         vaultItem.ID,
		VaultID:    vaultItem.VaultID,
		KeySalt:    vaultItem.KeySalt,
		ItemNonce:  vaultItem.ItemNonce,
		Ciphertext: vaultItem.Ciphertext,
		CreatedAt:  vaultItem.CreatedAt,
		UpdatedAt:  vaultItem.UpdatedAt,
	}, nil
}

func (s *vaultItemsService) GetForVault(
	ctx context.Context,
	userID uuid.UUID,
	id uuid.UUID,
) ([]VaultItemResult, error) {
	rawVaultItems, err := s.vaultItemsRepository.GetForVault(ctx, userID, id)
	if err != nil {
		s.logger.Error().Err(err).
			Str("vault_id", id.String()).
			Msg("failed to query vault items")
		return nil, ErrVaultItemsQueryFailed
	}

	vaultItems := make([]VaultItemResult, len(rawVaultItems))
	for i, vaultItem := range rawVaultItems {
		vaultItems[i] = VaultItemResult{
			ID:         vaultItem.ID,
			VaultID:    vaultItem.VaultID,
			KeySalt:    vaultItem.KeySalt,
			ItemNonce:  vaultItem.ItemNonce,
			Ciphertext: vaultItem.Ciphertext,
			CreatedAt:  vaultItem.CreatedAt,
			UpdatedAt:  vaultItem.UpdatedAt,
		}
	}

	return vaultItems, nil
}

func (s *vaultItemsService) Delete(ctx context.Context, userID uuid.UUID, vaultID uuid.UUID, id uuid.UUID) error {
	if err := s.vaultItemsRepository.Delete(ctx, userID, vaultID, id); err != nil {
		if errors.Is(err, repositories.ErrVaultItemNotFound) {
			return ErrVaultItemNotFound
		}

		s.logger.Error().Err(err).
			Str("vault_item_id", id.String()).
			Msg("failed to delete vault item")
		return ErrVaultItemDeletionFailed
	}

	return nil
}
