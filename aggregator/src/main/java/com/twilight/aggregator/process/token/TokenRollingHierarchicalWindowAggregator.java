package com.twilight.aggregator.process.token;

import java.util.Iterator;

import org.apache.flink.streaming.api.functions.windowing.ProcessWindowFunction;
import org.apache.flink.streaming.api.windowing.windows.TimeWindow;
import org.apache.flink.util.Collector;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.twilight.aggregator.model.TokenRollingMetric;

public class TokenRollingHierarchicalWindowAggregator
        extends ProcessWindowFunction<TokenRollingMetric, TokenRollingMetric, String, TimeWindow> {
    private static final long serialVersionUID = 1L;
    private final String windowName;
    private transient Logger log;

    public TokenRollingHierarchicalWindowAggregator(String windowName) {
        this.windowName = windowName;
    }

    @Override
    public void open(org.apache.flink.configuration.Configuration parameters) throws Exception {
        super.open(parameters);
        this.log = LoggerFactory.getLogger(TokenRollingHierarchicalWindowAggregator.class);
    }

    @Override
    public void process(String key, Context context, Iterable<TokenRollingMetric> elements,
            Collector<TokenRollingMetric> out) {
        log.info("TokenRollingHierarchicalWindowAggregator: Process: {}", elements);
        TokenRollingMetric aggregated = null;
        Iterator<TokenRollingMetric> iterator = elements.iterator();
        long windowEndTime = alignWindowEndTime(context.window().getEnd(), windowName);

        while (iterator.hasNext()) {
            TokenRollingMetric current = iterator.next();
            if (aggregated == null) {
                // Initialize with the first metric
                aggregated = new TokenRollingMetric(current.getTokenId(), current.getVolumeUsd(),
                        current.getTokenPriceUsd());
                aggregated.setTimeWindow(windowName);
                aggregated.setEndTime((windowEndTime / 1000) * 1000); // 精确到秒
            } else {
                // 累加值

                // 最新值使用最后一个窗口的值
                aggregated.setTokenPriceUsd(current.getTokenPriceUsd());
                aggregated.setEndTime(current.getEndTime());

                // Aggregate metrics
                aggregated.setVolumeUsd(aggregated.getVolumeUsd() + current.getVolumeUsd());
            }

        }
        log.info("TokenRollingHierarchicalWindowAggregator: Aggregated: {}", aggregated);
        if (aggregated != null) {
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
            case "1h":
                // Align to hour
                alignedTime = (endTime / 3600000) * 3600000;
                break;
        }
        return alignedTime;
    }

}
