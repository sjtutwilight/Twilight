// scripts/deploy.js
const fs = require("fs");
const path = require("path");
const hre = require("hardhat");
const { ethers } = hre;

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

async function main() {
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
    { name: "Wrapped Ether", symbol: "WETH", decimals: 18 },
    { name: "USD Coin", symbol: "USDC", decimals: 18 },
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
  // Get balances of all accounts in all tokens
  // Create initial liquidity pairs
  console.log("\nCreating initial liquidity pairs...");
  const pairs = [
    { token0: 0, token1: 1, amount0: "100", amount1: "100000" }, // WETH-USDC
    { token0: 0, token1: 2, amount0: "100", amount1: "100000" }, // WETH-DAI
    { token0: 1, token1: 2, amount0: "100000", amount1: "100000" }, // USDC-DAI
    { token0: 0, token1: 3, amount0: "100", amount1: "50000" }, // WETH-TWI
    { token0: 0, token1: 4, amount0: "100", amount1: "5" },  // WETH-WBTC
    { token0: 1, token1: 3, amount0: "100000", amount1: "50000" }, // USDC-TWI
    { token0: 1, token1: 4, amount0: "100000", amount1: "5" }, // USDC-WBTC
  ];
  
 
  for (const pair of pairs) {
    const token0 = tokens[pair.token0];
    const token1 = tokens[pair.token1];
    const token0Decimals = await token0.decimals();
    const token1Decimals = await token1.decimals();
    
    const amount0 = ethers.parseUnits(pair.amount0, token0Decimals);
    const amount1 = ethers.parseUnits(pair.amount1, token1Decimals);

    console.log(`\nSetting up ${await token0.symbol()}-${await token1.symbol()} pair...`);
    
    // Get token addresses
    const token0Address = await token0.getAddress();
    const token1Address = await token1.getAddress();
    const routerAddress = await twSwapRouter.getAddress();

    // Check balances before approval
    const balance0 = await token0.balanceOf(accounts[0].address);
    const balance1 = await token1.balanceOf(accounts[0].address);
    console.log(`Balances before - ${await token0.symbol()}: ${ethers.formatUnits(balance0, token0Decimals)}, ${await token1.symbol()}: ${ethers.formatUnits(balance1, token1Decimals)}`);

    // Approve router
    console.log("Approving router...");
    await token0.approve(routerAddress, ethers.MaxUint256);
    await token1.approve(routerAddress, ethers.MaxUint256);
    
    // Verify approvals
    const allowance0 = await token0.allowance(accounts[0].address, routerAddress);
    const allowance1 = await token1.allowance(accounts[0].address, routerAddress);
    console.log(`Allowances - ${await token0.symbol()}: ${ethers.formatUnits(allowance0, token0Decimals)}, ${await token1.symbol()}: ${ethers.formatUnits(allowance1, token1Decimals)}`);

    // Add liquidity
    console.log("Adding liquidity...");
    try {
      const addLiquidityTx = await twSwapRouter.addLiquidity(
        token0Address,
        token1Address,
        amount0,
        amount1,
        0, // amountAMin
        0, // amountBMin
        accounts[0].address,
        ethers.MaxUint256 // deadline
      );
      
      console.log("Waiting for transaction...");
      const receipt = await addLiquidityTx.wait();
      console.log(`Transaction confirmed: ${receipt.hash}`);
      
      // Get pair address
      const pairAddress = await twSwapFactory.getPair(token0Address, token1Address);
      
      console.log(`Created pair ${await token0.symbol()}-${await token1.symbol()} at ${pairAddress}`);
    } catch (error) {
      console.error(`Error adding liquidity for ${await token0.symbol()}-${await token1.symbol()}:`, error);
      throw error; // Re-throw to stop the script
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
      decimals: await t.decimals()
    }))),
    pairs: await Promise.all(pairs.map(async pair => {
      const token0 = tokens[pair.token0];
      const token1 = tokens[pair.token1];
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

  return deploymentInfo;
}

main()
  .then((addresses) => {
    console.log("Deployed addresses:", addresses);
    process.exit(0);
  })
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });