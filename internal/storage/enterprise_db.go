package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

// EnterpriseTransfer represents a transfer between enterprises or to an employee
type EnterpriseTransfer struct {
	ID               int64   `json:"id"`
	FromEnterpriseID int     `json:"from_enterprise_id"`
	ToEnterpriseID   int     `json:"to_enterprise_id"`
	ToEmployeeID     int     `json:"to_employee_id,omitempty"`
	Amount           float64 `json:"amount"`
	Status           string  `json:"status"`
	Purpose          string  `json:"purpose"`
	Comment          string  `json:"comment,omitempty"`
	RequestedBy      int     `json:"requested_by"`
	RequestedAt      int64   `json:"requested_at"`
	ProcessedBy      int     `json:"processed_by,omitempty"`
	ProcessedAt      int64   `json:"processed_at,omitempty"`
}

// SalaryProject represents a salary project submission
type SalaryProject struct {
	ID             int64   `json:"id"`
	EnterpriseID   int     `json:"enterprise_id"`
	EnterpriseName string  `json:"enterprise_name"`
	EmployeeCount  int     `json:"employee_count"`
	TotalAmount    float64 `json:"total_amount"`
	DocumentURL    string  `json:"document_url"`
	Comment        string  `json:"comment,omitempty"`
	Status         string  `json:"status"` // Status can be: "pending", "approved", "rejected"
	SubmittedBy    int     `json:"submitted_by"`
	SubmittedAt    int64   `json:"submitted_at"`
	ProcessedBy    int     `json:"processed_by,omitempty"`
	ProcessedAt    int64   `json:"processed_at,omitempty"`
}

// Enterprise represents a company or organization
type Enterprise struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedAt   int64  `json:"created_at"`
}

// EnterpriseUser represents a relationship between a user and an enterprise
type EnterpriseUser struct {
	UserID       int    `json:"user_id"`
	EnterpriseID int    `json:"enterprise_id"`
	Role         string `json:"role"`
	AssignedAt   int64  `json:"assigned_at"`
}

// SalaryPayment represents a payment to an employee in a salary project
type SalaryPayment struct {
	ID               int64   `json:"id"`
	ProjectID        int64   `json:"project_id"`
	EmployeeName     string  `json:"employee_name"`
	EmployeePosition string  `json:"employee_position"`
	Amount           float64 `json:"amount"`
	AccountNumber    string  `json:"account_number"`
	BankName         string  `json:"bank_name"`
	PaymentPurpose   string  `json:"payment_purpose"`
	DocumentURL      string  `json:"document_url"`
	Status           string  `json:"status"` // Status can be: "pending", "approved", "rejected"
	CreatedAt        int64   `json:"created_at"`
}

// EnsureEnterpriseTablesExist creates required tables for enterprise functionality
func EnsureEnterpriseTablesExist() error {
	// Create enterprises table
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS enterprises (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			created_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create enterprise_users table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS enterprise_users (
			user_id INTEGER NOT NULL,
			enterprise_id INTEGER NOT NULL,
			role TEXT NOT NULL,
			assigned_at INTEGER NOT NULL,
			PRIMARY KEY (user_id, enterprise_id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (enterprise_id) REFERENCES enterprises(id)
		)
	`)
	if err != nil {
		return err
	}

	// Create enterprise_transfers table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS enterprise_transfers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			from_enterprise_id INTEGER NOT NULL,
			to_enterprise_id INTEGER NOT NULL,
			to_employee_id INTEGER,
			amount REAL NOT NULL,
			status TEXT NOT NULL,
			purpose TEXT NOT NULL,
			comment TEXT,
			requested_by INTEGER NOT NULL,
			requested_at INTEGER NOT NULL,
			processed_by INTEGER,
			processed_at INTEGER,
			FOREIGN KEY (from_enterprise_id) REFERENCES enterprises(id),
			FOREIGN KEY (to_enterprise_id) REFERENCES enterprises(id),
			FOREIGN KEY (requested_by) REFERENCES users(id),
			FOREIGN KEY (processed_by) REFERENCES users(id)
		)
	`)
	if err != nil {
		return err
	}

	// Create salary_projects table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS salary_projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			enterprise_id INTEGER NOT NULL,
			enterprise_name TEXT NOT NULL,
			employee_count INTEGER NOT NULL,
			total_amount REAL NOT NULL,
			document_url TEXT NOT NULL,
			comment TEXT,
			status TEXT NOT NULL,
			submitted_by INTEGER NOT NULL,
			submitted_at INTEGER NOT NULL,
			processed_by INTEGER,
			processed_at INTEGER,
			FOREIGN KEY (enterprise_id) REFERENCES enterprises(id),
			FOREIGN KEY (submitted_by) REFERENCES users(id),
			FOREIGN KEY (processed_by) REFERENCES users(id)
		)
	`)
	if err != nil {
		return err
	}

	// Create salary_payments table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS salary_payments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL,
			employee_name TEXT NOT NULL,
			employee_position TEXT NOT NULL,
			amount REAL NOT NULL,
			account_number TEXT NOT NULL,
			bank_name TEXT NOT NULL,
			payment_purpose TEXT NOT NULL,
			document_url TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at INTEGER NOT NULL,
			FOREIGN KEY (project_id) REFERENCES salary_projects(id)
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

// CheckUserEnterpriseAuthorization checks if a user is authorized for a specific enterprise
func CheckUserEnterpriseAuthorization(userID, enterpriseID int) bool {
	if err := EnsureEnterpriseTablesExist(); err != nil {
		log.Printf("Error ensuring enterprise tables exist: %v", err)
		return false
	}

	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM enterprise_users
		WHERE user_id = ? AND enterprise_id = ?
	`, userID, enterpriseID).Scan(&count)

	if err != nil {
		log.Printf("Error checking enterprise authorization: %v", err)
		return false
	}

	return count > 0
}

// GetEnterpriseTransfers retrieves transfers for a specific enterprise
func GetEnterpriseTransfers(enterpriseID int, status string) ([]EnterpriseTransfer, error) {
	if err := EnsureEnterpriseTablesExist(); err != nil {
		return nil, err
	}

	query := `
		SELECT 
			id, from_enterprise_id, to_enterprise_id, to_employee_id,
			amount, status, purpose, comment,
			requested_by, requested_at, processed_by, processed_at
		FROM enterprise_transfers
		WHERE from_enterprise_id = ?
	`
	args := []interface{}{enterpriseID}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY requested_at DESC"

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transfers []EnterpriseTransfer
	for rows.Next() {
		var transfer EnterpriseTransfer
		var toEmployeeID sql.NullInt64
		var comment sql.NullString
		var processedBy sql.NullInt64
		var processedAt sql.NullInt64

		err := rows.Scan(
			&transfer.ID, &transfer.FromEnterpriseID, &transfer.ToEnterpriseID, &toEmployeeID,
			&transfer.Amount, &transfer.Status, &transfer.Purpose, &comment,
			&transfer.RequestedBy, &transfer.RequestedAt, &processedBy, &processedAt,
		)
		if err != nil {
			return nil, err
		}

		// Convert nullable fields
		if toEmployeeID.Valid {
			transfer.ToEmployeeID = int(toEmployeeID.Int64)
		}
		if comment.Valid {
			transfer.Comment = comment.String
		}
		if processedBy.Valid {
			transfer.ProcessedBy = int(processedBy.Int64)
		}
		if processedAt.Valid {
			transfer.ProcessedAt = processedAt.Int64
		}

		transfers = append(transfers, transfer)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transfers, nil
}

// GetEnterpriseSalaryProjects retrieves salary projects for a specific enterprise
func GetEnterpriseSalaryProjects(enterpriseID int) ([]SalaryProject, error) {
	if err := EnsureEnterpriseTablesExist(); err != nil {
		return nil, err
	}

	query := `
		SELECT 
			id, enterprise_id, enterprise_name, employee_count,
			total_amount, document_url, comment, status,
			submitted_by, submitted_at, processed_by, processed_at
		FROM salary_projects
		WHERE enterprise_id = ?
		ORDER BY submitted_at DESC
	`

	rows, err := DB.Query(query, enterpriseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []SalaryProject
	for rows.Next() {
		var project SalaryProject
		var comment sql.NullString
		var processedBy sql.NullInt64
		var processedAt sql.NullInt64

		err := rows.Scan(
			&project.ID, &project.EnterpriseID, &project.EnterpriseName, &project.EmployeeCount,
			&project.TotalAmount, &project.DocumentURL, &comment, &project.Status,
			&project.SubmittedBy, &project.SubmittedAt, &processedBy, &processedAt,
		)
		if err != nil {
			return nil, err
		}

		// Convert nullable fields
		if comment.Valid {
			project.Comment = comment.String
		}
		if processedBy.Valid {
			project.ProcessedBy = int(processedBy.Int64)
		}
		if processedAt.Valid {
			project.ProcessedAt = processedAt.Int64
		}

		projects = append(projects, project)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

// GetUserEnterprises retrieves all enterprises associated with a user
func GetUserEnterprises(userID int) ([]Enterprise, error) {
	if err := EnsureEnterpriseTablesExist(); err != nil {
		return nil, err
	}

	query := `
		SELECT e.id, e.name, e.description, e.created_at
		FROM enterprises e
		JOIN enterprise_users eu ON e.id = eu.enterprise_id
		WHERE eu.user_id = ?
		ORDER BY e.name
	`

	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enterprises []Enterprise
	for rows.Next() {
		var enterprise Enterprise
		var description sql.NullString

		err := rows.Scan(
			&enterprise.ID,
			&enterprise.Name,
			&description,
			&enterprise.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if description.Valid {
			enterprise.Description = description.String
		}

		enterprises = append(enterprises, enterprise)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return enterprises, nil
}

// SaveEnterpriseTransfer saves a new enterprise transfer request to the database
func SaveEnterpriseTransfer(transfer *EnterpriseTransfer) (int64, error) {
	if err := EnsureEnterpriseTablesExist(); err != nil {
		return 0, err
	}

	result, err := DB.Exec(`
		INSERT INTO enterprise_transfers (
			from_enterprise_id, to_enterprise_id, to_employee_id,
			amount, status, purpose, comment,
			requested_by, requested_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		transfer.FromEnterpriseID, transfer.ToEnterpriseID, transfer.ToEmployeeID,
		transfer.Amount, transfer.Status, transfer.Purpose, transfer.Comment,
		transfer.RequestedBy, time.Now().Unix(),
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// SaveSalaryProject saves a new salary project to the database
func SaveSalaryProject(project *SalaryProject) (int64, error) {
	if err := EnsureEnterpriseTablesExist(); err != nil {
		return 0, err
	}

	// Set default status to pending
	project.Status = "pending"
	project.SubmittedAt = time.Now().Unix()

	result, err := DB.Exec(`
		INSERT INTO salary_projects (
			enterprise_id, enterprise_name, employee_count,
			total_amount, document_url, comment, status,
			submitted_by, submitted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		project.EnterpriseID,
		project.EnterpriseName,
		project.EmployeeCount,
		project.TotalAmount,
		project.DocumentURL,
		project.Comment,
		project.Status,
		project.SubmittedBy,
		project.SubmittedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetPendingSalaryProjects retrieves all pending salary projects
func GetPendingSalaryProjects() ([]SalaryProject, error) {
	if err := EnsureEnterpriseTablesExist(); err != nil {
		return nil, err
	}

	query := `
		SELECT 
			id, enterprise_id, enterprise_name, employee_count,
			total_amount, document_url, comment, status,
			submitted_by, submitted_at, processed_by, processed_at
		FROM salary_projects
		WHERE status = 'pending'
		ORDER BY submitted_at DESC
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []SalaryProject
	for rows.Next() {
		var project SalaryProject
		var comment sql.NullString
		var processedBy sql.NullInt64
		var processedAt sql.NullInt64

		err := rows.Scan(
			&project.ID,
			&project.EnterpriseID,
			&project.EnterpriseName,
			&project.EmployeeCount,
			&project.TotalAmount,
			&project.DocumentURL,
			&comment,
			&project.Status,
			&project.SubmittedBy,
			&project.SubmittedAt,
			&processedBy,
			&processedAt,
		)
		if err != nil {
			return nil, err
		}

		if comment.Valid {
			project.Comment = comment.String
		}
		if processedBy.Valid {
			project.ProcessedBy = int(processedBy.Int64)
		}
		if processedAt.Valid {
			project.ProcessedAt = processedAt.Int64
		}

		projects = append(projects, project)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

// GetPendingTransfers retrieves all pending transfer requests
func GetPendingTransfers() ([]EnterpriseTransfer, error) {
	if err := EnsureEnterpriseTablesExist(); err != nil {
		return nil, err
	}

	query := `
		SELECT 
			id, from_enterprise_id, to_enterprise_id, to_employee_id,
			amount, status, purpose, comment,
			requested_by, requested_at, processed_by, processed_at
		FROM enterprise_transfers
		WHERE status = 'pending'
		ORDER BY requested_at DESC
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transfers []EnterpriseTransfer
	for rows.Next() {
		var transfer EnterpriseTransfer
		var toEmployeeID sql.NullInt64
		var comment sql.NullString
		var processedBy sql.NullInt64
		var processedAt sql.NullInt64

		err := rows.Scan(
			&transfer.ID, &transfer.FromEnterpriseID, &transfer.ToEnterpriseID, &toEmployeeID,
			&transfer.Amount, &transfer.Status, &transfer.Purpose, &comment,
			&transfer.RequestedBy, &transfer.RequestedAt, &processedBy, &processedAt,
		)
		if err != nil {
			return nil, err
		}

		// Convert nullable fields
		if toEmployeeID.Valid {
			transfer.ToEmployeeID = int(toEmployeeID.Int64)
		}
		if comment.Valid {
			transfer.Comment = comment.String
		}
		if processedBy.Valid {
			transfer.ProcessedBy = int(processedBy.Int64)
		}
		if processedAt.Valid {
			transfer.ProcessedAt = processedAt.Int64
		}

		transfers = append(transfers, transfer)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transfers, nil
}

// ApproveSalaryProject approves a salary project submission
func ApproveSalaryProject(projectID, adminID int64, comment string) error {
	if err := EnsureEnterpriseTablesExist(); err != nil {
		return err
	}

	// Check if project exists and is in pending status
	var status string
	err := DB.QueryRow("SELECT status FROM salary_projects WHERE id = ?", projectID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("salary project not found")
		}
		return err
	}

	if status != "pending" {
		return errors.New("only pending salary projects can be approved")
	}

	// Update the salary project status
	_, err = DB.Exec(`
		UPDATE salary_projects
		SET status = 'approved', processed_by = ?, processed_at = ?, comment = CASE WHEN ? <> '' THEN ? ELSE comment END
		WHERE id = ?
	`, adminID, time.Now().Unix(), comment, comment, projectID)
	if err != nil {
		return err
	}

	// Log the approval
	var projectInfo struct {
		EnterpriseID   int
		EnterpriseName string
		TotalAmount    float64
		SubmittedBy    int64
	}

	err = DB.QueryRow(`
		SELECT enterprise_id, enterprise_name, total_amount, submitted_by
		FROM salary_projects WHERE id = ?
	`, projectID).Scan(&projectInfo.EnterpriseID, &projectInfo.EnterpriseName, &projectInfo.TotalAmount, &projectInfo.SubmittedBy)

	if err == nil {
		metadata := fmt.Sprintf("Approved salary project #%d for enterprise %s (ID: %d), submitted by user ID %d",
			projectID, projectInfo.EnterpriseName, projectInfo.EnterpriseID, projectInfo.SubmittedBy)
		if comment != "" {
			metadata += ". Comment: " + comment
		}
		LogTransaction(adminID, "salary_project_approval", &projectInfo.TotalAmount, metadata)
	}

	return nil
}

// RejectSalaryProject rejects a salary project submission
func RejectSalaryProject(projectID, adminID int64, reason string) error {
	if err := EnsureEnterpriseTablesExist(); err != nil {
		return err
	}

	// Check if project exists and is in pending status
	var status string
	err := DB.QueryRow("SELECT status FROM salary_projects WHERE id = ?", projectID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("salary project not found")
		}
		return err
	}

	if status != "pending" {
		return errors.New("only pending salary projects can be rejected")
	}

	// Update the salary project status
	_, err = DB.Exec(`
		UPDATE salary_projects
		SET status = 'rejected', processed_by = ?, processed_at = ?, comment = ?
		WHERE id = ?
	`, adminID, time.Now().Unix(), reason, projectID)
	if err != nil {
		return err
	}

	// Log the rejection
	var projectInfo struct {
		EnterpriseID   int
		EnterpriseName string
		TotalAmount    float64
		SubmittedBy    int64
	}

	err = DB.QueryRow(`
		SELECT enterprise_id, enterprise_name, total_amount, submitted_by
		FROM salary_projects WHERE id = ?
	`, projectID).Scan(&projectInfo.EnterpriseID, &projectInfo.EnterpriseName, &projectInfo.TotalAmount, &projectInfo.SubmittedBy)

	if err == nil {
		metadata := fmt.Sprintf("Rejected salary project #%d for enterprise %s (ID: %d), submitted by user ID %d. Reason: %s",
			projectID, projectInfo.EnterpriseName, projectInfo.EnterpriseID, projectInfo.SubmittedBy, reason)
		LogTransaction(adminID, "salary_project_rejection", &projectInfo.TotalAmount, metadata)
	}

	return nil
}

// ApproveEnterpriseTransfer approves an enterprise transfer request
func ApproveEnterpriseTransfer(transferID, adminID int64, comment string) error {
	if err := EnsureEnterpriseTablesExist(); err != nil {
		return err
	}

	// Check if transfer exists and is in pending status
	var status string
	err := DB.QueryRow("SELECT status FROM enterprise_transfers WHERE id = ?", transferID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("transfer request not found")
		}
		return err
	}

	if status != "pending" {
		return errors.New("only pending transfers can be approved")
	}

	// Update the transfer status
	_, err = DB.Exec(`
		UPDATE enterprise_transfers
		SET status = 'approved', processed_by = ?, processed_at = ?, comment = CASE WHEN ? <> '' THEN ? ELSE comment END
		WHERE id = ?
	`, adminID, time.Now().Unix(), comment, comment, transferID)
	if err != nil {
		return err
	}

	// Log the approval
	var transferInfo struct {
		FromEnterpriseID int
		ToEnterpriseID   int
		ToEmployeeID     sql.NullInt64
		Amount           float64
		Purpose          string
		RequestedBy      int64
	}

	err = DB.QueryRow(`
		SELECT from_enterprise_id, to_enterprise_id, to_employee_id, amount, purpose, requested_by
		FROM enterprise_transfers WHERE id = ?
	`, transferID).Scan(
		&transferInfo.FromEnterpriseID,
		&transferInfo.ToEnterpriseID,
		&transferInfo.ToEmployeeID,
		&transferInfo.Amount,
		&transferInfo.Purpose,
		&transferInfo.RequestedBy,
	)

	if err == nil {
		recipientText := fmt.Sprintf("enterprise ID %d", transferInfo.ToEnterpriseID)
		if transferInfo.ToEmployeeID.Valid {
			recipientText = fmt.Sprintf("employee ID %d at enterprise ID %d",
				transferInfo.ToEmployeeID.Int64, transferInfo.ToEnterpriseID)
		}

		metadata := fmt.Sprintf("Approved transfer #%d from enterprise ID %d to %s, requested by user ID %d. Purpose: %s",
			transferID, transferInfo.FromEnterpriseID, recipientText, transferInfo.RequestedBy, transferInfo.Purpose)
		if comment != "" {
			metadata += ". Comment: " + comment
		}
		LogTransaction(adminID, "transfer_approval", &transferInfo.Amount, metadata)
	}

	return nil
}

// RejectEnterpriseTransfer rejects an enterprise transfer request
func RejectEnterpriseTransfer(transferID, adminID int64, reason string) error {
	if err := EnsureEnterpriseTablesExist(); err != nil {
		return err
	}

	// Check if transfer exists and is in pending status
	var status string
	err := DB.QueryRow("SELECT status FROM enterprise_transfers WHERE id = ?", transferID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("transfer request not found")
		}
		return err
	}

	if status != "pending" {
		return errors.New("only pending transfers can be rejected")
	}

	// Update the transfer status
	_, err = DB.Exec(`
		UPDATE enterprise_transfers
		SET status = 'rejected', processed_by = ?, processed_at = ?, comment = ?
		WHERE id = ?
	`, adminID, time.Now().Unix(), reason, transferID)
	if err != nil {
		return err
	}

	// Log the rejection
	var transferInfo struct {
		FromEnterpriseID int
		ToEnterpriseID   int
		ToEmployeeID     sql.NullInt64
		Amount           float64
		Purpose          string
		RequestedBy      int64
	}

	err = DB.QueryRow(`
		SELECT from_enterprise_id, to_enterprise_id, to_employee_id, amount, purpose, requested_by
		FROM enterprise_transfers WHERE id = ?
	`, transferID).Scan(
		&transferInfo.FromEnterpriseID,
		&transferInfo.ToEnterpriseID,
		&transferInfo.ToEmployeeID,
		&transferInfo.Amount,
		&transferInfo.Purpose,
		&transferInfo.RequestedBy,
	)

	if err == nil {
		recipientText := fmt.Sprintf("enterprise ID %d", transferInfo.ToEnterpriseID)
		if transferInfo.ToEmployeeID.Valid {
			recipientText = fmt.Sprintf("employee ID %d at enterprise ID %d",
				transferInfo.ToEmployeeID.Int64, transferInfo.ToEnterpriseID)
		}

		metadata := fmt.Sprintf("Rejected transfer #%d from enterprise ID %d to %s, requested by user ID %d. Reason: %s",
			transferID, transferInfo.FromEnterpriseID, recipientText, transferInfo.RequestedBy, reason)
		LogTransaction(adminID, "transfer_rejection", &transferInfo.Amount, metadata)
	}

	return nil
}

// UpdateSalaryProjectStatus updates the status of a salary project
func UpdateSalaryProjectStatus(projectID int64, status string, processedBy int) error {
	if err := EnsureEnterpriseTablesExist(); err != nil {
		return err
	}

	if status != "approved" && status != "rejected" {
		return errors.New("invalid status")
	}

	result, err := DB.Exec(`
		UPDATE salary_projects
		SET status = ?, processed_by = ?, processed_at = ?
		WHERE id = ? AND status = 'pending'
	`,
		status,
		processedBy,
		time.Now().Unix(),
		projectID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("project not found or already processed")
	}

	return nil
}

// CreateEnterprise creates a new enterprise
func CreateEnterprise(name string, description string) (int64, error) {
	if err := EnsureEnterpriseTablesExist(); err != nil {
		return 0, err
	}

	result, err := DB.Exec(`
		INSERT INTO enterprises (name, description, created_at)
		VALUES (?, ?, ?)
	`,
		name,
		description,
		time.Now().Unix(),
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// AssociateUserWithEnterprise associates a user with an enterprise
func AssociateUserWithEnterprise(userID int, enterpriseID int, role string) error {
	if err := EnsureEnterpriseTablesExist(); err != nil {
		return err
	}

	_, err := DB.Exec(`
		INSERT INTO enterprise_users (user_id, enterprise_id, role, assigned_at)
		VALUES (?, ?, ?, ?)
	`,
		userID,
		enterpriseID,
		role,
		time.Now().Unix(),
	)
	if err != nil {
		return err
	}

	return nil
}

// SaveSalaryPayment saves a new salary payment to the database
func SaveSalaryPayment(payment *SalaryPayment) (int64, error) {
	if err := EnsureEnterpriseTablesExist(); err != nil {
		return 0, err
	}

	// Set default status to pending
	payment.Status = "pending"
	payment.CreatedAt = time.Now().Unix()

	result, err := DB.Exec(`
		INSERT INTO salary_payments (
			project_id, employee_name, employee_position,
			amount, account_number, bank_name,
			payment_purpose, document_url, status,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		payment.ProjectID,
		payment.EmployeeName,
		payment.EmployeePosition,
		payment.Amount,
		payment.AccountNumber,
		payment.BankName,
		payment.PaymentPurpose,
		payment.DocumentURL,
		payment.Status,
		payment.CreatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetSalaryProjectPayments retrieves all payments for a specific salary project
func GetSalaryProjectPayments(projectID int64) ([]SalaryPayment, error) {
	if err := EnsureEnterpriseTablesExist(); err != nil {
		return nil, err
	}

	query := `
		SELECT 
			id, project_id, employee_name, employee_position,
			amount, account_number, bank_name,
			payment_purpose, document_url, status,
			created_at
		FROM salary_payments
		WHERE project_id = ?
		ORDER BY created_at DESC
	`

	rows, err := DB.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []SalaryPayment
	for rows.Next() {
		var payment SalaryPayment
		err := rows.Scan(
			&payment.ID,
			&payment.ProjectID,
			&payment.EmployeeName,
			&payment.EmployeePosition,
			&payment.Amount,
			&payment.AccountNumber,
			&payment.BankName,
			&payment.PaymentPurpose,
			&payment.DocumentURL,
			&payment.Status,
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

// SaveMultipleSalaryPayments saves multiple salary payments to the database
func SaveMultipleSalaryPayments(payments []*SalaryPayment) ([]int64, error) {
	if err := EnsureEnterpriseTablesExist(); err != nil {
		return nil, err
	}

	// Start a transaction
	tx, err := DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var paymentIDs []int64
	for _, payment := range payments {
		// Set default status to pending
		payment.Status = "pending"
		payment.CreatedAt = time.Now().Unix()

		result, err := tx.Exec(`
			INSERT INTO salary_payments (
				project_id, employee_name, employee_position,
				amount, account_number, bank_name,
				payment_purpose, document_url, status,
				created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			payment.ProjectID,
			payment.EmployeeName,
			payment.EmployeePosition,
			payment.Amount,
			payment.AccountNumber,
			payment.BankName,
			payment.PaymentPurpose,
			payment.DocumentURL,
			payment.Status,
			payment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		paymentIDs = append(paymentIDs, id)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return paymentIDs, nil
}
