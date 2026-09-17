package engine

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiter_Noop(t *testing.T) {
	limiter := NewRateLimiter(0, 0)
	if _, ok := limiter.(*noopRateLimiter); !ok {
		t.Error("expected noop limiter for rps=0")
	}
	if !limiter.Allow() {
		t.Error("noop limiter should always allow")
	}
	if err := limiter.Wait(context.Background()); err != nil {
		t.Errorf("noop limiter should not error: %v", err)
	}
}

func TestRateLimiter_BasicThrottle(t *testing.T) {
	// 100 RPS,burst=10
	limiter := NewRateLimiter(100, 10)

	// 前 10 个应该立即通过(burst)
	startTime := time.Now()
	allowed := 0
	for i := 0; i < 10; i++ {
		if limiter.Allow() {
			allowed++
		}
	}
	elapsed := time.Since(startTime)
	if allowed != 10 {
		t.Errorf("expected 10 immediate allows, got %d", allowed)
	}
	if elapsed > 50*time.Millisecond {
		t.Errorf("burst took too long: %v", elapsed)
	}
}

func TestRateLimiter_RateAccuracy(t *testing.T) {
	// 100 RPS
	limiter := NewRateLimiter(100, 1)

	// 在 1 秒内发送请求,计数
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	count := 0
	for {
		if err := limiter.Wait(ctx); err != nil {
			break
		}
		count++
	}

	// 允许 ±10% 误差
	if count < 85 || count > 115 {
		t.Errorf("expected ~100 requests in 1s, got %d", count)
	}
	t.Logf("sent %d requests in 1s at 100 RPS limit", count)
}

func TestRateLimiter_WaitCancellation(t *testing.T) {
	// 1 RPS,极低
	limiter := NewRateLimiter(1, 1)

	// 第一个 Allow 消耗 token
	limiter.Allow()

	// 接下来 Wait 应该被阻塞
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := limiter.Wait(ctx)
	if err == nil {
		t.Error("expected context deadline error")
	}
}