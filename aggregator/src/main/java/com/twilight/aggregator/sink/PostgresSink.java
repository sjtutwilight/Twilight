package com.twilight.aggregator.sink;

import java.sql.Connection;
import java.sql.PreparedStatement;

import javax.sql.DataSource;

import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.sink.RichSinkFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.twilight.aggregator.config.DatabaseConfig;
import com.twilight.aggregator.model.PairMetric;
import com.twilight.aggregator.model.TokenMetric;

public class PostgresSink<T> extends RichSinkFunction<T> {
    private static final Logger log = LoggerFactory.getLogger(PostgresSink.class);
    private static final long serialVersionUID = 1L;

    private transient DataSource dataSource;
    private final String insertSql;

    private PostgresSink(String insertSql) {
        this.insertSql = insertSql;
    }

    @Override
    public void open(Configuration parameters) throws Exception {
        dataSource = DatabaseConfig.getInstance().getHikariDataSource();
    }

    @Override
    public void invoke(T value, Context context) throws Exception {
        try (Connection conn = dataSource.getConnection();
             PreparedStatement stmt = conn.prepareStatement(insertSql)) {
            
            if (value instanceof PairMetric) {
                PairMetric metric = (PairMetric) value;
                stmt.setLong(1, metric.getPairId());
                stmt.setString(2, metric.getTimeWindow());
                stmt.setTimestamp(3, metric.getEndTime());
                stmt.setBigDecimal(4, metric.getToken0Reserve());
                stmt.setBigDecimal(5, metric.getToken1Reserve());
                stmt.setBigDecimal(6, metric.getReserveUsd());
                stmt.setBigDecimal(7, metric.getToken0VolumeUsd());
                stmt.setBigDecimal(8, metric.getToken1VolumeUsd());
                stmt.setBigDecimal(9, metric.getVolumeUsd());
                stmt.setInt(10, metric.getTxcnt());
            } else if (value instanceof TokenMetric) {
                TokenMetric metric = (TokenMetric) value;
                stmt.setLong(1, metric.getTokenId());
                stmt.setString(2, metric.getTimeWindow());
                stmt.setTimestamp(3, metric.getEndTime());
                stmt.setBigDecimal(4, metric.getVolumeUsd());
                stmt.setInt(5, metric.getTxcnt());
                stmt.setBigDecimal(6, metric.getTokenPriceUsd());
                stmt.setBigDecimal(7, metric.getBuyPressureUsd());
                stmt.setInt(8, metric.getBuyersCount());
                stmt.setInt(9, metric.getSellersCount());
                stmt.setBigDecimal(10, metric.getBuyVolumeUsd());
                stmt.setBigDecimal(11, metric.getSellVolumeUsd());
                stmt.setInt(12, metric.getMakersCount());
                stmt.setInt(13, metric.getBuyCount());
                stmt.setInt(14, metric.getSellCount());
            }
            
            stmt.executeUpdate();
        } catch (Exception e) {
            log.error("Error writing to database", e);
            throw e;
        }
    }

    public static PostgresSink<PairMetric> createPairMetricSink() {
        return new PostgresSink<>(
            "INSERT INTO twswap_pair_metric (pair_id, time_window, end_time, token0_reserve, token1_reserve, " +
            "reserve_usd, token0_volume_usd, token1_volume_usd, volume_usd, txcnt) " +
            "VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) " +
            "ON CONFLICT (pair_id, time_window, end_time) DO UPDATE SET " +
            "token0_reserve = EXCLUDED.token0_reserve, " +
            "token1_reserve = EXCLUDED.token1_reserve, " +
            "reserve_usd = EXCLUDED.reserve_usd, " +
            "token0_volume_usd = EXCLUDED.token0_volume_usd, " +
            "token1_volume_usd = EXCLUDED.token1_volume_usd, " +
            "volume_usd = EXCLUDED.volume_usd, " +
            "txcnt = EXCLUDED.txcnt"
        );
    }

    public static PostgresSink<TokenMetric> createTokenMetricSink() {
        return new PostgresSink<>(
            "INSERT INTO token_metric (token_id, time_window, end_time, volume_usd, txcnt, " +
            "token_price_usd, buy_pressure_usd, buyers_count, sellers_count, buy_volume_usd, sell_volume_usd, " +
            "makers_count, buy_count, sell_count) " +
            "VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) " +
            "ON CONFLICT (token_id, time_window, end_time) DO UPDATE SET " +
            "volume_usd = EXCLUDED.volume_usd, " +
            "txcnt = EXCLUDED.txcnt, " +
            "token_price_usd = EXCLUDED.token_price_usd, " +
            "buy_pressure_usd = EXCLUDED.buy_pressure_usd, " +
            "buyers_count = EXCLUDED.buyers_count, " +
            "sellers_count = EXCLUDED.sellers_count, " +
            "buy_volume_usd = EXCLUDED.buy_volume_usd, " +
            "sell_volume_usd = EXCLUDED.sell_volume_usd, " +
            "makers_count = EXCLUDED.makers_count, " +
            "buy_count = EXCLUDED.buy_count, " +
            "sell_count = EXCLUDED.sell_count"
        );
    }
} 