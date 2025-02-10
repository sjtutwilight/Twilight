package com.twilight.aggregator.source;

import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.async.ResultFuture;
import org.apache.flink.streaming.api.functions.async.RichAsyncFunction;
import redis.clients.jedis.Jedis;
import com.twilight.aggregator.config.RedisConfig;

import java.io.Serializable;
import java.util.Collections;

public class AsyncPriceLookupFunction extends RichAsyncFunction<String, Double> {
    private static final long serialVersionUID = 1L;
    private transient RedisConfig redisConfig;

    @Override
    public void open(Configuration parameters) {
        redisConfig = RedisConfig.getInstance();
    }

    @Override
    public void close() {
        if (redisConfig != null) {
            redisConfig.close();
        }
    }

    @Override
    public void asyncInvoke(String tokenAddress, final ResultFuture<Double> resultFuture) {
        try (Jedis jedis = redisConfig.getResource()) {
            String priceStr = jedis.get("token_price:" + tokenAddress.toLowerCase());
            if (priceStr == null) {
                resultFuture.complete(Collections.singleton(0.0));
                return;
            }

            try {
                double price = Double.parseDouble(priceStr);
                resultFuture.complete(Collections.singleton(price));
            } catch (NumberFormatException e) {
                resultFuture.complete(Collections.singleton(0.0));
            }
        } catch (Exception e) {
            resultFuture.complete(Collections.singleton(0.0));
        }
    }

    @Override
    public void timeout(String input, ResultFuture<Double> resultFuture) {
        resultFuture.complete(Collections.singleton(0.0));
    }
}
