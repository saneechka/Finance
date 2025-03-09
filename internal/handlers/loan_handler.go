package handlers

import (
	"finance/internal/models"
	db "finance/internal/storage"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLoan handles a loan request from a user
func RequestLoan(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var request models.LoanRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the user ID from the authenticated user
	request.UserID = int64(userID)

	// Validate loan amount
	if request.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "loan amount must be greater than zero"})
		return
	}

	// Validate loan term
	if request.TermMonths <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "loan term must be at least one month"})
		return
	}

	// Validate custom interest rate if provided
	if request.InterestRate != nil && *request.InterestRate < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "interest rate cannot be negative"})
		return
	}

	// Validate loan type
	if request.Type != models.StandardLoan && request.Type != models.InstallmentPlan {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid loan type, must be either 'standard' or 'installment'",
		})
		return
	}

	// Create the loan request
	loan, err := db.RequestLoan(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create loan request: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "loan request submitted successfully",
		"loan":    loan,
	})
}

// GetUserLoans retrieves all loans for the authenticated user
func GetUserLoans(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	loans, err := db.GetUserLoans(int64(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve loans: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"loans": loans})
}

// GetLoanDetails retrieves details for a specific loan
func GetLoanDetails(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	loanIDStr := c.Param("id")
	loanID, err := strconv.ParseInt(loanIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan ID"})
		return
	}

	// Get the loan
	loan, err := db.GetLoan(loanID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "loan not found: " + err.Error()})
		return
	}

	// Check if the user is authorized to view this loan
	if loan.UserID != int64(userID) && !hasRole(userID, "admin", "operator") {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not authorized to view this loan"})
		return
	}

	// Get loan payments
	payments, err := db.GetLoanPayments(loanID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve loan payments: " + err.Error()})
		return
	}

	// Calculate remaining amount
	var paidAmount float64
	for _, payment := range payments {
		paidAmount += payment.Amount
	}
	remainingAmount := loan.TotalPayable - paidAmount

	// Calculate payment progress
	progressPercent := (paidAmount / loan.TotalPayable) * 100

	// Add time-related information
	var timeInfo map[string]interface{}
	if loan.StartDate != nil && loan.EndDate != nil {
		now := time.Now()

		// Calculate elapsed and remaining duration
		elapsedDuration := now.Sub(*loan.StartDate)
		remainingDuration := loan.EndDate.Sub(now)

		// Express durations in days
		elapsedDays := int(elapsedDuration.Hours() / 24)
		remainingDays := int(remainingDuration.Hours() / 24)
		totalDays := int(loan.EndDate.Sub(*loan.StartDate).Hours() / 24)

		// Calculate time progress
		timeProgressPercent := float64(elapsedDays) / float64(totalDays) * 100

		timeInfo = map[string]interface{}{
			"elapsed_days":          elapsedDays,
			"remaining_days":        remainingDays,
			"total_days":            totalDays,
			"time_progress_percent": timeProgressPercent,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"loan":             loan,
		"payments":         payments,
		"paid_amount":      paidAmount,
		"remaining_amount": remainingAmount,
		"progress_percent": progressPercent,
		"time_info":        timeInfo,
	})
}

// MakeLoanPayment handles a payment on a loan
func MakeLoanPayment(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var paymentRequest models.LoanPaymentRequest
	if err := c.ShouldBindJSON(&paymentRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the payment amount
	if paymentRequest.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment amount must be greater than zero"})
		return
	}

	// Get the loan to verify ownership
	loan, err := db.GetLoan(paymentRequest.LoanID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "loan not found: " + err.Error()})
		return
	}

	// Check if the user is authorized to make payments on this loan
	if loan.UserID != int64(userID) && !hasRole(userID, "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not authorized to make payments on this loan"})
		return
	}

	// Make the payment
	payment, err := db.MakePayment(paymentRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process payment: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "payment processed successfully",
		"payment": payment,
	})
}

// GetPendingLoans retrieves all pending loan requests (admin only)
func GetPendingLoans(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is admin
	if !hasRole(userID, "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
		return
	}

	// Get pending loans
	loans, err := db.GetPendingLoans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve pending loans: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pending_loans": loans})
}

// ApproveLoan approves a loan request (admin only)
func ApproveLoan(c *gin.Context) {
	adminID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is admin
	if !hasRole(adminID, "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
		return
	}

	var request struct {
		LoanID int64 `json:"loan_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Approve the loan
	if err := db.ApproveLoan(request.LoanID, int64(adminID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve loan: " + err.Error()})
		return
	}

	// Activate the loan immediately after approval
	if err := db.ActivateLoan(request.LoanID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "loan approved but failed to activate: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "loan approved and activated successfully"})
}

// RejectLoan rejects a loan request (admin only)
func RejectLoan(c *gin.Context) {
	adminID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is admin
	if !hasRole(adminID, "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
		return
	}

	var request struct {
		LoanID int64 `json:"loan_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Reject the loan
	if err := db.RejectLoan(request.LoanID, int64(adminID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reject loan: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "loan rejected successfully"})
}

// GetLoanRates returns the current fixed interest rates for different loan terms
func GetLoanRates(c *gin.Context) {
	// Define the fixed rates
	rates := map[string]float64{
		"3":      5.0,  // 3-month loans: 5%
		"6":      7.5,  // 6-month loans: 7.5%
		"12":     10.0, // 12-month loans: 10%
		"24":     15.0, // 24-month loans: 15%
		"custom": 20.0, // Loans over 24 months: 20%
	}

	c.JSON(http.StatusOK, gin.H{
		"rates": rates,
		"note":  "Custom interest rates can be requested for special circumstances",
	})
}
