package main

import (
	"log"
	"net/http"

	"github.com/graphql-go/handler"
	"github.com/timpamungkas/loangraphql/graphqlhandler" // Import the local package
)

func main() {
	// Initialize the GraphQL schema.
	// The schema is defined in graphqlhandler/schema.go and loaded in its init() function.
	// We just need to make sure the package is imported so its init() runs.
	// graphqlhandler.Schema will be available after this.

	graphqlGQLHandler := handler.New(&handler.Config{
		Schema:   &graphqlhandler.Schema,
		Pretty:   true, // For pretty JSON output
		GraphiQL: true, // Enable GraphiQL interface (optional)
	})

	// Register the GraphQL handler
	http.Handle("/graphql", graphqlGQLHandler)

	port := "8080"
	log.Printf("GraphQL server starting on http://localhost:%s/graphql", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
