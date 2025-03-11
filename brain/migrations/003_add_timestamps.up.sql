-- Add timestamp columns to token_metric
ALTER TABLE token_metric 
ADD COLUMN IF NOT EXISTS create_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN IF NOT EXISTS update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- Add timestamp columns to token_holder
ALTER TABLE token_holder 
ADD COLUMN IF NOT EXISTS create_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN IF NOT EXISTS update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- Add timestamp columns to account_asset_view
ALTER TABLE account_asset_view 
ADD COLUMN IF NOT EXISTS create_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN IF NOT EXISTS update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- Create function to update timestamp
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.update_time = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for token_metric
CREATE TRIGGER update_token_metric_timestamp
    BEFORE UPDATE ON token_metric
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();

-- Create triggers for token_holder
CREATE TRIGGER update_token_holder_timestamp
    BEFORE UPDATE ON token_holder
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();

-- Create triggers for account_asset_view
CREATE TRIGGER update_account_asset_view_timestamp
    BEFORE UPDATE ON account_asset_view
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp(); 