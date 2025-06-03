package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/timpamungkas/loangraphql/graphqlhandler"
)

// DBService holds the database connection pool.
type DBService struct {
	Pool *pgxpool.Pool
}

// NewDBService creates a new DBService.
func NewDBService(pool *pgxpool.Pool) *DBService {
	if pool == nil {
		panic("pgxpool.Pool cannot be nil when creating DBService")
	}
	return &DBService{Pool: pool}
}

// CreateLoanApplicationDraft creates a new loan application with customer and loan details in a transaction.
func (s *DBService) CreateLoanApplicationDraft(ctx context.Context, customer graphqlhandler.CustomerData, loan graphqlhandler.ProposedLoanData, collateral graphqlhandler.CollateralData, createdBy string) (string, error) {
	customerID := uuid.New()
	loanID := uuid.New()
	now := time.Now()

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Rollback is a no-op if Commit has been called

	// Insert into customers table
	customerSQL := `
		INSERT INTO customers 
		    (id, full_name, date_of_birth, id_number, email, phone, 
		     address_street, address_city, address_zipcode, 
		     created_by, updated_by, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err = tx.Exec(ctx, customerSQL,
		customerID, customer.FullName, customer.DateOfBirth, customer.IDNumber, customer.Email, customer.Phone,
		customer.Address.Street, customer.Address.City, customer.Address.Zipcode,
		createdBy, createdBy, now, now,
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert customer: %w", err)
	}

	// Insert into loans table
	loanSQL := `
		INSERT INTO loans 
		    (id, customer_id, tenure, amount, loan_status, 
		     collateral_category, collateral_brand, collateral_variant, 
		     collateral_manufacturing_year, collateral_is_document_complete, 
		     created_by, updated_by, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	_, err = tx.Exec(ctx, loanSQL,
		loanID, customerID, loan.Tenure, loan.Amount, "DRAFT", // loan_status
		collateral.Category, collateral.Brand, collateral.Variant,
		collateral.ManufacturingYear, collateral.IsDocumentComplete,
		createdBy, createdBy, now, now,
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert loan: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return loanID.String(), nil
}

// GetLoanApplication retrieves a loan application and its associated customer data.
func (s *DBService) GetLoanApplication(ctx context.Context, loanUUID string) (*graphqlhandler.LoanApplicationData, error) {
	sql := `
		SELECT
            l.id, l.loan_status, l.tenure, l.amount,
            l.collateral_category, l.collateral_brand, l.collateral_variant,
            l.collateral_manufacturing_year, l.collateral_is_document_complete,
            l.created_at, l.updated_at,
            c.id, c.full_name, c.date_of_birth, c.id_number, c.email, c.phone,
            c.address_street, c.address_city, c.address_zipcode
        FROM loans l
        JOIN customers c ON l.customer_id = c.id
        WHERE l.id = $1 AND l.deleted = FALSE`

	var app graphqlhandler.LoanApplicationData
	// Need a placeholder for c.id as it's not directly in CustomerData
	var customerDBID uuid.UUID 
	// DateOfBirth from DB is string, CustomerData.DateOfBirth is string
	var dobString string 

	err := s.Pool.QueryRow(ctx, sql, loanUUID).Scan(
		&app.UUID, &app.Status, &app.ProposedLoan.Tenure, &app.ProposedLoan.Amount,
		&app.Collateral.Category, &app.Collateral.Brand, &app.Collateral.Variant,
		&app.Collateral.ManufacturingYear, &app.Collateral.IsDocumentComplete,
		&app.CreatedAt, &app.UpdatedAt,
		&customerDBID, &app.Customer.FullName, &dobString, &app.Customer.IDNumber, &app.Customer.Email, &app.Customer.Phone,
		&app.Customer.Address.Street, &app.Customer.Address.City, &app.Customer.Address.Zipcode,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Or a custom not found error: fmt.Errorf("loan application with ID %s not found", loanUUID)
		}
		return nil, fmt.Errorf("failed to query loan application: %w", err)
	}
    app.Customer.DateOfBirth = dobString // Assign scanned string date

	return &app, nil
}

// SubmitLoanApplication updates the loan status to 'SUBMITTED'.
func (s *DBService) SubmitLoanApplication(ctx context.Context, loanUUID string, updatedBy string) (bool, error) {
	sql := `
		UPDATE loans 
		SET loan_status = $1, updated_at = NOW(), updated_by = $2 
		WHERE id = $3 AND loan_status = 'DRAFT' AND deleted = FALSE`

	commandTag, err := s.Pool.Exec(ctx, sql, "SUBMITTED", updatedBy, loanUUID)
	if err != nil {
		return false, fmt.Errorf("failed to update loan application to submitted: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		// Could be not found, or not in DRAFT state, or already deleted.
		// Query separately if a more specific error is needed. For now, just indicate no change.
		return false, nil 
	}

	return true, nil
}

// CancelLoanApplication updates the loan status to 'CANCELLED' and marks it as deleted.
func (s *DBService) CancelLoanApplication(ctx context.Context, loanUUID string, updatedBy string) (bool, error) {
	sql := `
		UPDATE loans 
		SET loan_status = $1, updated_at = NOW(), updated_by = $2, deleted = TRUE, deleted_at = NOW() 
		WHERE id = $3 AND deleted = FALSE`

	commandTag, err := s.Pool.Exec(ctx, sql, "CANCELLED", updatedBy, loanUUID)
	if err != nil {
		return false, fmt.Errorf("failed to cancel loan application: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		// Could be not found or already deleted.
		return false, nil 
	}

	return true, nil
}
