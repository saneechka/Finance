package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"finance/internal/models"
	db "finance/internal/storage"

	"github.com/gin-gonic/gin"
)

// Helper function to safely extract user ID
func getUserID(c *gin.Context) (int, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, false
	}

	// Try to convert to int
	id, ok := userID.(int)
	if !ok {
		// If it's not an int, try to convert from float64 (common in JSON)
		if idFloat, ok := userID.(float64); ok {
			return int(idFloat), true
		}
		return 0, false
	}

	return id, true
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// CreateDeposit now requires authentication
func CreateDeposit(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var deposit models.Deposit
	if err := c.ShouldBindJSON(&deposit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Always set the client ID to the authenticated user's ID
	// This prevents users from creating deposits for other accounts
	deposit.ClientID = int64(userID)

	// Continue with the rest of the validation
	if deposit.BankName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bank_name is required"})
		return
	}

	
	if deposit.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than 0"})
		return
	}

	if deposit.Interest < 0 || deposit.Interest > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "interest must be between 0 and 100"})
		return
	}

	if err := db.SaveDeposit(deposit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save deposit"})
		return
	}

	// Log the transaction
	amount := deposit.Amount
	_, err := db.LogTransaction(deposit.ClientID, "create", &amount,
		fmt.Sprintf("Created deposit in %s", deposit.BankName))
	if err != nil {
		log.Printf("Error logging transaction: %v", err)
	}

	c.JSON(http.StatusCreated, deposit)
}

// isAdmin checks if a user has admin privileges by querying the database
func isAdmin(userID int) bool {
	isAdmin, err := db.IsUserAdmin(userID)
	if err != nil {
		// If there's an error, log it and assume the user is not an admin
		log.Printf("Error checking if user %d is admin: %v", userID, err)
		return false
	}
	return isAdmin
}

func DeleteDeposit(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var deposit models.Deposit
	if err := c.ShouldBindJSON(&deposit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Always set client ID to the authenticated user's ID
	// This ensures users can only modify their own deposits
	deposit.ClientID = int64(userID)

	if deposit.BankName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bank_name is required"})
		return
	}

	if deposit.DepositID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deposit_id is required"})
		return
	}

	// Pass the deposit_id to the DeleteDeposit function
	if err := db.DeleteDeposit(deposit.ClientID, deposit.BankName); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "deposit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete deposit"})
		return
	}

	// Log the transaction
	_, err := db.LogTransaction(deposit.ClientID, "delete", nil,
		fmt.Sprintf("Deleted deposit %d in %s", deposit.DepositID, deposit.BankName))
	if err != nil {
		log.Printf("Error logging transaction: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "deposit deleted successfully"})
}

func TransferBetweenAccounts(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var transfer models.Transfer
	if err := c.ShouldBindJSON(&transfer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Always set client ID to the authenticated user's ID
	// This ensures users can only transfer from their own accounts
	transfer.ClientID = int64(userID)

	if transfer.ClientID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id is required"})
		return
	}

	if transfer.BankName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bank_name is required"})
		return
	}

	if transfer.FromAccount <= 0 || transfer.ToAccount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid account IDs are required"})
		return
	}

	if transfer.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than 0"})
		return
	}

	if err := db.TransferBetweenAccounts(transfer); err != nil {
		switch err {
		case sql.ErrNoRows:
			c.JSON(http.StatusNotFound, gin.H{"error": "one or both accounts not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Log the transaction
	amount := transfer.Amount
	_, err := db.LogTransaction(transfer.ClientID, "transfer", &amount,
		fmt.Sprintf("Transfer from %d to %d", transfer.FromAccount, transfer.ToAccount))
	if err != nil {
		log.Printf("Error logging transaction: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "transfer completed successfully"})
}

func BlockDeposit(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var deposit models.Deposit
	if err := c.ShouldBindJSON(&deposit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Always set client ID to the authenticated user's ID
	// This ensures users can only block their own deposits
	deposit.ClientID = int64(userID)

	if deposit.ClientID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id is required"})
		return
	}

	if deposit.BankName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bank_name is required"})
		return
	}

	if deposit.DepositID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deposit_id is required"})
		return
	}

	if err := db.BlockDeposit(deposit.ClientID, deposit.BankName, deposit.DepositID); err != nil {
		switch {
		case err.Error() == "deposit not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "deposit not found"})
		case err.Error() == "deposit is already blocked":
			c.JSON(http.StatusBadRequest, gin.H{"error": "deposit is already blocked"})
		default:
			log.Printf("Error blocking deposit: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Log the transaction
	_, err := db.LogTransaction(deposit.ClientID, "block", nil,
		fmt.Sprintf("Blocked deposit %d in %s", deposit.DepositID, deposit.BankName))
	if err != nil {
		log.Printf("Error logging transaction: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "deposit blocked successfully"})
}

func UnblockDeposit(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var deposit models.Deposit
	if err := c.ShouldBindJSON(&deposit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Always set client ID to the authenticated user's ID
	// This ensures users can only unblock their own deposits
	deposit.ClientID = int64(userID)

	if deposit.ClientID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id is required"})
		return
	}

	if deposit.BankName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bank_name is required"})
		return
	}

	if deposit.DepositID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deposit_id is required"})
		return
	}

	if err := db.UnblockDeposit(deposit.ClientID, deposit.BankName, deposit.DepositID); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "deposit not found or not blocked"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unblock deposit"})
		return
	}

	// Log the transaction
	_, err := db.LogTransaction(deposit.ClientID, "unblock", nil,
		fmt.Sprintf("Unblocked deposit %d in %s", deposit.DepositID, deposit.BankName))
	if err != nil {
		log.Printf("Error logging transaction: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "deposit unblocked successfully"})
}

func FreezeDeposit(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var deposit models.Deposit
	if err := c.ShouldBindJSON(&deposit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Always set client ID to the authenticated user's ID
	// This ensures users can only freeze their own deposits
	deposit.ClientID = int64(userID)

	if deposit.ClientID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id is required"})
		return
	}

	if deposit.BankName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bank_name is required"})
		return
	}

	if deposit.DepositID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deposit_id is required"})
		return
	}

	if deposit.FreezeDuration <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "freeze_duration must be greater than 0"})
		return
	}

	log.Printf("Attempting to freeze deposit: ID=%d, ClientID=%d, Bank=%s, Duration=%d hours",
		deposit.DepositID, deposit.ClientID, deposit.BankName, deposit.FreezeDuration)

	if err := db.FreezeDeposit(deposit.ClientID, deposit.BankName, deposit.DepositID, deposit.FreezeDuration); err != nil {
		log.Printf("Error while freezing deposit: %v", err)
		if strings.Contains(err.Error(), "deposit not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		switch {
		case err.Error() == "deposit not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "deposit not found"})
		case err.Error() == "deposit is already blocked":
			c.JSON(http.StatusBadRequest, gin.H{"error": "deposit is already blocked"})
		default:
			log.Printf("Error freezing deposit: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Log the transaction
	_, err := db.LogTransaction(deposit.ClientID, "freeze", nil,
		fmt.Sprintf("Froze deposit %d in %s for %d hours",
			deposit.DepositID, deposit.BankName, deposit.FreezeDuration))
	if err != nil {
		log.Printf("Error logging transaction: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("deposit frozen for %d hours", deposit.FreezeDuration),
	})
}

// GetDeposits retrieves all deposits for the authenticated user
func GetDeposits(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Get all deposits for this user
	deposits, err := db.GetDepositsByUserID(int64(userID))
	if err != nil {
		log.Printf("Error fetching deposits for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load deposits"})
		return
	}

	// Return empty array instead of null if no deposits found
	if deposits == nil {
		deposits = []models.Deposit{}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"deposits": deposits,
		},
	})
}
