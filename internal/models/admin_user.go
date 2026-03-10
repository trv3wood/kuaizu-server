package models

import "time"

// AdminUser represents an admin user in the database
type AdminUser struct {
	ID           int       `db:"id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	Nickname     *string   `db:"nickname"`
	Status       int       `db:"status"` // 1=enabled, 0=disabled
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
