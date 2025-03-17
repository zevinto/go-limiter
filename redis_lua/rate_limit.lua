local key = KEYS[1]               -- 限流的 Redis Key，例如 "rate_limit"
local rate = tonumber(ARGV[1])     -- 令牌生成速率，每秒补充多少个令牌
local capacity = tonumber(ARGV[2]) -- 令牌桶的最大容量
local now = tonumber(ARGV[3])      -- 当前时间戳（秒）

-- 确保令牌桶的键是数值
local tokens = tonumber(redis.call("get", key))
if not tokens then
    tokens = capacity  -- 如果 Redis 里没有这个 key，初始化为最大容量
end

-- 删除可能存在的旧哈希表
redis.call("del", key .. ":info")

-- 确保 `last_time` 也是数值
local last_time = tonumber(redis.call("hget", key .. ":info", "last_time"))
if not last_time then
    last_time = now  -- 如果 Redis 里没有存储 `last_time`，说明是第一次请求，设置为当前时间
end

-- 计算新的令牌数
local elapsed = now - last_time  -- 计算距离上次请求过去了多少秒
local new_tokens = math.min(capacity, tokens + elapsed * rate)  -- 按速率补充令牌

if new_tokens < 1 then
    -- 即使拒绝请求也更新 last_time，避免时间计算错误
    redis.call("hset", key .. ":info", "last_time", now)
    return 0  -- 如果桶里没令牌了，拒绝请求
else
    redis.call("set", key, new_tokens - 1) -- 消耗 1 个令牌
    redis.call("hset", key .. ":info", "last_time", now) -- 更新 last_time
    return 1  -- 允许请求
end

