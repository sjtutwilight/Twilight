package aggregator

import (
	"time"

	"twilight/pkg/types"
)

// MetricType represents the type of metric window
type MetricType string

const (
	MetricType30s MetricType = "30s"
	MetricType5m  MetricType = "5m"
	MetricType30m MetricType = "30m"
)

// MetricWindow represents a time window for metrics
type MetricWindow struct {
	Size  time.Duration
	Slide time.Duration
	Type  MetricType
}

// TokenMetricAccumulator accumulates token metrics within a window
type TokenMetricAccumulator struct {
	TokenID          int64
	TotalSupply      float64
	TradeVolumeUSD   float64
	TransactionCount int
	TotalLiquidity   float64
	PriceUSD         float64
	LastUpdateTime   time.Time
}

// PairMetricAccumulator accumulates pair metrics within a window
type PairMetricAccumulator struct {
	PairID           int64
	Token0Address    string
	Token1Address    string
	Reserve0         float64
	Reserve1         float64
	TotalSupply      float64
	ReserveUSD       float64
	Token0Volume     float64
	Token1Volume     float64
	VolumeUSD        float64
	TransactionCount int
	LastUpdateTime   time.Time
}

// EventMessage represents a processed event message from Kafka
type EventMessage struct {
	Event           types.Event
	BlockTimestamp  time.Time
	DecodedData     types.DecodedEvent
	TransactionHash string
}

// MetricResult represents the final metric to be written to database/kafka
type MetricResult struct {
	WindowStart  time.Time
	WindowEnd    time.Time
	MetricType   MetricType
	TokenMetrics []types.TokenMetric
	PairMetrics  []types.PairMetric
}
