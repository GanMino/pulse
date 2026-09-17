package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pulse/pulse/internal/engine"
	"github.com/pulse/pulse/internal/model"
	"github.com/pulse/pulse/internal/repository/sqlite"
	"gorm.io/gorm"
)

// RunService 管理测试运行的完整生命周期
// 协调 ScenarioRepository, TestRunRepository 和 Engine
type RunService struct {
	scenarioRepo  *sqlite.ScenarioRepository
	projectRepo   *sqlite.ProjectRepository
	testRunRepo   *sqlite.TestRunRepository
	reportRepo    *sqlite.ReportRepository
	userRepo      *sqlite.UserRepository
	scenarioSvc   *ScenarioService

	// 当前活跃的运行
	mu     sync.RWMutex
	active map[int64]*ActiveRun // test_run_id -> ActiveRun
}

// ActiveRun 是正在运行的测试
type ActiveRun struct {
	TestRun  *model.TestRun
	Engine   engine.Engine
	Scenario *model.Scenario
	cancel   context.CancelFunc
}

// NewRunService 创建 RunService
func NewRunService(
	scenarioRepo *sqlite.ScenarioRepository,
	projectRepo *sqlite.ProjectRepository,
	testRunRepo *sqlite.TestRunRepository,
	reportRepo *sqlite.ReportRepository,
	userRepo *sqlite.UserRepository,
	scenarioSvc *ScenarioService,
) *RunService {
	return &RunService{
		scenarioRepo: scenarioRepo,
		projectRepo:  projectRepo,
		testRunRepo:  testRunRepo,
		reportRepo:   reportRepo,
		userRepo:     userRepo,
		scenarioSvc:  scenarioSvc,
		active:       make(map[int64]*ActiveRun),
	}
}

// StartRunRequest 启动测试请求
type StartRunRequest struct {
	ScenarioID int64                  `json:"scenarioId"`
	Runtime    *RuntimeConfig         `json:"runtime,omitempty"` // 可选运行时覆盖
}

// StartRunResponse 启动测试响应
type StartRunResponse struct {
	RunID      int64  `json:"runId"`
	RunUUID    string `json:"uuid"`
	StartedAt  string `json:"startedAt"`
	ScenarioName string `json:"scenarioName"`
}

// StartRun 启动一次压测
func (s *RunService) StartRun(ctx context.Context, req StartRunRequest) (*StartRunResponse, error) {
	// 1. 加载场景
	scenario, err := s.scenarioRepo.GetByID(ctx, req.ScenarioID)
	if err != nil {
		if err == sqlite.ErrNotFound {
			return nil, ErrScenarioNotFound
		}
		return nil, fmt.Errorf("load scenario: %w", err)
	}

	// 2. 创建 TestRun 记录
	user, _ := s.userRepo.GetOrCreateDefault(ctx)

	runtimeCfg := req.Runtime
	if runtimeCfg == nil {
		runtimeCfg = &model.RuntimeConfig{}
	}

	run := &model.TestRun{
		ScenarioID:  scenario.ID,
		ProjectID:   scenario.ProjectID,
		Status:      model.TestRunStatusRunning,
		TriggerType: model.TestRunTriggerManual,
		TriggeredBy: user.ID,
		RuntimeConfig: runtimeCfg,
	}
	now := time.Now()
	run.StartedAt = &now

	if err := s.testRunRepo.Create(ctx, run); err != nil {
		return nil, fmt.Errorf("create test run: %w", err)
	}

	// 3. 创建引擎并启动
	engineConfig := engine.Config{
		MaxVUs:           20000,
		KeepAlive:        true,
		EnableHTTP2:      true,
		DialTimeout:      5 * time.Second,
		ResponseTimeout:  30 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
		SnapshotInterval: time.Second,
	}
	eng := engine.NewLocal(engineConfig)

	// 4. 转换场景为引擎格式
	engineScenario := s.convertScenario(scenario)

	// 5. 应用 runtime 覆盖
	if runtimeCfg.VUs > 0 {
		engineScenario.Load.VUs = runtimeCfg.VUs
	}
	if runtimeCfg.Duration != "" {
		duration, _ := time.ParseDuration(runtimeCfg.Duration)
		if duration > 0 {
			engineScenario.Load.Duration = duration
		}
	}
	if runtimeCfg.TargetRPS > 0 {
		engineScenario.Load.TargetRPS = runtimeCfg.TargetRPS
	}

	// 6. 创建 context(可取消)
	runCtx, cancel := context.WithCancel(context.Background())

	// 7. 启动引擎
	engineRun := &engine.Run{
		ID:       run.ID,
		Scenario: engineScenario,
		Runtime: engine.RuntimeOverride{
			VUs:       engineScenario.Load.VUs,
			Duration:  engineScenario.Load.Duration,
			TargetRPS: engineScenario.Load.TargetRPS,
		},
	}

	if err := eng.Start(runCtx, engineRun); err != nil {
		cancel()
		// 回滚 run 状态
		run.Status = model.TestRunStatusFailed
		run.ErrorMessage = err.Error()
		_ = s.testRunRepo.UpdateResult(ctx, run)
		return nil, fmt.Errorf("start engine: %w", err)
	}

	// 8. 记录活跃运行
	active := &ActiveRun{
		TestRun:  run,
		Engine:   eng,
		Scenario: scenario,
		cancel:   cancel,
	}
	s.mu.Lock()
	s.active[run.ID] = active
	s.mu.Unlock()

	// 9. 异步等待完成 + 更新 DB
	go s.waitForCompletion(run.ID)

	return &StartRunResponse{
		RunID:        run.ID,
		RunUUID:      run.UUID,
		StartedAt:    now.Format(time.RFC3339),
		ScenarioName: scenario.Name,
	}, nil
}

// bridgeEventsToWails 把引擎事件桥接到 Wails Events(可选注入)
// 注意:此方法需要 wails Runtime,我们在 App 启动时注册一次
// 这里用 interface 解耦,避免 service 包依赖 wails
type EventBridge interface {
	Emit(eventName string, data ...interface{})
}

// BridgeEngineEvents 把活跃引擎的事件和指标桥接到 Wails
// 由 App.Startup 调用,作为中间层
func (s *RunService) BridgeEngineEvents(runID int64, bridge EventBridge) {
	go func() {
		s.mu.RLock()
		active, ok := s.active[runID]
		s.mu.RUnlock()
		if !ok {
			return
		}

		// 转发事件
		go func() {
			for event := range active.Engine.Events() {
				bridge.Emit(event.Type+":"+fmt.Sprint(runID), event.Payload)
				if event.Type == engine.EventTestCompleted ||
					event.Type == engine.EventTestFailed ||
					event.Type == engine.EventTestStopped {
					return
				}
			}
		}()

		// 转发指标快照
		go func() {
			for snapshot := range active.Engine.Metrics() {
				bridge.Emit("metric:update:"+fmt.Sprint(runID), snapshot)
			}
		}()
	}()
}

// waitForCompletion 等待引擎完成,更新 DB
func (s *RunService) waitForCompletion(runID int64) {
	s.mu.RLock()
	active, ok := s.active[runID]
	s.mu.RUnlock()
	if !ok {
		return
	}

	// 等待引擎完成
	summary, err := active.Engine.Wait()

	ctx := context.Background()

	// 更新 TestRun 状态
	run := active.TestRun
	now := time.Now()
	run.FinishedAt = &now
	run.DurationMs = now.Sub(*run.StartedAt).Milliseconds()

	if summary != nil {
		run.Summary = &model.RunSummary{
			TotalRequests: summary.TotalRequests,
			TotalErrors:   summary.TotalErrors,
			ErrorRate:     summary.ErrorRate,
			AvgRPS:        summary.AvgRPS,
			PeakRPS:       summary.AvgRPS, // 简化:使用 avgRPS
			P50Ms:         summary.LatencyMs.P50,
			P90Ms:         summary.LatencyMs.P90,
			P95Ms:         summary.LatencyMs.P95,
			P99Ms:         summary.LatencyMs.P99,
			MaxMs:         summary.LatencyMs.Max,
			AvgMs:         summary.LatencyMs.Avg,
			StatusCodes:   summary.StatusCodes,
			PerRequest:    convertPerRequestToRequestStats(summary.PerRequest),
		}
	}

	if err != nil {
		run.Status = model.TestRunStatusFailed
		run.ErrorMessage = err.Error()
	} else {
		// 检查引擎状态
		switch active.Engine.Status() {
		case engine.StatusCompleted:
			run.Status = model.TestRunStatusCompleted
		case engine.StatusAborted:
			run.Status = model.TestRunStatusAborted
		case engine.StatusFailed:
			run.Status = model.TestRunStatusFailed
		default:
			run.Status = model.TestRunStatusCompleted
		}
	}

	// 持久化到 DB
	if err := s.testRunRepo.UpdateResult(ctx, run); err != nil {
		fmt.Printf("[RunService] failed to update test run: %v\n", err)
	}

	// 创建 Report 记录
	report := &model.Report{
		TestRunID: run.ID,
		Summary:   run.Summary,
		Status:    model.ReportStatusReady,
	}
	if err := s.reportRepo.Create(ctx, report); err != nil {
		fmt.Printf("[RunService] failed to create report: %v\n", err)
	}

	// 清理活跃列表
	s.mu.Lock()
	delete(s.active, runID)
	s.mu.Unlock()
}

// convertPerRequestToRequestStats 转换引擎格式为 model 格式
func convertPerRequestToRequestStats(perReq map[string]*engine.RequestSnapshot) map[string]*model.RequestStat {
	result := make(map[string]*model.RequestStat, len(perReq))
	for name, snap := range perReq {
		result[name] = &model.RequestStat{
			Name:   snap.Name,
			Count:  snap.Count,
			Errors: snap.Errors,
			AvgRPS: snap.RPS,
			P95Ms:  snap.LatencyMs.P95,
			P99Ms:  snap.LatencyMs.P99,
		}
	}
	return result
}

// PauseRun 暂停测试
func (s *RunService) PauseRun(ctx context.Context, runID int64) error {
	s.mu.RLock()
	active, ok := s.active[runID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("run %d not active", runID)
	}

	if err := active.Engine.Pause(); err != nil {
		return err
	}

	active.TestRun.Status = model.TestRunStatusPaused
	return s.testRunRepo.UpdateStatus(ctx, runID, model.TestRunStatusPaused)
}

// ResumeRun 恢复测试
func (s *RunService) ResumeRun(ctx context.Context, runID int64) error {
	s.mu.RLock()
	active, ok := s.active[runID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("run %d not active", runID)
	}

	if err := active.Engine.Resume(); err != nil {
		return err
	}

	active.TestRun.Status = model.TestRunStatusRunning
	return s.testRunRepo.UpdateStatus(ctx, runID, model.TestRunStatusRunning)
}

// StopRun 停止测试
func (s *RunService) StopRun(ctx context.Context, runID int64) error {
	s.mu.RLock()
	active, ok := s.active[runID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("run %d not active", runID)
	}

	if err := active.Engine.Stop(); err != nil {
		return err
	}

	active.TestRun.Status = model.TestRunStatusAborted
	return s.testRunRepo.UpdateStatus(ctx, runID, model.TestRunStatusAborted)
}

// GetRun 获取测试运行详情
func (s *RunService) GetRun(ctx context.Context, runID int64) (*model.TestRun, error) {
	return s.testRunRepo.GetByID(ctx, runID)
}

// GetRunStatus 获取测试运行状态
func (s *RunService) GetRunStatus(ctx context.Context, runID int64) (string, error) {
	// 优先从活跃列表获取
	s.mu.RLock()
	if active, ok := s.active[runID]; ok {
		s.mu.RUnlock()
		return string(active.Engine.Status()), nil
	}
	s.mu.RUnlock()

	// 从 DB 获取
	run, err := s.testRunRepo.GetByID(ctx, runID)
	if err != nil {
		return "", err
	}
	return string(run.Status), nil
}

// IsActive 检查 run 是否在活跃列表
func (s *RunService) IsActive(runID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.active[runID]
	return ok
}

// GetActiveEngine 获取活跃引擎(用于前端订阅事件)
func (s *RunService) GetActiveEngine(runID int64) (engine.Engine, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if active, ok := s.active[runID]; ok {
		return active.Engine, true
	}
	return nil, false
}

// convertScenario 将 model.Scenario 转换为 engine.Scenario
func (s *RunService) convertScenario(s *model.Scenario) engine.Scenario {
	result := engine.Scenario{
		ID:   s.ID,
		Name: s.Name,
		Load: engine.LoadConfig{
			VUs:       s.Config.Load.VUs,
			Duration:  parseDuration(s.Config.Load.Duration),
			TargetRPS: s.Config.Load.TargetRPS,
			ThinkTime: parseDuration(s.Config.Load.ThinkTime),
		},
		Vars: s.Config.Variables,
	}

	if s.Config.Load.RampUp != nil {
		result.Load.RampUp = &engine.RampUp{
			Type:     s.Config.Load.RampUp.Type,
			Duration: parseDuration(s.Config.Load.RampUp.Duration),
			Steps:    s.Config.Load.RampUp.Steps,
		}
	}

	// 转换请求
	for _, r := range s.Config.Requests {
		req := engine.Request{
			Name:    r.Name,
			Method:  r.Method,
			URL:     r.URL,
			Headers: r.Headers,
			Body:    r.Body,
			BodyType: "json",
			Extractors: r.Extractors,
			Assertions: engine.Assertions{
				Status:       r.Assertions.Status,
				MaxLatencyMs: r.Assertions.MaxLatencyMs,
				BodyContains: r.Assertions.BodyContains,
			},
			ThinkTime: parseDuration(r.ThinkTime),
			Weight:    r.Weight,
		}
		if r.BodyRaw != "" {
			req.BodyType = "raw"
			req.Body = nil
			// BodyRaw 通过 Headers 或 Body 传递,简化:MVP 中将 bodyRaw 序列化为 JSON 字符串
			req.Body = map[string]interface{}{"_raw": r.BodyRaw}
		}
		result.Requests = append(result.Requests, req)
	}

	return result
}

// parseDuration 解析 "5m" 格式字符串为 time.Duration
func parseDuration(s string) time.Duration {
	if s == "" {
		return 0
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d
}

// 编译期检查
var _ = gorm.ErrRecordNotFound