package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"github.com/Mangatsu/server/pkg/log"
	"go.uber.org/zap"
	"golang.org/x/crypto/argon2"
)

// Argon2idHash is a wrapper for the argon2id hashing algorithm.
// https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#password-hashing-algorithms
type Argon2idHash struct {
	// time represents the number of passed over the specified memory
	time uint32
	// CPU memory cost
	memory uint32
	// threads for parallelism
	threads uint8
	// keyLen of the generate hash key
	keyLen uint32
	// saltLen the length of the salt used
	saltLen uint32
}

// HashSalt is a wrapper for the hash and salt values.
type HashSalt struct {
	Hash []byte
	Salt []byte
}

// NewArgon2idHash constructor function for Argon2idHash.
func NewArgon2idHash(time, saltLen, memory uint32, threads uint8, keyLen uint32) *Argon2idHash {
	return &Argon2idHash{
		// time represents the number of passed over the specified memory
		time: time,
		// salt length
		saltLen: saltLen,
		// CPU memory cost (in KiB)
		memory: memory,
		// threads for parallelism
		threads: threads,
		// hash key length
		keyLen: keyLen,
	}
}

// DefaultArgon2idHash constructor function for Argon2idHash.
// https://tobtu.com/minimum-password-settings/
func DefaultArgon2idHash() *Argon2idHash {
	return &Argon2idHash{
		time:    2,
		saltLen: 16,
		memory:  19456,
		threads: 2,
		keyLen:  32,
	}
}

// randomSecret generates a random secret of a given length. Used for salt generation.
func randomSecret(length uint32) ([]byte, error) {
	secret := make([]byte, length)

	_, err := rand.Read(secret)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

// GenerateHash using the password and provided salt.
// If no salt value was provided, fallbacks to a random value of a given length.
func (a *Argon2idHash) GenerateHash(password, salt []byte) (*HashSalt, error) {
	if len(salt) == 0 {
		var err error
		if salt, err = randomSecret(a.saltLen); err != nil {
			return nil, err
		}
	}

	hash := argon2.IDKey(password, salt, a.time, a.memory, a.threads, a.keyLen)

	return &HashSalt{Hash: hash, Salt: salt}, nil
}

// Compare generated hash with stored hash.
func (a *Argon2idHash) Compare(hash, salt, password []byte) bool {
	hashSalt, err := a.GenerateHash(password, salt)
	if err != nil {
		log.Z.Debug("failed to generate hash", zap.Error(err))
		return false
	}

	// ConstantTimeCompare used to prevent timing attacks.
	return subtle.ConstantTimeCompare(hash, hashSalt.Hash) == 1
}
