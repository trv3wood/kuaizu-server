package models

import "time"

// Feedback represents a user feedback in the database
type Feedback struct {
	ID           int
	UserID       int
	Content      string
	ContactImage *string
	Status       int // 0=pending, 1=handled
	AdminReply   *string
	CreatedAt    time.Time
	UpdatedAt    time.Time

	// Joined fields
	UserNickname *string
}
