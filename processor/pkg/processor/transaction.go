package processor

import (
	"context"
	"strings"
	"time"

	"github.com/sjtutwilight/Twilight/common/pkg/types"
	"github.com/sjtutwilight/Twilight/processor/pkg/model"
)

// ProcessTransaction processes a transaction and its events with improved error handling
func (p *Processor) ProcessTransaction(ctx context.Context, tx *types.Transaction, events []types.Event) error {
	// Start a database transaction
	dbTx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return model.NewProcessorError("begin_transaction", err)
	}
	defer dbTx.Rollback()

	// Insert transaction using prepared statement
	var txID int64
	rows, err := p.insertTxStmt.QueryContext(ctx, tx)
	if err != nil {
		return model.NewProcessorError("insert_transaction", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&txID); err != nil {
			return model.NewProcessorError("scan_transaction_id", err)
		}
	}

	// Process events
	for i := range events {
		event := &events[i]
		event.TransactionID = txID
		event.CreateTime = time.Now()

		// Insert event record
		var eventID int64
		eventRows, err := p.insertEventStmt.QueryContext(ctx, event)
		if err != nil {
			return model.NewProcessorError("insert_event", err)
		}

		if eventRows.Next() {
			if err := eventRows.Scan(&eventID); err != nil {
				eventRows.Close()
				return model.NewProcessorError("scan_event_id", err)
			}
		}
		eventRows.Close()

		// Set the event ID
		event.ID = eventID

		// Process transfer events
		if strings.EqualFold(event.EventName, "Transfer") {
			if err := p.transferProcessor.ProcessEvent(ctx, dbTx, event); err != nil {
				return model.NewProcessorError("process_transfer_event", err)
			}
		}
	}

	// Commit transaction
	if err := dbTx.Commit(); err != nil {
		return model.NewProcessorError("commit_transaction", err)
	}

	return nil
}
