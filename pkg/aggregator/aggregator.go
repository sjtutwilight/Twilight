package aggregator

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"twilight/pkg/types"

	"github.com/jmoiron/sqlx"
)

// Constants
const (
	USDC_ADDRESS = "0x1234" // Replace with actual USDC address
)

// Aggregator handles metric aggregation
type Aggregator struct {
	db      *sqlx.DB
	windows []MetricWindow
}

// NewAggregator creates a new aggregator
func NewAggregator(db *sqlx.DB) *Aggregator {
	return &Aggregator{
		db: db,
		windows: []MetricWindow{
			{
				Size:  30 * time.Second,
				Slide: 6 * time.Second,
				Type:  MetricType30s,
			},
			{
				Size:  5 * time.Minute,
				Slide: 1 * time.Minute,
				Type:  MetricType5m,
			},
			{
				Size:  30 * time.Minute,
				Slide: 6 * time.Minute,
				Type:  MetricType30m,
			},
		},
	}
}

// ProcessEvent processes an event and updates metrics
func (a *Aggregator) ProcessEvent(ctx context.Context, msg EventMessage) error {
	// Decode event data if not already decoded
	if msg.DecodedData.Name == "" {
		decoded, err := types.DecodeEvent(msg.Event.EventData)
		if err != nil {
			return fmt.Errorf("failed to decode event data: %w", err)
		}
		msg.DecodedData = *decoded
	}

	// Process event based on type
	switch msg.Event.EventName {
	case "Swap":
		return a.processSwap(ctx, msg)
	case "Mint":
		return a.processMint(ctx, msg)
	case "Burn":
		return a.processBurn(ctx, msg)
	case "Sync":
		return a.processSync(ctx, msg)
	}

	return nil
}

// processSwap handles Swap events
func (a *Aggregator) processSwap(ctx context.Context, msg EventMessage) error {
	// Get pair info
	var pair types.Pair
	err := a.db.GetContext(ctx, &pair, `
		SELECT * FROM pair WHERE pair_address = $1
	`, msg.Event.ContractAddress)
	if err != nil {
		return fmt.Errorf("failed to get pair: %w", err)
	}

	// Extract amounts
	amount0In := types.ToBigInt(msg.DecodedData.Args["amount0In"])
	amount1In := types.ToBigInt(msg.DecodedData.Args["amount1In"])
	amount0Out := types.ToBigInt(msg.DecodedData.Args["amount0Out"])
	amount1Out := types.ToBigInt(msg.DecodedData.Args["amount1Out"])

	// Calculate volumes
	volume0 := new(big.Int).Add(amount0In, amount0Out)
	volume1 := new(big.Int).Add(amount1In, amount1Out)

	// Calculate USD values
	volumeUSD := a.calculateUSDVolume(pair, volume0, volume1)

	// Update metrics for each window
	for _, window := range a.windows {
		windowStart := msg.BlockTimestamp.Truncate(window.Size)

		// Update pair metrics
		_, err = a.db.ExecContext(ctx, `
			INSERT INTO pair_metric (
				pair_id, metric_type, start_time,
				token0_address, token1_address,
				token0_volume, token1_volume, volume_usd,
				transaction_count, update_time
			) VALUES (
				$1, $2, $3,
				$4, $5,
				$6, $7, $8,
				1, NOW()
			)
			ON CONFLICT (pair_id, metric_type, start_time)
			DO UPDATE SET
				token0_volume = pair_metric.token0_volume + EXCLUDED.token0_volume,
				token1_volume = pair_metric.token1_volume + EXCLUDED.token1_volume,
				volume_usd = pair_metric.volume_usd + EXCLUDED.volume_usd,
				transaction_count = pair_metric.transaction_count + 1,
				update_time = NOW()
		`, pair.ID, window.Type, windowStart,
			pair.Token0ID, pair.Token1ID,
			volume0.String(), volume1.String(), volumeUSD)
		if err != nil {
			return fmt.Errorf("failed to update pair metrics: %w", err)
		}

		// Update token metrics
		for _, tokenID := range []int64{pair.Token0ID, pair.Token1ID} {
			_, err = a.db.ExecContext(ctx, `
				INSERT INTO token_metric (
					token_id, metric_type, start_time,
					trade_volume_usd, transaction_count,
					update_time
				) VALUES (
					$1, $2, $3,
					$4, 1,
					NOW()
				)
				ON CONFLICT (token_id, metric_type, start_time)
				DO UPDATE SET
					trade_volume_usd = token_metric.trade_volume_usd + EXCLUDED.trade_volume_usd,
					transaction_count = token_metric.transaction_count + 1,
					update_time = NOW()
			`, tokenID, window.Type, windowStart, volumeUSD)
			if err != nil {
				return fmt.Errorf("failed to update token metrics: %w", err)
			}
		}
	}

	return nil
}

// processMint handles Mint events
func (a *Aggregator) processMint(ctx context.Context, msg EventMessage) error {
	// Get pair info
	var pair types.Pair
	err := a.db.GetContext(ctx, &pair, `
		SELECT * FROM pair WHERE pair_address = $1
	`, msg.Event.ContractAddress)
	if err != nil {
		return fmt.Errorf("failed to get pair: %w", err)
	}

	// Extract amounts
	amount0 := types.ToBigInt(msg.DecodedData.Args["amount0"])
	amount1 := types.ToBigInt(msg.DecodedData.Args["amount1"])

	// Calculate USD values
	liquidityUSD := a.calculateUSDLiquidity(pair, amount0, amount1)

	// Update metrics for each window
	for _, window := range a.windows {
		windowStart := msg.BlockTimestamp.Truncate(window.Size)

		// Update pair metrics
		_, err = a.db.ExecContext(ctx, `
			INSERT INTO pair_metric (
				pair_id, metric_type, start_time,
				reserve_usd, total_supply,
				transaction_count, update_time
			) VALUES (
				$1, $2, $3,
				$4, $5,
				1, NOW()
			)
			ON CONFLICT (pair_id, metric_type, start_time)
			DO UPDATE SET
				reserve_usd = pair_metric.reserve_usd + EXCLUDED.reserve_usd,
				total_supply = pair_metric.total_supply + EXCLUDED.total_supply,
				transaction_count = pair_metric.transaction_count + 1,
				update_time = NOW()
		`, pair.ID, window.Type, windowStart,
			liquidityUSD, pair.TotalSupply)
		if err != nil {
			return fmt.Errorf("failed to update pair metrics: %w", err)
		}

		// Update token metrics
		for _, tokenID := range []int64{pair.Token0ID, pair.Token1ID} {
			_, err = a.db.ExecContext(ctx, `
				INSERT INTO token_metric (
					token_id, metric_type, start_time,
					total_liquidity, transaction_count,
					update_time
				) VALUES (
					$1, $2, $3,
					$4, 1,
					NOW()
				)
				ON CONFLICT (token_id, metric_type, start_time)
				DO UPDATE SET
					total_liquidity = token_metric.total_liquidity + EXCLUDED.total_liquidity,
					transaction_count = token_metric.transaction_count + 1,
					update_time = NOW()
			`, tokenID, window.Type, windowStart, liquidityUSD)
			if err != nil {
				return fmt.Errorf("failed to update token metrics: %w", err)
			}
		}
	}

	return nil
}

// processBurn handles Burn events
func (a *Aggregator) processBurn(ctx context.Context, msg EventMessage) error {
	// Get pair info
	var pair types.Pair
	err := a.db.GetContext(ctx, &pair, `
		SELECT * FROM pair WHERE pair_address = $1
	`, msg.Event.ContractAddress)
	if err != nil {
		return fmt.Errorf("failed to get pair: %w", err)
	}

	// Extract amounts
	amount0 := types.ToBigInt(msg.DecodedData.Args["amount0"])
	amount1 := types.ToBigInt(msg.DecodedData.Args["amount1"])

	// Calculate USD values
	liquidityUSD := a.calculateUSDLiquidity(pair, amount0, amount1)

	// Update metrics for each window
	for _, window := range a.windows {
		windowStart := msg.BlockTimestamp.Truncate(window.Size)

		// Update pair metrics
		_, err = a.db.ExecContext(ctx, `
			INSERT INTO pair_metric (
				pair_id, metric_type, start_time,
				reserve_usd, total_supply,
				transaction_count, update_time
			) VALUES (
				$1, $2, $3,
				$4, $5,
				1, NOW()
			)
			ON CONFLICT (pair_id, metric_type, start_time)
			DO UPDATE SET
				reserve_usd = pair_metric.reserve_usd - EXCLUDED.reserve_usd,
				total_supply = pair_metric.total_supply - EXCLUDED.total_supply,
				transaction_count = pair_metric.transaction_count + 1,
				update_time = NOW()
		`, pair.ID, window.Type, windowStart,
			liquidityUSD, pair.TotalSupply)
		if err != nil {
			return fmt.Errorf("failed to update pair metrics: %w", err)
		}

		// Update token metrics
		for _, tokenID := range []int64{pair.Token0ID, pair.Token1ID} {
			_, err = a.db.ExecContext(ctx, `
				INSERT INTO token_metric (
					token_id, metric_type, start_time,
					total_liquidity, transaction_count,
					update_time
				) VALUES (
					$1, $2, $3,
					$4, 1,
					NOW()
				)
				ON CONFLICT (token_id, metric_type, start_time)
				DO UPDATE SET
					total_liquidity = token_metric.total_liquidity - EXCLUDED.total_liquidity,
					transaction_count = token_metric.transaction_count + 1,
					update_time = NOW()
			`, tokenID, window.Type, windowStart, liquidityUSD)
			if err != nil {
				return fmt.Errorf("failed to update token metrics: %w", err)
			}
		}
	}

	return nil
}

// processSync handles Sync events
func (a *Aggregator) processSync(ctx context.Context, msg EventMessage) error {
	// Get pair info
	var pair types.Pair
	err := a.db.GetContext(ctx, &pair, `
		SELECT * FROM pair WHERE pair_address = $1
	`, msg.Event.ContractAddress)
	if err != nil {
		return fmt.Errorf("failed to get pair: %w", err)
	}

	// Extract reserves
	reserve0 := types.ToBigInt(msg.DecodedData.Args["reserve0"])
	reserve1 := types.ToBigInt(msg.DecodedData.Args["reserve1"])

	// Calculate USD values
	reserveUSD := a.calculateReserveUSD(pair, reserve0, reserve1)
	priceUSD := a.calculateTokenPrice(pair, reserve0, reserve1)

	// Update metrics for each window
	for _, window := range a.windows {
		windowStart := msg.BlockTimestamp.Truncate(window.Size)

		// Update pair metrics
		_, err = a.db.ExecContext(ctx, `
			INSERT INTO pair_metric (
				pair_id, metric_type, start_time,
				reserve0, reserve1, reserve_usd,
				transaction_count, update_time
			) VALUES (
				$1, $2, $3,
				$4, $5, $6,
				1, NOW()
			)
			ON CONFLICT (pair_id, metric_type, start_time)
			DO UPDATE SET
				reserve0 = EXCLUDED.reserve0,
				reserve1 = EXCLUDED.reserve1,
				reserve_usd = EXCLUDED.reserve_usd,
				transaction_count = pair_metric.transaction_count + 1,
				update_time = NOW()
		`, pair.ID, window.Type, windowStart,
			reserve0.String(), reserve1.String(), reserveUSD)
		if err != nil {
			return fmt.Errorf("failed to update pair metrics: %w", err)
		}

		// Update token metrics with price
		if pair.Token0IsUSDC {
			_, err = a.db.ExecContext(ctx, `
				INSERT INTO token_metric (
					token_id, metric_type, start_time,
					price_usd, transaction_count,
					update_time
				) VALUES (
					$1, $2, $3,
					$4, 1,
					NOW()
				)
				ON CONFLICT (token_id, metric_type, start_time)
				DO UPDATE SET
					price_usd = EXCLUDED.price_usd,
					transaction_count = token_metric.transaction_count + 1,
					update_time = NOW()
			`, pair.Token1ID, window.Type, windowStart, priceUSD)
			if err != nil {
				return fmt.Errorf("failed to update token metrics: %w", err)
			}
		}
	}

	return nil
}

// Helper functions

// calculateUSDVolume calculates volume in USD
func (a *Aggregator) calculateUSDVolume(pair types.Pair, volume0, volume1 *big.Int) float64 {
	if pair.Token0IsUSDC {
		vol := new(big.Float).SetInt(volume0)
		result, _ := vol.Float64()
		return result
	}
	vol := new(big.Float).SetInt(volume1)
	result, _ := vol.Float64()
	return result
}

// calculateUSDLiquidity calculates liquidity in USD
func (a *Aggregator) calculateUSDLiquidity(pair types.Pair, amount0, amount1 *big.Int) float64 {
	if pair.Token0IsUSDC {
		liq := new(big.Float).SetInt(amount0)
		result, _ := liq.Float64()
		return result * 2 // Both sides are equal in a balanced pool
	}
	liq := new(big.Float).SetInt(amount1)
	result, _ := liq.Float64()
	return result * 2
}

// calculateReserveUSD calculates total reserves in USD
func (a *Aggregator) calculateReserveUSD(pair types.Pair, reserve0, reserve1 *big.Int) float64 {
	if pair.Token0IsUSDC {
		res := new(big.Float).SetInt(reserve0)
		result, _ := res.Float64()
		return result * 2
	}
	res := new(big.Float).SetInt(reserve1)
	result, _ := res.Float64()
	return result * 2
}

// calculateTokenPrice calculates token price in USD
func (a *Aggregator) calculateTokenPrice(pair types.Pair, reserve0, reserve1 *big.Int) float64 {
	if reserve0.Sign() == 0 || reserve1.Sign() == 0 {
		return 0
	}

	if pair.Token0IsUSDC {
		price := new(big.Float).SetInt(reserve0)
		price.Quo(price, new(big.Float).SetInt(reserve1))
		result, _ := price.Float64()
		return result
	}
	price := new(big.Float).SetInt(reserve1)
	price.Quo(price, new(big.Float).SetInt(reserve0))
	result, _ := price.Float64()
	return result
}
