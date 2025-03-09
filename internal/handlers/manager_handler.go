package handlers

import (
	"finance/internal/models"
	db "finance/internal/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
		if err := db.ApproveLoan(request.LoanID, int64(managerID)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve loan"})
			return
		}
		if err := db.ActivateLoan(request.LoanID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "loan approved but failed to activate"})
			return
		}
	} else {
		if err := db.RejectLoan(request.LoanID, int64(managerID)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reject loan"})
			return
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

	loans, err := db.GetPendingLoans() // Reusing existing function
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve loans"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"loans": loans})
}
