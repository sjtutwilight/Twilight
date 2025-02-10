package model

import (
	"context"
	"sync"
	"time"
)

type BatchProcessor struct {
	batchSize int
	timeout   time.Duration
	events    []Event
	mu        sync.Mutex
	processor func([]Event) error
}

type BatchConfig struct {
	BatchSize int
	Timeout   time.Duration
}

func NewBatchProcessor(config BatchConfig, processor func([]Event) error) *BatchProcessor {
	return &BatchProcessor{
		batchSize: config.BatchSize,
		timeout:   config.Timeout,
		events:    make([]Event, 0, config.BatchSize),
		processor: processor,
	}
}

func (bp *BatchProcessor) Add(event Event) error {
	bp.mu.Lock()
	bp.events = append(bp.events, event)
	currentSize := len(bp.events)
	bp.mu.Unlock()

	if currentSize >= bp.batchSize {
		return bp.process()
	}
	return nil
}

func (bp *BatchProcessor) process() error {
	bp.mu.Lock()
	if len(bp.events) == 0 {
		bp.mu.Unlock()
		return nil
	}
	events := bp.events
	bp.events = make([]Event, 0, bp.batchSize)
	bp.mu.Unlock()

	return bp.processor(events)
}

func (bp *BatchProcessor) Start(ctx context.Context) {
	ticker := time.NewTicker(bp.timeout)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			bp.process()
			return
		case <-ticker.C:
			bp.process()
		}
	}
}
