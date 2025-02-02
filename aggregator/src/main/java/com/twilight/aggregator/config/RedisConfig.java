package com.twilight.aggregator.config;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import redis.clients.jedis.Jedis;
import redis.clients.jedis.JedisPool;
import redis.clients.jedis.JedisPoolConfig;

public class RedisConfig extends BaseConfig {
    private static final Logger LOG = LoggerFactory.getLogger(RedisConfig.class);
    private static volatile RedisConfig instance;
    private volatile JedisPool jedisPool;
    private static final int MAX_RETRY_ATTEMPTS = 3;
    private static final long RETRY_DELAY_MS = 1000;

    private RedisConfig() {
        super();
        initializeConfig();
    }

    public static synchronized RedisConfig getInstance() {
        if (instance == null) {
            instance = new RedisConfig();
        }
        return instance;
    }

    private void initializeConfig() {
        // Redis配置已经在父类中加载
    }

    // Redis连接配置
    public String getRedisHost() {
        return getProperty("redis.host", "localhost");
    }

    public int getRedisPort() {
        return getIntProperty("redis.port", "6379");
    }

    public String getRedisPassword() {
        return getProperty("redis.password", "");
    }

    // Redis连接池配置
    public int getRedisMaxTotal() {
        return getIntProperty("redis.pool.maxTotal", "100");
    }

    public int getRedisMaxIdle() {
        return getIntProperty("redis.pool.maxIdle", "20");
    }

    public int getRedisMinIdle() {
        return getIntProperty("redis.pool.minIdle", "10");
    }

    public long getRedisMaxWaitMillis() {
        return getLongProperty("redis.pool.maxWaitMillis", "10000");
    }

    public boolean isRedisTestOnBorrow() {
        return getBooleanProperty("redis.pool.testOnBorrow", "true");
    }

    public boolean isRedisTestWhileIdle() {
        return getBooleanProperty("redis.pool.testWhileIdle", "true");
    }

    public long getRedisTimeBetweenEvictionRunsMillis() {
        return getLongProperty("redis.pool.timeBetweenEvictionRunsMillis", "30000");
    }

    // Redis连接池管理
    public synchronized JedisPool getJedisPool() {
        if (jedisPool == null) {
            int attempts = 0;
            while (attempts < MAX_RETRY_ATTEMPTS) {
                try {
                    JedisPoolConfig config = new JedisPoolConfig();
                    config.setMaxTotal(getRedisMaxTotal());
                    config.setMaxIdle(getRedisMaxIdle());
                    config.setMinIdle(getRedisMinIdle());
                    config.setMaxWaitMillis(getRedisMaxWaitMillis());
                    config.setTestOnBorrow(isRedisTestOnBorrow());
                    config.setTestWhileIdle(isRedisTestWhileIdle());
                    config.setTimeBetweenEvictionRunsMillis(getRedisTimeBetweenEvictionRunsMillis());
                    
                    // 添加连接测试配置
                    config.setTestOnReturn(true);
                    config.setTestOnCreate(true);
                    config.setJmxEnabled(false);
                    
                    // 设置驱逐策略
                    config.setMinEvictableIdleTimeMillis(60000);
                    config.setTimeBetweenEvictionRunsMillis(30000);
                    config.setNumTestsPerEvictionRun(3);
                    config.setBlockWhenExhausted(true);

                    String password = getRedisPassword();
                    if (password != null && !password.isEmpty()) {
                        jedisPool = new JedisPool(config, getRedisHost(), getRedisPort(), 2000, password);
                    } else {
                        jedisPool = new JedisPool(config, getRedisHost(), getRedisPort(), 2000);
                    }
                    
                    // 测试连接池
                    try (Jedis jedis = jedisPool.getResource()) {
                        jedis.ping();
                    }
                    
                    LOG.info("Redis connection pool initialized successfully");
                    break;
                } catch (Exception e) {
                    attempts++;
                    if (attempts >= MAX_RETRY_ATTEMPTS) {
                        LOG.error("Failed to initialize Redis connection pool after {} attempts", MAX_RETRY_ATTEMPTS, e);
                        throw new RuntimeException("Failed to initialize Redis connection pool", e);
                    }
                    LOG.warn("Failed to initialize Redis connection pool, attempt {}/{}, retrying in {} ms", 
                            attempts, MAX_RETRY_ATTEMPTS, RETRY_DELAY_MS, e);
                    try {
                        Thread.sleep(RETRY_DELAY_MS);
                    } catch (InterruptedException ie) {
                        Thread.currentThread().interrupt();
                        throw new RuntimeException("Interrupted while waiting to retry Redis connection", ie);
                    }
                }
            }
        }
        return jedisPool;
    }

    /**
     * 获取一个 Redis 连接。
     * 注意：调用者必须在 try-with-resources 块中使用此方法返回的连接，以确保资源被正确释放。
     * 示例：
     * try (Jedis jedis = redisConfig.getResource()) {
     *     // 使用 jedis
     * }
     */
    public Jedis getResource() {
        JedisPool pool = getJedisPool();
        int attempts = 0;
        while (attempts < MAX_RETRY_ATTEMPTS) {
            try {
                Jedis jedis = pool.getResource();
                // 测试连接是否可用
                jedis.ping();
                return jedis;
            } catch (Exception e) {
                attempts++;
                if (attempts >= MAX_RETRY_ATTEMPTS) {
                    LOG.error("Failed to get Redis connection after {} attempts", MAX_RETRY_ATTEMPTS, e);
                    throw new RuntimeException("Failed to get Redis connection", e);
                }
                LOG.warn("Failed to get Redis connection, attempt {}/{}, retrying in {} ms", 
                        attempts, MAX_RETRY_ATTEMPTS, RETRY_DELAY_MS, e);
                try {
                    Thread.sleep(RETRY_DELAY_MS);
                } catch (InterruptedException ie) {
                    Thread.currentThread().interrupt();
                    throw new RuntimeException("Interrupted while waiting to retry Redis connection", ie);
                }
            }
        }
        throw new RuntimeException("Failed to get Redis connection after " + MAX_RETRY_ATTEMPTS + " attempts");
    }

    public synchronized void close() {
        if (jedisPool != null) {
            try {
                jedisPool.close();
            } catch (Exception e) {
                LOG.error("Error closing Redis connection pool", e);
            } finally {
                jedisPool = null;
            }
        }
    }
}
