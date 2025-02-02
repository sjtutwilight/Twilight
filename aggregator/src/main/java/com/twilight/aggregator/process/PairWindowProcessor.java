package com.twilight.aggregator.process;

import java.math.BigDecimal;
import java.sql.Timestamp;
import java.util.Map;

import org.apache.flink.streaming.api.functions.windowing.ProcessWindowFunction;
import org.apache.flink.streaming.api.windowing.windows.TimeWindow;
import org.apache.flink.util.Collector;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.twilight.aggregator.config.FlinkConfig;
import com.twilight.aggregator.config.RedisConfig;
import com.twilight.aggregator.model.Pair;
import com.twilight.aggregator.model.PairMetric;
import com.twilight.aggregator.utils.EthereumUtils;

import redis.clients.jedis.Jedis;

public class PairWindowProcessor extends ProcessWindowFunction<Pair, PairMetric, String, TimeWindow> {
    private final String windowName;
    private final Logger log = LoggerFactory.getLogger(PairWindowProcessor.class);
    private final String usdcAddress;
    private final RedisConfig redisConfig;

    public PairWindowProcessor(String windowName) {
        this.windowName = windowName;
        this.redisConfig = RedisConfig.getInstance();
        this.usdcAddress = FlinkConfig.getInstance().getUsdcAddress().toLowerCase();
    }

    @Override
    public void process(String key, Context context, Iterable<Pair> elements, Collector<PairMetric> out) {
        PairMetric metric = new PairMetric();
        metric.setTimeWindow(windowName);
        metric.setEndTime(new Timestamp(context.window().getEnd()));
        
        for (Pair pair : elements) {
            updateMetric(metric, pair);
            metric.setPairId(pair.getPairId());
            //metric任意元素为空
            if (metric.getToken0Reserve() == null || metric.getToken1Reserve() == null || metric.getReserveUsd() == null||metric.getToken0VolumeUsd() == null||metric.getToken1VolumeUsd() == null||metric.getVolumeUsd() == null||metric.getTxcnt() == 0) {
                log.error("Pair metric is null for contract address: {}", pair.getPairAddress());
                return;
            }
        }

        out.collect(metric);
    }

    private void updateMetric(PairMetric metric, Pair pair) {
        switch (pair.getEventName()) {
            case "Sync":
                BigDecimal token0Reserve = EthereumUtils.convertWeiToEth(pair.getEventArgs().get("reserve0").toString());
                BigDecimal token1Reserve = EthereumUtils.convertWeiToEth(pair.getEventArgs().get("reserve1").toString());
                metric.setToken0Reserve(token0Reserve);
                metric.setToken1Reserve(token1Reserve);
                metric.setReserveUsd(calculateReserveUsd(pair,token0Reserve,token1Reserve));
                
                // Update token price in Redis if pair contains USDC
                updateTokenPriceInRedis(pair, token0Reserve, token1Reserve);
                
                metric.incrementTxCount();
                break;
            case "Swap":
                updateSwapMetrics(metric, pair);
                metric.incrementTxCount();
                break;
            case "Mint":
            case "Burn":
                metric.incrementTxCount();
                break;
        }
    }

    private BigDecimal calculateReserveUsd(Pair pair,BigDecimal token0Reserve,BigDecimal token1Reserve) {
        return token0Reserve.multiply(pair.getToken0PriceUsd())
            .add(token1Reserve.multiply(pair.getToken1PriceUsd()));
    }

    private void updateTokenPriceInRedis(Pair pair, BigDecimal token0Reserve, BigDecimal token1Reserve) {
        try (Jedis jedis = redisConfig.getResource()) {
            // Check if either token is USDC
            String token0Address = pair.getToken0Address().toLowerCase();
            String token1Address = pair.getToken1Address().toLowerCase();

            if (token0Address.equals(usdcAddress)) {
                // If token0 is USDC, update token1's price
                // Price = USDC reserve / token reserve
                BigDecimal tokenPrice = token0Reserve.divide(token1Reserve, 18, BigDecimal.ROUND_HALF_UP);
                jedis.set("token_price:" + token1Address, tokenPrice.toString());
                log.info("Updated price for token {}: {} USD", token1Address, tokenPrice);
            } else if (token1Address.equals(usdcAddress)) {
                // If token1 is USDC, update token0's price
                // Price = USDC reserve / token reserve
                BigDecimal tokenPrice = token1Reserve.divide(token0Reserve, 18, BigDecimal.ROUND_HALF_UP);
                jedis.set("token_price:" + token0Address, tokenPrice.toString());
                log.info("Updated price for token {}: {} USD", token0Address, tokenPrice);
            }
        } catch (Exception e) {
            log.error("Error updating token price in Redis: {}", e.getMessage());
        }
    }

    private void updateSwapMetrics(PairMetric metric, Pair pair) {
        Map<String, Object> args = pair.getEventArgs();
        BigDecimal amount0Total = EthereumUtils.convertWeiToEth(args.get("amount0In").toString())
            .add(EthereumUtils.convertWeiToEth(args.get("amount0Out").toString()));
        BigDecimal amount1Total = EthereumUtils.convertWeiToEth(args.get("amount1In").toString())
            .add(EthereumUtils.convertWeiToEth(args.get("amount1Out").toString()));

        BigDecimal volume0Usd = amount0Total.multiply(pair.getToken0PriceUsd());
        BigDecimal volume1Usd = amount1Total.multiply(pair.getToken1PriceUsd());

        metric.setToken0VolumeUsd(metric.getToken0VolumeUsd().add(volume0Usd));
        metric.setToken1VolumeUsd(metric.getToken1VolumeUsd().add(volume1Usd));
        metric.setVolumeUsd(metric.getVolumeUsd().add(volume0Usd.add(volume1Usd)));
    }
} 