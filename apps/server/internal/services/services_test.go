package services

import (
	"context"
	"time"

	"github.com/Pelfox/gophkeeper/apps/server/internal/models"
	"github.com/Pelfox/gophkeeper/apps/server/internal/repositories"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func testLogger() zerolog.Logger {
	return zerolog.Nop()
}

func fixedTime() time.Time {
	return time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC)
}

type fakeUsersRepository struct {
	createCalls int
	createInput repositories.CreateUserInput
	createUser  *models.User
	createErr   error

	findByEmailCalls int
	findByEmailEmail string
	findByEmailUser  *models.User
	findByEmailErr   error

	getByIDCalls int
	getByIDID    uuid.UUID
	getByIDUser  *models.User
	getByIDErr   error

	updateCalls int
	updateID    uuid.UUID
	updateInput repositories.UpdateUserInput
	updateUser  *models.User
	updateErr   error
}

func (r *fakeUsersRepository) Create(
	_ context.Context,
	input repositories.CreateUserInput,
) (*models.User, error) {
	r.createCalls++
	r.createInput = input
	return r.createUser, r.createErr
}

func (r *fakeUsersRepository) FindByEmail(
	_ context.Context,
	email string,
) (*models.User, error) {
	r.findByEmailCalls++
	r.findByEmailEmail = email
	return r.findByEmailUser, r.findByEmailErr
}

func (r *fakeUsersRepository) GetByID(
	_ context.Context,
	id uuid.UUID,
) (*models.User, error) {
	r.getByIDCalls++
	r.getByIDID = id
	return r.getByIDUser, r.getByIDErr
}

func (r *fakeUsersRepository) Update(
	_ context.Context,
	id uuid.UUID,
	input repositories.UpdateUserInput,
) (*models.User, error) {
	r.updateCalls++
	r.updateID = id
	r.updateInput = input
	return r.updateUser, r.updateErr
}

type fakeSessionsRepository struct {
	createCalls int
	createInput repositories.CreateSessionInput
	createItem  *models.Session
	createErr   error

	getForUserCalls int
	getForUserID    uuid.UUID
	getForUserItems []models.Session
	getForUserErr   error

	findByAccessTokenCalls int
	findByAccessTokenValue string
	findByAccessTokenItem  *models.Session
	findByAccessTokenErr   error
}

func (r *fakeSessionsRepository) Create(
	_ context.Context,
	input repositories.CreateSessionInput,
) (*models.Session, error) {
	r.createCalls++
	r.createInput = input
	if r.createErr != nil {
		return nil, r.createErr
	}
	if r.createItem != nil {
		return r.createItem, nil
	}
	return &models.Session{
		ID:          uuid.New(),
		UserID:      input.UserID,
		AccessToken: input.AccessToken,
		ExpiresAt:   input.ExpiresAt,
		CreatedAt:   fixedTime(),
	}, nil
}

func (r *fakeSessionsRepository) GetForUser(
	_ context.Context,
	userID uuid.UUID,
) ([]models.Session, error) {
	r.getForUserCalls++
	r.getForUserID = userID
	return r.getForUserItems, r.getForUserErr
}

func (r *fakeSessionsRepository) FindByAccessToken(
	_ context.Context,
	accessToken string,
) (*models.Session, error) {
	r.findByAccessTokenCalls++
	r.findByAccessTokenValue = accessToken
	return r.findByAccessTokenItem, r.findByAccessTokenErr
}

type fakeVaultsRepository struct {
	createCalls int
	createInput repositories.CreateVaultInput
	createItem  *models.Vault
	createErr   error

	getForUserCalls int
	getForUserID    uuid.UUID
	getForUserItems []models.Vault
	getForUserErr   error

	updateCalls int
	updateID    uuid.UUID
	updateOwner uuid.UUID
	updateInput repositories.UpdateVaultInput
	updateItem  *models.Vault
	updateErr   error

	deleteCalls int
	deleteID    uuid.UUID
	deleteOwner uuid.UUID
	deleteErr   error
}

func (r *fakeVaultsRepository) Create(
	_ context.Context,
	input repositories.CreateVaultInput,
) (*models.Vault, error) {
	r.createCalls++
	r.createInput = input
	return r.createItem, r.createErr
}

func (r *fakeVaultsRepository) GetForUser(
	_ context.Context,
	userID uuid.UUID,
) ([]models.Vault, error) {
	r.getForUserCalls++
	r.getForUserID = userID
	return r.getForUserItems, r.getForUserErr
}

func (r *fakeVaultsRepository) Update(
	_ context.Context,
	id uuid.UUID,
	ownerID uuid.UUID,
	input repositories.UpdateVaultInput,
) (*models.Vault, error) {
	r.updateCalls++
	r.updateID = id
	r.updateOwner = ownerID
	r.updateInput = input
	return r.updateItem, r.updateErr
}

func (r *fakeVaultsRepository) Delete(
	_ context.Context,
	id uuid.UUID,
	ownerID uuid.UUID,
) error {
	r.deleteCalls++
	r.deleteID = id
	r.deleteOwner = ownerID
	return r.deleteErr
}

type fakeKeyringsRepository struct {
	createCalls int
	createInput repositories.CreateKeyringInput
	createItem  *models.Keyring
	createErr   error

	getCalls int
	getUser  uuid.UUID
	getVault uuid.UUID
	getItem  *models.Keyring
	getErr   error

	deleteCalls int
	deleteUser  uuid.UUID
	deleteVault uuid.UUID
	deleteErr   error
}

func (r *fakeKeyringsRepository) Create(
	_ context.Context,
	input repositories.CreateKeyringInput,
) (*models.Keyring, error) {
	r.createCalls++
	r.createInput = input
	if r.createErr != nil {
		return nil, r.createErr
	}
	if r.createItem != nil {
		return r.createItem, nil
	}
	return &models.Keyring{
		UserID:                input.UserID,
		VaultID:               input.VaultID,
		EncryptionSalt:        input.EncryptionSalt,
		EncryptionNonce:       input.EncryptionNonce,
		EncryptedMasterKey:    input.EncryptedMasterKey,
		EncryptionTimeCost:    input.EncryptionTimeCost,
		EncryptionMemoryCost:  input.EncryptionMemoryCost,
		EncryptionParallelism: input.EncryptionParallelism,
		EncryptionKeySize:     input.EncryptionKeySize,
		CreatedAt:             fixedTime(),
		UpdatedAt:             fixedTime(),
	}, nil
}

func (r *fakeKeyringsRepository) Get(
	_ context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
) (*models.Keyring, error) {
	r.getCalls++
	r.getUser = userID
	r.getVault = vaultID
	return r.getItem, r.getErr
}

func (r *fakeKeyringsRepository) Delete(
	_ context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
) error {
	r.deleteCalls++
	r.deleteUser = userID
	r.deleteVault = vaultID
	return r.deleteErr
}

type fakeVaultItemsRepository struct {
	createCalls int
	createUser  uuid.UUID
	createInput repositories.CreateVaultItemInput
	createItem  *models.VaultItem
	createErr   error

	getByIDCalls int
	getByIDUser  uuid.UUID
	getByIDVault uuid.UUID
	getByIDItem  uuid.UUID
	getByIDValue *models.VaultItem
	getByIDErr   error

	getForVaultCalls int
	getForVaultUser  uuid.UUID
	getForVaultVault uuid.UUID
	getForVaultItems []models.VaultItem
	getForVaultErr   error

	updateCalls int
	updateUser  uuid.UUID
	updateVault uuid.UUID
	updateItem  uuid.UUID
	updateInput repositories.UpdateVaultItemInput
	updateValue *models.VaultItem
	updateErr   error

	deleteCalls int
	deleteUser  uuid.UUID
	deleteVault uuid.UUID
	deleteItem  uuid.UUID
	deleteErr   error
}

func (r *fakeVaultItemsRepository) Create(
	_ context.Context,
	userID uuid.UUID,
	input repositories.CreateVaultItemInput,
) (*models.VaultItem, error) {
	r.createCalls++
	r.createUser = userID
	r.createInput = input
	return r.createItem, r.createErr
}

func (r *fakeVaultItemsRepository) GetByID(
	_ context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
	itemID uuid.UUID,
) (*models.VaultItem, error) {
	r.getByIDCalls++
	r.getByIDUser = userID
	r.getByIDVault = vaultID
	r.getByIDItem = itemID
	return r.getByIDValue, r.getByIDErr
}

func (r *fakeVaultItemsRepository) GetForVault(
	_ context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
) ([]models.VaultItem, error) {
	r.getForVaultCalls++
	r.getForVaultUser = userID
	r.getForVaultVault = vaultID
	return r.getForVaultItems, r.getForVaultErr
}

func (r *fakeVaultItemsRepository) Update(
	_ context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
	itemID uuid.UUID,
	input repositories.UpdateVaultItemInput,
) (*models.VaultItem, error) {
	r.updateCalls++
	r.updateUser = userID
	r.updateVault = vaultID
	r.updateItem = itemID
	r.updateInput = input
	return r.updateValue, r.updateErr
}

func (r *fakeVaultItemsRepository) Delete(
	_ context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
	itemID uuid.UUID,
) error {
	r.deleteCalls++
	r.deleteUser = userID
	r.deleteVault = vaultID
	r.deleteItem = itemID
	return r.deleteErr
}
