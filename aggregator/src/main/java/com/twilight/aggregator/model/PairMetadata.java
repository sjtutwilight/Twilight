package com.twilight.aggregator.model;

import java.io.Serializable;

import lombok.Data;

@Data
public class PairMetadata implements Serializable {
    private static final long serialVersionUID = 1L;
    
    private Long pairId;
    private String pairAddress;
    private Long token0Id;
    private Long token1Id;
    private String token0Address;
    private String token1Address;

} 