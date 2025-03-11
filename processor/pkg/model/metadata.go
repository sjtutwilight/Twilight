package model

import (
	"log"
	"strconv"
)

// TokenMetadata represents token information stored in Redis
type TokenMetadata struct {
	ID       string `json:"id"`
	Address  string `json:"address"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Decimals string `json:"decimals"`
	ChainID  string `json:"chainId"`
}

// GetDecimalsAsInt returns the token decimals as an integer
func (t *TokenMetadata) GetDecimalsAsInt() int {
	if t.Decimals == "" {
		return 18 // 默认值
	}

	decimals, err := strconv.Atoi(t.Decimals)
	if err != nil {
		log.Printf("Warning: failed to parse decimals for token %s: %v, using default 18", t.Symbol, err)
		return 18
	}

	return decimals
}

// GetIDAsInt64 returns the token ID as int64
func (t *TokenMetadata) GetIDAsInt64() (int64, error) {
	return strconv.ParseInt(t.ID, 10, 64)
}

// AccountMetadata represents account information stored in Redis
type AccountMetadata struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Tag     string `json:"tag"`
}

// GetIDAsInt64 returns the account ID as int64
func (a *AccountMetadata) GetIDAsInt64() (int64, error) {
	return strconv.ParseInt(a.ID, 10, 64)
}

// PairMetadata represents pair information stored in Redis
type PairMetadata struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Token0  struct {
		ID      string `json:"id"`
		Address string `json:"address"`
		Symbol  string `json:"symbol"`
	} `json:"token0"`
	Token1 struct {
		ID      string `json:"id"`
		Address string `json:"address"`
		Symbol  string `json:"symbol"`
	} `json:"token1"`
	ChainID string `json:"chainId"`
}

// GetIDAsInt64 returns the pair ID as int64
func (p *PairMetadata) GetIDAsInt64() (int64, error) {
	return strconv.ParseInt(p.ID, 10, 64)
}
