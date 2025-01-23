package db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
)

type DeploymentConfig struct {
	Factory string `json:"factory"`
	Router  string `json:"router"`
	Tokens  []struct {
		Address  string `json:"address"`
		Symbol   string `json:"symbol"`
		Decimals string `json:"decimals"`
	} `json:"tokens"`
	Pairs []struct {
		Token0  string `json:"token0"`
		Token1  string `json:"token1"`
		Address string `json:"address"`
	} `json:"pairs"`
}

// InitializeFromDeployment reads deployment.json and initializes the database
func (p *Processor) InitializeFromDeployment(workspacePath string) error {
	// Read deployment.json
	deploymentPath := filepath.Join(workspacePath, "deployment.json")
	data, err := ioutil.ReadFile(deploymentPath)
	if err != nil {
		return &ProcessorError{Op: "read_deployment", Err: err}
	}

	var config DeploymentConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return &ProcessorError{Op: "parse_deployment", Err: err}
	}

	// Start transaction
	tx, err := p.db.Beginx()
	if err != nil {
		return &ProcessorError{Op: "begin_transaction", Err: err}
	}
	defer tx.Rollback()

	// Initialize factory
	if err := p.initializeFactory(tx, config.Factory); err != nil {
		return err
	}

	// Initialize tokens
	tokenMap, err := p.initializeTokens(tx, config.Tokens)
	if err != nil {
		return err
	}

	// Initialize pairs
	if err := p.initializePairs(tx, config.Pairs, tokenMap); err != nil {
		return err
	}

	return tx.Commit()
}

func (p *Processor) initializeFactory(tx *sqlx.Tx, factoryAddress string) error {
	_, err := tx.Exec(`
		INSERT INTO twswap_factory (
			chain_id, factory_address, pair_count, 
			volume_usd, liquidity_usd, txcnt, time_window, end_time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (chain_id, factory_address, time_window, end_time) DO NOTHING`,
		"local", factoryAddress, 0, 0, 0, 0, "total", time.Now())

	if err != nil {
		return &ProcessorError{Op: "insert_factory", Err: err}
	}
	return nil
}

func (p *Processor) initializeTokens(tx *sqlx.Tx, tokens []struct {
	Address  string `json:"address"`
	Symbol   string `json:"symbol"`
	Decimals string `json:"decimals"`
}) (map[string]int64, error) {
	tokenMap := make(map[string]int64)

	for _, token := range tokens {
		var tokenID int64
		err := tx.QueryRow(`
			INSERT INTO token (
				chain_id, token_address, token_symbol, 
				token_name, token_decimals, create_time, update_time
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (chain_id, token_address) 
			DO UPDATE SET update_time = EXCLUDED.update_time
			RETURNING id`,
			"local", token.Address, token.Symbol,
			token.Symbol, token.Decimals, time.Now(), time.Now(),
		).Scan(&tokenID)

		if err != nil {
			return nil, &ProcessorError{Op: "insert_token", Err: err}
		}
		tokenMap[token.Address] = tokenID
	}

	return tokenMap, nil
}

func (p *Processor) initializePairs(tx *sqlx.Tx, pairs []struct {
	Token0  string `json:"token0"`
	Token1  string `json:"token1"`
	Address string `json:"address"`
}, tokenMap map[string]int64) error {
	for _, pair := range pairs {
		token0ID, exists := tokenMap[pair.Token0]
		if !exists {
			return &ProcessorError{Op: "get_token0", Err: fmt.Errorf("token0 not found: %s", pair.Token0)}
		}
		token1ID, exists := tokenMap[pair.Token1]
		if !exists {
			return &ProcessorError{Op: "get_token1", Err: fmt.Errorf("token1 not found: %s", pair.Token1)}
		}

		// Insert pair
		_, err := tx.Exec(`
			INSERT INTO twswap_pair (
				chain_id, pair_address, token0_id, token1_id,
				fee_tier, created_at_timestamp, created_at_block_number
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (chain_id, pair_address, fee_tier) 
			DO UPDATE SET 
				token0_id = EXCLUDED.token0_id,
				token1_id = EXCLUDED.token1_id`,
			"local", pair.Address, token0ID, token1ID,
			"0.003", time.Now(), 0)

		if err != nil {
			return &ProcessorError{Op: "insert_pair", Err: err}
		}
	}
	return nil
}
