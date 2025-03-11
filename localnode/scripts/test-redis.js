// scripts/test-redis.js
const hre = require("hardhat");
const Redis = require('ioredis');
const { getTokenMetadata, getAccountMetadata, getPairMetadata } = require('./init/redis');

/**
 * Test Redis optimization by retrieving and displaying metadata
 */
async function main() {
  console.log("Testing Redis optimization...");
  
  // Initialize Redis connection
  const redis = new Redis({
    host: 'localhost',
    port: 6379
  });

  try {
    // Test token metadata
    console.log("\n=== TESTING TOKEN METADATA ===");
    const tokenMetadata = await redis.get('tokenMetadata');
    if (tokenMetadata) {
      const parsedTokenMetadata = JSON.parse(tokenMetadata);
      console.log(`Found ${parsedTokenMetadata.length} tokens in Redis`);
      
      // Display first token as example
      if (parsedTokenMetadata.length > 0) {
        console.log("Example token metadata:");
        console.log(JSON.stringify(parsedTokenMetadata[0], null, 2));
        
        // Test getTokenMetadata helper function
        const firstTokenAddress = parsedTokenMetadata[0].address;
        console.log(`\nTesting getTokenMetadata for address: ${firstTokenAddress}`);
        const tokenData = await getTokenMetadata(redis, firstTokenAddress);
        console.log("Result:", JSON.stringify(tokenData, null, 2));
      }
    } else {
      console.log("No token metadata found in Redis");
    }
    
    // Test account metadata
    console.log("\n=== TESTING ACCOUNT METADATA ===");
    const accountMetadata = await redis.get('accountMetadata');
    if (accountMetadata) {
      const parsedAccountMetadata = JSON.parse(accountMetadata);
      console.log(`Found ${parsedAccountMetadata.length} accounts in Redis`);
      
      // Display first account as example
      if (parsedAccountMetadata.length > 0) {
        console.log("Example account metadata:");
        console.log(JSON.stringify(parsedAccountMetadata[0], null, 2));
        
        // Test getAccountMetadata helper function
        const firstAccountAddress = parsedAccountMetadata[0].address;
        console.log(`\nTesting getAccountMetadata for address: ${firstAccountAddress}`);
        const accountData = await getAccountMetadata(redis, firstAccountAddress);
        console.log("Result:", JSON.stringify(accountData, null, 2));
      }
    } else {
      console.log("No account metadata found in Redis");
    }
    
    // Test pair metadata
    console.log("\n=== TESTING PAIR METADATA ===");
    const pairMetadata = await redis.get('pairMetadata');
    if (pairMetadata) {
      const parsedPairMetadata = JSON.parse(pairMetadata);
      console.log(`Found ${parsedPairMetadata.length} pairs in Redis`);
      
      // Display first pair as example
      if (parsedPairMetadata.length > 0) {
        console.log("Example pair metadata:");
        console.log(JSON.stringify(parsedPairMetadata[0], null, 2));
        
        // Test getPairMetadata helper function
        const firstPairAddress = parsedPairMetadata[0].address;
        console.log(`\nTesting getPairMetadata for address: ${firstPairAddress}`);
        const pairData = await getPairMetadata(redis, firstPairAddress);
        console.log("Result:", JSON.stringify(pairData, null, 2));
      }
    } else {
      console.log("No pair metadata found in Redis");
    }
    
    // Test token prices
    console.log("\n=== TESTING TOKEN PRICES ===");
    const tokenPriceKeys = await redis.keys('token_price:*');
    console.log(`Found ${tokenPriceKeys.length} token prices in Redis`);
    
    // Display a few token prices as examples
    if (tokenPriceKeys.length > 0) {
      const sampleSize = Math.min(tokenPriceKeys.length, 5);
      for (let i = 0; i < sampleSize; i++) {
        const key = tokenPriceKeys[i];
        const price = await redis.get(key);
        console.log(`${key}: ${price} USD`);
      }
    }
    
    console.log("\nRedis optimization test completed successfully");
  } catch (error) {
    console.error("Error during Redis test:", error);
  } finally {
    // Close Redis connection
    await redis.quit();
  }
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main()
  .then(() => {
    console.log("Redis test complete");
    process.exit(0);
  })
  .catch((error) => {
    console.error(error);
    process.exit(1);
  }); 