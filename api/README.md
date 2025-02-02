# Twilight Token Analytics API Documentation

## Overview

This document describes the GraphQL API endpoints for querying token analytics data, including token information, historical metrics, and time series data for analysis.

## API Endpoint

```
http://localhost:8091/query
```

GraphQL Playground (for development): `http://localhost:8091/`

## Schema Types

### Custom Scalars

```graphql
scalar DateTime  # ISO-8601 format, e.g., "2024-01-01T00:00:00Z"
scalar Decimal   # High-precision decimal numbers
```

### Token
```graphql
type Token {
  id: ID!
  chainId: String!
  chainName: String
  tokenAddress: String!
  tokenSymbol: String
  tokenName: String
  tokenDecimals: Int
  hypeScore: Int
  supplyUsd: Decimal         # FDV (Fully Diluted Valuation)
  liquidityUsd: Decimal      # Total liquidity estimation
  createTime: DateTime
  updateTime: DateTime
  currentPrice: Decimal      # Current token price in USD
  priceChange1h: Decimal     # Price change percentage in last hour
  volume1h: Decimal          # Trading volume in last hour
  buyPressure1h: Decimal     # Buy pressure in last hour
}
```

### TokenMetric
```graphql
type TokenMetric {
  id: ID!
  tokenId: ID!
  timeWindow: String         # e.g., "20s", "1min", "5min", "30min", "1h"
  endTime: DateTime
  volumeUsd: Decimal
  txcnt: Int
  tokenPriceUsd: Decimal
  buyPressureUsd: Decimal    # buy_volume - sell_volume
  buyersCount: Int          # Unique buyers count
  sellersCount: Int         # Unique sellers count
  buyVolumeUsd: Decimal
  sellVolumeUsd: Decimal
  makersCount: Int          # Total unique traders (buyers + sellers)
  buyCount: Int            # Number of buy transactions
  sellCount: Int           # Number of sell transactions
  updateTime: DateTime
  token: Token!             # Associated token information
}
```

### TokenStats
```graphql
type TokenStats {
  token: Token!
  currentPrice: Decimal!
  priceChange1h: Decimal!
  volume1h: Decimal!
  buyPressure1h: Decimal!
  lastUpdate: DateTime!
}
```

### TimeSeriesDataPoint
```graphql
type TimeSeriesDataPoint {
  timestamp: DateTime!
  tokenPriceUsd: Decimal
  buyVolumeUsd: Decimal
  sellVolumeUsd: Decimal
  volumeUsd: Decimal
  txcnt: Int
  makersCount: Int         # Total unique traders
  buyCount: Int           # Buy transactions count
  sellCount: Int          # Sell transactions count
  buyersCount: Int        # Unique buyers count
  sellersCount: Int       # Unique sellers count
}
```

## Available Queries

### 1. Token Stats Queries

#### Get Real-time Token Stats
```graphql
query TokenStats($tokenId: ID!) {
  tokenStats(tokenId: $tokenId) {
    token {
      tokenSymbol
      tokenName
      liquidityUsd
      supplyUsd
      hypeScore
    }
    currentPrice
    priceChange1h
    volume1h
    buyPressure1h
    lastUpdate
  }
}
```

#### Get Multiple Tokens Stats
```graphql
query TokensStats($tokenIds: [ID!]!) {
  tokensStats(tokenIds: $tokenIds) {
    token {
      tokenSymbol
      tokenName
      liquidityUsd
      supplyUsd
      hypeScore
    }
    currentPrice
    priceChange1h
    volume1h
    buyPressure1h
    lastUpdate
  }
}
```

### 2. Token Metrics by Window

#### Get Latest Window Metrics
```graphql
query TokenMetricsByWindow($tokenId: ID!, $timeWindow: String!) {
  tokenMetricsByWindow(tokenId: $tokenId, timeWindow: $timeWindow) {
    timeWindow
    endTime
    volumeUsd
    txcnt
    tokenPriceUsd
    buyPressureUsd
    buyersCount
    sellersCount
    makersCount
    buyCount
    sellCount
    token {
      tokenSymbol
      tokenName
    }
  }
}
```

### 3. Time Series Analytics

#### Get Detailed Time Series Data
```graphql
query TokenAnalytics(
  $tokenId: ID!
  $timeWindow: String!
  $from: DateTime!
  $to: DateTime!
) {
  tokenAnalytics(
    tokenId: $tokenId
    timeWindow: $timeWindow
    from: $from
    to: $to
  ) {
    token {
      tokenSymbol
      tokenName
    }
    timeWindow
    dataPoints {
      timestamp
      tokenPriceUsd
      buyVolumeUsd
      sellVolumeUsd
      volumeUsd
      txcnt
      makersCount
      buyCount
      sellCount
      buyersCount
      sellersCount
    }
  }
}
```

## Real-time Updates

### Token Stats Subscription
```graphql
subscription TokenStatsUpdated($tokenId: ID!) {
  tokenStatsUpdated(tokenId: $tokenId) {
    token {
      tokenSymbol
      tokenName
    }
    currentPrice
    priceChange1h
    volume1h
    buyPressure1h
    lastUpdate
  }
}
```

### Token Metrics Subscription
```graphql
subscription TokenMetricsUpdated($tokenId: ID!, $timeWindow: String!) {
  tokenMetricsUpdated(tokenId: $tokenId, timeWindow: $timeWindow) {
    timeWindow
    endTime
    volumeUsd
    txcnt
    tokenPriceUsd
    buyPressureUsd
    buyersCount
    sellersCount
    makersCount
    buyCount
    sellCount
  }
}
```

## Time Window Options

The following time window values are supported:
- `20s` - 20 seconds
- `1m` - 1 minute
- `5m` - 5 minutes
- `30m` - 30 minutes

## Query Examples

### Token Metrics Query
```graphql
query {
  tokenMetrics(
    tokenId: "1"
    timeWindow: "5m"
    fromTime: "2024-01-01T00:00:00Z"
    toTime: "2024-01-02T00:00:00Z"
    limit: 100
  ) {
    timeWindow
    endTime
    volumeUsd
    txcnt
    tokenPriceUsd
    buyPressureUsd
    buyersCount
    sellersCount
    token {
      tokenSymbol
      tokenName
    }
  }
}
```

### Token Analytics Query
```graphql
query {
  tokenAnalytics(
    tokenId: "1"
    timeWindow: "5m"
    from: "2024-01-01T00:00:00Z"
    to: "2024-01-02T00:00:00Z"
  ) {
    token {
      tokenSymbol
      tokenName
    }
    timeWindow
    dataPoints {
      timestamp
      tokenPriceUsd
      buyVolumeUsd
      sellVolumeUsd
      volumeUsd
      txcnt
    }
  }
}
```

## Error Handling

The API returns standard GraphQL errors in the following format:

```json
{
  "errors": [
    {
      "message": "Error message description",
      "path": ["fieldName"],
      "locations": [{"line": 2, "column": 3}]
    }
  ]
}
```

Common error cases:
1. Invalid token ID or address
2. Invalid time window value
3. Invalid date range
4. Database connection issues

## Best Practices

1. Always specify only the fields you need
2. Use pagination for large result sets
3. Use appropriate time windows based on your analysis needs:
   - Use smaller windows (20s, 1min) for real-time monitoring
   - Use larger windows (5min, 30min, 1h) for trend analysis
4. Include error handling in your application
5. Consider caching responses for frequently accessed data

## Rate Limiting

Currently, there are no rate limits implemented. However, please be mindful of your request frequency to ensure optimal performance for all users.

## Support

For any questions or issues, please open a GitHub issue in the repository. 