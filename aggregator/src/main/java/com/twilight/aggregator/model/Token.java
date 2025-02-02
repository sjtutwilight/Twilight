package com.twilight.aggregator.model;

import java.io.Serializable;
import java.math.BigDecimal;

import lombok.Data;

@Data
public class Token implements Serializable {
    private static final long serialVersionUID = 1L;
    
    private String tokenAddress;
    private Long tokenId;
    private BigDecimal tokenPriceUsd;
    private boolean buyOrSell;
    private BigDecimal amount;
    private String fromAddress;
} 