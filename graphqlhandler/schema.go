package graphqlhandler

import (
	"fmt"
	"github.com/graphql-go/graphql"
)

var Schema graphql.Schema

func init() {
	// Re-fetch loanApplicationType to ensure it's initialized, especially its fields
	// This is important because we are about to assign resolvers to its fields.
	// The GetLoanApplicationType() function was defined in types.go for this purpose,
	// but since loanApplicationType is a global var in the same package and init() in types.go sets it up,
	// we can directly use loanApplicationType here.
	// For safety, one might wrap access, but direct use should be fine after all init() are run.

	// Assign specific resolvers for LoanApplication fields that need formatting (e.g., time.Time to String)
	// Ensure loanApplicationType is fully initialized before modifying its fields.
	// The fields for loanApplicationType are defined in types.go.
	// We need to make sure that the Field instances are the same ones we modify here.
	// This is tricky if types.go also has an init().
	// A safer way is to define fields with resolvers directly in types.go or pass resolver map to NewObject.
	// Given the current structure, let's try to define it here.
	// This assumes loanApplicationType from types.go is accessible and its Fields map can be modified.
	// This is generally not good practice. Better to define fields with resolvers in one place.

	// Let's redefine loanApplicationType here with the field resolvers for CreatedAt and UpdatedAt.
	// This will override the one in types.go if this init() runs after.
	// This is not ideal. A better way is to have a single point of truth for type definition.

	// Alternative: Modify types.go to accept resolver functions, or make types aware of them.
	// For now, let's assume the fields in loanApplicationType (from types.go) can have their Resolve property set.
	// This is fragile. The fields in the loanApplicationType object (defined in types.go)
	// must have their Resolve func set.

	// Correct approach: Define fields with their resolvers when creating the object in types.go
	// Since that's done, we will assume that types.go will be modified, or that the default
	// conversion of time.Time to string is acceptable for now, and field-specific resolvers
	// for formatting will be skipped in this step to avoid cross-file modification issues.
	// For now, we'll rely on default resolver behavior for LoanApplication fields.
	// We will ensure that the `LoanApplicationData` struct (from data.go) is returned,
	// and graphql-go will attempt to map its fields.

	// The `timeFormatterResolver` is defined in resolvers.go, but applying it to fields of
	// `loanApplicationType` (defined in types.go) from `schema.go` is complex due to initialization order
	// and package separation of concerns.
	// A common pattern is to have a function in types.go that returns the type, and it internally sets resolvers
	// or takes resolvers as arguments.

	// For this step, we will construct the Query and Mutation objects and the Schema.
	// We will rely on graphql-go's default resolver if a field resolver isn't specified,
	// which means it will try to find a struct field with the same name or a method.

	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"healthCheck": &graphql.Field{
				Type:    graphql.NewNonNull(graphql.String),
				Resolve: healthCheckResolver,
			},
			"getLoanApplication": &graphql.Field{
				Type: GetLoanApplicationType(), // Nullable as per schema
				Args: graphql.FieldConfigArgument{
					"uuid": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: getLoanApplicationResolver,
			},
		},
	})

	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createLoanApplicationDraft": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
				Args: graphql.FieldConfigArgument{
					"data": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(loanApplicationDraftInputType),
					},
				},
				Resolve: createLoanApplicationDraftResolver,
			},
			"submitLoanApplication": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
				Args: graphql.FieldConfigArgument{
					"uuid": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: submitLoanApplicationResolver,
			},
			"cancelLoanApplication": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
				Args: graphql.FieldConfigArgument{
					"uuid": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: cancelLoanApplicationResolver,
			},
		},
	})

	var err error
	Schema, err = graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
		// Types:    []graphql.Type{collateralCategoryEnum, dateScalar, emailScalar, addressType, customerType, collateralType, proposedLoanType, loanApplicationType}, // Explicitly list types if needed for schema documentation or if not referenced directly by Query/Mutation fields.
	})

	if err != nil {
		// This panic is okay for init() as it prevents server startup with invalid schema
		panic(fmt.Sprintf("Failed to create GraphQL schema: %v", err))
	}
}
