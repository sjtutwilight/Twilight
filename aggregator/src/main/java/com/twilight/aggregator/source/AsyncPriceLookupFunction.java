package com.twilight.aggregator.source;

import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.async.ResultFuture;
import org.apache.flink.streaming.api.functions.async.RichAsyncFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.twilight.aggregator.config.RedisAsyncConfig;
import com.twilight.aggregator.model.ProcessEvent;
import io.lettuce.core.api.async.RedisAsyncCommands;
import io.lettuce.core.KeyValue;
import io.lettuce.core.RedisFuture;

import java.util.List;

public class AsyncPriceLookupFunction extends RichAsyncFunction<ProcessEvent, ProcessEvent> {
    private static final Logger log = LoggerFactory.getLogger(AsyncPriceLookupFunction.class);
    private static final long serialVersionUID = 1L;
    private transient RedisAsyncConfig redisConfig;
    private transient RedisAsyncCommands<String, String> asyncCommands;

    @Override
    public void open(Configuration parameters) {
        redisConfig = RedisAsyncConfig.getInstance();
        asyncCommands = redisConfig.getAsyncCommands();
        log.info("Initialized AsyncPriceLookupFunction");
    }

    @Override
    public void close() {
        // Redis connection is managed by RedisAsyncConfig singleton
    }

    @Override
    public void asyncInvoke(ProcessEvent event, final ResultFuture<ProcessEvent> resultFuture) {
        try {
            if (event.getToken0Id() == null || event.getToken1Id() == null) {
                // 如果token ID为空，直接返回原始事件
                log.debug("Token ID is null, skipping price lookup: {}", event);
                resultFuture.complete(java.util.Collections.singleton(event));
                return;
            }

            // 构建Redis查询
            RedisFuture<List<KeyValue<String, String>>> future = asyncCommands.mget(
                    "token_price:" + event.getToken0Address().toLowerCase(),
                    "token_price:" + event.getToken1Address().toLowerCase());
            // 处理异步结果
            future.thenAccept(prices -> {
                try {
                    if (prices == null || prices.isEmpty()) {
                        log.debug("No prices found for tokens: {}, {}", event.getToken0Id(), event.getToken1Id());
                        event.setToken0PriceUsd(0.0);
                        event.setToken1PriceUsd(0.0);
                        resultFuture.complete(java.util.Collections.singleton(event));
                        return;
                    }

                    // 处理token0价格
                    KeyValue<String, String> token0KV = prices.get(0);
                    String token0Price = token0KV != null ? token0KV.getValue() : null;
                    event.setToken0PriceUsd(token0Price != null ? Double.parseDouble(token0Price) : 0.0);

                    // 处理token1价格
                    KeyValue<String, String> token1KV = prices.get(1);
                    String token1Price = token1KV != null ? token1KV.getValue() : null;
                    event.setToken1PriceUsd(token1Price != null ? Double.parseDouble(token1Price) : 0.0);

                    resultFuture.complete(java.util.Collections.singleton(event));
                } catch (NumberFormatException e) {
                    log.warn("Invalid price format: {}", e.getMessage());
                    event.setToken0PriceUsd(0.0);
                    event.setToken1PriceUsd(0.0);
                    resultFuture.complete(java.util.Collections.singleton(event));
                }
            })
                    .exceptionally(throwable -> {
                        log.warn("Error querying Redis: {}", throwable.getMessage());
                        event.setToken0PriceUsd(0.0);
                        event.setToken1PriceUsd(0.0);
                        resultFuture.complete(java.util.Collections.singleton(event));
                        return null;
                    });
        } catch (Exception e) {
            log.warn("Error in asyncInvoke: {}", e.getMessage());
            event.setToken0PriceUsd(0.0);
            event.setToken1PriceUsd(0.0);
            resultFuture.complete(java.util.Collections.singleton(event));
        }
    }

    @Override
    public void timeout(ProcessEvent event, ResultFuture<ProcessEvent> resultFuture) {
        log.warn("Async operation timed out for event: {}", event);
        event.setToken0PriceUsd(0.0);
        event.setToken1PriceUsd(0.0);
        resultFuture.complete(java.util.Collections.singleton(event));
    }
}
