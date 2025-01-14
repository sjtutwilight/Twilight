package db

import (
	"context"
	"fmt"

	"twilight/pkg/types"

	"github.com/jmoiron/sqlx"
)

// Constants
const (
	USDC_ADDRESS = "0x1234" // Replace with actual USDC address
)

// EventProcessor handles processing of different event types
type EventProcessor struct {
	db *sqlx.DB
}

// NewEventProcessor creates a new event processor
func NewEventProcessor(db *sqlx.DB) *EventProcessor {
	return &EventProcessor{db: db}
}

// ProcessEvent processes an event based on its type
func (p *EventProcessor) ProcessEvent(ctx context.Context, tx *sqlx.Tx, event *types.Event) error {
	// Insert event using prepared statement
	if err := p.insertEvent(ctx, tx, event); err != nil {
		return &ProcessorError{
			Op:  "insert_event",
			Err: err,
		}
	}

	// Process specific event type
	var err error
	switch event.EventName {
	case "PairCreated":
		err = p.processPairCreated(ctx, tx, event)
	case "Sync":
		err = p.processSync(ctx, tx, event)
	}

	if err != nil {
		return &ProcessorError{
			Op:  "process_" + event.EventName,
			Err: err,
		}
	}

	return nil
}

// insertEvent inserts an event record
func (p *EventProcessor) insertEvent(ctx context.Context, tx *sqlx.Tx, event *types.Event) error {
	_, err := tx.NamedExecContext(ctx, `
		INSERT INTO event (
			transaction_id, chain_id, event_name, contract_address,
			log_index, event_data, create_time, block_number
		) VALUES (
			:transaction_id, :chain_id, :event_name, :contract_address,
			:log_index, :event_data, :create_time, :block_number
		)`, event)
	return err
}

// processPairCreated handles PairCreated events
func (p *EventProcessor) processPairCreated(ctx context.Context, tx *sqlx.Tx, event *types.Event) error {
	// Decode event data
	decoded, err := types.DecodeEvent(event.EventData)
	if err != nil {
		return fmt.Errorf("failed to decode PairCreated event: %w", err)
	}

	// Extract token addresses
	token0Addr := decoded.Args["token0"].(string)
	token1Addr := decoded.Args["token1"].(string)
	pairAddr := decoded.Args["pair"].(string)

	// Get token IDs
	var token0ID, token1ID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM token WHERE token_address = $1
	`, token0Addr).Scan(&token0ID)
	if err != nil {
		return fmt.Errorf("failed to get token0 ID: %w", err)
	}

	err = tx.QueryRowContext(ctx, `
		SELECT id FROM token WHERE token_address = $1
	`, token1Addr).Scan(&token1ID)
	if err != nil {
		return fmt.Errorf("failed to get token1 ID: %w", err)
	}

	// Check if token0 is USDC
	var token0IsUSDC bool
	err = tx.QueryRowContext(ctx, `
		SELECT token_address = $1 FROM token WHERE id = $2
	`, USDC_ADDRESS, token0ID).Scan(&token0IsUSDC)
	if err != nil {
		return fmt.Errorf("failed to check if token0 is USDC: %w", err)
	}

	// Create new pair record
	_, err = tx.ExecContext(ctx, `
		INSERT INTO pair (
			chain_id, pair_address, token0_id, token1_id, token0_is_usdc,
			reserve0, reserve1, total_supply,
			volume_token0, volume_token1,
			created_at_timestamp, created_at_block_number,
			update_time
		) VALUES (
			$1, $2, $3, $4, $5,
			0, 0, 0,
			0, 0,
			NOW(), $6,
			NOW()
		)
	`, event.ChainID, pairAddr, token0ID, token1ID, token0IsUSDC, event.BlockNumber)

	if err != nil {
		return fmt.Errorf("failed to create pair record: %w", err)
	}

	return nil
}

// processSync handles Sync events to update reserves
func (p *EventProcessor) processSync(ctx context.Context, tx *sqlx.Tx, event *types.Event) error {
	// Decode event data
	decoded, err := types.DecodeEvent(event.EventData)
	if err != nil {
		return fmt.Errorf("failed to decode Sync event: %w", err)
	}

	// Extract reserves
	reserve0 := types.ToBigInt(decoded.Args["reserve0"])
	reserve1 := types.ToBigInt(decoded.Args["reserve1"])

	// Update pair reserves
	_, err = tx.ExecContext(ctx, `
		UPDATE pair 
		SET reserve0 = $1,
			reserve1 = $2,
			update_time = NOW()
		WHERE pair_address = $3
	`, reserve0.String(), reserve1.String(), event.ContractAddress)
	if err != nil {
		return fmt.Errorf("failed to update pair reserves: %w", err)
	}

	return nil
}
