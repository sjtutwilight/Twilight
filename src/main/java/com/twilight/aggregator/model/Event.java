package com.twilight.aggregator.model;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.Serializable;
import java.math.BigDecimal;

public class Event implements Serializable {
    private static final long serialVersionUID = 1L;
    private static final ObjectMapper MAPPER = new ObjectMapper();

    private String eventName;
    private String contractAddress;
    private String pairAddress;
    private long blockTimestamp;
    private String transactionHash;
    private JsonNode decodedArgs;
    private BigDecimal usdcReserve;
    private BigDecimal tokenReserve;
    private BigDecimal volume;

    public static Event fromJson(String json) {
        try {
            Event event = new Event();
            JsonNode node = MAPPER.readTree(json);
            
            event.setEventName(node.path("event_name").asText());
            event.setContractAddress(node.path("contract_address").asText());
            event.setPairAddress(node.path("pair_address").asText());
            event.setBlockTimestamp(node.path("block_timestamp").asLong());
            event.setTransactionHash(node.path("transaction_hash").asText());
            event.setDecodedArgs(node.path("decoded_args"));
            
            // Parse numeric values
            String usdcReserveStr = node.path("usdc_reserve").asText("0");
            String tokenReserveStr = node.path("token_reserve").asText("0");
            String volumeStr = node.path("volume").asText("0");
            
            event.setUsdcReserve(new BigDecimal(usdcReserveStr));
            event.setTokenReserve(new BigDecimal(tokenReserveStr));
            event.setVolume(new BigDecimal(volumeStr));
            
            return event;
        } catch (Exception e) {
            throw new RuntimeException("Failed to parse event JSON: " + json, e);
        }
    }

    // Getters and Setters
    public String getEventName() {
        return eventName;
    }

    public void setEventName(String eventName) {
        this.eventName = eventName;
    }

    public String getContractAddress() {
        return contractAddress;
    }

    public void setContractAddress(String contractAddress) {
        this.contractAddress = contractAddress;
    }

    public String getPairAddress() {
        return pairAddress;
    }

    public void setPairAddress(String pairAddress) {
        this.pairAddress = pairAddress;
    }

    public long getBlockTimestamp() {
        return blockTimestamp;
    }

    public void setBlockTimestamp(long blockTimestamp) {
        this.blockTimestamp = blockTimestamp;
    }

    public String getTransactionHash() {
        return transactionHash;
    }

    public void setTransactionHash(String transactionHash) {
        this.transactionHash = transactionHash;
    }

    public JsonNode getDecodedArgs() {
        return decodedArgs;
    }

    public void setDecodedArgs(JsonNode decodedArgs) {
        this.decodedArgs = decodedArgs;
    }

    public BigDecimal getUsdcReserve() {
        return usdcReserve;
    }

    public void setUsdcReserve(BigDecimal usdcReserve) {
        this.usdcReserve = usdcReserve;
    }

    public BigDecimal getTokenReserve() {
        return tokenReserve;
    }

    public void setTokenReserve(BigDecimal tokenReserve) {
        this.tokenReserve = tokenReserve;
    }

    public BigDecimal getVolume() {
        return volume;
    }

    public void setVolume(BigDecimal volume) {
        this.volume = volume;
    }
} 