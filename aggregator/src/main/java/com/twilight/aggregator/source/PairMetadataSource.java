package com.twilight.aggregator.source;

import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.HashMap;
import java.util.Map;

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
        // 查询 pair 和 token 信息
        String pairSql = "SELECT " +
                "p.id as pair_id, " +
                "p.pair_address, " +
                "p.token0_id, " +
                "p.token1_id, " +
                "t0.token_address as token0_address, " +
                "t1.token_address as token1_address " +
                "FROM twswap_pair p " +
                "JOIN token t0 ON p.token0_id = t0.id " +
                "JOIN token t1 ON p.token1_id = t1.id " +
                "WHERE p.chain_id = '31337'"; // 假设 chain_id 为 31337，可以从配置中获取

        // 创建 PairMetadata 映射，用于后续添加 tag 信息
        Map<String, PairMetadata> metadataMap = new HashMap<>();

        try (Connection conn = dataSource.getConnection();
                PreparedStatement stmt = conn.prepareStatement(pairSql);
                ResultSet rs = stmt.executeQuery()) {

            while (rs.next()) {
                PairMetadata metadata = new PairMetadata();
                metadata.setPairId(rs.getLong("pair_id"));
                metadata.setPairAddress(rs.getString("pair_address"));
                metadata.setToken0Id(rs.getLong("token0_id"));
                metadata.setToken1Id(rs.getLong("token1_id"));
                metadata.setToken0Address(rs.getString("token0_address"));
                metadata.setToken1Address(rs.getString("token1_address"));

                // 将 metadata 添加到映射中
                metadataMap.put(metadata.getPairAddress().toLowerCase(), metadata);
            }
        }

        // 如果没有 pair 数据，直接返回
        if (metadataMap.isEmpty()) {
            LOG.warn("No pair metadata found in database");
            return;
        }

        // 查询 account 表中的 tag 信息
        String accountSql = "SELECT address, " +
                "smart_money_tag, cex_tag, big_whale_tag, fresh_wallet_tag " +
                "FROM account " +
                "WHERE smart_money_tag = true OR cex_tag = true OR " +
                "big_whale_tag = true OR fresh_wallet_tag = true";

        try (Connection conn = dataSource.getConnection();
                PreparedStatement stmt = conn.prepareStatement(accountSql);
                ResultSet rs = stmt.executeQuery()) {

            while (rs.next()) {
                String address = rs.getString("address");
                String tag = "all"; // 默认标签

                // 确定标签，只会有一个为 true
                if (rs.getBoolean("smart_money_tag")) {
                    tag = "smart_money";
                } else if (rs.getBoolean("cex_tag")) {
                    tag = "cex";
                } else if (rs.getBoolean("big_whale_tag")) {
                    tag = "whale"; // 将 big_whale_tag 对应的标签从 dex 改为 whale
                } else if (rs.getBoolean("fresh_wallet_tag")) {
                    tag = "fresh_wallet";
                }

                // 为所有 PairMetadata 添加地址标签
                for (PairMetadata metadata : metadataMap.values()) {
                    metadata.addAddressTag(address, tag);
                    LOG.debug("Added tag for address: {}, tag: {}", address, tag);
                }
            }
        }

        // 发送所有 PairMetadata 到下游
        for (PairMetadata metadata : metadataMap.values()) {
            ctx.collect(metadata);
            LOG.debug("Emitted pair metadata: {}", metadata);
        }
    }

    @Override
    public void cancel() {
        isRunning = false;
        shutdownScheduler();
    }
}