package com.twilight.aggregator.model;

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
    private Long blockTimestamp;
    
    @JsonProperty("fromAddress")
    private String fromAddress;
    
    @JsonProperty("toAddress")
    private String toAddress;
    
    @JsonProperty("inputData")
    private String inputData;
    
    @JsonProperty("transactionStatus")
    private String status;
    
    @JsonProperty("gasUsed")
    private Long gasUsed;
    
    @JsonProperty("gasPrice")
    private String gasPrice;
    
    @JsonProperty("chainID")
    private String chainId;
    
    @JsonProperty("nonce")
    private Long nonce;
    
    @JsonProperty("transactionValue")
    private String transactionValue;
    
    private Event[] events;

    // Getters and setters
    public String getTransactionHash() { return transactionHash; }
    public void setTransactionHash(String transactionHash) { this.transactionHash = transactionHash; }
    
    public Long getBlockNumber() { return blockNumber; }
    public void setBlockNumber(Long blockNumber) { this.blockNumber = blockNumber; }
    
    public String getBlockHash() { return blockHash; }
    public void setBlockHash(String blockHash) { this.blockHash = blockHash; }
    
    public Long getBlockTimestamp() { return blockTimestamp; }
    public void setBlockTimestamp(Long blockTimestamp) { this.blockTimestamp = blockTimestamp; }
    
    public String getFromAddress() { return fromAddress; }
    public void setFromAddress(String fromAddress) { this.fromAddress = fromAddress; }
    
    public String getToAddress() { return toAddress; }
    public void setToAddress(String toAddress) { this.toAddress = toAddress; }
    
    public String getInputData() { return inputData; }
    public void setInputData(String inputData) { this.inputData = inputData; }
    
    public String getStatus() { return status; }
    public void setStatus(String status) { this.status = status; }
    
    public Long getGasUsed() { return gasUsed; }
    public void setGasUsed(Long gasUsed) { this.gasUsed = gasUsed; }
    
    public String getGasPrice() { return gasPrice; }
    public void setGasPrice(String gasPrice) { this.gasPrice = gasPrice; }
    
    public String getChainId() { return chainId; }
    public void setChainId(String chainId) { this.chainId = chainId; }
    
    public Long getNonce() { return nonce; }
    public void setNonce(Long nonce) { this.nonce = nonce; }
    
    public String getTransactionValue() { return transactionValue; }
    public void setTransactionValue(String transactionValue) { this.transactionValue = transactionValue; }
    
    public Event[] getEvents() { return events; }
    public void setEvents(Event[] events) { this.events = events; }
} 