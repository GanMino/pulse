package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Collector 是聚合所有指标的收集器
// 它维护全局统计 + 每个请求的单独统计
type Collector struct {
	// 全局计数(原子)
	totalRequests atomic.Int64
	totalErrors   atomic.Int64
	totalBytesIn  atomic.Int64
	totalBytesOut atomic.Int64
	startedAt     time.Time

	// 全局延迟直方图
	histogram *HistogramWrapper

	// 状态码计数
	statusCodes sync.Map // map[int]int64

	// 每个请求的单独 collector
	perRequest sync.Map // map[string]*RequestCollector

	// RPS 计算(滑动窗口)
	windowSeconds int
	requestTimes  *SlidingWindow
}

// RequestCollector 是单个请求的指标收集器
type RequestCollector struct {
	Name       string
	totalCount atomic.Int64
	totalErr   atomic.Int64
	histogram  *HistogramWrapper
}

// NewCollector 创建全局 Collector
func NewCollector(windowSeconds int) *Collector {
	if windowSeconds <= 0 {
		windowSeconds = 10
	}
	return &Collector{
		histogram:     NewHistogram(60_000_000, 3),
		windowSeconds: windowSeconds,
		requestTimes:  NewSlidingWindow(time.Duration(windowSeconds) * time.Second),
		startedAt:     time.Now(),
	}
}

// ForRequest 获取或创建单个请求的 collector
func (c *Collector) ForRequest(reqName string) *RequestCollector {
	if v, ok := c.perRequest.Load(reqName); ok {
		return v.(*RequestCollector)
	}
	rc := &RequestCollector{
		Name:      reqName,
		histogram: NewHistogram(60_000_000, 3),
	}
	actual, _ := c.perRequest.LoadOrStore(reqName, rc)
	return actual.(*RequestCollector)
}

// Record 记录一次请求完成
func (c *Collector) Record(reqName string, statusCode int, latencyUs int64, bytesIn, bytesOut int64, isError bool) {
	// 全局
	c.totalRequests.Add(1)
	if isError {
		c.totalErrors.Add(1)
	}
	c.totalBytesIn.Add(bytesIn)
	c.totalBytesOut.Add(bytesOut)
	c.histogram.Record(latencyUs)
	c.requestTimes.Add(time.Now())

	// 状态码
	val, _ := c.statusCodes.LoadOrStore(statusCode, new(atomic.Int64))
	val.(*atomic.Int64).Add(1)

	// 单请求
	rc := c.ForRequest(reqName)
	rc.totalCount.Add(1)
	if isError {
		rc.totalErr.Add(1)
	}
	rc.histogram.Record(latencyUs)
}

// RPS 计算过去 windowSeconds 秒内的 RPS
func (c *Collector) RPS() float64 {
	return c.requestTimes.RPS()
}

// TotalRequests 总请求数
func (c *Collector) TotalRequests() int64 {
	return c.totalRequests.Load()
}

// TotalErrors 总错误数
func (c *Collector) TotalErrors() int64 {
	return c.totalErrors.Load()
}

// ErrorRate 错误率
func (c *Collector) ErrorRate() float64 {
	total := c.totalRequests.Load()
	if total == 0 {
		return 0
	}
	return float64(c.totalErrors.Load()) / float64(total)
}

// Elapsed 运行时长
func (c *Collector) Elapsed() time.Duration {
	return time.Since(c.startedAt)
}

// HistogramSnapshot 全局延迟快照
func (c *Collector) HistogramSnapshot() HistogramSnapshot {
	return c.histogram.Snapshot()
}

// StatusCodes 状态码分布(快照)
func (c *Collector) StatusCodes() map[string]int64 {
	result := make(map[string]int64)
	c.statusCodes.Range(func(k, v interface{}) bool {
		code := k.(int)
		count := v.(*atomic.Int64).Load()
		result[itoa(code)] = count
		return true
	})
	return result
}

// PerRequestSnapshot 所有单请求的快照
func (c *Collector) PerRequestSnapshot() map[string]RequestStats {
	result := make(map[string]RequestStats)
	c.perRequest.Range(func(k, v interface{}) bool {
		name := k.(string)
		rc := v.(*RequestCollector)
		snap := rc.histogram.Snapshot()
		result[name] = RequestStats{
			Name:    name,
			Count:   rc.totalCount.Load(),
			Errors:  rc.totalErr.Load(),
			Latency: snap,
		}
		return true
	})
	return result
}

// RequestStats 单个请求的统计
type RequestStats struct {
	Name    string            `json:"name"`
	Count   int64             `json:"count"`
	Errors  int64             `json:"errors"`
	Latency HistogramSnapshot `json:"latency"`
}

// itoa 是简单的 int → string
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	negative := i < 0
	if negative {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if negative {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}