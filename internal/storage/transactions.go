package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"finance/internal/utils"
)
 
// TransactionStatistics represents transaction statistics data
type TransactionStatistics struct {
	TotalTransactions int     `json:"total_transactions"`
	TotalAmount       float64 `json:"total_amount"`
	ActiveUsers       int     `json:"active_users"`
	AvgTransaction    float64 `json:"avg_transaction"`
}

// Transaction represents transaction history entry
type Transaction struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Type      string    `json:"type"`
	Amount    *float64  `json:"amount,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	CanCancel bool      `json:"can_cancel"`
}

// ActionLog represents a complete action log record
type ActionLog struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	Username    string     `json:"username"`
	Type        string     `json:"type"`
	Amount      *float64   `json:"amount,omitempty"`
	Metadata    string     `json:"metadata"`
	Timestamp   time.Time  `json:"timestamp"`
	CancelledBy *int       `json:"cancelled_by,omitempty"`
	CancelTime  *time.Time `json:"cancel_time,omitempty"`
}

// EnsureTransactionTablesExist creates necessary tables for tracking transactions
func EnsureTransactionTablesExist() error {
	// Create transaction history table
	transactionHistoryQuery := `
        CREATE TABLE IF NOT EXISTS transaction_history (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER NOT NULL,
            transaction_type TEXT NOT NULL,
            amount REAL,
            metadata TEXT,
            timestamp TIMESTAMP NOT NULL
        )
    `
	if _, err := db.Exec(transactionHistoryQuery); err != nil {
		return err
	}

	// Create cancellation tracking table
	cancellationTrackingQuery := `
        CREATE TABLE IF NOT EXISTS cancellation_tracking (
            operator_id INTEGER NOT NULL,
            user_id INTEGER NOT NULL,
            deposit_id INTEGER NOT NULL,
            transaction_id INTEGER NOT NULL,
            cancelled_at TIMESTAMP NOT NULL,
            PRIMARY KEY (operator_id, user_id, deposit_id)
        )
    `
	_, err := db.Exec(cancellationTrackingQuery)
	return err
}

// LogTransaction adds a transaction to the history with encryption
func LogTransaction(userID int64, txType string, amount *float64, metadata string) (int64, error) {
	if err := EnsureTransactionTablesExist(); err != nil {
		return 0, err
	}

	// Create log data structure
	logData := map[string]interface{}{
		"user_id":   userID,
		"type":      txType,
		"amount":    amount,
		"metadata":  metadata,
		"timestamp": time.Now(),
	}

	// Encrypt the metadata
	encryptedMetadata, err := utils.EncryptLogMessage(logData)
	if err != nil {
		// Fall back to plain text if encryption fails
		log.Printf("Failed to encrypt log: %v", err)
		encryptedMetadata = metadata
	}

	query := `
        INSERT INTO transaction_history (user_id, transaction_type, amount, metadata, timestamp)
        VALUES (?, ?, ?, ?, ?)
    `
	result, err := db.Exec(query, userID, txType, amount, encryptedMetadata, time.Now())
	if err != nil {
		return 0, err
	}

	// Also write to system log file
	logToFile(userID, txType, amount, metadata)

	return result.LastInsertId()
}

// logToFile writes logs to an encrypted file
func logToFile(userID int64, txType string, amount *float64, metadata string) {
	logData := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"user_id":   userID,
		"type":      txType,
		"amount":    amount,
		"metadata":  metadata,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(logData)
	if err != nil {
		log.Printf("Failed to marshal log data: %v", err)
		return
	}

	// Encrypt the log entry
	encryptedLog, err := utils.EncryptLogMessage(logData)
	if err != nil {
		log.Printf("Failed to encrypt log: %v", err)
		encryptedLog = string(jsonData)
	}

	// Open log file in append mode
	logPath := filepath.Join(os.TempDir(), "finance_system.log")
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}
	defer file.Close()

	// Write encrypted log
	file.WriteString(encryptedLog + "\n")
}

// GetTransactionStatistics returns statistics about transactions
func GetTransactionStatistics() (*TransactionStatistics, error) {
	stats := &TransactionStatistics{}

	// Ensure tables exist
	if err := EnsureTransactionTablesExist(); err != nil {
		return stats, err
	}

	// Get total transactions
	if err := db.QueryRow("SELECT COUNT(*) FROM transaction_history").Scan(&stats.TotalTransactions); err != nil {
		return stats, err
	}

	// Get total amount (for transfers)
	if err := db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM transaction_history WHERE transaction_type = 'transfer' AND amount IS NOT NULL").Scan(&stats.TotalAmount); err != nil {
		return stats, err
	}

	// Get active users (users with transactions in the last 30 days)
	if err := db.QueryRow("SELECT COUNT(DISTINCT user_id) FROM transaction_history WHERE timestamp > datetime('now', '-30 days')").Scan(&stats.ActiveUsers); err != nil {
		return stats, err
	}

	// Get average transaction amount
	if err := db.QueryRow("SELECT COALESCE(AVG(amount), 0) FROM transaction_history WHERE transaction_type = 'transfer' AND amount IS NOT NULL").Scan(&stats.AvgTransaction); err != nil {
		return stats, err
	}

	return stats, nil
}

// GetTransactionHistory returns transaction history with optional filters
func GetTransactionHistory(username string, txType string, date *time.Time) ([]Transaction, error) {
	transactions := []Transaction{}

	// Ensure tables exist
	if err := EnsureTransactionTablesExist(); err != nil {
		return transactions, err
	}

	// Build query with filters
	query := `
        SELECT th.id, th.user_id, u.username, th.transaction_type, th.amount, th.timestamp,
               (NOT EXISTS (SELECT 1 FROM cancellation_tracking ct WHERE ct.transaction_id = th.id)) AS can_cancel
        FROM transaction_history th
        LEFT JOIN users u ON th.user_id = u.id
        WHERE 1=1
    `
	args := []interface{}{}

	if username != "" {
		query += " AND u.username LIKE ?"
		args = append(args, "%"+username+"%")
	}

	if txType != "" {
		query += " AND th.transaction_type = ?"
		args = append(args, txType)
	}

	if date != nil {
		query += " AND date(th.timestamp) = date(?)"
		args = append(args, date.Format("2006-01-02"))
	}

	query += " ORDER BY th.timestamp DESC LIMIT 100"

	rows, err := db.Query(query, args...)
	if err != nil {
		return transactions, err
	}
	defer rows.Close()

	// Process rows
	for rows.Next() {
		var tx Transaction
		var canCancel bool

		err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&tx.Username,
			&tx.Type,
			&tx.Amount,
			&tx.Timestamp,
			&canCancel,
		)
		if err != nil {
			return nil, err
		}

		// Exclude delete operations from cancellations
		tx.CanCancel = canCancel && tx.Type != "delete"

		transactions = append(transactions, tx)
	}

	if err = rows.Err(); err != nil {
		return transactions, err
	}

	return transactions, nil
}

// CancelTransaction allows an operator to cancel a transaction
func CancelTransaction(transactionID int64, operatorID int) error {
	if err := EnsureTransactionTablesExist(); err != nil {
		return err
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get transaction details
	var txDetails Transaction
	var metadata string

	err = tx.QueryRow(`
        SELECT id, user_id, transaction_type, amount, metadata
        FROM transaction_history
        WHERE id = ?
    `, transactionID).Scan(&txDetails.ID, &txDetails.UserID, &txDetails.Type, &txDetails.Amount, &metadata)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("transaction not found")
		}
		return err
	}

	// Check if transaction type is valid for cancellation (can't cancel delete operations)
	if txDetails.Type == "delete" {
		return errors.New("delete operations cannot be cancelled")
	}

	// Check if this operator has already used their cancellation for this user/deposit
	var depositID int64

	// Extract deposit ID from metadata
	if err := tx.QueryRow(`SELECT deposit_id FROM deposits WHERE client_id = ?`, txDetails.UserID).Scan(&depositID); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("deposit not found for this user")
		}
		return err
	}

	// Check if cancellation has already been used
	var count int
	err = tx.QueryRow(`
        SELECT COUNT(*) FROM cancellation_tracking
        WHERE operator_id = ? AND user_id = ? AND deposit_id = ?
    `, operatorID, txDetails.UserID, depositID).Scan(&count)

	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("you have already used your cancellation for this account")
	}

	// Perform cancellation based on transaction type
	switch txDetails.Type {
	case "transfer":
		// Reverse the transfer
		// This would require more details about which accounts were involved
		// For simplicity, we'll just log the cancellation

	case "freeze":
		// Unfreeze the deposit
		_, err = tx.Exec(`
            UPDATE deposits
            SET is_frozen = 0, freeze_duration = 0, freeze_until = NULL
            WHERE client_id = ? AND deposit_id = ?
        `, txDetails.UserID, depositID)

	case "block":
		// Unblock the deposit
		_, err = tx.Exec(`
            UPDATE deposits
            SET is_blocked = 0
            WHERE client_id = ? AND deposit_id = ?
        `, txDetails.UserID, depositID)

	case "unblock":
		// Re-block the deposit
		_, err = tx.Exec(`
            UPDATE deposits
            SET is_blocked = 1
            WHERE client_id = ? AND deposit_id = ?
        `, txDetails.UserID, depositID)

	default:
		return errors.New("unsupported transaction type for cancellation")
	}

	if err != nil {
		return err
	}

	// Record the cancellation
	_, err = tx.Exec(`
        INSERT INTO cancellation_tracking (operator_id, user_id, deposit_id, transaction_id, cancelled_at)
        VALUES (?, ?, ?, ?, ?)
    `, operatorID, txDetails.UserID, depositID, transactionID, time.Now())

	if err != nil {
		return err
	}

	// Log the cancellation as a new transaction
	_, err = tx.Exec(`
        INSERT INTO transaction_history (user_id, transaction_type, amount, metadata, timestamp)
        VALUES (?, ?, ?, ?, ?)
    `, txDetails.UserID, "cancel_"+txDetails.Type, txDetails.Amount,
		fmt.Sprintf("Cancelled by operator %d, original tx: %d", operatorID, transactionID),
		time.Now())

	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetAllActionLogs retrieves and decrypts action logs
func GetAllActionLogs(startDate, endDate *time.Time, username, actionType string) ([]ActionLog, error) {
	logs := []ActionLog{}

	// Ensure tables exist
	if err := EnsureTransactionTablesExist(); err != nil {
		return logs, err
	}

	// Build query with filters
	query := `
		SELECT th.id, th.user_id, u.username, th.transaction_type, th.amount, th.metadata, th.timestamp,
			   ct.operator_id as cancelled_by, ct.cancelled_at as cancel_time
		FROM transaction_history th
		LEFT JOIN users u ON th.user_id = u.id
		LEFT JOIN cancellation_tracking ct ON th.id = ct.transaction_id
		WHERE 1=1
	`
	args := []interface{}{}

	if username != "" {
		query += " AND u.username LIKE ?"
		args = append(args, "%"+username+"%")
	}

	if actionType != "" {
		query += " AND th.transaction_type = ?"
		args = append(args, actionType)
	}

	if startDate != nil {
		query += " AND th.timestamp >= ?"
		args = append(args, startDate)
	}

	if endDate != nil {
		query += " AND th.timestamp <= ?"
		args = append(args, endDate)
	}

	query += " ORDER BY th.timestamp DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return logs, err
	}
	defer rows.Close()

	// Process rows
	for rows.Next() {
		var log ActionLog
		var cancelledBy sql.NullInt64
		var cancelTime sql.NullTime
		var encryptedMetadata string

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Username,
			&log.Type,
			&log.Amount,
			&encryptedMetadata,
			&log.Timestamp,
			&cancelledBy,
			&cancelTime,
		)
		if err != nil {
			return nil, err
		}

		// Try to decrypt metadata
		decryptedMetadata, err := utils.DecryptLogMessage(encryptedMetadata)
		if err != nil {
			// If decryption fails, use encrypted value
			log.Metadata = encryptedMetadata
		} else {
			log.Metadata = decryptedMetadata
		}

		// Handle nullable fields
		if cancelledBy.Valid {
			cancelledByInt := int(cancelledBy.Int64)
			log.CancelledBy = &cancelledByInt
		}
		if cancelTime.Valid {
			log.CancelTime = &cancelTime.Time
		}

		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return logs, err
	}

	return logs, nil
}

// CancelAllUserActions cancels all actions for a specific user
func CancelAllUserActions(userID, adminID int) (int, error) {
	if err := EnsureTransactionTablesExist(); err != nil {
		return 0, err
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Get all user's deposits
	rows, err := tx.Query(`SELECT deposit_id FROM deposits WHERE client_id = ?`, userID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var depositIDs []int64
	for rows.Next() {
		var depositID int64
		if err := rows.Scan(&depositID); err != nil {
			return 0, err
		}
		depositIDs = append(depositIDs, depositID)
	}

	if err = rows.Err(); err != nil {
		return 0, err
	}

	// If no deposits found, nothing to cancel
	if len(depositIDs) == 0 {
		return 0, errors.New("no deposits found for this user")
	}

	// Get uncancelled transactions for this user
	cancelledTransactions := 0

	for _, depositID := range depositIDs {
		// First check if admin already cancelled an action on this deposit
		var count int
		err := tx.QueryRow(`
			SELECT COUNT(*) FROM cancellation_tracking 
			WHERE user_id = ? AND deposit_id = ? AND operator_id = ?
		`, userID, depositID, adminID).Scan(&count)

		if err != nil {
			return 0, err
		}

		if count > 0 {
			// Admin already used cancellation for this deposit, skip
			continue
		}

		// Find latest uncancelled transaction for this deposit
		var txID int64
		var txType string

		err = tx.QueryRow(`
			SELECT th.id, th.transaction_type 
			FROM transaction_history th
			LEFT JOIN cancellation_tracking ct ON th.id = ct.transaction_id
			WHERE th.user_id = ? 
			AND th.transaction_type NOT IN ('delete', 'cancel_transfer', 'cancel_freeze', 'cancel_block', 'cancel_unblock')
			AND ct.transaction_id IS NULL
			ORDER BY th.timestamp DESC
			LIMIT 1
		`, userID).Scan(&txID, &txType)

		if err != nil {
			if err == sql.ErrNoRows {
				// No transactions to cancel for this deposit
				continue
			}
			return 0, err
		}

		// Perform cancellation based on transaction type
		switch txType {
		case "transfer":
			// For transfer we'd need to reverse it, but here we just log

		case "freeze":
			// Unfreeze the deposit
			_, err = tx.Exec(`
				UPDATE deposits
				SET is_frozen = 0, freeze_duration = 0, freeze_until = NULL
				WHERE client_id = ? AND deposit_id = ?
			`, userID, depositID)

		case "block":
			// Unblock the deposit
			_, err = tx.Exec(`
				UPDATE deposits
				SET is_blocked = 0
				WHERE client_id = ? AND deposit_id = ?
			`, userID, depositID)

		case "unblock":
			// Re-block the deposit
			_, err = tx.Exec(`
				UPDATE deposits
				SET is_blocked = 1
				WHERE client_id = ? AND deposit_id = ?
			`, userID, depositID)
		}

		if err != nil {
			return 0, err
		}

		// Record the cancellation
		_, err = tx.Exec(`
			INSERT INTO cancellation_tracking (operator_id, user_id, deposit_id, transaction_id, cancelled_at)
			VALUES (?, ?, ?, ?, ?)
		`, adminID, userID, depositID, txID, time.Now())

		if err != nil {
			return 0, err
		}

		// Log the cancellation action
		_, err = tx.Exec(`
			INSERT INTO transaction_history (user_id, transaction_type, metadata, timestamp)
			VALUES (?, ?, ?, ?)
		`, userID, "cancel_"+txType, fmt.Sprintf("Admin action: cancelled by admin %d", adminID), time.Now())

		if err != nil {
			return 0, err
		}

		cancelledTransactions++
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return cancelledTransactions, nil
}
