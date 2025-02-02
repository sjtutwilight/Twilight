package resolver

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	db *pgxpool.Pool
}

// NewResolver creates a new resolver with the given database connection
func NewResolver(db *pgxpool.Pool) *Resolver {
	if db == nil {
		panic("database connection is required")
	}
	return &Resolver{
		db: db,
	}
}
