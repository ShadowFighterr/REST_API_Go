package utils

import "errors"

type ContextKey string

func AuthorizeUser(allowedRoles []string, userRole string) (bool, error) {
	for _, role := range allowedRoles {
		if role == userRole {
			return true, nil
		}
	}
	return false, errors.New("User is not authorized")
}
