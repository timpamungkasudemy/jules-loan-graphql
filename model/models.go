package model

import (
	"time"
	// No external project dependencies for the models themselves, only standard library.
)

// Address represents an address.
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Zipcode string `json:"zipcode"`
}

// Customer represents customer data.
// Note: ID here is the UUID string for the customer record itself.
// IDNumber is the national/document ID number.
type Customer struct {
	ID          string     `json:"id"` // UUID for the customer record
	FullName    string     `json:"full_name"`
	DateOfBirth string     `json:"date_of_birth"` // Keep as string, validation/conversion at boundary
	IDNumber    string     `json:"id_number"`     // National/document ID
	Email       string     `json:"email,omitempty"`
	Phone       string     `json:"phone"`
	Address     Address    `json:"address"` // Embedded struct
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CreatedBy   string     `json:"created_by"`
	UpdatedBy   string     `json:"updated_by"`
	Deleted     bool       `json:"-"` // Often excluded from JSON response unless specifically needed
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// Collateral represents loan collateral.
type Collateral struct {
	Category           string `json:"category"` // CAR, MOTORCYCLE
	Brand              string `json:"brand"`
	Variant            string `json:"variant"`
	ManufacturingYear  int    `json:"manufacturing_year"`
	IsDocumentComplete bool   `json:"is_document_complete"`
}

// ProposedLoan represents the terms of a proposed loan.
type ProposedLoan struct {
	Tenure int     `json:"tenure"`
	Amount float64 `json:"amount"`
}

// LoanApplication represents the entire loan application.
// This struct will be used for database interaction and can also be used
// as a base for GraphQL responses, potentially with some fields omitted or transformed.
type LoanApplication struct {
	ID             string       `json:"id"` // UUID for the loan record
	CustomerID     string       `json:"customer_id"` // Foreign key to Customer.ID
	Status         string       `json:"status"`      // DRAFT, SUBMITTED, CANCELLED
	ProposedLoan   ProposedLoan `json:"proposed_loan"` // Embedded struct
	Collateral     Collateral   `json:"collateral"`    // Embedded struct
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	CreatedBy      string       `json:"created_by"`
	UpdatedBy      string       `json:"updated_by"`
	Deleted        bool         `json:"-"`
	DeletedAt      *time.Time   `json:"deleted_at,omitempty"`

	// Customer details can be included here when fetching a full loan application view.
	// This matches how LoanApplicationData was structured previously in graphqlhandler.
	CustomerData Customer `json:"customer"`
}

// Input types for mutations (mirroring GraphQL inputs but in Go, used by resolvers before DB interaction)
// These are distinct from the DB models above where some fields are auto-generated (ID, CreatedAt etc)

type AddressInput struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Zipcode string `json:"zipcode"`
}

type CustomerInput struct {
	FullName    string       `json:"full_name"`
	DateOfBirth string       `json:"date_of_birth"`
	IDNumber    string       `json:"id_number"`
	Email       string       `json:"email,omitempty"`
	Phone       string       `json:"phone"`
	Address     AddressInput `json:"address"`
}

type CollateralInput struct {
	Category           string `json:"category"`
	Brand              string `json:"brand"`
	Variant            string `json:"variant"`
	ManufacturingYear  int    `json:"manufacturing_year"`
	IsDocumentComplete bool   `json:"is_document_complete"`
}

type ProposedLoanInput struct {
	Tenure int     `json:"tenure"`
	Amount float64 `json:"amount"`
}

// LoanApplicationDraftInput is the combined input for creating a new draft.
type LoanApplicationDraftInput struct {
	ProposedLoan ProposedLoanInput `json:"proposed_loan"`
	Collateral   CollateralInput   `json:"collateral"`
	Customer     CustomerInput     `json:"customer"`
}
