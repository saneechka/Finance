package models

import "time"

// Transaction represents a financial transaction in the system
type Transaction struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Type            string    `json:"type"`
	Amount          *float64  `json:"amount,omitempty"`
	Description     string    `json:"description"`
	TransactionDate time.Time `json:"transaction_date"`
}
