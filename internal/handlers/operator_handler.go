package handlers

import (
	"finance/internal/storage"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// hasRole checks if a user has any of the specified roles
func hasRole(userID int, roles ...string) bool {
	hasRole, err := storage.CheckUserRole(userID, roles...)
	if err != nil {
		log.Printf("Error checking user roles: %v", err)
		return false
	}
	return hasRole
}

// GetTransactionStatistics retrieves statistics about transactions for operators
func GetTransactionStatistics(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is an operator or admin
	if !hasRole(userID, "operator", "admin", "manager") {
		c.JSON(http.StatusForbidden, gin.H{"error": "operator privileges required"})
		return
	}

	// Parse period parameter if present
	period := c.DefaultQuery("period", "month")
	var startDate, endDate time.Time

	now := time.Now()
	endDate = now

	switch period {
	case "day":
		startDate = now.AddDate(0, 0, -1)
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	default: // month
		startDate = now.AddDate(0, -1, 0)
	}

	// Get transaction statistics
	stats, err := storage.GetTransactionStatistics()
	if err != nil {
		log.Printf("Error getting transaction statistics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get transaction statistics"})
		return
	}

	// Get transaction counts by type
	typeCounts, err := storage.GetTransactionCountsByType(startDate, endDate)
	if err != nil {
		log.Printf("Error getting transaction type counts: %v", err)
		// Continue anyway, we'll just return without the type breakdown
	}

	c.JSON(http.StatusOK, gin.H{
		"statistics": stats,
		"period": gin.H{
			"start": startDate.Format("2006-01-02"),
			"end":   endDate.Format("2006-01-02"),
			"name":  period,
		},
		"by_type": typeCounts,
	})
}

// GetTransactions retrieves transactions for operator review
func GetTransactions(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is an operator or admin
	if !hasRole(userID, "operator", "admin", "manager") {
		c.JSON(http.StatusForbidden, gin.H{"error": "operator privileges required"})
		return
	}

	// Parse filter parameters
	username := c.Query("username")
	txType := c.Query("type")

	var date *time.Time
	if dateStr := c.Query("date"); dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			date = &parsedDate
		}
	}

	// Get transaction history with filters
	transactions, err := storage.GetTransactionHistory(username, txType, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// CancelTransaction allows an operator to cancel a transaction

// GetUserActions retrieves user actions for operator review
func GetUserActions(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is an operator or admin
	if !hasRole(userID, "operator", "admin", "manager") {
		c.JSON(http.StatusForbidden, gin.H{"error": "operator privileges required"})
		return
	}

	// Parse filter parameters
	username := c.Query("username")
	actionType := c.Query("type")

	// Use GetAllActionLogs with filters
	logs, err := storage.GetAllActionLogs(nil, nil, username, actionType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user actions"})
		return
	}

	// Convert logs to a more suitable format for the response
	actions := make([]map[string]interface{}, len(logs))
	for i, log := range logs {
		// Determine if this action is cancellable
		isCancelled := log.CancelledBy != nil
		isLastAction := true // We'll assume it's the last action for simplicity

		actions[i] = map[string]interface{}{
			"id":             log.ID,
			"user_id":        log.UserID,
			"username":       log.Username,
			"type":           log.Type,
			"amount":         log.Amount,
			"timestamp":      log.Timestamp,
			"cancelled":      isCancelled,
			"is_last_action": isLastAction,
		}
	}

	c.JSON(http.StatusOK, gin.H{"actions": actions})
}

// GetUsers retrieves users for operator management
func GetUsers(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is an operator or admin
	if !hasRole(userID, "operator", "admin", "manager") {
		c.JSON(http.StatusForbidden, gin.H{"error": "operator privileges required"})
		return
	}

	// Get search term if provided
	searchTerm := c.Query("search")

	// Fetch users from database with optional search filter
	users, err := storage.GetAllUsers(searchTerm)
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}

	// Format the response
	var formattedUsers []map[string]interface{}

	for _, user := range users {
		// Get last action for this user
		transactions, err := storage.GetUserTransactions(user.ID, 1)
		if err != nil {
			log.Printf("Error fetching transactions for user %d: %v", user.ID, err)
			continue
		}

		var lastAction map[string]interface{}
		var hasCancellableAction bool

		if len(transactions) > 0 {
			tx := transactions[0]

			// Check if action is cancellable (not deleted and not already cancelled)
			hasCancellableAction = tx.CanCancel && tx.Type != "delete"

			lastAction = map[string]interface{}{
				"id":        tx.ID,
				"type":      tx.Type,
				"timestamp": tx.Timestamp,
				"amount":    tx.Amount,
			}
		}

		formattedUser := map[string]interface{}{
			"id":                     user.ID,
			"username":               user.Username,
			"email":                  user.Email,
			"role":                   user.Role,
			"has_cancellable_action": hasCancellableAction,
			"last_action":            lastAction,
		}

		formattedUsers = append(formattedUsers, formattedUser)
	}

	c.JSON(http.StatusOK, gin.H{
		"users": formattedUsers,
	})
}

// GetUserLastAction retrieves the last action for a specific user
func GetUserLastAction(c *gin.Context) {
	operatorID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is an operator or admin
	if !hasRole(operatorID, "operator", "admin", "manager") {
		c.JSON(http.StatusForbidden, gin.H{"error": "operator privileges required"})
		return
	}

	// Parse user ID from URL parameter
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get user details
	user, err := storage.GetUserByID(int(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Get last cancellable action for this user from the database
	transactions, err := storage.GetUserTransactions(int(userID), 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user transactions"})
		return
	}

	if len(transactions) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no transactions found for this user"})
		return
	}

	// Return the last transaction
	c.JSON(http.StatusOK, gin.H{
		"username": user.Username,
		"user_id":  userID,
		"action":   transactions[0],
	})
}

// CancelLastOperation allows an operator to cancel a user's last operation
func CancelLastOperation(c *gin.Context) {
	operatorID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is an operator or admin
	if !hasRole(operatorID, "operator", "admin", "manager") {
		c.JSON(http.StatusForbidden, gin.H{"error": "operator privileges required"})
		return
	}

	// Parse request body
	var request struct {
		UserID   int `json:"user_id"`
		ActionID int `json:"action_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	// Validate parameters
	if request.UserID <= 0 || request.ActionID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id or action_id"})
		return
	}

	// Call storage function to cancel the transaction
	err := storage.CancelTransaction(int64(request.ActionID), operatorID)
	if err != nil {
		if err.Error() == "delete operations cannot be cancelled" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Операции удаления не могут быть отменены"})
			return
		}

		log.Printf("Error cancelling transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log this operator action
	logMsg := fmt.Sprintf("Operator %d cancelled action %d for user %d", operatorID, request.ActionID, request.UserID)
	storage.LogTransaction(int64(operatorID), "operator_cancel", nil, logMsg)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Операция успешно отменена",
	})
}

// GetRecentActions retrieves recent actions for the operator dashboard
func GetRecentActions(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is an operator or admin
	if !hasRole(userID, "operator", "admin", "manager") {
		c.JSON(http.StatusForbidden, gin.H{"error": "operator privileges required"})
		return
	}

	// Get limit parameter (default to 10)
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// Get recent transactions directly from storage
	transactions, err := storage.GetRecentTransactions(limit)
	if err != nil {
		log.Printf("Error getting recent actions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get recent actions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recent_actions": transactions,
	})
}
