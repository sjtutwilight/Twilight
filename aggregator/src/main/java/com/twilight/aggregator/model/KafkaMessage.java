package com.twilight.aggregator.model;

import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;

import lombok.Data;

@Data
public class KafkaMessage {
    @JsonProperty("transaction")
    private Transaction transaction;
    
    @JsonProperty("events")
    private List<Event> events;

    // Getters and Setters
    public Transaction getTransaction() {
        return transaction;
    }

    public void setTransaction(Transaction transaction) {
        this.transaction = transaction;
    }

    public List<Event> getEvents() {
        return events;
    }

    public void setEvents(List<Event> events) {
        this.events = events;
    }
} 