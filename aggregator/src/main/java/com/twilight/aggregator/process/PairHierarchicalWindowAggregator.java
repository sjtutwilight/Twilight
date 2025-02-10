package com.twilight.aggregator.process;

import java.util.Iterator;

import org.apache.flink.streaming.api.functions.windowing.ProcessWindowFunction;
import org.apache.flink.streaming.api.windowing.windows.TimeWindow;
import org.apache.flink.util.Collector;

import com.twilight.aggregator.model.PairMetric;

public class PairHierarchicalWindowAggregator
        extends ProcessWindowFunction<PairMetric, PairMetric, String, TimeWindow> {
    private final String windowName;

    public PairHierarchicalWindowAggregator(String windowName) {
        this.windowName = windowName;
    }

    @Override
    public void open(org.apache.flink.configuration.Configuration parameters) throws Exception {
        super.open(parameters);
    }

    @Override
    public void process(String key, Context context, Iterable<PairMetric> elements, Collector<PairMetric> out) {
        PairMetric aggregated = null;
        Iterator<PairMetric> iterator = elements.iterator();
        long windowEndTime = alignWindowEndTime(context.window().getEnd(), windowName);

        while (iterator.hasNext()) {
            PairMetric current = iterator.next();
            if (aggregated == null) {
                // Initialize with the first metric
                aggregated = new PairMetric();
                aggregated.setPairId(current.getPairId());
                aggregated.setTimeWindow(windowName);
                aggregated.setEndTime(current.getEndTime());
                // 累加值初始化
                aggregated.setToken0VolumeUsd(current.getToken0VolumeUsd());
                aggregated.setToken1VolumeUsd(current.getToken1VolumeUsd());
                aggregated.setVolumeUsd(current.getVolumeUsd());
                aggregated.setTxcnt(current.getTxcnt());
                // 最新值初始化
                aggregated.setToken0Reserve(current.getToken0Reserve());
                aggregated.setToken1Reserve(current.getToken1Reserve());
                aggregated.setReserveUsd(current.getReserveUsd());
            } else {
                // 累加值
                aggregated.setToken0VolumeUsd(aggregated.getToken0VolumeUsd() + current.getToken0VolumeUsd());
                aggregated.setToken1VolumeUsd(aggregated.getToken1VolumeUsd() + current.getToken1VolumeUsd());
                aggregated.setVolumeUsd(aggregated.getVolumeUsd() + current.getVolumeUsd());
                aggregated.setTxcnt(aggregated.getTxcnt() + current.getTxcnt());
                // 最新值使用最后一个窗口的值
                aggregated.setToken0Reserve(current.getToken0Reserve());
                aggregated.setToken1Reserve(current.getToken1Reserve());
                aggregated.setReserveUsd(current.getReserveUsd());
                aggregated.setEndTime(current.getEndTime());
            }
        }

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
            case "30min":
                // Align to 30 minutes
                alignedTime = (endTime / 1800000) * 1800000;
                break;
        }
        return alignedTime;
    }
}