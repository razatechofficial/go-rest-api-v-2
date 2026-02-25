package domain

import "time"

type UserID string

type User struct {
	ID        UserID
	Name      string
	Email     string
	Password  string // bcrypt hash — never plain text
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time // pointer — nil means not deleted
}
