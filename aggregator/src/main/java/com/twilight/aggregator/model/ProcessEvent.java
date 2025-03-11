package com.twilight.aggregator.model;

import java.io.Serializable;
import java.util.Map;

import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
public class ProcessEvent implements Serializable {
    private static final long serialVersionUID = 1L;

    private String eventName;
    private String contractAddress;
    private Map<String, String> decodedArgs;
    private String fromAddress;
    private String fromAddressTag;
    private Long timestamp;

    private Long pairId;
    private Long token0Id;
    private Long token1Id;
    private String token0Address;
    private String token1Address;
    private double token0PriceUsd;
    private double token1PriceUsd;

}
