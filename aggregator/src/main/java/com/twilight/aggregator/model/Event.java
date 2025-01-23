package com.twilight.aggregator.model;

import java.io.Serializable;
import java.util.List;
import java.util.Map;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.ObjectMapper;

public class Event implements Serializable {
    private static final long serialVersionUID = 1L;
    private static transient final ObjectMapper MAPPER = new ObjectMapper();

    @JsonProperty("chainId")
    private String chainId;
    
    @JsonProperty("eventName")
    private String eventName;
    
    @JsonProperty("contractAddress")
    private String contractAddress;
    
    @JsonProperty("logIndex")
    private int logIndex;
    
    @JsonProperty("eventData")
    private String eventData;
    
    @JsonProperty("blockNumber")
    private long blockNumber;

    @JsonProperty("topics")
    private List<String> topics;
 
    @JsonProperty("decodedArgs")
    private Map<String, Object> decodedArgs;

    // Default constructor required by Flink
    public Event() {}

    public static Event fromJson(String json) {
        try {
            return MAPPER.readValue(json, Event.class);
        } catch (Exception e) {
            throw new RuntimeException("Failed to parse event from JSON: " + json, e);
        }
    }

    // Getters and setters
    public String getChainId() {
        return chainId;
    }

    public void setChainId(String chainId) {
        this.chainId = chainId;
    }

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

    public int getLogIndex() {
        return logIndex;
    }

    public void setLogIndex(int logIndex) {
        this.logIndex = logIndex;
    }

    public String getEventData() {
        return eventData;
    }

    public void setEventData(String eventData) {
        this.eventData = eventData;
    }

    public long getBlockNumber() {
        return blockNumber;
    }

    public void setBlockNumber(long blockNumber) {
        this.blockNumber = blockNumber;
    }
    
    public List<String> getTopics() {
        return topics;
    }

    public void setTopics(List<String> topics) {
        this.topics = topics;
    }

    public Map<String, Object> getDecodedArgs() {
        return decodedArgs;
    }

    public void setDecodedArgs(Map<String, Object> decodedArgs) {
        this.decodedArgs = decodedArgs;
    }
} 