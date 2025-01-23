package com.twilight.aggregator.job;

import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public abstract class BaseJob {
    protected static final Logger logger = LoggerFactory.getLogger(BaseJob.class);
    protected final StreamExecutionEnvironment env;

    protected BaseJob() {
        this.env = StreamExecutionEnvironment.getExecutionEnvironment();
        configureEnvironment();
    }

    protected void configureEnvironment() {
        env.setParallelism(1);
        env.enableCheckpointing(60000); // 1 minute
    }

    public abstract void execute() throws Exception;
} 