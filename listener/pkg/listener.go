package chain

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"twilight/pkg/handlers"
	"twilight/pkg/types"
)

// Use the EventHandler type from handlers package
type EventHandler = handlers.EventHandler

type Contract struct {
	Address  common.Address
	Name     string
	ABI      *abi.ABI
	Handlers map[string]EventHandler
}

type EventListener struct {
	client    *ethclient.Client
	chainID   string
	contracts map[string]*Contract
	eventChan chan types.TransactionEvent
}

func NewEventListener(client *ethclient.Client, chainID string) *EventListener {
	return &EventListener{
		client:    client,
		chainID:   chainID,
		contracts: make(map[string]*Contract),
		eventChan: make(chan types.TransactionEvent),
	}
}

func (l *EventListener) AddContract(address string, name string, contractABI interface{}, handlers map[string]EventHandler) {
	addr := common.HexToAddress(address)
	var parsedABI *abi.ABI
	if contractABI != nil {
		if abiObj, ok := contractABI.(*abi.ABI); ok {
			parsedABI = abiObj
		}
	}

	l.contracts[address] = &Contract{
		Address:  addr,
		Name:     name,
		ABI:      parsedABI,
		Handlers: handlers,
	}
}

func (l *EventListener) GetEventChannel() <-chan types.TransactionEvent {
	return l.eventChan
}

func (l *EventListener) Start(ctx context.Context) (<-chan types.TransactionEvent, <-chan error) {
	errorChan := make(chan error)

	go func() {
		defer close(l.eventChan)
		defer close(errorChan)

		latestBlock, err := l.client.BlockNumber(ctx)
		if err != nil {
			errorChan <- fmt.Errorf("failed to get latest block: %w", err)
			return
		}

		log.Printf("Starting from block %d", latestBlock)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				currentBlock, err := l.client.BlockNumber(ctx)
				if err != nil {
					errorChan <- fmt.Errorf("failed to get current block: %w", err)
					continue
				}

				if currentBlock > latestBlock {
					if err := l.processNewBlocks(ctx, latestBlock+1, currentBlock); err != nil {
						errorChan <- fmt.Errorf("failed to process blocks: %w", err)
					}
					latestBlock = currentBlock
				}

				time.Sleep(time.Second)
			}
		}
	}()

	return l.eventChan, errorChan
}

func (l *EventListener) processNewBlocks(ctx context.Context, fromBlock, toBlock uint64) error {
	for blockNum := fromBlock; blockNum <= toBlock; blockNum++ {
		block, err := l.client.BlockByNumber(ctx, new(big.Int).SetUint64(blockNum))
		if err != nil {
			return fmt.Errorf("failed to get block %d: %w", blockNum, err)
		}

		for _, tx := range block.Transactions() {
			receipt, err := l.client.TransactionReceipt(ctx, tx.Hash())
			if err != nil {
				log.Printf("Failed to get receipt for tx %s: %v", tx.Hash().Hex(), err)
				continue
			}

			kafkaTx, err := types.NewKafkaTransaction(l.chainID, receipt, tx)
			if err != nil {
				log.Printf("Failed to create Kafka transaction: %v", err)
				continue
			}

			var events []types.KafkaEvent
			for _, eventLog := range receipt.Logs {
				if contract, ok := l.contracts[eventLog.Address.Hex()]; ok {
					if handler, ok := contract.Handlers[eventLog.Topics[0].Hex()]; ok {
						decodedData, err := handler(*eventLog)
						if err != nil {
							log.Printf("Failed to handle event: %v", err)
							continue
						}

						decodedJSON, err := json.Marshal(decodedData)
						if err != nil {
							log.Printf("Failed to marshal decoded data: %v", err)
							continue
						}

						event := types.KafkaEvent{
							ChainID:         l.chainID,
							EventName:       contract.Name,
							ContractAddress: eventLog.Address.Hex(),
							LogIndex:        int(eventLog.Index),
							BlockNumber:     int64(eventLog.BlockNumber),
							Topics:          make([]string, len(eventLog.Topics)),
							Data:            hex.EncodeToString(eventLog.Data),
							DecodedArgs:     string(decodedJSON),
							CreateTime:      time.Now(),
						}

						for i, topic := range eventLog.Topics {
							event.Topics[i] = topic.Hex()
						}

						events = append(events, event)
					}
				}
			}

			if len(events) > 0 {
				l.eventChan <- types.TransactionEvent{
					Transaction: *kafkaTx,
					Events:      events,
				}
				log.Printf("Processed transaction %s with %d events", tx.Hash().Hex(), len(events))
			}
		}
	}

	return nil
}
