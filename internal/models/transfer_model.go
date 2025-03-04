package models

type Transfer struct {
	ClientID    int64   `json:"client_id"`
	BankName    string  `json:"bank_name"`
	FromAccount int64   `json:"from_account"`
	ToAccount   int64   `json:"to_account"`
	Amount      float64 `json:"amount"`
}
