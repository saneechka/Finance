package storage

import (
	"database/sql"
	"errors"
	"finance/internal/models"
	"time"
)

// EnsureAccountsTableExists creates the accounts table if it doesn't exist
func EnsureAccountsTableExists() error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL UNIQUE,
			balance REAL DEFAULT 0.0,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`
	_, err := db.Exec(createTableQuery)
	return err
}

// CreateAccount creates a new account for a user
func CreateAccount(userID int64) error {
	if err := EnsureAccountsTableExists(); err != nil {
		return err
	}

	now := time.Now()
	query := `
		INSERT INTO accounts (user_id, balance, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id) DO NOTHING
	`

	_, err := db.Exec(query, userID, 0.0, now, now)
	return err
}

// GetAccount retrieves account details for a user
func GetAccount(userID int64) (*models.Account, error) {
	if err := EnsureAccountsTableExists(); err != nil {
		return nil, err
	}

	query := `
		SELECT id, user_id, balance, created_at, updated_at
		FROM accounts
		WHERE user_id = ?
	`

	var account models.Account
	err := db.QueryRow(query, userID).Scan(
		&account.ID,
		&account.UserID,
		&account.Balance,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// If account doesn't exist, create it automatically
			err = CreateAccount(userID)
			if err != nil {
				return nil, err
			}
			// Try fetching again
			return GetAccount(userID)
		}
		return nil, err
	}

	return &account, nil
}

// UpdateAccountBalance updates the account balance for a user
func UpdateAccountBalance(userID int64, amount float64) error {
	if err := EnsureAccountsTableExists(); err != nil {
		return err
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// First check if the account exists, if not create it
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM accounts WHERE user_id = ?)", userID).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		now := time.Now()
		_, err = tx.Exec(
			"INSERT INTO accounts (user_id, balance, created_at, updated_at) VALUES (?, ?, ?, ?)",
			userID, amount, now, now,
		)
		if err != nil {
			return err
		}
	} else {
		// Update existing account
		_, err = tx.Exec(
			"UPDATE accounts SET balance = balance + ?, updated_at = ? WHERE user_id = ?",
			amount, time.Now(), userID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// CheckSufficientFunds verifies if an account has enough funds
func CheckSufficientFunds(userID int64, amount float64) (bool, error) {
	query := `SELECT balance >= ? FROM accounts WHERE user_id = ?`

	var sufficient bool
	err := db.QueryRow(query, amount, userID).Scan(&sufficient)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, errors.New("account not found")
		}
		return false, err
	}

	return sufficient, nil
}

// TransferFunds transfers funds between accounts
func TransferFunds(fromUserID int64, toUserID int64, amount float64) error {
	if fromUserID == toUserID {
		return errors.New("cannot transfer funds to the same account")
	}

	if amount <= 0 {
		return errors.New("transfer amount must be positive")
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if source account has sufficient funds
	var balance float64
	err = tx.QueryRow("SELECT balance FROM accounts WHERE user_id = ?", fromUserID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("source account not found")
		}
		return err
	}

	if balance < amount {
		return errors.New("insufficient funds")
	}

	now := time.Now()

	// Deduct from source account
	_, err = tx.Exec(
		"UPDATE accounts SET balance = balance - ?, updated_at = ? WHERE user_id = ?",
		amount, now, fromUserID,
	)
	if err != nil {
		return err
	}

	// Check if destination account exists, create if not
	var destExists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM accounts WHERE user_id = ?)", toUserID).Scan(&destExists)
	if err != nil {
		return err
	}

	if !destExists {
		_, err = tx.Exec(
			"INSERT INTO accounts (user_id, balance, created_at, updated_at) VALUES (?, ?, ?, ?)",
			toUserID, amount, now, now,
		)
	} else {
		// Add to destination account
		_, err = tx.Exec(
			"UPDATE accounts SET balance = balance + ?, updated_at = ? WHERE user_id = ?",
			amount, now, toUserID,
		)
	}
	if err != nil {
		return err
	}

	return tx.Commit()
}
