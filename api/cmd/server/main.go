package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/sjtutwilight/Twilight/api/graph/generated"
	"github.com/sjtutwilight/Twilight/api/graph/resolver"
	"github.com/sjtutwilight/Twilight/api/internal/config"
	"github.com/sjtutwilight/Twilight/api/internal/middleware"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create new schema and resolver
	r := resolver.NewResolver(db)
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: r}))

	// Create a new router
	mux := http.NewServeMux()

	// Add CORS middleware
	corsHandler := middleware.CorsMiddleware()(srv)

	// Setup routes
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", corsHandler)

	// Get port from environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8091"
	}

	// Start server
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
