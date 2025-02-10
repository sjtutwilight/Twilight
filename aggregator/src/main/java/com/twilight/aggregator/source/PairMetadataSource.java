package com.twilight.aggregator.source;

import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

import javax.sql.DataSource;

import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.source.RichSourceFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.twilight.aggregator.config.DatabaseConfig;
import com.twilight.aggregator.model.PairMetadata;

public class PairMetadataSource extends RichSourceFunction<PairMetadata> {
    private static transient Logger LOG;
    private static final long serialVersionUID = 1L;

    private final long refreshInterval;
    private volatile boolean isRunning = true;
    private transient DataSource dataSource;
    private transient ScheduledExecutorService scheduler;

    public PairMetadataSource(long refreshInterval) {
        this.refreshInterval = refreshInterval;
    }

    @Override
    public void open(Configuration parameters) {
        dataSource = DatabaseConfig.getInstance().getHikariDataSource();
        LOG = LoggerFactory.getLogger(PairMetadataSource.class);
        scheduler = Executors.newSingleThreadScheduledExecutor();
    }

    @Override
    public void run(SourceContext<PairMetadata> ctx) throws Exception {
        scheduler.scheduleAtFixedRate(() -> {
            if (!isRunning) {
                return;
            }
            try {
                queryAndEmit(ctx);
            } catch (Exception e) {
                LOG.error("Error querying pair metadata", e);
                if (e instanceof InterruptedException) {
                    Thread.currentThread().interrupt();
                }
            }
        }, 0, refreshInterval, TimeUnit.MILLISECONDS);

        // Wait until the source is canceled
        while (isRunning) {
            try {
                Thread.sleep(1000);
            } catch (InterruptedException e) {
                // Restore the interrupted status
                Thread.currentThread().interrupt();
                break;
            }
        }

        // Shutdown the scheduler gracefully
        shutdownScheduler();
    }

    private void shutdownScheduler() {
        if (scheduler != null && !scheduler.isShutdown()) {
            try {
                scheduler.shutdown();
                if (!scheduler.awaitTermination(5, TimeUnit.SECONDS)) {
                    scheduler.shutdownNow();
                }
            } catch (InterruptedException e) {
                scheduler.shutdownNow();
                Thread.currentThread().interrupt();
            }
        }
    }

    private void queryAndEmit(SourceContext<PairMetadata> ctx) throws SQLException {
        String sql = "SELECT " +
                "p.id as pair_id, " +
                "p.pair_address, " +
                "p.token0_id, " +
                "p.token1_id, " +
                "t0.token_address as token0_address, " +
                "t1.token_address as token1_address " +
                "FROM twswap_pair p " +
                "JOIN token t0 ON p.token0_id = t0.id " +
                "JOIN token t1 ON p.token1_id = t1.id " +
                "WHERE p.chain_id = '31337'"; // 假设 chain_id 为 1，可以从配置中获取

        try (Connection conn = dataSource.getConnection();
                PreparedStatement stmt = conn.prepareStatement(sql);
                ResultSet rs = stmt.executeQuery()) {

            while (rs.next()) {
                PairMetadata metadata = new PairMetadata();
                metadata.setPairId(rs.getLong("pair_id"));
                metadata.setPairAddress(rs.getString("pair_address"));
                metadata.setToken0Id(rs.getLong("token0_id"));
                metadata.setToken1Id(rs.getLong("token1_id"));
                metadata.setToken0Address(rs.getString("token0_address"));
                metadata.setToken1Address(rs.getString("token1_address"));

                synchronized (ctx.getCheckpointLock()) {
                    ctx.collect(metadata);
                }
            }
            LOG.info("Successfully loaded pair metadata from database");
        } catch (SQLException e) {
            LOG.error("Failed to query pair metadata: {}", e.getMessage());
            throw e;
        }
    }

    @Override
    public void cancel() {
        isRunning = false;
        shutdownScheduler();
    }
}