package models

type Deposit struct {
DepositID int64 `json:"deposit_id"`
ClientID int64 `json:"client_id"`
	BankName string  `json:"bank_name"`
	Amount   float64 `json:"amount"`
	Interest float64 `json:"interest"`
	CreateData string `json:"create_data"`
}

