package handlers

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Event topic hashes
const (
	PairCreatedTopic = "0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9"
	SwapTopic        = "0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822"
	MintTopic        = "0x4c209b5fc8ad50758f13e2e1088ba56a560dff690a1c6fef26394f4c03821c4f"
	BurnTopic        = "0xdccd412f0b1252819cb1fd330b93224ca42612892bb3f4f789976e6d81936496"
	SyncTopic        = "0x1c411e9a96e071241c2f21f7726b17ae89e3cab4c78be50e062b03a9fffbbad1"
	TransferTopic    = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	ApprovalTopic    = "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"
)

// EventHandler is a function type that processes an Ethereum log and returns decoded data
type EventHandler func(log types.Log) (map[string]interface{}, error)

// PairCreatedHandler handles the PairCreated event from TWSwapFactory
func PairCreatedHandler(log types.Log) (map[string]interface{}, error) {
	if len(log.Data) < 32 {
		return nil, fmt.Errorf("PairCreated handler: data length = %d, want >= 32", len(log.Data))
	}

	pair := new(big.Int).SetBytes(log.Data[:32])

	return map[string]interface{}{
		"token0": log.Topics[1].Hex(),
		"token1": log.Topics[2].Hex(),
		"pair":   pair.String(),
	}, nil
}

// SwapHandler handles the Swap event from TWSwapPair
func SwapHandler(log types.Log) (map[string]interface{}, error) {
	if len(log.Data) < 128 {
		return nil, fmt.Errorf("Swap handler: data length = %d, want >= 128", len(log.Data))
	}
	amount0In := new(big.Int).SetBytes(log.Data[0:32])
	amount1In := new(big.Int).SetBytes(log.Data[32:64])
	amount0Out := new(big.Int).SetBytes(log.Data[64:96])
	amount1Out := new(big.Int).SetBytes(log.Data[96:128])

	return map[string]interface{}{
		"sender":     common.HexToAddress(log.Topics[1].Hex()).Hex(),
		"to":         common.HexToAddress(log.Topics[2].Hex()).Hex(),
		"amount0In":  amount0In.String(),
		"amount1In":  amount1In.String(),
		"amount0Out": amount0Out.String(),
		"amount1Out": amount1Out.String(),
	}, nil
}

// MintHandler handles the Mint event from TWSwapPair
func MintHandler(log types.Log) (map[string]interface{}, error) {
	if len(log.Data) < 64 {
		return nil, fmt.Errorf("Mint handler: data length = %d, want >= 64", len(log.Data))
	}
	amount0 := new(big.Int).SetBytes(log.Data[0:32])
	amount1 := new(big.Int).SetBytes(log.Data[32:64])

	return map[string]interface{}{
		"sender":  common.HexToAddress(log.Topics[1].Hex()).Hex(),
		"amount0": amount0.String(),
		"amount1": amount1.String(),
	}, nil
}

// BurnHandler handles the Burn event from TWSwapPair
func BurnHandler(log types.Log) (map[string]interface{}, error) {
	if len(log.Data) < 64 {
		return nil, fmt.Errorf("Burn handler: data length = %d, want >= 64", len(log.Data))
	}
	amount0 := new(big.Int).SetBytes(log.Data[0:32])
	amount1 := new(big.Int).SetBytes(log.Data[32:64])

	return map[string]interface{}{
		"sender":  common.HexToAddress(log.Topics[1].Hex()).Hex(),
		"to":      common.HexToAddress(log.Topics[2].Hex()).Hex(),
		"amount0": amount0.String(),
		"amount1": amount1.String(),
	}, nil
}

// SyncHandler handles the Sync event from TWSwapPair
func SyncHandler(log types.Log) (map[string]interface{}, error) {
	if len(log.Data) < 64 {
		return nil, fmt.Errorf("Sync handler: data length = %d, want >= 64", len(log.Data))
	}
	reserve0 := new(big.Int).SetBytes(log.Data[0:32])
	reserve1 := new(big.Int).SetBytes(log.Data[32:64])

	return map[string]interface{}{
		"reserve0": reserve0.String(),
		"reserve1": reserve1.String(),
	}, nil
}

// TransferHandler handles the Transfer event from ERC20 tokens
func TransferHandler(log types.Log) (map[string]interface{}, error) {
	if len(log.Data) < 32 {
		return nil, fmt.Errorf("Transfer handler: data length = %d, want >= 32", len(log.Data))
	}

	if len(log.Topics) < 3 {
		return nil, fmt.Errorf("Transfer handler: topics length = %d, want >= 3", len(log.Topics))
	}

	value := new(big.Int).SetBytes(log.Data[:32])
	return map[string]interface{}{
		"from":  log.Topics[1].Hex(),
		"to":    log.Topics[2].Hex(),
		"value": value.String(),
	}, nil
}

// ApprovalHandler handles the Approval event from ERC20 tokens
func ApprovalHandler(log types.Log) (map[string]interface{}, error) {
	if len(log.Data) < 32 {
		return nil, fmt.Errorf("Approval handler: data length = %d, want >= 32", len(log.Data))
	}

	value := new(big.Int).SetBytes(log.Data[:32])
	return map[string]interface{}{
		"owner":   log.Topics[1].Hex(),
		"spender": log.Topics[2].Hex(),
		"value":   value.String(),
	}, nil
}

// GetPairHandlers returns a map of event handlers for TWSwapPair contract
func GetPairHandlers() map[string]EventHandler {
	return map[string]EventHandler{
		SwapTopic: SwapHandler,
		MintTopic: MintHandler,
		BurnTopic: BurnHandler,
		SyncTopic: SyncHandler,
	}
}

// GetTokenHandlers returns a map of event handlers for ERC20 token contract
func GetTokenHandlers() map[string]EventHandler {
	return map[string]EventHandler{
		TransferTopic: TransferHandler,
		ApprovalTopic: ApprovalHandler,
	}
}
