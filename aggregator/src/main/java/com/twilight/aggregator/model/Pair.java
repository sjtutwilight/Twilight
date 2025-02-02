package com.twilight.aggregator.model;

import java.io.Serializable;
import java.math.BigDecimal;
import java.util.Map;

import lombok.Data;

@Data
public class Pair implements Serializable {
    private static final long serialVersionUID = 1L;
    
    private Long pairId;
    private String pairAddress;
    private String token0Address;
    private String token1Address;
    private BigDecimal token0PriceUsd;
    private BigDecimal token1PriceUsd;

    private String eventName;
    private Map<String, Object> eventArgs;
    private String fromAddress;
} 