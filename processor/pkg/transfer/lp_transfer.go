package transfer

import (
	"context"
	"fmt"
	"math/big"

	"github.com/jmoiron/sqlx"
	"github.com/sjtutwilight/Twilight/common/pkg/types"
)

// handleLPTransfer processes LP token transfers
func (p *Processor) handleLPTransfer(
	ctx context.Context,
	tx *sqlx.Tx,
	pairAddr string,
	from string,
	to string,
	value *big.Int,
	event *types.Event,
) error {
	// Get pair metadata
	pairMeta := p.cacheManager.GetPairMetadata(pairAddr)
	if pairMeta == nil {
		return fmt.Errorf("pair metadata not found for address: %s", pairAddr)
	}

	// Convert value to decimal (assuming 18 decimals for LP tokens)
	valueDecimal := new(big.Float).SetInt(value)
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	valueDecimal.Quo(valueDecimal, divisor)

	// For LP tokens, we don't have a direct price, so we'll use 0 for now
	// In a real implementation, we would calculate the LP token price based on reserves and totalSupply
	valueUSD := big.NewFloat(0)

	// Create a descriptive name for the LP token
	lpName := fmt.Sprintf("%s-%s LP", pairMeta.Token0.Symbol, pairMeta.Token1.Symbol)

	// Get pair ID as int64
	pairID, err := p.cacheManager.GetPairIDAsInt64(pairAddr)
	if err != nil {
		return fmt.Errorf("failed to get pair ID: %w", err)
	}

	// Update account assets
	if err := p.updateAccountAssets(ctx, tx, pairID, from, to, valueDecimal, valueUSD, "defiPosition", lpName); err != nil {
		return err
	}

	return nil
}
