package graphqlhandler

import (
	"fmt"
	"github.com/graphql-go/graphql"
	// Resolver is in the same package (resolvers.go), no need for explicit import of its definition
)

// Schema will be initialized by calling BuildSchema from main.go
var Schema graphql.Schema

// BuildSchema creates and returns the GraphQL schema, configured with the provided resolver.
func BuildSchema(resolver *Resolver) (graphql.Schema, error) {
	// It's good practice to ensure the resolver and its dependencies are valid.
	if resolver == nil {
		return graphql.Schema{}, fmt.Errorf("BuildSchema: provided resolver is nil")
	}
	if resolver.DB == nil {
		// Depending on whether all resolvers need DB, this check might be too strict
		// or could be done per-resolver method if some don't need DB.
		// For this application, all primary resolvers (CRUD operations) will need DB.
		return graphql.Schema{}, fmt.Errorf("BuildSchema: resolver.DB is nil")
	}

	// Note: The GetLoanApplicationType() function (from types.go) returns the GraphQL type for loan applications.
	// If its fields (like createdAt, updatedAt) need specific resolvers (e.g., timeFormatterResolver),
	// those should ideally be set within GetLoanApplicationType() itself or when its fields are defined in types.go.
	// This BuildSchema function will primarily wire up the main query and mutation resolvers.

	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"healthCheck": &graphql.Field{
				Type:    graphql.NewNonNull(graphql.String),
				Resolve: healthCheckResolver, // Remains a package-level function
			},
			"getLoanApplication": &graphql.Field{
				Type: GetLoanApplicationType(), // Assumes GetLoanApplicationType() is defined in types.go
				Args: graphql.FieldConfigArgument{
					"uuid": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: resolver.GetLoanApplication, // Use method from resolver instance
			},
		},
	})

	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createLoanApplicationDraft": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID), // Returns the new loan application UUID
				Args: graphql.FieldConfigArgument{
					"data": &graphql.ArgumentConfig{
						// Assumes loanApplicationDraftInputType is defined in types.go
						Type: graphql.NewNonNull(loanApplicationDraftInputType),
					},
				},
				Resolve: resolver.CreateLoanApplicationDraft, // Use method from resolver instance
			},
			"submitLoanApplication": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
				Args: graphql.FieldConfigArgument{
					"uuid": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: resolver.SubmitLoanApplication, // Use method from resolver instance
			},
			"cancelLoanApplication": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
				Args: graphql.FieldConfigArgument{
					"uuid": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: resolver.CancelLoanApplication, // Use method from resolver instance
			},
		},
	})

	// Create the new schema
	newSchema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
		// Types: []graphql.Type{...} // List any types here if they are not discoverable through query/mutation fields
	})

	if err != nil {
		return graphql.Schema{}, fmt.Errorf("failed to create GraphQL schema: %w", err)
	}

	return newSchema, nil
}
