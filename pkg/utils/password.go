package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

func VerifyPassword(password, encodedHash string) error {
	parts := strings.Split(encodedHash, ".")
	if len(parts) != 2 {
		return HandleError(errors.New("Invalid encoded hash format"), "Internal error")
	}
	saltBase64 := parts[0]
	hashedPasswordBase64 := parts[1]

	salt, err := base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {
		return HandleError(err, "Internal error")
	}

	hashedPassword, err := base64.StdEncoding.DecodeString(hashedPasswordBase64)
	if err != nil {
		return HandleError(err, "Internal error")
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	if len(hash) != len(hashedPassword) {
		return HandleError(errors.New("hashed password length mismatch"), "Error validating password")
	}

	if subtle.ConstantTimeCompare(hash, hashedPassword) != 1 {
		return HandleError(errors.New("incorrect username or password"), "Error validating password")
	}
	return nil
}

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", HandleError(errors.New("password is required"), "Please enter a password")
	}

	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", HandleError(err, "Error generating salt")
	}

	hashedPassword := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	saltBase64 := base64.StdEncoding.EncodeToString(salt)
	hashedPasswordBase64 := base64.StdEncoding.EncodeToString(hashedPassword)

	return fmt.Sprintf("%s.%s", saltBase64, hashedPasswordBase64), nil
}
