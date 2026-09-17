package engine

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPool_BasicSubmit(t *testing.T) {
	pool := NewPool(0, 10)
	defer pool.Close()

	var counter atomic.Int64
	for i := 0; i < 100; i++ {
		pool.Submit(func() {
			counter.Add(1)
		})
	}

	// 等待所有任务完成
	deadline := time.Now().Add(5 * time.Second)
	for counter.Load() < 100 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if counter.Load() != 100 {
		t.Errorf("counter = %d, want 100", counter.Load())
	}
}

func TestPool_DynamicExpand(t *testing.T) {
	pool := NewPool(2, 50)
	defer pool.Close()

	var maxConcurrent atomic.Int64
	var current atomic.Int64

	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pool.Submit(func() {
				cur := current.Add(1)
				if cur > maxConcurrent.Load() {
					maxConcurrent.Store(cur)
				}
				time.Sleep(5 * time.Millisecond)
				current.Add(-1)
			})
		}()
	}

	wg.Wait()
	pool.Wait()

	if maxConcurrent.Load() == 0 {
		t.Error("expected some concurrent execution")
	}
	if maxConcurrent.Load() > 50 {
		t.Errorf("max concurrent = %d, exceeds pool max of 50", maxConcurrent.Load())
	}
	t.Logf("max concurrent execution: %d", maxConcurrent.Load())
}

func TestPool_Close_StopsAccepting(t *testing.T) {
	pool := NewPool(0, 5)

	pool.Close()

	if pool.Submit(func() {}) {
		t.Error("expected Submit to return false after Close")
	}
}

func TestPool_Size(t *testing.T) {
	pool := NewPool(2, 10)
	defer pool.Close()

	if pool.Size() < 2 {
		t.Errorf("size = %d, want >= 2", pool.Size())
	}

	// 触发扩容
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pool.Submit(func() {
				time.Sleep(10 * time.Millisecond)
			})
		}()
	}
	wg.Wait()

	t.Logf("pool size after submit: %d", pool.Size())
}

func TestPool_ParallelSafety(t *testing.T) {
	pool := NewPool(5, 20)
	defer pool.Close()

	var wg sync.WaitGroup
	var errors atomic.Int64

	// 100 个 goroutine 并发提交
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				if !pool.Submit(func() {
					// 简单的任务
					_ = time.Now().UnixNano()
				}) {
					errors.Add(1)
				}
			}
		}()
	}

	wg.Wait()
	pool.Wait()

	if errors.Load() > 0 {
		t.Errorf("got %d submit errors", errors.Load())
	}
}