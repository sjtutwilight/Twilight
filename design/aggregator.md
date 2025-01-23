
# 链聚合器
使用flink进行指标计算，java作为开发语言。从chain_transaction中解析出event列表，遍历列表进行处理sink到数据库。数据库结构在TableStucture.md
## pair指标处理
keyBy:contractAddress
窗口：多个滚动时间窗口，窗口时间为20s,1min,5min,30min
### 指标计算
指标为窗口期累加值，初始值为0：
- token0_volume_usd
- token1_volume_usd
- volume_usd
- txcnt
指标在窗口内更新状态：
- token0_reserve
- token1_reserve
- reserve_usd
## 事件处理逻辑
### Sync:
- 更新 token0_reserve, token1_reserve，如果涉及usdc,则刷新price
- 计算并更新 reserve_usd = token0_reserve * token0_price + token1_reserve * token1_price
- 增加 txcnt
### Swap:
- 更新 token0_volume_usd += amount0In * token0_price + amount0Out * token0_price
- 更新 token1_volume_usd 同理
- 增加 txcnt
### Mint/Burn:
- 仅增加 txcnt
- ## 缓存
- 先查缓存，缓存不存在再查数据库
- ### price缓存
- 在状态中维护各token 的price以供算子使用
- 当有usdc参与的交易对出现sync时，更新该交易对token的price
- ### token、pair 元数据
- 启动时从表中直接load到内存中
  
