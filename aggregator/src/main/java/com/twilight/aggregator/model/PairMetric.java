package com.twilight.aggregator.model;

import java.io.Serializable;
import java.math.BigDecimal;
import java.sql.Timestamp;

import org.apache.flink.streaming.api.windowing.time.Time;

import lombok.Data;

@Data
public class PairMetric implements Serializable {
    private static final long serialVersionUID = 1L;

    private Long pairId;
    private String timeWindow;
    private Timestamp endTime;
    private BigDecimal token0Reserve = BigDecimal.ZERO;
    private BigDecimal token1Reserve = BigDecimal.ZERO;
    private BigDecimal reserveUsd = BigDecimal.ZERO;
    private BigDecimal token0VolumeUsd = BigDecimal.ZERO;
    private BigDecimal token1VolumeUsd = BigDecimal.ZERO;
    private BigDecimal volumeUsd = BigDecimal.ZERO;
    private int txcnt = 0;

    public PairMetric() {
        this.token0Reserve = BigDecimal.ZERO;
        this.token1Reserve = BigDecimal.ZERO;
        this.reserveUsd = BigDecimal.ZERO;
        this.token0VolumeUsd = BigDecimal.ZERO;
        this.token1VolumeUsd = BigDecimal.ZERO;
        this.volumeUsd = BigDecimal.ZERO;
        this.txcnt = 0;
    }

    public PairMetric(Long pairId, String timeWindow, Time windowSize, long endTimeMillis) {
        this();
        this.pairId = pairId;
        this.timeWindow = timeWindow;
        this.endTime = new Timestamp(endTimeMillis);
    }


    public void incrementTxCount() {
        this.txcnt++;
    }
} 