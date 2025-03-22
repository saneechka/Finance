package storage

import (
	"database/sql"
	"errors"
	"finance/internal/models"
	"fmt"
	"time"
)

// GetUsersForOperator retrieves a list of users with their last actions
func GetUsersForOperator(searchTerm string) ([]models.User, error) {
	// Create users table if it doesn't exist
	if err := EnsureUserTableExists(); err != nil {
		return nil, err
	}

	// Create actions tables if they don't exist
	if err := ensureUserActionsTableExists(); err != nil {
		return nil, err
	}

	// Build query
	query := `
		SELECT 
			u.id, 
			u.username, 
			u.email, 
			u.role, 
			(
				SELECT COUNT(*) > 0 FROM user_actions ua 
				WHERE ua.user_id = u.id AND ua.cancelled = 0
			) AS has_cancellable_action
		FROM users u
	`

	args := []interface{}{}

	if searchTerm != "" {
		query += " WHERE u.username LIKE ? OR u.id = ?"
		args = append(args, "%"+searchTerm+"%")

		// Try to convert search term to int for ID search
		var userID int
		_, err := fmt.Sscanf(searchTerm, "%d", &userID)
		if err == nil {
			args = append(args, userID)
		} else {
			args = append(args, 0) // Impossible ID if can't convert
		}
	}

	query += " ORDER BY u.id DESC LIMIT 100"

	// Execute query
	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse results
	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.HasCancellableAction,
		)
		if err != nil {
			return nil, err
		}

		// Get the last action for each user
		lastAction, _, err := GetUserLastAction(user.ID)
		if err == nil && lastAction != nil {
			user.LastAction = lastAction
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// ensureUserActionsTableExists creates the user actions table if it doesn't exist
func ensureUserActionsTableExists() error {
	query := `
		CREATE TABLE IF NOT EXISTS user_actions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			type VARCHAR(50) NOT NULL,
			amount REAL,
			metadata TEXT,
			unix_timestamp BIGINT NOT NULL,
			cancelled BOOLEAN DEFAULT 0,
			cancelled_by INTEGER,
			cancel_timestamp BIGINT
		)
	`
	_, err := DB.Exec(query)
	return err
}

// GetUserActionsForOperator retrieves a list of user actions with filters
func GetUserActionsForOperator(username string, actionType string) ([]models.UserAction, error) {
	// Ensure tables exist
	if err := ensureUserActionsTableExists(); err != nil {
		return nil, err
	}

	// Build query
	query := `
		SELECT 
			a.id, 
			a.user_id, 
			u.username, 
			a.type, 
			a.amount, 
			a.metadata, 
			a.unix_timestamp, 
			a.cancelled,
			a.cancelled_by,
			a.cancel_timestamp,
			(
				a.id = (
					SELECT MAX(id) FROM user_actions 
					WHERE user_id = a.user_id AND cancelled = 0
				)
			) AS is_last_action
		FROM user_actions a
		LEFT JOIN users u ON a.user_id = u.id
		WHERE 1=1
	`

	args := []interface{}{}

	if username != "" {
		query += " AND u.username LIKE ?"
		args = append(args, "%"+username+"%")
	}

	if actionType != "" {
		query += " AND a.type = ?"
		args = append(args, actionType)
	}

	query += " ORDER BY a.unix_timestamp DESC LIMIT 100"

	// Execute query
	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse results
	var actions []models.UserAction
	for rows.Next() {
		var action models.UserAction
		var cancelled int
		var cancelledBy sql.NullInt64
		var cancelTimestamp sql.NullInt64
		var isLastAction int

		err := rows.Scan(
			&action.ID,
			&action.UserID,
			&action.Username,
			&action.Type,
			&action.Amount,
			&action.Metadata,
			&action.Timestamp,
			&cancelled,
			&cancelledBy,
			&cancelTimestamp,
			&isLastAction,
		)
		if err != nil {
			return nil, err
		}

		action.Cancelled = cancelled == 1
		action.IsLastAction = isLastAction == 1

		if cancelledBy.Valid {
			action.CancelledBy = int(cancelledBy.Int64)
		}

		if cancelTimestamp.Valid {
			action.CancelTime = time.Unix(cancelTimestamp.Int64, 0).Format(time.RFC3339)
		}

		actions = append(actions, action)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return actions, nil
}

// GetUserLastAction gets the last uncancelled action of a user
func GetUserLastAction(userID int) (*models.UserAction, string, error) {
	// Ensure tables exist
	if err := ensureUserActionsTableExists(); err != nil {
		return nil, "", err
	}

	// Get the username
	var username string
	err := DB.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	if err != nil {
		return nil, "", err
	}

	// Get the last uncancelled action
	query := `
		SELECT 
			id, 
			user_id, 
			type, 
			amount, 
			metadata, 
			unix_timestamp
		FROM user_actions
		WHERE user_id = ? AND cancelled = 0
		ORDER BY unix_timestamp DESC
		LIMIT 1
	`

	var action models.UserAction
	err = DB.QueryRow(query, userID).Scan(
		&action.ID,
		&action.UserID,
		&action.Type,
		&action.Amount,
		&action.Metadata,
		&action.Timestamp,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, username, nil // No actions but we found the user
		}
		return nil, username, err
	}

	return &action, username, nil
}

// GetActionDetails gets details of a specific action
func GetActionDetails(actionID int) (*models.UserAction, string, error) {
	// Ensure tables exist
	if err := ensureUserActionsTableExists(); err != nil {
		return nil, "", err
	}

	// Get action details with username
	query := `
		SELECT 
			a.id, 
			a.user_id, 
			u.username, 
			a.type, 
			a.amount, 
			a.metadata, 
			a.unix_timestamp, 
			a.cancelled,
			a.cancelled_by,
			a.cancel_timestamp
		FROM user_actions a
		LEFT JOIN users u ON a.user_id = u.id
		WHERE a.id = ?
	`

	var action models.UserAction
	var username string
	var cancelled int
	var cancelledBy sql.NullInt64
	var cancelTimestamp sql.NullInt64

	err := DB.QueryRow(query, actionID).Scan(
		&action.ID,
		&action.UserID,
		&username,
		&action.Type,
		&action.Amount,
		&action.Metadata,
		&action.Timestamp,
		&cancelled,
		&cancelledBy,
		&cancelTimestamp,
	)

	if err != nil {
		return nil, "", err
	}

	action.Cancelled = cancelled == 1

	if cancelledBy.Valid {
		action.CancelledBy = int(cancelledBy.Int64)
	}

	if cancelTimestamp.Valid {
		action.CancelTime = time.Unix(cancelTimestamp.Int64, 0).Format(time.RFC3339)
	}

	return &action, username, nil
}

// CancelUserAction cancels a user's action by an operator
func CancelUserAction(userID, actionID, operatorID int) error {
	// Ensure tables exist
	if err := ensureUserActionsTableExists(); err != nil {
		return err
	}

	// Start a transaction
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get the action to check if it's cancellable
	var actionType string
	var cancelled int
	var isLastAction int

	query := `
		SELECT 
			type, 
			cancelled,
			id = (SELECT MAX(id) FROM user_actions WHERE user_id = ? AND cancelled = 0)
		FROM user_actions
		WHERE id = ? AND user_id = ?
	`

	err = tx.QueryRow(query, userID, actionID, userID).Scan(&actionType, &cancelled, &isLastAction)
	if err != nil {
		return err
	}

	// Check if already cancelled
	if cancelled == 1 {
		return errors.New("action already cancelled")
	}

	// Check if it's the last action
	if isLastAction != 1 {
		return errors.New("only the last action can be cancelled")
	}

	// Update the action as cancelled
	now := time.Now().Unix()
	_, err = tx.Exec(
		"UPDATE user_actions SET cancelled = 1, cancelled_by = ?, cancel_timestamp = ? WHERE id = ?",
		operatorID, now, actionID,
	)
	if err != nil {
		return err
	}

	// Record the cancellation as a new action
	_, err = tx.Exec(
		"INSERT INTO user_actions (user_id, type, metadata, unix_timestamp) VALUES (?, ?, ?, ?)",
		userID, "cancel_"+actionType,
		fmt.Sprintf("Cancelled by operator ID: %d", operatorID),
		now,
	)
	if err != nil {
		return err
	}

	// If the action was related to a deposit, update the deposit status
	switch actionType {
	case "freeze":
		// Unfreeze the deposit
		_, err = tx.Exec(
			"UPDATE deposits SET is_frozen = 0, freeze_until = NULL WHERE client_id = ?",
			userID,
		)
	case "block":
		// Unblock the deposit
		_, err = tx.Exec(
			"UPDATE deposits SET is_blocked = 0 WHERE client_id = ?",
			userID,
		)
	case "unblock":
		// Re-block the deposit
		_, err = tx.Exec(
			"UPDATE deposits SET is_blocked = 1 WHERE client_id = ?",
			userID,
		)
	case "transfer":
		// For transfers, we would need to implement a reversal mechanism
		// This is a simplified version
		_, err = tx.Exec(
			"INSERT INTO user_actions (user_id, type, amount, metadata, unix_timestamp) VALUES (?, ?, ?, ?, ?)",
			userID, "reverse_transfer", 0,
			fmt.Sprintf("Automatic reversal of transfer ID: %d", actionID),
			now,
		)
	}

	if err != nil {
		return err
	}

	// Commit the transaction
	return tx.Commit()
}
