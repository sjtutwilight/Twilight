package com.twilight.aggregator.process;

import org.apache.flink.api.common.functions.FlatMapFunction;
import org.apache.flink.util.Collector;

import com.twilight.aggregator.model.Event;
import com.twilight.aggregator.model.KafkaMessage;

public class EventExtractor implements FlatMapFunction<KafkaMessage, Event> {

    @Override
    public void flatMap(KafkaMessage message, Collector<Event> out) {
        if (message.getEvents() != null && message.getTransaction() != null) {
            String fromAddress = message.getTransaction().getFromAddress();
            for (Event event : message.getEvents()) {
                event.setFromAddress(fromAddress);
                out.collect(event);
            }
        } else {
        }
    }
}