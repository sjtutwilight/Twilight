package com.twilight.aggregator.utils;

import java.math.BigDecimal;
import java.math.BigInteger;

public class EthereumUtils {
    private static final BigDecimal WEI_TO_ETH = new BigDecimal("1000000000000000000"); // 10^18

    public static BigDecimal convertWeiToEth(String weiAmount) {
        if (weiAmount == null || weiAmount.isEmpty()) {
            return BigDecimal.ZERO;
        }
        try {
            return new BigDecimal(weiAmount).divide(WEI_TO_ETH, 18, BigDecimal.ROUND_HALF_UP);
        } catch (NumberFormatException e) {
            return BigDecimal.ZERO;
        }
    }

    public static BigDecimal convertWeiToEth(BigInteger weiAmount) {
        if (weiAmount == null) {
            return BigDecimal.ZERO;
        }
        return new BigDecimal(weiAmount).divide(WEI_TO_ETH, 18, BigDecimal.ROUND_HALF_UP);
    }
} 