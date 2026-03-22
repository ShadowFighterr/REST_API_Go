package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func SignToken(userId int, username, position string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtExpiresIn := os.Getenv("JWT_EXPIRES_IN")
	claims := jwt.MapClaims{
		"sub":  userId,
		"name": username,
		"role": position,
	}

	if jwtExpiresIn != "" {
		duration, err := time.ParseDuration(jwtExpiresIn)
		if err != nil {
			return "", HandleError(err, "Internal error")
		}
		claims["exp"] = jwt.NewNumericDate(time.Now().Add(duration))
	} else {
		claims["exp"] = jwt.NewNumericDate(time.Now().Add(24 * time.Hour))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", HandleError(err, "Internal error")
	}

	return signedToken, nil
}
