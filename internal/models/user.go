package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password,omitempty"` // omitempty ensures password isn't included in responses
	Email     string    `json:"email,omitempty"`
	Role      string    `json:"role,omitempty"` // New field for user role (admin or client)
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
