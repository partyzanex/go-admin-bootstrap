package usecase

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

const (
	argon2Time    = 3
	argon2Memory  = 64 * 1024
	argon2Threads = 4
	argon2KeyLen  = 32
	argon2SaltLen = 16

	argon2idPrefix = "$argon2id$"
	bcryptPrefix   = "$2"

	minPasswordLen = 8
)

var (
	errPasswordTooShort    = errors.New("password must be at least 8 characters")
	errPasswordNoUpper     = errors.New("password must contain at least one uppercase letter")
	errPasswordNoLower     = errors.New("password must contain at least one lowercase letter")
	errPasswordNoDigit     = errors.New("password must contain at least one digit")
	errInvalidArgon2Hash   = errors.New("invalid argon2id hash format")
	errArgon2VersionChange = errors.New("argon2id version mismatch")
)

func hashPassword(password string) (string, error) {
	salt := make([]byte, argon2SaltLen)

	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	return fmt.Sprintf(
		"%sv=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2idPrefix,
		argon2.Version,
		argon2Memory, argon2Time, argon2Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func comparePassword(hashedPassword, password string) error {
	switch {
	case strings.HasPrefix(hashedPassword, argon2idPrefix):
		return compareArgon2(hashedPassword, password)
	case strings.HasPrefix(hashedPassword, bcryptPrefix):
		return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	default:
		return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	}
}

func compareArgon2(encoded, password string) error {
	parts := strings.Split(encoded, "$")

	// $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	const expectedParts = 6
	if len(parts) != expectedParts {
		return errInvalidArgon2Hash
	}

	var version int

	var memory, time uint32

	var threads uint8

	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return fmt.Errorf("parsing argon2 version: %w", err)
	}

	if version != argon2.Version {
		return errArgon2VersionChange
	}

	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return fmt.Errorf("parsing argon2 params: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return fmt.Errorf("decoding salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return fmt.Errorf("decoding hash: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, argon2KeyLen)

	if subtle.ConstantTimeCompare(hash, expectedHash) != 1 {
		return bcrypt.ErrMismatchedHashAndPassword
	}

	return nil
}

func validatePasswordComplexity(password string) error {
	if len(password) < minPasswordLen {
		return errPasswordTooShort
	}

	var hasUpper, hasLower, hasDigit bool

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasUpper {
		return errPasswordNoUpper
	}

	if !hasLower {
		return errPasswordNoLower
	}

	if !hasDigit {
		return errPasswordNoDigit
	}

	return nil
}
