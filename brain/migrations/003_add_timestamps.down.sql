-- Drop triggers
DROP TRIGGER IF EXISTS update_token_metric_timestamp ON token_metric;
DROP TRIGGER IF EXISTS update_token_holder_timestamp ON token_holder;
DROP TRIGGER IF EXISTS update_account_asset_view_timestamp ON account_asset_view;

-- Drop function
DROP FUNCTION IF EXISTS update_timestamp();

-- Drop timestamp columns from token_metric
ALTER TABLE token_metric 
DROP COLUMN IF EXISTS create_time,
DROP COLUMN IF EXISTS update_time;

-- Drop timestamp columns from token_holder
ALTER TABLE token_holder 
DROP COLUMN IF EXISTS create_time,
DROP COLUMN IF EXISTS update_time;

-- Drop timestamp columns from account_asset_view
ALTER TABLE account_asset_view 
DROP COLUMN IF EXISTS create_time,
DROP COLUMN IF EXISTS update_time; 