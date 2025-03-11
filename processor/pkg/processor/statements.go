package processor

import (
	"github.com/sjtutwilight/Twilight/processor/pkg/model"
)

// prepareStatements prepares SQL statements for reuse
func (p *Processor) prepareStatements() error {
	var err error

	// Prepare transaction insert statement
	p.insertTxStmt, err = p.db.PrepareNamed(`
		INSERT INTO transaction (
			chain_id, transaction_hash, block_number, block_timestamp, 
			from_address, to_address, method_name, transaction_status, 
			gas_used, input_data, create_time
		) VALUES (
			:chain_id, :transaction_hash, :block_number, :block_timestamp,
			:from_address, :to_address, :method_name, :transaction_status,
			:gas_used, :input_data, :create_time
		) RETURNING id`)
	if err != nil {
		return model.NewProcessorError("prepare_tx_stmt", err)
	}

	// Prepare event insert statement
	p.insertEventStmt, err = p.db.PrepareNamed(`
		INSERT INTO event (
			transaction_id, chain_id, event_name, contract_address,
			log_index, event_data, create_time, block_number
		) VALUES (
			:transaction_id, :chain_id, :event_name, :contract_address,
			:log_index, :event_data, :create_time, :block_number
		) RETURNING id`)
	if err != nil {
		return model.NewProcessorError("prepare_event_stmt", err)
	}

	return nil
}
