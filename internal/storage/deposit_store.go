// Add these new functions to your deposit_store.go file

package storage

import (
	"database/sql"
	"errors"
	// ... existing imports
)

// Add this error variable
var ErrInsufficientFunds = errors.New("insufficient funds for transfer")

// VerifyAccountsForTransfer checks if both accounts exist and belong to the right user
func VerifyAccountsForTransfer(userID int64, fromAccount, toAccount int64) (fromExists bool, toExists bool, err error) {
	// First check if the from account exists and belongs to the user
	var fromOwnerID int64
	err = db.QueryRow("SELECT client_id FROM deposits WHERE deposit_id = $1", fromAccount).Scan(&fromOwnerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, false, nil // From account does not exist
		}
		return false, false, err
	}

	// Check if the from account belongs to the user
	if fromOwnerID != userID {
		return false, false, nil // User does not own the from account
	}

	// Now check if the to account exists
	var toOwnerID int64
	err = db.QueryRow("SELECT client_id FROM deposits WHERE deposit_id = $1", toAccount).Scan(&toOwnerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, false, nil // From exists, to does not
		}
		return true, false, err
	}

	// Both accounts exist
	return true, true, nil
}

// CheckAccountsBlockedOrFrozen checks if either account is blocked or frozen
func CheckAccountsBlockedOrFrozen(fromAccount, toAccount int64) (bool, error) {
	var fromStatus, toStatus string

	// Check from account status
	err := db.QueryRow("SELECT status FROM deposits WHERE deposit_id = $1", fromAccount).Scan(&fromStatus)
	if err != nil {
		return false, err
	}

	if fromStatus == "blocked" || fromStatus == "frozen" {
		return true, nil
	}

	// Check to account status
	err = db.QueryRow("SELECT status FROM deposits WHERE deposit_id = $1", toAccount).Scan(&toStatus)
	if err != nil {
		return false, err
	}

	if toStatus == "blocked" || toStatus == "frozen" {
		return true, nil
	}

	// Neither account is blocked or frozen
	return false, nil
}
