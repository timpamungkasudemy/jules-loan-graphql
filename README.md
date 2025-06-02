# Loan GraphQL API Server (Go)

This project implements a GraphQL server for managing loan applications, built using Go and the `graphql-go/graphql` library.

## Prerequisites

- Go (version 1.21 or higher recommended, as used in `go.mod`)

## Setup & Running the Application

1.  **Clone the repository (if you haven't already):**
    ```bash
    git clone <repository-url>
    cd <repository-directory>
    ```

2.  **Download dependencies:**
    Open your terminal in the project root directory and run:
    ```bash
    go mod tidy
    ```
    This command ensures all necessary Go modules are downloaded and your `go.mod` and `go.sum` files are up to date.

3.  **Start the server:**
    From the project root directory, run:
    ```bash
    go run cmd/main.go
    ```
    The server will start, and by default, it will be accessible at `http://localhost:8080/graphql`.

    You should see a log message like:
    ```
    YYYY/MM/DD HH:MM:SS GraphQL server starting on http://localhost:8080/graphql
    ```

    You can then access this URL in your browser (if GraphiQL is enabled, which it is by default in this setup) or send GraphQL requests to it using a client like Postman, Insomnia, or `curl`.

## GraphQL Schema

Unlike projects using libraries like `gqlgen`, this server does **not** use `.graphqls` or `.gql` schema definition files that are then used to generate Go code. Instead, the GraphQL schema (types, queries, mutations) is defined directly in Go code.

**Key files for schema definition:**

-   **`graphqlhandler/types.go`**: Defines the GraphQL object types (e.g., `LoanApplication`, `Customer`), input types (e.g., `CustomerInput`), and enums (e.g., `CollateralCategory`) using the `graphql-go/graphql` library.
-   **`graphqlhandler/resolvers.go`**: Contains the resolver functions that provide the logic for fetching and manipulating data for your GraphQL queries and mutations. It also includes input validation logic.
-   **`graphqlhandler/schema.go`**: Constructs the overall GraphQL schema by assembling the query and mutation objects from their respective resolver functions and type definitions. The main `graphql.Schema` object is initialized here.
-   **`graphqlhandler/data.go`**: For this example, it holds the in-memory data storage. In a real application, resolvers would interact with a database or other persistent storage.

**Modifying the Schema:**

If you need to add or change the GraphQL schema (e.g., add a new field, type, query, or mutation):

1.  **Update Go Type Definitions:** Modify the relevant `graphql.Object` or `graphql.InputObject` definitions in `graphqlhandler/types.go`.
2.  **Update/Add Resolver Functions:** If you add new fields or operations, implement or update the corresponding resolver functions in `graphqlhandler/resolvers.go`. Ensure any new input types or fields have appropriate validation.
3.  **Update Schema Construction:** Adjust the schema definition in `graphqlhandler/schema.go` to include any new query fields, mutation fields, or types.
4.  **Update Data Structures (if applicable):** If your changes affect the underlying data model, update the Go structs in `graphqlhandler/data.go` (or your database interaction layer).

**Code Generation:**

There is **no separate script or command to generate Go source files from a GraphQL schema file** in this project setup because the schema is defined directly in Go. All changes are made manually to the Go source files mentioned above.

## Testing

You can test the GraphQL API by sending requests to `http://localhost:8080/graphql`. The GraphiQL interface (accessible in a browser) is enabled by default, allowing you to write and execute queries and mutations interactively.

**Example Health Check Query:**
```graphql
query {
  healthCheck
}
```

**Example Create Loan Application Draft Mutation:**
(Refer to the schema for the exact structure of `LoanApplicationDraftInput`)
```graphql
mutation CreateDraft {
  createLoanApplicationDraft(data: {
    proposed_loan: {
      tenure: 12,
      amount: 5000.00
    },
    collateral: {
      category: CAR,
      brand: "Toyota",
      variant: "Camry",
      manufacturing_year: 2021,
      is_document_complete: true
    },
    customer: {
      full_name: "John Doe",
      date_of_birth: "1990-01-15",
      id_number: "ID123456789",
      email: "john.doe@example.com",
      phone: "1234567890",
      address: {
        street: "123 Main St",
        city: "Anytown",
        zipcode: "12345"
      }
    }
  })
}
```
This mutation will return the UUID of the newly created draft. You can then use this UUID with the `getLoanApplication` query.
