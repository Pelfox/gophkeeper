package crypto

import (
	"bytes"
	"errors"
	"testing"
)

// TestCreateAndDeriveMasterKey verifies that a newly created encrypted master
// key can be derived with the original vault password.
func TestCreateAndDeriveMasterKey(t *testing.T) {
	vaultPassword := []byte("4539b23bfb6162d1f87d4ac3e9d32d8c")

	result, err := CreateMasterKey(vaultPassword)
	if err != nil {
		t.Fatalf("CreateMasterKey returned error: %v", err)
	}
	if len(result.MasterKey) != masterKeySize {
		t.Fatalf(
			"unexpected master key size: got %d want %d",
			len(result.MasterKey),
			masterKeySize,
		)
	}
	if len(result.EncryptionSalt) != encryptionSaltSize {
		t.Fatalf(
			"unexpected encryption salt size: got %d want %d",
			len(result.EncryptionSalt),
			encryptionSaltSize,
		)
	}
	if len(result.EncryptionNonce) == 0 {
		t.Fatal("expected encryption nonce to be populated")
	}
	if len(result.EncryptedMasterKey) == 0 {
		t.Fatal("expected encrypted master key to be populated")
	}

	derivedMasterKey, err := DeriveMasterKey(DeriveMasterKeyParameters{
		Password:             vaultPassword,
		EncryptionSalt:       result.EncryptionSalt,
		EncryptionNonce:      result.EncryptionNonce,
		EncryptedMasterKey:   result.EncryptedMasterKey,
		EncryptionParameters: result.EncryptionParameters,
	})
	if err != nil {
		t.Fatalf("DeriveMasterKey returned error: %v", err)
	}
	if !bytes.Equal(derivedMasterKey, result.MasterKey) {
		t.Fatal("derived master key does not match created master key")
	}
}

// TestDeriveMasterKeyWrongPassword verifies that encrypted master key
// derivation fails with the wrong vault password.
func TestDeriveMasterKeyWrongPassword(t *testing.T) {
	result, err := CreateMasterKey([]byte("correct-password"))
	if err != nil {
		t.Fatalf("CreateMasterKey returned error: %v", err)
	}

	_, err = DeriveMasterKey(DeriveMasterKeyParameters{
		Password:             []byte("wrong-password"),
		EncryptionSalt:       result.EncryptionSalt,
		EncryptionNonce:      result.EncryptionNonce,
		EncryptedMasterKey:   result.EncryptedMasterKey,
		EncryptionParameters: result.EncryptionParameters,
	})
	if err == nil {
		t.Fatal("expected DeriveMasterKey to fail")
	}
}

// TestDeriveMasterKeyInvalidNonce verifies that encrypted master key
// derivation validates nonce size before decrypting.
func TestDeriveMasterKeyInvalidNonce(t *testing.T) {
	result, err := CreateMasterKey([]byte("password"))
	if err != nil {
		t.Fatalf("CreateMasterKey returned error: %v", err)
	}

	_, err = DeriveMasterKey(DeriveMasterKeyParameters{
		Password:             []byte("password"),
		EncryptionSalt:       result.EncryptionSalt,
		EncryptionNonce:      []byte{1, 2, 3},
		EncryptedMasterKey:   result.EncryptedMasterKey,
		EncryptionParameters: result.EncryptionParameters,
	})
	if !errors.Is(err, ErrInvalidNonceSize) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrInvalidNonceSize)
	}
}

// TestDeriveMasterKeyTamperedCiphertext verifies that authentication fails
// when encrypted master key bytes are modified.
func TestDeriveMasterKeyTamperedCiphertext(t *testing.T) {
	password := []byte("password")
	result, err := CreateMasterKey(password)
	if err != nil {
		t.Fatalf("CreateMasterKey returned error: %v", err)
	}

	tampered := append([]byte(nil), result.EncryptedMasterKey...)
	tampered[0] ^= 0xff

	_, err = DeriveMasterKey(DeriveMasterKeyParameters{
		Password:             password,
		EncryptionSalt:       result.EncryptionSalt,
		EncryptionNonce:      result.EncryptionNonce,
		EncryptedMasterKey:   tampered,
		EncryptionParameters: result.EncryptionParameters,
	})
	if err == nil {
		t.Fatal("expected DeriveMasterKey to fail")
	}
}
