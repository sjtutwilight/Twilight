-- -- Insert USDC token
-- INSERT INTO token (chain_id, token_address, token_name, token_symbol, token_decimals, trust_score, create_time, update_time)
-- VALUES ('31337', '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48', 'USD Coin', 'USDC', 6, 1.0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
-- ON CONFLICT (chain_id, token_address) DO NOTHING;

-- -- Insert WETH token
-- INSERT INTO token (chain_id, token_address, token_name, token_symbol, token_decimals, trust_score, create_time, update_time)
-- VALUES ('31337', '0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2', 'Wrapped Ether', 'WETH', 18, 1.0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
-- ON CONFLICT (chain_id, token_address) DO NOTHING;

-- -- Insert USDT token
-- INSERT INTO token (chain_id, token_address, token_name, token_symbol, token_decimals, trust_score, create_time, update_time)
-- VALUES ('31337', '0xdAC17F958D2ee523a2206206994597C13D831ec7', 'Tether USD', 'USDT', 6, 1.0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
-- ON CONFLICT (chain_id, token_address) DO NOTHING;

-- -- Insert DAI token
-- INSERT INTO token (chain_id, token_address, token_name, token_symbol, token_decimals, trust_score, create_time, update_time)
-- VALUES ('31337', '0x6B175474E89094C44Da98b954EedeAC495271d0F', 'Dai Stablecoin', 'DAI', 18, 1.0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
-- ON CONFLICT (chain_id, token_address) DO NOTHING;

-- -- Insert initial factory data
-- INSERT INTO twswap_factory (chain_id, factory_address, time_window, end_time, pair_count, volume_usd, liquidity_usd, txcnt, update_time)
-- VALUES ('31337', '0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f', 'total', CURRENT_TIMESTAMP, 0, 0, 0, 0, CURRENT_TIMESTAMP)
-- ON CONFLICT (chain_id, factory_address, time_window, end_time) DO NOTHING; 