package com.twilight.aggregator.process;

import java.math.BigDecimal;
import java.util.Collection;
import java.util.concurrent.atomic.AtomicInteger;

import org.apache.flink.api.common.state.BroadcastState;
import org.apache.flink.api.common.state.MapStateDescriptor;
import org.apache.flink.api.common.state.ReadOnlyBroadcastState;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.async.ResultFuture;
import org.apache.flink.streaming.api.functions.co.BroadcastProcessFunction;
import org.apache.flink.util.Collector;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.twilight.aggregator.model.Event;
import com.twilight.aggregator.model.Pair;
import com.twilight.aggregator.model.PairMetadata;
import com.twilight.aggregator.source.AsyncPriceLookupFunction;

public class PairMetadataProcessor extends BroadcastProcessFunction<Event, PairMetadata, Pair> {
    private final MapStateDescriptor<String, PairMetadata> pairMetadataDescriptor;
    private final Logger log = LoggerFactory.getLogger(PairMetadataProcessor.class);
    private final AsyncPriceLookupFunction priceLookup;

    public PairMetadataProcessor(MapStateDescriptor<String, PairMetadata> descriptor) {
        this.pairMetadataDescriptor = descriptor;
        this.priceLookup = new AsyncPriceLookupFunction();
        try {
            this.priceLookup.open(new Configuration());
        } catch (Exception e) {
            log.error("Failed to initialize price lookup function", e);
            throw new RuntimeException(e);
        }
    }

    @Override
    public void processElement(Event event, ReadOnlyContext ctx, Collector<Pair> out) throws Exception {
        String eventName = event.getEventName();
        if (isSwapEvent(eventName)) {
            ReadOnlyBroadcastState<String, PairMetadata> broadcastState = ctx.getBroadcastState(pairMetadataDescriptor);
            PairMetadata metadata = broadcastState.get(event.getContractAddress());
            
            if (metadata == null) {
                log.error("Pair metadata not found for contract address: {}", event.getContractAddress());
                return;
            }

            Pair pair = new Pair();
            pair.setPairId(metadata.getPairId());
            pair.setPairAddress(metadata.getPairAddress());
            pair.setToken0Address(metadata.getToken0Address());
            pair.setToken1Address(metadata.getToken1Address());
            pair.setEventName(eventName);
            pair.setEventArgs(event.getDecodedArgs());
            pair.setFromAddress(event.getFromAddress());

            // 使用AtomicInteger来跟踪完成的价格查询数量
            AtomicInteger completedQueries = new AtomicInteger(0);
            ResultFuture<BigDecimal> token0Future = new ResultFuture<BigDecimal>() {
                @Override
                public void complete(Collection<BigDecimal> result) {
                    pair.setToken0PriceUsd(result.iterator().next());
                    if (completedQueries.incrementAndGet() == 2) {
                        log.info("PairMetadataProcessor processElement: {}", pair);
                        out.collect(pair);
                    }
                }

                @Override
                public void completeExceptionally(Throwable error) {
                    log.error("Error getting token0 price", error);
                    pair.setToken0PriceUsd(BigDecimal.ZERO);
                    if (completedQueries.incrementAndGet() == 2) {
                        out.collect(pair);
                    }
                }
            };

            ResultFuture<BigDecimal> token1Future = new ResultFuture<BigDecimal>() {
                @Override
                public void complete(Collection<BigDecimal> result) {
                    pair.setToken1PriceUsd(result.iterator().next());
                    if (completedQueries.incrementAndGet() == 2) {
                        out.collect(pair);
                    }
                }

                @Override
                public void completeExceptionally(Throwable error) {
                    log.error("Error getting token1 price", error);
                    pair.setToken1PriceUsd(BigDecimal.ZERO);
                    if (completedQueries.incrementAndGet() == 2) {
                        out.collect(pair);
                    }
                }
            };

            // 并行执行两个价格查询
            priceLookup.asyncInvoke(metadata.getToken0Address(), token0Future);
            priceLookup.asyncInvoke(metadata.getToken1Address(), token1Future);
        }
    }

    @Override
    public void processBroadcastElement(PairMetadata metadata, Context ctx, Collector<Pair> out) throws Exception {
        BroadcastState<String, PairMetadata> broadcastState = ctx.getBroadcastState(pairMetadataDescriptor);
        broadcastState.put(metadata.getPairAddress(), metadata);
    }

    private boolean isSwapEvent(String eventName) {
        return eventName.equals("Swap") || 
               eventName.equals("Sync") || 
               eventName.equals("Mint") || 
               eventName.equals("Burn");
    }
} 