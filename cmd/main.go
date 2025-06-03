package main

import (
	"context" // Added for pgxpool and general context usage
	"log"
	"net/http"
	"os"
	// "fmt" // Not strictly needed if using log.Fatalf

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Driver for postgres
	_ "github.com/golang-migrate/migrate/v4/source/file"       // Driver for file source
	"github.com/graphql-go/handler"
	"github.com/jackc/pgx/v5/pgxpool" // Added for database connection pool

	"github.com/timpamungkas/loangraphql/db"             // Added for DBService
	"github.com/timpamungkas/loangraphql/graphqlhandler" // Import the local package
)

func main() {
	log.Println("Starting application...")
	ctx := context.Background() // Create a background context

	// --- Database Migration ---
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Fallback to a default URL if DATABASE_URL is not set.
		// IMPORTANT: Advise users to set DATABASE_URL in production.
		databaseURL = "postgres://user:password@localhost:5432/yourdb?sslmode=disable"
		log.Printf("WARNING: DATABASE_URL environment variable not set. Using default: %s", databaseURL)
		log.Println("WARNING: For production, please set the DATABASE_URL environment variable.")
	}

	log.Println("Attempting to run database migrations from file://db/migrations...")
	m, err := migrate.New(
		"file://db/migrations",
		databaseURL,
	)
	if err != nil {
		log.Fatalf("Failed to initialize migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("An error occurred while running database migrations: %v", err)
	} else if err == migrate.ErrNoChange {
		log.Println("No new database migrations to apply.")
	} else {
		log.Println("Database migrations applied successfully.")
	}
	// Source and database handles are closed automatically by migrate.New and m.Up()
	// unless m.Close() is explicitly called. For this simple case, it's fine.

	// --- Database Connection Pool ---
	log.Println("Initializing database connection pool...")
	dbConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("Failed to parse database URL for pgxpool: %v", err)
	}
	// Example: Configure pool settings (optional)
	// dbConfig.MaxConns = 5
	// dbConfig.MinConns = 1
	// dbConfig.MaxConnLifetime = time.Hour
	// dbConfig.MaxConnIdleTime = time.Minute * 30

	pool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		log.Fatalf("Unable to create database connection pool: %v", err)
	}
	defer pool.Close() // Ensure pool is closed when main function exits

	// Ping the database to verify the connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v. Please check connection and credentials.", err)
	}
	log.Println("Successfully connected to the database and connection pool initialized.")

	// --- Initialize Services and Resolver ---
	log.Println("Initializing services...")
	dbService := db.NewDBService(pool)
	appResolver := &graphqlhandler.Resolver{DB: dbService}
	log.Println("Services initialized.")

	// --- Build GraphQL Schema ---
	log.Println("Building GraphQL schema...")
	gqlSchema, err := graphqlhandler.BuildSchema(appResolver)
	if err != nil {
		log.Fatalf("Failed to build GraphQL schema: %v", err)
	}
	// Assign to the global variable in graphqlhandler package.
	// This provides backward compatibility if any part of graphqlhandler (e.g. types.go)
	// still implicitly relies on it, though ideally it shouldn't.
	graphqlhandler.Schema = gqlSchema
	log.Println("GraphQL schema built successfully.")

	// --- Setup GraphQL HTTP Handler ---
	log.Println("Setting up GraphQL HTTP handler...")
	graphqlGQLHandler := handler.New(&handler.Config{
		Schema:   &gqlSchema, // Use the dynamically built schema
		Pretty:   true,
		GraphiQL: true, // Enable GraphiQL interface
	})
	http.Handle("/graphql", graphqlGQLHandler)
	log.Println("GraphQL HTTP handler configured.")

	// --- Start HTTP Server ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	log.Printf("GraphQL server starting on http://localhost:%s/graphql", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
