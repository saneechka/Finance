package storage

import (
	"database/sql"
	"errors"
	"finance/internal/models"
	"fmt"
	"time"
)


func EnsureDepositsTableExists() error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS deposits (
			deposit_id INTEGER PRIMARY KEY AUTOINCREMENT,
			client_id INTEGER NOT NULL,
			bank_name TEXT NOT NULL,
			amount REAL NOT NULL,
			interest REAL NOT NULL,
			is_blocked INTEGER DEFAULT 0,
			is_frozen INTEGER DEFAULT 0,
			freeze_duration INTEGER DEFAULT 0,
			freeze_until TIMESTAMP,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			UNIQUE(client_id, bank_name, deposit_id)
		)
	`
	_, err := db.Exec(createTableQuery)
	return err
}

// SaveDeposit stores a new deposit in the database
func SaveDeposit(deposit models.Deposit) error {
	if err := EnsureDepositsTableExists(); err != nil {
		return err
	}

	now := time.Now()
	query := `
		INSERT INTO deposits (
			client_id, bank_name, amount, interest,
			is_blocked, is_frozen, freeze_duration,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.Exec(
		query,
		deposit.ClientID,
		deposit.BankName,
		deposit.Amount,
		deposit.Interest,
		boolToInt(deposit.IsBlocked),
		boolToInt(deposit.IsFrozen),
		deposit.FreezeDuration,
		now,
		now,
	)

	if err != nil {
		return err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	deposit.DepositID = lastID
	return nil
}

// GetDeposit retrieves a deposit by its ID, client ID, and bank name
func GetDeposit(clientID int64, bankName string, depositID int64) (models.Deposit, error) {
	var deposit models.Deposit

	query := `
		SELECT deposit_id, client_id, bank_name, amount, interest, 
		       is_blocked, is_frozen, freeze_duration, freeze_until
		FROM deposits
		WHERE client_id = ? AND bank_name = ? AND deposit_id = ?
	`

	var isBlocked, isFrozen int
	var freezeUntil sql.NullTime
	err := db.QueryRow(query, clientID, bankName, depositID).Scan(
		&deposit.DepositID,
		&deposit.ClientID,
		&deposit.BankName,
		&deposit.Amount,
		&deposit.Interest,
		&isBlocked,
		&isFrozen,
		&deposit.FreezeDuration,
		&freezeUntil,
	)




	if err != nil {
		if err == sql.ErrNoRows {
			return deposit, errors.New("deposit not found")
		}
		return deposit, err
	}

	deposit.IsBlocked = isBlocked == 1
	deposit.IsFrozen = isFrozen == 1

	if freezeUntil.Valid {
		deposit.FreezeUntil = freezeUntil.Time
	}

	return deposit, nil
}

// GetDepositsByUserID retrieves all deposits for a specific user
func GetDepositsByUserID(userID int64) ([]models.Deposit, error) {
	if err := EnsureDepositsTableExists(); err != nil {
		return nil, err
	}

	query := `
		SELECT 
			deposit_id, 
			client_id, 
			bank_name, 
			amount, 
			interest, 
			is_blocked, 
			is_frozen, 
			freeze_duration, 
			freeze_until,
			created_at,
			updated_at
		FROM deposits 
		WHERE client_id = ?
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("database query error: %v", err)
	}
	defer rows.Close()

	deposits := []models.Deposit{}
	for rows.Next() {
		var deposit models.Deposit
		var isBlocked, isFrozen int
		var freezeUntil sql.NullTime
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&deposit.DepositID,
			&deposit.ClientID,
			&deposit.BankName,
			&deposit.Amount,
			&deposit.Interest,
			&isBlocked,
			&isFrozen,
			&deposit.FreezeDuration,
			&freezeUntil,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning deposit row: %v", err)
		}

		deposit.IsBlocked = isBlocked == 1
		deposit.IsFrozen = isFrozen == 1
		if freezeUntil.Valid {
			deposit.FreezeUntil = freezeUntil.Time
		}

		deposits = append(deposits, deposit)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deposit rows: %v", err)
	}

	return deposits, nil
}

// DeleteDeposit removes a deposit from the database
func DeleteDeposit(clientID int64, bankName string) error {
	if err := EnsureDepositsTableExists(); err != nil {
		return err
	}

	query := `DELETE FROM deposits WHERE client_id = ? AND bank_name = ?`
	result, err := db.Exec(query, clientID, bankName)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// BlockDeposit marks a deposit as blocked
func BlockDeposit(clientID int64, bankName string, depositID int64) error {
	deposit, err := GetDeposit(clientID, bankName, depositID)
	if err != nil {
		return err
	}

	if deposit.IsBlocked {
		return errors.New("deposit is already blocked")
	}

	query := `
		UPDATE deposits 
		SET is_blocked = 1, updated_at = ?
		WHERE client_id = ? AND bank_name = ? AND deposit_id = ?
	`

	_, err = db.Exec(query, time.Now(), clientID, bankName, depositID)
	return err
}

// UnblockDeposit removes the block from a deposit
func UnblockDeposit(clientID int64, bankName string, depositID int64) error {
	deposit, err := GetDeposit(clientID, bankName, depositID)
	if err != nil {
		return err
	}

	if !deposit.IsBlocked {
		return sql.ErrNoRows
	}

	query := `
		UPDATE deposits 
		SET is_blocked = 0, updated_at = ?
		WHERE client_id = ? AND bank_name = ? AND deposit_id = ?
	`

	_, err = db.Exec(query, time.Now(), clientID, bankName, depositID)
	return err
}

// FreezeDeposit marks a deposit as frozen for the specified duration
func FreezeDeposit(clientID int64, bankName string, depositID int64, freezeDuration int) error {
	deposit, err := GetDeposit(clientID, bankName, depositID)
	if err != nil {
		return err
	}

	if deposit.IsBlocked {
		return errors.New("deposit is already blocked")
	}

	// Calculate the time when the freeze will end
	freezeUntil := time.Now().Add(time.Duration(freezeDuration) * time.Hour)

	query := `
		UPDATE deposits 
		SET is_frozen = 1, freeze_duration = ?, freeze_until = ?, updated_at = ?
		WHERE client_id = ? AND bank_name = ? AND deposit_id = ?
	`

	_, err = db.Exec(
		query,
		freezeDuration,
		freezeUntil,
		time.Now(),
		clientID,
		bankName,
		depositID,
	)

	return err
}

// TransferBetweenAccounts transfers funds between accounts
func TransferBetweenAccounts(transfer models.Transfer) error {
	if err := EnsureDepositsTableExists(); err != nil {
		return err
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if source account exists and has sufficient funds
	var sourceAmount float64
	sourceQuery := `
		SELECT amount FROM deposits 
		WHERE client_id = ? AND bank_name = ? AND deposit_id = ? AND is_blocked = 0
	`
	err = tx.QueryRow(sourceQuery, transfer.ClientID, transfer.BankName, transfer.FromAccount).Scan(&sourceAmount)
	if err != nil {
		return err
	}

	if sourceAmount < transfer.Amount {
		return fmt.Errorf("insufficient funds: available %.2f, needed %.2f", sourceAmount, transfer.Amount)
	}

	// Check if destination account exists
	var exists int
	destQuery := `SELECT COUNT(*) FROM deposits WHERE client_id = ? AND bank_name = ? AND deposit_id = ?`
	err = tx.QueryRow(destQuery, transfer.ClientID, transfer.BankName, transfer.ToAccount).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return sql.ErrNoRows
	}

	now := time.Now()

	// Update source account
	_, err = tx.Exec(`
		UPDATE deposits SET amount = amount - ?, updated_at = ?
		WHERE client_id = ? AND bank_name = ? AND deposit_id = ?
	`, transfer.Amount, now, transfer.ClientID, transfer.BankName, transfer.FromAccount)
	if err != nil {
		return err
	}

	// Update destination account
	_, err = tx.Exec(`
		UPDATE deposits SET amount = amount + ?, updated_at = ?
		WHERE client_id = ? AND bank_name = ? AND deposit_id = ?
	`, transfer.Amount, now, transfer.ClientID, transfer.BankName, transfer.ToAccount)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}

// Helper function to convert bool to int for SQLite

