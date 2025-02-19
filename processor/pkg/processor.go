package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sjtutwilight/Twilight/common/pkg/types"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Custom error types
type ProcessorError struct {
	Op  string // Operation that failed
	Err error  // Original error
}

func (e *ProcessorError) Error() string {
	if e.Err == nil {
		return e.Op
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Processor handles database operations with improved connection handling
type Processor struct {
	db *sqlx.DB
	// Prepared statements
	insertTxStmt    *sqlx.NamedStmt
	insertEventStmt *sqlx.NamedStmt
	// Processors
	accountProcessor *AccountProcessor
	pairProcessor    *PairProcessor
}

// NewProcessor creates a new database processor with prepared statements
func NewProcessor(connStr string) (*Processor, error) {
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, &ProcessorError{Op: "connect_db", Err: err}
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	p := &Processor{
		db:               db,
		accountProcessor: NewAccountProcessor(db),
		pairProcessor:    NewPairProcessor(db),
	}

	if err := p.prepareStatements(); err != nil {
		db.Close()
		return nil, err
	}

	// Initialize processors
	ctx := context.Background()
	if err := p.accountProcessor.Initialize(ctx); err != nil {
		db.Close()
		return nil, &ProcessorError{Op: "initialize_account_processor", Err: err}
	}

	return p, nil
}

// prepareStatements prepares SQL statements for reuse
func (p *Processor) prepareStatements() error {
	var err error

	// Prepare transaction insert statement
	p.insertTxStmt, err = p.db.PrepareNamed(`
		INSERT INTO transaction (
			chain_id, transaction_hash, block_number, block_timestamp, 
			from_address, to_address, method_name, transaction_status, 
			gas_used, input_data, create_time
		) VALUES (
			:chain_id, :transaction_hash, :block_number, :block_timestamp,
			:from_address, :to_address, :method_name, :transaction_status,
			:gas_used, :input_data, :create_time
		) RETURNING id`)
	if err != nil {
		return &ProcessorError{Op: "prepare_tx_stmt", Err: err}
	}

	// Prepare event insert statement
	p.insertEventStmt, err = p.db.PrepareNamed(`
		INSERT INTO event (
			transaction_id, chain_id, event_name, contract_address,
			log_index, event_data, create_time, block_number
		) VALUES (
			:transaction_id, :chain_id, :event_name, :contract_address,
			:log_index, :event_data, :create_time, :block_number
		)`)
	if err != nil {
		return &ProcessorError{Op: "prepare_event_stmt", Err: err}
	}

	return nil
}

// ProcessTransaction processes a transaction and its events with improved error handling
func (p *Processor) ProcessTransaction(ctx context.Context, tx *types.Transaction, events []types.Event) error {
	// Start a database transaction
	dbTx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return &ProcessorError{Op: "begin_transaction", Err: err}
	}
	defer dbTx.Rollback()

	// Insert transaction using prepared statement
	var txID int64
	rows, err := p.insertTxStmt.QueryContext(ctx, tx)
	if err != nil {
		return &ProcessorError{Op: "insert_transaction", Err: err}
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&txID); err != nil {
			return &ProcessorError{Op: "scan_transaction_id", Err: err}
		}
	}

	// Process events
	for _, event := range events {
		event.TransactionID = txID
		event.CreateTime = time.Now()

		accountId := p.accountProcessor.accountMap[strings.ToLower(tx.FromAddress)]
		// Insert event record
		if err := p.insertEventRecord(ctx, dbTx, &event); err != nil {
			return &ProcessorError{Op: "insert_event", Err: err}
		}

		// // Process pair-related events
		// if err := p.pairProcessor.ProcessPairEvent(ctx, dbTx, &event); err != nil {
		// 	return &ProcessorError{Op: "process_pair_event", Err: err}
		// }

		// Process account-related events using decodedArgs directly
		if err := p.accountProcessor.ProcessEvent(ctx, dbTx, map[string]interface{}{
			"eventName":       event.EventName,
			"contractAddress": event.ContractAddress,
			"decodedArgs":     event.DecodedArgs,
		}, accountId); err != nil {
			return &ProcessorError{Op: "process_account_event", Err: err}
		}
	}

	// Commit transaction
	if err := dbTx.Commit(); err != nil {
		return &ProcessorError{Op: "commit_transaction", Err: err}
	}

	return nil
}

func (p *Processor) insertEventRecord(ctx context.Context, tx *sqlx.Tx, event *types.Event) error {
	_, err := p.insertEventStmt.Exec(event)
	if err != nil {
		return fmt.Errorf("failed to insert event: %v", err)
	}
	return nil
}

// Close closes the database connection and prepared statements
func (p *Processor) Close() error {
	if p.insertTxStmt != nil {
		p.insertTxStmt.Close()
	}
	if p.insertEventStmt != nil {
		p.insertEventStmt.Close()
	}
	return p.db.Close()
}
