package listener

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/viper"
)

// Data represents blockchain data
type Data struct {
	Block    *types.Block
	Receipts map[common.Hash]*types.Receipt
}

// DataSource interface defines methods for data sources
type DataSource interface {
	Start(context.Context) error
	Stop() error
}

// ConnectionMethod interface defines methods for connection strategies
type ConnectionMethod interface {
	Connect() error
	Fetch() (Data, error)
	Close() error
}

// LocalNodeSource implements DataSource for a local Ethereum node
type LocalNodeSource struct {
	client           *ethclient.Client
	connectionMethod ConnectionMethod
	lastBlockNumber  uint64
	mu               sync.Mutex
	stopChan         chan struct{}
	dataChan         chan<- Data
}

// NewLocalNodeSource creates a new local node data source
func NewLocalNodeSource(nodeURL string) (*LocalNodeSource, error) {
	client, err := ethclient.Dial(nodeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	source := &LocalNodeSource{
		client:   client,
		stopChan: make(chan struct{}),
	}

	// Create and set the connection method (polling by default)
	pollingMethod := NewPollingMethod(client)
	source.connectionMethod = pollingMethod

	return source, nil
}

// SetDataChannel sets the data channel for the source
func (s *LocalNodeSource) SetDataChannel(dataChan chan<- Data) {
	s.dataChan = dataChan
}

// Start begins the data source operation
func (s *LocalNodeSource) Start(ctx context.Context) error {
	if s.dataChan == nil {
		return fmt.Errorf("data channel not set")
	}

	if err := s.connectionMethod.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	pollingInterval := time.Duration(viper.GetInt("chain.polling.interval_ms")) * time.Millisecond
	if pollingInterval <= 0 {
		pollingInterval = 1 * time.Second // Default polling interval
	}

	ticker := time.NewTicker(pollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.stopChan:
			return nil
		case <-ticker.C:
			data, err := s.connectionMethod.Fetch()
			if err != nil {
				log.Printf("Error fetching data: %v", err)
				continue
			}

			// Send the data to the channel
			select {
			case s.dataChan <- data:
				log.Printf("Fetched block %d with %d transactions", data.Block.NumberU64(), len(data.Block.Transactions()))
			default:
				log.Printf("Data channel full, dropping block %d", data.Block.NumberU64())
			}
		}
	}
}

// Stop halts the data source operation
func (s *LocalNodeSource) Stop() error {
	close(s.stopChan)
	return s.connectionMethod.Close()
}

// PollingMethod implements ConnectionMethod using polling
type PollingMethod struct {
	client           *ethclient.Client
	lastBlockNumber  uint64
	maxBlocksPerPoll int
	mu               sync.Mutex
}

// NewPollingMethod creates a new polling connection method
func NewPollingMethod(client *ethclient.Client) *PollingMethod {
	maxBlocksPerPoll := viper.GetInt("chain.polling.max_blocks_per_poll")
	if maxBlocksPerPoll <= 0 {
		maxBlocksPerPoll = 100 // Default max blocks per poll
	}

	return &PollingMethod{
		client:           client,
		maxBlocksPerPoll: maxBlocksPerPoll,
	}
}

// Connect initializes the connection
func (p *PollingMethod) Connect() error {
	// For polling, we just need to ensure we can connect to the client
	_, err := p.client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}
	return nil
}

// Fetch retrieves new blocks
func (p *PollingMethod) Fetch() (Data, error) {
	ctx := context.Background()

	p.mu.Lock()
	lastProcessed := p.lastBlockNumber
	p.mu.Unlock()

	currentBlock, err := p.client.BlockNumber(ctx)
	if err != nil {
		return Data{}, fmt.Errorf("failed to get current block number: %w", err)
	}

	// Ensure we don't process more than maxBlocksPerPoll
	endBlock := currentBlock
	if endBlock > lastProcessed+uint64(p.maxBlocksPerPoll) {
		endBlock = lastProcessed + uint64(p.maxBlocksPerPoll)
	}

	if lastProcessed >= endBlock {
		return Data{}, fmt.Errorf("no new blocks to process")
	}

	// For simplicity, we'll just fetch the latest block
	// In a real implementation, we would fetch all blocks from lastProcessed+1 to endBlock
	block, err := p.client.BlockByNumber(ctx, big.NewInt(int64(endBlock)))
	if err != nil {
		return Data{}, fmt.Errorf("failed to get block %d: %w", endBlock, err)
	}

	// Fetch receipts for all transactions in the block
	receipts := make(map[common.Hash]*types.Receipt)
	for _, tx := range block.Transactions() {
		receipt, err := p.client.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			log.Printf("Failed to get receipt for tx %s: %v", tx.Hash().Hex(), err)
			continue
		}
		receipts[tx.Hash()] = receipt
	}

	p.mu.Lock()
	p.lastBlockNumber = endBlock
	p.mu.Unlock()

	return Data{
		Block:    block,
		Receipts: receipts,
	}, nil
}

// Close terminates the connection
func (p *PollingMethod) Close() error {
	// No specific cleanup needed for polling
	return nil
}
