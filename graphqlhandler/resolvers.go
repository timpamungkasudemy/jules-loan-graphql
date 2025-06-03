package graphqlhandler

import (
	"context" // Added for p.Context
	"fmt"
	"regexp"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/timpamungkas/loangraphql/db" // Added for DBService
	// "github.com/google/uuid" // No longer needed here, DB layer handles UUID generation
)

// Resolver struct holds dependencies for resolver methods, like a DB connection.
type Resolver struct {
	DB *db.DBService
}

// --- Validation Helpers (copied as is from original) ---
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
	email, _ := input["email"].(string)
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
	if email != "" && !isValidEmail(email) {
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
		return fmt.Errorf("is_document_complete is required and must be a boolean")
	}
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

// --- Resolver Methods ---

var healthCheckResolver = func(p graphql.ResolveParams) (interface{}, error) {
	return "OK", nil
}

func (r *Resolver) CreateLoanApplicationDraft(p graphql.ResolveParams) (interface{}, error) {
	dataArg, ok := p.Args["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing 'data' argument")
	}

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

	customerAddrMap, _ := customerInput["address"].(map[string]interface{})

	// Ensure email is correctly handled if missing from input (GraphQL might omit it if not provided by client)
	emailStr := ""
	if emailVal, ok := customerInput["email"]; ok {
		emailStr, _ = emailVal.(string)
	}

	customerData := CustomerData{
		FullName:    customerInput["full_name"].(string),
		DateOfBirth: customerInput["date_of_birth"].(string),
		IDNumber:    customerInput["id_number"].(string),
		Email:       emailStr,
		Phone:       customerInput["phone"].(string),
		Address: AddressData{
			Street:  customerAddrMap["street"].(string),
			City:    customerAddrMap["city"].(string),
			Zipcode: customerAddrMap["zipcode"].(string),
		},
	}
	proposedLoanData := ProposedLoanData{
		Tenure: proposedLoanInput["tenure"].(int),
		Amount: proposedLoanInput["amount"].(float64),
	}
	collateralData := CollateralData{
		Category:           collateralInput["category"].(string),
		Brand:              collateralInput["brand"].(string),
		Variant:            collateralInput["variant"].(string),
		ManufacturingYear:  collateralInput["manufacturing_year"].(int),
		IsDocumentComplete: collateralInput["is_document_complete"].(bool),
	}

	// Use p.Context for the context argument
	// Placeholder "system_user_resolver" for createdBy
	loanUUID, err := r.DB.CreateLoanApplicationDraft(p.Context, customerData, proposedLoanData, collateralData, "system_user_resolver")
	if err != nil {
		return nil, fmt.Errorf("failed to create loan application draft in DB: %w", err)
	}
	return loanUUID, nil
}

func (r *Resolver) GetLoanApplication(p graphql.ResolveParams) (interface{}, error) {
	uuidArg, ok := p.Args["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("missing 'uuid' argument")
	}

	loanApp, err := r.DB.GetLoanApplication(p.Context, uuidArg)
	if err != nil {
		return nil, fmt.Errorf("failed to get loan application from DB: %w", err)
	}
	if loanApp == nil { // DB method returns nil, nil for not found
		return nil, nil
	}
	return loanApp, nil
}

func (r *Resolver) SubmitLoanApplication(p graphql.ResolveParams) (interface{}, error) {
	uuidArg, ok := p.Args["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("missing 'uuid' argument")
	}

	// Placeholder "system_user_resolver" for updatedBy
	success, err := r.DB.SubmitLoanApplication(p.Context, uuidArg, "system_user_resolver")
	if err != nil {
		return nil, fmt.Errorf("failed to submit loan application in DB: %w", err)
	}
	// The DB method returns false, nil if not found or not in DRAFT state.
	// This needs to be translated to a GraphQL error or specific response.
	// For now, if not successful and no error, means condition not met (e.g., not found or wrong state).
	if !success && err == nil {
	    return false, fmt.Errorf("loan application with UUID '%s' not found, not in DRAFT state, or already deleted", uuidArg)
	}
	return success, nil
}

func (r *Resolver) CancelLoanApplication(p graphql.ResolveParams) (interface{}, error) {
	uuidArg, ok := p.Args["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("missing 'uuid' argument")
	}

	// Placeholder "system_user_resolver" for updatedBy
	success, err := r.DB.CancelLoanApplication(p.Context, uuidArg, "system_user_resolver")
	if err != nil {
		return nil, fmt.Errorf("failed to cancel loan application in DB: %w", err)
	}
	// Similar to submit, if not successful and no error, means condition not met.
	if !success && err == nil {
	    return false, fmt.Errorf("loan application with UUID '%s' not found or already deleted", uuidArg)
	}
	return success, nil
}

var timeFormatterResolver = func(p graphql.ResolveParams) (interface{}, error) {
	if t, ok := p.Source.(*LoanApplicationData); ok {
		// Ensure field names match the GraphQL schema
		if p.Info.FieldName == "createdAt" || p.Info.FieldName == "created_at" {
			return t.CreatedAt.Format(time.RFC3339), nil
		}
		if p.Info.FieldName == "updatedAt" || p.Info.FieldName == "updated_at" {
			return t.UpdatedAt.Format(time.RFC3339), nil
		}
	} else if tTime, ok := p.Source.(time.Time); ok { // If source is already time.Time
        return tTime.Format(time.RFC3339), nil
    }
    
    // Fallback or error if type assertion fails or field name doesn't match
    // Check the type of p.Source if the above assertions fail
    // log.Printf("timeFormatterResolver: Unhandled type for p.Source: %T for field %s", p.Source, p.Info.FieldName)
	return nil, fmt.Errorf("failed to format time, source type: %T for field %s", p.Source, p.Info.FieldName)
}
