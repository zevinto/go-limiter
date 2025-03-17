package main

import (
	"fmt"
	"golang.org/x/time/rate"
	"sync"
)

// 令牌桶限流（Token Bucket）

type TokenBucketLimiter struct {
	limiter *rate.Limiter
}

func NewTokenBucketLimiter(rateLimit rate.Limit, bucketSize int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		limiter: rate.NewLimiter(rateLimit, bucketSize),
	}
}

func (t *TokenBucketLimiter) AllowRequest() bool {
	return t.limiter.Allow()
}

func main() {
	limiter := NewTokenBucketLimiter(2, 5) // 2 次/秒，最大 5 个令牌
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if limiter.AllowRequest() {
				fmt.Printf("Request %d Allowed\n", i)
			} else {
				fmt.Printf("Request %d Denied\n", i)
			}
		}(i)
		//time.Sleep(500 * time.Millisecond)
	}

	wg.Wait()
}
