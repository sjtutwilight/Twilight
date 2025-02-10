package com.twilight.aggregator.process;

import java.util.Map;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;

import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.windowing.ProcessWindowFunction;
import org.apache.flink.streaming.api.windowing.windows.TimeWindow;
import org.apache.flink.util.Collector;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.twilight.aggregator.config.RedisConfig;
import com.twilight.aggregator.model.Pair;
import com.twilight.aggregator.model.PairMetric;
import com.twilight.aggregator.utils.EthereumUtils;

import redis.clients.jedis.Jedis;

public class PairWindowProcessor extends ProcessWindowFunction<Pair, PairMetric, String, TimeWindow> {
    private static final long serialVersionUID = 1L;
    private final String windowName;
    private transient Logger log;
    private String usdcAddress;
    private transient RedisConfig redisConfig;
    private transient ExecutorService executor;
    private static final int THREAD_POOL_SIZE = 2;

    public PairWindowProcessor(String windowName) {
        this.windowName = windowName;
    }

    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        this.log = LoggerFactory.getLogger(PairWindowProcessor.class);
        this.redisConfig = RedisConfig.getInstance();
        this.usdcAddress = "0x74A6379d012ce53E3b0718C05dD72a3De87F0c6a";
        this.executor = Executors.newFixedThreadPool(THREAD_POOL_SIZE, r -> {
            Thread t = new Thread(r, "pair-price-update-thread");
            t.setDaemon(true);
            return t;
        });
    }

    @Override
    public void close() throws Exception {
        if (executor != null) {
            executor.shutdown();
            if (!executor.awaitTermination(5, TimeUnit.SECONDS)) {
                executor.shutdownNow();
            }
        }
        super.close();
    }

    @Override
    public void process(String key, Context context, Iterable<Pair> elements, Collector<PairMetric> out) {
        PairMetric metric = new PairMetric();
        metric.setTimeWindow(windowName);
        metric.setEndTime(context.window().getEnd());

        for (Pair pair : elements) {
            updateMetric(metric, pair);
            metric.setPairId(pair.getPairId());

        }

        out.collect(metric);
    }

    private void updateMetric(PairMetric metric, Pair pair) {
        switch (pair.getEventName()) {
            case "Sync":
                double token0Reserve = EthereumUtils.convertWeiToEth(pair.getReserve0());
                double token1Reserve = EthereumUtils.convertWeiToEth(pair.getReserve1());
                metric.setToken0Reserve(token0Reserve);
                metric.setToken1Reserve(token1Reserve);
                metric.setReserveUsd(calculateReserveUsd(pair, token0Reserve, token1Reserve));

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

    private double calculateReserveUsd(Pair pair, double token0Reserve, double token1Reserve) {
        return token0Reserve * pair.getToken0PriceUsd() + token1Reserve * pair.getToken1PriceUsd();
    }

    private void updateTokenPriceInRedis(Pair pair, double token0Reserve, double token1Reserve) {
        String token0Address = pair.getToken0Address().toLowerCase();
        String token1Address = pair.getToken1Address().toLowerCase();

        if (!token0Address.equals(usdcAddress) && !token1Address.equals(usdcAddress)) {
            return; // Skip if neither token is USDC
        }

        CompletableFuture.runAsync(() -> {
            try (Jedis jedis = redisConfig.getResource()) {
                if (token0Address.equals(usdcAddress)) {
                    // If token0 is USDC, update token1's price
                    double tokenPrice = token0Reserve / token1Reserve;
                    jedis.set("token_price:" + token1Address, String.valueOf(tokenPrice));
                    log.debug("Updated price for token {}: {} USD", token1Address, tokenPrice);
                } else if (token1Address.equals(usdcAddress)) {
                    // If token1 is USDC, update token0's price
                    double tokenPrice = token1Reserve / token0Reserve;
                    jedis.set("token_price:" + token0Address, String.valueOf(tokenPrice));
                    log.debug("Updated price for token {}: {} USD", token0Address, tokenPrice);
                }
            } catch (Exception e) {
                log.error("Failed to update token price in Redis for pair {}", pair.getPairAddress());
            }
        }, executor).orTimeout(2, TimeUnit.SECONDS)
                .exceptionally(throwable -> {
                    log.error("Async token price update failed for pair {}", pair.getPairAddress());
                    return null;
                });
    }

    private void updateSwapMetrics(PairMetric metric, Pair pair) {
        double amount0In = EthereumUtils.convertWeiToEth(pair.getAmount0In());
        double amount0Out = EthereumUtils.convertWeiToEth(pair.getAmount0Out());
        double amount1In = EthereumUtils.convertWeiToEth(pair.getAmount1In());
        double amount1Out = EthereumUtils.convertWeiToEth(pair.getAmount1Out());

        double amount0Total = amount0In + amount0Out;
        double amount1Total = amount1In + amount1Out;

        double volume0Usd = amount0Total * pair.getToken0PriceUsd();
        double volume1Usd = amount1Total * pair.getToken1PriceUsd();

        metric.setToken0VolumeUsd(metric.getToken0VolumeUsd() + volume0Usd);
        metric.setToken1VolumeUsd(metric.getToken1VolumeUsd() + volume1Usd);
        metric.setVolumeUsd(metric.getVolumeUsd() + volume0Usd + volume1Usd);
    }
}