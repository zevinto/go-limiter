package main

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// 滑动窗口限流（Sliding Window）

var ctx = context.Background()

type SlidingWindowLimiter struct {
	client    *redis.Client
	key       string
	limit     int
	windowSec int64
}

func NewSlidingWindowLimiter(client *redis.Client, key string, limit int, windowSec int64) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		client:    client,
		key:       key,
		limit:     limit,
		windowSec: windowSec,
	}
}

func (l *SlidingWindowLimiter) AllowRequest() bool {
	now := time.Now().Unix()
	pipe := l.client.TxPipeline()
	// 移除窗口外的请求记录
	pipe.ZRemRangeByScore(ctx, l.key, "0", fmt.Sprintf("%d", now-l.windowSec))
	// 获取窗口内的请求数
	countCmd := pipe.ZCard(ctx, l.key)
	// 添加当前请求
	pipe.ZAdd(ctx, l.key, redis.Z{Score: float64(now), Member: now})
	// 设置过期时间
	pipe.Expire(ctx, l.key, time.Duration(l.windowSec)*time.Second)

	_, _ = pipe.Exec(ctx)
	count := countCmd.Val()
	return count < int64(l.limit)
}

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "root",
		DB:       0,
	})

	limiter := NewSlidingWindowLimiter(client, "sliding_window", 5, 10)

	for i := 0; i < 10; i++ {
		if limiter.AllowRequest() {
			fmt.Println("请求通过")
		} else {
			fmt.Println("请求被拒绝")
		}
		time.Sleep(1 * time.Second)
	}
}
