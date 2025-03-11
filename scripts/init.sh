#!/bin/bash
set -e

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
until PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -c '\q'; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done

echo "PostgreSQL is up - executing schema creation"

# Create database if it doesn't exist
PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -tc "SELECT 1 FROM pg_database WHERE datname = 'twilight'" | grep -q 1 || PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -c "CREATE DATABASE twilight"

# Connect to the database and create schema
PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -d twilight << EOF

-- Drop tables if they exist (for clean initialization)
DROP TABLE IF EXISTS account_asset_view CASCADE;
DROP TABLE IF EXISTS account_transfer_history CASCADE;
DROP TABLE IF EXISTS token_holder CASCADE;
DROP TABLE IF EXISTS token_rolling_metric CASCADE;
DROP TABLE IF EXISTS token_recent_metric CASCADE;
DROP TABLE IF EXISTS token_metric CASCADE;
DROP TABLE IF EXISTS transfer_event CASCADE;
DROP TABLE IF EXISTS event CASCADE;
DROP TABLE IF EXISTS transaction CASCADE;
DROP TABLE IF EXISTS twswap_pair_metric CASCADE;
DROP TABLE IF EXISTS twswap_pair CASCADE;
DROP TABLE IF EXISTS twswap_factory CASCADE;
DROP TABLE IF EXISTS account_asset CASCADE;
DROP TABLE IF EXISTS token CASCADE;
DROP TABLE IF EXISTS account CASCADE;
DROP TABLE IF EXISTS action_executions CASCADE;
DROP TABLE IF EXISTS condition_executions CASCADE;
DROP TABLE IF EXISTS events CASCADE;
DROP TABLE IF EXISTS strategies CASCADE;

-- Create account table
CREATE TABLE account (
    id SERIAL PRIMARY KEY,
    chain_id INTEGER NOT NULL,
    chain_name VARCHAR(100) NOT NULL,
    address VARCHAR(128) NOT NULL,
    entity VARCHAR(255),
    smart_money_tag BOOLEAN NOT NULL DEFAULT FALSE,
    cex_tag BOOLEAN NOT NULL DEFAULT FALSE,
    big_whale_tag BOOLEAN NOT NULL DEFAULT FALSE,
    fresh_wallet_tag BOOLEAN NOT NULL DEFAULT FALSE,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (chain_id, address)
);

-- Create account_asset table
CREATE TABLE account_asset (
    account_id INTEGER NOT NULL,
    asset_type VARCHAR(50) NOT NULL,
    biz_id INTEGER NOT NULL,
    biz_name VARCHAR(255),
    value DECIMAL(24,4),
    extend_info JSON,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (account_id, asset_type, biz_id),
    CONSTRAINT fk_asset_account FOREIGN KEY (account_id) REFERENCES account(id)
);

CREATE INDEX idx_account_asset_asset_type ON account_asset(asset_type);

-- Create account_transfer_history table
CREATE TABLE account_transfer_history (
    account_id INTEGER NOT NULL,
    isBuy INTEGER NOT NULL,
    end_time TIMESTAMP NOT NULL,
    txcnt INTEGER,
    total_value_usd DECIMAL(24,4),
    PRIMARY KEY (account_id, end_time),
    CONSTRAINT fk_ath_account FOREIGN KEY (account_id) REFERENCES account(id)
);

-- Create transaction table
CREATE TABLE transaction (
    id BIGSERIAL PRIMARY KEY,
    chain_id VARCHAR(64),
    transaction_hash VARCHAR(128) NOT NULL,
    block_number BIGINT,
    block_timestamp TIMESTAMP,
    from_address VARCHAR(128),
    to_address VARCHAR(128),
    method_name VARCHAR(64),
    transaction_status INT CHECK (transaction_status IN (0, 1, 2)),
    gas_used BIGINT,
    input_data TEXT,
    create_time TIMESTAMP,
    UNIQUE (chain_id, transaction_hash)
);

CREATE INDEX idx_transaction_from ON transaction (from_address);
CREATE INDEX idx_transaction_to ON transaction (to_address);

-- Create event table
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

CREATE INDEX idx_event_tx_log ON event (transaction_id, log_index);

-- Create token table
CREATE TABLE token (
    id SERIAL PRIMARY KEY,
    chain_id INTEGER NOT NULL,
    chain_name VARCHAR(100),
    token_symbol VARCHAR(50) NOT NULL,
    token_catagory VARCHAR(50),
    token_decimals INTEGER,
    token_address VARCHAR(255) NOT NULL,
    issuer VARCHAR(255),
    create_time TIMESTAMP,
    update_time TIMESTAMP
);

CREATE INDEX idx_token_chain_id ON token(chain_id);
CREATE INDEX idx_token_token_symbol ON token(token_symbol);
CREATE INDEX idx_token_token_address ON token(token_address);

-- Create transfer_event table
CREATE TABLE transfer_event (
    id SERIAL PRIMARY KEY,
    token_id INTEGER NOT NULL,
    event_id INTEGER NOT NULL,
    from_address VARCHAR(255) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    block_timestamp TIMESTAMP NOT NULL,
    token_symbol VARCHAR(50),
    value_usd DOUBLE PRECISION,
    CONSTRAINT fk_te_event FOREIGN KEY (event_id) REFERENCES event(id),
    CONSTRAINT fk_te_token FOREIGN KEY (token_id) REFERENCES token(id)
);
CREATE INDEX idx_te_token_timestamp ON transfer_event(token_id, block_timestamp);
CREATE INDEX idx_te_from_address ON transfer_event(from_address);
CREATE INDEX idx_te_to_address ON transfer_event(to_address);

-- Create twswap_factory table
CREATE TABLE twswap_factory (
    id BIGSERIAL PRIMARY KEY,
    chain_id VARCHAR(64),
    factory_address VARCHAR(128) NOT NULL,
    time_window VARCHAR(16),
    end_time TIMESTAMP,
    pair_count INT,
    volume_usd DECIMAL(24,4),
    liquidity_usd DECIMAL(24,4),
    txcnt INT,
    update_time TIMESTAMP,
    UNIQUE (chain_id, factory_address, time_window, end_time)
);

-- Create token_recent_metric table
CREATE TABLE token_recent_metric (
    token_id INTEGER NOT NULL,
    time_window VARCHAR(50) NOT NULL,
    end_time TIMESTAMP NOT NULL,
    tag VARCHAR(50) NOT NULL,
    txcnt INTEGER,
    buy_count INTEGER,
    sell_count INTEGER,
    volume_usd DECIMAL(24,4),
    buy_volume_usd DECIMAL(24,4),
    sell_volume_usd DECIMAL(24,4),
    buy_pressure_usd DECIMAL(24,4),
    token_price_usd DECIMAL(24,4),
    PRIMARY KEY (token_id, time_window, end_time, tag),
    CONSTRAINT fk_trm_recent_token FOREIGN KEY (token_id) REFERENCES token(id)
);

CREATE INDEX idx_token_recent_end_time ON token_recent_metric(end_time);

CREATE TABLE token_metric (
    token_id INTEGER PRIMARY KEY,
    token_price DOUBLE PRECISION,
    token_age INTEGER,
    liquidity_usd DOUBLE PRECISION,
    security_score INTEGER,
    fdv DOUBLE PRECISION,
    mcap DOUBLE PRECISION,
    create_time       TIMESTAMP,
  update_time       TIMESTAMP,
    CONSTRAINT fk_tmview_token FOREIGN KEY (token_id) REFERENCES token(id)
);
CREATE INDEX idx_tmv_mcap ON token_metric_view(mcap);

-- Create token_rolling_metric table
CREATE TABLE token_rolling_metric (
    token_id INTEGER NOT NULL,
    end_time TIMESTAMP NOT NULL,
    token_price_usd DECIMAL(24,4),
    mcap DECIMAL(24,4),
    PRIMARY KEY (token_id, end_time),
    CONSTRAINT fk_trm_token FOREIGN KEY (token_id) REFERENCES token(id)
);

CREATE INDEX idx_trm_end_time ON token_rolling_metric(end_time);

CREATE TABLE token_holder(
    token_id INTEGER NOT NULL,
    account_id INTEGER NOT NULL,
	  token_address VARCHAR(255) NOT NULL,
    account_address VARCHAR(255) NOT NULL,
    entity VARCHAR(255),
    value_usd DOUBLE PRECISION,
    ownership REAL,           -- 持仓比例（浮点数）
      create_time       TIMESTAMP,
  update_time       TIMESTAMP,
    PRIMARY KEY (token_id, account_id),
    CONSTRAINT fk_th_token FOREIGN KEY (token_id) REFERENCES token(id),
    CONSTRAINT fk_th_account FOREIGN KEY (account_id) REFERENCES account(id)
);

-- Create twswap_pair table
CREATE TABLE twswap_pair (
    id BIGSERIAL PRIMARY KEY,
    chain_id VARCHAR(64),
    pair_address VARCHAR(128) NOT NULL,
    pair_name VARCHAR(64),
    token0_id BIGINT REFERENCES token(id),
    token1_id BIGINT REFERENCES token(id),
    fee_tier VARCHAR(16) DEFAULT '0.3%',
    created_at_timestamp TIMESTAMP,
    created_at_block_number BIGINT,
    UNIQUE (chain_id, pair_address)
);

-- Create twswap_pair_metric table
CREATE TABLE twswap_pair_metric (
    id BIGSERIAL PRIMARY KEY,
    pair_id BIGINT NOT NULL REFERENCES twswap_pair(id) ON DELETE CASCADE,
    time_window VARCHAR(16),
    end_time TIMESTAMP,
    token0_reserve DECIMAL(24,4),
    token1_reserve DECIMAL(24,4),
    reserve_usd DECIMAL(24,4),
    token0_volume_usd DECIMAL(24,4),
    token1_volume_usd DECIMAL(24,4),
    volume_usd DECIMAL(24,4),
    txcnt INT,
    UNIQUE (pair_id, time_window, end_time)
);

-- Create strategies table
CREATE TABLE strategies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    config JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN DEFAULT true
);

-- Create events table
CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    processed BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create condition_executions table
CREATE TABLE condition_executions (
    id SERIAL PRIMARY KEY,
    strategy_id INTEGER REFERENCES strategies(id),
    event_id INTEGER REFERENCES events(id),
    condition_type VARCHAR(50) NOT NULL,
    result BOOLEAN NOT NULL,
    execution_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create action_executions table
CREATE TABLE action_executions (
    id SERIAL PRIMARY KEY,
    strategy_id INTEGER REFERENCES strategies(id),
    event_id INTEGER REFERENCES events(id),
    action_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create view tables (for CQRS)
CREATE TABLE account_asset_view (
    account_id INTEGER NOT NULL,
    asset_type VARCHAR(50) NOT NULL,
    biz_id INTEGER NOT NULL,
    biz_name VARCHAR(255),
    value DECIMAL(24,4),
    value_usd DECIMAL(24,4),
    asset_price DECIMAL(24,4),
      create_time       TIMESTAMP,
  update_time       TIMESTAMP,
    PRIMARY KEY (account_id, asset_type, biz_id),
    CONSTRAINT fk_aav_account FOREIGN KEY (account_id) REFERENCES account(id)
);

CREATE INDEX idx_aav_biz_name ON account_asset_view(account_id, biz_name);



GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO twilight;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO twilight;

EOF

echo "Database schema created successfully!" 