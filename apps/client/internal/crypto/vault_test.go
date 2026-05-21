package crypto

import (
	"bytes"
	"errors"
	"testing"
)

// TestEncryptAndDecryptItem verifies that encrypted vault item bytes can be
// decrypted with the same master key.
func TestEncryptAndDecryptItem(t *testing.T) {
	masterKey := bytes.Repeat([]byte{1}, masterKeySize)
	plaintext := []byte("secret value")

	result, err := EncryptItem(EncryptItemParameters{
		ItemValue: plaintext,
		MasterKey: masterKey,
	})
	if err != nil {
		t.Fatalf("EncryptItem returned error: %v", err)
	}
	if len(result.KeySalt) != keySaltSize {
		t.Fatalf(
			"unexpected key salt size: got %d want %d",
			len(result.KeySalt),
			keySaltSize,
		)
	}
	if len(result.ItemNonce) == 0 {
		t.Fatal("expected item nonce to be populated")
	}
	if len(result.Ciphertext) == 0 {
		t.Fatal("expected ciphertext to be populated")
	}

	decrypted, err := DecryptItem(DecryptItemParameters{
		Ciphertext: result.Ciphertext,
		ItemNonce:  result.ItemNonce,
		KeySalt:    result.KeySalt,
		MasterKey:  masterKey,
	})
	if err != nil {
		t.Fatalf("DecryptItem returned error: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("unexpected plaintext: got %q want %q", decrypted, plaintext)
	}
}

// TestDecryptItemInvalidMasterKeySize verifies that decryption validates
// master key size before deriving the item key.
func TestDecryptItemInvalidMasterKeySize(t *testing.T) {
	_, err := DecryptItem(DecryptItemParameters{
		Ciphertext: []byte{1, 2, 3},
		ItemNonce:  bytes.Repeat([]byte{1}, 24),
		KeySalt:    bytes.Repeat([]byte{2}, keySaltSize),
		MasterKey:  []byte{1, 2, 3},
	})
	if !errors.Is(err, ErrInvalidMasterKeySize) {
		t.Fatalf(
			"unexpected error: got %v want %v",
			err,
			ErrInvalidMasterKeySize,
		)
	}
}

// TestDecryptItemInvalidNonceSize verifies that decryption validates item
// nonce size before opening the ciphertext.
func TestDecryptItemInvalidNonceSize(t *testing.T) {
	masterKey := bytes.Repeat([]byte{1}, masterKeySize)

	_, err := DecryptItem(DecryptItemParameters{
		Ciphertext: []byte{1, 2, 3},
		ItemNonce:  []byte{1, 2, 3},
		KeySalt:    bytes.Repeat([]byte{2}, keySaltSize),
		MasterKey:  masterKey,
	})
	if !errors.Is(err, ErrInvalidNonceSize) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrInvalidNonceSize)
	}
}

// TestDecryptItemTamperedCiphertext verifies that authentication fails when
// encrypted item bytes are modified.
func TestDecryptItemTamperedCiphertext(t *testing.T) {
	masterKey := bytes.Repeat([]byte{1}, masterKeySize)
	result, err := EncryptItem(EncryptItemParameters{
		ItemValue: []byte("secret value"),
		MasterKey: masterKey,
	})
	if err != nil {
		t.Fatalf("EncryptItem returned error: %v", err)
	}

	tampered := append([]byte(nil), result.Ciphertext...)
	tampered[0] ^= 0xff

	_, err = DecryptItem(DecryptItemParameters{
		Ciphertext: tampered,
		ItemNonce:  result.ItemNonce,
		KeySalt:    result.KeySalt,
		MasterKey:  masterKey,
	})
	if err == nil {
		t.Fatal("expected DecryptItem to fail")
	}
}
