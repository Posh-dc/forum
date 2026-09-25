package backEnd

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"

	"golang.org/x/crypto/argon2"
)

const (
	saltLength = 16
	hashLenght = 32

	argonTime    = 3
	argonMemory  = 64 * 1024
	argonThreads = 4
)

func GenerateSessionId() (string, error) {

	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// RANDOM SALT GENERATOR
func GenerateSalt() ([]byte, error) {

	salt := make([]byte, saltLength)

	_, err := rand.Read(salt)

	if err != nil {
		return nil, err
	}

	return salt, nil
}

//

// GENERATING HASHED PASSWORD
func HashPassword(password string) (string, error) {

	salt, generatingSaltErr := GenerateSalt()

	if generatingSaltErr != nil {

		return "", generatingSaltErr

	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		argonTime,
		argonMemory,
		argonThreads,
		hashLenght,
	)

	encodeSalt := base64.RawURLEncoding.EncodeToString(salt)
	encodeHash := base64.RawURLEncoding.EncodeToString(hash)

	return fmt.Sprintf("%s.%s", encodeSalt, encodeHash), nil
}

// GENERATING VERIFICATION CODE TO SEND TO USER EMAIL
func GenerateVerificationCode() (string, error) {

	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}

// GENERAL HASHING FUNCTION
func GeneralHashFunction(data string) string {

	hash := sha256.Sum256([]byte(data))

	return hex.EncodeToString(hash[:])
}


// FUNCTION THAT VERIFIES GENERAL HASH
func VerifyGeneralHash(data string, expectedHashed string) bool {

	actualHash := GeneralHashFunction(data)

	return subtle.ConstantTimeCompare(
		[]byte(actualHash),
		[]byte(expectedHashed),
	) == 1
}
