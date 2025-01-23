package com.twilight.aggregator.processor;

import java.io.Serializable;
import java.math.BigDecimal;
import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Timestamp;
import java.util.HashMap;
import java.util.Map;
import java.util.Objects;

import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.windowing.ProcessWindowFunction;
import org.apache.flink.streaming.api.windowing.time.Time;
import org.apache.flink.streaming.api.windowing.windows.TimeWindow;
import org.apache.flink.util.Collector;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.twilight.aggregator.config.AppConfig;
import com.twilight.aggregator.model.Event;
import com.twilight.aggregator.model.PairMetric;

public class PairEventProcessor extends ProcessWindowFunction<Event, PairMetric, String, TimeWindow> implements Serializable {
    private static final long serialVersionUID = 2L;
    private static final BigDecimal WEI_TO_ETH = new BigDecimal("1000000000000000000"); // 10^18
    private static final Logger LOG = LoggerFactory.getLogger(PairEventProcessor.class);
    private static final String USDC_ADDRESS = "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48".toLowerCase();

    private final long windowSizeMillis;
    private final String windowName;
    
    // Mark all non-serializable fields as transient
    private transient Connection connection;
    private transient Map<String, String> pairIdCache;
    private transient Map<String, TokenMetadata> tokenMetadataCache;
    private transient Map<String, PairMetadata> pairMetadataCache;
    private transient Map<String, BigDecimal> tokenPriceCache;
    private transient AppConfig config;

    // Token metadata class
    private static class TokenMetadata implements Serializable {
        private static final long serialVersionUID = 1L;
        private final String address;
        private final int decimals;
        private final String symbol;

        public TokenMetadata(String address, int decimals, String symbol) {
            this.address = address;
            this.decimals = decimals;
            this.symbol = symbol;
        }

        public String getAddress() {
            return address;
        }

        public int getDecimals() {
            return decimals;
        }

        @Override
        public boolean equals(Object o) {
            if (this == o) return true;
            if (o == null || getClass() != o.getClass()) return false;
            TokenMetadata that = (TokenMetadata) o;
            return decimals == that.decimals && 
                   Objects.equals(address, that.address) && 
                   Objects.equals(symbol, that.symbol);
        }

        @Override
        public int hashCode() {
            return Objects.hash(address, decimals, symbol);
        }
    }

    // Pair metadata class
    private static class PairMetadata implements Serializable {
        private static final long serialVersionUID = 1L;
        private final String pairId;
        private final String token0Address;
        private final String token1Address;
        private final String feeTier;

        public PairMetadata(String pairId, String token0Address, String token1Address, String feeTier) {
            this.pairId = pairId;
            this.token0Address = token0Address;
            this.token1Address = token1Address;
            this.feeTier = feeTier;
        }

        public String getPairId() {
            return pairId;
        }

        public String getToken0Address() {
            return token0Address;
        }

        public String getToken1Address() {
            return token1Address;
        }

        @Override
        public boolean equals(Object o) {
            if (this == o) return true;
            if (o == null || getClass() != o.getClass()) return false;
            PairMetadata that = (PairMetadata) o;
            return Objects.equals(pairId, that.pairId) && 
                   Objects.equals(token0Address, that.token0Address) && 
                   Objects.equals(token1Address, that.token1Address) && 
                   Objects.equals(feeTier, that.feeTier);
        }

        @Override
        public int hashCode() {
            return Objects.hash(pairId, token0Address, token1Address, feeTier);
        }
    }

    public PairEventProcessor(Time windowSize, String windowName) {
        this.windowSizeMillis = windowSize.toMilliseconds();
        this.windowName = windowName;
    }

    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        // Initialize all transient fields in open()
        config = AppConfig.builder().build();
        initializeConnection();
        pairIdCache = new HashMap<>();
        tokenMetadataCache = new HashMap<>();
        pairMetadataCache = new HashMap<>();
        tokenPriceCache = new HashMap<>();
        loadTokenMetadata();
        loadPairMetadata();
    }

    @Override
    public void close() throws Exception {
        if (connection != null && !connection.isClosed()) {
            try {
                connection.close();
            } catch (SQLException e) {
                LOG.error("Error closing connection", e);
            }
        }
        super.close();
    }

    private void initializeConnection() throws SQLException {
        try {
            connection = DriverManager.getConnection(
                config.getJdbcUrl(),
                config.getJdbcUser(),
                config.getJdbcPassword()
            );
        } catch (SQLException e) {
            LOG.error("Failed to initialize database connection", e);
            throw e;
        }
    }

    private String getPairId(String pairAddress) {
        PairMetadata metadata = pairMetadataCache.get(pairAddress.toLowerCase());
        return metadata != null ? metadata.getPairId() : null;
    }

    private String getTokenAddress(String pairAddress, boolean isToken0) {
        PairMetadata metadata = pairMetadataCache.get(pairAddress.toLowerCase());
        if (metadata == null) {
            LOG.warn("Pair metadata not found for address: {}", pairAddress);
            return null;
        }
        return isToken0 ? metadata.getToken0Address() : metadata.getToken1Address();
    }

    private void loadTokenMetadata() {
        String sql = "SELECT token_address, token_decimals, token_symbol FROM token";
        try (PreparedStatement stmt = connection.prepareStatement(sql)) {
            try (ResultSet rs = stmt.executeQuery()) {
                while (rs.next()) {
                    String address = rs.getString("token_address").toLowerCase();
                    TokenMetadata metadata = new TokenMetadata(
                        address,
                        rs.getInt("token_decimals"),
                        rs.getString("token_symbol")
                    );
                    tokenMetadataCache.put(address, metadata);
                }
            }
        } catch (SQLException e) {
            LOG.error("Failed to load token metadata", e);
        }
    }

    private void loadPairMetadata() {
        String sql = "SELECT id, pair_address, token0_id, token1_id, fee_tier FROM twswap_pair";
        try (PreparedStatement stmt = connection.prepareStatement(sql)) {
            try (ResultSet rs = stmt.executeQuery()) {
                while (rs.next()) {
                    String pairAddress = rs.getString("pair_address").toLowerCase();
                    PairMetadata metadata = new PairMetadata(
                        rs.getString("id"),
                        getTokenAddressById(rs.getLong("token0_id")),
                        getTokenAddressById(rs.getLong("token1_id")),
                        rs.getString("fee_tier")
                    );
                    pairMetadataCache.put(pairAddress, metadata);
                }
            }
        } catch (SQLException e) {
            LOG.error("Failed to load pair metadata", e);
        }
    }

    private String getTokenAddressById(long tokenId) {
        String sql = "SELECT token_address FROM token WHERE id = ?";
        try (PreparedStatement stmt = connection.prepareStatement(sql)) {
            stmt.setLong(1, tokenId);
            try (ResultSet rs = stmt.executeQuery()) {
                if (rs.next()) {
                    return rs.getString("token_address").toLowerCase();
                }
            }
        } catch (SQLException e) {
            LOG.error("Failed to get token address for id: " + tokenId, e);
        }
        return null;
    }

    private BigDecimal normalizeAmount(BigDecimal amount, int decimals) {
        return amount.movePointLeft(decimals);
    }

    private BigDecimal parseHexValue(String hexValue) {
        if (hexValue.startsWith("0x")) {
            return new BigDecimal(new java.math.BigInteger(hexValue.substring(2), 16));
        }
        return new BigDecimal(hexValue);
    }

    private String getEventValue(Map<String, Object> eventData, String key) {
        Object value = eventData.get(key);
        return value != null ? value.toString() : null;
    }

    private BigDecimal convertWeiToEth(BigDecimal weiAmount) {
        if (weiAmount == null) {
            return BigDecimal.ZERO;
        }
        return weiAmount.divide(WEI_TO_ETH, 18, BigDecimal.ROUND_HALF_UP);
    }

    private void updateTokenPrice(String pairAddress, String tokenAddress, BigDecimal reserve0, BigDecimal reserve1) {
        PairMetadata metadata = pairMetadataCache.get(pairAddress.toLowerCase());
        if (metadata == null) {
            LOG.warn("Pair metadata not found for address: {}", pairAddress);
            return;
        }

        String token0Address = metadata.getToken0Address();
        String token1Address = metadata.getToken1Address();

        if (token0Address.equals(USDC_ADDRESS)) {
            // token1 price = USDC amount / token1 amount
            BigDecimal price = reserve0.divide(reserve1, 18, BigDecimal.ROUND_HALF_UP);
            tokenPriceCache.put(token1Address, price);
            LOG.info("Updated price for token {}: {}", token1Address, price);
        } else if (token1Address.equals(USDC_ADDRESS)) {
            // token0 price = USDC amount / token0 amount
            BigDecimal price = reserve1.divide(reserve0, 18, BigDecimal.ROUND_HALF_UP);
            tokenPriceCache.put(token0Address, price);
            LOG.info("Updated price for token {}: {}", token0Address, price);
        }
    }

    private BigDecimal getTokenPrice(String pairAddress, boolean isToken0) {
        
            String tokenAddress = getTokenAddress(pairAddress, isToken0).toLowerCase();
            return tokenPriceCache.getOrDefault(tokenAddress, BigDecimal.ONE);
         
    }

    @Override
    public void process(
            String pairAddress,
            ProcessWindowFunction<Event, PairMetric, String, TimeWindow>.Context context,
            Iterable<Event> events,
            Collector<PairMetric> out
    ) throws Exception {
        String pairId = getPairId(pairAddress);
        if (pairId == null) {
            LOG.warn("Pair not found in database: {}", pairAddress);
            return;
        }

        // Initialize metrics
        BigDecimal token0VolumeUsd = BigDecimal.ZERO;
        BigDecimal token1VolumeUsd = BigDecimal.ZERO;
        BigDecimal volumeUsd = BigDecimal.ZERO;
        int txCount = 0;
        
        // Latest state
        BigDecimal token0Reserve = BigDecimal.ZERO;
        BigDecimal token1Reserve = BigDecimal.ZERO;
        BigDecimal reserveUsd = BigDecimal.ZERO;

        // Process events in order
        for (Event event : events) {
            txCount++; // Every event increases transaction count
            
            Map<String, Object> eventData = event.getDecodedArgs();
            if (eventData == null) {
                LOG.warn("No decoded args found for event: {}", event.getEventName());
                continue;
            }
            
            switch (event.getEventName().toLowerCase()) {
                case "sync":
                    // Update reserves from eventData
                    String reserve0 = getEventValue(eventData, "reserve0");
                    String reserve1 = getEventValue(eventData, "reserve1");
                    if (reserve0 != null && reserve1 != null) {
                        token0Reserve = convertWeiToEth(new BigDecimal(reserve0));
                        token1Reserve = convertWeiToEth(new BigDecimal(reserve1));
                        
                        // Update token prices if this is a USDC pair
                        updateTokenPrice(pairAddress, getTokenAddress(pairAddress, true), token0Reserve, token1Reserve);
                        
                        // Calculate reserve USD using cached prices
                        reserveUsd = calculateReserveUsd(token0Reserve, token1Reserve, 
                                getTokenPrice(pairAddress, true), 
                                getTokenPrice(pairAddress, false));
                    }
                    break;
                    
                case "swap":
                    // Get amounts from eventData
                    String amount0InStr = getEventValue(eventData, "amount0In");
                    String amount0OutStr = getEventValue(eventData, "amount0Out");
                    String amount1InStr = getEventValue(eventData, "amount1In");
                    String amount1OutStr = getEventValue(eventData, "amount1Out");
                    
                    if (amount0InStr != null && amount0OutStr != null && 
                        amount1InStr != null && amount1OutStr != null) {
                        BigDecimal amount0In = convertWeiToEth(new BigDecimal(amount0InStr));
                        BigDecimal amount0Out = convertWeiToEth(new BigDecimal(amount0OutStr));
                        BigDecimal amount1In = convertWeiToEth(new BigDecimal(amount1InStr));
                        BigDecimal amount1Out = convertWeiToEth(new BigDecimal(amount1OutStr));
                        
                        // Calculate volumes using cached prices
                        BigDecimal token0Volume = calculateTokenVolume(
                                amount0In, 
                                amount0Out, 
                                getTokenPrice(pairAddress, true)
                        );
                        BigDecimal token1Volume = calculateTokenVolume(
                                amount1In, 
                                amount1Out, 
                                getTokenPrice(pairAddress, false)
                        );
                        
                        token0VolumeUsd = token0VolumeUsd.add(token0Volume);
                        token1VolumeUsd = token1VolumeUsd.add(token1Volume);
                        volumeUsd = token0VolumeUsd.add(token1VolumeUsd);
                    }
                    break;
                    
                case "mint":
                case "burn":
                    // These events only contribute to transaction count
                    break;
                    
                default:
                    LOG.warn("Unknown event type: {} for pair: {}", event.getEventName(), pairAddress);
                    break;
            }
        }

        // Create and populate metric
        PairMetric metric = new PairMetric();
        metric.setPairId(Long.parseLong(pairId));
        metric.setTimeWindow(windowName);
        metric.setEndTime(new Timestamp(context.window().getEnd()));
        
        // Set reserves
        metric.setToken0Reserve(token0Reserve);
        metric.setToken1Reserve(token1Reserve);
        metric.setReserveUsd(reserveUsd);
        
        // Set volumes
        metric.setToken0VolumeUsd(token0VolumeUsd);
        metric.setToken1VolumeUsd(token1VolumeUsd);
        metric.setVolumeUsd(volumeUsd);
        
        // Set transaction count
        metric.setTxcnt(txCount);

        LOG.info("Processed metrics for pair {}: {}", pairAddress, metric);
        out.collect(metric);
    }

    private BigDecimal calculateTokenVolume(BigDecimal amountIn, BigDecimal amountOut, BigDecimal tokenPrice) {
        if (tokenPrice == null || tokenPrice.compareTo(BigDecimal.ZERO) <= 0) {
            return BigDecimal.ZERO;
        }
        return (amountIn.add(amountOut)).multiply(tokenPrice);
    }

    private BigDecimal calculateReserveUsd(
            BigDecimal token0Reserve, 
            BigDecimal token1Reserve, 
            BigDecimal token0Price, 
            BigDecimal token1Price
    ) {
        BigDecimal reserve0Usd = BigDecimal.ZERO;
        BigDecimal reserve1Usd = BigDecimal.ZERO;
        
        if (token0Price != null && token0Price.compareTo(BigDecimal.ZERO) > 0) {
            reserve0Usd = token0Reserve.multiply(token0Price);
        }
        
        if (token1Price != null && token1Price.compareTo(BigDecimal.ZERO) > 0) {
            reserve1Usd = token1Reserve.multiply(token1Price);
        }
        
        return reserve0Usd.add(reserve1Usd);
    }
} 