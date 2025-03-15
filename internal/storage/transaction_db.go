package storage

import (
	"database/sql"
	"finance/internal/models"
	"time"
)

// EnsureTransactionsTableExists creates the transactions table if it doesn't exist
func EnsureTransactionsTableExists() error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			transaction_type TEXT NOT NULL,
			amount REAL,
			description TEXT,
			created_at TIMESTAMP NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`
	_, err := db.Exec(createTableQuery)
	return err
}

// LogTransaction records a transaction to the database
// func LogTransaction(userID int64, transactionType string, amount *float64, description string) (*models.Transaction, error) {
// 	if err := EnsureTransactionsTableExists(); err != nil {
// 		return nil, err
// 	}

// 	now := time.Now()
// 	transaction := &models.Transaction{
// 		UserID:          userID,
// 		Type:            transactionType,
// 		Amount:          amount,
// 		Description:     description,
// 		TransactionDate: now,
// 	}

// 	query := `
// 		INSERT INTO transactions (user_id, transaction_type, amount, description, created_at)
// 		VALUES (?, ?, ?, ?, ?)
// 	`
// 	var args []interface{}
// 	if amount != nil {
// 		args = []interface{}{userID, transactionType, *amount, description, now}
// 	} else {
// 		args = []interface{}{userID, transactionType, nil, description, now}
// 	}

// 	result, err := db.Exec(query, args...)
// 	if err != nil {
// 		return nil, err
// 	}

// 	id, err := result.LastInsertId()
// 	if err != nil {
// 		return nil, err
// 	}
// 	transaction.ID = id

// 	return transaction, nil
// }

// LogTransactionTx logs a transaction within an existing database transaction
func LogTransactionTx(tx *sql.Tx, userID int64, transactionType string, amount *float64, description string) (*models.Transaction, error) {
	if tx == nil {
		//return LogTransaction(userID, transactionType, amount, description)
	}

	now := time.Now()
	transaction := &models.Transaction{
		UserID:          userID,
		Type:            transactionType,
		Amount:          amount,
		Description:     description,
		TransactionDate: now,
	}

	query := `
		INSERT INTO transactions (user_id, transaction_type, amount, description, created_at)
		VALUES (?, ?, ?, ?, ?)
	`
	var args []interface{}
	if amount != nil {
		args = []interface{}{userID, transactionType, *amount, description, now}
	} else {
		args = []interface{}{userID, transactionType, nil, description, now}
	}

	result, err := tx.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	transaction.ID = id

	return transaction, nil
}

// // GetUserTransactions retrieves all transactions for a specific user
// func GetUserTransactions(userID int64) ([]*models.Transaction, error) {
// 	if err := EnsureTransactionsTableExists(); err != nil {
// 		return nil, err
// 	}

// 	query := `
// 		SELECT id, user_id, transaction_type, amount, description, created_at
// 		FROM transactions
// 		WHERE user_id = ?
// 		ORDER BY created_at DESC
// 	`

// 	rows, err := db.Query(query, userID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	transactions := []*models.Transaction{}

// 	for rows.Next() {
// 		transaction := &models.Transaction{}
// 		var nullAmount sql.NullFloat64

// 		err := rows.Scan(
// 			&transaction.ID,
// 			&transaction.UserID,
// 			&transaction.Type,
// 			&nullAmount,
// 			&transaction.Description,
// 			&transaction.TransactionDate,
// 		)

// 		if err != nil {
// 			return nil, err
// 		}

// 		if nullAmount.Valid {
// 			amount := nullAmount.Float64
// 			transaction.Amount = &amount
// 		}

// 		transactions = append(transactions, transaction)
// 	}

// 	if err = rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	return transactions, nil
// }

// Helper function to convert bool to int for SQLite

