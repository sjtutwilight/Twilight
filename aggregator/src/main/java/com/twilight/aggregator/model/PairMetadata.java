package com.twilight.aggregator.model;

import java.io.Serializable;
import java.util.Map;
import java.util.HashMap;

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

    // 存储地址到标签的映射
    private Map<String, String> addressTagMap = new HashMap<>();

    // 添加地址标签
    public void addAddressTag(String address, String tag) {
        if (address != null && tag != null) {
            addressTagMap.put(address.toLowerCase(), tag);
        }
    }

    // 获取地址标签
    public String getAddressTag(String address) {
        if (address == null) {
            return "all";
        }
        return addressTagMap.getOrDefault(address.toLowerCase(), "all");
    }
}