package com.twilight.aggregator.utils;

public class EthereumUtils {
    private static final double WEI_TO_ETH = 1e18; // 10^18

    public static double convertWeiToEth(String weiAmount) {
        if (weiAmount == null) {
            return 0.0;
        }
        try {
            return Double.parseDouble(weiAmount) / WEI_TO_ETH;
        } catch (NumberFormatException e) {
            return 0.0;
        }
    }

    public static double convertWeiToEth(long weiAmount) {
        return weiAmount / WEI_TO_ETH;
    }
}