package com.twilight.aggregator.process;

import org.apache.flink.api.common.functions.FlatMapFunction;
import org.apache.flink.util.Collector;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.twilight.aggregator.model.Event;
import com.twilight.aggregator.model.Transaction;
public class EventExtractor implements FlatMapFunction<Transaction, Event> {
    private static final Logger log = LoggerFactory.getLogger(EventExtractor.class);
    @Override
    public void flatMap(Transaction transaction, Collector<Event> out) {
        if (transaction.getEvents() != null) {
            for (Event event : transaction.getEvents()) {
                event.setFromAddress(transaction.getFromAddress());
                out.collect(event);
            }
        }
    }
} 