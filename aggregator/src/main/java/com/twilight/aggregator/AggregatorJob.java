package com.twilight.aggregator;

import java.time.Duration;

import org.apache.flink.api.common.eventtime.WatermarkStrategy;
import org.apache.flink.api.common.state.MapStateDescriptor;

import org.apache.flink.connector.kafka.source.enumerator.initializer.OffsetsInitializer;
import org.apache.flink.connector.kafka.source.KafkaSource;
import org.apache.flink.streaming.api.datastream.BroadcastStream;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.datastream.KeyedStream;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;
import org.apache.flink.util.OutputTag;
import org.apache.flink.streaming.api.datastream.SingleOutputStreamOperator;

import com.twilight.aggregator.config.FlinkConfig;
import com.twilight.aggregator.model.ProcessEvent;
import com.twilight.aggregator.model.KafkaMessage;
import com.twilight.aggregator.model.Pair;
import com.twilight.aggregator.model.PairMetadata;
import com.twilight.aggregator.model.PairMetric;
import com.twilight.aggregator.model.Token;
import com.twilight.aggregator.model.TokenRecentMetric;
import com.twilight.aggregator.process.EventExtractor;

import com.twilight.aggregator.serialization.KafkaMessageDeserializer;
import com.twilight.aggregator.sink.PostgresSink;
import com.twilight.aggregator.source.PairMetadataSource;
import com.twilight.aggregator.process.pair.PairWindowManager;
import com.twilight.aggregator.process.token.TokenWindowManager;
import com.twilight.aggregator.model.TokenRollingMetric;
import com.twilight.aggregator.process.EventEnrichmentProcessor;
import org.apache.flink.streaming.api.datastream.AsyncDataStream;
import java.util.concurrent.TimeUnit;
import com.twilight.aggregator.source.AsyncPriceLookupFunction;
import com.twilight.aggregator.process.EventSplitProcessor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class AggregatorJob {
        private static final Logger log = LoggerFactory.getLogger(AggregatorJob.class);
        private static final MapStateDescriptor<String, PairMetadata> PAIR_METADATA_DESCRIPTOR = new MapStateDescriptor<>(
                        "PairMetadataDescriptor", String.class, PairMetadata.class);

        private static final FlinkConfig config = FlinkConfig.getInstance();

        public static void main(String[] args) throws Exception {
                // Set up the execution environment
                final StreamExecutionEnvironment env = StreamExecutionEnvironment.getExecutionEnvironment();

                // Configure Kafka source
                KafkaSource<KafkaMessage> source = KafkaSource.<KafkaMessage>builder()
                                .setBootstrapServers(config.getKafkaBootstrapServers())
                                .setTopics(config.getKafkaTopic())
                                .setGroupId(config.getKafkaGroupId())
                                .setStartingOffsets(OffsetsInitializer.latest())
                                .setValueOnlyDeserializer(new KafkaMessageDeserializer())
                                .build();

                // Configure watermark strategies
                WatermarkStrategy<KafkaMessage> kafkaWatermarkStrategy = WatermarkStrategy
                                .<KafkaMessage>forBoundedOutOfOrderness(Duration.ofSeconds(5))
                                .withTimestampAssigner((event, timestamp) -> {
                                        long eventTime = event.getTransaction().getTimestamp();
                                        return eventTime;
                                })
                                .withIdleness(Duration.ofSeconds(30));

                WatermarkStrategy<Token> tokenWatermarkStrategy = WatermarkStrategy
                                .<Token>forBoundedOutOfOrderness(Duration.ofSeconds(5))
                                .withTimestampAssigner((token, timestamp) -> {
                                        long eventTime = token.getTimestamp();
                                        return eventTime;
                                })
                                .withIdleness(Duration.ofSeconds(30));

                WatermarkStrategy<Pair> pairWatermarkStrategy = WatermarkStrategy
                                .<Pair>forBoundedOutOfOrderness(Duration.ofSeconds(5))
                                .withTimestampAssigner((pair, timestamp) -> {
                                        long eventTime = pair.getTimestamp();
                                        return eventTime;
                                })
                                .withIdleness(Duration.ofSeconds(30));

                // Extract events from transactions
                DataStream<ProcessEvent> eventStream = env
                                .fromSource(source, kafkaWatermarkStrategy, "Kafka Source")
                                .flatMap(new EventExtractor())
                                .name("Event Extractor");

                // Create broadcast state for pair metadata
                BroadcastStream<PairMetadata> pairMetadataBroadcast = createPairMetadataBroadcast(env);
                DataStream<ProcessEvent> enrichedEventStream = eventStream
                                .connect(pairMetadataBroadcast)
                                .process(new EventEnrichmentProcessor(PAIR_METADATA_DESCRIPTOR))
                                .name("Event Enrichment");

                // Enrich events with price data using async I/O
                enrichedEventStream = AsyncDataStream.unorderedWait(
                                enrichedEventStream,
                                new AsyncPriceLookupFunction(),
                                5000, // 增加超时时间到5秒
                                TimeUnit.MILLISECONDS,
                                500 // 增加容量到500
                ).name("Price Lookup");

                // Split events into token and pair streams using side outputs
                EventSplitProcessor splitProcessor = new EventSplitProcessor();

                // 先处理事件分割，获取原始的处理结果流
                SingleOutputStreamOperator<Pair> processedStream = enrichedEventStream
                                .process(splitProcessor);
                log.info("Created processed stream with EventSplitProcessor");

                // 从处理结果流获取侧流（token流）
                OutputTag<Token> tokenOutputTag = splitProcessor.getTokenOutput();

                DataStream<Token> tokenStream = processedStream
                                .getSideOutput(tokenOutputTag)
                                .assignTimestampsAndWatermarks(tokenWatermarkStrategy)
                                .name("Token Stream");
                log.debug("Created token stream from side output");

                // 处理主流（pair流）
                SingleOutputStreamOperator<Pair> pairStream = processedStream
                                .assignTimestampsAndWatermarks(pairWatermarkStrategy)
                                .name("Pair Stream");
                log.debug("Created pair stream from processed stream");

                // Key the streams
                KeyedStream<Token, String> keyedTokenStream = tokenStream
                                .keyBy(Token::getTokenAddress);
                log.debug("Created keyed token stream");

                // Process tokens
                DataStream<TokenRecentMetric> tokenRecentMetrics = TokenWindowManager
                                .createSlidingHierarchicalWindows(keyedTokenStream);
                log.debug("Created token recent metrics stream");

                // DataStream<TokenRollingMetric> tokenRollingMetrics = TokenWindowManager
                // .createRollingHierarchicalWindows(keyedTokenStream);
                // log.debug("Created token rolling metrics stream");

                // Process pairs
                DataStream<PairMetric> pairMetrics = PairWindowManager.createHierarchicalWindows(
                                pairStream.keyBy(Pair::getPairId));
                log.debug("Created pair metrics stream");

                // Add sinks
                log.debug("Adding sink for pair metrics");
                pairMetrics
                                .map(metric -> {
                                        log.debug("Sending PairMetric to sink: {}", metric);
                                        return metric;
                                })
                                .addSink(PostgresSink.createPairMetricSink())
                                .name("Pair Metrics Sink");

                // log.debug("Adding sink for token recent metrics");
                tokenRecentMetrics
                                .map(metric -> {
                                        log.debug("Sending TokenRecentMetric to sink: {}", metric);
                                        return metric;
                                })
                                .addSink(PostgresSink.createTokenRecentMetricSink())
                                .name("Token Recent Metrics Sink");

                log.debug("Adding sink for token rolling metrics");
                // tokenRollingMetrics
                // .map(metric -> {
                // log.debug("Sending TokenRollingMetric to sink: {}", metric);
                // if (metric.getTokenId() == null) {
                // log.error("TokenRollingMetric has null tokenId: {}", metric);
                // }
                // if (metric.getTimeWindow() == null) {
                // log.error("TokenRollingMetric has null timeWindow: {}", metric);
                // // 如果timeWindow为null，设置默认值
                // metric.setTimeWindow("20s");
                // log.debug("Set default timeWindow for TokenRollingMetric: {}", metric);
                // }
                // return metric;
                // })
                // .filter(metric -> metric != null) // 过滤掉null值
                // .addSink(PostgresSink.createTokenRollingMetricSink())
                // .name("Token Rolling Metrics Sink");

                System.out.println("Job started...");
                log.info("Job started...");

                env.execute("DeFi Metrics Aggregator");
        }

        private static BroadcastStream<PairMetadata> createPairMetadataBroadcast(StreamExecutionEnvironment env) {
                // Create pair metadata source
                PairMetadataSource source = new PairMetadataSource(config.getPairMetadataRefreshInterval());
                return env.addSource(source)
                                .setParallelism(1)
                                .broadcast(PAIR_METADATA_DESCRIPTOR);
        }
}
