package com.twilight.aggregator.model;

import java.io.Serializable;

import lombok.Data;

@Data
public class Token implements Serializable {
    private static final long serialVersionUID = 1L;

    private String tokenAddress;
    private Long tokenId;
    private double tokenPriceUsd;
    private boolean buyOrSell;
    private double amount;
    private String fromAddress;
}