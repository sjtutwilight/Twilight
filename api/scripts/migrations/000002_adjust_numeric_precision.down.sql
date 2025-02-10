-- Revert numeric columns in token_metric table
ALTER TABLE token_metric
    ALTER COLUMN volume_usd TYPE NUMERIC(38,18),
    ALTER COLUMN token_price_usd TYPE NUMERIC(38,18),
    ALTER COLUMN buy_pressure_usd TYPE NUMERIC(38,18),
    ALTER COLUMN buy_volume_usd TYPE NUMERIC(38,18),
    ALTER COLUMN sell_volume_usd TYPE NUMERIC(38,18);

-- Revert numeric columns in token table
ALTER TABLE token
    ALTER COLUMN supply_usd TYPE NUMERIC(38,18),
    ALTER COLUMN liquidity_usd TYPE NUMERIC(38,18),
    ALTER COLUMN current_price TYPE NUMERIC(38,18),
    ALTER COLUMN price_change_1h TYPE NUMERIC(38,18),
    ALTER COLUMN volume_1h TYPE NUMERIC(38,18),
    ALTER COLUMN buy_pressure_1h TYPE NUMERIC(38,18);

-- Revert numeric columns in twswap_pair_metric table
ALTER TABLE twswap_pair_metric
    ALTER COLUMN token0_reserve TYPE NUMERIC(38,18),
    ALTER COLUMN token1_reserve TYPE NUMERIC(38,18),
    ALTER COLUMN reserve_usd TYPE NUMERIC(38,18),
    ALTER COLUMN token0_volume_usd TYPE NUMERIC(38,18),
    ALTER COLUMN token1_volume_usd TYPE NUMERIC(38,18),
    ALTER COLUMN volume_usd TYPE NUMERIC(38,18);

-- Revert numeric columns in twswap_factory table
ALTER TABLE twswap_factory
    ALTER COLUMN volume_usd TYPE NUMERIC(38,18),
    ALTER COLUMN liquidity_usd TYPE NUMERIC(38,18); 