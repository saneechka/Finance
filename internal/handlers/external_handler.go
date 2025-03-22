package handlers

import (
	"finance/internal/storage"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SubmitSalaryProject handles the submission of salary project documents by external specialists
func SubmitSalaryProject(c *gin.Context) {
	// Authentication and role check
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	if !hasRole(userID, "external") {
		c.JSON(http.StatusForbidden, gin.H{"error": "external specialist privileges required"})
		return
	}
	// Parse request body
	var request struct {
		EnterpriseID   int     `json:"enterprise_id" binding:"required"`
		EnterpriseName string  `json:"enterprise_name" binding:"required"`
		EmployeeCount  int     `json:"employee_count" binding:"required"`
		TotalAmount    float64 `json:"total_amount" binding:"required"`
		DocumentURL    string  `json:"document_url" binding:"required"`
		Comment        string  `json:"comment"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}
	// Validate parameters
	if request.EnterpriseID <= 0 || request.EmployeeCount <= 0 || request.TotalAmount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid enterprise data"})
		return
	}
	// Create salary project object
	salaryProject := &storage.SalaryProject{
		EnterpriseID:   request.EnterpriseID,
		EnterpriseName: request.EnterpriseName,
		EmployeeCount:  request.EmployeeCount,
		TotalAmount:    request.TotalAmount,
		DocumentURL:    request.DocumentURL,
		Comment:        request.Comment,
		SubmittedBy:    int(userID),
	}
	// Save the salary project to the database
	projectID, err := storage.SaveSalaryProject(salaryProject)
	if err != nil {
		log.Printf("Error saving salary project: %v", err)
		if err.Error() == "database not initialized" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection error"})
			return
		} else if err.Error() == "enterprise not found" || strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "enterprise not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit salary project: " + err.Error()})
		return
	}
	// Log the salary project submission
	metadata := "Enterprise: " + request.EnterpriseName +
		", Employees: " + strconv.Itoa(request.EmployeeCount) +
		", Document: " + request.DocumentURL
	if request.Comment != "" {
		metadata += ", Comment: " + request.Comment
	}
	storage.LogTransaction(int64(userID), "salary_project_submission", &request.TotalAmount, metadata)

	c.JSON(http.StatusOK, gin.H{
		"message":       "salary project documents submitted successfully",
		"submission_id": projectID,
		"enterprise_id": request.EnterpriseID,
		"timestamp":     time.Now().Unix(),
	})
}

// RequestEnterpriseTransfer handles transfer requests to other enterprises
func RequestEnterpriseTransfer(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	if !hasRole(userID, "external") {
		c.JSON(http.StatusForbidden, gin.H{"error": "external specialist privileges required"})
		return
	}
	// Parse request body
	var request struct {
		FromEnterpriseID int     `json:"from_enterprise_id" binding:"required"`
		ToEnterpriseID   int     `json:"to_enterprise_id" binding:"required"`
		ToEmployeeID     int     `json:"to_employee_id"`
		Amount           float64 `json:"amount" binding:"required"`
		TransferPurpose  string  `json:"transfer_purpose" binding:"required"`
		Comment          string  `json:"comment"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}
	// Validate parameters
	if request.FromEnterpriseID <= 0 || request.ToEnterpriseID <= 0 || request.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transfer data"})
		return
	}
	// Check if the requesting user is authorized for this enterprise
	isAuthorized := storage.CheckUserEnterpriseAuthorization(userID, request.FromEnterpriseID)
	if !isAuthorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized for this enterprise"})
		return
	}
	// Create a unique ID for the transfer request
	transferID := time.Now().UnixNano()
	// Determine if this is an enterprise-to-enterprise or enterprise-to-employee transfer
	transferType := "enterprise_transfer"
	transferDetails := "To Enterprise ID: " + strconv.Itoa(request.ToEnterpriseID)
	if request.ToEmployeeID > 0 {
		transferType = "employee_transfer"
		transferDetails = "To Employee ID: " + strconv.Itoa(request.ToEmployeeID) + " at Enterprise ID: " + strconv.Itoa(request.ToEnterpriseID)
	}
	// Add transfer purpose and optional comment to metadata
	metadata := transferDetails + ", Purpose: " + request.TransferPurpose
	if request.Comment != "" {
		metadata += ", Comment: " + request.Comment
	}
	// Log the transfer request (not executing it automatically)
	storage.LogTransaction(int64(userID), transferType+"_request", &request.Amount, metadata)

	c.JSON(http.StatusOK, gin.H{
		"message":            "transfer request submitted successfully",
		"transfer_id":        transferID,
		"from_enterprise_id": request.FromEnterpriseID,
		"to_enterprise_id":   request.ToEnterpriseID,
		"amount":             request.Amount,
		"status":             "pending",
		"timestamp":          time.Now().Unix(),
	})
}

// GetEnterpriseTransfers retrieves all transfer requests for an enterprise
func GetEnterpriseTransfers(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	if !hasRole(userID, "external") {
		c.JSON(http.StatusForbidden, gin.H{"error": "external specialist privileges required"})
		return
	}
	// Get enterprise ID from query params
	enterpriseIDStr := c.Query("enterprise_id")
	if enterpriseIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "enterprise_id parameter is required"})
		return
	}
	enterpriseID, err := strconv.Atoi(enterpriseIDStr)
	if err != nil || enterpriseID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid enterprise_id"})
		return
	}
	// Check if the requesting user is authorized for this enterprise
	isAuthorized := storage.CheckUserEnterpriseAuthorization(userID, enterpriseID)
	if !isAuthorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized for this enterprise"})
		return
	}
	// Get transfer status filter if provided
	status := c.Query("status")
	// Retrieve transfers for the enterprise
	transfers, err := storage.GetEnterpriseTransfers(enterpriseID, status)
	if err != nil {
		log.Printf("Error fetching enterprise transfers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve transfers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"transfers": transfers,
		"count":     len(transfers),
	})
}

// GetSalaryProjects retrieves all salary project submissions for an enterprise
func GetSalaryProjects(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	if !hasRole(userID, "external") {
		c.JSON(http.StatusForbidden, gin.H{"error": "external specialist privileges required"})
		return
	}
	// Get enterprise ID from query params
	enterpriseIDStr := c.Query("enterprise_id")
	if enterpriseIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "enterprise_id parameter is required"})
		return
	}
	enterpriseID, err := strconv.Atoi(enterpriseIDStr)
	if err != nil || enterpriseID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid enterprise_id"})
		return
	}
	// Check if the requesting user is authorized for this enterprise
	isAuthorized := storage.CheckUserEnterpriseAuthorization(userID, enterpriseID)
	if !isAuthorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized for this enterprise"})
		return
	}
	// Retrieve salary projects for the enterprise
	projects, err := storage.GetEnterpriseSalaryProjects(enterpriseID)
	if err != nil {
		log.Printf("Error fetching enterprise salary projects: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve salary projects"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"salary_projects": projects,
		"count":           len(projects),
	})
}

// GetUserEnterprises retrieves all enterprises associated with the authenticated user
func GetUserEnterprises(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	if !hasRole(userID, "external") {
		c.JSON(http.StatusForbidden, gin.H{"error": "external specialist privileges required"})
		return
	}
	enterprises, err := storage.GetUserEnterprises(userID)
	if err != nil {
		log.Printf("Error fetching user enterprises: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve enterprises"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"enterprises": enterprises,
		"count":       len(enterprises),
	})
}
