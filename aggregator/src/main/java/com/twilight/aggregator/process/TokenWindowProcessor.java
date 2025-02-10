package com.twilight.aggregator.process;

import java.util.HashSet;
import java.util.Set;

import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.windowing.ProcessWindowFunction;
import org.apache.flink.streaming.api.windowing.windows.TimeWindow;
import org.apache.flink.util.Collector;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.twilight.aggregator.model.Token;
import com.twilight.aggregator.model.TokenMetric;

public class TokenWindowProcessor extends ProcessWindowFunction<Token, TokenMetric, String, TimeWindow> {
    private static final long serialVersionUID = 1L;
    private final String windowName;
    private transient Logger log;

    public TokenWindowProcessor(String windowName) {
        this.windowName = windowName;
    }

    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        this.log = LoggerFactory.getLogger(TokenWindowProcessor.class);
    }

    @Override
    public void process(String key, Context context, Iterable<Token> elements, Collector<TokenMetric> out) {
        TokenMetric metric = new TokenMetric();
        metric.setTimeWindow(windowName);
        metric.setEndTime(context.window().getEnd());

        Set<String> buyers = new HashSet<>();
        Set<String> sellers = new HashSet<>();
        double buyVolumeUsd = 0.0;
        double sellVolumeUsd = 0.0;

        for (Token token : elements) {
            metric.setTokenId(token.getTokenId());
            double volumeUsd = token.getAmount() * token.getTokenPriceUsd();
            metric.setVolumeUsd(metric.getVolumeUsd() + volumeUsd);
            metric.setTokenPriceUsd(token.getTokenPriceUsd()); // 使用最新价格

            if (token.isBuyOrSell()) {
                buyers.add(token.getFromAddress());
                buyVolumeUsd += volumeUsd;
                metric.incrementBuyCount();
            } else {
                sellers.add(token.getFromAddress());
                sellVolumeUsd += volumeUsd;
                metric.incrementSellCount();
            }

            metric.incrementTxCount();
        }

        metric.setBuyersCount(buyers.size());
        metric.setSellersCount(sellers.size());
        metric.calculateMakersCount();
        metric.setBuyVolumeUsd(buyVolumeUsd);
        metric.setSellVolumeUsd(sellVolumeUsd);
        metric.setBuyPressureUsd(buyVolumeUsd - sellVolumeUsd);

        // Check if any metric value is invalid
        if (metric.getTokenId() == null || metric.getVolumeUsd() == 0.0 || metric.getTokenPriceUsd() == 0.0) {
            log.error("Token metric has invalid values for token ID: {}", metric.getTokenId());
            return;
        }

        out.collect(metric);
    }
}