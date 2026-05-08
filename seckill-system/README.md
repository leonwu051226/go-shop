# Seckill System 项目日志

## 2026-05-07 23:46 CST — 数据库接入与核心接口完成

### 已完成工作

1. **Viper 配置读取** (`pkg/config/config.go`)
   - 使用 Viper 读取 `config/config.yaml`
   - 配置结构涵盖 Server、MySQL、Redis、RabbitMQ
   - MySQL 密码已对齐 Docker Compose 中的 `seckill123456`，端口使用 `3308`

2. **GORM 数据库连接** (`pkg/database/mysql.go`)
   - 完成 MySQL DSN 拼接与 GORM 初始化
   - 支持设置 MaxOpenConns / MaxIdleConns

3. **核心数据模型** (`internal/model/`)
   - `User`：用户表（ID、Username、PasswordHash、时间戳）
   - `Product`：普通商品表（ID、Name、Description、Price、Stock、时间戳）
   - `SeckillProduct`：秒杀商品表（ID、ProductID、SeckillPrice、Stock、StartTime、EndTime、Status、时间戳）
   - `Order`：订单表（ID、OrderNo、UserID、ProductID、SeckillProductID、Quantity、TotalPrice、Status、时间戳）

4. **自动迁移** (`cmd/seckill/main.go`)
   - 程序启动时自动执行 `AutoMigrate`，自动建表
   - 包含测试数据初始化（默认用户 admin / 123456，3 个商品及对应的秒杀活动）

5. **基础业务接口**
   - `POST /api/v1/register` — 用户注册
   - `POST /api/v1/login` — 用户登录，返回 JWT Token
   - `GET /api/v1/seckill/products` — 获取秒杀商品列表（带关联商品信息）

6. **Bug 修复**
   - 修复 `main.go` 中 Server Port 格式问题：`config.yaml` 中配置为 `8080`，程序自动补全为 `:8080`，避免 Gin 监听报错

### 验证结果

```bash
# Ping
curl http://localhost:8080/ping
# {"message":"pong"}

# 登录
curl -X POST http://localhost:8080/api/v1/login -H "Content-Type: application/json" -d '{"username":"admin","password":"123456"}'
# {"code":0,"msg":"success","data":{"token":"..."}}

# 秒杀商品列表
curl http://localhost:8080/api/v1/seckill/products
# {"code":0,"msg":"success","data":[...]}
```

### 技术栈

- Go 1.26.2 + Gin
- GORM + MySQL 8.0
- JWT (golang-jwt/jwt/v5)
- bcrypt 密码哈希
- Viper 配置管理

---

## 2026-05-07 23:57 CST — 高并发秒杀核心逻辑完成

### 已完成工作

1. **Redis 客户端接入** (`pkg/database/redis.go`)
   - 使用 go-redis/v9 连接 Docker 中的 Redis 7.0
   - 端口调整为 `6380`（避免与本地 Redis 冲突）

2. **RabbitMQ 接入** (`pkg/rabbitmq/rabbitmq.go`)
   - 使用 amqp091-go 连接 RabbitMQ 3
   - 声明 `seckill.order.exchange` (direct) + `seckill.order.queue`
   - 实现 `PublishOrderMessage` 方法，消息持久化投递

3. **库存预热逻辑** (`internal/service/seckill.go`)
   - `PreheatStock()`：项目启动时自动将 MySQL 中秒杀商品库存加载到 Redis
   - Redis Key 设计：`seckill:stock:{seckill_product_id}` (String)
   - 用户参与记录：`seckill:users:{seckill_product_id}` (Set)

4. **Lua 原子扣减脚本** (`internal/service/seckill.go`)
   - 单脚本内完成：检查重复购买 → 检查库存 → 扣减库存 → 记录用户
   - 通过 `redis.NewScript` 注册，保证 Redis 端原子执行
   - 返回值：1=成功, 0=库存不足, -1=已购买, -2=系统错误

5. **JWT 认证中间件** (`internal/middleware/jwt.go`)
   - 解析 `Authorization: Bearer <token>` 头部
   - 验证通过后注入 `user_id` 和 `username` 到 Gin Context

6. **秒杀抢购接口** (`internal/api/seckill.go`)
   - `POST /api/v1/seckill/do`（需 JWT 认证）
   - 流程：解析 UserID → 执行 Lua 扣减 → 生成 OrderID → 发 RabbitMQ → 返回 "queuing"
   - 防重机制：同一用户同一商品只能抢购一次

7. **基础设施端口调整**
   - MySQL: `3308`（避免与本地 MySQL 冲突）
   - Redis: `6380`（避免与本地 Redis 冲突）

### 验证结果

```bash
# 1. 登录获取 Token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456"}' | grep -o '"token":"[^"]*"' | sed 's/"token":"//;s/"$//')

# 2. 获取秒杀商品列表
curl http://localhost:8080/api/v1/seckill/products
# {"code":0,"msg":"success","data":[{"id":1,"product_id":1,"seckill_price":4999,"stock":10,...},...]}

# 3. 执行秒杀（带 JWT Token）
curl -X POST http://localhost:8080/api/v1/seckill/do \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"seckill_product_id":1}'
# {"code":0,"msg":"success","data":{"order_id":"...","status":"queuing","message":"your order is being processed"}}

# 4. 重复抢购（应失败）
curl -X POST http://localhost:8080/api/v1/seckill/do \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"seckill_product_id":1}'
# {"code":403,"msg":"you have already purchased this item"}
```

### 技术栈更新

- go-redis/v9
- amqp091-go (RabbitMQ)
- google/uuid

### 数据验证

- Redis 库存：`seckill:stock:1` 从 10 → 9 ✅
- RabbitMQ 队列：`seckill.order.queue` messages_ready = 1 ✅
- 防重集合：`seckill:users:1` 包含用户 ID 1 ✅

---

## 2026-05-08 11:45 CST — RabbitMQ 消费者与事务落库完成

### 已完成工作

1. **订单模型幂等性增强** (`internal/model/order.go`)
   - 添加联合唯一索引 `idx_user_seckill` (user_id + seckill_product_id)
   - 配合已有的 `order_no` 唯一索引，形成双重幂等保障

2. **RabbitMQ 消费者** (`pkg/rabbitmq/consumer.go`)
   - `StartOrderConsumer(db)`：在独立 Goroutine 中持续监听 `seckill.order.queue`
   - 使用独立 Channel，避免与生产者通道并发冲突
   - 手动 ACK/NACK 机制：
     - 成功处理 → `Ack`
     - 重复消息（唯一索引冲突）→ `Ack`（避免重复消费）
     - 库存不足 → `Nack` + 不重新入队（丢弃）
     - 系统错误 → `Nack` + 重新入队（重试）

3. **事务安全落库** (`pkg/rabbitmq/consumer.go` `processOrder`)
   - GORM Transaction 包裹两步操作：
     1. **乐观锁扣减库存**：`UPDATE seckill_products SET stock = stock - 1 WHERE id = ? AND stock > 0`
     2. **插入订单记录**：利用唯一索引保证同一条消息不会生成两个订单
   - 事务失败自动回滚，确保数据一致性

4. **main.go 集成**
   - 启动消费者 Goroutine：`go rabbitmq.StartOrderConsumer(db)`

### 验证结果

```bash
# 启动服务后，消费者会自动处理队列中的消息
# 日志输出示例：
# 2026/05/08 11:45:08 Order consumer started, listening on queue: seckill.order.queue
# 2026/05/08 11:45:08 [Consumer] Order processed successfully. orderNo=f8485f3e-c484-4cb9-af8a-a7808b9a47d4
```

### 如何验证消息已被成功处理

**1. 查看 MySQL 订单表**

```bash
docker exec seckill-mysql mysql -u seckill -pseckill123456 -D seckill -e "SELECT id, order_no, user_id, seckill_product_id, status FROM orders;"
```

预期输出（包含一条已处理订单）：
```
+----+--------------------------------------+---------+------------------+--------+
| id | order_no                             | user_id | seckill_product_id | status |
+----+--------------------------------------+---------+------------------+--------+
|  1 | f8485f3e-c484-4cb9-af8a-a7808b9a47d4 |       1 |                1 |      1 |
+----+--------------------------------------+---------+------------------+--------+
```

**2. 查看 MySQL 库存扣减**

```bash
docker exec seckill-mysql mysql -u seckill -pseckill123456 -D seckill -e "SELECT id, stock FROM seckill_products WHERE id = 1;"
```

预期输出（stock 从 10 变为 9）：
```
+----+-------+
| id | stock |
+----+-------+
|  1 |     9 |
+----+-------+
```

**3. 查看 RabbitMQ 队列状态**

```bash
docker exec seckill-rabbitmq rabbitmqctl list_queues name messages_ready messages_unacknowledged
```

预期输出（消息已被消费完毕）：
```
name	messages_ready	messages_unacknowledged
seckill.order.queue	0	0
```
