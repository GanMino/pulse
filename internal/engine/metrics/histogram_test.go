package metrics

import (
	"math/rand"
	"testing"
	"time"
)

func TestHistogram_Basic(t *testing.T) {
	h := NewHistogram(60_000_000, 3) // 60s 范围

	// 记录 100 个值
	for i := 1; i <= 100; i++ {
		h.Record(int64(i * 1000)) // 1ms - 100ms
	}

	snap := h.Snapshot()
	if snap.Count != 100 {
		t.Errorf("count = %d, want 100", snap.Count)
	}
	if snap.Min < 0.5 || snap.Min > 1.5 {
		t.Errorf("min = %f ms, want ~1 ms", snap.Min)
	}
	if snap.Max < 95 || snap.Max > 105 {
		t.Errorf("max = %f ms, want ~100 ms", snap.Max)
	}
	if snap.P50 < 45 || snap.P50 > 55 {
		t.Errorf("p50 = %f ms, want ~50 ms", snap.P50)
	}
}

func TestHistogram_Percentiles(t *testing.T) {
	h := NewHistogram(60_000_000, 3)

	// 1000 个值,正态分布,均值 50ms
	r := rand.New(rand.NewSource(42))
	for i := 0; i < 1000; i++ {
		// 简单正态:均值 50000us,标准差 10000us
		val := int64(50000 + r.NormFloat64()*10000)
		if val < 1 {
			val = 1
		}
		h.Record(val)
	}

	snap := h.Snapshot()
	// P50 应接近 50ms,允许一定误差
	if snap.P50 < 40 || snap.P50 > 60 {
		t.Errorf("p50 = %f ms, want ~50 ms (±10)", snap.P50)
	}
	if snap.P99 > snap.Max {
		t.Errorf("P99 (%f) > Max (%f)", snap.P99, snap.Max)
	}
	t.Logf("Count=%d Min=%.2fms Max=%.2fms P50=%.2fms P95=%.2fms P99=%.2fms",
		snap.Count, snap.Min, snap.Max, snap.P50, snap.P95, snap.P99)
}

func TestHistogram_Reset(t *testing.T) {
	h := NewHistogram(60_000_000, 3)
	h.Record(5000)
	h.Record(10000)

	if h.Count() != 2 {
		t.Errorf("count = %d, want 2", h.Count())
	}

	h.Reset()
	if h.Count() != 0 {
		t.Errorf("after reset count = %d, want 0", h.Count())
	}
}

func TestHistogram_ClampMaxValue(t *testing.T) {
	h := NewHistogram(1_000_000, 3) // 1s 最大值

	// 记录超过最大值
	h.Record(5_000_000) // 5s

	snap := h.Snapshot()
	if snap.Max > 1100 { // 略大于 1s(因为截断)
		t.Errorf("max = %f ms, expected clamp near 1000 ms", snap.Max)
	}
}

func TestHistogram_ConcurrentRecord(t *testing.T) {
	h := NewHistogram(60_000_000, 3)

	done := make(chan struct{})
	for w := 0; w < 10; w++ {
		go func() {
			for i := 0; i < 1000; i++ {
				h.Record(int64(1000 + i))
			}
			done <- struct{}{}
		}()
	}

	for w := 0; w < 10; w++ {
		<-done
	}

	if h.Count() != 10000 {
		t.Errorf("count = %d, want 10000", h.Count())
	}

	// 防止 time 报警告
	_ = time.Millisecond
}