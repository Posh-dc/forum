package backEnd

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

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

func GenerateSalt() ([]byte, error) {

	salt := make([]byte, saltLength)

	_, err := rand.Read(salt)

	if err != nil {
		return nil, err
	}

	return salt, nil
}

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
