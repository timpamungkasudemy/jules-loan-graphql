package graphqlhandler

import (
	"sync"
	"time"
)

// Using maps for simple in-memory storage
// In a real app, use a database

var (
	loanApplications      = make(map[string]*LoanApplicationData)
	loanApplicationsMutex = &sync.RWMutex{}
)

// Internal data structures for storage (matching GraphQL types but as Go structs)
// These are separate from the graphql.Object definitions but will hold the data.

type AddressData struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Zipcode string `json:"zipcode"`
}

type CustomerData struct {
	FullName    string      `json:"full_name"`
	DateOfBirth string      `json:"date_of_birth"` // Store as string, validate in resolver
	IDNumber    string      `json:"id_number"`
	Email       string      `json:"email,omitempty"` // Store as string, validate in resolver
	Phone       string      `json:"phone"`
	Address     AddressData `json:"address"`
}

type CollateralData struct {
	Category           string `json:"category"` // CAR, MOTORCYCLE
	Brand              string `json:"brand"`
	Variant            string `json:"variant"`
	ManufacturingYear  int    `json:"manufacturing_year"`
	IsDocumentComplete bool   `json:"is_document_complete"`
}

type ProposedLoanData struct {
	Tenure int     `json:"tenure"`
	Amount float64 `json:"amount"`
}

type LoanApplicationData struct {
	UUID         string           `json:"uuid"`
	Status       string           `json:"status"` // DRAFT, SUBMITTED, CANCELLED
	ProposedLoan ProposedLoanData `json:"proposed_loan"`
	Collateral   CollateralData   `json:"collateral"`
	Customer     CustomerData     `json:"customer"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}
