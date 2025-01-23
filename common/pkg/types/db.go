package types

import "time"

// Pair represents a trading pair
type Pair struct {
	ID                   int64     `db:"id"`
	ChainID              string    `db:"chain_id"`
	PairAddress          string    `db:"pair_address"`
	Token0ID             int64     `db:"token0_id"`
	Token1ID             int64     `db:"token1_id"`
	Token0IsUSDC         bool      `db:"token0_is_usdc"`
	Reserve0             float64   `db:"reserve0"`
	Reserve1             float64   `db:"reserve1"`
	TotalSupply          float64   `db:"total_supply"`
	VolumeToken0         float64   `db:"volume_token0"`
	VolumeToken1         float64   `db:"volume_token1"`
	CreatedAtTimestamp   time.Time `db:"created_at_timestamp"`
	CreatedAtBlockNumber int64     `db:"created_at_block_number"`
	UpdateTime           time.Time `db:"update_time"`
}

// Token represents a token
type Token struct {
	ID           int64     `db:"id"`
	ChainID      string    `db:"chain_id"`
	TokenAddress string    `db:"token_address"`
	TokenSymbol  string    `db:"token_symbol"`
	TokenName    string    `db:"token_name"`
	Decimals     int       `db:"decimals"`
	TrustScore   float64   `db:"trust_score"`
	CreateTime   time.Time `db:"create_time"`
	UpdateTime   time.Time `db:"update_time"`
}

// TokenMetric represents token metrics
type TokenMetric struct {
	ID               int64     `db:"id"`
	TokenID          int64     `db:"token_id"`
	MetricType       string    `db:"metric_type"`
	StartTime        time.Time `db:"start_time"`
	TotalSupply      float64   `db:"total_supply"`
	TradeVolumeUSD   float64   `db:"trade_volume_usd"`
	TransactionCount int       `db:"transaction_count"`
	TotalLiquidity   float64   `db:"total_liquidity"`
	PriceUSD         float64   `db:"price_usd"`
	UpdateTime       time.Time `db:"update_time"`
}

// PairMetric represents pair metrics
type PairMetric struct {
	ID               int64     `db:"id"`
	PairID           int64     `db:"pair_id"`
	MetricType       string    `db:"metric_type"`
	StartTime        time.Time `db:"start_time"`
	Token0Address    string    `db:"token0_address"`
	Token1Address    string    `db:"token1_address"`
	Reserve0         float64   `db:"reserve0"`
	Reserve1         float64   `db:"reserve1"`
	TotalSupply      float64   `db:"total_supply"`
	ReserveUSD       float64   `db:"reserve_usd"`
	Token0Volume     float64   `db:"token0_volume"`
	Token1Volume     float64   `db:"token1_volume"`
	VolumeUSD        float64   `db:"volume_usd"`
	TransactionCount int       `db:"transaction_count"`
	UpdateTime       time.Time `db:"update_time"`
}

// TWSwapFactory represents the factory contract
type TWSwapFactory struct {
	ID                int64     `db:"id"`
	ChainID           string    `db:"chain_id"`
	FactoryAddress    string    `db:"factory_address"`
	PairCount         int       `db:"pair_count"`
	TotalVolumeUSD    float64   `db:"total_volume_usd"`
	TotalLiquidityUSD float64   `db:"total_liquidity_usd"`
	TransactionCount  int       `db:"transaction_count"`
	UpdateTime        time.Time `db:"update_time"`
}
