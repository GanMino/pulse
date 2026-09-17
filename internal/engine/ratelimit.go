package engine

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter 是 RPS 限流器接口
type RateLimiter interface {
	// Wait 阻塞直到可以放行(会考虑 burst)
	Wait(ctx context.Context) error
	// Allow 非阻塞,尝试获取一个 token
	Allow() bool
}

// tokenBucketRateLimiter 基于 token bucket 的限流器
type tokenBucketRateLimiter struct {
	limiter *rate.Limiter
}

// NewRateLimiter 创建 RPS 限流器
// 参数:rps 每秒允许的请求数,burst 突发容量
func NewRateLimiter(rps, burst int) RateLimiter {
	if rps <= 0 {
		return &noopRateLimiter{} // 不限流
	}
	if burst <= 0 {
		burst = rps // 默认 burst = rps
	}
	return &tokenBucketRateLimiter{
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
	}
}

func (l *tokenBucketRateLimiter) Wait(ctx context.Context) error {
	return l.limiter.Wait(ctx)
}

func (l *tokenBucketRateLimiter) Allow() bool {
	return l.limiter.Allow()
}

// noopRateLimiter 不限流(用于 rps=0 的情况)
type noopRateLimiter struct{}

func (l *noopRateLimiter) Wait(_ context.Context) error { return nil }
func (l *noopRateLimiter) Allow() bool                  { return true }

// RateLimiterFromTarget 根据目标 RPS 创建对应 limiter
// 如果 rps <= 0,返回 noop
func RateLimiterFromTarget(rps int) RateLimiter {
	return NewRateLimiter(rps, rps)
}

// EnsureCompileTimeInterface 实现保证(避免误改接口)
var _ RateLimiter = (*tokenBucketRateLimiter)(nil)
var _ RateLimiter = (*noopRateLimiter)(nil)

// 防止 time 包未使用(预留扩展)
var _ = time.Second