package models

import "time"

// LoanType represents the different types of loans available
type LoanType string

const (
	// Regular loan types
	StandardLoan    LoanType = "standard"
	InstallmentPlan LoanType = "installment"
)

// LoanTerm represents loan duration in months
type LoanTerm int

const (
	ThreeMonths  LoanTerm = 3
	SixMonths    LoanTerm = 6
	TwelveMonths LoanTerm = 12
	TwoYears     LoanTerm = 24
	Custom       LoanTerm = 0 // For custom duration
)


type LoanStatus string

const (
	Pending   LoanStatus = "pending"
	Approved  LoanStatus = "approved"
	Active    LoanStatus = "active"
	Completed LoanStatus = "completed"
	Rejected  LoanStatus = "rejected"
	Default   LoanStatus = "default"
)

// Loan represents a loan or installment plan
type Loan struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	Username       string     `json:"username,omitempty"`
	Type           LoanType   `json:"type"`
	Amount         float64    `json:"amount"`
	Term           int        `json:"term_months"` 
	InterestRate   float64    `json:"interest_rate"`
	TotalPayable   float64    `json:"total_payable"`
	MonthlyPayment float64    `json:"monthly_payment"`
	Status         LoanStatus `json:"status"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ApprovedBy     *int64     `json:"approved_by,omitempty"`
	RejectedBy     *int64     `json:"rejected_by,omitempty"`
	ApprovedAt     *time.Time `json:"approved_at,omitempty"`
	RejectedAt     *time.Time `json:"rejected_at,omitempty"`

}


type Payment struct {
	ID        int64     `json:"id"`
	LoanID    int64     `json:"loan_id"`
	Amount    float64   `json:"amount"`
	Date      time.Time `json:"date"`
	CreatedAt time.Time `json:"created_at"`
}

//sample of request
type LoanRequest struct {
	UserID       int64    `json:"user_id"`
	Type         LoanType `json:"type"`
	Amount       float64  `json:"amount"`
	TermMonths   int      `json:"term_months"`
	InterestRate *float64 `json:"interest_rate,omitempty"` // Optional custom rate
}

// LoanPaymentRequest represents a request to make a payment on a loan
type LoanPaymentRequest struct {
	LoanID int64   `json:"loan_id"`
	Amount float64 `json:"amount"`
}
