package graphqlhandler

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
)

// --- Validation Helpers ---
func isValidDate(dateStr string) bool {
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

func isValidEmail(emailStr string) bool {
	if emailStr == "" { // Email is optional
		return true
	}
	return regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(emailStr)
}

func validateAddressInput(input map[string]interface{}) error {
	street, _ := input["street"].(string)
	city, _ := input["city"].(string)
	zipcode, _ := input["zipcode"].(string)

	if len(street) == 0 || len(street) > 200 {
		return fmt.Errorf("street must be 1-200 characters")
	}
	if len(city) == 0 || len(city) > 100 {
		return fmt.Errorf("city must be 1-100 characters")
	}
	if len(zipcode) < 3 || len(zipcode) > 10 {
		return fmt.Errorf("zipcode must be 3-10 characters")
	}
	return nil
}

func validateCustomerInput(input map[string]interface{}) error {
	fullName, _ := input["full_name"].(string)
	dob, _ := input["date_of_birth"].(string)
	idNumber, _ := input["id_number"].(string)
	email, _ := input["email"].(string) // email can be empty string if not provided
	phone, _ := input["phone"].(string)

	if !regexp.MustCompile(`^[a-zA-Z ]{3,100}$`).MatchString(fullName) {
		return fmt.Errorf("full_name must be 3-100 characters, alphabet and space only")
	}
	if !isValidDate(dob) {
		return fmt.Errorf("date_of_birth must be in YYYY-MM-DD format")
	}
	if len(idNumber) == 0 || len(idNumber) > 25 {
		return fmt.Errorf("id_number must be 1-25 characters")
	}
	if email != "" && !isValidEmail(email) { // Validate only if email is provided
		return fmt.Errorf("email is not valid")
	}
	if !regexp.MustCompile(`^[0-9]{6,30}$`).MatchString(phone) {
		return fmt.Errorf("phone must be 6-30 digits")
	}

	addressInput, ok := input["address"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("address is required")
	}
	if err := validateAddressInput(addressInput); err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}
	return nil
}

func validateCollateralInput(input map[string]interface{}) error {
	brand, _ := input["brand"].(string)
	variant, _ := input["variant"].(string)
	mfgYear, okInt := input["manufacturing_year"].(int)

	if len(brand) == 0 {
		return fmt.Errorf("brand is required")
	}
	if len(variant) == 0 {
		return fmt.Errorf("variant is required")
	}
	currentYear := time.Now().Year()
	if !okInt || mfgYear < 2020 || mfgYear > currentYear {
		return fmt.Errorf("manufacturing_year must be between 2020 and %d", currentYear)
	}
	_, okBool := input["is_document_complete"].(bool)
	if !okBool {
		// This case should ideally be caught by GraphQL type system for non-null boolean
		return fmt.Errorf("is_document_complete is required and must be a boolean")
	}
	// Category is enum, handled by GraphQL type system
	return nil
}

func validateProposedLoanInput(input map[string]interface{}) error {
	tenure, okInt := input["tenure"].(int)
	amount, okFloat := input["amount"].(float64)

	if !okInt || tenure < 3 || tenure > 60 || tenure%3 != 0 {
		return fmt.Errorf("tenure must be between 3 and 60, and divisible by 3")
	}
	if !okFloat || amount < 100 || amount > 50000 {
		return fmt.Errorf("amount must be between 100 and 50000")
	}
	return nil
}

// --- Resolver Functions ---

var healthCheckResolver = func(p graphql.ResolveParams) (interface{}, error) {
	return "OK", nil
}

var createLoanApplicationDraftResolver = func(p graphql.ResolveParams) (interface{}, error) {
	dataArg, ok := p.Args["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing 'data' argument")
	}

	// Validate inputs
	proposedLoanInput, _ := dataArg["proposed_loan"].(map[string]interface{})
	collateralInput, _ := dataArg["collateral"].(map[string]interface{})
	customerInput, _ := dataArg["customer"].(map[string]interface{})

	if err := validateProposedLoanInput(proposedLoanInput); err != nil {
		return nil, fmt.Errorf("invalid proposed_loan: %w", err)
	}
	if err := validateCollateralInput(collateralInput); err != nil {
		return nil, fmt.Errorf("invalid collateral: %w", err)
	}
	if err := validateCustomerInput(customerInput); err != nil {
		return nil, fmt.Errorf("invalid customer: %w", err)
	}

	// Map input to data structure
	appUUID := uuid.New().String()
	now := time.Now()

	customerAddrInput := customerInput["address"].(map[string]interface{})
	newApp := &LoanApplicationData{
		UUID:   appUUID,
		Status: "DRAFT",
		ProposedLoan: ProposedLoanData{
			Tenure: proposedLoanInput["tenure"].(int),
			Amount: proposedLoanInput["amount"].(float64),
		},
		Collateral: CollateralData{
			Category:           collateralInput["category"].(string),
			Brand:              collateralInput["brand"].(string),
			Variant:            collateralInput["variant"].(string),
			ManufacturingYear:  collateralInput["manufacturing_year"].(int),
			IsDocumentComplete: collateralInput["is_document_complete"].(bool),
		},
		Customer: CustomerData{
			FullName:    customerInput["full_name"].(string),
			DateOfBirth: customerInput["date_of_birth"].(string),
			IDNumber:    customerInput["id_number"].(string),
			Email:       customerInput["email"].(string), // Already asserted as string or empty
			Phone:       customerInput["phone"].(string),
			Address: AddressData{
				Street:  customerAddrInput["street"].(string),
				City:    customerAddrInput["city"].(string),
				Zipcode: customerAddrInput["zipcode"].(string),
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	loanApplicationsMutex.Lock()
	loanApplications[appUUID] = newApp
	loanApplicationsMutex.Unlock()

	return appUUID, nil
}

var getLoanApplicationResolver = func(p graphql.ResolveParams) (interface{}, error) {
	uuidArg, ok := p.Args["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("missing 'uuid' argument")
	}

	loanApplicationsMutex.RLock()
	app, exists := loanApplications[uuidArg]
	loanApplicationsMutex.RUnlock()

	if !exists {
		return nil, nil // GraphQL spec: return null if not found for nullable type
	}
	// Map internal struct to the format expected by graphql.Field resolver
	// This mapping is implicitly handled if LoanApplicationData fields match LoanApplication type fields
	// and their Go types are compatible with what graphql-go expects (e.g. string for Date, Email)
	return app, nil
}

var submitLoanApplicationResolver = func(p graphql.ResolveParams) (interface{}, error) {
	uuidArg, ok := p.Args["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("missing 'uuid' argument")
	}

	loanApplicationsMutex.Lock()
	defer loanApplicationsMutex.Unlock()

	app, exists := loanApplications[uuidArg]
	if !exists {
		return false, fmt.Errorf("loan application with UUID '%s' not found", uuidArg)
	}

	if app.Status != "DRAFT" {
		// Depending on business logic, could allow submission from other statuses or return error
		return false, fmt.Errorf("loan application status is '%s', cannot submit", app.Status)
	}

	app.Status = "SUBMITTED"
	app.UpdatedAt = time.Now()
	loanApplications[uuidArg] = app // Re-assign pointer if needed, though map stores pointer

	return true, nil
}

var cancelLoanApplicationResolver = func(p graphql.ResolveParams) (interface{}, error) {
	uuidArg, ok := p.Args["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("missing 'uuid' argument")
	}

	loanApplicationsMutex.Lock()
	defer loanApplicationsMutex.Unlock()

	app, exists := loanApplications[uuidArg]
	if !exists {
		return false, fmt.Errorf("loan application with UUID '%s' not found", uuidArg)
	}

	// Add business logic here, e.g., cannot cancel if already processed
	if app.Status == "CANCELLED" {
		return true, nil // Already cancelled
	}
	if app.Status != "DRAFT" && app.Status != "SUBMITTED" { // Example: cannot cancel if in terminal state other than cancelled
		return false, fmt.Errorf("loan application status is '%s', cannot cancel", app.Status)
	}

	app.Status = "CANCELLED"
	app.UpdatedAt = time.Now()
	loanApplications[uuidArg] = app

	return true, nil
}

// Field resolver for LoanApplication.createdAt and LoanApplication.updatedAt to format time.Time
var timeFormatterResolver = func(p graphql.ResolveParams) (interface{}, error) {
	if t, ok := p.Source.(*LoanApplicationData); ok {
		if p.Info.FieldName == "created_at" {
			return t.CreatedAt.Format(time.RFC3339), nil
		}
		if p.Info.FieldName == "updated_at" {
			return t.UpdatedAt.Format(time.RFC3339), nil
		}
	}
	return nil, fmt.Errorf("failed to format time")
}

// Need to update loanApplicationType in types.go to use this resolver for createdAt and updatedAt
// This subtask cannot modify types.go, so this is a note for future adjustment or it will be done when schema is built.
// For now, the schema will be built with string types for these and expect the LoanApplicationData to provide them as strings.
// Let's adjust LoanApplicationData to store string for CreatedAt/UpdatedAt to simplify for now.
// (Correction: The LoanApplicationData struct already has time.Time, the resolver above is correct for formatting it.)
// The LoanApplication GraphQL type in types.go uses graphql.String for created_at/updated_at.
// So, the resolver for GetLoanApplication needs to ensure these are strings when it returns LoanApplicationData.
// The `app` returned by getLoanApplicationResolver is a *LoanApplicationData.
// graphql-go will look for fields like "CreatedAt" on this struct.
// If LoanApplicationData.CreatedAt is time.Time, and loanApplicationType.Fields["created_at"] is graphql.String,
// we need a resolver for the *field* "created_at" on the *type* "LoanApplication".

// This will be done in schema.go by assigning this resolver to the fields of loanApplicationType.
// For now, the type definition in types.go has these as graphql.String.
// The LoanApplicationData struct has time.Time.
// The default resolver will try to convert time.Time to string. The format might not be RFC3339.
// So, assigning timeFormatterResolver to these fields in schema.go is the correct approach.
