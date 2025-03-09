package handlers

import (
	// Add this import
	db "finance/internal/storage"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Get transaction statistics for operators
func GetTransactionStatistics(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Verify user is an operator or admin
	if !hasRole(userID, "operator", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "operator privileges required"})
		return
	}

	statistics, err := db.GetTransactionStatistics()
	if err != nil {
		log.Printf("Error fetching transaction statistics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get transaction statistics"})
		return
	}

	c.JSON(http.StatusOK, statistics)
}

// Get transaction history for operators
func GetTransactions(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Verify user is an operator or admin
	if !hasRole(userID, "operator", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "operator privileges required"})
		return
	}

	// Get filter parameters
	username := c.Query("username")
	txType := c.Query("type")
	dateStr := c.Query("date")

	var date *time.Time
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			date = &parsedDate
		}
	}

	transactions, err := db.GetTransactionHistory(username, txType, date)
	if err != nil {
		log.Printf("Error fetching transactions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// Cancel a transaction (operators can cancel one transaction per account)
func CancelTransaction(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Verify user is an operator or admin
	if !hasRole(userID, "operator", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "operator privileges required"})
		return
	}

	var request struct {
		TransactionID int64 `json:"transaction_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := db.CancelTransaction(request.TransactionID, userID)
	if err != nil {
		log.Printf("Error cancelling transaction: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transaction cancelled successfully"})
}
