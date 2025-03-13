package models

import "time"

// User represents a user in the system
type User struct {
	ID                   int         `json:"id"`
	Username             string      `json:"username"`
	Password             string      `json:"-"`
	Email                string      `json:"email"`
	Role                 string      `json:"role"`
	Approved             bool        `json:"approved"`
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`
	LastAction           *UserAction `json:"last_action,omitempty"`
	HasCancellableAction bool        `json:"has_cancellable_action"`
}

// UserAction represents an action taken by a user
type UserAction struct {
	ID           int     `json:"id"`
	UserID       int     `json:"user_id"`
	Username     string  `json:"username,omitempty"`
	Type         string  `json:"type"`
	Amount       float64 `json:"amount"`
	Metadata     string  `json:"metadata"`
	Timestamp    int64   `json:"timestamp"`
	Cancelled    bool    `json:"cancelled"`
	CancelledBy  int     `json:"cancelled_by,omitempty"`
	CancelTime   string  `json:"cancel_time,omitempty"`
	IsLastAction bool    `json:"is_last_action,omitempty"`
}
