// scripts/deploy.js
const fs = require("fs");
const path = require("path");
const hre = require("hardhat");
const { ethers } = hre;
const { Client } = require('pg');
const deployment = require('../../deployment.json');
const Redis = require('ioredis');

async function updateInitCodeHash() {
  // Get TWSwapPair bytecode
  const TWSwapPair = await ethers.getContractFactory("TWSwapPair");
  const bytecode = TWSwapPair.bytecode;
  
  // Calculate init code hash
  const COMPUTED_INIT_CODE_HASH = ethers.keccak256(bytecode);
  console.log("Computed INIT_CODE_HASH:", COMPUTED_INIT_CODE_HASH);

  // Read TWSwapLibrary.sol
  const libraryPath = path.join(__dirname, "../contracts/TWSwap/libraries/TWSwapLibrary.sol");
  let libraryContent = fs.readFileSync(libraryPath, 'utf8');

  // Replace the old hash with the new one
  const oldHashPattern = /uint256 initCodeHash = 0x[a-fA-F0-9]+;/;
  const newHashValue = `uint256 initCodeHash = ${COMPUTED_INIT_CODE_HASH};`;
  libraryContent = libraryContent.replace(oldHashPattern, newHashValue);

  // Write back to file
  fs.writeFileSync(libraryPath, libraryContent);
  console.log("Updated INIT_CODE_HASH in TWSwapLibrary.sol");

  return COMPUTED_INIT_CODE_HASH;
}

// Add new function for periodic token updates
async function startTokenMetricsUpdater(deploymentInfo, redis, pgClient) {
  const updateTokenMetrics = async () => {
    try {
      // Get all tokens and their contracts
      for (const token of deploymentInfo.tokens) {
        // Get token contract
        const tokenContract = await ethers.getContractAt("MyERC20", token.address);
        
        // Calculate random hype score (0-100)
        const hypeScore = Math.floor(Math.random() * 101);
        
        // Get token price from Redis
        const tokenPrice = await redis.get(`token_price:${token.address.toLowerCase()}`);
        const priceInUsd = tokenPrice ? parseFloat(tokenPrice) : 0;
        
        // Get total supply and calculate supply_usd
        const totalSupply = await tokenContract.totalSupply();
        const supplyUsd = (Number(ethers.formatUnits(totalSupply, token.decimals)) * priceInUsd).toString();
        
        // Calculate liquidity_usd by summing all pairs' reserves
        let liquidityUsd = 0;
        for (const pair of deploymentInfo.pairs) {
          if (pair.token0.toLowerCase() === token.address.toLowerCase() || 
              pair.token1.toLowerCase() === token.address.toLowerCase()) {
            const pairContract = await ethers.getContractAt("TWSwapPair", pair.address);
            const [reserve0, reserve1] = await pairContract.getReserves();
            
            // Determine which reserve belongs to our token
            const isToken0 = pair.token0.toLowerCase() === token.address.toLowerCase();
            const tokenReserve = isToken0 ? reserve0 : reserve1;
            
            // Add to liquidity (reserve * price)
            liquidityUsd += Number(ethers.formatUnits(tokenReserve, token.decimals)) * priceInUsd;
          }
        }
        
        // Update token in database
        await pgClient.query(
          `UPDATE token 
           SET hype_score = $1, 
               supply_usd = $2, 
               liquidity_usd = $3,
               update_time = NOW()
           WHERE token_address = $4`,
          [hypeScore, supplyUsd, liquidityUsd.toString(), token.address]
        );
        
        console.log(`Updated metrics for ${token.symbol}:`, {
          hypeScore,
          supplyUsd,
          liquidityUsd: liquidityUsd.toString()
        });
      }
    } catch (error) {
      console.error('Error updating token metrics:', error);
    }
  };

  // Run initial update
  await updateTokenMetrics();
  
  // Schedule updates every minute
  setInterval(updateTokenMetrics, 60000);
}

async function main() {
  try {
    // First update the INIT_CODE_HASH
    const initCodeHash = await updateInitCodeHash();
    console.log("Using INIT_CODE_HASH:", initCodeHash);

    console.log("\nStarting initialization...");

    // Get signers (accounts)
    const accounts = await ethers.getSigners();
    console.log(`Using ${accounts.length} accounts for testing`);

    // Deploy CloneFactory
    console.log("Deploying CloneFactory...");
    const CloneFactory = await ethers.getContractFactory("CloneFactory");
    const cloneFactory = await CloneFactory.deploy();
    await cloneFactory.waitForDeployment();
    console.log("CloneFactory deployed to:", await cloneFactory.getAddress());

    // Deploy MyERC20 implementation
    console.log("Deploying MyERC20 implementation...");
    const MyERC20 = await ethers.getContractFactory("MyERC20");
    const myERC20Implementation = await MyERC20.deploy();
    await myERC20Implementation.waitForDeployment();
    console.log("MyERC20 implementation deployed to:", await myERC20Implementation.getAddress());

    // Deploy TWSwap contracts
    console.log("Deploying TWSwap contracts...");
    const TWSwapFactory = await ethers.getContractFactory("TWSwapFactory");
    const twSwapFactory = await TWSwapFactory.deploy();
    await twSwapFactory.waitForDeployment();
    console.log("TWSwapFactory deployed to:", await twSwapFactory.getAddress());

    const TWSwapRouter = await ethers.getContractFactory("TWSwapRouter");
    const twSwapRouter = await TWSwapRouter.deploy(await twSwapFactory.getAddress());
    await twSwapRouter.waitForDeployment();
    console.log("TWSwapRouter deployed to:", await twSwapRouter.getAddress());

    // Deploy tokens using CloneFactory
    console.log("Deploying tokens via CloneFactory...");
    const tokenConfigs = [
      { name: "USD Coin", symbol: "USDC", decimals: 18 },
      { name: "Wrapped Ether", symbol: "WETH", decimals: 18 },
     
      { name: "Dai Stablecoin", symbol: "DAI", decimals: 18 },
      { name: "Twilight Token", symbol: "TWI", decimals: 18 },
      { name: "Wrapped Bitcoin", symbol: "WBTC", decimals: 18}
    ];

    const tokens = [];
    for (const config of tokenConfigs) {
      console.log(`\nDeploying ${config.symbol}...`);
      
      // Clone the implementation
      const tx = await cloneFactory.clone(await myERC20Implementation.getAddress());
      const receipt = await tx.wait();
      
      // Get the cloned token address from the event
      const event = receipt.logs.find(
        log => log.fragment && log.fragment.name === 'CloneCreated'
      );
      if (!event) {
        throw new Error(`Failed to get CloneCreated event for ${config.symbol}`);
      }
      const tokenAddress = event.args.instance;
      
      // Initialize the cloned token
      const token = await ethers.getContractAt("MyERC20", tokenAddress);
      await token.initialize(config.name, config.symbol);
      console.log(`${config.symbol} deployed and initialized at:`, tokenAddress);

      // Mint tokens to all accounts
      const mintAmount = ethers.parseUnits("1000000", config.decimals);
      for (const account of accounts) {
        await token.mint(account.address, mintAmount);
        console.log(`Minted ${ethers.formatUnits(mintAmount, config.decimals)} ${config.symbol} to ${account.address}`);
      }

      tokens.push(token);
    }

    // Initialize pairs array with correct token order and amounts
    const pairs = [
      {
        token0: 'USDC',
        token1: 'WETH',
        amount0: '3000',    // 3000 USDC
        amount1: '1'       // 1 WETH
      },
      {
        token0: 'WETH',
        token1: 'DAI',
        amount0: '1',      // 1 WETH
        amount1: '3000'    // 3000 DAI
      },
      {
        token0: 'USDC',
        token1: 'DAI',
        amount0: '1000',   // 1000 USDC
        amount1: '1000'    // 1000 DAI
      },
      {
        token0: 'TWI',
        token1: 'WETH',
       
        amount0: '60',      // 60 TWI
        amount1: '1'       // 1 WETH
      },
      {
        token0: 'WBTC',
        token1: 'WETH',
       
        amount0: '1',     // 1 WBTC
        amount1: '40'       // 40 WETH
      },
      {
        token0: 'USDC',
        token1: 'TWI',
        amount0: '5000',  // 5000 USDC
        amount1: '100'     // 100 TWI
      },
      {
        token0: 'USDC',
        token1: 'WBTC',
        amount0: '120000', // 120000 USDC
        amount1: '1'       // 1 WBTC
      }
    ];

    // Map token symbols to token contracts
    const tokenMap = {
      'USDC': tokens[0],
      'WETH': tokens[1],
      'DAI': tokens[2],
      'TWI': tokens[3],
      'WBTC': tokens[4]
    };

    // Add liquidity for each pair
    for (const pairConfig of pairs) {
      try {
        // 根据 pair 配置从 tokenMap 中获取对应 token 合约
        const tokenAContract = tokenMap[pairConfig.token0];
        const tokenBContract = tokenMap[pairConfig.token1];
        // 获取 token 地址和小数位
        const tokenAAddress = await tokenAContract.getAddress();
        const tokenASymbol = await tokenAContract.symbol();
        const tokenBAddress = await tokenBContract.getAddress();
        const tokenBSymbol = await tokenBContract.symbol();
        const tokenADecimals = await tokenAContract.decimals();
        const tokenBDecimals = await tokenBContract.decimals();
      
        // 将配置的数量转换为 BigNumber（注意数量配置字符串必须与各自的小数位匹配）
        const amountADesired = ethers.parseUnits(pairConfig.amount0, tokenADecimals);
        const amountBDesired = ethers.parseUnits(pairConfig.amount1, tokenBDecimals);
      
        // Get router address
        const routerAddress = await twSwapRouter.getAddress();
        console.log(`Router address: ${routerAddress}`);
      
        // Approve router 转账流动性所需的 token（注意：需要分别对两个 token 进行授权）
        console.log(`Approving tokens for router ${routerAddress}...`);
        const approveATx = await tokenAContract.approve(routerAddress, ethers.MaxUint256);
        await approveATx.wait();
        console.log(`Approved ${pairConfig.token0}`);
        
        const approveBTx = await tokenBContract.approve(routerAddress, ethers.MaxUint256);
        await approveBTx.wait();
        console.log(`Approved ${pairConfig.token1}`);
      
        // 调用 Router 的 addLiquidity，注意 deadline 单位为秒
        const deadline = Math.floor(Date.now() / 1000) + 1000;
        console.log('Adding liquidity...');
        console.log(`tokenA: ${tokenASymbol},address: ${tokenAAddress},amount: ${ethers.formatUnits(amountADesired, tokenADecimals)},tokenB: ${tokenBSymbol},address: ${tokenBAddress},amount: ${ethers.formatUnits(amountBDesired, tokenBDecimals)}`);
        const addLiquidityTx = await twSwapRouter.addLiquidity(
          tokenAAddress,   // tokenA（已排序）
          tokenBAddress,   // tokenB（已排序）
          amountADesired,  // 数量与 tokenA 对应
          amountBDesired,  // 数量与 tokenB 对应
          0,              // amountAMin：可设为 0，或根据报价设定
          0,              // amountBMin：可设为 0，或根据报价设定
          accounts[0].address, // 流动性接收者
          deadline        // 截止时间
        );
        console.log("等待流动性注入交易确认...");
        const receipt = await addLiquidityTx.wait();
        console.log(`流动性注入交易成功：${receipt.hash}`);
      
        // 获取 pair 地址并验证 reserves
        const pairAddress = await twSwapFactory.getPair(tokenAAddress, tokenBAddress);
        const pairContract = await ethers.getContractAt("TWSwapPair", pairAddress);
        const reserves = await pairContract.getReserves();
        console.log(`Pair ${pairConfig.token0}-${pairConfig.token1} 地址：${pairAddress}`);
        console.log(`注入后 reserves：reserve0 = ${reserves.reserve0}, reserve1 = ${reserves.reserve1}`);
      } catch (error) {
        console.error(`Error adding liquidity for pair ${pairConfig.token0}-${pairConfig.token1}:`, error);
        throw error;
      }
    }

    console.log("\nInitialization complete!");

    // Save deployment info
    const deploymentInfo = {
      factory: await twSwapFactory.getAddress(),
      router: await twSwapRouter.getAddress(),
      initCodeHash: initCodeHash,
      tokens: await Promise.all(tokens.map(async t => ({
        address: await t.getAddress(),
        symbol: await t.symbol(),
        decimals: await t.decimals(),
        id: (tokens.indexOf(t) + 1).toString()
      }))),
      pairs: await Promise.all(pairs.map(async pair => {
        const token0 = tokenMap[pair.token0];
        const token1 = tokenMap[pair.token1];
        const token0Address = await token0.getAddress();
        const token1Address = await token1.getAddress();
        const pairAddress = await twSwapFactory.getPair(token0Address, token1Address);
        return {
          token0: token0Address,
          token1: token1Address,
          address: pairAddress
        };
      }))
    };

    // Convert BigInt values to strings for JSON serialization
    const deploymentInfoSerializable = {
      ...deploymentInfo,
      factory: deploymentInfo.factory.toString(),
      router: deploymentInfo.router.toString(),
      initCodeHash: deploymentInfo.initCodeHash.toString(),
      tokens: deploymentInfo.tokens.map(token => ({
        ...token,
        address: token.address.toString(),
        decimals: token.decimals.toString()
      })),
      pairs: deploymentInfo.pairs.map(pair => ({
        token0: pair.token0.toString(),
        token1: pair.token1.toString(),
        address: pair.address.toString()
      }))
    };

    // Save deployment info to file in project root
    const deploymentPath = path.join(__dirname, "../../deployment.json");
    fs.writeFileSync(deploymentPath, JSON.stringify(deploymentInfoSerializable, null, 2));
    console.log(`\nDeployment info saved to ${deploymentPath}`);

    // Initialize Redis and PostgreSQL connections
    console.log("\nInitializing connections...");
    const redis = new Redis({
      host: 'localhost',
      port: 6379
    });

    const pgClient = new Client({
      host: 'localhost',
      port: 5432,
      database: 'twilight',
      user: 'twilight',
      password: 'twilight123'
    });
    await pgClient.connect();

    try {
      // Initialize Redis with token prices
      console.log("\nInitializing token prices in Redis...");
      await initializeRedis(redis, tokens, tokenMap, pairs, twSwapFactory);

      // Initialize database tables
      console.log("\nInitializing database tables...");
      await initializeDatabase(pgClient, deploymentInfo);

      // Start periodic token metrics updates
      console.log("\nStarting periodic token metrics updates...");
      await startTokenMetricsUpdater(deploymentInfo, redis, pgClient);
      
      console.log("Token metrics updater started successfully");
      
      // Keep the process running
      process.stdin.resume();
      
      // Handle graceful shutdown
      process.on('SIGINT', async () => {
        console.log('\nReceived SIGINT. Shutting down gracefully...');
        await redis.quit();
        await pgClient.end();
        process.exit(0);
      });

      return deploymentInfo;
    } catch (error) {
      console.error('Error during initialization:', error);
      await redis.quit();
      await pgClient.end();
      process.exit(1);
    }
  } catch (error) {
    console.error('Error:', error);
    process.exit(1);
  }
}

// Helper function to initialize Redis
async function initializeRedis(redis, tokens, tokenMap, pairs, twSwapFactory) {
  try {
    // Clear existing token prices
    console.log("Clearing existing token prices from Redis...");
    const keys = await redis.keys('token_price:*');
    if (keys.length > 0) {
      await redis.del(...keys);
    }

    // Get USDC token contract and address
    const usdcAddress = await tokens[0].getAddress();
    
    // For each pair with USDC, calculate and store token price
    for (const pair of pairs) {
      if (!pair.token0.includes('USDC') && !pair.token1.includes('USDC')) {
        continue;
      }

      const token0 = tokenMap[pair.token0];
      const token1 = tokenMap[pair.token1];
      const token0Address = await token0.getAddress();
      const token1Address = await token1.getAddress();
      
      const pairAddress = await twSwapFactory.getPair(token0Address, token1Address);
      const pairContract = await ethers.getContractAt("TWSwapPair", pairAddress);
      const [reserve0, reserve1] = await pairContract.getReserves();
      
      // If token0 is USDC, calculate token1's price
      if (pair.token0 === 'USDC') {
        const price = Number(pair.amount0) / Number(pair.amount1);
        await redis.set(`token_price:${token1Address.toLowerCase()}`, price.toString());
        console.log(`Set price for ${pair.token1}: ${price} USD`);
      }
      // If token1 is USDC, calculate token0's price
      else if (pair.token1 === 'USDC') {
        const price = Number(pair.amount1) / Number(pair.amount0);
        await redis.set(`token_price:${token0Address.toLowerCase()}`, price.toString());
        console.log(`Set price for ${pair.token0}: ${price} USD`);
      }
    }

    // Set USDC price to 1
    await redis.set(`token_price:${usdcAddress.toLowerCase()}`, "1");
    console.log("Set USDC price: 1 USD");
    console.log("Successfully initialized token prices in Redis");
  } catch (error) {
    console.error("Error initializing Redis:", error);
    throw error;
  }
}

// Helper function to initialize database
async function initializeDatabase(pgClient, deploymentInfo) {
  try {
    // 清空表数据
    await pgClient.query('TRUNCATE TABLE twswap_pair RESTART IDENTITY CASCADE');
    await pgClient.query('TRUNCATE TABLE token RESTART IDENTITY CASCADE');

    // 初始化 token 表
    const tokens = deploymentInfo.tokens;
    for (const token of tokens) {
      const result = await pgClient.query(
        `INSERT INTO token (
          chain_id, chain_name, token_address, token_symbol, 
          token_name, token_decimals, create_time, update_time
        ) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING id`,
        ['31337', 'ethereum', token.address, token.symbol, token.symbol, token.decimals]
      );
      token.id = result.rows[0].id;
    }

    // 初始化 pair 表
    const pairs = deploymentInfo.pairs;
    for (const pair of pairs) {
      const token0 = tokens.find(t => t.address.toLowerCase() === pair.token0.toLowerCase());
      const token1 = tokens.find(t => t.address.toLowerCase() === pair.token1.toLowerCase());

      if (!token0 || !token1) {
        console.warn(`Token not found for pair: ${pair.address}`);
        continue;
      }

      await pgClient.query(
        `INSERT INTO twswap_pair (
          chain_id, pair_address, token0_id, token1_id, 
          fee_tier, created_at_timestamp, created_at_block_number
        ) VALUES ($1, $2, $3, $4, $5, NOW(), $6)`,
        ['31337', pair.address, token0.id, token1.id, '0.3%', 0]
      );
    }

    console.log('Successfully initialized token and pair metadata');
  } catch (error) {
    console.error('Error initializing database:', error);
    throw error;
  }
}

main()
  .then((addresses) => {
    console.log("Deployment and initialization complete");
  })
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });