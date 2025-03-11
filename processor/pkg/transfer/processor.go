package transfer

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sjtutwilight/Twilight/common/pkg/types"
	"github.com/sjtutwilight/Twilight/processor/pkg/cache"
)

// Processor handles processing of transfer events
type Processor struct {
	db           *sqlx.DB
	cacheManager *cache.Manager
	// Add methods as fields
	insertTransferEvent func(ctx context.Context, tx *sqlx.Tx, event *types.Event, tokenID int64, fromAddr string, toAddr string, valueUSD *big.Float, tokenSymbol string) error
	updateAccountAssets func(ctx context.Context, tx *sqlx.Tx, bizID int64, fromAddr string, toAddr string, value *big.Float, valueUSD *big.Float, assetType string, bizName string) error
	upsertAccountAsset  func(ctx context.Context, tx *sqlx.Tx, accountID int64, assetType string, bizID int64, bizName string, valueDelta float64, valueUSDDelta float64) error
}

// NewProcessor creates a new transfer processor
func NewProcessor(db *sqlx.DB, cacheManager *cache.Manager) *Processor {
	p := &Processor{
		db:           db,
		cacheManager: cacheManager,
	}

	// Initialize method fields
	p.insertTransferEvent = p.insertTransferEventImpl
	p.updateAccountAssets = p.updateAccountAssetsImpl
	p.upsertAccountAsset = p.upsertAccountAssetImpl

	return p
}

// ProcessEvent processes a transfer event
func (p *Processor) ProcessEvent(ctx context.Context, tx *sqlx.Tx, event *types.Event) error {
	// Check if the event is a Transfer event
	if event.EventName != "Transfer" {
		return nil
	}

	// Get the contract address
	contractAddr := strings.ToLower(event.ContractAddress)

	// Parse the event data
	from := strings.ToLower(event.DecodedArgs["from"].(string))
	to := strings.ToLower(event.DecodedArgs["to"].(string))
	valueStr := event.DecodedArgs["value"].(string)

	// Convert value to big.Int
	value := new(big.Int)
	value.SetString(valueStr, 10)

	// Check if this is a token or pair contract
	if p.cacheManager.IsToken(contractAddr) {
		return p.handleTokenTransfer(ctx, tx, contractAddr, from, to, value, event)
	} else if p.cacheManager.IsPair(contractAddr) {
		return p.handleLPTransfer(ctx, tx, contractAddr, from, to, value, event)
	}

	// Unknown contract, log and skip
	log.Printf("Unknown contract address in transfer event: %s", contractAddr)
	return nil
}

// insertTransferEventImpl is the implementation of insertTransferEvent
func (p *Processor) insertTransferEventImpl(
	ctx context.Context,
	tx *sqlx.Tx,
	event *types.Event,
	tokenID int64,
	fromAddr string,
	toAddr string,
	valueUSD *big.Float,
	tokenSymbol string,
) error {
	// Convert valueUSD to string with 4 decimal places
	valueUSDStr := valueUSD.Text('f', 4)
	valueUSDFloat, _ := strconv.ParseFloat(valueUSDStr, 64)

	// Insert transfer event without checking account IDs
	_, err := tx.ExecContext(ctx, `
		INSERT INTO transfer_event (
			token_id, event_id, 
			from_address, to_address, block_timestamp, token_symbol, value_usd
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)`,
		tokenID, event.ID,
		fromAddr, toAddr, event.CreateTime, tokenSymbol, valueUSDFloat,
	)

	if err != nil {
		log.Printf("Error inserting transfer event: %v\nToken ID: %d, Event ID: %d, From: %s, To: %s",
			err, tokenID, event.ID, fromAddr, toAddr)
		return fmt.Errorf("failed to insert transfer event: %w", err)
	}

	return nil
}

// updateAccountAssetsImpl is the implementation of updateAccountAssets
func (p *Processor) updateAccountAssetsImpl(
	ctx context.Context,
	tx *sqlx.Tx,
	bizID int64,
	fromAddr string,
	toAddr string,
	value *big.Float,
	valueUSD *big.Float,
	assetType string,
	bizName string,
) error {
	// Get account metadata
	fromAccount := p.cacheManager.GetAccountMetadata(fromAddr)
	toAccount := p.cacheManager.GetAccountMetadata(toAddr)

	// Update sender's assets if it's a tracked account
	if fromAccount != nil && fromAddr != "0x0000000000000000000000000000000000000000" {
		// Negate value for sender
		negValue := new(big.Float).Neg(value)
		negValueStr := negValue.Text('f', 4) // Use 4 decimal places for consistency
		negValueFloat, _ := strconv.ParseFloat(negValueStr, 64)

		// Negate USD value for sender
		negValueUSD := new(big.Float).Neg(valueUSD)
		negValueUSDStr := negValueUSD.Text('f', 4) // Use 4 decimal places for consistency
		negValueUSDFloat, _ := strconv.ParseFloat(negValueUSDStr, 64)

		// Get account ID as int64
		fromAccountID, err := fromAccount.GetIDAsInt64()
		if err != nil {
			return fmt.Errorf("failed to parse from account ID: %w", err)
		}

		// Update or insert sender's asset
		if err := p.upsertAccountAsset(ctx, tx, fromAccountID, assetType, bizID, bizName, negValueFloat, negValueUSDFloat); err != nil {
			log.Printf("Error updating sender's asset: %v\nAccount: %s, Asset Type: %s, Biz ID: %d, Biz Name: %s, Value: %f, Value USD: %f",
				err, fromAddr, assetType, bizID, bizName, negValueFloat, negValueUSDFloat)
			return err
		}
		log.Printf("Updated sender's asset: Account: %s, Asset Type: %s, Biz Name: %s, Value: %f, Value USD: %f",
			fromAddr, assetType, bizName, negValueFloat, negValueUSDFloat)
	} else if fromAddr != "0x0000000000000000000000000000000000000000" {
		// Log that we're skipping an untracked sender
		log.Printf("Skipping asset update for untracked sender: %s", fromAddr)
	}

	// Update receiver's assets if it's a tracked account
	if toAccount != nil && toAddr != "0x0000000000000000000000000000000000000000" {
		// Convert value to float for receiver with 4 decimal places
		valueStr := value.Text('f', 4)
		valueFloat, _ := strconv.ParseFloat(valueStr, 64)

		// Convert USD value to float for receiver with 4 decimal places
		valueUSDStr := valueUSD.Text('f', 4)
		valueUSDFloat, _ := strconv.ParseFloat(valueUSDStr, 64)

		// Get account ID as int64
		toAccountID, err := toAccount.GetIDAsInt64()
		if err != nil {
			return fmt.Errorf("failed to parse to account ID: %w", err)
		}

		// Update or insert receiver's asset
		if err := p.upsertAccountAsset(ctx, tx, toAccountID, assetType, bizID, bizName, valueFloat, valueUSDFloat); err != nil {
			log.Printf("Error updating receiver's asset: %v\nAccount: %s, Asset Type: %s, Biz ID: %d, Biz Name: %s, Value: %f, Value USD: %f",
				err, toAddr, assetType, bizID, bizName, valueFloat, valueUSDFloat)
			return err
		}
		log.Printf("Updated receiver's asset: Account: %s, Asset Type: %s, Biz Name: %s, Value: %f, Value USD: %f",
			toAddr, assetType, bizName, valueFloat, valueUSDFloat)
	} else if toAddr != "0x0000000000000000000000000000000000000000" {
		// Log that we're skipping an untracked receiver
		log.Printf("Skipping asset update for untracked receiver: %s", toAddr)
	}

	return nil
}

// upsertAccountAssetImpl is the implementation of upsertAccountAsset
func (p *Processor) upsertAccountAssetImpl(
	ctx context.Context,
	tx *sqlx.Tx,
	accountID int64,
	assetType string,
	bizID int64,
	bizName string,
	valueDelta float64,
	valueUSDDelta float64,
) error {
	now := time.Now()

	// Try to update existing asset
	result, err := tx.ExecContext(ctx, `
		UPDATE account_asset 
		SET value = ROUND(CAST(value + $1 AS DECIMAL(24,4)), 4), 
		    update_time = $2
		WHERE account_id = $3 AND asset_type = $4 AND biz_id = $5
	`, valueDelta, now, accountID, assetType, bizID)

	if err != nil {
		return fmt.Errorf("failed to update account asset: %w", err)
	}

	// Check if the update affected any rows
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// If no rows were affected, insert a new asset
	if rowsAffected == 0 {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO account_asset (
				account_id, asset_type, biz_id, biz_name, 
				value, create_time, update_time
			) VALUES (
				$1, $2, $3, $4, 
				ROUND(CAST($5 AS DECIMAL(24,4)), 4), 
				$6, $7
			)`,
			accountID, assetType, bizID, bizName,
			valueDelta, now, now,
		)

		if err != nil {
			return fmt.Errorf("failed to insert account asset: %w", err)
		}
	}

	return nil
}
