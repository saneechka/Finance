package handlers

import (
	"finance/internal/models"
	db "finance/internal/storage"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterManagerRoutes registers all manager endpoints
func RegisterManagerRoutes(router *gin.RouterGroup) {
	// Operator functionality (reuse operator handlers)
	router.GET("/transactions/statistics", GetTransactionStatistics)
	router.GET("/transactions", GetTransactions)
	router.POST("/transactions/cancel", CancelTransaction)

	// Loan management
	router.POST("/loans/review", ManagerReviewLoan)
	router.GET("/loans/pending", GetManagerLoans)
	router.POST("/loans/process", ProcessLoanRequest)
}

// GetTransactionStatistics handler for managers to get statistics (reused from operator)
func GetTransactionStatistics(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is a manager or operator
	if !hasRole(userID, "manager", "operator", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient privileges"})
		return
	}

	// Get statistics
	stats, err := db.GetTransactionStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetTransactions handler for managers to view transactions (reused from operator)
func GetTransactions(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is a manager or operator
	if !hasRole(userID, "manager", "operator", "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient privileges"})
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
	transactions, err := db.GetTransactionHistory(username, txType, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// CancelTransaction handler for managers to cancel transactions (reused from operator)
func CancelTransaction(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is a manager or operator
	if !hasRole(userID, "manager", "operator") {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient privileges"})
		return
	}

	var request struct {
		TransactionID int64 `json:"transaction_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Attempt to cancel the transaction
	if err := db.CancelTransaction(request.TransactionID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction cancelled successfully"})
}

// ManagerReviewLoan handles loan review by managers
func ManagerReviewLoan(c *gin.Context) {
	managerID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is a manager
	if !hasRole(managerID, "manager") {
		c.JSON(http.StatusForbidden, gin.H{"error": "manager privileges required"})
		return
	}

	var request struct {
		LoanID  int64  `json:"loan_id" binding:"required"`
		Action  string `json:"action" binding:"required"`
		Comment string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate action
	if request.Action != "approve" && request.Action != "reject" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action, must be either 'approve' or 'reject'"})
		return
	}

	// Get the loan to review
	loan, err := db.GetLoan(request.LoanID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "loan not found"})
		return
	}

	// Verify loan is in pending status
	if loan.Status != models.Pending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only pending loans can be reviewed"})
		return
	}

	// Process the review
	if request.Action == "approve" {
		// First approve the loan (changes status to Approved)
		if err := db.ManagerApproveLoan(request.LoanID, int64(managerID)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve loan"})
			return
		}

		// Then explicitly activate it (changes status to Active)
		if err := db.ActivateLoan(request.LoanID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "loan approved but failed to activate"})
			return
		}

		// Get user details for notification
		user, err := db.GetUserByID(int(loan.UserID))
		if err == nil && user != nil {
			// Log the approval with username
			metadata := fmt.Sprintf("Loan/installment #%d for %s approved and activated by manager #%d",
				request.LoanID, user.Username, managerID)
			db.LogTransaction(int64(managerID), "loan_manager_approval", &loan.Amount, metadata)
		}
	} else {
		// For rejection, comment is required
		if request.Comment == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "comment is required for rejection"})
			return
		}

		if err := db.ManagerRejectLoan(request.LoanID, int64(managerID), request.Comment); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reject loan"})
			return
		}

		// Get user details for notification
		user, err := db.GetUserByID(int(loan.UserID))
		if err == nil && user != nil {
			// Log the rejection with username and reason
			metadata := fmt.Sprintf("Loan/installment #%d for %s rejected by manager #%d. Reason: %s",
				request.LoanID, user.Username, managerID, request.Comment)
			db.LogTransaction(int64(managerID), "loan_manager_rejection", &loan.Amount, metadata)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "loan " + request.Action + "d successfully",
		"loan_id": request.LoanID,
	})
}

// GetManagerLoans retrieves loans that need manager review
func GetManagerLoans(c *gin.Context) {
	managerID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is a manager
	if !hasRole(managerID, "manager") {
		c.JSON(http.StatusForbidden, gin.H{"error": "manager privileges required"})
		return
	}

	status := c.Query("status")
	if status == "" {
		status = string(models.Pending)
	}

	loans, err := db.GetLoansByStatus(models.LoanStatus(status))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve loans"})
		return
	}

	// Enhance response with additional information
	loansWithUsernames := make([]map[string]interface{}, len(loans))
	for i, loan := range loans {
		// Get username for each loan
		user, err := db.GetUserByID(int(loan.UserID))
		username := "Unknown"
		if err == nil && user != nil {
			username = user.Username
		}

		loansWithUsernames[i] = map[string]interface{}{
			"id":              loan.ID,
			"user_id":         loan.UserID,
			"username":        username,
			"type":            loan.Type,
			"amount":          loan.Amount,
			"term":            loan.Term,
			"interest_rate":   loan.InterestRate,
			"total_payable":   loan.TotalPayable,
			"monthly_payment": loan.MonthlyPayment,
			"status":          loan.Status,
			"created_at":      loan.CreatedAt,
			"needs_review":    loan.Status == models.Pending,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"loans":         loansWithUsernames,
		"total_pending": len(loans),
	})
}

// ProcessLoanRequest handles new loan and installment requests
func ProcessLoanRequest(c *gin.Context) {
	managerID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	if !hasRole(managerID, "manager") {
		c.JSON(http.StatusForbidden, gin.H{"error": "manager privileges required"})
		return
	}

	var request struct {
		UserID       int64   `json:"user_id" binding:"required"`
		Type         string  `json:"type" binding:"required"` // "loan" or "installment"
		Amount       float64 `json:"amount" binding:"required"`
		Duration     int     `json:"duration" binding:"required"` // months
		Action       string  `json:"action" binding:"required"`   // "approve" or "reject"
		Comment      string  `json:"comment"`
		InterestRate float64 `json:"interest_rate,omitempty"` // only for loans
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request type
	if request.Type != "standard" && request.Type != "installment" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type, must be 'standard' or 'installment'"})
		return
	}

	if request.Action != "approve" && request.Action != "reject" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
		return
	}

	if request.Action == "approve" {
		// Create loan request
		loanRequest := models.LoanRequest{
			UserID:     request.UserID,
			Type:       models.LoanType(request.Type),
			Amount:     request.Amount,
			TermMonths: request.Duration,
		}

		// Set interest rate if provided
		if request.InterestRate > 0 {
			loanRequest.InterestRate = &request.InterestRate
		}

		// Create the loan
		loan, err := db.RequestLoan(loanRequest)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create " + request.Type})
			return
		}

		// Two-step process: First approve the loan
		if err := db.ManagerApproveLoan(loan.ID, int64(managerID)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve " + request.Type})
			return
		}

		// Then explicitly activate it
		if err := db.ActivateLoan(loan.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": request.Type + " approved but failed to activate"})
			return
		}

		// Log the approval
		userIDStr := strconv.FormatInt(request.UserID, 10)
		amount := request.Amount
		db.LogTransaction(int64(managerID), request.Type+"_approval", &amount,
			"Approved "+request.Type+" for user "+userIDStr)

		c.JSON(http.StatusOK, gin.H{
			"message": request.Type + " approved successfully",
			"user_id": request.UserID,
			"amount":  request.Amount,
			"loan_id": loan.ID,
		})
	} else {
		// Just log the rejection (no loan is created)
		userIDStr := strconv.FormatInt(request.UserID, 10)
		amount := request.Amount
		db.LogTransaction(int64(managerID), request.Type+"_rejection", &amount,
			"Rejected "+request.Type+" for user "+userIDStr+": "+request.Comment)

		c.JSON(http.StatusOK, gin.H{
			"message": request.Type + " rejected",
			"user_id": request.UserID,
		})
	}
}
