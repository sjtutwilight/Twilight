package com.twilight.aggregator.model;

import java.io.Serializable;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@JsonIgnoreProperties(ignoreUnknown = true)
public class TokenMetric implements Serializable {
    private static final long serialVersionUID = 1L;

    private Long tokenId;
    private String timeWindow;
    private long endTime;
    private double volumeUsd = 0.0;
    private int txcnt = 0;
    private double tokenPriceUsd = 0.0;
    private double buyPressureUsd = 0.0;
    private int buyersCount = 0;
    private int sellersCount = 0;
    private double buyVolumeUsd = 0.0;
    private double sellVolumeUsd = 0.0;
    private Integer makersCount = 0;
    private Integer buyCount = 0;
    private Integer sellCount = 0;

    public TokenMetric(Long tokenId, String timeWindow, long endTimeMillis) {
        this.tokenId = tokenId;
        this.timeWindow = timeWindow;
        this.endTime = endTimeMillis;
    }

    public void incrementTxCount() {
        this.txcnt++;
    }

    public void incrementBuyCount() {
        this.buyCount++;
    }

    public void incrementSellCount() {
        this.sellCount++;
    }

    public void calculateMakersCount() {
        this.makersCount = this.buyersCount + this.sellersCount;
    }

    public void setVolumeUsd(double volumeUsd) {
        this.volumeUsd = volumeUsd;
    }

    public void setTokenPriceUsd(double tokenPriceUsd) {
        this.tokenPriceUsd = tokenPriceUsd;
    }

    public void setBuyPressureUsd(double buyPressureUsd) {
        this.buyPressureUsd = buyPressureUsd;
    }

    public void setBuyVolumeUsd(double buyVolumeUsd) {
        this.buyVolumeUsd = buyVolumeUsd;
    }

    public void setSellVolumeUsd(double sellVolumeUsd) {
        this.sellVolumeUsd = sellVolumeUsd;
    }
}