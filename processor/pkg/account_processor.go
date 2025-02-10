package db

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/jmoiron/sqlx"
)

// AccountProcessor handles processing of account-related events
type AccountProcessor struct {
	db             *sqlx.DB
	accountMap     map[string]int64 // address -> account_id
	tokenMap       map[string]int64 // token_address -> token_id
	pairMap        map[string]int64 // address -> pair_id
	liquidityDelta *big.Float       // Changed from *big.Int to *big.Float
}

// NewAccountProcessor creates a new account processor
func NewAccountProcessor(db *sqlx.DB) *AccountProcessor {
	return &AccountProcessor{
		db:             db,
		accountMap:     make(map[string]int64),
		tokenMap:       make(map[string]int64),
		pairMap:        make(map[string]int64),
		liquidityDelta: new(big.Float),
	}
}

// Initialize loads account and pair mappings from database
func (p *AccountProcessor) Initialize(ctx context.Context) error {
	// Load account mappings
	accounts := []struct {
		ID      int64  `db:"id"`
		Address string `db:"address"`
	}{}

	if err := p.db.SelectContext(ctx, &accounts, `
		SELECT id, address FROM account
	`); err != nil {
		return fmt.Errorf("failed to load accounts: %w", err)
	}

	for _, acc := range accounts {
		p.accountMap[strings.ToLower(acc.Address)] = acc.ID
	}

	// Load token mappings
	tokens := []struct {
		ID      int64  `db:"id"`
		Address string `db:"token_address"`
	}{}

	if err := p.db.SelectContext(ctx, &tokens, `
		SELECT id, token_address FROM token
	`); err != nil {
		return fmt.Errorf("failed to load tokens: %w", err)
	}

	for _, token := range tokens {
		p.tokenMap[strings.ToLower(token.Address)] = token.ID
	}

	// Load pair mappings with token information
	pairs := []struct {
		ID          int64  `db:"id"`
		PairAddress string `db:"pair_address"`
		Token0ID    int64  `db:"token0_id"`
		Token1ID    int64  `db:"token1_id"`
	}{}

	if err := p.db.SelectContext(ctx, &pairs, `
		SELECT id, pair_address, token0_id, token1_id 
		FROM twswap_pair
	`); err != nil {
		return fmt.Errorf("failed to load pairs: %w", err)
	}

	for _, pair := range pairs {
		p.pairMap[strings.ToLower(pair.PairAddress)] = pair.ID
	}

	return nil
}

// ProcessEvent processes an event based on its type
func (p *AccountProcessor) ProcessEvent(ctx context.Context, tx *sqlx.Tx, event map[string]interface{}, accountId int64) error {
	eventName, ok := event["eventName"].(string)
	if !ok {
		return fmt.Errorf("missing event name")
	}

	contractAddr := strings.ToLower(event["contractAddress"].(string))
	decodedArgs := event["decodedArgs"].(map[string]interface{})
	switch eventName {
	case "Transfer":
		return p.handleTransfer(ctx, tx, contractAddr, decodedArgs, accountId)
	case "Sync":
		return p.handleSync(ctx, tx, contractAddr, decodedArgs, accountId)
	}

	return nil
}

// convertToDecimal converts a string representation of a big number to decimal format
// by dividing by 10^decimals
func convertToDecimal(value string, decimals int) string {
	// Parse the string value into a big.Int
	n := new(big.Int)
	n.SetString(value, 10)

	// Create a big.Float from the big.Int
	f := new(big.Float).SetInt(n)

	// Calculate divisor (10^decimals)
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))

	// Perform division
	result := new(big.Float).Quo(f, divisor)

	// Convert to string with high precision
	return result.Text('f', 18)
}

// handleTransfer processes Transfer events
func (p *AccountProcessor) handleTransfer(ctx context.Context, tx *sqlx.Tx, contractAddr string, args map[string]interface{}, accountId int64) error {
	from := strings.ToLower(args["from"].(string))
	to := strings.ToLower(args["to"].(string))
	valueStr := args["value"].(string)

	// Convert value to decimal format (assuming 18 decimals for most tokens)
	decimalValue := convertToDecimal(valueStr, 18)
	value := new(big.Float)
	value.SetString(decimalValue)

	// Check if this is a pair contract
	if pairID, isPair := p.pairMap[contractAddr]; isPair {
		// Handle LP token transfer
		if from == "0x0000000000000000000000000000000000000000" {
			// Mint LP tokens
			p.liquidityDelta.Set(value)
		} else if to == "0x0000000000000000000000000000000000000000" {
			// Burn LP tokens

			negValue := new(big.Float).Neg(value)
			// Update LP token balance
			if err := p.updateLPBalance(ctx, tx, accountId, pairID, negValue); err != nil {
				return err

			}
		}
		return nil
	}

	// Handle regular token transfer
	if fromID, exists := p.accountMap[from]; exists {
		negValue := new(big.Float).Neg(value)
		if err := p.updateTokenBalance(ctx, tx, fromID, contractAddr, negValue); err != nil {
			return err
		}
	}

	if toID, exists := p.accountMap[to]; exists {
		if err := p.updateTokenBalance(ctx, tx, toID, contractAddr, value); err != nil {
			return err
		}
	}

	return nil
}

// handleSync processes Sync events
func (p *AccountProcessor) handleSync(ctx context.Context, tx *sqlx.Tx, contractAddr string, args map[string]interface{}, accountId int64) error {
	if p.liquidityDelta.Cmp(big.NewFloat(0)) == 0 {
		return nil
	}

	pairID, exists := p.pairMap[contractAddr]
	if !exists {
		return fmt.Errorf("unknown pair address: %s", contractAddr)
	}

	// Convert reserves to decimal format
	reserve0 := convertToDecimal(args["reserve0"].(string), 18)
	reserve1 := convertToDecimal(args["reserve1"].(string), 18)

	// Calculate price as reserve ratio
	r0 := new(big.Float)
	r1 := new(big.Float)
	r0.SetString(reserve0)
	r1.SetString(reserve1)
	price := new(big.Float).Quo(r0, r1)

	// Find the account that received the LP tokens

	// Update the LP token balance and average price
	if err := p.updateLPBalanceAndPrice(ctx, tx, accountId, pairID, p.liquidityDelta, price); err != nil {
		return err
	}

	// Reset liquidityDelta
	p.liquidityDelta = big.NewFloat(0)

	return nil
}

// updateTokenBalance updates the token balance for an account
func (p *AccountProcessor) updateTokenBalance(ctx context.Context, tx *sqlx.Tx, accountID int64, tokenAddr string, delta *big.Float) error {
	// Convert delta to decimal string
	deltaStr := delta.Text('f', 18)

	// Get token ID from tokenMap
	tokenID, exists := p.tokenMap[strings.ToLower(tokenAddr)]
	if !exists {
		return fmt.Errorf("unknown token address: %s", tokenAddr)
	}

	_, err := tx.ExecContext(ctx, `
		UPDATE account_asset 
		SET balance = balance + $1, update_time = NOW()
		WHERE account_id = $2 AND asset_type = 'token_holding' AND bizId = $3
	`, deltaStr, accountID, tokenID)

	return err
}

// updateLPBalance updates the LP token balance for an account
func (p *AccountProcessor) updateLPBalance(ctx context.Context, tx *sqlx.Tx, accountID int64, pairID int64, delta *big.Float) error {
	deltaStr := delta.Text('f', 18)

	_, err := tx.ExecContext(ctx, `
		UPDATE account_asset 
		SET balance = balance + $1, update_time = NOW()
		WHERE account_id = $2 AND asset_type = 'lp' AND bizId = $3
	`, deltaStr, accountID, pairID)

	return err
}

// updateLPBalanceAndPrice updates both LP token balance and average price
func (p *AccountProcessor) updateLPBalanceAndPrice(ctx context.Context, tx *sqlx.Tx, accountId int64, pairID int64, delta *big.Float, newPrice *big.Float) error {
	// Get current balance and average price
	var currentBalance float64
	var currentAvgPrice float64
	err := tx.QueryRowContext(ctx, `
		SELECT balance, 
			   COALESCE((extension_info->>'average_price')::float, 0) as avg_price
		FROM account_asset 
		WHERE account_id = $1 AND asset_type = 'lp' AND bizId = $2
	`, accountId, pairID).Scan(&currentBalance, &currentAvgPrice)
	if err != nil {
		return fmt.Errorf("failed to get current LP data: %w", err)
	}

	// Convert delta to float64
	deltaFloat, _ := delta.Float64()
	newPriceFloat, _ := newPrice.Float64()

	// Calculate new average price
	newBalance := currentBalance + deltaFloat
	newAvgPrice := ((currentBalance * currentAvgPrice) + (deltaFloat * newPriceFloat)) / newBalance

	// Update both balance and average price
	_, err = tx.ExecContext(ctx, `
		UPDATE account_asset 
		SET balance = $1, 
			extension_info = jsonb_set(
				COALESCE(extension_info, '{}'::jsonb),
				'{average_price}',
				$2::text::jsonb
			),
			update_time = NOW()
		WHERE account_id = $3 AND asset_type = 'lp' AND bizId = $4
	`, newBalance, newAvgPrice, accountId, pairID)

	return err
}
