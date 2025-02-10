package com.twilight.aggregator.model;

import java.io.Serializable;
import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@JsonIgnoreProperties(ignoreUnknown = true)
public class Pair implements Serializable {
    private static final long serialVersionUID = 1L;

    private Long pairId;
    private String pairAddress;
    private String token0Address;
    private String token1Address;
    private double token0PriceUsd;
    private double token1PriceUsd;
    private String eventName;

    // Swap event args
    private String amount0In;
    private String amount0Out;
    private String amount1In;
    private String amount1Out;

    // Sync event args
    private String reserve0;
    private String reserve1;

    // Mint/Burn event args
    private String sender;
    private String to;
    private String amount0;
    private String amount1;

    private String fromAddress;
}