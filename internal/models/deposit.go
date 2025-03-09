package models

import "time"

// Deposit represents a bank deposit
type Deposit struct {
	DepositID      int64     `json:"deposit_id"`
	ClientID       int64     `json:"client_id"`
	BankName       string    `json:"bank_name"`
	Amount         float64   `json:"amount"`
	Interest       float64   `json:"interest"`
	IsBlocked      bool      `json:"is_blocked"`
	IsFrozen       bool      `json:"is_frozen"`
	FreezeDuration int       `json:"freeze_duration"`
	FreezeUntil    time.Time `json:"freeze_until,omitempty"`
}

// Transfer represents a transfer between accounts
type Transfer struct {
	ClientID    int64   `json:"client_id"`
	BankName    string  `json:"bank_name"`
	FromAccount int64   `json:"from_account"`
	ToAccount   int64   `json:"to_account"`
	Amount      float64 `json:"amount"`
	DepositID   int64   `json:"deposit_id"`
}
