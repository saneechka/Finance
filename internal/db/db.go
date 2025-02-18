package db

import (
	"database/sql"
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
}

func SaveDeposit(d models.Deposit) error {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	_, err := DB.Exec(`INSERT INTO deposits (client_id, bank_name, amount, interest, create_data) VALUES (?, ?, ?, ?, ?)`,
		d.ClientID, d.BankName, d.Amount, d.Interest, currentTime)
	return err
}
