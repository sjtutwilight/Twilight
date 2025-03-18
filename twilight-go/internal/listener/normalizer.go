package listener

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/sjtutwilight/twilight-go/pkg/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// NormalizedData represents standardized data format
type NormalizedData struct {
	ChainEvent *types.ChainEvent
}

// Normalizer interface defines methods for data normalization
type Normalizer interface {
	Normalize(Data) (NormalizedData, error)
}

// DexTransactionNormalizer implements Normalizer for DEX transactions
type DexTransactionNormalizer struct {
	contracts map[common.Address]*Contract
}

// NewDexTransactionNormalizer creates a new DEX transaction normalizer
func NewDexTransactionNormalizer() *DexTransactionNormalizer {
	return &DexTransactionNormalizer{
		contracts: make(map[common.Address]*Contract),
	}
}

// AddContract adds a contract to the normalizer
func (n *DexTransactionNormalizer) AddContract(address common.Address, contract *Contract) {
	n.contracts[address] = contract
}

// Normalize converts raw blockchain data to standardized format
func (n *DexTransactionNormalizer) Normalize(data Data) (NormalizedData, error) {
	block := data.Block
	if block == nil {
		return NormalizedData{}, fmt.Errorf("block data is nil")
	}

	// Process each transaction in the block
	for _, tx := range block.Transactions() {
		receipt, exists := data.Receipts[tx.Hash()]
		if !exists {
			continue
		}

		// Create a chain event for the transaction
		chainEvent := n.createChainEvent(block, tx, receipt)

		// 初始化空的事件数组，确保不为null
		chainEvent.Events = []struct {
			EventName       string                 `json:"eventName"`
			ContractAddress string                 `json:"contractAddress"`
			LogIndex        int                    `json:"logIndex"`
			BlockNumber     int64                  `json:"blockNumber"`
			Topics          []string               `json:"topics"`
			EventData       string                 `json:"eventData"`
			DecodedArgs     map[string]interface{} `json:"decodedArgs"`
		}{}

		// Process logs
		for _, logEntry := range receipt.Logs {
			if contract, exists := n.contracts[logEntry.Address]; exists {
				// Check if this log matches any of our registered event signatures
				if contract.IsInterestedInLog(logEntry) {
					eventName := contract.GetEventName(logEntry.Topics[0])

					// Add the event to the chain event
					n.addEventToChainEvent(chainEvent, logEntry, eventName)
				}
			}
		}

		// 即使没有找到事件，也返回带有空事件数组的数据
		return NormalizedData{
			ChainEvent: chainEvent,
		}, nil
	}

	return NormalizedData{}, fmt.Errorf("no transactions to normalize")
}

// createChainEvent creates a new chain event from block, transaction, and receipt
func (n *DexTransactionNormalizer) createChainEvent(block *ethtypes.Block, tx *ethtypes.Transaction, receipt *ethtypes.Receipt) *types.ChainEvent {
	signer := ethtypes.LatestSignerForChainID(tx.ChainId())
	from, err := ethtypes.Sender(signer, tx)
	if err != nil {
		from = common.Address{}
	}

	var to string
	if tx.To() != nil {
		to = tx.To().Hex()
	}

	event := &types.ChainEvent{}
	event.Transaction.BlockNumber = block.Number().Int64()
	event.Transaction.BlockHash = block.Hash().Hex()
	event.Transaction.Timestamp = time.Now().Unix()
	event.Transaction.TransactionHash = tx.Hash().Hex()
	event.Transaction.TransactionIndex = int(receipt.TransactionIndex)

	if receipt.Status == 1 {
		event.Transaction.TransactionStatus = "success"
	} else {
		event.Transaction.TransactionStatus = "failed"
	}

	event.Transaction.GasUsed = int64(receipt.GasUsed)
	event.Transaction.GasPrice = tx.GasPrice().String()
	event.Transaction.Nonce = tx.Nonce()
	event.Transaction.FromAddress = from.Hex()
	event.Transaction.ToAddress = to
	event.Transaction.TransactionValue = tx.Value().String()
	event.Transaction.InputData = "0x" + hex.EncodeToString(tx.Data())
	event.Transaction.ChainID = tx.ChainId().String()

	return event
}

// addEventToChainEvent adds an event to a chain event
func (n *DexTransactionNormalizer) addEventToChainEvent(chainEvent *types.ChainEvent, log *ethtypes.Log, eventName string) {
	eventData := struct {
		EventName       string                 `json:"eventName"`
		ContractAddress string                 `json:"contractAddress"`
		LogIndex        int                    `json:"logIndex"`
		BlockNumber     int64                  `json:"blockNumber"`
		Topics          []string               `json:"topics"`
		EventData       string                 `json:"eventData"`
		DecodedArgs     map[string]interface{} `json:"decodedArgs"`
	}{
		EventName:       eventName,
		ContractAddress: log.Address.Hex(),
		LogIndex:        int(log.Index),
		BlockNumber:     int64(log.BlockNumber),
		Topics:          make([]string, len(log.Topics)),
		EventData:       "0x" + hex.EncodeToString(log.Data),
		DecodedArgs:     make(map[string]interface{}),
	}

	for i, topic := range log.Topics {
		eventData.Topics[i] = topic.Hex()
	}

	// Decode event arguments based on event type
	switch eventName {
	case "Mint":
		if len(log.Topics) >= 2 && len(log.Data) >= 64 {
			sender := common.HexToAddress(log.Topics[1].Hex())
			amount0 := new(big.Int).SetBytes(log.Data[:32])
			amount1 := new(big.Int).SetBytes(log.Data[32:64])

			eventData.DecodedArgs = map[string]interface{}{
				"sender":  sender.Hex(),
				"amount0": amount0.String(),
				"amount1": amount1.String(),
			}
		}
	case "Burn":
		if len(log.Topics) >= 2 && len(log.Data) >= 64 {
			sender := common.HexToAddress(log.Topics[1].Hex())
			amount0 := new(big.Int).SetBytes(log.Data[:32])
			amount1 := new(big.Int).SetBytes(log.Data[32:64])

			eventData.DecodedArgs = map[string]interface{}{
				"sender":  sender.Hex(),
				"amount0": amount0.String(),
				"amount1": amount1.String(),
			}
		}
	case "Swap":
		if len(log.Data) >= 128 {
			amount0In := new(big.Int).SetBytes(log.Data[:32])
			amount1In := new(big.Int).SetBytes(log.Data[32:64])
			amount0Out := new(big.Int).SetBytes(log.Data[64:96])
			amount1Out := new(big.Int).SetBytes(log.Data[96:128])

			eventData.DecodedArgs = map[string]interface{}{
				"amount0In":  amount0In.String(),
				"amount1In":  amount1In.String(),
				"amount0Out": amount0Out.String(),
				"amount1Out": amount1Out.String(),
			}
		}
	case "Sync":
		if len(log.Data) >= 64 {
			reserve0 := new(big.Int).SetBytes(log.Data[:32])
			reserve1 := new(big.Int).SetBytes(log.Data[32:64])

			eventData.DecodedArgs = map[string]interface{}{
				"reserve0": reserve0.String(),
				"reserve1": reserve1.String(),
			}
		}
	case "Transfer":
		if len(log.Topics) >= 3 && len(log.Data) >= 32 {
			from := common.HexToAddress(log.Topics[1].Hex())
			to := common.HexToAddress(log.Topics[2].Hex())
			value := new(big.Int).SetBytes(log.Data[:32])

			eventData.DecodedArgs = map[string]interface{}{
				"from":  from.Hex(),
				"to":    to.Hex(),
				"value": value.String(),
			}
		}
	}

	chainEvent.Events = append(chainEvent.Events, eventData)
}
