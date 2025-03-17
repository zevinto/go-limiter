package main

import (
	"fmt"
	"sync"
	"time"
)

// 并发控制（Semaphore）

type SemaphoreLimiter struct {
	sem chan struct{}
}

func NewSemaphoreLimiter(maxConcurrency int) *SemaphoreLimiter {
	return &SemaphoreLimiter{
		sem: make(chan struct{}, maxConcurrency),
	}
}

func (s *SemaphoreLimiter) Acquire() {
	s.sem <- struct{}{}
}
func (s *SemaphoreLimiter) Release() {
	<-s.sem
}

func main() {
	limiter := NewSemaphoreLimiter(3) // 允许最多 3 个并发任务
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			limiter.Acquire()
			fmt.Printf("Task %d is running\n", i)
			time.Sleep(2 * time.Second)
			fmt.Printf("Task %d is finished\n", i)
			limiter.Release()
		}(i)
	}

	wg.Wait()
}
