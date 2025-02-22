package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"finance/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func init() {
	var err error
	DB, err = sql.Open("sqlite3", "file:deposits.db?cache=shared&_busy_timeout=9999999")
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS deposits(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		client_id INTEGER NOT NULL,
		bank_name TEXT NOT NULL,
		amount REAL NOT NULL,
		interest REAL NOT NULL,
		create_data TEXT NOT NULL
	);`

	if _, err := DB.Exec(createTableSQL); err != nil {
		log.Fatal(err)
	}

	transferTableSQL := `CREATE TABLE IF NOT EXISTS transfers(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		client_id INTEGER NOT NULL,
		bank_name TEXT NOT NULL,
		from_account INTEGER NOT NULL,
		to_account INTEGER NOT NULL,
		amount REAL NOT NULL,
		transfer_date TEXT NOT NULL
	);`
	if _, err := DB.Exec(transferTableSQL); err != nil {
		log.Fatal(err)
	}
}

func SaveDeposit(d models.Deposit) error {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	_, err := DB.Exec(`INSERT INTO deposits (client_id, bank_name, amount, interest, create_data) VALUES (?, ?, ?, ?, ?)`,
		d.ClientID, d.BankName, d.Amount, d.Interest, currentTime)
	return err
}

func DeleteDeposit(clientID int64, bankName string) error {
	result, err := DB.Exec(`DELETE FROM deposits WHERE client_id = ? AND bank_name = ?`,
		clientID, bankName)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func TransferBetweenAccounts(t models.Transfer) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	
	var count int
	err = tx.QueryRow(`SELECT COUNT(*) FROM deposits 
		WHERE client_id = ? AND bank_name = ? 
		AND id IN (?, ?)`,
		t.ClientID, t.BankName, t.FromAccount, t.ToAccount).Scan(&count)
	if err != nil {
		return err
	}
	if count != 2 {
		return sql.ErrNoRows
	}

	// Subtract from source account
	result, err := tx.Exec(`UPDATE deposits 
		SET amount = amount - ? 
		WHERE id = ? AND client_id = ? AND bank_name = ? AND amount >= ?`,
		t.Amount, t.FromAccount, t.ClientID, t.BankName, t.Amount)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return fmt.Errorf("insufficient funds or account not found")
	}

	// Add to destination account
	_, err = tx.Exec(`UPDATE deposits 
		SET amount = amount + ? 
		WHERE id = ? AND client_id = ? AND bank_name = ?`,
		t.Amount, t.ToAccount, t.ClientID, t.BankName)
	if err != nil {
		return err
	}

	// Record transfer
	_, err = tx.Exec(`INSERT INTO transfers 
		(client_id, bank_name, from_account, to_account, amount, transfer_date) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		t.ClientID, t.BankName, t.FromAccount, t.ToAccount, t.Amount,
		time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		return err
	}

	return tx.Commit()
}
