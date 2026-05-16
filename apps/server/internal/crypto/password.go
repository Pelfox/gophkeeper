package crypto

import "github.com/alexedwards/argon2id"

// HashPassword hashes the given string password using Argon2id algorithm.
func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

// VerifyPassword verifies the given password string against hashed string.
func VerifyPassword(password string, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}
