package com.twilight.aggregator.process;

import java.math.BigDecimal;
import java.sql.Timestamp;
import java.util.HashSet;
import java.util.Set;

import org.apache.flink.streaming.api.functions.windowing.ProcessWindowFunction;  
import org.apache.flink.streaming.api.windowing.windows.TimeWindow;
import org.apache.flink.util.Collector;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.twilight.aggregator.model.Token;
import com.twilight.aggregator.model.TokenMetric;
public class TokenWindowProcessor extends ProcessWindowFunction<Token, TokenMetric, String, TimeWindow> {
    private final String windowName;
    private final Logger log = LoggerFactory.getLogger(TokenWindowProcessor.class);
    public TokenWindowProcessor(String windowName) {
        this.windowName = windowName;
    }

    @Override
    public void process(String key, Context context, Iterable<Token> elements, Collector<TokenMetric> out) {
        TokenMetric metric = new TokenMetric();
        metric.setTimeWindow(windowName);
        metric.setEndTime(new Timestamp(context.window().getEnd()));

        Set<String> buyers = new HashSet<>();
        Set<String> sellers = new HashSet<>();

        for (Token token : elements) {
            metric.setTokenId(token.getTokenId());
            updateMetric(metric, token, buyers, sellers);
        }

        metric.setBuyersCount(buyers.size());
        metric.setSellersCount(sellers.size());
        log.info("Token metric: {}", metric.toString());
        out.collect(metric);
    }

    private void updateMetric(TokenMetric metric, Token token, Set<String> buyers, Set<String> sellers) {
        BigDecimal volumeUsd = token.getAmount().multiply(token.getTokenPriceUsd());
        metric.setVolumeUsd(metric.getVolumeUsd().add(volumeUsd));
        metric.setTokenPriceUsd(token.getTokenPriceUsd());

        if (token.isBuyOrSell()) {
            metric.setBuyVolumeUsd(metric.getBuyVolumeUsd().add(volumeUsd));
            buyers.add(token.getFromAddress());
            metric.incrementBuyCount();
        } else {
            metric.setSellVolumeUsd(metric.getSellVolumeUsd().add(volumeUsd));
            sellers.add(token.getFromAddress());
            metric.incrementSellCount();
        }
        
        metric.setBuyPressureUsd(metric.getBuyVolumeUsd().subtract(metric.getSellVolumeUsd()));
        metric.incrementTxCount();
        
        // Update counts after processing each token
        metric.setBuyersCount(buyers.size());
        metric.setSellersCount(sellers.size());
        metric.calculateMakersCount();
    }
} 