-- 清除现有策略
DELETE FROM strategies;

-- 插入示例策略
INSERT INTO strategies (id, name, description, config) VALUES 
('price_alert', 'Price Alert Strategy', 'Monitors price changes and sends notifications', '{
  "id": "price_alert",
  "name": "Price Alert Strategy",
  "description": "Monitors price changes and sends notifications",
  "startNode": {
    "id": "start",
    "type": "trigger",
    "config": {
      "eventType": "price_update"
    },
    "nextNodes": ["check_threshold"]
  },
  "nodes": [
    {
      "id": "check_threshold",
      "type": "condition",
      "config": {
        "type": "threshold",
        "field": "data.price",
        "operator": "gt",
        "threshold": 1000
      },
      "nextNodes": ["check_frequency"]
    },
    {
      "id": "check_frequency",
      "type": "condition",
      "config": {
        "type": "frequency",
        "count": 3,
        "window": "5m"
      },
      "nextNodes": ["notify_user", "execute_chain"]
    },
    {
      "id": "notify_user",
      "type": "action",
      "config": {
        "type": "user",
        "user_id": "user123",
        "channel": "telegram"
      }
    },
    {
      "id": "execute_chain",
      "type": "action",
      "config": {
        "type": "chain",
        "contract": "0x1234...5678",
        "method": "executeRiskControl"
      }
    }
  ]
}'),
('liquidation_risk', 'Liquidation Risk Strategy', 'Monitors liquidation risks and takes preventive actions', '{
  "id": "liquidation_risk",
  "name": "Liquidation Risk Strategy",
  "description": "Monitors liquidation risks and takes preventive actions",
  "startNode": {
    "id": "start",
    "type": "trigger",
    "config": {
      "eventType": "health_factor_update"
    },
    "nextNodes": ["check_health"]
  },
  "nodes": [
    {
      "id": "check_health",
      "type": "condition",
      "config": {
        "type": "threshold",
        "field": "data.healthFactor",
        "operator": "lt",
        "threshold": 1.1
      },
      "nextNodes": ["notify_risk"]
    },
    {
      "id": "notify_risk",
      "type": "action",
      "config": {
        "type": "user",
        "user_id": "user456",
        "channel": "email"
      }
    }
  ]
}'); 