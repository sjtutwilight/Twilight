package com.twilight.aggregator.process;

import java.io.Serializable;
import java.util.Collection;
import java.util.Map;

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
import com.twilight.aggregator.model.PairMetadata;
import com.twilight.aggregator.model.Token;
import com.twilight.aggregator.source.AsyncPriceLookupFunction;
import com.twilight.aggregator.utils.EthereumUtils;
import com.twilight.aggregator.utils.SerializableResultFuture;

public class TokenMetadataProcessor extends BroadcastProcessFunction<Event, PairMetadata, Token> {
    private static final long serialVersionUID = 1L;
    private final MapStateDescriptor<String, PairMetadata> pairMetadataDescriptor;
    private transient Logger log;
    private transient AsyncPriceLookupFunction priceLookup;

    public TokenMetadataProcessor(MapStateDescriptor<String, PairMetadata> descriptor) {
        this.pairMetadataDescriptor = descriptor;
    }

    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        this.priceLookup = new AsyncPriceLookupFunction();
        this.priceLookup.open(parameters);
        this.log = LoggerFactory.getLogger(TokenMetadataProcessor.class);
    }

    @Override
    public void close() throws Exception {
        if (priceLookup != null) {
            priceLookup.close();
        }
        super.close();
    }

    @Override
    public void processElement(Event event, ReadOnlyContext ctx, Collector<Token> out) throws Exception {
        if ("Swap".equals(event.getEventName())) {
            ReadOnlyBroadcastState<String, PairMetadata> broadcastState = ctx.getBroadcastState(pairMetadataDescriptor);
            PairMetadata metadata = broadcastState.get(event.getContractAddress());

            if (metadata == null) {
                log.error("Pair metadata is null for contract address: {}", event.getContractAddress());
                return;
            }
            log.debug("fromAddress: {}", event.getFromAddress());
            processSwapEvent(metadata, event.getDecodedArgs(), event.getFromAddress(), out);
        }
    }

    private void processSwapEvent(PairMetadata metadata, Map<String, String> args, String fromAddress,
            Collector<Token> out) {
        // Process token0
        double amount0In = EthereumUtils.convertWeiToEth(args.get("amount0In"));
        double amount0Out = EthereumUtils.convertWeiToEth(args.get("amount0Out"));
        if (amount0In > 0 || amount0Out > 0) {
            Token token0 = new Token();
            token0.setTokenAddress(metadata.getToken0Address());
            token0.setTokenId(metadata.getToken0Id());
            token0.setAmount(amount0In + amount0Out);
            token0.setBuyOrSell(amount0Out > 0);
            token0.setFromAddress(fromAddress);

            // 异步获取token0价格
            ResultFuture<Double> token0Future = new SerializableResultFuture<>(
                    null,
                    result -> {
                        if (result != null && !result.isEmpty()) {
                            double price = result.iterator().next();
                            token0.setTokenPriceUsd(price);
                            log.debug("Processing token0: address={}, price={}, amount={}, from={}",
                                    token0.getTokenAddress(), price, token0.getAmount(), token0.getFromAddress());
                            out.collect(token0);
                        } else {
                            log.warn("No price result for token0: {}", token0.getTokenAddress());
                            token0.setTokenPriceUsd(0.0);
                            out.collect(token0);
                        }
                    },
                    error -> {
                        log.error("Error getting token0 price for address {}: {}",
                                token0.getTokenAddress(), error.getMessage());
                        token0.setTokenPriceUsd(0.0);
                        out.collect(token0);
                    });
            priceLookup.asyncInvoke(metadata.getToken0Address(), token0Future);
        }

        // Process token1
        double amount1In = EthereumUtils.convertWeiToEth(args.get("amount1In"));
        double amount1Out = EthereumUtils.convertWeiToEth(args.get("amount1Out"));
        if (amount1In > 0 || amount1Out > 0) {
            Token token1 = new Token();
            token1.setTokenAddress(metadata.getToken1Address());
            token1.setTokenId(metadata.getToken1Id());
            token1.setAmount(amount1In + amount1Out);
            token1.setBuyOrSell(amount1Out > 0);
            token1.setFromAddress(fromAddress);

            // 异步获取token1价格
            ResultFuture<Double> token1Future = new SerializableResultFuture<>(
                    null,
                    result -> {
                        if (result != null && !result.isEmpty()) {
                            double price = result.iterator().next();
                            token1.setTokenPriceUsd(price);
                            log.debug("Processing token1: address={}, price={}, amount={}, from={}",
                                    token1.getTokenAddress(), price, token1.getAmount(), token1.getFromAddress());
                            out.collect(token1);
                        } else {
                            log.warn("No price result for token1: {}", token1.getTokenAddress());
                            token1.setTokenPriceUsd(0.0);
                            out.collect(token1);
                        }
                    },
                    error -> {
                        log.error("Error getting token1 price for address {}: {}",
                                token1.getTokenAddress(), error.getMessage());
                        token1.setTokenPriceUsd(0.0);
                        out.collect(token1);
                    });
            priceLookup.asyncInvoke(metadata.getToken1Address(), token1Future);
        }
    }

    @Override
    public void processBroadcastElement(PairMetadata metadata, Context ctx, Collector<Token> out) throws Exception {
        BroadcastState<String, PairMetadata> broadcastState = ctx.getBroadcastState(pairMetadataDescriptor);
        broadcastState.put(metadata.getPairAddress(), metadata);
    }
}