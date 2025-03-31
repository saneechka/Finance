// Add these new functions to your deposit_store.go file

package storage

import (
	"database/sql"
	"errors"
	"log"

	
)
// var ErrAccountNotFound = errors.New("account not found")
// Add this error variable
var ErrInsufficientFunds = errors.New("insufficient funds for transfer")

// VerifyAccountsForTransfer checks if both accounts exist and belong to the right user
func VerifyAccountsForTransfer(userID int64, fromAccount, toAccount int64) (fromExists bool, toExists bool, err error) {
	// First check if the from account exists and belongs to the user
	var fromOwnerID int64
	err = DB.QueryRow("SELECT client_id FROM deposits WHERE deposit_id = ?", fromAccount).Scan(&fromOwnerID)
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
	err = DB.QueryRow("SELECT client_id FROM deposits WHERE deposit_id = ?", toAccount).Scan(&toOwnerID)
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
	var fromBlocked, fromFrozen, toBlocked, toFrozen int

	// Check from account status
	err := DB.QueryRow("SELECT is_blocked, is_frozen FROM deposits WHERE deposit_id = ?", fromAccount).Scan(&fromBlocked, &fromFrozen)
	if err != nil {
		log.Printf("Error checking from account status: %v", err)
		return false, err
	}

	log.Printf("From account (ID: %d) status - blocked: %d, frozen: %d", fromAccount, fromBlocked, fromFrozen)

	if fromBlocked == 1 || fromFrozen == 1 {
		log.Printf("From account is blocked or frozen")
		return true, nil
	}

	// Check to account status
	err = DB.QueryRow("SELECT is_blocked, is_frozen FROM deposits WHERE deposit_id = ?", toAccount).Scan(&toBlocked, &toFrozen)
	if err != nil {
		log.Printf("Error checking to account status: %v", err)
		return false, err
	}

	log.Printf("To account (ID: %d) status - blocked: %d, frozen: %d", toAccount, toBlocked, toFrozen)

	if toBlocked == 1 || toFrozen == 1 {
		log.Printf("To account is blocked or frozen")
		return true, nil
	}

	// Neither account is blocked or frozen
	log.Printf("Neither account is blocked or frozen")
	return false, nil
}
