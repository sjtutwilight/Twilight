package com.twilight.aggregator.process;

import java.util.Iterator;

import org.apache.flink.streaming.api.functions.windowing.ProcessWindowFunction;
import org.apache.flink.streaming.api.windowing.windows.TimeWindow;
import org.apache.flink.util.Collector;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.twilight.aggregator.model.TokenMetric;

public class TokenHierarchicalWindowAggregator
        extends ProcessWindowFunction<TokenMetric, TokenMetric, String, TimeWindow> {
    private static final long serialVersionUID = 1L;
    private final String windowName;
    private transient Logger log;

    public TokenHierarchicalWindowAggregator(String windowName) {
        this.windowName = windowName;
    }

    @Override
    public void open(org.apache.flink.configuration.Configuration parameters) throws Exception {
        super.open(parameters);
        this.log = LoggerFactory.getLogger(TokenHierarchicalWindowAggregator.class);
    }

    @Override
    public void process(String key, Context context, Iterable<TokenMetric> elements, Collector<TokenMetric> out) {
        TokenMetric aggregated = null;
        Iterator<TokenMetric> iterator = elements.iterator();
        long windowEndTime = alignWindowEndTime(context.window().getEnd(), windowName);

        while (iterator.hasNext()) {
            TokenMetric current = iterator.next();
            if (aggregated == null) {
                // Initialize with the first metric
                aggregated = new TokenMetric();
                aggregated.setTokenId(current.getTokenId());
                aggregated.setTimeWindow(windowName);
                aggregated.setEndTime(current.getEndTime());
                // 累加值初始化
                aggregated.setVolumeUsd(current.getVolumeUsd());
                aggregated.setTxcnt(current.getTxcnt());
                aggregated.setBuyVolumeUsd(current.getBuyVolumeUsd());
                aggregated.setSellVolumeUsd(current.getSellVolumeUsd());
                aggregated.setBuyersCount(current.getBuyersCount());
                aggregated.setSellersCount(current.getSellersCount());
                // 最新值初始化
                aggregated.setTokenPriceUsd(current.getTokenPriceUsd());
            } else {
                // 累加值
                aggregated.setVolumeUsd(aggregated.getVolumeUsd() + current.getVolumeUsd());
                aggregated.setTxcnt(aggregated.getTxcnt() + current.getTxcnt());
                aggregated.setBuyVolumeUsd(aggregated.getBuyVolumeUsd() + current.getBuyVolumeUsd());
                aggregated.setSellVolumeUsd(aggregated.getSellVolumeUsd() + current.getSellVolumeUsd());
                aggregated.setBuyersCount(aggregated.getBuyersCount() + current.getBuyersCount());
                aggregated.setSellersCount(aggregated.getSellersCount() + current.getSellersCount());
                // 最新值使用最后一个窗口的值
                aggregated.setTokenPriceUsd(current.getTokenPriceUsd());
                aggregated.setEndTime(current.getEndTime());
            }

            // Aggregate metrics
            aggregated.setVolumeUsd(aggregated.getVolumeUsd() + current.getVolumeUsd());
            aggregated.setBuyPressureUsd(aggregated.getBuyPressureUsd() + current.getBuyPressureUsd());
            aggregated.setBuyersCount(aggregated.getBuyersCount() + current.getBuyersCount());
            aggregated.setSellersCount(aggregated.getSellersCount() + current.getSellersCount());
            aggregated.setBuyVolumeUsd(aggregated.getBuyVolumeUsd() + current.getBuyVolumeUsd());
            aggregated.setSellVolumeUsd(aggregated.getSellVolumeUsd() + current.getSellVolumeUsd());
            aggregated.setBuyCount(aggregated.getBuyCount() + current.getBuyCount());
            aggregated.setSellCount(aggregated.getSellCount() + current.getSellCount());
            aggregated.incrementTxCount();
        }

        if (aggregated != null) {
            // 计算买入压力
            aggregated.setBuyPressureUsd(aggregated.getBuyVolumeUsd() - aggregated.getSellVolumeUsd());
            aggregated.calculateMakersCount();
            out.collect(aggregated);
        }
    }

    private long alignWindowEndTime(long endTime, String windowName) {
        long alignedTime = endTime;
        switch (windowName) {
            case "1min":
                // Align to minute
                alignedTime = (endTime / 60000) * 60000;
                break;
            case "5min":
                // Align to 5 minutes
                alignedTime = (endTime / 300000) * 300000;
                break;
            case "30min":
                // Align to 30 minutes
                alignedTime = (endTime / 1800000) * 1800000;
                break;
        }
        return alignedTime;
    }
}