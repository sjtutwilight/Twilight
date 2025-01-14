```sql
CREATE TABLE transaction (
  id BIGSERIAL PRIMARY KEY,
  chain_id VARCHAR(64),
  transaction_hash VARCHAR(128) NOT NULL,
  block_number BIGINT,
  block_timestamp TIMESTAMP,
  from_address VARCHAR(128),
  to_address VARCHAR(128),
  method_name VARCHAR(64),
  transaction_status INT,
  gas_used BIGINT,
  input_data TEXT,
  create_time TIMESTAMP,

  -- 若需要强唯一，可加:
  UNIQUE (chain_id, transaction_hash)
);
CREATE TABLE event (
  id BIGSERIAL PRIMARY KEY,
  transaction_id BIGINT REFERENCES transaction(id) ON DELETE CASCADE,
  chain_id VARCHAR(64),
  event_name VARCHAR(64),
  contract_address VARCHAR(128),
  log_index INT,
  event_data TEXT,
  create_time TIMESTAMP,
  block_number BIGINT
);
CREATE TABLE token (
  id BIGSERIAL PRIMARY KEY,
  chain_id VARCHAR(64),
  token_address VARCHAR(128) NOT NULL,
  token_symbol VARCHAR(32),
  token_name VARCHAR(128),
  token_decimals INT,
  trust_score DECIMAL(5,2),
  create_time TIMESTAMP,
  update_time TIMESTAMP,

  UNIQUE (chain_id, token_address)
);
CREATE TABLE twswap_factory (
  id BIGSERIAL PRIMARY KEY,
  chain_id VARCHAR(64),
  factory_address VARCHAR(128) NOT NULL,
  pair_count INT,
  total_volume_usd DECIMAL(38,18),
  total_liquidity_usd DECIMAL(38,18),
  transaction_count INT,
  update_time TIMESTAMP,

  UNIQUE (chain_id, factory_address)
);
CREATE TABLE token_metric (
  id BIGSERIAL PRIMARY KEY,
  token_id BIGINT NOT NULL REFERENCES token(id) ON DELETE CASCADE,
  metric_type VARCHAR(16) NOT NULL,    -- e.g. '10s','1min','5min','30min','1h'
  start_time TIMESTAMP NOT NULL,       -- 窗口聚合开始时间

  total_supply        DECIMAL(38,18),
  trade_volume_usd    DECIMAL(38,18),
  transaction_count   INT,
  total_liquidity     DECIMAL(38,18),
  token_price_usd     DECIMAL(38,18),
  update_time         TIMESTAMP
);
CREATE TABLE pair (
  id BIGSERIAL PRIMARY KEY,
  chain_id VARCHAR(64),
  pair_address VARCHAR(128) NOT NULL,
  token0_id BIGINT REFERENCES token(id),
  token1_id BIGINT REFERENCES token(id),
  token0_reserve DECIMAL(38,18),
  token1_reserve DECIMAL(38,18),
  total_supply DECIMAL(38,18),
  volume_token0 DECIMAL(38,18),
  volume_token1 DECIMAL(38,18),
  created_at_timestamp TIMESTAMP,
  created_at_block_number BIGINT,
  update_time TIMESTAMP,

  UNIQUE (chain_id, pair_address)
);
CREATE TABLE pair_metric (
  id BIGSERIAL PRIMARY KEY,
  pair_id BIGINT NOT NULL REFERENCES pair(id) ON DELETE CASCADE,
  metric_type VARCHAR(16) NOT NULL,          -- e.g. '10s','1min','5min','30min','1h'
  start_timestamp TIMESTAMP NOT NULL,        -- 窗口开始时间
  token0_address VARCHAR(128),
  token1_address VARCHAR(128),

  reserve0 DECIMAL(38,18),
  reserve1 DECIMAL(38,18),
  total_supply DECIMAL(38,18),
  reserve_usd DECIMAL(38,18),
  token0_volume DECIMAL(38,18),
  token1_volume DECIMAL(38,18),
  volume_usd DECIMAL(38,18),
  transaction_count INT,
  update_time TIMESTAMP
);


```