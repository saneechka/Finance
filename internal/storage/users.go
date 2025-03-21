package storage

import (
	"database/sql"
	"fmt"
	"strings"
)

// User represents a user in the system
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// GetAllUsers retrieves all users, optionally filtered by search term
func GetAllUsers(searchTerm string) ([]User, error) {
	users := []User{}

	// Base query
	query := `
		SELECT id, username, email, role 
		FROM users 
		WHERE 1=1
	`

	args := []interface{}{}

	// Add search condition if provided
	if searchTerm != "" {
		query += " AND (username LIKE ? OR email LIKE ? OR id = ?)"
		searchPattern := "%" + searchTerm + "%"

		// Try to convert searchTerm to int for ID matching
		var searchID int
		_, err := fmt.Sscanf(searchTerm, "%d", &searchID)
		if err != nil {
			searchID = 0 // Default to 0 if not a valid number
		}

		args = append(args, searchPattern, searchPattern, searchID)
	}

	query += " ORDER BY id DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Role); err != nil {
			return users, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}




// func NewGetUserTransactions(userID int64)(*User,error){
// 	var new_user User
// 	err:=db.QuerruRow(
		
// 	)
// }



// GetUserByID retrieves a user by their ID
func GetUserByID(userID int) (*User, error) {
	var user User

	err := db.QueryRow(
		"SELECT id, username, email, role FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// GetUserTransactions retrieves transactions for a specific user with limit
func GetUserTransactions(userID int, limit int) ([]Transaction, error) {
	transactions := []Transaction{}

	query := `
		SELECT th.id, th.user_id, u.username, th.transaction_type, th.amount, th.timestamp,
			   (NOT EXISTS (SELECT 1 FROM cancellation_tracking ct WHERE ct.transaction_id = th.id)) AS can_cancel
		FROM transaction_history th
		LEFT JOIN users u ON th.user_id = u.id
		WHERE th.user_id = ?
		ORDER BY th.timestamp DESC
	`

	if limit > 0 {
		query += " LIMIT ?"
	}

	var rows *sql.Rows
	var err error

	if limit > 0 {
		rows, err = db.Query(query, userID, limit)
	} else {
		rows, err = db.Query(query, userID)
	}

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

		// Set cancellable flag
		tx.CanCancel = canCancel

		transactions = append(transactions, tx)
	}

	return transactions, rows.Err()
}

// IsUserAdmin checks if a user has admin role

// CheckUserRole checks if a user has any of the specified roles
func CheckUserRole(userID int, roles ...string) (bool, error) {
	var role string
	err := db.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // User not found
		}
		return false, err
	}

	userRole := strings.ToLower(role)
	for _, r := range roles {
		if userRole == strings.ToLower(r) {
			return true, nil
		}
	}

	return false, nil
}
