package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/twilight/api/graph/generated"
	"github.com/twilight/api/graph/resolver"
	"github.com/twilight/api/internal/db"
)

const defaultPort = "8091"

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Initialize database connection
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	// Get database connection
	dbConn := db.GetDB()
	if dbConn == nil {
		log.Fatal("Failed to get database connection")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Create resolver with database connection
	resolver := resolver.NewResolver(dbConn)

	// Create GraphQL server
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: resolver,
	}))

	// Create a new CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		Debug:            true,
	})

	// Setup HTTP handlers with CORS
	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", srv)

	// Wrap the mux with CORS middleware
	handler := c.Handler(mux)

	// Start server
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
