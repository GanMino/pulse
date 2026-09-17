package metrics

import (
	"sync"
	"time"
)

// SlidingWindow 是一个简单的滑动窗口,用于 RPS 计算
// 它在 window 时间内保留每个请求的时间戳,然后统计 RPS
type SlidingWindow struct {
	window  time.Duration
	buckets int
	mu      sync.RWMutex
	ring    []bucket
	head    int
	tail    int
	size    int
}

type bucket struct {
	startTime time.Time
	count     int64
}

// NewSlidingWindow 创建新的滑动窗口
func NewSlidingWindow(window time.Duration) *SlidingWindow {
	const bucketSize = 100 * time.Millisecond
	buckets := int(window / bucketSize)
	if buckets < 1 {
		buckets = 1
	}
	return &SlidingWindow{
		window:  window,
		buckets: buckets,
		ring:    make([]bucket, buckets),
	}
}

// Add 记录一次事件
func (s *SlidingWindow) Add(t time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 计算应该放入哪个 bucket
	bucketIdx := int(t.UnixMilli()/100%int64(s.buckets)) % s.buckets
	b := &s.ring[bucketIdx]

	// 如果 bucket 已过期(早于 window),重置
	if !b.startTime.IsZero() && t.Sub(b.startTime) > s.window {
		b.count = 0
		s.size--
	}

	b.startTime = t.Truncate(100 * time.Millisecond)
	b.count++
	if s.size < s.buckets {
		s.size++
	}
}

// RPS 计算最近 window 内的 RPS
func (s *SlidingWindow) RPS() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-s.window)

	total := int64(0)
	for i := 0; i < s.buckets; i++ {
		b := s.ring[i]
		if !b.startTime.IsZero() && b.startTime.After(cutoff) {
			total += b.count
		}
	}

	return float64(total) / s.window.Seconds()
}

// Count 获取窗口内事件总数
func (s *SlidingWindow) Count() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-s.window)
	total := int64(0)
	for i := 0; i < s.buckets; i++ {
		b := s.ring[i]
		if !b.startTime.IsZero() && b.startTime.After(cutoff) {
			total += b.count
		}
	}
	return total
}