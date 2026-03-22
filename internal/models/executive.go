package models

import "database/sql"

type Executive struct {
	ID                    int            `json:"id,omitempty" db:"id"`
	FirstName             string         `json:"first_name,omitempty" db:"first_name"`
	LastName              string         `json:"last_name,omitempty" db:"last_name"`
	Position              string         `json:"position,omitempty" db:"position"`
	Email                 string         `json:"email,omitempty" db:"email"`
	Username              string         `json:"username,omitempty" db:"username"`
	Password              string         `json:"password,omitempty" db:"password"`
	PasswordChangedAt     sql.NullString `json:"password_changed_at,omitempty" db:"password_changed_at"`
	UserCreatedAt         sql.NullString `json:"user_created_at,omitempty" db:"user_created_at"`
	PasswordResetCode     sql.NullString `json:"password_reset_code,omitempty" db:"password_reset_token"`
	PasswordCodeExpiresAt sql.NullString `json:"password_code_expires_at,omitempty" db:"password_token_expires"`
	InactiveStatus        bool           `json:"inactive_status,omitempty" db:"inactive_status"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type UpdatePasswordResponse struct {
	Token           string `json:"token"`
	PasswordUpdated bool   `json:"password_updated"`
}
