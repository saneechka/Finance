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

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func CreateDeposit(c *gin.Context) {
	var deposit models.Deposit
	if err := c.ShouldBindJSON(&deposit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if deposit.ClientID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id is required"})
		return
	}

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
	c.JSON(http.StatusCreated, deposit)
}

func DeleteDeposit(c *gin.Context) {
	var deposit models.Deposit
	if err := c.ShouldBindJSON(&deposit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if deposit.ClientID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id is required"})
		return
	}

	if deposit.BankName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bank_name is required"})
		return
	}

	if err := db.DeleteDeposit(deposit.ClientID, deposit.BankName); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "deposit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete deposit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deposit deleted successfully"})
}

func TransferBetweenAccounts(c *gin.Context) {
	var transfer models.Transfer
	if err := c.ShouldBindJSON(&transfer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	c.JSON(http.StatusOK, gin.H{"message": "transfer completed successfully"})
}

func BlockDeposit(c *gin.Context) {
	var deposit models.Deposit
	if err := c.ShouldBindJSON(&deposit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	c.JSON(http.StatusOK, gin.H{"message": "deposit blocked successfully"})
}

func UnblockDeposit(c *gin.Context) {
	var deposit models.Deposit
	if err := c.ShouldBindJSON(&deposit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	c.JSON(http.StatusOK, gin.H{"message": "deposit unblocked successfully"})
}

func FreezeDeposit(c *gin.Context) {
	var deposit models.Deposit
	if err := c.ShouldBindJSON(&deposit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("deposit frozen for %d hours", deposit.FreezeDuration),
	})
}
