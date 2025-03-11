package com.twilight.aggregator.model;

import java.io.Serializable;
import java.sql.Timestamp;

import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
public class TokenRollingMetric implements Serializable {
    private static final long serialVersionUID = 1L;

    private Long tokenId;
    private String timeWindow;
    private long endTime = System.currentTimeMillis();
    private double volumeUsd;
    private double tokenPriceUsd;

    public TokenRollingMetric(Long tokenId, double volumeUsd, double tokenPriceUsd) {
        this.tokenId = tokenId;
        this.volumeUsd = volumeUsd;
        this.tokenPriceUsd = tokenPriceUsd;
        this.endTime = System.currentTimeMillis();
    }

    @Override
    public String toString() {
        return "TokenRollingMetric(" +
                "tokenId=" + tokenId +
                ", timeWindow=" + timeWindow +
                ", endTime=" + new Timestamp(endTime) +
                ", volumeUsd=" + volumeUsd +
                ", tokenPriceUsd=" + tokenPriceUsd +
                ')';
    }
}
