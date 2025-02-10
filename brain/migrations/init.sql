-- 创建策略表
CREATE TABLE IF NOT EXISTS strategies (
    id TEXT PRIMARY KEY,
    name TEXT,
    start_node_id TEXT,
    nodes JSONB
);



-- 插入测试策略
INSERT INTO strategies (id, name, start_node_id, nodes) VALUES (
    'lpRiskAlert',
    'LP无常损失预警策略',
    'node1',
    '[
       {
         "id": "node1",
         "type": "eventTrigger",
         "config": {"eventType":"priceChange"},
         "next": ["node2"]
       },
       {
         "id": "node2",
         "type": "condition",
         "config": {"rule":"checkIfUnstablePool","threshold":3.0},
         "next": ["node3","nodeEnd"]
       },
       {
         "id": "node3",
         "type": "action",
         "config": {"actionType":"notifyUsers","groupId":"lp_holders_pool_A","message":"LP池风险预警"},
         "next": ["node4"]
       },
       {
         "id": "node4",
         "type": "delay",
         "config": {"durationMs":5000},
         "next": ["node5"]
       },
       {
         "id": "node5",
         "type": "condition",
         "config": {"rule":"checkRiskLevelIncreased","threshold":5.0},
         "next": ["node6","nodeEnd"]
       },
       {
         "id": "node6",
         "type": "action",
         "config": {"actionType":"mockChainOp","operation":"autoWithdrawLiquidity"},
         "next": ["nodeEnd"]
       },
       {
         "id": "nodeEnd",
         "type": "end",
         "config": {},
         "next": []
       }
     ]'
);

