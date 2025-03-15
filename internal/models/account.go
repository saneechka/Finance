package models

import "time"

// Account represents a user's main financial account
type Account struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AccountTransaction represents a transaction request
type AccountTransaction struct {
	UserID      int64   `json:"user_id"`
	Amount      float64 `json:"amount" binding:"required"`
	Description string  `json:"description"`
}

// TransferRequest represents a transfer between accounts
type TransferRequest struct {
	FromUserID  int64   `json:"from_user_id"`
	ToUserID    int64   `json:"to_user_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	Description string  `json:"description"`
}
