package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

var (
	// ErrInvalidNonceSize is returned when size of the provided encryption
	// nonce is invalid.
	ErrInvalidNonceSize = errors.New("size of the provided nonce is invalid")
)

const (
	// masterKeySize describes the size of the slice used for generating user's
	// vault's master key, in bytes.
	masterKeySize = 32
	// encryptionSaltSize describes the size of the slice used for generating
	// encryption key, in bytes.
	encryptionSaltSize = 16
)

// EncryptionKDFParameters describe the parameters of the algorithm that is
// used to generate an encryption key.
type EncryptionKDFParameters struct {
	// TimeCost describes amount of passes of the given MemoryCost.
	TimeCost uint32
	// MemoryCost describes how much memory should be used.
	MemoryCost uint32
	// Parallelism describes how much threads should be used.
	Parallelism uint8
	// KeySize describes the size of the returned byte slice.
	KeySize uint32
}

// encryptionKeyParameters describes the parameters that is used by encryption
// key generator.
var encryptionKeyParameters = EncryptionKDFParameters{
	TimeCost:    1,         // 1 pass
	MemoryCost:  64 * 1024, // 64 MiB
	Parallelism: 4,         // 4 threads
	KeySize:     32,        // 32 bytes
}

// MasterKeyEncryptionResult describes the result of the key generation and
// encryption process.
type MasterKeyEncryptionResult struct {
	// MasterKey holds an actual plaintext value, representing user's vault's
	// master encryption key. This value should never be stored without a
	// proper encryption.
	MasterKey []byte
	// EncryptionSalt holds the raw bytes that were used for the encryption of
	// master key.
	EncryptionSalt []byte
	// EncryptionNonce holds the raw bytes that were used by the encryption
	// algorithm to encrypt master key.
	EncryptionNonce []byte
	// EncryptedMasterKey holds an actual encrypted master key. This value is
	// safe to be stored in the database.
	EncryptedMasterKey []byte
	// EncryptionParameters describe the parameters of the underlying algorithm
	// that was used to generate an encryption key.
	EncryptionParameters EncryptionKDFParameters
}

// CreateMasterKey creates a random master key for the user's vault deriving
// the encryption of it from their password. This master key is used for
// encrypting and decrypting all vault's elements.
func CreateMasterKey(password []byte) (*MasterKeyEncryptionResult, error) {
	masterKey := make([]byte, masterKeySize)
	if _, err := rand.Read(masterKey); err != nil {
		return nil, fmt.Errorf("failed to create master key: %w", err)
	}

	// Creating a random salt for the encryption of the master key. This value
	// should be stored in the database.
	encryptionSalt := make([]byte, encryptionSaltSize)
	if _, err := rand.Read(encryptionSalt); err != nil {
		return nil, fmt.Errorf("failed to create encryption salt: %w", err)
	}

	// Creating an encryption key that later will be used to encrypt master key.
	// We're deriving it from user's password and a random encryption salt. This
	// value shouldn't be stored.
	encryptionKey := argon2.IDKey(
		password,
		encryptionSalt,
		encryptionKeyParameters.TimeCost,
		encryptionKeyParameters.MemoryCost,
		encryptionKeyParameters.Parallelism,
		encryptionKeyParameters.KeySize,
	)

	aead, err := chacha20poly1305.NewX(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive AEAD from encryption key: %w", err)
	}

	// Creating a random nonce for the AEAD seal. This value should be stored
	// in the database.
	encryptionNonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(encryptionNonce); err != nil {
		return nil, fmt.Errorf("failed to create nonce for an encryption key: %w", err)
	}

	encryptedMasterKey := aead.Seal(nil, encryptionNonce, masterKey, nil)
	result := MasterKeyEncryptionResult{
		MasterKey:            masterKey,
		EncryptionSalt:       encryptionSalt,
		EncryptionNonce:      encryptionNonce,
		EncryptedMasterKey:   encryptedMasterKey,
		EncryptionParameters: encryptionKeyParameters,
	}

	return &result, nil
}

// DeriveMasterKeyParameters describe all parameters that are needed to derive
// a master key for the user.
type DeriveMasterKeyParameters struct {
	// Password is user's password.
	Password []byte
	// EncryptionSalt holds a raw slice of bytes that were used to generate to
	// generate encryption key using KDF.
	EncryptionSalt []byte
	// EncryptionNonce holds a raw slice of bytes that were used as a nonce to
	// encrypt the master key.
	EncryptionNonce []byte
	// EncryptedMasterKey holds a raw slice of bytes that represents an
	// encrypted master key.
	EncryptedMasterKey []byte
	// EncryptionParameters describe the parameters for the KDF that were used
	// to generate encryption key.
	EncryptionParameters EncryptionKDFParameters
}

// DeriveMasterKey derives user's master key for the vault from the given
// parameters. The returned slice of bytes is the master key.
func DeriveMasterKey(params DeriveMasterKeyParameters) ([]byte, error) {
	if len(params.EncryptionNonce) != chacha20poly1305.NonceSizeX {
		return nil, ErrInvalidNonceSize
	}

	// Deriving an encryption key for the master key.
	encryptionKey := argon2.IDKey(
		params.Password,
		params.EncryptionSalt,
		params.EncryptionParameters.TimeCost,
		params.EncryptionParameters.MemoryCost,
		params.EncryptionParameters.Parallelism,
		params.EncryptionParameters.KeySize,
	)

	aead, err := chacha20poly1305.NewX(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive AEAD from encryption key: %w", err)
	}

	masterKey, err := aead.Open(nil, params.EncryptionNonce, params.EncryptedMasterKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt the master key: %w", err)
	}

	return masterKey, nil
}
