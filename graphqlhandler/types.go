package graphqlhandler

import (
	"github.com/graphql-go/graphql"
)

// Enum for CollateralCategory
var collateralCategoryEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "CollateralCategory",
	Values: graphql.EnumValueConfigMap{
		"CAR": &graphql.EnumValueConfig{
			Value: "CAR",
		},
		"MOTORCYCLE": &graphql.EnumValueConfig{
			Value: "MOTORCYCLE",
		},
	},
})

// Custom Scalars (as String for now, validation in resolvers or later custom scalar type)
var dateScalar = graphql.String  // Placeholder for Date scalar
var emailScalar = graphql.String // Placeholder for Email scalar

// Address Type
var addressType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Address",
	Fields: graphql.Fields{
		"street":  &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"city":    &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"zipcode": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
	},
})

// Address Input Type
var addressInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AddressInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"street":  &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		"city":    &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		"zipcode": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
	},
})

// Customer Type
var customerType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Customer",
	Fields: graphql.Fields{
		"full_name":     &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"date_of_birth": &graphql.Field{Type: graphql.NewNonNull(dateScalar)},
		"id_number":     &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"email":         &graphql.Field{Type: emailScalar},
		"phone":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"address":       &graphql.Field{Type: graphql.NewNonNull(addressType)},
	},
})

// Customer Input Type
var customerInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CustomerInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"full_name":     &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		"date_of_birth": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(dateScalar)},
		"id_number":     &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		"email":         &graphql.InputObjectFieldConfig{Type: emailScalar},
		"phone":         &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		"address":       &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(addressInputType)},
	},
})

// Collateral Type
var collateralType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Collateral",
	Fields: graphql.Fields{
		"category":             &graphql.Field{Type: graphql.NewNonNull(collateralCategoryEnum)},
		"brand":                &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"variant":              &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"manufacturing_year":   &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
		"is_document_complete": &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
	},
})

// Collateral Input Type
var collateralInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CollateralInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"category":             &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(collateralCategoryEnum)},
		"brand":                &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		"variant":              &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		"manufacturing_year":   &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Int)},
		"is_document_complete": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Boolean)},
	},
})

// Proposed Loan Type
var proposedLoanType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ProposedLoan",
	Fields: graphql.Fields{
		"tenure": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
		"amount": &graphql.Field{Type: graphql.NewNonNull(graphql.Float)},
	},
})

// Proposed Loan Input Type
var proposedLoanInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "ProposedLoanInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"tenure": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Int)},
		"amount": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Float)},
	},
})

// Loan Application Type
var loanApplicationType *graphql.Object // Forward declaration for potential self-reference or ordering

func init() { // Use init to resolve potential circular dependencies if any type refers to LoanApplication itself.
	loanApplicationType = graphql.NewObject(graphql.ObjectConfig{
		Name: "LoanApplication",
		Fields: graphql.Fields{
			"uuid":          &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
			"status":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"proposed_loan": &graphql.Field{Type: graphql.NewNonNull(proposedLoanType)},
			"collateral":    &graphql.Field{Type: graphql.NewNonNull(collateralType)},
			"customer":      &graphql.Field{Type: graphql.NewNonNull(customerType)},
			"created_at":    &graphql.Field{Type: graphql.NewNonNull(graphql.String)}, // Using String for simplicity
			"updated_at":    &graphql.Field{Type: graphql.NewNonNull(graphql.String)}, // Using String for simplicity
		},
	})
}

// Loan Application Draft Input Type (for create mutation)
var loanApplicationDraftInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "LoanApplicationDraftInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"proposed_loan": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(proposedLoanInputType)},
		"collateral":    &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(collateralInputType)},
		"customer":      &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(customerInputType)},
	},
})

func GetLoanApplicationType() *graphql.Object {
	// This function is needed because loanApplicationType is initialized in init()
	// and might not be directly addressable if other types need it during their own global var initialization.
	// However, with all types in this file, direct use of loanApplicationType should be fine after init() runs.
	// For safety and to ensure it's initialized, especially if types were split into more files:
	if loanApplicationType == nil {
		// This case should ideally not happen if init() functions are correctly managed by Go runtime.
		// Re-run init logic or panic if critical. For now, let's assume init() handles it.
		// The init() function for loanApplicationType should handle its setup.
	}
	return loanApplicationType
}
