package com.twilight.aggregator.model;

import java.io.Serializable;
import java.math.BigDecimal;
import java.sql.Timestamp;

import lombok.Data;

@Data
public class TokenMetric implements Serializable {
    private static final long serialVersionUID = 1L;

    private Long tokenId;
    private String timeWindow;
    private Timestamp endTime;
    private BigDecimal volumeUsd = BigDecimal.ZERO;
    private Integer txcnt = 0;
    private BigDecimal tokenPriceUsd = BigDecimal.ZERO;
    private BigDecimal buyPressureUsd = BigDecimal.ZERO;
    private Integer buyersCount = 0;
    private Integer sellersCount = 0;
    private BigDecimal buyVolumeUsd = BigDecimal.ZERO;
    private BigDecimal sellVolumeUsd = BigDecimal.ZERO;
    private Integer makersCount = 0;
    private Integer buyCount = 0;
    private Integer sellCount = 0;

    public TokenMetric() {
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
} 