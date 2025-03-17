package main

// Redis + Lua 限流
// https://github.com/redis/redis-doc
import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "root",
		DB:       0,
	})

	// 测试 Redis 连接
	_, err := client.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Failed to connect to Redis:", err)
		return
	}

	// 清理可能存在的旧 key
	if err := client.Del(ctx, "rate_limit").Err(); err != nil {
		fmt.Println("Failed to clean up old key:", err)
		return
	}

	scriptBytes, err := os.ReadFile("rate_limit.lua")
	if err != nil {
		fmt.Println("Failed to read Lua script:", err)
		return
	}

	// 创建脚本对象
	luaScript := redis.NewScript(string(scriptBytes))

	// 获取当前时间戳
	now := time.Now().Unix()

	// 执行lua脚本
	result, err := luaScript.Run(ctx, client, []string{"rate_limit"}, 1, 10, now).Result()
	if err != nil {
		fmt.Println("Failed to execute Lua script:", err)
		return
	}

	if result.(int64) == 1 {
		fmt.Println("Request Allowed")
	} else {
		fmt.Println("Rate Limit Exceeded")
	}
}
