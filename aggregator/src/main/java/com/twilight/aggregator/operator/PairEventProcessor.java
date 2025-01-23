// package com.twilight.aggregator.operator;

// import java.io.Serializable;
// import java.math.BigDecimal;
// import java.math.RoundingMode;
// import java.sql.Connection;
// import java.sql.DriverManager;
// import java.sql.PreparedStatement;
// import java.sql.ResultSet;
// import java.sql.SQLException;
// import java.util.HashMap;
// import java.util.Map;

// import org.apache.flink.configuration.Configuration;
// import org.apache.flink.streaming.api.functions.windowing.ProcessWindowFunction;
// import org.apache.flink.streaming.api.windowing.time.Time;
// import org.apache.flink.streaming.api.windowing.windows.TimeWindow;
// import org.apache.flink.util.Collector;
// import org.slf4j.Logger;
// import org.slf4j.LoggerFactory;

// import com.fasterxml.jackson.databind.JsonNode;
// import com.fasterxml.jackson.databind.ObjectMapper;
// import com.twilight.aggregator.config.AppConfig;
// import com.twilight.aggregator.model.Event;
// import com.twilight.aggregator.model.PairMetric;

// public class PairEventProcessor extends ProcessWindowFunction<Event, PairMetric, String, TimeWindow> implements Serializable {
//     private static final long serialVersionUID = 1L;
//     private static final Logger LOG = LoggerFactory.getLogger(PairEventProcessor.class);
//     private static final ObjectMapper objectMapper = new ObjectMapper();
//     private static final BigDecimal USDC_DECIMALS = new BigDecimal("18");

//     private final String windowName;
//     private final Time windowSize;
//     private final String jdbcUrl;
//     private final String jdbcUser;
//     private final String jdbcPassword;

//     private transient Connection connection;
//     private transient Map<String, TokenInfo> tokenCache;
//     private transient Map<String, UsdcPairInfo> usdcPairCache;
//     private transient Map<String, Long> pairIdCache;
//     private transient AppConfig config;

//     private static class TokenInfo implements Serializable {
//         String symbol;
//         int decimals;
//         String usdcPairAddress;
//     }

//     private static class UsdcPairInfo implements Serializable {
//         String token0;
//         String token1;
//         Event lastSyncEvent;
//     }

//     public PairEventProcessor(String windowName, Time windowSize, String jdbcUrl, String jdbcUser, String jdbcPassword) {
//         this.windowName = windowName;
//         this.windowSize = windowSize;
//         this.jdbcUrl = jdbcUrl;
//         this.jdbcUser = jdbcUser;
//         this.jdbcPassword = jdbcPassword;
//     }

//     @Override
//     public void open(Configuration parameters) throws Exception {
//         super.open(parameters);
//         config = AppConfig.builder().build();
//         initializeConnection();
//         tokenCache = new HashMap<>();
//         usdcPairCache = new HashMap<>();
//         pairIdCache = new HashMap<>();
//         loadTokenInfo();
//         LOG.info("PairEventProcessor initialization completed");
//     }

//     @Override
//     public void process(String pairAddress, Context context, Iterable<Event> events, Collector<PairMetric> out) throws Exception {
//         if (pairAddress == null) {
//             LOG.warn("Received null pair address");
//             return;
//         }

//         Long pairId = getPairId(pairAddress);
//         if (pairId == null) {
//             LOG.warn("Pair not found in database: {}", pairAddress);
//             return;
//         }

//         PairMetric metric = new PairMetric(pairId, windowName, windowSize, context.window().getEnd());
//         metric.setToken0Reserve(BigDecimal.ZERO);
//         metric.setToken1Reserve(BigDecimal.ZERO);
//         metric.setReserveUsd(BigDecimal.ZERO);
//         metric.setToken0VolumeUsd(BigDecimal.ZERO);
//         metric.setToken1VolumeUsd(BigDecimal.ZERO);
//         metric.setVolumeUsd(BigDecimal.ZERO);
//         metric.setTxcnt(0);

//         Event lastSyncEvent = null;
//         for (Event event : events) {
//             String eventName = event.getEventName().toLowerCase();
//             switch (eventName) {
//                 case "sync":
//                     lastSyncEvent = event;
//                     break;
//                 case "swap":
//                     processSwapEvent(event, metric);
//                     break;
//                 case "mint":
//                     processMintEvent(metric, objectMapper.valueToTree(event.getDecodedArgs()));
//                     break;
//                 case "burn":
//                     processBurnEvent(metric, objectMapper.valueToTree(event.getDecodedArgs()));
//                     break;
//             }
//             metric.incrementTxCount();
//         }

//         // Process the last sync event to get final reserves
//         if (lastSyncEvent != null) {
//             processSyncEvent(lastSyncEvent, metric);
//         }

//         saveMetric(metric);
//         out.collect(metric);
//     }

//     @Override
//     public void close() throws Exception {
//         if (connection != null && !connection.isClosed()) {
//             connection.close();
//         }
//         super.close();
//     }

//     private void initializeConnection() throws SQLException {
//         connection = DriverManager.getConnection(
//             config.getJdbcUrl(),
//             config.getJdbcUser(),
//             config.getJdbcPassword()
//         );
//     }

//     private void loadTokenInfo() throws SQLException {
//         String sql = "SELECT p.pair_address, " +
//                     "t0.token_address as token0_address, t0.token_symbol as token0_symbol, t0.token_decimals as token0_decimals, " +
//                     "t1.token_address as token1_address, t1.token_symbol as token1_symbol, t1.token_decimals as token1_decimals " +
//                     "FROM twswap_pair p " +
//                     "JOIN token t0 ON p.token0_id = t0.id " +
//                     "JOIN token t1 ON p.token1_id = t1.id";
        
//         LOG.debug("Loading token information with query: {}", sql);
//         try (PreparedStatement stmt = connection.prepareStatement(sql)) {
//             ResultSet rs = stmt.executeQuery();
//             int tokenCount = 0;
//             int usdcPairCount = 0;
            
//             while (rs.next()) {
//                 String token0Address = rs.getString("token0_address");
//                 String token1Address = rs.getString("token1_address");
//                 String pairAddress = rs.getString("pair_address");

//                 // Create and store token0 info
//                 TokenInfo token0Info = new TokenInfo();
//                 token0Info.symbol = rs.getString("token0_symbol");
//                 token0Info.decimals = rs.getInt("token0_decimals");
//                 tokenCache.put(token0Address, token0Info);

//                 // Create and store token1 info
//                 TokenInfo token1Info = new TokenInfo();
//                 token1Info.symbol = rs.getString("token1_symbol");
//                 token1Info.decimals = rs.getInt("token1_decimals");
//                 tokenCache.put(token1Address, token1Info);

//                 tokenCount += 2;

//                 // Initialize USDC pair cache if either token is USDC
//                 if ("USDC".equals(token0Info.symbol) || "USDC".equals(token1Info.symbol)) {
//                     UsdcPairInfo pairInfo = new UsdcPairInfo();
//                     pairInfo.token0 = token0Address;
//                     pairInfo.token1 = token1Address;
//                     usdcPairCache.put(pairAddress, pairInfo);
//                     usdcPairCount++;

//                     // Set USDC pair address for non-USDC token
//                     if ("USDC".equals(token0Info.symbol)) {
//                         token1Info.usdcPairAddress = pairAddress;
//                     } else {
//                         token0Info.usdcPairAddress = pairAddress;
//                     }
//                 }
//             }
//             LOG.info("Loaded {} tokens and {} USDC pairs", tokenCount, usdcPairCount);
//         }
//     }

//     private void saveMetric(PairMetric metric) throws SQLException {
//         String sql = "INSERT INTO twswap_pair_metric (pair_id, time_window, end_time, token0_reserve, token1_reserve, " +
//                     "reserve_usd, token0_volume_usd, token1_volume_usd, volume_usd, txcnt) " +
//                     "VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)";
        
//         try (PreparedStatement stmt = connection.prepareStatement(sql)) {
//             stmt.setLong(1, metric.getPairId());
//             stmt.setString(2, windowName);
//             stmt.setTimestamp(3, metric.getEndTime());
//             stmt.setBigDecimal(4, metric.getToken0Reserve());
//             stmt.setBigDecimal(5, metric.getToken1Reserve());
//             stmt.setBigDecimal(6, metric.getReserveUsd());
//             stmt.setBigDecimal(7, metric.getToken0VolumeUsd());
//             stmt.setBigDecimal(8, metric.getToken1VolumeUsd());
//             stmt.setBigDecimal(9, metric.getVolumeUsd());
//             stmt.setInt(10, metric.getTxcnt());
            
//             stmt.executeUpdate();
//         }
//     }

//     private Long getPairId(String pairAddress) throws SQLException {
//         if (pairIdCache.containsKey(pairAddress)) {
//             return pairIdCache.get(pairAddress);
//         }

//         String sql = "SELECT id FROM twswap_pair WHERE pair_address = ?";
//         try (PreparedStatement stmt = connection.prepareStatement(sql)) {
//             stmt.setString(1, pairAddress);
//             try (ResultSet rs = stmt.executeQuery()) {
//                 if (rs.next()) {
//                     Long pairId = rs.getLong("id");
//                     pairIdCache.put(pairAddress, pairId);
//                     return pairId;
//                 }
//             }
//         }
//         return null;
//     }

//     private String[] getTokenAddresses(String pairAddress) throws SQLException {
//         String sql = "SELECT t0.token_address as token0_address, t1.token_address as token1_address " +
//                     "FROM twswap_pair p " +
//                     "JOIN token t0 ON p.token0_id = t0.id " +
//                     "JOIN token t1 ON p.token1_id = t1.id " +
//                     "WHERE p.pair_address = ?";
        
//         try (PreparedStatement stmt = connection.prepareStatement(sql)) {
//             stmt.setString(1, pairAddress);
//             ResultSet rs = stmt.executeQuery();
//             if (rs.next()) {
//                 return new String[] {
//                     rs.getString("token0_address"),
//                     rs.getString("token1_address")
//                 };
//             }
//         }
//         return null;
//     }

//     private BigDecimal normalizeAmount(BigDecimal amount, int decimals) {
//         return amount.divide(BigDecimal.TEN.pow(decimals), 18, RoundingMode.HALF_UP);
//     }

//     private void processSyncEvent(Event event, PairMetric metric) {
//         try {
//             String pairAddress = event.getContractAddress();
//             if (pairAddress == null) {
//                 LOG.warn("Pair address is null for sync event");
//                 return;
//             }

//             // Get token addresses from cache
//             String[] tokenAddresses = getTokenAddresses(pairAddress);
//             if (tokenAddresses == null) {
//                 LOG.warn("Token addresses not found for pair {}", pairAddress);
//                 return;
//             }

//             // Get token decimals from cache
//             TokenInfo token0Info = tokenCache.get(tokenAddresses[0]);
//             TokenInfo token1Info = tokenCache.get(tokenAddresses[1]);
//             if (token0Info == null || token1Info == null) {
//                 LOG.warn("Token info not found for tokens {} and {}", tokenAddresses[0], tokenAddresses[1]);
//                 return;
//             }

//             // Parse reserve values from event data
//             JsonNode eventData = objectMapper.valueToTree(event.getDecodedArgs());
//             if (eventData == null || eventData.isEmpty()) {
//                 LOG.warn("No decoded args found for sync event");
//                 return;
//             }

//             LOG.debug("Processing sync event data: {}", eventData);

//             BigDecimal reserve0 = new BigDecimal(eventData.get("reserve0").asText());
//             BigDecimal reserve1 = new BigDecimal(eventData.get("reserve1").asText());

//             // Scale down reserves using token decimals
//             reserve0 = normalizeAmount(reserve0, token0Info.decimals);
//             reserve1 = normalizeAmount(reserve1, token1Info.decimals);

//             LOG.info("Scaled reserves for pair {}: reserve0={}, reserve1={}", pairAddress, reserve0, reserve1);

//             // Update metric with scaled reserves
//             metric.setToken0Reserve(reserve0);
//             metric.setToken1Reserve(reserve1);

//             // Calculate and update reserve USD
//             calculateReserveUsd(metric, pairAddress);
//             LOG.debug("Updated reserves and USD value for pair {} - reserve0={}, reserve1={}, reserveUsd={}", 
//                 pairAddress, reserve0, reserve1, metric.getReserveUsd());

//             // Store sync event in USDC pair cache if applicable
//             UsdcPairInfo usdcPairInfo = usdcPairCache.get(pairAddress);
//             if (usdcPairInfo != null) {
//                 usdcPairInfo.lastSyncEvent = event;
//                 LOG.debug("Updated USDC pair cache for pair {}", pairAddress);
//             }

//         } catch (Exception e) {
//             LOG.error("Error processing sync event: {}", e.getMessage(), e);
//             // Set reserves to zero in case of error
//             metric.setToken0Reserve(BigDecimal.ZERO);
//             metric.setToken1Reserve(BigDecimal.ZERO);
//             metric.setReserveUsd(BigDecimal.ZERO);
//         }
//     }

//     private void processSwapEvent(Event event, PairMetric metric) {
//         try {
//             JsonNode data = objectMapper.valueToTree(event.getDecodedArgs());
//             if (data == null || data.isEmpty()) {
//                 LOG.warn("No decoded args found for swap event");
//                 return;
//             }

//             BigDecimal amount0In = new BigDecimal(data.get("amount0In").asText());
//             BigDecimal amount1In = new BigDecimal(data.get("amount1In").asText());
//             BigDecimal amount0Out = new BigDecimal(data.get("amount0Out").asText());
//             BigDecimal amount1Out = new BigDecimal(data.get("amount1Out").asText());

//             LOG.debug("Processing swap amounts - in: ({}, {}), out: ({}, {})", 
//                 amount0In, amount1In, amount0Out, amount1Out);

//             // Calculate and update volumes in USD
//             calculateSwapVolumes(metric, amount0In, amount1In, amount0Out, amount1Out);

//         } catch (Exception e) {
//             LOG.error("Error processing swap event: {}", e.getMessage(), e);
//         }
//     }

//     private void processMintEvent(PairMetric metric, JsonNode data) {
//         try {
//             BigDecimal amount0 = new BigDecimal(data.get("amount0").asText());
//             BigDecimal amount1 = new BigDecimal(data.get("amount1").asText());

//             LOG.debug("Processing mint amounts - token0={}, token1={}", amount0, amount1);

//             String pairAddress = getPairAddress(metric.getPairId());
//             if (pairAddress == null) {
//                 LOG.warn("Pair address not found for ID: {}", metric.getPairId());
//                 return;
//             }

//             String[] tokenAddresses = getTokenAddresses(pairAddress);
//             if (tokenAddresses == null) {
//                 LOG.warn("Token addresses not found for pair: {}", pairAddress);
//                 return;
//             }

//             TokenInfo token0Info = tokenCache.get(tokenAddresses[0]);
//             TokenInfo token1Info = tokenCache.get(tokenAddresses[1]);

//             if (token0Info != null && token1Info != null) {
//                 // Scale down amounts by token decimals
//                 BigDecimal scaledAmount0 = normalizeAmount(amount0, token0Info.decimals);
//                 BigDecimal scaledAmount1 = normalizeAmount(amount1, token1Info.decimals);

//                 // Update volumes with scaled amounts
//                 metric.setToken0VolumeUsd(metric.getToken0VolumeUsd().add(scaledAmount0));
//                 metric.setToken1VolumeUsd(metric.getToken1VolumeUsd().add(scaledAmount1));

//                 LOG.debug("Updated scaled volumes - token0={}, token1={}", 
//                     metric.getToken0VolumeUsd(), metric.getToken1VolumeUsd());
//             } else {
//                 LOG.warn("Token info not found for addresses: {} {}", tokenAddresses[0], tokenAddresses[1]);
//                 metric.setToken0VolumeUsd(BigDecimal.ZERO);
//                 metric.setToken1VolumeUsd(BigDecimal.ZERO);
//             }
//         } catch (Exception e) {
//             LOG.error("Error processing mint volumes: {}", e.getMessage());
//             metric.setToken0VolumeUsd(BigDecimal.ZERO);
//             metric.setToken1VolumeUsd(BigDecimal.ZERO);
//         }
//     }

//     private void processBurnEvent(PairMetric metric, JsonNode data) {
//         try {
//             BigDecimal amount0 = new BigDecimal(data.get("amount0").asText());
//             BigDecimal amount1 = new BigDecimal(data.get("amount1").asText());

//             LOG.debug("Processing burn amounts - token0={}, token1={}", amount0, amount1);

//             String pairAddress = getPairAddress(metric.getPairId());
//             if (pairAddress == null) {
//                 LOG.warn("Pair address not found for ID: {}", metric.getPairId());
//                 return;
//             }

//             String[] tokenAddresses = getTokenAddresses(pairAddress);
//             if (tokenAddresses == null) {
//                 LOG.warn("Token addresses not found for pair: {}", pairAddress);
//                 return;
//             }

//             TokenInfo token0Info = tokenCache.get(tokenAddresses[0]);
//             TokenInfo token1Info = tokenCache.get(tokenAddresses[1]);

//             if (token0Info != null && token1Info != null) {
//                 // Scale down amounts by token decimals
//                 BigDecimal scaledAmount0 = normalizeAmount(amount0, token0Info.decimals);
//                 BigDecimal scaledAmount1 = normalizeAmount(amount1, token1Info.decimals);

//                 // Update volumes with scaled amounts
//                 metric.setToken0VolumeUsd(metric.getToken0VolumeUsd().add(scaledAmount0));
//                 metric.setToken1VolumeUsd(metric.getToken1VolumeUsd().add(scaledAmount1));

//                 LOG.debug("Updated scaled volumes - token0={}, token1={}", 
//                     metric.getToken0VolumeUsd(), metric.getToken1VolumeUsd());
//             } else {
//                 LOG.warn("Token info not found for addresses: {} {}", tokenAddresses[0], tokenAddresses[1]);
//                 metric.setToken0VolumeUsd(BigDecimal.ZERO);
//                 metric.setToken1VolumeUsd(BigDecimal.ZERO);
//             }
//         } catch (Exception e) {
//             LOG.error("Error processing burn volumes: {}", e.getMessage());
//             metric.setToken0VolumeUsd(BigDecimal.ZERO);
//             metric.setToken1VolumeUsd(BigDecimal.ZERO);
//         }
//     }

//     private void calculateReserveUsd(PairMetric metric, String pairAddress) {
//         try {
//             String[] tokenAddresses = getTokenAddresses(pairAddress);
//             if (tokenAddresses == null) {
//                 LOG.warn("Token addresses not found for pair: {}", pairAddress);
//                 metric.setReserveUsd(BigDecimal.ZERO);
//                 return;
//             }

//             TokenInfo token0Info = tokenCache.get(tokenAddresses[0]);
//             TokenInfo token1Info = tokenCache.get(tokenAddresses[1]);

//             if (token0Info == null || token1Info == null) {
//                 LOG.warn("Token info not found for addresses: {} {}", tokenAddresses[0], tokenAddresses[1]);
//                 metric.setReserveUsd(BigDecimal.ZERO);
//                 return;
//             }

//             // Calculate USD value
//             BigDecimal reserveUsd = BigDecimal.ZERO;

//             // If token0 is USDC
//             if ("USDC".equals(token0Info.symbol)) {
//                 reserveUsd = metric.getToken0Reserve().multiply(new BigDecimal("2"));
//                 LOG.debug("Calculated USD value using token0 (USDC): {}", reserveUsd);
//             }
//             // If token1 is USDC
//             else if ("USDC".equals(token1Info.symbol)) {
//                 reserveUsd = metric.getToken1Reserve().multiply(new BigDecimal("2"));
//                 LOG.debug("Calculated USD value using token1 (USDC): {}", reserveUsd);
//             }
//             // If neither token is USDC, try to find USDC price through token0's USDC pair
//             else if (token0Info.usdcPairAddress != null) {
//                 UsdcPairInfo usdcPairInfo = usdcPairCache.get(token0Info.usdcPairAddress);
//                 if (usdcPairInfo != null && usdcPairInfo.lastSyncEvent != null) {
//                     try {
//                         JsonNode syncData = objectMapper.valueToTree(usdcPairInfo.lastSyncEvent.getDecodedArgs());
//                         if (syncData != null && !syncData.isEmpty()) {
//                             BigDecimal usdcReserve = normalizeAmount(new BigDecimal(syncData.get("reserve1").asText()), USDC_DECIMALS.intValue());
//                             BigDecimal tokenReserve = normalizeAmount(new BigDecimal(syncData.get("reserve0").asText()), token0Info.decimals);
//                             if (tokenReserve.compareTo(BigDecimal.ZERO) > 0) {
//                                 BigDecimal priceInUsdc = usdcReserve.divide(tokenReserve, 18, RoundingMode.HALF_UP);
//                                 BigDecimal token0Value = metric.getToken0Reserve().multiply(priceInUsdc);
//                                 BigDecimal token1Value = metric.getToken1Reserve().multiply(priceInUsdc);
//                                 reserveUsd = token0Value.add(token1Value);
//                                 LOG.debug("Calculated USD value using token0's USDC pair: {}", reserveUsd);
//                             }
//                         }
//                     } catch (Exception e) {
//                         LOG.error("Error calculating USD value using token0's USDC pair", e);
//                     }
//                 }
//             }

//             metric.setReserveUsd(reserveUsd);
//             LOG.debug("Set reserve USD for pair {}: {}", pairAddress, reserveUsd);

//         } catch (SQLException e) {
//             LOG.error("Error calculating reserve USD for pair {}: {}", pairAddress, e.getMessage());
//             metric.setReserveUsd(BigDecimal.ZERO);
//         } catch (Exception e) {
//             LOG.error("Unexpected error calculating reserve USD for pair {}: {}", pairAddress, e.getMessage());
//             metric.setReserveUsd(BigDecimal.ZERO);
//         }
//     }

//     private void calculateSwapVolumes(PairMetric metric, BigDecimal amount0In, BigDecimal amount1In,
//             BigDecimal amount0Out, BigDecimal amount1Out) {
//         try {
//             String pairAddress = getPairAddress(metric.getPairId());
//             if (pairAddress == null) {
//                 LOG.warn("Pair address not found for ID: {}", metric.getPairId());
//                 return;
//             }

//             String[] tokenAddresses = getTokenAddresses(pairAddress);
//             if (tokenAddresses == null) {
//                 LOG.warn("Token addresses not found for pair: {}", pairAddress);
//                 return;
//             }

//             TokenInfo token0Info = tokenCache.get(tokenAddresses[0]);
//             TokenInfo token1Info = tokenCache.get(tokenAddresses[1]);

//             if (token0Info == null || token1Info == null) {
//                 LOG.warn("Token info not found for addresses: {} {}", tokenAddresses[0], tokenAddresses[1]);
//                 return;
//             }

//             // Scale down amounts by token decimals
//             BigDecimal scaledAmount0In = normalizeAmount(amount0In, token0Info.decimals);
//             BigDecimal scaledAmount0Out = normalizeAmount(amount0Out, token0Info.decimals);
//             BigDecimal scaledAmount1In = normalizeAmount(amount1In, token1Info.decimals);
//             BigDecimal scaledAmount1Out = normalizeAmount(amount1Out, token1Info.decimals);

//             // Calculate total volumes for each token
//             BigDecimal token0Volume = scaledAmount0In.add(scaledAmount0Out);
//             BigDecimal token1Volume = scaledAmount1In.add(scaledAmount1Out);

//             // Update token volumes
//             metric.setToken0VolumeUsd(metric.getToken0VolumeUsd().add(token0Volume));
//             metric.setToken1VolumeUsd(metric.getToken1VolumeUsd().add(token1Volume));

//             // Calculate USD volume based on USDC pair if available
//             BigDecimal volumeUsd = BigDecimal.ZERO;
            
//             // If token0 is USDC
//             if ("USDC".equals(token0Info.symbol)) {
//                 volumeUsd = token0Volume;
//             }
//             // If token1 is USDC
//             else if ("USDC".equals(token1Info.symbol)) {
//                 volumeUsd = token1Volume;
//             }
//             // If neither token is USDC, try to find USDC price through token0's USDC pair
//             else if (token0Info.usdcPairAddress != null) {
//                 UsdcPairInfo usdcPairInfo = usdcPairCache.get(token0Info.usdcPairAddress);
//                 if (usdcPairInfo != null && usdcPairInfo.lastSyncEvent != null) {
//                     JsonNode syncData = objectMapper.valueToTree(usdcPairInfo.lastSyncEvent.getDecodedArgs());
//                     if (syncData != null && !syncData.isEmpty()) {
//                         BigDecimal usdcReserve = normalizeAmount(new BigDecimal(syncData.get("reserve1").asText()), USDC_DECIMALS.intValue());
//                         BigDecimal tokenReserve = normalizeAmount(new BigDecimal(syncData.get("reserve0").asText()), token0Info.decimals);
//                         if (tokenReserve.compareTo(BigDecimal.ZERO) > 0) {
//                             BigDecimal priceInUsdc = usdcReserve.divide(tokenReserve, 18, RoundingMode.HALF_UP);
//                             volumeUsd = token0Volume.multiply(priceInUsdc);
//                         }
//                     }
//                 }
//             }

//             // Update total volume in USD
//             metric.setVolumeUsd(metric.getVolumeUsd().add(volumeUsd));
            
//             LOG.debug("Updated volumes for pair {} - token0={}, token1={}, usd={}", 
//                 pairAddress, metric.getToken0VolumeUsd(), metric.getToken1VolumeUsd(), metric.getVolumeUsd());

//         } catch (Exception e) {
//             LOG.error("Error calculating swap volumes: {}", e.getMessage(), e);
//         }
//     }

//     private String getPairAddress(Long pairId) throws SQLException {
//         String sql = "SELECT pair_address FROM twswap_pair WHERE id = ?";
//         try (PreparedStatement stmt = connection.prepareStatement(sql)) {
//             stmt.setLong(1, pairId);
//             try (ResultSet rs = stmt.executeQuery()) {
//                 if (rs.next()) {
//                     return rs.getString("pair_address");
//                 }
//             }
//         }
//         return null;
//     }
// } 