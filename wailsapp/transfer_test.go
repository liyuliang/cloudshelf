package main

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestParallelRunLimitsConcurrency 验证四路限流确实生效且错误按任务下标归位。
func TestParallelRunLimitsConcurrency(t *testing.T) {
	var cur int32
	var peak int32
	var mu sync.Mutex

	const n = 12
	errs := parallelRun(n, maxConcurrent, func(i int) error {
		c := atomic.AddInt32(&cur, 1)
		mu.Lock()
		if c > peak {
			peak = c
		}
		mu.Unlock()
		time.Sleep(5 * time.Millisecond)
		atomic.AddInt32(&cur, -1)
		if i == 7 {
			return errors.New("boom")
		}
		return nil
	})

	mu.Lock()
	gotPeak := peak
	mu.Unlock()
	if gotPeak > maxConcurrent {
		t.Fatalf("并发峰值 = %d，超过上限 %d", gotPeak, maxConcurrent)
	}
	if gotPeak < 2 {
		t.Fatalf("并发峰值 = %d，任务似乎没有真正并行", gotPeak)
	}
	if len(errs) != n {
		t.Fatalf("错误切片长度 = %d，期望 %d", len(errs), n)
	}
	fails := 0
	for i, e := range errs {
		if e != nil {
			fails++
			if i != 7 {
				t.Fatalf("错误出现在非预期下标 %d: %v", i, e)
			}
		}
	}
	if fails != 1 {
		t.Fatalf("失败任务数 = %d，期望 1", fails)
	}
	if cur != 0 {
		t.Fatalf("仍有 %d 个任务未退出", cur)
	}
}
