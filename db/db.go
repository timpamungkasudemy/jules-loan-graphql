package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/timpamungkas/loangraphql/model" // Changed import
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
// It now returns the fully populated LoanApplication model.
func (s *DBService) CreateLoanApplicationDraft(ctx context.Context, customerIn model.CustomerInput, loanIn model.ProposedLoanInput, collateralIn model.CollateralInput, createdBy string) (*model.LoanApplication, error) {
	customerUUID := uuid.New()
	loanUUID := uuid.New()
	now := time.Now()

	// Prepare customer data for insertion
	dbCustomer := model.Customer{
		ID:          customerUUID.String(),
		FullName:    customerIn.FullName,
		DateOfBirth: customerIn.DateOfBirth, // Assuming YYYY-MM-DD string format
		IDNumber:    customerIn.IDNumber,
		Email:       customerIn.Email,
		Phone:       customerIn.Phone,
		Address: model.Address{ // Convert from AddressInput
			Street:  customerIn.Address.Street,
			City:    customerIn.Address.City,
			Zipcode: customerIn.Address.Zipcode,
		},
		CreatedBy: createdBy,
		UpdatedBy: createdBy, // Initially same as createdBy
		CreatedAt: now,
		UpdatedAt: now,
		Deleted:   false,
		DeletedAt: nil,
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert into customers table
	customerSQL := `
		INSERT INTO customers
		    (id, full_name, date_of_birth, id_number, email, phone,
		     address_street, address_city, address_zipcode,
		     created_by, updated_by, created_at, updated_at, deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	_, err = tx.Exec(ctx, customerSQL,
		dbCustomer.ID, dbCustomer.FullName, dbCustomer.DateOfBirth, dbCustomer.IDNumber, dbCustomer.Email, dbCustomer.Phone,
		dbCustomer.Address.Street, dbCustomer.Address.City, dbCustomer.Address.Zipcode,
		dbCustomer.CreatedBy, dbCustomer.UpdatedBy, dbCustomer.CreatedAt, dbCustomer.UpdatedAt, dbCustomer.Deleted,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert customer: %w", err)
	}

	// Prepare loan application data for insertion
	dbLoanApp := &model.LoanApplication{
		ID:         loanUUID.String(),
		CustomerID: dbCustomer.ID,
		Status:     "DRAFT",
		ProposedLoan: model.ProposedLoan{ // Convert from ProposedLoanInput
			Tenure: loanIn.Tenure,
			Amount: loanIn.Amount,
		},
		Collateral: model.Collateral{ // Convert from CollateralInput
			Category:           collateralIn.Category,
			Brand:              collateralIn.Brand,
			Variant:            collateralIn.Variant,
			ManufacturingYear:  collateralIn.ManufacturingYear,
			IsDocumentComplete: collateralIn.IsDocumentComplete,
		},
		CreatedBy:    createdBy,
		UpdatedBy:    createdBy, // Initially same as createdBy
		CreatedAt:    now,
		UpdatedAt:    now,
		Deleted:      false,
		DeletedAt:    nil,
		CustomerData: dbCustomer, // Embed the full customer data for the return
	}

	// Insert into loans table
	loanSQL := `
		INSERT INTO loans
		    (id, customer_id, tenure, amount, loan_status,
		     collateral_category, collateral_brand, collateral_variant,
		     collateral_manufacturing_year, collateral_is_document_complete,
		     created_by, updated_by, created_at, updated_at, deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
	_, err = tx.Exec(ctx, loanSQL,
		dbLoanApp.ID, dbLoanApp.CustomerID, dbLoanApp.ProposedLoan.Tenure, dbLoanApp.ProposedLoan.Amount, dbLoanApp.Status,
		dbLoanApp.Collateral.Category, dbLoanApp.Collateral.Brand, dbLoanApp.Collateral.Variant,
		dbLoanApp.Collateral.ManufacturingYear, dbLoanApp.Collateral.IsDocumentComplete,
		dbLoanApp.CreatedBy, dbLoanApp.UpdatedBy, dbLoanApp.CreatedAt, dbLoanApp.UpdatedAt, dbLoanApp.Deleted,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert loan: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return dbLoanApp, nil
}

// GetLoanApplication retrieves a loan application and its associated customer data.
func (s *DBService) GetLoanApplication(ctx context.Context, loanUUID string) (*model.LoanApplication, error) {
	sql := `
		SELECT
            l.id, l.status, l.tenure, l.amount,
            l.collateral_category, l.collateral_brand, l.collateral_variant,
            l.collateral_manufacturing_year, l.collateral_is_document_complete,
            l.created_at, l.updated_at, l.created_by, l.updated_by, l.deleted, l.deleted_at, l.customer_id,
            c.id, c.full_name, c.date_of_birth, c.id_number, c.email, c.phone,
            c.address_street, c.address_city, c.address_zipcode,
            c.created_at, c.updated_at, c.created_by, c.updated_by, c.deleted, c.deleted_at
        FROM loans l
        JOIN customers c ON l.customer_id = c.id
        WHERE l.id = $1 AND l.deleted = FALSE` // Query for non-deleted loans

	app := model.LoanApplication{} // Target struct

	err := s.Pool.QueryRow(ctx, sql, loanUUID).Scan(
		// Loan fields
		&app.ID, &app.Status, &app.ProposedLoan.Tenure, &app.ProposedLoan.Amount,
		&app.Collateral.Category, &app.Collateral.Brand, &app.Collateral.Variant,
		&app.Collateral.ManufacturingYear, &app.Collateral.IsDocumentComplete,
		&app.CreatedAt, &app.UpdatedAt, &app.CreatedBy, &app.UpdatedBy, &app.Deleted, &app.DeletedAt, &app.CustomerID,
		// Customer fields (for app.CustomerData)
		&app.CustomerData.ID, &app.CustomerData.FullName, &app.CustomerData.DateOfBirth,
		&app.CustomerData.IDNumber, &app.CustomerData.Email, &app.CustomerData.Phone,
		&app.CustomerData.Address.Street, &app.CustomerData.Address.City, &app.CustomerData.Address.Zipcode,
		&app.CustomerData.CreatedAt, &app.CustomerData.UpdatedAt, &app.CustomerData.CreatedBy, &app.CustomerData.UpdatedBy,
		&app.CustomerData.Deleted, &app.CustomerData.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Standard way to indicate not found
		}
		return nil, fmt.Errorf("failed to query loan application: %w", err)
	}
	// Ensure CustomerID in CustomerData is consistent if it wasn't scanned directly (it was: c.id -> app.CustomerData.ID)
	// and Loan's CustomerID is also set (it was: l.customer_id -> app.CustomerID)

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

	return commandTag.RowsAffected() > 0, nil
}

// CancelLoanApplication updates the loan status to 'CANCELLED' and marks it as deleted (soft delete).
func (s *DBService) CancelLoanApplication(ctx context.Context, loanUUID string, updatedBy string) (bool, error) {
	sql := `
		UPDATE loans
		SET loan_status = $1, updated_at = NOW(), updated_by = $2, deleted = TRUE, deleted_at = NOW()
		WHERE id = $3 AND deleted = FALSE`

	commandTag, err := s.Pool.Exec(ctx, sql, "CANCELLED", updatedBy, loanUUID)
	if err != nil {
		return false, fmt.Errorf("failed to cancel loan application: %w", err)
	}

	return commandTag.RowsAffected() > 0, nil
}
