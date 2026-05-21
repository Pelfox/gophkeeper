package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

var (
	// ErrInvalidMasterKeySize is returned when the size of the provided master
	// key for decryption is invalid.
	ErrInvalidMasterKeySize = errors.New("size of the provided master key is invalid")
)

const (
	// keySaltSize specifies the size of the key salt for item, in bytes.
	keySaltSize = 32
	// itemKeySize specifies the size of the item key, in bytes.
	itemKeySize = 32
)

// EncryptItemParameters describes the parameters for item encryption.
type EncryptItemParameters struct {
	// ItemValue is the slice of raw bytes to be encrypted with the MasterKey.
	ItemValue []byte
	// MasterKey is user's vault's master key, used for encryption of the item.
	MasterKey []byte
}

// ItemEncryptionResult holds the result of the item encryption.
type ItemEncryptionResult struct {
	// KeySalt holds the slice of raw bytes for the salt for item key
	// generation.
	KeySalt []byte
	// ItemNonce holds the slice of raw bytes for the AEAD nonce.
	ItemNonce []byte
	// Ciphertext holds the slice of raw bytes for the encrypted vault item
	// value.
	Ciphertext []byte
}

// EncryptItem encrypts the given item for the vault, using the provided master
// key. Returns the slice of raw bytes that can be stored in the database.
func EncryptItem(params EncryptItemParameters) (*ItemEncryptionResult, error) {
	keySalt := make([]byte, keySaltSize)
	if _, err := rand.Read(keySalt); err != nil {
		return nil, fmt.Errorf("failed to generate key salt: %w", err)
	}

	itemKeyHkdf := hkdf.New(sha256.New, params.MasterKey, keySalt, nil)
	itemKey := make([]byte, itemKeySize)
	if _, err := itemKeyHkdf.Read(itemKey); err != nil {
		return nil, fmt.Errorf("failed to generate item key: %w", err)
	}

	aead, err := chacha20poly1305.NewX(itemKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive AEAD from item key: %w", err)
	}

	itemNonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(itemNonce); err != nil {
		return nil, fmt.Errorf("failed to generate item nonce: %w", err)
	}

	ciphertext := aead.Seal(nil, itemNonce, params.ItemValue, nil)
	result := ItemEncryptionResult{
		KeySalt:    keySalt,
		ItemNonce:  itemNonce,
		Ciphertext: ciphertext,
	}

	return &result, nil
}

// DecryptItemParameters describes the parameters, used for decrypting vault
// item.
type DecryptItemParameters struct {
	// Ciphertext is the raw slice of bytes that represents an encrypted vault
	// item.
	Ciphertext []byte
	// ItemNonce holds the raw slice of bytes that is used for AEAD nonce.
	ItemNonce []byte
	// KeySalt holds the raw slice of bytes that is used as a item key salt.
	KeySalt []byte
	// MasterKey is user's vault's master key, used to unlock (decrypt) items.
	MasterKey []byte
}

// DecryptItem decrypts the given ciphertext of an vault item.
func DecryptItem(params DecryptItemParameters) ([]byte, error) {
	if len(params.MasterKey) != chacha20poly1305.KeySize {
		return nil, ErrInvalidMasterKeySize
	}

	itemKeyHkdf := hkdf.New(sha256.New, params.MasterKey, params.KeySalt, nil)
	itemKey := make([]byte, itemKeySize)
	if _, err := itemKeyHkdf.Read(itemKey); err != nil {
		return nil, fmt.Errorf("failed to derive item key: %w", err)
	}

	aead, err := chacha20poly1305.NewX(itemKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive AEAD from item key: %w", err)
	}

	if len(params.ItemNonce) != aead.NonceSize() {
		return nil, ErrInvalidNonceSize
	}

	plaintext, err := aead.Open(nil, params.ItemNonce, params.Ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open the ciphertext: %w", err)
	}

	return plaintext, nil
}
