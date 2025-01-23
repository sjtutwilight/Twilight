package com.twilight.aggregator.job;

import org.apache.flink.api.common.eventtime.WatermarkStrategy;
import org.apache.flink.api.common.functions.FlatMapFunction;
import org.apache.flink.api.common.serialization.SimpleStringSchema;
import org.apache.flink.api.common.typeinfo.TypeHint;
import org.apache.flink.api.common.typeinfo.TypeInformation;
import org.apache.flink.api.java.functions.KeySelector;
import org.apache.flink.connector.jdbc.JdbcConnectionOptions;
import org.apache.flink.connector.jdbc.JdbcExecutionOptions;
import org.apache.flink.connector.jdbc.JdbcSink;
import org.apache.flink.connector.kafka.source.KafkaSource;
import org.apache.flink.connector.kafka.source.enumerator.initializer.OffsetsInitializer;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.datastream.SingleOutputStreamOperator;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;
import org.apache.flink.streaming.api.windowing.assigners.TumblingEventTimeWindows;
import org.apache.flink.streaming.api.windowing.time.Time;
import org.apache.flink.util.Collector;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.twilight.aggregator.config.AppConfig;
import com.twilight.aggregator.model.Event;
import com.twilight.aggregator.model.KafkaMessage;
import com.twilight.aggregator.model.PairMetric;
import com.twilight.aggregator.processor.PairEventProcessor;

public class PairMetricsJob {
    private static final Logger LOG = LoggerFactory.getLogger(PairMetricsJob.class);

    static {
        // Add JVM arguments to allow reflective access
        System.setProperty("--add-opens", "java.base/java.util=ALL-UNNAMED");
        System.setProperty("--add-opens", "java.base/java.lang=ALL-UNNAMED");
        System.setProperty("--add-opens", "java.base/java.lang.invoke=ALL-UNNAMED");
    }

    public static void main(String[] args) throws Exception {
        // Load configuration using builder pattern
        AppConfig config = AppConfig.builder()
                .withEnvironment(System.getProperty("env", "dev"))
                .build();
                
        LOG.info("Starting PairMetricsJob with config: {}", config);

        // Create execution environment
        StreamExecutionEnvironment env = StreamExecutionEnvironment.getExecutionEnvironment();
        
        // Configure environment based on the running environment
        if (config.getEnvironment() == AppConfig.Environment.DEV) {
            env.setParallelism(1);
            env.enableCheckpointing(30000); // 30 seconds
        } else {
            env.enableCheckpointing(60000); // 1 minute
        }

        // Create Kafka source
        KafkaSource<String> source = KafkaSource.<String>builder()
                .setBootstrapServers(config.getKafkaBootstrapServers())
                .setTopics(config.getInputTopic())
                .setGroupId(config.getKafkaGroupId())
                .setStartingOffsets(OffsetsInitializer.latest())
                .setValueOnlyDeserializer(new SimpleStringSchema())
                .build();

        // Create the data stream
        DataStream<String> stream = env.fromSource(
                source,
                WatermarkStrategy.<String>forMonotonousTimestamps()
                        .withTimestampAssigner((event, timestamp) -> System.currentTimeMillis()),
                "Kafka Source"
        );

        // Deserialize JSON to KafkaMessage and flatMap events
        SingleOutputStreamOperator<Event> events = stream
            .map(json -> {
                ObjectMapper mapper = new ObjectMapper();
                return mapper.readValue(json, KafkaMessage.class);
            })
            .returns(TypeInformation.of(KafkaMessage.class))
            .flatMap(new FlatMapFunction<KafkaMessage, Event>() {
                private static final long serialVersionUID = 1L;
                
                @Override
                public void flatMap(KafkaMessage message, Collector<Event> out) throws Exception {
                    if (message.getEvents() != null) {
                        for (Event event : message.getEvents()) {
                            // Only collect pair-related events (Sync, Swap, Mint, Burn)
                            String eventName = event.getEventName().toLowerCase();
                            if (eventName.equals("sync") || eventName.equals("swap") || 
                                eventName.equals("mint") || eventName.equals("burn")) {
                                out.collect(event);
                            }
                        }
                    }
                }
            })
            .returns(TypeInformation.of(new TypeHint<Event>() {}));

        // Process events with different time windows
        processEventsWithWindow(events, Time.seconds(20), "10s");
        processEventsWithWindow(events, Time.minutes(1), "1min");
        processEventsWithWindow(events, Time.minutes(5), "5min");
        processEventsWithWindow(events, Time.minutes(30), "30min");
        processEventsWithWindow(events, Time.hours(1), "1h");
        // Execute the job
        env.execute("Pair Metrics Aggregator");
    }

    private static void processEventsWithWindow(
            DataStream<Event> events,
            Time windowSize,
            String windowName
    ) {
        events
            .keyBy(new KeySelector<Event, String>() {
                private static final long serialVersionUID = 1L;
                
                @Override
                public String getKey(Event event) {
                    return event.getContractAddress();
                }
            })
            .window(TumblingEventTimeWindows.of(windowSize))
            .process(new PairEventProcessor(windowSize, windowName))
            .name(String.format("Process-%s-window", windowName))
            .addSink(JdbcSink.sink(
                "INSERT INTO twswap_pair_metric (" +
                    "pair_id, time_window, end_time, " +
                    "token0_reserve, token1_reserve, reserve_usd, " +
                    "token0_volume_usd, token1_volume_usd, volume_usd, txcnt" +
                ") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) " +
                "ON CONFLICT (pair_id, time_window, end_time) DO UPDATE SET " +
                    "token0_reserve = EXCLUDED.token0_reserve, " +
                    "token1_reserve = EXCLUDED.token1_reserve, " +
                    "reserve_usd = EXCLUDED.reserve_usd, " +
                    "token0_volume_usd = EXCLUDED.token0_volume_usd, " +
                    "token1_volume_usd = EXCLUDED.token1_volume_usd, " +
                    "volume_usd = EXCLUDED.volume_usd, " +
                    "txcnt = EXCLUDED.txcnt",
                (statement, metric) -> {
                    statement.setLong(1, metric.getPairId());
                    statement.setString(2, metric.getTimeWindow());
                    statement.setTimestamp(3, metric.getEndTime());
                    statement.setBigDecimal(4, metric.getToken0Reserve());
                    statement.setBigDecimal(5, metric.getToken1Reserve());
                    statement.setBigDecimal(6, metric.getReserveUsd());
                    statement.setBigDecimal(7, metric.getToken0VolumeUsd());
                    statement.setBigDecimal(8, metric.getToken1VolumeUsd());
                    statement.setBigDecimal(9, metric.getVolumeUsd());
                    statement.setInt(10, metric.getTxcnt());
                },
                JdbcExecutionOptions.builder()
                    .withBatchSize(1000)
                    .withBatchIntervalMs(200)
                    .withMaxRetries(3)
                    .build(),
                new JdbcConnectionOptions.JdbcConnectionOptionsBuilder()
                    .withUrl(AppConfig.builder().build().getJdbcUrl())
                    .withDriverName("org.postgresql.Driver")
                    .withUsername(AppConfig.builder().build().getJdbcUser())
                    .withPassword(AppConfig.builder().build().getJdbcPassword())
                    .build()
            ))
            .name(String.format("Save-%s-metrics", windowName));
    }
} 