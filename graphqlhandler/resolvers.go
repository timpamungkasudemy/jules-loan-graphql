package graphqlhandler

import (
	"context" // Added for p.Context
	"fmt"
	"regexp"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/timpamungkas/loangraphql/db"    // Added for DBService
	"github.com/timpamungkas/loangraphql/model" // Added for model types
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

	proposedLoanMap, _ := dataArg["proposed_loan"].(map[string]interface{})
	collateralMap, _ := dataArg["collateral"].(map[string]interface{})
	customerMap, _ := dataArg["customer"].(map[string]interface{})

	if err := validateProposedLoanInput(proposedLoanMap); err != nil {
		return nil, fmt.Errorf("invalid proposed_loan: %w", err)
	}
	if err := validateCollateralInput(collateralMap); err != nil {
		return nil, fmt.Errorf("invalid collateral: %w", err)
	}
	if err := validateCustomerInput(customerMap); err != nil {
		return nil, fmt.Errorf("invalid customer: %w", err)
	}

	customerAddrMap, _ := customerMap["address"].(map[string]interface{})
	emailStr := ""
	if emailVal, ok := customerMap["email"]; ok {
		emailStr, _ = emailVal.(string)
	}

	customerIn := model.CustomerInput{
		FullName:    customerMap["full_name"].(string),
		DateOfBirth: customerMap["date_of_birth"].(string),
		IDNumber:    customerMap["id_number"].(string),
		Email:       emailStr,
		Phone:       customerMap["phone"].(string),
		Address: model.AddressInput{
			Street:  customerAddrMap["street"].(string),
			City:    customerAddrMap["city"].(string),
			Zipcode: customerAddrMap["zipcode"].(string),
		},
	}
	proposedLoanIn := model.ProposedLoanInput{
		Tenure: proposedLoanMap["tenure"].(int),
		Amount: proposedLoanMap["amount"].(float64),
	}
	collateralIn := model.CollateralInput{
		Category:           collateralMap["category"].(string),
		Brand:              collateralMap["brand"].(string),
		Variant:            collateralMap["variant"].(string),
		ManufacturingYear:  collateralMap["manufacturing_year"].(int),
		IsDocumentComplete: collateralMap["is_document_complete"].(bool),
	}

	createdLoanApp, err := r.DB.CreateLoanApplicationDraft(p.Context, customerIn, proposedLoanIn, collateralIn, "system_user_resolver")
	if err != nil {
		return nil, fmt.Errorf("failed to create loan application draft in DB: %w", err)
	}
	// The GraphQL schema for this mutation expects an ID (string) to be returned.
	return createdLoanApp.ID, nil
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
	if loanApp == nil {
		return nil, nil
	}
	// This returns *model.LoanApplication. The GraphQL type returned by GetLoanApplicationType()
	// in types.go needs to be compatible with this structure.
	return loanApp, nil
}

func (r *Resolver) SubmitLoanApplication(p graphql.ResolveParams) (interface{}, error) {
	uuidArg, ok := p.Args["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("missing 'uuid' argument")
	}

	success, err := r.DB.SubmitLoanApplication(p.Context, uuidArg, "system_user_resolver")
	if err != nil {
		return nil, fmt.Errorf("failed to submit loan application in DB: %w", err)
	}
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

	success, err := r.DB.CancelLoanApplication(p.Context, uuidArg, "system_user_resolver")
	if err != nil {
		return nil, fmt.Errorf("failed to cancel loan application in DB: %w", err)
	}
	if !success && err == nil {
	    return false, fmt.Errorf("loan application with UUID '%s' not found or already deleted", uuidArg)
	}
	return success, nil
}

var timeFormatterResolver = func(p graphql.ResolveParams) (interface{}, error) {
	// Check if the source is *model.LoanApplication
	if loanApp, ok := p.Source.(*model.LoanApplication); ok {
		if p.Info.FieldName == "createdAt" || p.Info.FieldName == "created_at" {
			return loanApp.CreatedAt.Format(time.RFC3339), nil
		}
		if p.Info.FieldName == "updatedAt" || p.Info.FieldName == "updated_at" {
			return loanApp.UpdatedAt.Format(time.RFC3339), nil
		}
		// If we need to format dates from the nested CustomerData
		if p.Info.FieldName == "customer" { // This would be for the whole customer object
			// If specific fields within customer need formatting, the GraphQL schema
			// would need resolvers on the Customer type's fields.
			// For example, if customer.created_at needed formatting.
			// This resolver is for fields on LoanApplication type.
		}
	} else if tTime, ok := p.Source.(time.Time); ok {
        return tTime.Format(time.RFC3339), nil
    }
    // Fallback for other fields on LoanApplication that might be time.Time but not handled above,
    // or if there's a type mismatch.
	// Or, if called on a field of CustomerData that got passed here somehow.
	// For CustomerData.CreatedAt/UpdatedAt, specific resolvers on CustomerType fields are better.
	// This specific resolver is attached to LoanApplicationType's fields.

	// Attempt to access nested customer fields if appropriate for the field name
	// This part is tricky because this resolver is attached to LoanApplication fields.
	// If p.Info.FieldName refers to a field *within* CustomerData (e.g. "customer.created_at" - not standard GQL path)
	// it won't work directly.
	// However, if schema design passes CustomerData as source for its own fields, then this might be relevant.
	// For now, this resolver should primarily handle LoanApplication.CreatedAt and LoanApplication.UpdatedAt.

	// If source is *model.Customer (e.g. if this resolver was mistakenly used for Customer fields)
	if cust, ok := p.Source.(*model.Customer); ok {
	    if p.Info.FieldName == "createdAt" || p.Info.FieldName == "created_at" {
			return cust.CreatedAt.Format(time.RFC3339), nil
		}
		if p.Info.FieldName == "updatedAt" || p.Info.FieldName == "updated_at" {
			return cust.UpdatedAt.Format(time.RFC3339), nil
		}
	}

	return nil, fmt.Errorf("timeFormatterResolver: unhandled source type %T or field %s", p.Source, p.Info.FieldName)
}
