package com.twilight.aggregator.process;

import org.apache.flink.api.common.state.MapStateDescriptor;
import org.apache.flink.streaming.api.functions.co.BroadcastProcessFunction;
import org.apache.flink.util.Collector;
import org.apache.flink.api.common.state.ReadOnlyBroadcastState;
import com.twilight.aggregator.model.PairMetadata;
import com.twilight.aggregator.model.ProcessEvent;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class EventEnrichmentProcessor extends BroadcastProcessFunction<ProcessEvent, PairMetadata, ProcessEvent> {
    private final MapStateDescriptor<String, PairMetadata> pairMetadataDescriptor;
    private static final Logger log = LoggerFactory.getLogger(EventEnrichmentProcessor.class);

    public EventEnrichmentProcessor(MapStateDescriptor<String, PairMetadata> descriptor) {
        this.pairMetadataDescriptor = descriptor;
    }

    @Override
    public void processElement(ProcessEvent event, ReadOnlyContext ctx, Collector<ProcessEvent> out) {
        try {
            if (event == null || event.getContractAddress() == null) {
                log.error("Invalid event or contract address is null");
                return;
            }
            String contractAddress = event.getContractAddress().toLowerCase();
            ReadOnlyBroadcastState<String, PairMetadata> broadcastState = ctx.getBroadcastState(pairMetadataDescriptor);
            PairMetadata metadata = broadcastState.get(contractAddress);

            if (metadata == null) {
                return;
            }

            if (metadata.getPairId() != null) {
                event.setPairId(metadata.getPairId());
            }
            if (metadata.getToken0Id() != null) {
                event.setToken0Id(metadata.getToken0Id());
            }
            if (metadata.getToken1Id() != null) {
                event.setToken1Id(metadata.getToken1Id());
            }
            if (metadata.getToken0Address() != null) {
                event.setToken0Address(metadata.getToken0Address());
            }
            if (metadata.getToken1Address() != null) {
                event.setToken1Address(metadata.getToken1Address());
            }

            if (event.getFromAddress() != null) {
                String fromAddressTag = metadata.getAddressTag(event.getFromAddress());
                event.setFromAddressTag(fromAddressTag);
            } else {
                event.setFromAddressTag("all");
            }

            boolean isValid = event.getPairId() != null &&
                    event.getToken0Id() != null &&
                    event.getToken1Id() != null &&
                    event.getToken0Address() != null &&
                    event.getToken1Address() != null;

            if (!isValid) {
                log.error("Missing required fields after enrichment: {}", event);
                return;
            }

            out.collect(event);
        } catch (Exception e) {
            log.error("Error processing event for contract {}: {}", event.getContractAddress(), e.getMessage());
        }
    }

    @Override
    public void processBroadcastElement(PairMetadata metadata, Context ctx, Collector<ProcessEvent> out) {
        try {
            if (metadata == null || metadata.getPairAddress() == null) {
                log.error("Invalid metadata or pair address is null");
                return;
            }

            String pairAddress = metadata.getPairAddress().toLowerCase();
            ctx.getBroadcastState(pairMetadataDescriptor).put(pairAddress, metadata);
        } catch (Exception e) {
            log.error("Error broadcasting metadata for pair {}: {}", metadata.getPairAddress(), e.getMessage());
        }
    }
}
