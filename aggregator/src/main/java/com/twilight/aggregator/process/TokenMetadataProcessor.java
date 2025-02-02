package com.twilight.aggregator.process;

import java.math.BigDecimal;
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

public class TokenMetadataProcessor extends BroadcastProcessFunction<Event, PairMetadata, Token> {
    private static final long serialVersionUID = 1L;
    private final MapStateDescriptor<String, PairMetadata> pairMetadataDescriptor;
    private final Logger log = LoggerFactory.getLogger(TokenMetadataProcessor.class);
    private final AsyncPriceLookupFunction priceLookup;

    public TokenMetadataProcessor(MapStateDescriptor<String, PairMetadata> descriptor) {
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
    public void processElement(Event event, ReadOnlyContext ctx, Collector<Token> out) throws Exception {
        if ("Swap".equals(event.getEventName())) {
            ReadOnlyBroadcastState<String, PairMetadata> broadcastState = ctx.getBroadcastState(pairMetadataDescriptor);
            PairMetadata metadata = broadcastState.get(event.getContractAddress());
            
            if (metadata == null) {
                log.error("Pair metadata is null for contract address: {}", event.getContractAddress());
                return;
            }

            Map<String, Object> args = event.getDecodedArgs();
            processSwapEvent(metadata, args, event.getFromAddress(), out);
        }
    }

    private void processSwapEvent(PairMetadata metadata, Map<String, Object> args, String fromAddress, Collector<Token> out) {
        // Process token0
        BigDecimal amount0In = EthereumUtils.convertWeiToEth(args.get("amount0In").toString());
        BigDecimal amount0Out = EthereumUtils.convertWeiToEth(args.get("amount0Out").toString());
        if (amount0In.compareTo(BigDecimal.ZERO) > 0 || amount0Out.compareTo(BigDecimal.ZERO) > 0) {
            Token token0 = new Token();
            token0.setTokenAddress(metadata.getToken0Address());
            token0.setTokenId(metadata.getToken0Id());
            token0.setAmount(amount0In.add(amount0Out));
            token0.setBuyOrSell(amount0Out.compareTo(BigDecimal.ZERO) > 0);
            token0.setFromAddress(fromAddress);

            // 异步获取token0价格
            ResultFuture<BigDecimal> token0Future = new ResultFuture<BigDecimal>() {
                @Override
                public void complete(Collection<BigDecimal> result) {
                    token0.setTokenPriceUsd(result.iterator().next());
                    if (isValidToken(token0)) {
                        log.info("TokenMetadataProcessor processElement: {}", token0);
                        out.collect(token0);
                    }
                }

                @Override
                public void completeExceptionally(Throwable error) {
                    log.error("Error getting token0 price", error);
                    token0.setTokenPriceUsd(BigDecimal.ZERO);
                    if (isValidToken(token0)) {
                        log.info("TokenMetadataProcessor processElement: {}", token0);
                        out.collect(token0);
                    }
                }
            };
            priceLookup.asyncInvoke(metadata.getToken0Address(), token0Future);
        }

        // Process token1
        BigDecimal amount1In = EthereumUtils.convertWeiToEth(args.get("amount1In").toString());
        BigDecimal amount1Out = EthereumUtils.convertWeiToEth(args.get("amount1Out").toString());
        if (amount1In.compareTo(BigDecimal.ZERO) > 0 || amount1Out.compareTo(BigDecimal.ZERO) > 0) {
            Token token1 = new Token();
            token1.setTokenAddress(metadata.getToken1Address());
            token1.setTokenId(metadata.getToken1Id());
            token1.setAmount(amount1In.add(amount1Out));
            token1.setBuyOrSell(amount1Out.compareTo(BigDecimal.ZERO) > 0);
            token1.setFromAddress(fromAddress);

            // 异步获取token1价格
            ResultFuture<BigDecimal> token1Future = new ResultFuture<BigDecimal>() {
                @Override
                public void complete(Collection<BigDecimal> result) {
                    token1.setTokenPriceUsd(result.iterator().next());
                    if (isValidToken(token1)) {
                        out.collect(token1);
                    }
                }

                @Override
                public void completeExceptionally(Throwable error) {
                    log.error("Error getting token1 price", error);
                    token1.setTokenPriceUsd(BigDecimal.ZERO);
                    if (isValidToken(token1)) {
                        out.collect(token1);
                    }
                }
            };
            priceLookup.asyncInvoke(metadata.getToken1Address(), token1Future);
        }
    }

    private boolean isValidToken(Token token) {
        return token.getTokenAddress() != null && 
               token.getTokenId() != null && 
               token.getAmount() != null && 
               token.getFromAddress() != null;
    }

    @Override
    public void processBroadcastElement(PairMetadata metadata, Context ctx, Collector<Token> out) throws Exception {
        BroadcastState<String, PairMetadata> broadcastState = ctx.getBroadcastState(pairMetadataDescriptor);
        broadcastState.put(metadata.getPairAddress(), metadata);
    }
} 