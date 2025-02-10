package com.twilight.aggregator;

import java.util.Properties;
import java.time.Duration;

import org.apache.flink.api.common.eventtime.WatermarkStrategy;
import org.apache.flink.api.common.state.MapStateDescriptor;
import org.apache.flink.api.common.typeinfo.TypeInformation;
import org.apache.flink.api.java.typeutils.MapTypeInfo;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.connector.kafka.source.enumerator.initializer.OffsetsInitializer;
import org.apache.flink.connector.kafka.source.KafkaSource;
import org.apache.flink.streaming.api.CheckpointingMode;
import org.apache.flink.streaming.api.datastream.BroadcastStream;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.datastream.KeyedStream;
import org.apache.flink.streaming.api.environment.CheckpointConfig.ExternalizedCheckpointCleanup;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;
import org.apache.flink.streaming.api.windowing.assigners.TumblingProcessingTimeWindows;
import org.apache.flink.streaming.api.windowing.time.Time;
import org.apache.flink.streaming.api.windowing.triggers.ProcessingTimeTrigger;

import com.twilight.aggregator.config.FlinkConfig;
import com.twilight.aggregator.model.Event;
import com.twilight.aggregator.model.KafkaMessage;
import com.twilight.aggregator.model.Pair;
import com.twilight.aggregator.model.PairMetadata;
import com.twilight.aggregator.model.PairMetric;
import com.twilight.aggregator.model.Token;
import com.twilight.aggregator.model.TokenMetric;
import com.twilight.aggregator.process.EventExtractor;
import com.twilight.aggregator.process.PairHierarchicalWindowAggregator;
import com.twilight.aggregator.process.PairMetadataProcessor;
import com.twilight.aggregator.process.PairWindowProcessor;
import com.twilight.aggregator.process.TokenHierarchicalWindowAggregator;
import com.twilight.aggregator.process.TokenMetadataProcessor;
import com.twilight.aggregator.process.TokenWindowProcessor;
import com.twilight.aggregator.serialization.KafkaMessageDeserializer;
import com.twilight.aggregator.sink.PostgresSink;
import com.twilight.aggregator.source.PairMetadataSource;

public class AggregatorJob {
    private static final MapStateDescriptor<String, PairMetadata> PAIR_METADATA_DESCRIPTOR = new MapStateDescriptor<>(
            "PairMetadata", String.class, PairMetadata.class);

    private static final FlinkConfig config = FlinkConfig.getInstance();

    public static void main(String[] args) throws Exception {
        // 创建流执行环境
        Configuration configuration = new Configuration();
        // 配置本地 Web UI，使用不同的端口避免冲突
        configuration.setString("rest.bind-port", "8083");
        configuration.setString("rest.address", "localhost");

        // 资源配置
        configuration.setString("taskmanager.memory.network.fraction", "0.3");
        configuration.setInteger("taskmanager.numberOfTaskSlots", 1);
        configuration.setString("taskmanager.memory.task.heap.size", "2048m");
        configuration.setString("taskmanager.memory.task.off-heap.size", "512m");
        configuration.setString("taskmanager.memory.network.min", "512mb");
        configuration.setString("taskmanager.memory.network.max", "512mb");
        configuration.setString("jobmanager.memory.process.size", "1024m");

        // 缓冲区配置
        configuration.setString("taskmanager.network.memory.buffer-size", "32kb");
        configuration.setString("taskmanager.network.memory.floating-buffers-per-gate", "1024");
        configuration.setString("taskmanager.network.memory.buffers-per-channel", "8");
        configuration.setString("taskmanager.network.detailed-metrics", "false");
        configuration.setBoolean("taskmanager.network.blocking-shuffle.compression.enabled", false);

        // 配置状态后端
        configuration.setString("state.backend", "hashmap");
        configuration.setString("state.backend.incremental", "true");
        configuration.setString("state.checkpoints.num-retained", "2");

        // 配置重启策略
        configuration.setString("restart-strategy", "fixed-delay");
        configuration.setString("restart-strategy.fixed-delay.attempts", "3");
        configuration.setString("restart-strategy.fixed-delay.delay", "10s");

        final StreamExecutionEnvironment env = StreamExecutionEnvironment
                .createLocalEnvironmentWithWebUI(configuration);
        env.setParallelism(1);
        env.setBufferTimeout(100); // 增加缓冲超时时间

        // 配置序列化
        env.getConfig().enableObjectReuse();

        // 创建 Kafka source with JSON deserializer
        KafkaSource<KafkaMessage> source = KafkaSource.<KafkaMessage>builder()
                .setBootstrapServers(config.getKafkaBootstrapServers())
                .setTopics(config.getKafkaTopic())
                .setGroupId(config.getKafkaGroupId())
                .setStartingOffsets(OffsetsInitializer.latest())
                .setValueOnlyDeserializer(new KafkaMessageDeserializer())
                .setProperties(new Properties() {
                    {
                        setProperty("fetch.min.bytes", "1");
                        setProperty("fetch.max.wait.ms", "100");
                        setProperty("max.poll.records", "500");
                        setProperty("enable.auto.commit", "false");
                        setProperty("auto.offset.reset", "latest");
                        setProperty("isolation.level", "read_committed");
                    }
                })
                .build();

        // Create main stream with simple watermark strategy
        WatermarkStrategy<KafkaMessage> watermarkStrategy = WatermarkStrategy
                .<KafkaMessage>forBoundedOutOfOrderness(Duration.ofSeconds(5))
                .withTimestampAssigner((message, timestamp) -> message.getTransaction().getTimestamp())
                .withIdleness(Duration.ofMinutes(1)); // 允许窗口在无数据时保持对齐

        // Extract events from transactions
        DataStream<Event> eventStream = env
                .fromSource(source, watermarkStrategy, "Kafka Source")
                .flatMap(new EventExtractor())
                .name("Event Extractor");

        // Create broadcast state for pair metadata
        BroadcastStream<PairMetadata> pairMetadataBroadcast = createPairMetadataBroadcast(env);

        // Process pairs
        DataStream<PairMetric> pairMetrics = processWithWindow(
                eventStream
                        .connect(pairMetadataBroadcast)
                        .process(new PairMetadataProcessor(PAIR_METADATA_DESCRIPTOR))
                        .keyBy(pair -> pair.getPairAddress()),
                "pair");

        // Process tokens
        DataStream<TokenMetric> tokenMetrics = processWithWindow(
                eventStream
                        .connect(pairMetadataBroadcast)
                        .process(new TokenMetadataProcessor(PAIR_METADATA_DESCRIPTOR))
                        .keyBy(Token::getTokenAddress),
                "token");

        // Add sinks
        pairMetrics.addSink(PostgresSink.createPairMetricSink());
        tokenMetrics.addSink(PostgresSink.createTokenMetricSink());
        System.out.println("Job started..."); // 确保这行打印出现

        env.execute("DeFi Metrics Aggregator");
    }

    private static BroadcastStream<PairMetadata> createPairMetadataBroadcast(StreamExecutionEnvironment env) {
        // Create pair metadata source
        PairMetadataSource source = new PairMetadataSource(config.getPairMetadataRefreshInterval());
        return env.addSource(source)
                .setParallelism(1)
                .broadcast(PAIR_METADATA_DESCRIPTOR);
    }

    private static <T, M> DataStream<M> processWithWindow(
            KeyedStream<T, String> stream,
            String metricType) {

        DataStream<M> result = null;

        if ("token".equals(metricType)) {
            @SuppressWarnings("unchecked")
            KeyedStream<Token, String> tokenStream = (KeyedStream<Token, String>) stream;

            // Base window (20s)
            DataStream<TokenMetric> baseWindow = tokenStream
                    .window(TumblingProcessingTimeWindows.of(
                            Time.seconds(20)))
                    .process(new TokenWindowProcessor("20s"))
                    .name("20s-token-window");

            // 1-minute window from 20s windows
            DataStream<TokenMetric> oneMinWindow = baseWindow
                    .keyBy(metric -> metric.getTokenId().toString())
                    .window(TumblingProcessingTimeWindows.of(
                            Time.minutes(1)))
                    .process(new TokenHierarchicalWindowAggregator("1min"))
                    .name("1min-token-window");

            // 5-minute window from 1-minute windows
            DataStream<TokenMetric> fiveMinWindow = oneMinWindow
                    .keyBy(metric -> metric.getTokenId().toString())
                    .window(TumblingProcessingTimeWindows.of(
                            Time.minutes(5)))
                    .process(new TokenHierarchicalWindowAggregator("5min"))
                    .name("5min-token-window");

            // 30-minute window from 5-minute windows
            DataStream<TokenMetric> thirtyMinWindow = fiveMinWindow
                    .keyBy(metric -> metric.getTokenId().toString())
                    .window(TumblingProcessingTimeWindows.of(
                            Time.minutes(30)))
                    .process(new TokenHierarchicalWindowAggregator("30min"))
                    .name("30min-token-window");

            // Merge all window results
            result = (DataStream<M>) baseWindow
                    .union(oneMinWindow)
                    .union(fiveMinWindow)
                    .union(thirtyMinWindow);

        } else if ("pair".equals(metricType)) {
            @SuppressWarnings("unchecked")
            KeyedStream<Pair, String> pairStream = (KeyedStream<Pair, String>) stream;

            // Base window (20s)
            DataStream<PairMetric> baseWindow = pairStream
                    .window(TumblingProcessingTimeWindows.of(
                            Time.seconds(20)))
                    .process(new PairWindowProcessor("20s"))
                    .name("20s-pair-window");

            // 1-minute window from 20s windows
            DataStream<PairMetric> oneMinWindow = baseWindow
                    .keyBy(metric -> metric.getPairId().toString())
                    .window(TumblingProcessingTimeWindows.of(
                            Time.minutes(1)))
                    .process(new PairHierarchicalWindowAggregator("1min"))
                    .name("1min-pair-window");

            // 5-minute window from 1-minute windows
            DataStream<PairMetric> fiveMinWindow = oneMinWindow
                    .keyBy(metric -> metric.getPairId().toString())
                    .window(TumblingProcessingTimeWindows.of(
                            Time.minutes(5)))
                    .process(new PairHierarchicalWindowAggregator("5min"))
                    .name("5min-pair-window");

            // 30-minute window from 5-minute windows
            DataStream<PairMetric> thirtyMinWindow = fiveMinWindow
                    .keyBy(metric -> metric.getPairId().toString())
                    .window(TumblingProcessingTimeWindows.of(
                            Time.minutes(30)))
                    .process(new PairHierarchicalWindowAggregator("30min"))
                    .name("30min-pair-window");

            // Merge all window results
            result = (DataStream<M>) baseWindow
                    .union(oneMinWindow)
                    .union(fiveMinWindow)
                    .union(thirtyMinWindow);
        }

        return result;
    }
}
