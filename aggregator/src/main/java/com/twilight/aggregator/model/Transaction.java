package com.twilight.aggregator.model;

import java.util.ArrayList;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;
@JsonIgnoreProperties(ignoreUnknown = true)
public class Transaction {
    @JsonProperty("transactionHash")
    private String transactionHash;
    
    @JsonProperty("blockNumber")
    private Long blockNumber;
    
    @JsonProperty("blockHash")
    private String blockHash;
    
    @JsonProperty("timestamp")
    private Long timestamp;
    
    @JsonProperty("transactionStatus")
    private String transactionStatus;
    
    @JsonProperty("gasUsed")
    private Long gasUsed;
    
    @JsonProperty("gasPrice")
    private String gasPrice;
    
    @JsonProperty("nonce")
    private Integer nonce;
    
    @JsonProperty("fromAddress")
    private String fromAddress;
    
    @JsonProperty("toAddress")
    private String toAddress;
    
    @JsonProperty("transactionValue")
    private String transactionValue;
    
    @JsonProperty("inputData")
    private String inputData;
    
    @JsonProperty("chainID")
    private String chainID;
    
    @JsonProperty("events")
    private List<Event> events;

    // Getters and setters omitted for brevity
    public String getTransactionHash() {
        return transactionHash;
    }

    public void setTransactionHash(String transactionHash) {
        this.transactionHash = transactionHash;
    }

    public Long getBlockNumber() {
        return blockNumber;
    }

    public void setBlockNumber(Long blockNumber) {
        this.blockNumber = blockNumber;
    }

    public String getBlockHash() {
        return blockHash;
    }

    public void setBlockHash(String blockHash) {
        this.blockHash = blockHash;
    }

    public Long getTimestamp() {
        return timestamp;
    }

    public void setTimestamp(Long timestamp) {
        this.timestamp = timestamp;
    }

    public String getTransactionStatus() {
        return transactionStatus;
    }

    public void setTransactionStatus(String transactionStatus) {
        this.transactionStatus = transactionStatus;
    }

    public Long getGasUsed() {
        return gasUsed;
    }

    public void setGasUsed(Long gasUsed) {
        this.gasUsed = gasUsed;
    }

    public String getGasPrice() {
        return gasPrice;
    }

    public void setGasPrice(String gasPrice) {
        this.gasPrice = gasPrice;
    }

    public Integer getNonce() {
        return nonce;
    }

    public void setNonce(Integer nonce) {
        this.nonce = nonce;
    }

    public String getFromAddress() {
        return fromAddress;
    }

    public void setFromAddress(String fromAddress) {
        this.fromAddress = fromAddress;
    }

    public String getToAddress() {
        return toAddress;
    }

    public void setToAddress(String toAddress) {
        this.toAddress = toAddress;
    }

    public String getTransactionValue() {
        return transactionValue;
    }

    public void setTransactionValue(String transactionValue) {
        this.transactionValue = transactionValue;
    }

    public String getInputData() {
        return inputData;
    }

    public void setInputData(String inputData) {
        this.inputData = inputData;
    }

    public String getChainID() {
        return chainID;
    }

    public void setChainID(String chainID) {
        this.chainID = chainID;
    }

    public List<Event> getEvents() {
        return new ArrayList<>(events);
    }

    public void setEvents(List<Event> events) {
        this.events = events;
    }
}
