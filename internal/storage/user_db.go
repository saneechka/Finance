package storage

import (
	"errors"
	"finance/internal/models"
	"time"
)

// EnsureUserTableExists creates the users table if it doesn't exist in the same database as deposits
func EnsureUserTableExists() error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			email TEXT,
			role TEXT DEFAULT 'client',
			approved INTEGER DEFAULT 0,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`
	_, err := db.Exec(createTableQuery)
	return err
}

// SaveUser stores a new user in the database
func SaveUser(user *models.User) error {
	// Ensure user table exists
	if err := EnsureUserTableExists(); err != nil {
		return err
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", user.Username).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("username already exists")
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// By default, new users are not approved
	query := `
		INSERT INTO users (username, password, email, role, approved, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.Exec(
		query,
		user.Username,
		user.Password,
		user.Email,
		user.Role,
		boolToInt(user.Approved), // Default is false (0)
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return err
	}

	// Get the last inserted ID
	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = int(lastID)
	return nil
}

// GetUserByUsername retrieves a user by their username
func GetUserByUsername(username string) (*models.User, error) {
	// Ensure user table exists
	if err := EnsureUserTableExists(); err != nil {
		return nil, err
	}

	user := &models.User{}
	query := `
		SELECT id, username, password, email, role, approved, created_at, updated_at
		FROM users
		WHERE username = ?
	`
	var approved int
	err := db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.Role,
		&approved,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	user.Approved = approved == 1
	return user, nil
}

// GetUserByID retrieves a user by their ID
func GetUserByID(id int) (*models.User, error) {
	// Ensure user table exists
	if err := EnsureUserTableExists(); err != nil {
		return nil, err
	}

	user := &models.User{}
	query := `
		SELECT id, username, password, email, role, approved, created_at, updated_at
		FROM users
		WHERE id = ?
	`
	var approved int
	err := db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.Role,
		&approved,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	user.Approved = approved == 1
	return user, nil
}

// IsUserAdmin checks if a user has administrator privileges
func IsUserAdmin(userID int) (bool, error) {
	var role string
	err := db.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)
	if err != nil {
		return false, err
	}

	return role == "admin", nil
}

// ApproveUser approves a user by their ID
func ApproveUser(userID int) error {
	query := `UPDATE users SET approved = 1, updated_at = ? WHERE id = ?`
	_, err := db.Exec(query, time.Now(), userID)
	return err
}

// RejectUser deletes a user by their ID (alternative to approval)
func RejectUser(userID int) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := db.Exec(query, userID)
	return err
}

// GetPendingUsers returns a list of users waiting for approval
func GetPendingUsers() ([]models.User, error) {
	users := []models.User{}

	query := `
		SELECT id, username, email, role, created_at, updated_at
		FROM users
		WHERE approved = 0
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
