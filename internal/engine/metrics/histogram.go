// Package metrics 实现 Pulse 的指标采集
package metrics

import (
	"sync"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
)

// HistogramWrapper 包装 HDR Histogram,提供微秒级别的延迟记录和分位数计算
type HistogramWrapper struct {
	mu          sync.Mutex
	histogram    *hdrhistogram.Histogram
	rangeMax     int64 // 最大记录值(微秒)
	sigFigs      int   // 有效位数
	count        int64
}

// NewHistogram 创建 Histogram
// rangeMax: 最大记录值(微秒)
// sigFigs: 有效位数(2-5,推荐 3)
func NewHistogram(rangeMax int64, sigFigs int) *HistogramWrapper {
	if rangeMax <= 0 {
		rangeMax = 60_000_000 // 默认 60 秒
	}
	if sigFigs < 1 || sigFigs > 5 {
		sigFigs = 3
	}
	// HDR Histogram 最小值为 1(微秒)
	h := hdrhistogram.New(1, rangeMax, sigFigs)
	return &HistogramWrapper{
		histogram: h,
		rangeMax:  rangeMax,
		sigFigs:   sigFigs,
	}
}

// Record 记录一个延迟值(微秒)
func (r *HistogramWrapper) Record(latencyUs int64) {
	if latencyUs < 1 {
		latencyUs = 1
	}
	if latencyUs > r.rangeMax {
		latencyUs = r.rangeMax // 截断到最大可记录值
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.histogram.RecordValue(latencyUs)
	r.count++
}

// Count 返回总记录数
func (r *HistogramWrapper) Count() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.count
}

// Snapshot 不可变快照
type HistogramSnapshot struct {
	Count int64   `json:"count"`
	Min   float64 `json:"min"` // 微秒
	Max   float64 `json:"max"`
	Mean  float64 `json:"mean"`
	P50   float64 `json:"p50"`
	P75   float64 `json:"p75"`
	P90   float64 `json:"p90"`
	P95   float64 `json:"p95"`
	P99   float64 `json:"p99"`
	P999  float64 `json:"p999"`
	StdDev float64 `json:"stdDev"`
}

// Snapshot 获取当前快照(微秒)
func (r *HistogramWrapper) Snapshot() HistogramSnapshot {
	r.mu.Lock()
	defer r.mu.Unlock()

	return HistogramSnapshot{
		Count: r.count,
		Min:   float64(r.histogram.Min()) / 1000.0, // us → ms
		Max:   float64(r.histogram.Max()) / 1000.0,
		Mean:  float64(r.histogram.Mean()) / 1000.0,
		P50:   float64(r.histogram.ValueAtQuantile(50)) / 1000.0,
		P75:   float64(r.histogram.ValueAtQuantile(75)) / 1000.0,
		P90:   float64(r.histogram.ValueAtQuantile(90)) / 1000.0,
		P95:   float64(r.histogram.ValueAtQuantile(95)) / 1000.0,
		P99:   float64(r.histogram.ValueAtQuantile(99)) / 1000.0,
		P999:  float64(r.histogram.ValueAtQuantile(99.9)) / 1000.0,
		StdDev: float64(r.histogram.StdDev()) / 1000.0,
	}
}

// Reset 清空直方图(用于周期重置,避免长跑测试内存增长)
func (r *HistogramWrapper) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.histogram.Reset()
	r.count = 0
}

// 防止 time 包未使用
var _ = time.Second