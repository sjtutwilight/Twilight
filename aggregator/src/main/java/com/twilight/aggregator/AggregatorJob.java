package com.twilight.aggregator;

import org.apache.flink.api.common.eventtime.WatermarkStrategy;
import org.apache.flink.api.common.state.MapStateDescriptor;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.configuration.RestOptions;
import org.apache.flink.connector.kafka.source.KafkaSource;
import org.apache.flink.streaming.api.CheckpointingMode;
import org.apache.flink.streaming.api.datastream.BroadcastStream;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.datastream.KeyedStream;
import org.apache.flink.streaming.api.environment.CheckpointConfig.ExternalizedCheckpointCleanup;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;
import org.apache.flink.streaming.api.functions.windowing.ProcessWindowFunction;
import org.apache.flink.streaming.api.windowing.assigners.TumblingProcessingTimeWindows;
import org.apache.flink.streaming.api.windowing.time.Time;
import org.apache.flink.streaming.api.windowing.windows.TimeWindow;

import com.twilight.aggregator.config.FlinkConfig;
import com.twilight.aggregator.model.Event;
import com.twilight.aggregator.model.PairMetadata;
import com.twilight.aggregator.model.PairMetric;
import com.twilight.aggregator.model.Token;
import com.twilight.aggregator.model.TokenMetric;
import com.twilight.aggregator.model.Transaction;
import com.twilight.aggregator.process.EventExtractor;
import com.twilight.aggregator.process.PairMetadataProcessor;
import com.twilight.aggregator.process.PairWindowProcessor;
import com.twilight.aggregator.process.TokenMetadataProcessor;
import com.twilight.aggregator.process.TokenWindowProcessor;
import com.twilight.aggregator.sink.PostgresSink;
import com.twilight.aggregator.source.PairMetadataSource;

public class AggregatorJob {
    private static final MapStateDescriptor<String, PairMetadata> PAIR_METADATA_DESCRIPTOR =
        new MapStateDescriptor<>("PairMetadata", String.class, PairMetadata.class);
    
    private static final FlinkConfig config = FlinkConfig.getInstance();

    public static void main(String[] args) throws Exception {
        // 创建流执行环境
        Configuration configuration = new Configuration();
        configuration.setString(RestOptions.BIND_PORT, "8090-8099");
        configuration.setString(RestOptions.ADDRESS, "localhost");
        configuration.setString("taskmanager.memory.network.fraction", "0.1");
        configuration.setInteger("taskmanager.numberOfTaskSlots", 4);
        configuration.setString("taskmanager.memory.task.heap.size", "1024m");
        configuration.setString("taskmanager.memory.task.off-heap.size", "1024m");
        configuration.setString("jobmanager.memory.process.size", "1024m");

        final StreamExecutionEnvironment env = StreamExecutionEnvironment.getExecutionEnvironment(configuration);
        env.setParallelism(FlinkConfig.getInstance().getParallelism());
        
        // 配置 JSON 序列化
        env.getConfig().enableObjectReuse();
        
        // 设置检查点和状态后端配置
        env.enableCheckpointing(FlinkConfig.getInstance().getCheckpointInterval());
        env.getCheckpointConfig().setCheckpointingMode(CheckpointingMode.EXACTLY_ONCE);
        env.getCheckpointConfig().setMinPauseBetweenCheckpoints(5000);
        env.getCheckpointConfig().setCheckpointTimeout(60000);
        env.getCheckpointConfig().setMaxConcurrentCheckpoints(1);
        env.getCheckpointConfig().enableExternalizedCheckpoints(ExternalizedCheckpointCleanup.RETAIN_ON_CANCELLATION);
        
        // 配置缓冲区
        configuration.setInteger("taskmanager.memory.network.max", 64);
        configuration.setInteger("taskmanager.memory.network.min", 64);
        configuration.setBoolean("taskmanager.network.detailed-metrics", true);
        
        // 创建 Kafka source
        KafkaSource<Transaction> source = FlinkConfig.getInstance().buildKafkaSource();
        
        // Create main stream
        DataStream<Transaction> transactionStream = env
            .fromSource(source, WatermarkStrategy.noWatermarks(), "Kafka Source");

        // Create broadcast state for pair metadata
        BroadcastStream<PairMetadata> pairMetadataBroadcast = createPairMetadataBroadcast(env);

        // Extract events
        DataStream<Event> eventStream = transactionStream
            .flatMap(new EventExtractor());

        // Process pairs
        DataStream<PairMetric> pairMetrics = processWithWindow(
            eventStream
                .connect(pairMetadataBroadcast)
                .process(new PairMetadataProcessor(PAIR_METADATA_DESCRIPTOR))
                .keyBy(pair -> pair.getPairAddress()),
            "pair"
        );

        // Process tokens
        DataStream<TokenMetric> tokenMetrics = processWithWindow(
            eventStream
                .connect(pairMetadataBroadcast)
                .process(new TokenMetadataProcessor(PAIR_METADATA_DESCRIPTOR))
                .keyBy(Token::getTokenAddress),
            "token"
        );

        // Add sinks
        pairMetrics.addSink(PostgresSink.createPairMetricSink()); 
        tokenMetrics.addSink(PostgresSink.createTokenMetricSink());
        System.out.println("Job started...");  // 确保这行打印出现

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
        String[] windowSizes = config.getWindowSizes();
        String[] windowNames = config.getWindowNames();
        
        for (int i = 0; i < windowSizes.length; i++) {
            Time windowSize = parseWindowSize(windowSizes[i]);
            String windowName = windowNames[i];
            
            ProcessWindowFunction<T, M, String, TimeWindow> windowFunction = null;
            if ("pair".equals(metricType)) {
                windowFunction = (ProcessWindowFunction<T, M, String, TimeWindow>) new PairWindowProcessor(windowName);
            } else if ("token".equals(metricType)) {
                windowFunction = (ProcessWindowFunction<T, M, String, TimeWindow>) new TokenWindowProcessor(windowName);
            }
            
            DataStream<M> windowedStream = stream
                .window(TumblingProcessingTimeWindows.of(windowSize))
                .process(windowFunction);
            
            result = (result == null) ? windowedStream : result.union(windowedStream);
        }
        
        return result;
    }

    private static Time parseWindowSize(String windowSize) {
        if (!windowSize.matches("\\d+[smh]")) {
            throw new IllegalArgumentException("Invalid window size format: " + windowSize + 
                ". Must be a number followed by s(seconds), m(minutes), or h(hours)");
        }

        String value = windowSize.substring(0, windowSize.length() - 1);
        String unit = windowSize.substring(windowSize.length() - 1);
        long size = Long.parseLong(value);

        switch (unit.toLowerCase()) {
            case "s":
                return Time.seconds(size);
            case "m":
                return Time.minutes(size);
            case "h":
                return Time.hours(size);
            default:
                throw new IllegalArgumentException("Unsupported time unit: " + unit);
        }
    }
} 
