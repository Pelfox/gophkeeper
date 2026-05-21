package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Pelfox/gophkeeper/apps/client/internal/crypto"
	"github.com/Pelfox/gophkeeper/shared/protocol"
	"github.com/google/uuid"
)

// PlaintextVaultItemType describes the kind of item stored in the vault.
type PlaintextVaultItemType string

const (
	// PlaintextVaultItemTypePassword stores a password.
	PlaintextVaultItemTypePassword PlaintextVaultItemType = "password"
	// PlaintextVaultItemTypeTextNote stores a text note.
	PlaintextVaultItemTypeTextNote PlaintextVaultItemType = "text_note"
	// PlaintextVaultItemTypeBinaryFile stores binary file data.
	PlaintextVaultItemTypeBinaryFile PlaintextVaultItemType = "binary_file"
	// PlaintextVaultItemTypeBankCard stores bank card details.
	PlaintextVaultItemTypeBankCard PlaintextVaultItemType = "bank_card"
)

var (
	// ErrInvalidVaultItemType is returned when plaintext vault item type is
	// unknown.
	ErrInvalidVaultItemType = errors.New("invalid vault item type")
)

// PlaintextVaultItem describes the JSON document encrypted before being sent
// to the server.
type PlaintextVaultItem struct {
	// Type describes the kind of item stored in the vault.
	Type PlaintextVaultItemType `json:"type"`
	// Name is the human-friendly item name.
	Name string `json:"name"`
	// Payload contains item-specific fields.
	Payload any `json:"payload"`
}

// VaultItem describes decrypted vault item data along with server metadata.
type VaultItem struct {
	// ID is vault item's ID.
	ID uuid.UUID
	// VaultID is parent vault's ID.
	VaultID uuid.UUID
	// CreatedAt is the time the item was created.
	CreatedAt time.Time
	// UpdatedAt is the time the item was last updated.
	UpdatedAt time.Time
	// Plaintext contains decrypted item data.
	Plaintext PlaintextVaultItem
}

// PasswordPayload describes a stored password.
type PasswordPayload struct {
	// Password is the password to store.
	Password string `json:"password"`
	// Website is an optional website associated with the password.
	Website *string `json:"website,omitempty"`
}

// TextNotePayload describes stored plain text note.
type TextNotePayload struct {
	// Text is the note text to store.
	Text string `json:"text"`
}

// BinaryFilePayload describes stored binary file contents.
type BinaryFilePayload struct {
	// Path is the original file path.
	Path string `json:"path"`
	// Data is the file contents. It is encoded as base64 in JSON.
	Data []byte `json:"data"`
}

// BankCardPayload describes stored bank card details.
type BankCardPayload struct {
	// Number is the bank card number.
	Number string `json:"number"`
	// CVV is the bank card security code.
	CVV string `json:"cvv"`
	// ExpirationDate is the card expiration date.
	ExpirationDate string `json:"expiration_date"`
	// HolderName is the card holder's name.
	HolderName string `json:"holder_name"`
}

type plaintextPayloadDecoderType = func(json.RawMessage) (any, error)

var plaintextVaultItemPayloadDecoders = map[PlaintextVaultItemType]plaintextPayloadDecoderType{
	PlaintextVaultItemTypePassword:   decodeVaultItemPayload[PasswordPayload],
	PlaintextVaultItemTypeTextNote:   decodeVaultItemPayload[TextNotePayload],
	PlaintextVaultItemTypeBinaryFile: decodeVaultItemPayload[BinaryFilePayload],
	PlaintextVaultItemTypeBankCard:   decodeVaultItemPayload[BankCardPayload],
}

// Valid indicates whether plaintext vault item type is supported.
func (t PlaintextVaultItemType) Valid() bool {
	_, ok := plaintextVaultItemPayloadDecoders[t]
	return ok
}

// CreateVaultItem encrypts plaintext item data locally and stores encrypted
// item material on the server.
func (a *App) CreateVaultItem(
	ctx context.Context,
	vault protocol.ProtocolVault,
	vaultPassword string,
	item PlaintextVaultItem,
) (*protocol.CreateVaultItemResponse, error) {
	if !item.Type.Valid() {
		return nil, ErrInvalidVaultItemType
	}

	masterKey, err := a.deriveVaultMasterKey(ctx, vault, vaultPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to unlock vault: %w", err)
	}

	encryptedItem, err := encryptPlaintextVaultItem(masterKey, item)
	if err != nil {
		return nil, err
	}

	resp, err := a.Client.PostVaultsIdItemsWithResponse(ctx, vault.ID.String(), protocol.CreateVaultItemRequest{
		KeySalt:    encryptedItem.KeySalt,
		ItemNonce:  encryptedItem.ItemNonce,
		Ciphertext: encryptedItem.Ciphertext,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to perform vault item creation request: %w", err)
	}

	if resp.JSON201 != nil {
		return resp.JSON201, nil
	}

	if resp.JSON400 != nil {
		return nil, fmt.Errorf("vault item creation failed: %s", resp.JSON400.Message)
	}

	if resp.JSON401 != nil {
		return nil, fmt.Errorf("unauthorized: %s", resp.JSON401.Message)
	}

	if resp.JSON404 != nil {
		return nil, fmt.Errorf("vault not found: %s", resp.JSON404.Message)
	}

	if resp.JSON422 != nil {
		return nil, fmt.Errorf("invalid request: %s", resp.JSON422.Message)
	}

	if resp.JSON500 != nil {
		return nil, fmt.Errorf("server error: %s", resp.JSON500.Message)
	}

	return nil, fmt.Errorf("unexpected response from server: %s", resp.Status())
}

// UpdateVaultItem encrypts plaintext item data locally and updates existing
// encrypted item material on the server.
func (a *App) UpdateVaultItem(
	ctx context.Context,
	vault protocol.ProtocolVault,
	vaultPassword string,
	item VaultItem,
	plaintextItem PlaintextVaultItem,
) (*protocol.UpdateVaultItemResponse, error) {
	if !plaintextItem.Type.Valid() {
		return nil, ErrInvalidVaultItemType
	}

	masterKey, err := a.deriveVaultMasterKey(ctx, vault, vaultPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to unlock vault: %w", err)
	}

	encryptedItem, err := encryptPlaintextVaultItem(masterKey, plaintextItem)
	if err != nil {
		return nil, err
	}

	resp, err := a.Client.PatchVaultsIdItemsItemIdWithResponse(
		ctx,
		vault.ID.String(),
		item.ID.String(),
		protocol.UpdateVaultItemRequest{
			KeySalt:    encryptedItem.KeySalt,
			ItemNonce:  encryptedItem.ItemNonce,
			Ciphertext: encryptedItem.Ciphertext,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to perform vault item update request: %w", err)
	}

	if resp.JSON200 != nil {
		return resp.JSON200, nil
	}

	if resp.JSON400 != nil {
		return nil, fmt.Errorf("vault item update failed: %s", resp.JSON400.Message)
	}

	if resp.JSON401 != nil {
		return nil, fmt.Errorf("unauthorized: %s", resp.JSON401.Message)
	}

	if resp.JSON404 != nil {
		return nil, fmt.Errorf("vault item not found: %s", resp.JSON404.Message)
	}

	if resp.JSON422 != nil {
		return nil, fmt.Errorf("invalid request: %s", resp.JSON422.Message)
	}

	if resp.JSON500 != nil {
		return nil, fmt.Errorf("server error: %s", resp.JSON500.Message)
	}

	return nil, fmt.Errorf("unexpected response from server: %s", resp.Status())
}

// DeleteVaultItem removes a vault item from the server.
func (a *App) DeleteVaultItem(
	ctx context.Context,
	vault protocol.ProtocolVault,
	item VaultItem,
) error {
	resp, err := a.Client.DeleteVaultsIdItemsItemIdWithResponse(
		ctx,
		vault.ID.String(),
		item.ID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to perform vault item deletion request: %w", err)
	}

	if resp.StatusCode() == http.StatusNoContent {
		return nil
	}

	if resp.JSON400 != nil {
		return fmt.Errorf("vault item deletion failed: %s", resp.JSON400.Message)
	}

	if resp.JSON401 != nil {
		return fmt.Errorf("unauthorized: %s", resp.JSON401.Message)
	}

	if resp.JSON404 != nil {
		return fmt.Errorf("vault item not found: %s", resp.JSON404.Message)
	}

	return fmt.Errorf("unexpected response from server: %s", resp.Status())
}

// ListVaultItems retrieves encrypted vault items from the server and decrypts
// them locally.
func (a *App) ListVaultItems(
	ctx context.Context,
	vault protocol.ProtocolVault,
	vaultPassword string,
) ([]VaultItem, error) {
	masterKey, err := a.deriveVaultMasterKey(ctx, vault, vaultPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to unlock vault: %w", err)
	}

	resp, err := a.Client.GetVaultsIdItemsWithResponse(ctx, vault.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to perform vault items list request: %w", err)
	}

	if resp.JSON400 != nil {
		return nil, fmt.Errorf("invalid vault: %s", resp.JSON400.Message)
	}

	if resp.JSON401 != nil {
		return nil, fmt.Errorf("unauthorized: %s", resp.JSON401.Message)
	}

	if resp.JSON500 != nil {
		return nil, fmt.Errorf("server error: %s", resp.JSON500.Message)
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response from server: %s", resp.Status())
	}

	items := make([]VaultItem, 0, len(*resp.JSON200))
	for _, encryptedItem := range *resp.JSON200 {
		plaintext, err := crypto.DecryptItem(crypto.DecryptItemParameters{
			Ciphertext: encryptedItem.Ciphertext,
			ItemNonce:  encryptedItem.ItemNonce,
			KeySalt:    encryptedItem.KeySalt,
			MasterKey:  masterKey,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt vault item %q: %w", encryptedItem.ID, err)
		}

		item, err := decodePlaintextVaultItem(plaintext)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal vault item %q: %w", encryptedItem.ID, err)
		}

		items = append(items, VaultItem{
			ID:        encryptedItem.ID,
			VaultID:   encryptedItem.VaultID,
			CreatedAt: encryptedItem.CreatedAt,
			UpdatedAt: encryptedItem.UpdatedAt,
			Plaintext: item,
		})
	}

	return items, nil
}

func decodePlaintextVaultItem(plaintext []byte) (PlaintextVaultItem, error) {
	var rawItem struct {
		Type    PlaintextVaultItemType `json:"type"`
		Name    string                 `json:"name"`
		Payload json.RawMessage        `json:"payload"`
	}
	if err := json.Unmarshal(plaintext, &rawItem); err != nil {
		return PlaintextVaultItem{}, err
	}

	decoderFunc, ok := plaintextVaultItemPayloadDecoders[rawItem.Type]
	if !ok {
		return PlaintextVaultItem{}, ErrInvalidVaultItemType
	}

	payload, err := decoderFunc(rawItem.Payload)
	if err != nil {
		return PlaintextVaultItem{}, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return PlaintextVaultItem{
		Type:    rawItem.Type,
		Name:    rawItem.Name,
		Payload: payload,
	}, nil
}

func decodeVaultItemPayload[T any](rawPayload json.RawMessage) (any, error) {
	var payload T
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (a *App) deriveVaultMasterKey(
	ctx context.Context,
	vault protocol.ProtocolVault,
	vaultPassword string,
) ([]byte, error) {
	keyringResp, err := a.Client.GetVaultsIdKeyringWithResponse(ctx, vault.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to perform vault keyring request: %w", err)
	}

	if keyringResp.JSON400 != nil {
		return nil, fmt.Errorf("invalid vault: %s", keyringResp.JSON400.Message)
	}

	if keyringResp.JSON401 != nil {
		return nil, fmt.Errorf("unauthorized: %s", keyringResp.JSON401.Message)
	}

	if keyringResp.JSON404 != nil {
		return nil, fmt.Errorf("vault not found: %s", keyringResp.JSON404.Message)
	}

	if keyringResp.JSON500 != nil {
		return nil, fmt.Errorf("server error: %s", keyringResp.JSON500.Message)
	}

	if keyringResp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response from server: %s", keyringResp.Status())
	}

	keyring := keyringResp.JSON200
	if keyring.EncryptionParallelism > 255 {
		return nil, fmt.Errorf("unsupported keyring parallelism: %d", keyring.EncryptionParallelism)
	}

	masterKey, err := crypto.DeriveMasterKey(crypto.DeriveMasterKeyParameters{
		Password:           []byte(vaultPassword),
		EncryptionSalt:     keyring.EncryptionSalt,
		EncryptionNonce:    keyring.EncryptionNonce,
		EncryptedMasterKey: keyring.EncryptedMasterKey,
		EncryptionParameters: crypto.EncryptionKDFParameters{
			TimeCost:    keyring.EncryptionTimeCost,
			MemoryCost:  keyring.EncryptionMemoryCost,
			Parallelism: uint8(keyring.EncryptionParallelism),
			KeySize:     keyring.EncryptionKeySize,
		},
	})
	if err != nil {
		return nil, err
	}

	return masterKey, nil
}

func encryptPlaintextVaultItem(
	masterKey []byte,
	item PlaintextVaultItem,
) (*crypto.ItemEncryptionResult, error) {
	plaintext, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plaintext item: %w", err)
	}

	encryptedItem, err := crypto.EncryptItem(crypto.EncryptItemParameters{
		ItemValue: plaintext,
		MasterKey: masterKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt vault item: %w", err)
	}

	return encryptedItem, nil
}
