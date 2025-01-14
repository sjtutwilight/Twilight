package db

import (
	"context"
	"encoding/json"
	"testing"

	"twilight/pkg/types"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("postgres", "host=localhost port=5432 user=twilight password=twilight dbname=twilight_test sslmode=disable")
	require.NoError(t, err)

	// Clean up existing data
	_, err = db.Exec(`
		TRUNCATE TABLE transaction, event, token, twswap_factory, token_metric, pair CASCADE
	`)
	require.NoError(t, err)

	return db
}

func setupTestData(t *testing.T, db *sqlx.DB) (token0ID, token1ID int64) {
	// Insert test tokens
	result := db.QueryRow(`
		INSERT INTO token (chain_id, token_address, symbol, name, decimals, trust_score, create_time, update_time)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id
	`, "31337", "0x1234", "USDC", "USD Coin", 6, 1.0)
	err := result.Scan(&token0ID)
	require.NoError(t, err)

	result = db.QueryRow(`
		INSERT INTO token (chain_id, token_address, symbol, name, decimals, trust_score, create_time, update_time)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id
	`, "31337", "0x5678", "WETH", "Wrapped Ether", 18, 1.0)
	err = result.Scan(&token1ID)
	require.NoError(t, err)

	// Insert token metrics
	_, err = db.Exec(`
		INSERT INTO token_metric (token_id, totalsupply, tradevolumeusd, txcount, totalliquidity, price, update_time)
		VALUES ($1, 0, 0, 0, 0, 0, NOW()), ($2, 0, 0, 0, 0, 0, NOW())
	`, token0ID, token1ID)
	require.NoError(t, err)

	// Insert factory
	_, err = db.Exec(`
		INSERT INTO twswap_factory (chain_id, factory_address, paircount, totalvolumeusd, totalliquidityusd, txcount, update_time)
		VALUES ($1, $2, 0, 0, 0, 0, NOW())
	`, "31337", "0xfactory")
	require.NoError(t, err)

	return token0ID, token1ID
}

func TestProcessPairCreated(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	token0ID, token1ID := setupTestData(t, db)
	processor := &Processor{db: db}

	// Create test event
	eventData := map[string]interface{}{
		"name":            "PairCreated",
		"contractAddress": "0xfactory",
		"args": map[string]interface{}{
			"token0": "0x1234",
			"token1": "0x5678",
			"pair":   "0xpair",
		},
	}
	eventJSON, err := json.Marshal(eventData)
	require.NoError(t, err)

	event := &types.Event{
		ChainID:         "31337",
		EventName:       "PairCreated",
		ContractAddress: "0xfactory",
		Data:            string(eventJSON),
	}

	// Start transaction
	tx, err := db.BeginTxx(context.Background(), nil)
	require.NoError(t, err)
	defer tx.Rollback()

	// Process event
	err = processor.processPairCreated(context.Background(), tx, event)
	require.NoError(t, err)

	// Verify factory update
	var factory types.TWSwapFactory
	err = tx.Get(&factory, "SELECT * FROM twswap_factory WHERE factory_address = $1", "0xfactory")
	require.NoError(t, err)
	assert.Equal(t, 1, factory.PairCount)
	assert.Equal(t, 1, factory.TxCount)

	// Verify pair creation
	var pair types.Pair
	err = tx.Get(&pair, "SELECT * FROM pair WHERE pair_address = $1", "0xpair")
	require.NoError(t, err)
	assert.Equal(t, token0ID, pair.Token0ID)
	assert.Equal(t, token1ID, pair.Token1ID)
	assert.Equal(t, float64(0), pair.Reserve0)
	assert.Equal(t, float64(0), pair.Reserve1)
}

func TestProcessSwap(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	token0ID, token1ID := setupTestData(t, db)
	processor := &Processor{db: db}

	// Create test pair
	_, err := db.Exec(`
		INSERT INTO pair (
			chain_id, pair_address, token0, token1,
			reserve0, reserve1, totalsupply,
			volumetoken0, volumetoken1,
			createdattimestamp, createdatblocknumber,
			update_time
		) VALUES (
			$1, $2, $3, $4,
			1000, 1000, 1000,
			0, 0,
			NOW(), 1,
			NOW()
		)
	`, "31337", "0xpair", token0ID, token1ID)
	require.NoError(t, err)

	// Create test event
	eventData := map[string]interface{}{
		"name":            "Swap",
		"contractAddress": "0xpair",
		"args": map[string]interface{}{
			"amount0In":  "100",
			"amount1In":  "0",
			"amount0Out": "0",
			"amount1Out": "98",
			"sender":     "0xsender",
			"to":         "0xrecipient",
		},
	}
	eventJSON, err := json.Marshal(eventData)
	require.NoError(t, err)

	event := &types.Event{
		ChainID:         "31337",
		EventName:       "Swap",
		ContractAddress: "0xpair",
		Data:            string(eventJSON),
	}

	// Start transaction
	tx, err := db.BeginTxx(context.Background(), nil)
	require.NoError(t, err)
	defer tx.Rollback()

	// Process event
	err = processor.processSwap(context.Background(), tx, event)
	require.NoError(t, err)

	// Verify pair update
	var pair types.Pair
	err = tx.Get(&pair, "SELECT * FROM pair WHERE pair_address = $1", "0xpair")
	require.NoError(t, err)
	assert.Equal(t, float64(100), pair.VolumeToken0)
	assert.Equal(t, float64(98), pair.VolumeToken1)

	// Verify token metrics
	var metric0, metric1 types.TokenMetric
	err = tx.Get(&metric0, "SELECT * FROM token_metric WHERE token_id = $1", token0ID)
	require.NoError(t, err)
	assert.Equal(t, 1, metric0.TxCount)

	err = tx.Get(&metric1, "SELECT * FROM token_metric WHERE token_id = $1", token1ID)
	require.NoError(t, err)
	assert.Equal(t, 1, metric1.TxCount)
}

func TestProcessSync(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	token0ID, token1ID := setupTestData(t, db)
	processor := &Processor{db: db}

	// Create test pair
	_, err := db.Exec(`
		INSERT INTO pair (
			chain_id, pair_address, token0, token1,
			reserve0, reserve1, totalsupply,
			volumetoken0, volumetoken1,
			createdattimestamp, createdatblocknumber,
			update_time
		) VALUES (
			$1, $2, $3, $4,
			1000, 1000, 1000,
			0, 0,
			NOW(), 1,
			NOW()
		)
	`, "31337", "0xpair", token0ID, token1ID)
	require.NoError(t, err)

	// Create test event
	eventData := map[string]interface{}{
		"name":            "Sync",
		"contractAddress": "0xpair",
		"args": map[string]interface{}{
			"reserve0": "1100",
			"reserve1": "900",
		},
	}
	eventJSON, err := json.Marshal(eventData)
	require.NoError(t, err)

	event := &types.Event{
		ChainID:         "31337",
		EventName:       "Sync",
		ContractAddress: "0xpair",
		Data:            string(eventJSON),
	}

	// Start transaction
	tx, err := db.BeginTxx(context.Background(), nil)
	require.NoError(t, err)
	defer tx.Rollback()

	// Process event
	err = processor.processSync(context.Background(), tx, event)
	require.NoError(t, err)

	// Verify pair update
	var pair types.Pair
	err = tx.Get(&pair, "SELECT * FROM pair WHERE pair_address = $1", "0xpair")
	require.NoError(t, err)
	assert.Equal(t, float64(1100), pair.Reserve0)
	assert.Equal(t, float64(900), pair.Reserve1)
}

func TestProcessMintBurn(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	token0ID, token1ID := setupTestData(t, db)
	processor := &Processor{db: db}

	// Create test pair
	_, err := db.Exec(`
		INSERT INTO pair (
			chain_id, pair_address, token0, token1,
			reserve0, reserve1, totalsupply,
			volumetoken0, volumetoken1,
			createdattimestamp, createdatblocknumber,
			update_time
		) VALUES (
			$1, $2, $3, $4,
			1000, 1000, 1000,
			0, 0,
			NOW(), 1,
			NOW()
		)
	`, "31337", "0xpair", token0ID, token1ID)
	require.NoError(t, err)

	// Test Mint
	mintData := map[string]interface{}{
		"name":            "Mint",
		"contractAddress": "0xpair",
		"args": map[string]interface{}{
			"sender":  "0xsender",
			"amount0": "100",
			"amount1": "100",
		},
	}
	mintJSON, err := json.Marshal(mintData)
	require.NoError(t, err)

	mintEvent := &types.Event{
		ChainID:         "31337",
		EventName:       "Mint",
		ContractAddress: "0xpair",
		Data:            string(mintJSON),
	}

	tx, err := db.BeginTxx(context.Background(), nil)
	require.NoError(t, err)
	defer tx.Rollback()

	err = processor.processMint(context.Background(), tx, mintEvent)
	require.NoError(t, err)

	var pair types.Pair
	err = tx.Get(&pair, "SELECT * FROM pair WHERE pair_address = $1", "0xpair")
	require.NoError(t, err)
	assert.Equal(t, float64(1100), pair.Reserve0)
	assert.Equal(t, float64(1100), pair.Reserve1)

	// Test Burn
	burnData := map[string]interface{}{
		"name":            "Burn",
		"contractAddress": "0xpair",
		"args": map[string]interface{}{
			"sender":  "0xsender",
			"amount0": "50",
			"amount1": "50",
			"to":      "0xrecipient",
		},
	}
	burnJSON, err := json.Marshal(burnData)
	require.NoError(t, err)

	burnEvent := &types.Event{
		ChainID:         "31337",
		EventName:       "Burn",
		ContractAddress: "0xpair",
		Data:            string(burnJSON),
	}

	err = processor.processBurn(context.Background(), tx, burnEvent)
	require.NoError(t, err)

	err = tx.Get(&pair, "SELECT * FROM pair WHERE pair_address = $1", "0xpair")
	require.NoError(t, err)
	assert.Equal(t, float64(1050), pair.Reserve0)
	assert.Equal(t, float64(1050), pair.Reserve1)
}
