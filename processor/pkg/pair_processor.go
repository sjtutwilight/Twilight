package db

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/jmoiron/sqlx"
	"github.com/twilight/common/pkg/types"
)

type PairProcessor struct {
	db *sqlx.DB
}

// Event data structures for validation
type SyncEvent struct {
	Reserve0 *big.Int `json:"reserve0"`
	Reserve1 *big.Int `json:"reserve1"`
}

type SwapEvent struct {
	Sender     string   `json:"sender"`
	Amount0In  *big.Int `json:"amount0In"`
	Amount1In  *big.Int `json:"amount1In"`
	Amount0Out *big.Int `json:"amount0Out"`
	Amount1Out *big.Int `json:"amount1Out"`
	To         string   `json:"to"`
}

type MintEvent struct {
	Sender    string   `json:"sender"`
	Amount0   *big.Int `json:"amount0"`
	Amount1   *big.Int `json:"amount1"`
	Liquidity *big.Int `json:"liquidity"`
}

type BurnEvent struct {
	Sender    string   `json:"sender"`
	Amount0   *big.Int `json:"amount0"`
	Amount1   *big.Int `json:"amount1"`
	Liquidity *big.Int `json:"liquidity"`
	To        string   `json:"to"`
}

func NewPairProcessor(db *sqlx.DB) *PairProcessor {
	return &PairProcessor{db: db}
}

// ProcessPairEvent validates pair events
func (p *PairProcessor) ProcessPairEvent(ctx context.Context, tx *sqlx.Tx, event *types.Event) error {
	// First validate the event is for a known pair
	var exists bool
	err := tx.Get(&exists, `
		SELECT EXISTS(
			SELECT 1 FROM twswap_pair 
			WHERE chain_id = $1 AND pair_address = $2
		)`,
		"local", event.ContractAddress)
	if err != nil {
		return fmt.Errorf("failed to check pair existence: %v", err)
	}
	if !exists {
		return fmt.Errorf("unknown pair address: %s", event.ContractAddress)
	}

	// Validate event data format based on event type
	switch event.EventName {
	case "Sync":
		var syncEvent SyncEvent
		if err := json.Unmarshal([]byte(event.EventData), &syncEvent); err != nil {
			return fmt.Errorf("invalid Sync event data: %v", err)
		}
		if syncEvent.Reserve0 == nil || syncEvent.Reserve1 == nil {
			return fmt.Errorf("Sync event missing reserves")
		}

	case "Swap":
		var swapEvent SwapEvent
		if err := json.Unmarshal([]byte(event.EventData), &swapEvent); err != nil {
			return fmt.Errorf("invalid Swap event data: %v", err)
		}
		if swapEvent.Amount0In == nil || swapEvent.Amount1In == nil ||
			swapEvent.Amount0Out == nil || swapEvent.Amount1Out == nil {
			return fmt.Errorf("Swap event missing amounts")
		}
		if swapEvent.Sender == "" || swapEvent.To == "" {
			return fmt.Errorf("Swap event missing addresses")
		}

	case "Mint":
		var mintEvent MintEvent
		if err := json.Unmarshal([]byte(event.EventData), &mintEvent); err != nil {
			return fmt.Errorf("invalid Mint event data: %v", err)
		}
		if mintEvent.Amount0 == nil || mintEvent.Amount1 == nil || mintEvent.Liquidity == nil {
			return fmt.Errorf("Mint event missing amounts")
		}
		if mintEvent.Sender == "" {
			return fmt.Errorf("Mint event missing sender")
		}

	case "Burn":
		var burnEvent BurnEvent
		if err := json.Unmarshal([]byte(event.EventData), &burnEvent); err != nil {
			return fmt.Errorf("invalid Burn event data: %v", err)
		}
		if burnEvent.Amount0 == nil || burnEvent.Amount1 == nil || burnEvent.Liquidity == nil {
			return fmt.Errorf("Burn event missing amounts")
		}
		if burnEvent.Sender == "" || burnEvent.To == "" {
			return fmt.Errorf("Burn event missing addresses")
		}

	default:
		return fmt.Errorf("unknown pair event: %s", event.EventName)
	}

	// The event is already saved in the events table by the main processor
	return nil
}
