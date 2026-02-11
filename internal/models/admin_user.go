package models

import "time"

// AdminUser represents an admin user in the database
type AdminUser struct {
	ID           int
	Username     string
	PasswordHash string
	Nickname     *string
	Status       int // 1=enabled, 0=disabled
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
