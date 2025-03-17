限流（Rate Limiting）主要用于控制系统的请求流量，防止流量突增导致系统崩溃。常见的限流方式包括以下几种：

---

### **1. 计数器（Counter）限流**
**原理：** 设定一个固定时间窗口（如 1 秒），在该时间内请求数达到上限后拒绝新的请求。  
**适用场景：** 简单的限流需求，如 API 调用限制。  
**实现方式：**
- Redis `INCR` 计数
- 数据库计数

**缺点：** 可能会出现“临界突发”问题（如 59 秒 100 次请求，60 秒 100 次请求，导致短时间内 200 次请求）。

---

### **2. 滑动窗口（Sliding Window）限流**
**原理：** 采用更精细的时间窗口，如按 1 秒切片统计过去 N 秒内的请求总量，保证限流更均匀。  
**适用场景：** 适用于对流量更加均衡控制的场景。  
**实现方式：**
- Redis `ZSET` 记录时间戳 + `ZREMRANGEBYSCORE` 清理过期请求

**优点：** 相比固定窗口，滑动窗口可以减少临界突发流量。

---

### **3. 令牌桶（Token Bucket）限流**
**原理：** 令牌按照固定速率放入桶中，请求需消耗令牌，桶满后不再放入新令牌。  
**适用场景：** 允许突发流量，但需要控制平均速率的情况，如 API 访问控制。  
**实现方式：**
- Redis `SETNX` + `EXPIRE` 实现分布式令牌桶
- Guava RateLimiter（Java）

**优点：**
- 允许短时间突发流量（只要桶里有令牌）
- 控制请求速率较稳定

**缺点：** 当流量持续高于令牌填充速率时，会导致请求被限制。

---

### **4. 漏桶（Leaky Bucket）限流**
**原理：** 以固定速率处理请求（漏水速率恒定），超出容量的请求被丢弃。  
**适用场景：** 适用于处理请求均匀的场景，如限流 API 调用、限流流量等。  
**实现方式：**
- Redis `LIST` 维护请求队列，定时消费

**优点：**
- 能够平滑处理流量，避免流量突增
- 适合流量均匀处理的业务

**缺点：** 不能应对流量突发，即使短时间内有大量请求，仍需等待队列处理。

---

### **5. 并发限制（Semaphore / 并发数控制）**
**原理：** 通过信号量（Semaphore）限制同时进行的请求数。  
**适用场景：** 限制并发执行的任务，如数据库连接池、线程池。  
**实现方式：**
- Redis `SETNX` 控制并发数
- Java `Semaphore` 控制线程并发

**优点：**
- 控制系统资源的使用，防止超载
- 适用于需要严格限制并发数的业务

**缺点：** 可能会导致队列等待，影响响应时间。

---

### **6. Nginx 限流**
**原理：** 通过 Nginx 配置 `limit_req_zone` 和 `limit_req` 限制请求速率。  
**适用场景：** 适用于 Web 服务器/API 网关层面的限流。  
**实现方式：**
```nginx
http {
    limit_req_zone $binary_remote_addr zone=one:10m rate=10r/s;
    
    server {
        location /api/ {
            limit_req zone=one burst=5 nodelay;
        }
    }
}
```
**优点：**
- 直接在 Nginx 层面限制请求，减少后端压力
- 适用于 API 请求限流

**缺点：** 不能按业务逻辑限流，需结合 Redis 实现更复杂策略。

---

### **7. MQ（消息队列）削峰**
**原理：** 高并发时，业务请求进入 MQ（如 Kafka、RabbitMQ），消费者按固定速率消费，避免瞬时高流量压垮系统。  
**适用场景：** 适用于需要异步处理的高并发业务，如订单处理、日志收集等。  
**实现方式：**
- 生产者推送请求到 MQ，消费者按限流规则消费

**优点：**
- 削峰填谷，缓解瞬时压力
- 保证系统稳定性

**缺点：** 需要处理延迟问题，不适用于强实时业务。

---

### **8. 基于用户 / IP / 业务类型的限流**
**原理：** 按不同维度进行限流，如用户 ID、IP 地址、接口类型等。  
**适用场景：** 适用于防止单个用户或 IP 滥用资源，如防 DDoS、API 访问控制等。  
**实现方式：**
- Redis `HINCRBY` 统计每个用户或 IP 的访问频率
- 配合防火墙（如 WAF）实现安全防护

**优点：** 精细化限流，防止特定用户滥用系统资源。

---

### **9. Redis + Lua 限流**
**原理：** 通过 Redis + Lua 脚本实现精准限流，如令牌桶算法或滑动窗口限流。  
**适用场景：** 适用于高性能限流需求，如 API 网关、微服务架构。  
**实现方式：** Redis `EVAL` 执行 Lua 脚本，保证限流操作的原子性。

示例：Redis 令牌桶 Lua 脚本
```lua
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local tokens = tonumber(redis.call("get", key) or capacity)
local last_time = tonumber(redis.call("hget", key .. ":info", "last_time") or now)

local elapsed = now - last_time
local new_tokens = math.min(capacity, tokens + elapsed * rate)

if new_tokens < 1 then
    return 0  -- 拒绝请求
else
    redis.call("set", key, new_tokens - 1)
    redis.call("hset", key, "last_time", now)
    return 1  -- 允许请求
end
```

**优点：**
- 支持高性能分布式限流
- 确保操作的原子性

---

### **总结**
| 限流方式 | 适用场景 | 允许突发流量 | 实现难度 |  
|----------|---------|-------------|----------|  
| 计数器限流 | 简单 API 限流 | ❌ | 低 |  
| 滑动窗口 | 均匀流量控制 | ✅ | 中 |  
| 令牌桶 | 短时突发流量 | ✅ | 中 |  
| 漏桶 | 流量均匀处理 | ❌ | 中 |  
| 并发控制 | 线程池、数据库连接池 | ✅ | 低 |  
| Nginx 限流 | API 网关层 | ❌ | 低 |  
| MQ 削峰 | 高并发请求 | ✅ | 高 |  
| 用户/IP 限流 | 防滥用 | ❌ | 低 |  
| Redis + Lua | 高性能限流 | ✅ | 高 |  

