package storage

import (
	"database/sql"
	"errors"
	"finance/internal/models"
	"fmt"
	"math"
	"time"
)

// EnsureLoansTableExists creates the loans and payments tables if they don't exist
func EnsureLoansTableExists() error {
	// Create loans table
	loansTableQuery := `
		CREATE TABLE IF NOT EXISTS loans (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			loan_type TEXT NOT NULL,
			amount REAL NOT NULL,
			term_months INTEGER NOT NULL,
			interest_rate REAL NOT NULL,
			total_payable REAL NOT NULL,
			monthly_payment REAL NOT NULL,
			status TEXT NOT NULL,
			start_date TIMESTAMP,
			end_date TIMESTAMP,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			approved_by INTEGER,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`
	if _, err := db.Exec(loansTableQuery); err != nil {
		return err
	}

	// Create payments table
	paymentsTableQuery := `
		CREATE TABLE IF NOT EXISTS loan_payments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			loan_id INTEGER NOT NULL,
			amount REAL NOT NULL,
			payment_date TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL,
			FOREIGN KEY (loan_id) REFERENCES loans(id)
		)
	`
	_, err := db.Exec(paymentsTableQuery)
	return err
}

// Calculate fixed interest rate for specific loan terms
func getFixedInterestRate(termMonths int) float64 {
	switch termMonths {
	case 3:
		return 5.0 // 5% for 3-month loans
	case 6:
		return 7.5 // 7.5% for 6-month loans
	case 12:
		return 10.0 // 10% for 12-month loans
	case 24:
		return 15.0 // 15% for 24-month loans
	default:
		if termMonths > 24 {
			return 20.0 // 20% for loans over 24 months
		}
		return 12.5 // Default rate for other terms
	}
}

// Calculate loan parameters
func calculateLoanParameters(amount float64, termMonths int, interestRate float64) (float64, float64) {
	// Convert annual interest rate to monthly rate
	monthlyRate := interestRate / 100 / 12

	// Calculate total amount payable with compound interest
	totalPayable := amount * math.Pow(1+monthlyRate, float64(termMonths))

	// Calculate monthly payment
	monthlyPayment := totalPayable / float64(termMonths)

	return totalPayable, monthlyPayment
}

// RequestLoan creates a loan request in the database
func RequestLoan(request models.LoanRequest) (*models.Loan, error) {
	if err := EnsureLoansTableExists(); err != nil {
		return nil, err
	}

	// Validate request
	if request.Amount <= 0 {
		return nil, errors.New("loan amount must be greater than zero")
	}

	if request.TermMonths <= 0 {
		return nil, errors.New("loan term must be at least one month")
	}

	// Determine interest rate - use provided rate or get fixed rate based on term
	var interestRate float64
	if request.InterestRate != nil {
		interestRate = *request.InterestRate
	} else {
		interestRate = getFixedInterestRate(request.TermMonths)
	}

	// Calculate total payable amount and monthly payment
	totalPayable, monthlyPayment := calculateLoanParameters(request.Amount, request.TermMonths, interestRate)

	// Create loan record
	now := time.Now()
	loan := &models.Loan{
		UserID:         request.UserID,
		Type:           request.Type,
		Amount:         request.Amount,
		Term:           request.TermMonths,
		InterestRate:   interestRate,
		TotalPayable:   totalPayable,
		MonthlyPayment: monthlyPayment,
		Status:         models.Pending, // All loans start as pending
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Insert into database
	query := `
		INSERT INTO loans (
			user_id, loan_type, amount, term_months, interest_rate, 
			total_payable, monthly_payment, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := db.Exec(
		query,
		loan.UserID,
		loan.Type,
		loan.Amount,
		loan.Term,
		loan.InterestRate,
		loan.TotalPayable,
		loan.MonthlyPayment,
		loan.Status,
		loan.CreatedAt,
		loan.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	loan.ID = id

	// Log the transaction
	metadata := fmt.Sprintf("%s loan requested for %d months with %.2f%% interest",
		loan.Type, loan.Term, loan.InterestRate)
	LogTransaction(loan.UserID, "loan_request", &loan.Amount, metadata)

	return loan, nil
}

// ApproveLoan approves a loan request
func ApproveLoan(loanID int64, approverID int64) error {
	if err := EnsureLoansTableExists(); err != nil {
		return err
	}

	// Get the loan to make sure it exists and is in pending status
	loan, err := GetLoan(loanID)
	if err != nil {
		return err
	}

	if loan.Status != models.Pending {
		return fmt.Errorf("loan is not pending approval (current status: %s)", loan.Status)
	}

	// Calculate start and end dates
	now := time.Now()
	startDate := now
	endDate := now.AddDate(0, loan.Term, 0)

	// Update the loan status
	query := `
		UPDATE loans 
		SET status = ?, start_date = ?, end_date = ?, updated_at = ? 
		WHERE id = ?
	`
	_, err = db.Exec(
		query,
		models.Approved,
		startDate,
		endDate,
		now,
		loanID,
	)
	if err != nil {
		return err
	}

	// Log the transaction
	metadata := fmt.Sprintf("Loan #%d approved by admin #%d", loanID, approverID)
	LogTransaction(loan.UserID, "loan_approved", &loan.Amount, metadata)

	return nil
}

// RejectLoan rejects a loan request
func RejectLoan(loanID int64, approverID int64) error {
	if err := EnsureLoansTableExists(); err != nil {
		return err
	}

	// Get the loan to make sure it exists and is in pending status
	loan, err := GetLoan(loanID)
	if err != nil {
		return err
	}

	if loan.Status != models.Pending {
		return fmt.Errorf("loan is not pending approval (current status: %s)", loan.Status)
	}

	// Update the loan status
	now := time.Now()
	query := `UPDATE loans SET status = ?, updated_at = ? WHERE id = ?`
	_, err = db.Exec(query, models.Rejected, now, loanID)
	if err != nil {
		return err
	}

	// Log the transaction
	metadata := fmt.Sprintf("Loan #%d rejected by admin #%d", loanID, approverID)
	LogTransaction(loan.UserID, "loan_rejected", &loan.Amount, metadata)

	return nil
}

// ActivateLoan changes a loan from approved to active
func ActivateLoan(loanID int64) error {
	if err := EnsureLoansTableExists(); err != nil {
		return err
	}

	// Get the loan to make sure it exists and is approved
	loan, err := GetLoan(loanID)
	if err != nil {
		return err
	}

	// Only managers can activate loans, so it must be in approved status
	if loan.Status != models.Approved {
		return fmt.Errorf("loan is not approved (current status: %s)", loan.Status)
	}

	// Update the loan status
	now := time.Now()
	query := `UPDATE loans SET status = ?, updated_at = ? WHERE id = ?`
	_, err = db.Exec(query, models.Active, now, loanID)
	if err != nil {
		return err
	}

	// Log the transaction
	metadata := fmt.Sprintf("Loan #%d activated", loanID)
	LogTransaction(loan.UserID, "loan_activated", &loan.Amount, metadata)

	return nil
}

// MakePayment records a payment against a loan
func MakePayment(payment models.LoanPaymentRequest) (*models.Payment, error) {
	if err := EnsureLoansTableExists(); err != nil {
		return nil, err
	}

	// Get the loan to make sure it exists
	loan, err := GetLoan(payment.LoanID)
	if err != nil {
		return nil, err
	}

	// Only allow payments on active loans
	if loan.Status != models.Active {
		return nil, fmt.Errorf("cannot make payment on loan with status: %s", loan.Status)
	}

	// Validate payment amount
	if payment.Amount <= 0 {
		return nil, errors.New("payment amount must be greater than zero")
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Calculate total payments made so far
	var totalPayments float64
	err = tx.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM loan_payments
		WHERE loan_id = ?
	`, payment.LoanID).Scan(&totalPayments)
	if err != nil {
		return nil, err
	}

	// Add current payment
	totalPayments += payment.Amount

	// Insert the payment record
	now := time.Now()
	result, err := tx.Exec(`
		INSERT INTO loan_payments (loan_id, amount, payment_date, created_at)
		VALUES (?, ?, ?, ?)
	`, payment.LoanID, payment.Amount, now, now)
	if err != nil {
		return nil, err
	}

	paymentID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Check if loan is fully paid
	if totalPayments >= loan.TotalPayable {
		// Mark loan as completed
		_, err = tx.Exec(`
			UPDATE loans SET status = ?, updated_at = ? WHERE id = ?
		`, models.Completed, now, payment.LoanID)
		if err != nil {
			return nil, err
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	// Create the payment object to return
	newPayment := &models.Payment{
		ID:        paymentID,
		LoanID:    payment.LoanID,
		Amount:    payment.Amount,
		Date:      now,
		CreatedAt: now,
	}

	// Log the transaction
	metadata := fmt.Sprintf("Payment made on loan #%d", payment.LoanID)
	LogTransaction(loan.UserID, "loan_payment", &payment.Amount, metadata)

	return newPayment, nil
}

// GetLoan retrieves a loan by its ID
func GetLoan(loanID int64) (*models.Loan, error) {
	if err := EnsureLoansTableExists(); err != nil {
		return nil, err
	}

	query := `
		SELECT id, user_id, loan_type, amount, term_months, interest_rate, 
		       total_payable, monthly_payment, status, start_date, end_date,
		       created_at, updated_at
		FROM loans
		WHERE id = ?
	`

	loan := &models.Loan{}
	var startDate, endDate sql.NullTime
	var status string

	err := db.QueryRow(query, loanID).Scan(
		&loan.ID,
		&loan.UserID,
		&loan.Type,
		&loan.Amount,
		&loan.Term,
		&loan.InterestRate,
		&loan.TotalPayable,
		&loan.MonthlyPayment,
		&status,
		&startDate,
		&endDate,
		&loan.CreatedAt,
		&loan.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("loan not found")
		}
		return nil, err
	}

	loan.Status = models.LoanStatus(status)

	if startDate.Valid {
		loan.StartDate = &startDate.Time
	}

	if endDate.Valid {
		loan.EndDate = &endDate.Time
	}

	return loan, nil
}

// GetUserLoans retrieves all loans for a specific user
func GetUserLoans(userID int64) ([]*models.Loan, error) {
	if err := EnsureLoansTableExists(); err != nil {
		return nil, err
	}

	query := `
		SELECT id, user_id, loan_type, amount, term_months, interest_rate, 
		       total_payable, monthly_payment, status, start_date, end_date,
		       created_at, updated_at
		FROM loans
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	loans := []*models.Loan{}

	for rows.Next() {
		loan := &models.Loan{}
		var startDate, endDate sql.NullTime
		var status string

		err := rows.Scan(
			&loan.ID,
			&loan.UserID,
			&loan.Type,
			&loan.Amount,
			&loan.Term,
			&loan.InterestRate,
			&loan.TotalPayable,
			&loan.MonthlyPayment,
			&status,
			&startDate,
			&endDate,
			&loan.CreatedAt,
			&loan.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		loan.Status = models.LoanStatus(status)

		if startDate.Valid {
			loan.StartDate = &startDate.Time
		}

		if endDate.Valid {
			loan.EndDate = &endDate.Time
		}

		loans = append(loans, loan)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return loans, nil
}

// GetLoanPayments retrieves all payments for a specific loan
func GetLoanPayments(loanID int64) ([]*models.Payment, error) {
	if err := EnsureLoansTableExists(); err != nil {
		return nil, err
	}

	query := `
		SELECT id, loan_id, amount, payment_date, created_at
		FROM loan_payments
		WHERE loan_id = ?
		ORDER BY payment_date DESC
	`

	rows, err := db.Query(query, loanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	payments := []*models.Payment{}

	for rows.Next() {
		payment := &models.Payment{}

		err := rows.Scan(
			&payment.ID,
			&payment.LoanID,
			&payment.Amount,
			&payment.Date,
			&payment.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		payments = append(payments, payment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return payments, nil
}

// GetPendingLoans gets all loans with pending status
func GetPendingLoans() ([]*models.Loan, error) {
	if err := EnsureLoansTableExists(); err != nil {
		return nil, err
	}

	query := `
		SELECT l.id, l.user_id, l.loan_type, l.amount, l.term_months, 
		       l.interest_rate, l.total_payable, l.monthly_payment, l.status,
		       l.start_date, l.end_date, l.created_at, l.updated_at, u.username
		FROM loans l
		JOIN users u ON l.user_id = u.id
		WHERE l.status = ?
		ORDER BY l.created_at DESC
	`

	rows, err := db.Query(query, models.Pending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	loans := []*models.Loan{}

	for rows.Next() {
		loan := &models.Loan{}
		var startDate, endDate sql.NullTime
		var username string
		var status string

		err := rows.Scan(
			&loan.ID,
			&loan.UserID,
			&loan.Type,
			&loan.Amount,
			&loan.Term,
			&loan.InterestRate,
			&loan.TotalPayable,
			&loan.MonthlyPayment,
			&status,
			&startDate,
			&endDate,
			&loan.CreatedAt,
			&loan.UpdatedAt,
			&username,
		)

		if err != nil {
			return nil, err
		}

		loan.Status = models.LoanStatus(status)

		if startDate.Valid {
			loan.StartDate = &startDate.Time
		}

		if endDate.Valid {
			loan.EndDate = &endDate.Time
		}

		loans = append(loans, loan)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return loans, nil
}

// Function to let managers approve loans
func ManagerApproveLoan(loanID int64, managerID int64) error {
	if err := EnsureLoansTableExists(); err != nil {
		return err
	}

	// Get the loan to make sure it exists and is in pending status
	loan, err := GetLoan(loanID)
	if err != nil {
		return err
	}

	if loan.Status != models.Pending {
		return fmt.Errorf("loan is not pending approval (current status: %s)", loan.Status)
	}

	// Calculate start and end dates
	now := time.Now()
	startDate := now
	endDate := now.AddDate(0, loan.Term, 0)

	// Update the loan status to approved but not yet active
	query := `
        UPDATE loans 
        SET status = ?, start_date = ?, end_date = ?, updated_at = ?, approved_by = ? 
        WHERE id = ?
    `
	_, err = db.Exec(
		query,
		models.Approved,
		startDate,
		endDate,
		now,
		managerID,
		loanID,
	)
	if err != nil {
		return err
	}

	// Log the transaction
	metadata := fmt.Sprintf("Loan #%d approved by manager #%d", loanID, managerID)
	LogTransaction(loan.UserID, "loan_approved_manager", &loan.Amount, metadata)

	return nil
}

// Function to let managers reject loans
func ManagerRejectLoan(loanID int64, managerID int64, reason string) error {
	if err := EnsureLoansTableExists(); err != nil {
		return err
	}

	// Get the loan to make sure it exists and is in pending status
	loan, err := GetLoan(loanID)
	if err != nil {
		return err
	}

	if loan.Status != models.Pending {
		return fmt.Errorf("loan is not pending approval (current status: %s)", loan.Status)
	}

	// Update the loan status
	now := time.Now()
	query := `UPDATE loans SET status = ?, updated_at = ? WHERE id = ?`
	_, err = db.Exec(query, models.Rejected, now, loanID)
	if err != nil {
		return err
	}

	// Log the transaction with the rejection reason
	metadata := fmt.Sprintf("Loan #%d rejected by manager #%d. Reason: %s", loanID, managerID, reason)
	LogTransaction(loan.UserID, "loan_rejected_manager", &loan.Amount, metadata)

	return nil
}

// GetLoansByStatus retrieves loans with a specific status
func GetLoansByStatus(status models.LoanStatus) ([]*models.Loan, error) {
	if err := EnsureLoansTableExists(); err != nil {
		return nil, err
	}

	query := `
        SELECT l.id, l.user_id, l.loan_type, l.amount, l.term_months, 
               l.interest_rate, l.total_payable, l.monthly_payment, l.status,
               l.start_date, l.end_date, l.created_at, l.updated_at, u.username
        FROM loans l
        JOIN users u ON l.user_id = u.id
        WHERE l.status = ?
        ORDER BY l.created_at DESC
    `

	rows, err := db.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	loans := []*models.Loan{}

	for rows.Next() {
		loan := &models.Loan{}
		var startDate, endDate sql.NullTime
		var username string
		var status string

		err := rows.Scan(
			&loan.ID,
			&loan.UserID,
			&loan.Type,
			&loan.Amount,
			&loan.Term,
			&loan.InterestRate,
			&loan.TotalPayable,
			&loan.MonthlyPayment,
			&status,
			&startDate,
			&endDate,
			&loan.CreatedAt,
			&loan.UpdatedAt,
			&username,
		)

		if err != nil {
			return nil, err
		}

		loan.Status = models.LoanStatus(status)
		loan.Username = username

		if startDate.Valid {
			loan.StartDate = &startDate.Time
		}

		if endDate.Valid {
			loan.EndDate = &endDate.Time
		}

		loans = append(loans, loan)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return loans, nil
}
