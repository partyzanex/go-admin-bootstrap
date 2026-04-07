package usecase

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// --- hashPassword tests ---

func TestHashPassword_ProducesArgon2idFormat(t *testing.T) {
	hash, err := hashPassword("TestPass123")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(hash, "$argon2id$"))

	// Verify the format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	parts := strings.Split(hash, "$")
	assert.Equal(t, 6, len(parts), "expected 6 parts in argon2id hash, got: %s", hash)
	assert.Equal(t, "", parts[0])
	assert.Equal(t, "argon2id", parts[1])
	assert.True(t, strings.HasPrefix(parts[2], "v="))
	assert.True(t, strings.Contains(parts[3], "m="))
	assert.True(t, strings.Contains(parts[3], "t="))
	assert.True(t, strings.Contains(parts[3], "p="))
	assert.NotEmpty(t, parts[4]) // salt
	assert.NotEmpty(t, parts[5]) // hash
}

func TestHashPassword_DifferentSaltsEachTime(t *testing.T) {
	hash1, err := hashPassword("TestPass123")
	require.NoError(t, err)
	hash2, err := hashPassword("TestPass123")
	require.NoError(t, err)
	assert.NotEqual(t, hash1, hash2, "same password should produce different hashes due to random salt")
}

// --- comparePassword tests ---

func TestComparePassword_Argon2id_Correct(t *testing.T) {
	hash, err := hashPassword("MySecure1")
	require.NoError(t, err)
	err = comparePassword(hash, "MySecure1")
	assert.NoError(t, err)
}

func TestComparePassword_Argon2id_Wrong(t *testing.T) {
	hash, err := hashPassword("MySecure1")
	require.NoError(t, err)
	err = comparePassword(hash, "WrongPass1")
	assert.Error(t, err)
}

func TestComparePassword_Bcrypt_Correct(t *testing.T) {
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte("BcryptPass1"), bcrypt.DefaultCost)
	require.NoError(t, err)
	err = comparePassword(string(bcryptHash), "BcryptPass1")
	assert.NoError(t, err)
}

func TestComparePassword_Bcrypt_Wrong(t *testing.T) {
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte("BcryptPass1"), bcrypt.DefaultCost)
	require.NoError(t, err)
	err = comparePassword(string(bcryptHash), "WrongPass1")
	assert.Error(t, err)
}

func TestComparePassword_InvalidArgon2Hash(t *testing.T) {
	err := comparePassword("$argon2id$bad-format", "password")
	assert.Error(t, err)
}

// --- validatePasswordComplexity tests ---

func TestValidatePasswordComplexity_Valid(t *testing.T) {
	err := validatePasswordComplexity("StrongP1!")
	assert.NoError(t, err)
}

func TestValidatePasswordComplexity_TooShort(t *testing.T) {
	err := validatePasswordComplexity("Short1A")
	assert.ErrorIs(t, err, errPasswordTooShort)
}

func TestValidatePasswordComplexity_NoUppercase(t *testing.T) {
	err := validatePasswordComplexity("lowercase1abc")
	assert.ErrorIs(t, err, errPasswordNoUpper)
}

func TestValidatePasswordComplexity_NoLowercase(t *testing.T) {
	err := validatePasswordComplexity("UPPERCASE1ABC")
	assert.ErrorIs(t, err, errPasswordNoLower)
}

func TestValidatePasswordComplexity_NoDigit(t *testing.T) {
	err := validatePasswordComplexity("NoDigitsHere")
	assert.ErrorIs(t, err, errPasswordNoDigit)
}

func TestValidatePasswordComplexity_NoSpecialChar(t *testing.T) {
	err := validatePasswordComplexity("StrongP1")
	assert.ErrorIs(t, err, errPasswordNoSpecialChar)
}

func TestValidatePasswordComplexity_ExactlyMinLength(t *testing.T) {
	// Exactly 8 chars with all requirements: upper, lower, digit, special
	err := validatePasswordComplexity("Abcde1!A")
	assert.NoError(t, err)
}

func TestValidatePasswordComplexity_Empty(t *testing.T) {
	err := validatePasswordComplexity("")
	assert.ErrorIs(t, err, errPasswordTooShort)
}

func TestComparePassword_DefaultBranch_UnknownFormat(t *testing.T) {
	// String starts with neither $argon2id$ nor $2 — falls into the default branch (bcrypt)
	err := comparePassword("notavalidhash", "password")
	assert.Error(t, err)
}

func TestCompareArgon2_VersionParseError(t *testing.T) {
	// v= contains a non-numeric value
	err := compareArgon2("$argon2id$v=abc$m=65536,t=3,p=4$c2FsdA$aGFzaA", "password")
	assert.ErrorContains(t, err, "parsing argon2 version")
}

func TestCompareArgon2_VersionMismatch(t *testing.T) {
	// argon2.Version == 19, using 18
	err := compareArgon2("$argon2id$v=18$m=65536,t=3,p=4$c2FsdA$aGFzaA", "password")
	assert.ErrorIs(t, err, errArgon2VersionChange)
}

func TestCompareArgon2_ParamsParseError(t *testing.T) {
	err := compareArgon2("$argon2id$v=19$badparams$c2FsdA$aGFzaA", "password")
	assert.ErrorContains(t, err, "parsing argon2 params")
}

func TestCompareArgon2_InvalidSaltBase64(t *testing.T) {
	// !!! — invalid base64
	err := compareArgon2("$argon2id$v=19$m=65536,t=3,p=4$!!!$aGFzaA", "password")
	assert.ErrorContains(t, err, "decoding salt")
}

func TestCompareArgon2_InvalidHashBase64(t *testing.T) {
	// c2FsdA — valid base64 salt ("salt"), !!! — invalid hash
	err := compareArgon2("$argon2id$v=19$m=65536,t=3,p=4$c2FsdA$!!!", "password")
	assert.ErrorContains(t, err, "decoding hash")
}
