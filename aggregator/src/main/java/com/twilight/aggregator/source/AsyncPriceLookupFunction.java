package com.twilight.aggregator.source;

import java.math.BigDecimal;
import java.util.Collections;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;

import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.async.ResultFuture;
import org.apache.flink.streaming.api.functions.async.RichAsyncFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.twilight.aggregator.config.RedisConfig;

import redis.clients.jedis.Jedis;

public class AsyncPriceLookupFunction extends RichAsyncFunction<String, BigDecimal> {
    private static final long serialVersionUID = 1L;
    private static final Logger log = LoggerFactory.getLogger(AsyncPriceLookupFunction.class);
    private static final String PRICE_KEY_PREFIX = "token_price:";
    private transient RedisConfig redisConfig;
    private transient volatile ExecutorService executor;
    private static final int THREAD_POOL_SIZE = 10;

    private synchronized void initializeIfNeeded() {
        if (redisConfig == null) {
            redisConfig = RedisConfig.getInstance();
            if (redisConfig == null) {
                throw new IllegalStateException("Failed to initialize RedisConfig");
            }
        }
        if (executor == null || executor.isShutdown()) {
            executor = Executors.newFixedThreadPool(THREAD_POOL_SIZE, r -> {
                Thread t = new Thread(r, "async-price-lookup-thread");
                t.setDaemon(true);
                return t;
            });
            log.info("Initialized thread pool with size {}", THREAD_POOL_SIZE);
        }
    }

    @Override
    public void open(Configuration parameters) {
        initializeIfNeeded();
        log.info("AsyncPriceLookupFunction opened with thread pool size {}: {}", THREAD_POOL_SIZE, redisConfig);
    }

    @Override
    public void asyncInvoke(String tokenAddress, final ResultFuture<BigDecimal> resultFuture) {
        // 确保在调用前重新检查初始化状态
        initializeIfNeeded();
        
        // 创建本地变量引用，避免并发问题
        final ExecutorService currentExecutor = executor;
        if (currentExecutor == null || currentExecutor.isShutdown()) {
            log.error("Thread pool is not available for token {}", tokenAddress);
            resultFuture.complete(Collections.singleton(BigDecimal.ZERO));
            return;
        }

        CompletableFuture.supplyAsync(() -> {
            try (Jedis jedis = redisConfig.getResource()) {
                return getTokenPrice(jedis, tokenAddress);
            } catch (Exception e) {
                log.error("Error getting price for token {}", tokenAddress, e);
                return BigDecimal.ZERO;
            }
        }, currentExecutor).thenAccept(price -> {
            try {
                resultFuture.complete(Collections.singleton(price));
            } catch (Exception e) {
                log.error("Error completing future for token {}", tokenAddress, e);
                resultFuture.complete(Collections.singleton(BigDecimal.ZERO));
            }
        }).exceptionally(throwable -> {
            log.error("Async operation failed for token {}", tokenAddress, throwable);
            resultFuture.complete(Collections.singleton(BigDecimal.ZERO));
            return null;
        });
    }

    private BigDecimal getTokenPrice(Jedis jedis, String tokenAddress) {
        if (tokenAddress == null) return BigDecimal.ZERO;
        try {
            String price = jedis.get(PRICE_KEY_PREFIX + tokenAddress.toLowerCase());
            log.info("getTokenPrice for {}: {}", tokenAddress, price);
            return price != null ? new BigDecimal(price) : BigDecimal.ZERO;
        } catch (Exception e) {
            log.error("Error getting price for token {}", tokenAddress, e);
            return BigDecimal.ZERO;
        }
    }

    @Override
    public void close() throws Exception {
        ExecutorService currentExecutor = executor;
        if (currentExecutor != null && !currentExecutor.isShutdown()) {
            try {
                currentExecutor.shutdown();
                if (!currentExecutor.awaitTermination(5, TimeUnit.SECONDS)) {
                    currentExecutor.shutdownNow();
                }
            } catch (InterruptedException e) {
                currentExecutor.shutdownNow();
                Thread.currentThread().interrupt();
            } finally {
                executor = null;
            }
        }

        if (redisConfig != null) {
            try {
                redisConfig.close();
            } finally {
                redisConfig = null;
            }
        }
    }
} 
