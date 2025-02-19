package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.64

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	pgx "github.com/jackc/pgx/v5"
	"github.com/sjtutwilight/Twilight/api/graph/generated"
	"github.com/sjtutwilight/Twilight/api/graph/model"
)

// Transaction is the resolver for the transaction field.
func (r *queryResolver) Transaction(ctx context.Context, id string) (*model.Transaction, error) {
	panic(fmt.Errorf("not implemented: Transaction - transaction"))
}

// TransactionByHash is the resolver for the transactionByHash field.
func (r *queryResolver) TransactionByHash(ctx context.Context, chainID string, hash string) (*model.Transaction, error) {
	panic(fmt.Errorf("not implemented: TransactionByHash - transactionByHash"))
}

// Transactions is the resolver for the transactions field.
func (r *queryResolver) Transactions(ctx context.Context, chainID *string, fromBlock *int, toBlock *int, limit *int, offset *int) ([]*model.Transaction, error) {
	panic(fmt.Errorf("not implemented: Transactions - transactions"))
}

// Token is the resolver for the token field.
func (r *queryResolver) Token(ctx context.Context, id string) (*model.Token, error) {
	query := `
		SELECT 
			id, chain_id, chain_name, token_address, token_symbol, token_name, 
			token_decimals, hype_score, supply_usd, liquidity_usd, 
			create_time, update_time
		FROM token 
		WHERE id = $1
	`

	// Get a connection from the pool
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %v", err)
	}
	defer conn.Release()

	// Execute query
	row := conn.QueryRow(ctx, query, id)

	token := &model.Token{}
	var createTime, updateTime time.Time
	var supplyUsd, liquidityUsd sql.NullString
	var tokenSymbol, tokenName, chainName sql.NullString
	var tokenDecimals, hypeScore sql.NullInt32

	err = row.Scan(
		&token.ID,
		&token.ChainID,
		&chainName,
		&token.TokenAddress,
		&tokenSymbol,
		&tokenName,
		&tokenDecimals,
		&hypeScore,
		&supplyUsd,
		&liquidityUsd,
		&createTime,
		&updateTime,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan token: %v", err)
	}

	// Handle nullable fields
	if chainName.Valid {
		token.ChainName = &chainName.String
	}
	if tokenSymbol.Valid {
		token.TokenSymbol = &tokenSymbol.String
	}
	if tokenName.Valid {
		token.TokenName = &tokenName.String
	}
	if tokenDecimals.Valid {
		val := int(tokenDecimals.Int32)
		token.TokenDecimals = &val
	}
	if hypeScore.Valid {
		val := int(hypeScore.Int32)
		token.HypeScore = &val
	}
	if supplyUsd.Valid {
		token.SupplyUsd = &supplyUsd.String
	}
	if liquidityUsd.Valid {
		token.LiquidityUsd = &liquidityUsd.String
	}

	token.CreateTime = createTime.Format(time.RFC3339)
	token.UpdateTime = updateTime.Format(time.RFC3339)

	return token, nil
}

// TokenByAddress is the resolver for the tokenByAddress field.
func (r *queryResolver) TokenByAddress(ctx context.Context, chainID string, address string) (*model.Token, error) {
	query := `
		SELECT 
			id, chain_id, chain_name, token_address, token_symbol, token_name, 
			token_decimals, hype_score, supply_usd, liquidity_usd, 
			create_time, update_time
		FROM token 
		WHERE chain_id = $1 AND token_address = $2
	`

	// Get a connection from the pool
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %v", err)
	}
	defer conn.Release()

	// Execute query
	row := conn.QueryRow(ctx, query, chainID, address)

	token := &model.Token{}
	var createTime, updateTime time.Time
	var supplyUsd, liquidityUsd sql.NullString
	var tokenSymbol, tokenName, chainName sql.NullString
	var tokenDecimals, hypeScore sql.NullInt32

	err = row.Scan(
		&token.ID,
		&token.ChainID,
		&chainName,
		&token.TokenAddress,
		&tokenSymbol,
		&tokenName,
		&tokenDecimals,
		&hypeScore,
		&supplyUsd,
		&liquidityUsd,
		&createTime,
		&updateTime,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan token: %v", err)
	}

	// Handle nullable fields
	if chainName.Valid {
		token.ChainName = &chainName.String
	}
	if tokenSymbol.Valid {
		token.TokenSymbol = &tokenSymbol.String
	}
	if tokenName.Valid {
		token.TokenName = &tokenName.String
	}
	if tokenDecimals.Valid {
		val := int(tokenDecimals.Int32)
		token.TokenDecimals = &val
	}
	if hypeScore.Valid {
		val := int(hypeScore.Int32)
		token.HypeScore = &val
	}
	if supplyUsd.Valid {
		token.SupplyUsd = &supplyUsd.String
	}
	if liquidityUsd.Valid {
		token.LiquidityUsd = &liquidityUsd.String
	}

	token.CreateTime = createTime.Format(time.RFC3339)
	token.UpdateTime = updateTime.Format(time.RFC3339)

	return token, nil
}

// Tokens is the resolver for the tokens field.
func (r *queryResolver) Tokens(ctx context.Context, chainID *string, limit *int, offset *int) ([]*model.Token, error) {
	query := `
		SELECT 
			id, chain_id, chain_name, token_address, token_symbol, token_name, 
			token_decimals, hype_score, supply_usd, liquidity_usd, 
			create_time, update_time
		FROM token
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 0

	// Add chain_id filter if provided
	if chainID != nil {
		argCount++
		query += fmt.Sprintf(" AND chain_id = $%d", argCount)
		args = append(args, *chainID)
	}

	// Add ORDER BY clause
	query += " ORDER BY create_time DESC"

	// Add pagination
	if limit != nil {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, *limit)
	}
	if offset != nil {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, *offset)
	}

	// Get a connection from the pool
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %v", err)
	}
	defer conn.Release()

	// Execute query
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tokens: %v", err)
	}
	defer rows.Close()

	var tokens []*model.Token
	for rows.Next() {
		token := &model.Token{}
		var createTime, updateTime time.Time
		var supplyUsd, liquidityUsd sql.NullString
		var tokenSymbol, tokenName, chainName sql.NullString
		var tokenDecimals, hypeScore sql.NullInt32

		err := rows.Scan(
			&token.ID,
			&token.ChainID,
			&chainName,
			&token.TokenAddress,
			&tokenSymbol,
			&tokenName,
			&tokenDecimals,
			&hypeScore,
			&supplyUsd,
			&liquidityUsd,
			&createTime,
			&updateTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan token row: %v", err)
		}

		// Handle nullable fields
		if chainName.Valid {
			token.ChainName = &chainName.String
		}
		if tokenSymbol.Valid {
			token.TokenSymbol = &tokenSymbol.String
		}
		if tokenName.Valid {
			token.TokenName = &tokenName.String
		}
		if tokenDecimals.Valid {
			val := int(tokenDecimals.Int32)
			token.TokenDecimals = &val
		}
		if hypeScore.Valid {
			val := int(hypeScore.Int32)
			token.HypeScore = &val
		}
		if supplyUsd.Valid {
			token.SupplyUsd = &supplyUsd.String
		}
		if liquidityUsd.Valid {
			token.LiquidityUsd = &liquidityUsd.String
		}

		token.CreateTime = createTime.Format(time.RFC3339)
		token.UpdateTime = updateTime.Format(time.RFC3339)

		tokens = append(tokens, token)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating token rows: %v", err)
	}

	return tokens, nil
}

// Pair is the resolver for the pair field.
func (r *queryResolver) Pair(ctx context.Context, id string) (*model.TwswapPair, error) {
	panic(fmt.Errorf("not implemented: Pair - pair"))
}

// PairByAddress is the resolver for the pairByAddress field.
func (r *queryResolver) PairByAddress(ctx context.Context, chainID string, address string) (*model.TwswapPair, error) {
	panic(fmt.Errorf("not implemented: PairByAddress - pairByAddress"))
}

// Pairs is the resolver for the pairs field.
func (r *queryResolver) Pairs(ctx context.Context, chainID *string, token0Address *string, token1Address *string, limit *int, offset *int) ([]*model.TwswapPair, error) {
	panic(fmt.Errorf("not implemented: Pairs - pairs"))
}

// TokenMetrics is the resolver for the tokenMetrics field.
func (r *queryResolver) TokenMetrics(ctx context.Context, tokenID string, timeWindow string, fromTime *string, toTime *string, limit *int) ([]*model.TokenMetric, error) {
	// Log input parameters for debugging
	log.Printf("TokenMetrics called with tokenID: %s, timeWindow: %s", tokenID, timeWindow)

	// Validate time window
	validTimeWindows := map[string]bool{
		"20s": true, "1m": true, "5m": true, "30m": true,
	}
	if !validTimeWindows[timeWindow] {
		return nil, fmt.Errorf("invalid time window: %s. Valid values are: 20s, 1m, 5m, 30m", timeWindow)
	}

	query := `
		SELECT 
			tm.id, tm.token_id, tm.time_window, tm.end_time,
			tm.volume_usd, tm.txcnt, tm.token_price_usd,
			tm.buy_pressure_usd, tm.buyers_count, tm.sellers_count,
			tm.buy_volume_usd, tm.sell_volume_usd, 
			tm.makers_count, tm.buy_count, tm.sell_count,
			tm.update_time,
			t.id as t_id, t.chain_id, t.chain_name, t.token_address, 
			t.token_symbol, t.token_name, t.token_decimals, t.hype_score,
			t.supply_usd, t.liquidity_usd, t.create_time, t.update_time as t_update_time
		FROM token_metric tm
		JOIN token t ON t.id = tm.token_id
		WHERE tm.token_id = $1 
		AND tm.time_window = $2
	`
	args := []interface{}{tokenID, timeWindow}
	argCount := 2

	if fromTime != nil {
		query += fmt.Sprintf(" AND tm.end_time >= $%d", argCount+1)
		args = append(args, *fromTime)
		argCount++
	}

	if toTime != nil {
		query += fmt.Sprintf(" AND tm.end_time <= $%d", argCount+1)
		args = append(args, *toTime)
		argCount++
	}

	query += " ORDER BY tm.end_time DESC"

	if limit != nil {
		query += fmt.Sprintf(" LIMIT $%d", argCount+1)
		args = append(args, *limit)
	}

	// Get a connection from the pool
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %v", err)
	}
	defer conn.Release()

	// Execute query
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query token metrics: %v", err)
	}
	defer rows.Close()

	var metrics []*model.TokenMetric
	for rows.Next() {
		m := &model.TokenMetric{}
		token := &model.Token{}
		var endTime sql.NullTime
		var updateTime sql.NullTime
		var createTime sql.NullTime
		var tokenUpdateTime sql.NullTime
		var volumeUsd, tokenPriceUsd, buyPressureUsd, buyVolumeUsd, sellVolumeUsd sql.NullString
		var txcnt, buyersCount, sellersCount, makersCount, buyCount, sellCount sql.NullInt32
		var tokenSymbol, tokenName, chainName sql.NullString
		var tokenDecimals, hypeScore sql.NullInt32
		var supplyUsd, liquidityUsd sql.NullString

		err := rows.Scan(
			&m.ID,
			&m.TokenID,
			&m.TimeWindow,
			&endTime,
			&volumeUsd,
			&txcnt,
			&tokenPriceUsd,
			&buyPressureUsd,
			&buyersCount,
			&sellersCount,
			&buyVolumeUsd,
			&sellVolumeUsd,
			&makersCount,
			&buyCount,
			&sellCount,
			&updateTime,
			&token.ID,
			&token.ChainID,
			&chainName,
			&token.TokenAddress,
			&tokenSymbol,
			&tokenName,
			&tokenDecimals,
			&hypeScore,
			&supplyUsd,
			&liquidityUsd,
			&createTime,
			&tokenUpdateTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan token metric: %v", err)
		}

		// Handle metric nullable fields
		if volumeUsd.Valid {
			m.VolumeUsd = &volumeUsd.String
		}
		if txcnt.Valid {
			val := int(txcnt.Int32)
			m.Txcnt = &val
		}
		if tokenPriceUsd.Valid {
			m.TokenPriceUsd = &tokenPriceUsd.String
		}
		if buyPressureUsd.Valid {
			m.BuyPressureUsd = &buyPressureUsd.String
		}
		if buyersCount.Valid {
			val := int(buyersCount.Int32)
			m.BuyersCount = &val
		}
		if sellersCount.Valid {
			val := int(sellersCount.Int32)
			m.SellersCount = &val
		}
		if buyVolumeUsd.Valid {
			m.BuyVolumeUsd = &buyVolumeUsd.String
		}
		if sellVolumeUsd.Valid {
			m.SellVolumeUsd = &sellVolumeUsd.String
		}
		if makersCount.Valid {
			val := int(makersCount.Int32)
			m.MakersCount = &val
		}
		if buyCount.Valid {
			val := int(buyCount.Int32)
			m.BuyCount = &val
		}
		if sellCount.Valid {
			val := int(sellCount.Int32)
			m.SellCount = &val
		}

		// Handle token nullable fields
		if chainName.Valid {
			token.ChainName = &chainName.String
		}
		if tokenSymbol.Valid {
			token.TokenSymbol = &tokenSymbol.String
		}
		if tokenName.Valid {
			token.TokenName = &tokenName.String
		}
		if tokenDecimals.Valid {
			val := int(tokenDecimals.Int32)
			token.TokenDecimals = &val
		}
		if hypeScore.Valid {
			val := int(hypeScore.Int32)
			token.HypeScore = &val
		}
		if supplyUsd.Valid {
			token.SupplyUsd = &supplyUsd.String
		}
		if liquidityUsd.Valid {
			token.LiquidityUsd = &liquidityUsd.String
		}

		// Handle nullable time fields
		if endTime.Valid {
			m.EndTime = endTime.Time.Format(time.RFC3339)
		}
		if updateTime.Valid {
			m.UpdateTime = updateTime.Time.Format(time.RFC3339)
		}
		if createTime.Valid {
			token.CreateTime = createTime.Time.Format(time.RFC3339)
		}
		if tokenUpdateTime.Valid {
			token.UpdateTime = tokenUpdateTime.Time.Format(time.RFC3339)
		}

		m.Token = token
		metrics = append(metrics, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating token metrics rows: %v", err)
	}

	return metrics, nil
}

// TokenAnalytics is the resolver for the tokenAnalytics field.
func (r *queryResolver) TokenAnalytics(ctx context.Context, tokenID string, timeWindow string, from string, to string) (*model.TokenAnalytics, error) {
	// Validate time window
	validTimeWindows := map[string]bool{
		"20s": true, "1m": true, "5m": true, "30m": true,
	}
	if !validTimeWindows[timeWindow] {
		return nil, fmt.Errorf("invalid time window: %s. Valid values are: 20s, 1m, 5m, 30m", timeWindow)
	}

	// First get the token information
	tokenQuery := `
		SELECT 
			id, chain_id, chain_name, token_address, token_symbol, token_name, 
			token_decimals, hype_score, supply_usd, liquidity_usd, 
			create_time, update_time
		FROM token 
		WHERE id = $1
	`

	// Get a connection from the pool
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %v", err)
	}
	defer conn.Release()

	// Execute token query
	row := conn.QueryRow(ctx, tokenQuery, tokenID)

	token := &model.Token{}
	var createTime, updateTime time.Time
	var supplyUsd, liquidityUsd sql.NullString
	var tokenSymbol, tokenName, chainName sql.NullString
	var tokenDecimals, hypeScore sql.NullInt32

	err = row.Scan(
		&token.ID,
		&token.ChainID,
		&chainName,
		&token.TokenAddress,
		&tokenSymbol,
		&tokenName,
		&tokenDecimals,
		&hypeScore,
		&supplyUsd,
		&liquidityUsd,
		&createTime,
		&updateTime,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("token not found: %s", tokenID)
		}
		return nil, fmt.Errorf("failed to scan token: %v", err)
	}

	// Handle nullable fields
	if chainName.Valid {
		token.ChainName = &chainName.String
	}
	if tokenSymbol.Valid {
		token.TokenSymbol = &tokenSymbol.String
	}
	if tokenName.Valid {
		token.TokenName = &tokenName.String
	}
	if tokenDecimals.Valid {
		val := int(tokenDecimals.Int32)
		token.TokenDecimals = &val
	}
	if hypeScore.Valid {
		val := int(hypeScore.Int32)
		token.HypeScore = &val
	}
	if supplyUsd.Valid {
		token.SupplyUsd = &supplyUsd.String
	}
	if liquidityUsd.Valid {
		token.LiquidityUsd = &liquidityUsd.String
	}

	token.CreateTime = createTime.Format(time.RFC3339)
	token.UpdateTime = updateTime.Format(time.RFC3339)

	// Now get the time series data
	metricsQuery := `
		SELECT 
			end_time,
			token_price_usd,
			buy_volume_usd,
			sell_volume_usd,
			volume_usd,
			txcnt
		FROM token_metric
		WHERE token_id = $1 
		AND time_window = $2
		AND end_time >= $3::timestamp
		AND end_time <= $4::timestamp
		ORDER BY end_time ASC
	`

	// Log query and parameters for debugging
	log.Printf("Executing metrics query: %s with args: [%s, %s, %s, %s]", metricsQuery, tokenID, timeWindow, from, to)

	// Execute metrics query
	rows, err := conn.Query(ctx, metricsQuery, tokenID, timeWindow, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query token metrics: %v", err)
	}
	defer rows.Close()

	var dataPoints []*model.TimeSeriesDataPoint
	for rows.Next() {
		dp := &model.TimeSeriesDataPoint{}
		var timestamp time.Time
		var tokenPriceUsd, buyVolumeUsd, sellVolumeUsd, volumeUsd sql.NullString
		var txcnt sql.NullInt32

		err := rows.Scan(
			&timestamp,
			&tokenPriceUsd,
			&buyVolumeUsd,
			&sellVolumeUsd,
			&volumeUsd,
			&txcnt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric row: %v", err)
		}

		dp.Timestamp = timestamp.Format(time.RFC3339)
		if tokenPriceUsd.Valid {
			dp.TokenPriceUsd = &tokenPriceUsd.String
		}
		if buyVolumeUsd.Valid {
			dp.BuyVolumeUsd = &buyVolumeUsd.String
		}
		if sellVolumeUsd.Valid {
			dp.SellVolumeUsd = &sellVolumeUsd.String
		}
		if volumeUsd.Valid {
			dp.VolumeUsd = &volumeUsd.String
		}
		if txcnt.Valid {
			val := int(txcnt.Int32)
			dp.Txcnt = &val
		}

		dataPoints = append(dataPoints, dp)
	}

	// Log the number of data points found
	log.Printf("Found %d data points for token %s with time window %s", len(dataPoints), tokenID, timeWindow)

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating metric rows: %v", err)
	}

	return &model.TokenAnalytics{
		Token:      token,
		TimeWindow: timeWindow,
		DataPoints: dataPoints,
	}, nil
}

// PairMetrics is the resolver for the pairMetrics field.
func (r *queryResolver) PairMetrics(ctx context.Context, pairID string, timeWindow string, fromTime *string, toTime *string, limit *int) ([]*model.TwswapPairMetric, error) {
	// Validate time window
	validTimeWindows := map[string]bool{
		"1m": true, "5m": true, "15m": true, "30m": true,
		"1h": true, "4h": true, "12h": true,
		"1d": true, "1w": true,
	}
	if !validTimeWindows[timeWindow] {
		return nil, fmt.Errorf("invalid time window: %s", timeWindow)
	}

	query := `
		SELECT 
			id, pair_id, time_window, end_time,
			token0_reserve, token1_reserve, reserve_usd,
			token0_volume_usd, token1_volume_usd, volume_usd,
			txcnt
		FROM twswap_pair_metrics
		WHERE pair_id = $1 
		AND time_window = $2
	`
	args := []interface{}{pairID, timeWindow}
	argCount := 2

	if fromTime != nil {
		query += fmt.Sprintf(" AND end_time >= $%d", argCount+1)
		args = append(args, *fromTime)
		argCount++
	}

	if toTime != nil {
		query += fmt.Sprintf(" AND end_time <= $%d", argCount+1)
		args = append(args, *toTime)
		argCount++
	}

	query += " ORDER BY end_time DESC"

	if limit != nil {
		query += fmt.Sprintf(" LIMIT $%d", argCount+1)
		args = append(args, *limit)
	}

	// Get a connection from the pool
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %v", err)
	}
	defer conn.Release()

	// Execute query using the acquired connection
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query pair metrics: %v", err)
	}
	defer rows.Close()

	var metrics []*model.TwswapPairMetric
	for rows.Next() {
		m := &model.TwswapPairMetric{}
		var endTime time.Time

		err := rows.Scan(
			&m.ID,
			&m.PairID,
			&m.TimeWindow,
			&endTime,
			&m.Token0Reserve,
			&m.Token1Reserve,
			&m.ReserveUsd,
			&m.Token0VolumeUsd,
			&m.Token1VolumeUsd,
			&m.VolumeUsd,
			&m.Txcnt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pair metric: %v", err)
		}

		m.EndTime = endTime.Format(time.RFC3339)
		metrics = append(metrics, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pair metrics rows: %v", err)
	}

	return metrics, nil
}

// FactoryMetrics is the resolver for the factoryMetrics field.
func (r *queryResolver) FactoryMetrics(ctx context.Context, chainID string, factoryAddress string, timeWindow string, fromTime *string, toTime *string, limit *int) ([]*model.TwswapFactory, error) {
	// Validate time window
	validTimeWindows := map[string]bool{
		"1m": true, "5m": true, "15m": true, "30m": true,
		"1h": true, "4h": true, "12h": true,
		"1d": true, "1w": true,
	}
	if !validTimeWindows[timeWindow] {
		return nil, fmt.Errorf("invalid time window: %s", timeWindow)
	}

	query := `
		SELECT 
			id, chain_id, factory_address, time_window, 
			end_time, pair_count, volume_usd, liquidity_usd, 
			txcnt, update_time
		FROM twswap_factory_metrics
		WHERE chain_id = $1 
		AND factory_address = $2 
		AND time_window = $3
	`
	args := []interface{}{chainID, factoryAddress, timeWindow}
	argCount := 3

	if fromTime != nil {
		query += fmt.Sprintf(" AND end_time >= $%d", argCount+1)
		args = append(args, *fromTime)
		argCount++
	}

	if toTime != nil {
		query += fmt.Sprintf(" AND end_time <= $%d", argCount+1)
		args = append(args, *toTime)
		argCount++
	}

	query += " ORDER BY end_time DESC"

	if limit != nil {
		query += fmt.Sprintf(" LIMIT $%d", argCount+1)
		args = append(args, *limit)
	}

	// Get a connection from the pool
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %v", err)
	}
	defer conn.Release()

	// Execute query using the acquired connection
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query factory metrics: %v", err)
	}
	defer rows.Close()

	var metrics []*model.TwswapFactory
	for rows.Next() {
		m := &model.TwswapFactory{}
		var endTime time.Time
		var updateTime time.Time

		err := rows.Scan(
			&m.ID,
			&m.ChainID,
			&m.FactoryAddress,
			&m.TimeWindow,
			&endTime,
			&m.PairCount,
			&m.VolumeUsd,
			&m.LiquidityUsd,
			&m.Txcnt,
			&updateTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan factory metric: %v", err)
		}

		m.EndTime = endTime.Format(time.RFC3339)
		m.UpdateTime = updateTime.Format(time.RFC3339)
		metrics = append(metrics, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating factory metrics rows: %v", err)
	}

	return metrics, nil
}

// TokenStats is the resolver for the tokenStats field.
func (r *queryResolver) TokenStats(ctx context.Context, tokenID string) (*model.TokenStats, error) {
	// 获取最新的 5m 时间窗口数据
	query := `
		WITH current_metric AS (
			SELECT 
				tm.token_price_usd,
				tm.volume_usd,
				tm.buy_pressure_usd,
				tm.end_time,
				t.*
			FROM token_metric tm
			JOIN token t ON t.id = tm.token_id
			WHERE tm.token_id = $1 
			AND tm.time_window = '5m'
			ORDER BY tm.end_time DESC
			LIMIT 1
		),
		prev_hour_metric AS (
			SELECT token_price_usd
			FROM token_metric
			WHERE token_id = $1 
			AND time_window = '1h'
			ORDER BY end_time DESC
			LIMIT 1
		)
		SELECT 
			c.*,
			p.token_price_usd as prev_price
		FROM current_metric c
		LEFT JOIN prev_hour_metric p ON true
	`

	// Get a connection from the pool
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %v", err)
	}
	defer conn.Release()

	// Execute query
	row := conn.QueryRow(ctx, query, tokenID)

	stats := &model.TokenStats{}
	token := &model.Token{}
	var endTime, createTime, updateTime time.Time
	var currentPrice, prevPrice, volumeUsd, buyPressureUsd sql.NullString
	var tokenSymbol, tokenName, chainName sql.NullString
	var tokenDecimals, hypeScore sql.NullInt32
	var supplyUsd, liquidityUsd sql.NullString

	err = row.Scan(
		&currentPrice,
		&volumeUsd,
		&buyPressureUsd,
		&endTime,
		&token.ID,
		&token.ChainID,
		&chainName,
		&token.TokenAddress,
		&tokenSymbol,
		&tokenName,
		&tokenDecimals,
		&hypeScore,
		&supplyUsd,
		&liquidityUsd,
		&createTime,
		&updateTime,
		&prevPrice,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("token not found: %s", tokenID)
		}
		return nil, fmt.Errorf("failed to scan token stats: %v", err)
	}

	// Handle token nullable fields
	if chainName.Valid {
		token.ChainName = &chainName.String
	}
	if tokenSymbol.Valid {
		token.TokenSymbol = &tokenSymbol.String
	}
	if tokenName.Valid {
		token.TokenName = &tokenName.String
	}
	if tokenDecimals.Valid {
		val := int(tokenDecimals.Int32)
		token.TokenDecimals = &val
	}
	if hypeScore.Valid {
		val := int(hypeScore.Int32)
		token.HypeScore = &val
	}
	if supplyUsd.Valid {
		token.SupplyUsd = &supplyUsd.String
	}
	if liquidityUsd.Valid {
		token.LiquidityUsd = &liquidityUsd.String
	}

	token.CreateTime = createTime.Format(time.RFC3339)
	token.UpdateTime = updateTime.Format(time.RFC3339)

	// Set TokenStats fields
	stats.Token = token
	stats.LastUpdate = endTime.Format(time.RFC3339)

	if currentPrice.Valid {
		stats.CurrentPrice = currentPrice.String
	} else {
		stats.CurrentPrice = "0"
	}

	if volumeUsd.Valid {
		stats.Volume1h = volumeUsd.String
	} else {
		stats.Volume1h = "0"
	}

	if buyPressureUsd.Valid {
		stats.BuyPressure1h = buyPressureUsd.String
	} else {
		stats.BuyPressure1h = "0"
	}

	// Calculate price change percentage
	if currentPrice.Valid && prevPrice.Valid {
		current, err := strconv.ParseFloat(currentPrice.String, 64)
		if err != nil {
			current = 0
		}
		prev, err := strconv.ParseFloat(prevPrice.String, 64)
		if err != nil {
			prev = 0
		}
		if prev > 0 {
			change := ((current - prev) / prev) * 100
			stats.PriceChange1h = fmt.Sprintf("%.2f", change)
		} else {
			stats.PriceChange1h = "0"
		}
	} else {
		stats.PriceChange1h = "0"
	}

	return stats, nil
}

// TokenMetricsByWindow is the resolver for the tokenMetricsByWindow field.
func (r *queryResolver) TokenMetricsByWindow(ctx context.Context, tokenID string, timeWindow string) (*model.TokenMetric, error) {
	// 使用新的索引来获取最新的指标数据
	query := `
		SELECT 
			tm.id, tm.token_id, tm.time_window, tm.end_time,
			tm.volume_usd, tm.txcnt, tm.token_price_usd,
			tm.buy_pressure_usd, tm.buyers_count, tm.sellers_count,
			tm.buy_volume_usd, tm.sell_volume_usd, 
			tm.makers_count, tm.buy_count, tm.sell_count,
			tm.update_time,
			t.id as t_id, t.chain_id, t.chain_name, t.token_address, 
			t.token_symbol, t.token_name, t.token_decimals, t.hype_score,
			t.supply_usd, t.liquidity_usd, t.create_time, t.update_time as t_update_time
		FROM token_metric tm
		JOIN token t ON t.id = tm.token_id
		WHERE tm.token_id = $1 
		AND tm.time_window = $2
		ORDER BY tm.end_time DESC
		LIMIT 1
	`

	// Get a connection from the pool
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %v", err)
	}
	defer conn.Release()

	// Execute query
	row := conn.QueryRow(ctx, query, tokenID, timeWindow)

	m := &model.TokenMetric{}
	token := &model.Token{}
	var endTime sql.NullTime
	var updateTime sql.NullTime
	var createTime sql.NullTime
	var tokenUpdateTime sql.NullTime
	var volumeUsd, tokenPriceUsd, buyPressureUsd, buyVolumeUsd, sellVolumeUsd sql.NullString
	var txcnt, buyersCount, sellersCount, makersCount, buyCount, sellCount sql.NullInt32
	var tokenSymbol, tokenName, chainName sql.NullString
	var tokenDecimals, hypeScore sql.NullInt32
	var supplyUsd, liquidityUsd sql.NullString

	err = row.Scan(
		&m.ID,
		&m.TokenID,
		&m.TimeWindow,
		&endTime,
		&volumeUsd,
		&txcnt,
		&tokenPriceUsd,
		&buyPressureUsd,
		&buyersCount,
		&sellersCount,
		&buyVolumeUsd,
		&sellVolumeUsd,
		&makersCount,
		&buyCount,
		&sellCount,
		&updateTime,
		&token.ID,
		&token.ChainID,
		&chainName,
		&token.TokenAddress,
		&tokenSymbol,
		&tokenName,
		&tokenDecimals,
		&hypeScore,
		&supplyUsd,
		&liquidityUsd,
		&createTime,
		&tokenUpdateTime,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan token metric: %v", err)
	}

	// Handle metric nullable fields
	if volumeUsd.Valid {
		m.VolumeUsd = &volumeUsd.String
	}
	if txcnt.Valid {
		val := int(txcnt.Int32)
		m.Txcnt = &val
	}
	if tokenPriceUsd.Valid {
		m.TokenPriceUsd = &tokenPriceUsd.String
	}
	if buyPressureUsd.Valid {
		m.BuyPressureUsd = &buyPressureUsd.String
	}
	if buyersCount.Valid {
		val := int(buyersCount.Int32)
		m.BuyersCount = &val
	}
	if sellersCount.Valid {
		val := int(sellersCount.Int32)
		m.SellersCount = &val
	}
	if buyVolumeUsd.Valid {
		m.BuyVolumeUsd = &buyVolumeUsd.String
	}
	if sellVolumeUsd.Valid {
		m.SellVolumeUsd = &sellVolumeUsd.String
	}
	if makersCount.Valid {
		val := int(makersCount.Int32)
		m.MakersCount = &val
	}
	if buyCount.Valid {
		val := int(buyCount.Int32)
		m.BuyCount = &val
	}
	if sellCount.Valid {
		val := int(sellCount.Int32)
		m.SellCount = &val
	}

	// Handle token nullable fields
	if chainName.Valid {
		token.ChainName = &chainName.String
	}
	if tokenSymbol.Valid {
		token.TokenSymbol = &tokenSymbol.String
	}
	if tokenName.Valid {
		token.TokenName = &tokenName.String
	}
	if tokenDecimals.Valid {
		val := int(tokenDecimals.Int32)
		token.TokenDecimals = &val
	}
	if hypeScore.Valid {
		val := int(hypeScore.Int32)
		token.HypeScore = &val
	}
	if supplyUsd.Valid {
		token.SupplyUsd = &supplyUsd.String
	}
	if liquidityUsd.Valid {
		token.LiquidityUsd = &liquidityUsd.String
	}

	// Handle nullable time fields
	if endTime.Valid {
		m.EndTime = endTime.Time.Format(time.RFC3339)
	}
	if updateTime.Valid {
		m.UpdateTime = updateTime.Time.Format(time.RFC3339)
	}
	if createTime.Valid {
		token.CreateTime = createTime.Time.Format(time.RFC3339)
	}
	if tokenUpdateTime.Valid {
		token.UpdateTime = tokenUpdateTime.Time.Format(time.RFC3339)
	}

	m.Token = token
	return m, nil
}

// TokensStats is the resolver for the tokensStats field.
func (r *queryResolver) TokensStats(ctx context.Context, tokenIds []string) ([]*model.TokenStats, error) {
	// 使用 IN 子句批量查询多个 token 的最新数据
	query := `
		WITH current_metrics AS (
			SELECT DISTINCT ON (tm.token_id)
				tm.token_id,
				tm.token_price_usd,
				tm.volume_usd,
				tm.buy_pressure_usd,
				tm.end_time,
				t.*
			FROM token_metric tm
			JOIN token t ON t.id = tm.token_id
			WHERE tm.token_id = ANY($1) 
			AND tm.time_window = '5m'
			ORDER BY tm.token_id, tm.end_time DESC
		),
		prev_hour_metrics AS (
			SELECT DISTINCT ON (token_id)
				token_id,
				token_price_usd as prev_price
			FROM token_metric
			WHERE token_id = ANY($1)
			AND time_window = '1h'
			ORDER BY token_id, end_time DESC
		)
		SELECT 
			c.*,
			p.prev_price
		FROM current_metrics c
		LEFT JOIN prev_hour_metrics p ON c.token_id = p.token_id
	`

	// Get a connection from the pool
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %v", err)
	}
	defer conn.Release()

	// Execute query
	rows, err := conn.Query(ctx, query, tokenIds)
	if err != nil {
		return nil, fmt.Errorf("failed to query tokens stats: %v", err)
	}
	defer rows.Close()

	var statsSlice []*model.TokenStats
	for rows.Next() {
		stats := &model.TokenStats{}
		token := &model.Token{}
		var endTime, createTime, updateTime time.Time
		var currentPrice, prevPrice, volumeUsd, buyPressureUsd sql.NullString
		var tokenSymbol, tokenName, chainName sql.NullString
		var tokenDecimals, hypeScore sql.NullInt32
		var supplyUsd, liquidityUsd sql.NullString

		err := rows.Scan(
			&token.ID,
			&currentPrice,
			&volumeUsd,
			&buyPressureUsd,
			&endTime,
			&token.ChainID,
			&chainName,
			&token.TokenAddress,
			&tokenSymbol,
			&tokenName,
			&tokenDecimals,
			&hypeScore,
			&supplyUsd,
			&liquidityUsd,
			&createTime,
			&updateTime,
			&prevPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan token stats: %v", err)
		}

		// Handle token nullable fields
		if chainName.Valid {
			token.ChainName = &chainName.String
		}
		if tokenSymbol.Valid {
			token.TokenSymbol = &tokenSymbol.String
		}
		if tokenName.Valid {
			token.TokenName = &tokenName.String
		}
		if tokenDecimals.Valid {
			val := int(tokenDecimals.Int32)
			token.TokenDecimals = &val
		}
		if hypeScore.Valid {
			val := int(hypeScore.Int32)
			token.HypeScore = &val
		}
		if supplyUsd.Valid {
			token.SupplyUsd = &supplyUsd.String
		}
		if liquidityUsd.Valid {
			token.LiquidityUsd = &liquidityUsd.String
		}

		token.CreateTime = createTime.Format(time.RFC3339)
		token.UpdateTime = updateTime.Format(time.RFC3339)

		// Set TokenStats fields
		stats.Token = token
		stats.LastUpdate = endTime.Format(time.RFC3339)

		if currentPrice.Valid {
			stats.CurrentPrice = currentPrice.String
		} else {
			stats.CurrentPrice = "0"
		}

		if volumeUsd.Valid {
			stats.Volume1h = volumeUsd.String
		} else {
			stats.Volume1h = "0"
		}

		if buyPressureUsd.Valid {
			stats.BuyPressure1h = buyPressureUsd.String
		} else {
			stats.BuyPressure1h = "0"
		}

		// Calculate price change percentage
		if currentPrice.Valid && prevPrice.Valid {
			current, err := strconv.ParseFloat(currentPrice.String, 64)
			if err != nil {
				current = 0
			}
			prev, err := strconv.ParseFloat(prevPrice.String, 64)
			if err != nil {
				prev = 0
			}
			if prev > 0 {
				change := ((current - prev) / prev) * 100
				stats.PriceChange1h = fmt.Sprintf("%.2f", change)
			} else {
				stats.PriceChange1h = "0"
			}
		} else {
			stats.PriceChange1h = "0"
		}

		statsSlice = append(statsSlice, stats)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating token stats rows: %v", err)
	}

	return statsSlice, nil
}

// TokenStatsUpdated is the resolver for the tokenStatsUpdated field.
func (r *subscriptionResolver) TokenStatsUpdated(ctx context.Context, tokenID string) (<-chan *model.TokenStats, error) {
	panic(fmt.Errorf("not implemented: TokenStatsUpdated - tokenStatsUpdated"))
}

// TokenMetricsUpdated is the resolver for the tokenMetricsUpdated field.
func (r *subscriptionResolver) TokenMetricsUpdated(ctx context.Context, tokenID string, timeWindow string) (<-chan *model.TokenMetric, error) {
	panic(fmt.Errorf("not implemented: TokenMetricsUpdated - tokenMetricsUpdated"))
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
