package com.twilight.aggregator.model;

import java.io.Serializable;
import java.math.BigDecimal;
import java.sql.Timestamp;

import org.apache.flink.streaming.api.windowing.time.Time;

public class PairMetric implements Serializable {
    private static final long serialVersionUID = 1L;

    private Long pairId;
    private String timeWindow;
    private Timestamp endTime;
    private BigDecimal token0Reserve;
    private BigDecimal token1Reserve;
    private BigDecimal reserveUsd;
    private BigDecimal token0VolumeUsd;
    private BigDecimal token1VolumeUsd;
    private BigDecimal volumeUsd;
    private int txcnt;

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

    // Getters and setters
    public Long getPairId() {
        return pairId;
    }

    public void setPairId(Long pairId) {
        this.pairId = pairId;
    }

    public String getTimeWindow() {
        return timeWindow;
    }

    public void setTimeWindow(String timeWindow) {
        this.timeWindow = timeWindow;
    }

    public Timestamp getEndTime() {
        return endTime;
    }

    public void setEndTime(Timestamp endTime) {
        this.endTime = endTime;
    }

    public BigDecimal getToken0Reserve() {
        return token0Reserve;
    }

    public void setToken0Reserve(BigDecimal token0Reserve) {
        this.token0Reserve = token0Reserve;
    }

    public BigDecimal getToken1Reserve() {
        return token1Reserve;
    }

    public void setToken1Reserve(BigDecimal token1Reserve) {
        this.token1Reserve = token1Reserve;
    }

    public BigDecimal getReserveUsd() {
        return reserveUsd;
    }

    public void setReserveUsd(BigDecimal reserveUsd) {
        this.reserveUsd = reserveUsd;
    }

    public BigDecimal getToken0VolumeUsd() {
        return token0VolumeUsd;
    }

    public void setToken0VolumeUsd(BigDecimal token0VolumeUsd) {
        this.token0VolumeUsd = token0VolumeUsd;
    }

    public BigDecimal getToken1VolumeUsd() {
        return token1VolumeUsd;
    }

    public void setToken1VolumeUsd(BigDecimal token1VolumeUsd) {
        this.token1VolumeUsd = token1VolumeUsd;
    }

    public BigDecimal getVolumeUsd() {
        return volumeUsd;
    }

    public void setVolumeUsd(BigDecimal volumeUsd) {
        this.volumeUsd = volumeUsd;
    }

    public int getTxcnt() {
        return txcnt;
    }

    public void setTxcnt(int txcnt) {
        this.txcnt = txcnt;
    }

    public void incrementTxCount() {
        this.txcnt++;
    }
} 