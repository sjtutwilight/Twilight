package db

import (
	"context"
	"fmt"
	"math/big"

	"github.com/sjtutwilight/Twilight/common/pkg/types"

	"github.com/jmoiron/sqlx"
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

	// The event is already saved in the events table by the main processor
	return nil
}
