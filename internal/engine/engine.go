// Package engine 实现 Pulse 自研的压测引擎
package engine

import (
	"context"
	"time"
)

// Engine 是压测引擎的抽象接口
// MVP 阶段只实现 Local(单机),后续可扩展 Remote(分布式)
type Engine interface {
	// Start 启动压测任务
	Start(ctx context.Context, run *Run) error

	// Pause 暂停压测(保留 VU pool,停止新请求)
	Pause() error

	// Resume 恢复压测
	Resume() error

	// Stop 优雅停止压测(最多等待 5s)
	Stop() error

	// Status 获取当前状态
	Status() Status

	// Metrics 获取实时指标 channel
	Metrics() <-chan *MetricSnapshot

	// Events 获取事件 channel(test:started, test:completed 等)
	Events() <-chan *Event

	// Summary 阻塞等待测试完成,返回汇总结果
	Wait() (*RunSummary, error)
}

// Status 是引擎状态
type Status string

const (
	StatusIdle      Status = "idle"
	StatusStarting  Status = "starting"
	StatusRunning   Status = "running"
	StatusPaused    Status = "paused"
	StatusStopping  Status = "stopping"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusAborted   Status = "aborted"
)

// Run 是单次测试运行的执行上下文
type Run struct {
	ID       int64
	Scenario Scenario
	Runtime  RuntimeOverride
}

// Scenario 是运行所需的核心配置(精简版)
type Scenario struct {
	ID       int64
	Name     string
	Requests []Request
	Load     LoadConfig
	Vars     map[string][]string
}

// Request 是单个 HTTP 请求定义
type Request struct {
	Name       string
	Method     string
	URL        string
	Headers    map[string]string
	Body       []byte
	BodyRaw    string
	BodyType   string // json / form / raw
	Extractors map[string]string
	Assertions Assertions
	ThinkTime  time.Duration
	Weight     int
}

// LoadConfig 压测负载
type LoadConfig struct {
	VUs       int
	Duration  time.Duration
	TargetRPS int // 0 = 不限
	RampUp    *RampUp
	ThinkTime time.Duration
	Iterations int
}

// RampUp 渐进式启动
type RampUp struct {
	Type     string        // linear / step / wave
	Duration time.Duration
	Steps    int
}

// RuntimeOverride 运行时参数覆盖
type RuntimeOverride struct {
	VUs       int
	Duration  time.Duration
	TargetRPS int
}

// Assertions 请求断言
type Assertions struct {
	Status        []int
	MaxLatencyMs  int
	BodyContains  []string
	HeaderCheck   map[string]string
}

// Event 是引擎发出的事件
type Event struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// Event 类型常量
const (
	EventTestStarted    = "test:started"
	EventTestPaused     = "test:paused"
	EventTestResumed    = "test:resumed"
	EventTestStopped    = "test:stopped"
	EventTestCompleted  = "test:completed"
	EventTestFailed     = "test:failed"
	EventRequestSent    = "request:sent"
	EventRequestDone    = "request:done"
	EventRequestError   = "request:error"
	EventRampUpProgress = "rampup:progress"
)

// MetricSnapshot 是一次指标快照(用于实时大屏)
type MetricSnapshot struct {
	Timestamp   time.Time                  `json:"timestamp"`
	ElapsedMs   int64                       `json:"elapsedMs"`
	VUsActive   int                         `json:"vusActive"`
	VUsTarget   int                         `json:"vusTarget"`
	RPS         float64                     `json:"rps"`
	TotalReq    int64                       `json:"totalRequests"`
	TotalErr    int64                       `json:"totalErrors"`
	ErrorRate   float64                     `json:"errorRate"`
	LatencyMs    LatencyMetrics              `json:"latencyMs"`
	StatusCodes map[string]int64            `json:"statusCodes"`
	PerRequest  map[string]*RequestSnapshot `json:"perRequest,omitempty"`
}

// LatencyMetrics 延迟统计
type LatencyMetrics struct {
	Min    float64 `json:"min"`
	Avg    float64 `json:"avg"`
	Max    float64 `json:"max"`
	P50    float64 `json:"p50"`
	P90    float64 `json:"p90"`
	P95    float64 `json:"p95"`
	P99    float64 `json:"p99"`
}

// RequestSnapshot 单个请求的指标快照
type RequestSnapshot struct {
	Name       string        `json:"name"`
	Count      int64         `json:"count"`
	Errors     int64         `json:"errors"`
	RPS        float64       `json:"rps"`
	LatencyMs  LatencyMetrics `json:"latencyMs"`
}

// RunSummary 是测试完成后的汇总
type RunSummary struct {
	RunID            int64                 `json:"runId"`
	TotalRequests    int64                 `json:"totalRequests"`
	TotalErrors      int64                 `json:"totalErrors"`
	ErrorRate        float64               `json:"errorRate"`
	DurationMs       int64                 `json:"durationMs"`
	AvgRPS           float64               `json:"avgRPS"`
	PeakRPS          float64               `json:"peakRPS"`
	LatencyMs        LatencyMetrics        `json:"latencyMs"`
	StatusCodes      map[string]int64      `json:"statusCodes"`
	PerRequest       map[string]*RequestSnapshot `json:"perRequest"`
	ThresholdResults []ThresholdResult     `json:"thresholdResults,omitempty"`
}

// ThresholdResult 阈值检查结果
type ThresholdResult struct {
	Name     string  `json:"name"`
	Expected float64 `json:"expected"`
	Actual   float64 `json:"actual"`
	Passed   bool    `json:"passed"`
}

// Config 是引擎的配置
type Config struct {
	MaxVUs           int
	KeepAlive        bool
	EnableHTTP2      bool
	DialTimeout      time.Duration
	ResponseTimeout  time.Duration
	TLSHandshakeTimeout time.Duration
	SnapshotInterval time.Duration
	HistogramRange   int64  // 微秒
	HistogramSigFigs int    // 有效位数
}