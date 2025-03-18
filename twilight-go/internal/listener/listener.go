package listener

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
)

// Listener is the main component that orchestrates data ingestion
type Listener struct {
	registry       *Registry
	normalizer     Normalizer
	producer       Producer
	reorgProcessor *ReorgProcessor
	contracts      map[common.Address]*Contract
	rawDataChan    chan Data
	dataChan       chan NormalizedData
	errorChan      chan error
	stopChan       chan struct{}
	wg             sync.WaitGroup
}

// NewListener creates a new listener
func NewListener(registry *Registry, normalizer Normalizer, producer Producer, reorgProcessor *ReorgProcessor) *Listener {
	return &Listener{
		registry:       registry,
		normalizer:     normalizer,
		producer:       producer,
		reorgProcessor: reorgProcessor,
		contracts:      make(map[common.Address]*Contract),
		rawDataChan:    make(chan Data, 100),           // Buffer size of 100
		dataChan:       make(chan NormalizedData, 100), // Buffer size of 100
		errorChan:      make(chan error, 10),
		stopChan:       make(chan struct{}),
	}
}

// Start begins the listener operation
func (l *Listener) Start(ctx context.Context) error {
	// Set data channels for all data sources
	sources := l.registry.GetAllSources()
	for _, source := range sources {
		if localNode, ok := source.(*LocalNodeSource); ok {
			localNode.SetDataChannel(l.rawDataChan)
		}
	}

	// Start all data sources
	for id, source := range sources {
		l.wg.Add(1)
		go func(id string, src DataSource) {
			defer l.wg.Done()
			if err := src.Start(ctx); err != nil {
				log.Printf("Error starting data source %s: %v", id, err)
				l.errorChan <- fmt.Errorf("error starting data source %s: %w", id, err)
			}
		}(id, source)
	}

	// Start the raw data processing goroutine
	l.wg.Add(1)
	go l.processRawData(ctx)

	// Start the normalized data processing goroutine
	l.wg.Add(1)
	go l.processData(ctx)

	return nil
}

// Stop halts the listener operation
func (l *Listener) Stop() error {
	close(l.stopChan)

	// Stop all data sources
	sources := l.registry.GetAllSources()
	for id, source := range sources {
		if err := source.Stop(); err != nil {
			log.Printf("Error stopping data source %s: %v", id, err)
		}
	}

	// Wait for all goroutines to finish
	l.wg.Wait()

	// Close channels
	close(l.dataChan)
	close(l.errorChan)

	return nil
}

// processRawData handles incoming raw data
func (l *Listener) processRawData(ctx context.Context) {
	defer l.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-l.stopChan:
			return
		case data := <-l.rawDataChan:
			// Process chain reorganization
			if data.Block != nil {
				if err := l.reorgProcessor.ProcessBlock(ctx, data.Block); err != nil {
					log.Printf("Error processing block for reorg: %v", err)
				}
			}

			// Normalize the data
			normalizedData, err := l.normalizer.Normalize(data)
			if err != nil {
				log.Printf("Error normalizing data: %v", err)
				continue
			}

			// Send the normalized data to the data channel
			select {
			case l.dataChan <- normalizedData:
				log.Printf("Normalized data for block %d", data.Block.NumberU64())
			default:
				log.Printf("Data channel full, dropping normalized data for block %d", data.Block.NumberU64())
			}
		}
	}
}

// processData handles incoming data
func (l *Listener) processData(ctx context.Context) {
	defer l.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-l.stopChan:
			return
		case data := <-l.dataChan:
			// Process the normalized data
			if err := l.handleNormalizedData(ctx, data); err != nil {
				log.Printf("Error handling normalized data: %v", err)
			}
		}
	}
}

// handleNormalizedData processes normalized data
func (l *Listener) handleNormalizedData(ctx context.Context, data NormalizedData) error {
	// Send the chain event to Kafka
	if err := l.producer.Produce(data.ChainEvent.Transaction.TransactionHash, data.ChainEvent); err != nil {
		return fmt.Errorf("failed to produce chain event: %w", err)
	}

	return nil
}

// AddContract adds a contract to monitor
func (l *Listener) AddContract(address common.Address, contract *Contract) {
	if normalizer, ok := l.normalizer.(*DexTransactionNormalizer); ok {
		normalizer.AddContract(address, contract)
	}
}

// GetDataChan returns the data channel
func (l *Listener) GetDataChan() <-chan NormalizedData {
	return l.dataChan
}

// GetErrorChan returns the error channel
func (l *Listener) GetErrorChan() <-chan error {
	return l.errorChan
}

// CreateLocalNodeListener creates a listener with a local node data source
func CreateLocalNodeListener(nodeURL string, kafkaBrokers []string, topic string, registry *Registry) (*Listener, error) {
	// Create the local node data source
	localNode, err := NewLocalNodeSource(nodeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create local node data source: %w", err)
	}

	// Register the data source with the provided registry
	if err := registry.Register("local_node", localNode); err != nil {
		return nil, fmt.Errorf("failed to register local node data source: %w", err)
	}

	// Create the normalizer
	normalizer := NewDexTransactionNormalizer()

	// Create the producer
	producer, err := NewKafkaProducer(kafkaBrokers, topic)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	// Create the reorg processor
	maxBlocksToKeep := viper.GetInt("chain.reorg.max_blocks_to_keep")
	if maxBlocksToKeep <= 0 {
		maxBlocksToKeep = 100 // Default value
	}
	reorgProcessor := NewReorgProcessor(producer, maxBlocksToKeep)

	// Create the listener
	listener := NewListener(registry, normalizer, producer, reorgProcessor)

	return listener, nil
}
