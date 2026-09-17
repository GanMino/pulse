package engine

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Local 是 MVP 阶段的压测引擎实现
// 实现 Engine 接口,运行在单台机器上
type Local struct {
	config Config

	// 运行时状态
	mu      sync.RWMutex
	status  Status
	run     *Run
	startAt time.Time

	// 内部组件
	scheduler *Scheduler
	client    *http.Client
	resolver  *VariableResolver
	limiter   RateLimiter
	collector *Collector

	// Channels
	metricsCh chan *MetricSnapshot
	eventsCh  chan *Event
	stopCh    chan struct{}
	doneCh    chan struct{}

	// 暂停标志
	pauseMu   sync.Mutex
	paused    bool

	// 最终结果
	resultMu sync.Mutex
	result   *RunSummary
}

// NewLocal 创建 Local Engine
func NewLocal(config Config) *Local {
	if config.MaxVUs <= 0 {
		config.MaxVUs = 20000
	}
	if config.SnapshotInterval == 0 {
		config.SnapshotInterval = time.Second
	}
	return &Local{
		config:    config,
		status:    StatusIdle,
		metricsCh: make(chan *MetricSnapshot, 100),
		eventsCh:  make(chan *Event, 100),
		stopCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
}

// Start 启动压测
func (l *Local) Start(ctx context.Context, run *Run) error {
	l.mu.Lock()
	if l.status != StatusIdle {
		l.mu.Unlock()
		return fmt.Errorf("engine already running (status=%s)", l.status)
	}
	l.status = StatusStarting
	l.run = run
	l.startAt = time.Now()
	l.mu.Unlock()

	l.emitEvent(EventTestStarted, map[string]interface{}{
		"runId":    run.ID,
		"scenario": run.Scenario.Name,
		"vus":      run.Scenario.Load.VUs,
		"duration": run.Scenario.Load.Duration.String(),
	})

	// 准备 HTTP 客户端
	l.client = NewHTTPClient(l.config)

	// 准备变量解析器
	l.resolver = NewVariableResolver(run.Scenario.Vars)

	// 准备限流器
	targetRPS := run.Scenario.Load.TargetRPS
	if run.Runtime != nil && run.Runtime.TargetRPS > 0 {
		targetRPS = run.Runtime.TargetRPS
	}
	if targetRPS > 0 {
		l.limiter = NewRateLimiter(targetRPS, targetRPS)
	}

	// 准备 Collector
	l.collector = NewCollector(10)

	// 准备 Scheduler
	vus := run.Scenario.Load.VUs
	duration2 := run.Scenario.Load.Duration
	if run.Runtime != nil {
		if run.Runtime.VUs > 0 {
			vus = run.Runtime.VUs
		}
		if run.Runtime.Duration > 0 {
			duration2 = run.Runtime.Duration
		}
	}

	loadCfg := LoadConfig{
		VUs:       vus,
		Duration:  duration2,
		TargetRPS: targetRPS,
		RampUp:    run.Scenario.Load.RampUp,
		ThinkTime: run.Scenario.Load.ThinkTime,
	}

	l.scheduler = NewScheduler(vus, loadCfg, l)

	// 启动调度器(在 goroutine 中)
	go l.schedulerEntry(ctx)

	// 启动指标快照循环
	go l.snapshotLoop(ctx)

	// 等待完成
	go l.waitForCompletion(ctx)

	return nil
}

// schedulerEntry 调度器入口
func (l *Local) schedulerEntry(ctx context.Context) {
	l.setStatus(StatusRunning)

	if err := l.scheduler.Run(ctx); err != nil {
		l.emitEvent(EventTestFailed, map[string]interface{}{"error": err.Error()})
		l.setStatus(StatusFailed)
		return
	}

	// 检查是否被中止
	l.mu.RLock()
	stopped := l.status == StatusStopping || l.status == StatusAborted
	l.mu.RUnlock()

	if stopped {
		l.setStatus(StatusAborted)
		l.emitEvent(EventTestStopped, nil)
	} else {
		l.setStatus(StatusCompleted)
		l.emitEvent(EventTestCompleted, nil)
	}
}

// waitForCompletion 等待测试完成,生成最终汇总
func (l *Local) waitForCompletion(ctx context.Context) {
	<-l.scheduler.Done()

	// 生成 RunSummary
	summary := l.collectSummary()
	l.resultMu.Lock()
	l.result = summary
	l.resultMu.Unlock()

	close(l.doneCh)
}

// snapshotLoop 每秒生成指标快照
func (l *Local) snapshotLoop(ctx context.Context) {
	ticker := time.NewTicker(l.config.SnapshotInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-l.stopCh:
			return
		case <-ticker.C:
			snapshot := l.collectSnapshot()
			select {
			case l.metricsCh <- snapshot:
			default:
				// channel 满了,丢弃(避免阻塞)
			}
		}
	}
}

// collectSnapshot 收集当前指标快照
func (l *Local) collectSnapshot() *MetricSnapshot {
	l.mu.RLock()
	run := l.run
	vusTarget := 0
	if run != nil {
		vusTarget = run.Scenario.Load.VUs
	}
	l.mu.RUnlock()

	elapsedMs := time.Since(l.startAt).Milliseconds()
	rps := l.collector.RPS()
	histSnap := l.collector.HistogramSnapshot()

	perReq := make(map[string]*RequestSnapshot)
	for name, stat := range l.collector.PerRequestSnapshot() {
		perReq[name] = &RequestSnapshot{
			Name:    stat.Name,
			Count:   stat.Count,
			Errors:  stat.Errors,
			RPS:     rps,
			LatencyMs: LatencyMetrics{
				Min: stat.Latency.Min,
				Avg: stat.Latency.Mean,
				Max: stat.Latency.Max,
				P50: stat.Latency.P50,
				P90: stat.Latency.P90,
				P95: stat.Latency.P95,
				P99: stat.Latency.P99,
			},
		}
	}

	return &MetricSnapshot{
		Timestamp:  time.Now(),
		ElapsedMs:  elapsedMs,
		VUsActive:  int(l.scheduler.ActiveVUs()),
		VUsTarget:  vusTarget,
		RPS:        rps,
		TotalReq:   l.collector.TotalRequests(),
		TotalErr:   l.collector.TotalErrors(),
		ErrorRate:  l.collector.ErrorRate(),
		LatencyMs: LatencyMetrics{
			Min:   histSnap.Min,
			Avg:   histSnap.Mean,
			Max:   histSnap.Max,
			P50:   histSnap.P50,
			P90:   histSnap.P90,
			P95:   histSnap.P95,
			P99:   histSnap.P99,
		},
		StatusCodes: l.collector.StatusCodes(),
		PerRequest:  perReq,
	}
}

// collectSummary 收集最终汇总
func (l *Local) collectSummary() *RunSummary {
	histSnap := l.collector.HistogramSnapshot()
	totalReqs := l.collector.TotalRequests()

	perReq := make(map[string]*RequestSnapshot)
	for name, stat := range l.collector.PerRequestSnapshot() {
		perReq[name] = &RequestSnapshot{
			Name:    stat.Name,
			Count:   stat.Count,
			Errors:  stat.Errors,
			LatencyMs: LatencyMetrics{
				Min: stat.Latency.Min,
				Avg: stat.Latency.Mean,
				Max: stat.Latency.Max,
				P50: stat.Latency.P50,
				P90: stat.Latency.P90,
				P95: stat.Latency.P95,
				P99: stat.Latency.P99,
			},
		}
	}

	durationMs := time.Since(l.startAt).Milliseconds()
	avgRPS := 0.0
	if durationMs > 0 {
		avgRPS = float64(totalReqs) * 1000.0 / float64(durationMs)
	}

	return &RunSummary{
		RunID:         l.run.ID,
		TotalRequests: totalReqs,
		TotalErrors:   l.collector.TotalErrors(),
		ErrorRate:     l.collector.ErrorRate(),
		DurationMs:    durationMs,
		AvgRPS:        avgRPS,
		LatencyMs: LatencyMetrics{
			Min:   histSnap.Min,
			Avg:   histSnap.Mean,
			Max:   histSnap.Max,
			P50:   histSnap.P50,
			P90:   histSnap.P90,
			P95:   histSnap.P95,
			P99:   histSnap.P99,
		},
		StatusCodes: l.collector.StatusCodes(),
		PerRequest:  perReq,
	}
}

// ============================================
// Engine 接口实现
// ============================================

// Pause 暂停
func (l *Local) Pause() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.status != StatusRunning {
		return fmt.Errorf("cannot pause in status %s", l.status)
	}
	l.status = StatusPaused
	l.scheduler.Pause()
	l.setPaused(true)
	l.emitEvent(EventTestPaused, nil)
	return nil
}

// Resume 恢复
func (l *Local) Resume() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.status != StatusPaused {
		return fmt.Errorf("cannot resume in status %s", l.status)
	}
	l.status = StatusRunning
	l.scheduler.Resume()
	l.setPaused(false)
	l.emitEvent(EventTestResumed, nil)
	return nil
}

// Stop 停止
func (l *Local) Stop() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.status == StatusIdle || l.status == StatusCompleted ||
		l.status == StatusAborted || l.status == StatusFailed {
		return fmt.Errorf("cannot stop in status %s", l.status)
	}
	l.status = StatusStopping
	l.scheduler.Stop()
	// 通知 stopCh(幂等)
	select {
	case <-l.stopCh:
		// 已经关闭
	default:
		close(l.stopCh)
	}
	return nil
}

// Status 获取状态
func (l *Local) Status() Status {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.status
}

// Metrics 获取指标 channel
func (l *Local) Metrics() <-chan *MetricSnapshot {
	return l.metricsCh
}

// Events 获取事件 channel
func (l *Local) Events() <-chan *Event {
	return l.eventsCh
}

// Wait 等待完成,返回汇总
func (l *Local) Wait() (*RunSummary, error) {
	<-l.doneCh
	l.resultMu.Lock()
	defer l.resultMu.Unlock()
	if l.result == nil {
		return nil, fmt.Errorf("no result available")
	}
	return l.result, nil
}

// ============================================
// 内部方法
// ============================================

func (l *Local) setStatus(s Status) {
	l.mu.Lock()
	l.status = s
	l.mu.Unlock()
}

func (l *Local) setPaused(p bool) {
	l.pauseMu.Lock()
	l.paused = p
	l.pauseMu.Unlock()
}

func (l *Local) isPaused() bool {
	l.pauseMu.Lock()
	defer l.pauseMu.Unlock()
	return l.paused
}

func (l *Local) emitEvent(eventType string, payload map[string]interface{}) {
	if payload == nil {
		payload = make(map[string]interface{})
	}
	payload["timestamp"] = time.Now()

	select {
	case l.eventsCh <- &Event{Type: eventType, Payload: payload}:
	default:
		// channel 满了,丢弃
	}
}

// ============================================
// VUCallback 接口实现
// ============================================

// OnVUStart 启动 VU 的工作循环
func (l *Local) OnVUStart(vuID int) error {
	run := l.run
	if run == nil {
		return fmt.Errorf("no run set")
	}
	if l.client == nil {
		return fmt.Errorf("http client not initialized")
	}

	vu := NewVU(vuID, run.Scenario, l.client, l.resolver, l.limiter, l.collector, l.isPaused)

	// 创建一个可取消的 context
	vuCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听 stopCh
	go func() {
		select {
		case <-l.stopCh:
			cancel()
		case <-vuCtx.Done():
		}
	}()

	// 运行 VU 主循环
	_ = vu.Run(vuCtx)
	return nil
}

// OnVUStop VU 停止时的回调
func (l *Local) OnVUStop(vuID int) {
	// 清理工作(可以加日志等)
}

// 编译期断言:Local 实现了 Engine 接口
var _ Engine = (*Local)(nil)