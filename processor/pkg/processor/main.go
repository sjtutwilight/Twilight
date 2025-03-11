package processor

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL驱动
	"github.com/sjtutwilight/Twilight/processor/pkg/cache"
	"github.com/sjtutwilight/Twilight/processor/pkg/model"
	"github.com/sjtutwilight/Twilight/processor/pkg/transfer"
)

// Processor handles database operations with improved connection handling
type Processor struct {
	db *sqlx.DB
	// Prepared statements
	insertTxStmt    *sqlx.NamedStmt
	insertEventStmt *sqlx.NamedStmt
	// Processors
	cacheManager      *cache.Manager
	transferProcessor *transfer.Processor
}

// NewProcessor creates a new database processor with prepared statements
func NewProcessor(connStr string, redisAddr string) (*Processor, error) {
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, model.NewProcessorError("connect_db", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Initialize cache manager
	cacheManager, err := cache.NewManager(redisAddr)
	if err != nil {
		db.Close()
		return nil, model.NewProcessorError("create_cache_manager", err)
	}

	// Load all caches
	ctx := context.Background()
	if err := cacheManager.LoadAllCaches(ctx); err != nil {
		cacheManager.Close()
		db.Close()
		return nil, model.NewProcessorError("load_caches", err)
	}

	p := &Processor{
		db:           db,
		cacheManager: cacheManager,
	}

	// Initialize transfer processor
	p.transferProcessor = transfer.NewProcessor(db, cacheManager)

	if err := p.prepareStatements(); err != nil {
		cacheManager.Close()
		db.Close()
		return nil, err
	}

	return p, nil
}

// Close closes the database connection and other resources
func (p *Processor) Close() error {
	// Close prepared statements
	if p.insertTxStmt != nil {
		if err := p.insertTxStmt.Close(); err != nil {
			return err
		}
	}

	if p.insertEventStmt != nil {
		if err := p.insertEventStmt.Close(); err != nil {
			return err
		}
	}

	// Close cache manager
	if p.cacheManager != nil {
		if err := p.cacheManager.Close(); err != nil {
			return err
		}
	}

	// Close database connection
	if p.db != nil {
		return p.db.Close()
	}

	return nil
}
