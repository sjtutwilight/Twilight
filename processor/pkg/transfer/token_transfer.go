package transfer

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/jmoiron/sqlx"
	"github.com/sjtutwilight/Twilight/common/pkg/types"
)

// handleTokenTransfer processes ERC20 token transfers
func (p *Processor) handleTokenTransfer(
	ctx context.Context,
	tx *sqlx.Tx,
	tokenAddr string,
	from string,
	to string,
	value *big.Int,
	event *types.Event,
) error {
	// Get token metadata
	tokenMeta := p.cacheManager.GetTokenMetadata(tokenAddr)
	if tokenMeta == nil {
		return fmt.Errorf("token metadata not found for address: %s", tokenAddr)
	}

	// Get token price
	tokenPrice, exists := p.cacheManager.GetTokenPrice(tokenAddr)
	if !exists {
		log.Printf("Warning: token price not found for address: %s, using 0", tokenAddr)
		tokenPrice = 0
	}

	// Calculate value in USD
	decimals := tokenMeta.GetDecimalsAsInt()

	// Convert value to decimal
	valueDecimal := new(big.Float).SetInt(value)
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))
	valueDecimal.Quo(valueDecimal, divisor)

	// Calculate USD value
	valueUSD := new(big.Float).Mul(valueDecimal, big.NewFloat(tokenPrice))

	// Get token ID as int64
	tokenID, err := p.cacheManager.GetTokenIDAsInt64(tokenAddr)
	if err != nil {
		return fmt.Errorf("failed to get token ID: %w", err)
	}

	// Insert transfer event record
	if err := p.insertTransferEvent(ctx, tx, event, tokenID, from, to, valueUSD, tokenMeta.Symbol); err != nil {
		return err
	}

	// Update account assets
	if err := p.updateAccountAssets(ctx, tx, tokenID, from, to, valueDecimal, valueUSD, "erc20", tokenMeta.Symbol); err != nil {
		return err
	}

	return nil
}
