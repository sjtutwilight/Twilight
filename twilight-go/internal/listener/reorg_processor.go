package listener

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// BlockInfo stores information about a processed block
type BlockInfo struct {
	Number     uint64
	Hash       common.Hash
	ParentHash common.Hash
}

// ReorgEvent represents a chain reorganization event
type ReorgEvent struct {
	OldBlocks []BlockInfo `json:"oldBlocks"`
	NewBlocks []BlockInfo `json:"newBlocks"`
}

// ReorgProcessor handles chain reorganization detection and notification
type ReorgProcessor struct {
	producer        Producer
	processedBlocks map[uint64]BlockInfo
	mu              sync.RWMutex
	maxBlocksToKeep int
}

// NewReorgProcessor creates a new chain reorganization processor
func NewReorgProcessor(producer Producer, maxBlocksToKeep int) *ReorgProcessor {
	if maxBlocksToKeep <= 0 {
		maxBlocksToKeep = 100 // Default value
	}

	return &ReorgProcessor{
		producer:        producer,
		processedBlocks: make(map[uint64]BlockInfo),
		maxBlocksToKeep: maxBlocksToKeep,
	}
}

// ProcessBlock checks for chain reorganization and records block information
func (p *ReorgProcessor) ProcessBlock(ctx context.Context, block *types.Block) error {
	blockNumber := block.NumberU64()
	blockHash := block.Hash()
	parentHash := block.ParentHash()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if we have processed a block at this height before
	if existingBlock, exists := p.processedBlocks[blockNumber]; exists {
		// If the hash is different, we have a reorg
		if existingBlock.Hash != blockHash {
			log.Printf("Chain reorganization detected at block %d: old hash %s, new hash %s",
				blockNumber, existingBlock.Hash.Hex(), blockHash.Hex())

			// Find the fork point
			forkPoint := p.findForkPoint(blockNumber, parentHash)

			// Collect old and new blocks
			oldBlocks := p.collectOldBlocks(forkPoint, blockNumber)
			newBlocks := []BlockInfo{{
				Number:     blockNumber,
				Hash:       blockHash,
				ParentHash: parentHash,
			}}

			// Create reorg event
			reorgEvent := ReorgEvent{
				OldBlocks: oldBlocks,
				NewBlocks: newBlocks,
			}

			// Notify about the reorg
			if err := p.producer.Produce("reorg", reorgEvent); err != nil {
				return fmt.Errorf("failed to produce reorg event: %w", err)
			}

			// Update the block info
			p.processedBlocks[blockNumber] = BlockInfo{
				Number:     blockNumber,
				Hash:       blockHash,
				ParentHash: parentHash,
			}
		}
	} else {
		// Record the new block
		p.processedBlocks[blockNumber] = BlockInfo{
			Number:     blockNumber,
			Hash:       blockHash,
			ParentHash: parentHash,
		}
	}

	// Clean up old blocks
	p.cleanupOldBlocks(blockNumber)

	return nil
}

// findForkPoint finds the point where the chain forked
func (p *ReorgProcessor) findForkPoint(blockNumber uint64, parentHash common.Hash) uint64 {
	// Start from the parent of the current block
	currentNumber := blockNumber - 1

	// Look for the fork point by checking if the parent hash matches
	for {
		if currentNumber == 0 {
			// We reached the genesis block
			return 0
		}

		if existingBlock, exists := p.processedBlocks[currentNumber]; exists {
			if existingBlock.Hash == parentHash {
				// We found the fork point
				return currentNumber
			}
			// Continue with the parent of the existing block
			parentHash = existingBlock.ParentHash
		}

		currentNumber--
	}
}

// collectOldBlocks collects blocks that are being replaced
func (p *ReorgProcessor) collectOldBlocks(forkPoint, blockNumber uint64) []BlockInfo {
	var oldBlocks []BlockInfo

	for i := forkPoint + 1; i <= blockNumber; i++ {
		if block, exists := p.processedBlocks[i]; exists {
			oldBlocks = append(oldBlocks, block)
		}
	}

	return oldBlocks
}

// cleanupOldBlocks removes old blocks to prevent memory leaks
func (p *ReorgProcessor) cleanupOldBlocks(currentBlockNumber uint64) {
	if currentBlockNumber <= uint64(p.maxBlocksToKeep) {
		return
	}

	// Remove blocks older than maxBlocksToKeep
	for i := uint64(0); i < currentBlockNumber-uint64(p.maxBlocksToKeep); i++ {
		delete(p.processedBlocks, i)
	}
}
