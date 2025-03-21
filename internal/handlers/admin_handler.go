package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"finance/internal/storage"
	"finance/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetAllActionLogs retrieves all action logs for admin viewing
func GetAllActionLogs(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is admin
	isAdmin, err := storage.IsUserAdmin(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify admin status"})
		return
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
		return
	}

	// Get query parameters for filtering
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	username := c.Query("username")
	actionType := c.Query("type")

	var startDate, endDate *time.Time
	if startDateStr != "" {
		parsedTime, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			startDate = &parsedTime
		}
	}

	if endDateStr != "" {
		parsedTime, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			// Set to end of day
			parsedTime = parsedTime.Add(24*time.Hour - time.Second)
			endDate = &parsedTime
		}
	}

	// Get logs from database
	logs, err := storage.GetAllActionLogs(startDate, endDate, username, actionType)
	if err != nil {
		log.Printf("Error fetching action logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get action logs"})
		return
	}

	// Also check if we have a system log file to send
	systemLogs, err := readSystemLogFile()

	c.JSON(http.StatusOK, gin.H{
		"database_logs": logs,
		"system_logs":   systemLogs,
	})
}

// readSystemLogFile reads and decrypts the system log file if it exists
func readSystemLogFile() ([]string, error) {
	logFilePath := filepath.Join(os.TempDir(), "finance_system.log")

	// Check if file exists
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		return []string{}, nil
	}

	// Read encrypted log file
	encryptedLines, err := os.ReadFile(logFilePath)
	if err != nil {
		return nil, err
	}

	// Split by lines
	encryptedEntries := strings.Split(string(encryptedLines), "\n")

	// Decrypt each line
	decryptedLines := make([]string, 0, len(encryptedEntries))
	for _, entry := range encryptedEntries {
		if entry == "" {
			continue
		}

		// Try to decrypt
		decryptedEntry, err := utils.DecryptLogMessage(entry)
		if err != nil {
			// If decryption fails, use the encrypted entry
			decryptedLines = append(decryptedLines, "Encrypted: "+entry)
		} else {
			decryptedLines = append(decryptedLines, decryptedEntry)
		}
	}

	return decryptedLines, nil
}

// CancelAllUserActions allows admin to cancel actions for a specific user
func CancelAllUserActions(c *gin.Context) {
	adminID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is admin
	isAdmin, err := storage.IsUserAdmin(adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify admin status"})
		return
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
		return
	}

	var request struct {
		UserID int `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cancel all user actions
	count, err := storage.CancelAllUserActions(request.UserID, adminID)
	if err != nil {
		log.Printf("Error cancelling user actions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "user actions cancelled successfully",
		"cancelled_count": count,
	})
}

// GetPendingExternalRequests retrieves all pending requests from external specialists
func GetPendingExternalRequests(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is admin
	isAdmin, err := storage.IsUserAdmin(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify admin status"})
		return
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
		return
	}

	// Get request type from query parameter (salary_project or transfer)
	requestType := c.Query("type")
	if requestType != "" && requestType != "salary_project" && requestType != "transfer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request type, must be 'salary_project' or 'transfer'"})
		return
	}

	var salaryProjects []storage.SalaryProject
	var transfers []storage.EnterpriseTransfer

	if requestType == "" || requestType == "salary_project" {
		// Get pending salary projects
		salaryProjects, err = storage.GetPendingSalaryProjects()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve pending salary projects"})
			return
		}
	}

	if requestType == "" || requestType == "transfer" {
		// Get pending transfers
		transfers, err = storage.GetPendingTransfers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve pending transfers"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"salary_projects": salaryProjects,
		"transfers":       transfers,
	})
}

// ApproveExternalRequest approves a request from an external specialist
func ApproveExternalRequest(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is admin
	isAdmin, err := storage.IsUserAdmin(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify admin status"})
		return
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
		return
	}

	// Parse request body
	var request struct {
		RequestID   int64  `json:"request_id" binding:"required"`
		RequestType string `json:"request_type" binding:"required"`
		Comment     string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	// Validate request type
	if request.RequestType != "salary_project" && request.RequestType != "transfer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request type, must be 'salary_project' or 'transfer'"})
		return
	}

	// Process based on request type
	switch request.RequestType {
	case "salary_project":
		err = storage.ApproveSalaryProject(request.RequestID, int64(userID), request.Comment)
	case "transfer":
		err = storage.ApproveEnterpriseTransfer(request.RequestID, int64(userID), request.Comment)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to approve %s: %v", request.RequestType, err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    fmt.Sprintf("%s approved successfully", request.RequestType),
		"request_id": request.RequestID,
	})
}

// RejectExternalRequest rejects a request from an external specialist
func RejectExternalRequest(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is admin
	isAdmin, err := storage.IsUserAdmin(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify admin status"})
		return
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
		return
	}

	// Parse request body
	var request struct {
		RequestID   int64  `json:"request_id" binding:"required"`
		RequestType string `json:"request_type" binding:"required"`
		Reason      string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	// Validate request type
	if request.RequestType != "salary_project" && request.RequestType != "transfer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request type, must be 'salary_project' or 'transfer'"})
		return
	}

	// Require a reason for rejection
	if request.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reason is required when rejecting a request"})
		return
	}

	// Process based on request type
	switch request.RequestType {
	case "salary_project":
		err = storage.RejectSalaryProject(request.RequestID, int64(userID), request.Reason)
	case "transfer":
		err = storage.RejectEnterpriseTransfer(request.RequestID, int64(userID), request.Reason)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to reject %s: %v", request.RequestType, err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    fmt.Sprintf("%s rejected successfully", request.RequestType),
		"request_id": request.RequestID,
	})
}
