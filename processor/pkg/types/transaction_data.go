package types

import (
	"github.com/sjtutwilight/Twilight/common/pkg/types"
)

// TransactionData represents a transaction and its events
type TransactionData struct {
	Transaction types.Transaction
	Events      []types.Event
}
