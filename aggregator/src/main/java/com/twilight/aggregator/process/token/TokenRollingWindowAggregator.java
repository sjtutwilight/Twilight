package com.twilight.aggregator.process.token;

import org.apache.flink.api.common.functions.AggregateFunction;
import com.twilight.aggregator.model.Token;
import com.twilight.aggregator.model.TokenRollingMetric;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class TokenRollingWindowAggregator
        implements AggregateFunction<Token, TokenRollingMetric, TokenRollingMetric> {

    @Override
    public TokenRollingMetric createAccumulator() {
        return new TokenRollingMetric(null, 0.0, 0.0);
    }

    @Override
    public TokenRollingMetric add(Token token, TokenRollingMetric accumulator) {
        if (accumulator.getTokenId() == null) {
            accumulator.setTokenId(token.getTokenId());
        }

        // Calculate volume
        double volumeUsd = token.getAmount() * token.getTokenPriceUsd();
        accumulator.setVolumeUsd(accumulator.getVolumeUsd() + volumeUsd);

        // Update token price
        accumulator.setTokenPriceUsd(token.getTokenPriceUsd());

        return accumulator;
    }

    @Override
    public TokenRollingMetric getResult(TokenRollingMetric accumulator) {
        Logger log = LoggerFactory.getLogger(TokenRollingWindowAggregator.class);
        // Ensure end_time is set to current time rounded to seconds
        accumulator.setEndTime((System.currentTimeMillis() / 1000) * 1000);
        log.info("TokenRollingWindowAggregator: Get result: {}", accumulator);
        return accumulator;
    }

    @Override
    public TokenRollingMetric merge(TokenRollingMetric a, TokenRollingMetric b) {
        TokenRollingMetric merged = new TokenRollingMetric(a.getTokenId(), a.getVolumeUsd() + b.getVolumeUsd(),
                b.getTokenPriceUsd());

        // Set the latest end_time
        merged.setEndTime(Math.max(a.getEndTime(), b.getEndTime()));

        return merged;
    }
}