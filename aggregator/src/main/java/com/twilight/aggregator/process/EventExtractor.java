package com.twilight.aggregator.process;

import org.apache.flink.api.common.functions.FlatMapFunction;
import org.apache.flink.util.Collector;

import com.twilight.aggregator.model.Event;
import com.twilight.aggregator.model.ProcessEvent;
import com.twilight.aggregator.model.KafkaMessage;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class EventExtractor implements FlatMapFunction<KafkaMessage, ProcessEvent> {
    private static final Logger log = LoggerFactory.getLogger(EventExtractor.class);

    @Override
    public void flatMap(KafkaMessage message, Collector<ProcessEvent> out) {
        if (message.getEvents() != null && message.getTransaction() != null) {
            String fromAddress = message.getTransaction().getFromAddress();
            Long timestamp = message.getTransaction().getTimestamp();

            for (Event event : message.getEvents()) {
                if (event.getEventName().equals("Approval")) {
                    break;
                }
                ProcessEvent processEvent = new ProcessEvent();
                // 基础字段
                processEvent.setEventName(event.getEventName());
                processEvent.setContractAddress(event.getContractAddress());
                processEvent.setDecodedArgs(event.getDecodedArgs());
                processEvent.setFromAddress(fromAddress);
                processEvent.setTimestamp(timestamp);
                // 各字段都不为null，有null则log.error
                if (processEvent.getEventName() == null || processEvent.getContractAddress() == null
                        || processEvent.getDecodedArgs() == null || processEvent.getFromAddress() == null
                        || processEvent.getTimestamp() == null) {
                    log.error("EventExtractor: Process event is null: {}", processEvent);
                }
                out.collect(processEvent);
            }
        }
    }
}