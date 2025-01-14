package types

import "time"

// Transaction represents a blockchain transaction
type Transaction struct {
	ID              int64     `db:"id"`
	ChainID         string    `db:"chain_id"`
	TransactionHash string    `db:"transaction_hash"`
	BlockNumber     int64     `db:"block_number"`
	BlockTimestamp  time.Time `db:"block_timestamp"`
	FromAddress     string    `db:"from_address"`
	ToAddress       string    `db:"to_address"`
	MethodName      string    `db:"method_name"`
	Status          int       `db:"transaction_status"`
	GasUsed         uint64    `db:"gas_used"`
	InputData       string    `db:"input_data"`
	CreateTime      time.Time `db:"create_time"`
}

// Event represents a blockchain event
type Event struct {
	ID              int64     `db:"id"`
	TransactionID   int64     `db:"transaction_id"`
	ChainID         string    `db:"chain_id"`
	EventName       string    `db:"event_name"`
	ContractAddress string    `db:"contract_address"`
	LogIndex        int       `db:"log_index"`
	EventData       string    `db:"event_data"` // JSON encoded event data
	BlockNumber     int64     `db:"block_number"`
	CreateTime      time.Time `db:"create_time"`
}
