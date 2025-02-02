package com.twilight.aggregator.serialization;

import org.apache.flink.api.common.serialization.DeserializationSchema;
import org.apache.flink.api.common.typeinfo.TypeInformation;

import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.twilight.aggregator.model.KafkaMessage;
import com.twilight.aggregator.model.Transaction;

public class TransactionDeserializer implements DeserializationSchema<Transaction> {
    private static final ObjectMapper MAPPER = new ObjectMapper()
        .configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false)
        .configure(DeserializationFeature.ACCEPT_EMPTY_STRING_AS_NULL_OBJECT, true)
        .configure(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS, false)
        .findAndRegisterModules();

    @Override
    public Transaction deserialize(byte[] message) {
        try {
            if (message == null || message.length == 0) {
                return null;
            }
            KafkaMessage kafkaMessage = MAPPER.readValue(message, KafkaMessage.class);
            if (kafkaMessage == null) {
                return null;
            }
            Transaction transaction = kafkaMessage.getTransaction();
            if (transaction != null) {
                transaction.setEvents(kafkaMessage.getEvents());
            }
            return transaction;
        } catch (Exception e) {
            throw new RuntimeException("Failed to deserialize transaction: " + new String(message), e);
        }
    }

    @Override
    public boolean isEndOfStream(Transaction nextElement) {
        return false;
    }

    @Override
    public TypeInformation<Transaction> getProducedType() {
        return TypeInformation.of(Transaction.class);
    }
} 