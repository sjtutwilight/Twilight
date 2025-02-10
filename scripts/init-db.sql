CREATE TABLE  if not exists account (
  id          BIGSERIAL PRIMARY KEY,
  chain_id    VARCHAR(64) NOT NULL,
  address     VARCHAR(128) NOT NULL ,
  balance     DECIMAL(30,10) DEFAULT 0, --ETH
  public_key  VARCHAR(256),
  create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(chain_id,address)
);
CREATE TABLE  if not exists account_asset (
  id         BIGSERIAL PRIMARY KEY,
  account_id BIGINT NOT NULL REFERENCES account(id) ON DELETE CASCADE,
  asset_type VARCHAR(32) NOT NULL,  -- 'token_holding' 或 'lp'
  bizId      VARCHAR(128) NOT NULL, -- if token_holding ,为token_id else if lp 为pair_id
  balance    DECIMAL(30,10) DEFAULT 0, --token balance
  extension_info       JSONB,                 -- 扩展信息,目前存lp注入的平均价格“average_price”
  create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(account_id, asset_type, bizId)
);
CREATE TABLE if not exists transaction (
  id                BIGSERIAL       PRIMARY KEY,
  chain_id          VARCHAR(64),
  transaction_hash  VARCHAR(128)    NOT NULL,
  block_number      BIGINT,
  block_timestamp   TIMESTAMP,
  from_address      VARCHAR(128),
  to_address        VARCHAR(128),
  method_name       VARCHAR(64),
  transaction_status INT,
  gas_used          BIGINT,
  input_data        TEXT,
  create_time       TIMESTAMP,

  -- 唯一性约束
  UNIQUE (chain_id, transaction_hash)
);
CREATE TABLE if not exists event (
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
CREATE TABLE if not exists token (
  id                BIGSERIAL     PRIMARY KEY,
  chain_id          VARCHAR(64),
  chain_name        VARCHAR(128),
  token_address     VARCHAR(128) NOT NULL,
  token_symbol      VARCHAR(32),
  token_name        VARCHAR(128),
  token_decimals    INT,
  hype_score        INT,                 -- 随机或自定义算法
  supply_usd        DECIMAL(24,4),      -- FDV
  liquidity_usd     DECIMAL(24,4),      -- 总流动性估算
  create_time       TIMESTAMP,
  update_time       TIMESTAMP,
  UNIQUE (chain_id, token_address)
);
CREATE TABLE if not exists twswap_factory (
  id                   BIGSERIAL     PRIMARY KEY,
  chain_id             VARCHAR(64),
  factory_address      VARCHAR(128)  NOT NULL,
  time_window          VARCHAR(16),  -- 如 '30s', '5min', '1h', 'total'
  end_time             TIMESTAMP,    -- 窗口结束时间
  pair_count           INT,
  volume_usd           DECIMAL(24,4),
  liquidity_usd        DECIMAL(24,4),
  txcnt                INT,
  -- 若需记录最后更新时刻
  update_time          TIMESTAMP,

  UNIQUE (chain_id, factory_address, time_window, end_time)
);
CREATE TABLE if not exists token_metric (
  id                  BIGSERIAL PRIMARY KEY,
  token_id            BIGINT NOT NULL REFERENCES token(id) ON DELETE CASCADE,
  time_window         VARCHAR(16),   -- '20s','1min','5min','30min','1h'
  end_time            TIMESTAMP,
  volume_usd          DECIMAL(24,4),  -- 指定窗口内交易量
  txcnt               INT,             -- 交易笔数
  token_price_usd     DECIMAL(24,4),  -- token价格
  buy_pressure_usd    DECIMAL(24,4),  -- (buy_volume - sell_volume)
  buyers_count        INT,
  sellers_count       INT,
  buy_volume_usd      DECIMAL(24,4),
  sell_volume_usd     DECIMAL(24,4),
  update_time         TIMESTAMP,
  UNIQUE (token_id, time_window, end_time)
);
CREATE TABLE if not exists twswap_token_metric (
  id              BIGSERIAL     PRIMARY KEY,
  token_id        BIGINT        NOT NULL REFERENCES token(id) ON DELETE CASCADE,
  time_window     VARCHAR(16),  -- '30s','5min','30min','1h','total'
  end_time        TIMESTAMP,    -- 窗口结束时间

  volume_usd      DECIMAL(24,4),
  txcnt           INT,
  liquidity_usd   DECIMAL(24,4),
  token_price_usd DECIMAL(24,4),

  update_time     TIMESTAMP,

  UNIQUE (token_id, time_window, end_time)
);
CREATE TABLE if not exists twswap_pair (
  id                     BIGSERIAL     PRIMARY KEY,
  chain_id               VARCHAR(64),
  pair_address           VARCHAR(128)  NOT NULL,
  token0_id              BIGINT        REFERENCES token(id),
  token1_id              BIGINT        REFERENCES token(id),
  fee_tier               VARCHAR(16)   DEFAULT '0.3%',  -- 目前默认 0.3%
  created_at_timestamp   TIMESTAMP,
  created_at_block_number BIGINT,

  UNIQUE (chain_id, pair_address, fee_tier)
);
CREATE TABLE if not exists twswap_pair_metric (
  id                  BIGSERIAL     PRIMARY KEY,
  pair_id             BIGINT        NOT NULL REFERENCES twswap_pair(id) ON DELETE CASCADE,
  time_window         VARCHAR(16),  -- '20s','1min','5min','30min'
  end_time            TIMESTAMP,    -- 窗口截止时间

  token0_reserve      DECIMAL(24,4),
  token1_reserve      DECIMAL(24,4),
  reserve_usd         DECIMAL(24,4),
  token0_volume_usd   DECIMAL(24,4),
  token1_volume_usd   DECIMAL(24,4),
  volume_usd          DECIMAL(24,4),
  txcnt               INT,

  UNIQUE (pair_id, time_window, end_time)
);

ALTER TABLE token_metric ADD COLUMN makers_count INT;
ALTER TABLE token_metric ADD COLUMN buy_count INT;
ALTER TABLE token_metric ADD COLUMN sell_count INT;
CREATE INDEX idx_token_metric_time_window ON token_metric (token_id, time_window, end_time);
CREATE INDEX idx_token_metric_latest ON token_metric (token_id, time_window, end_time DESC);