package main

import (
	"context"
	"fmt"
	"time"

	"github.com/pulse/pulse/internal/config"
	"github.com/pulse/pulse/internal/event"
	"github.com/pulse/pulse/internal/repository/sqlite"
	"github.com/pulse/pulse/internal/service"
	"github.com/pulse/pulse/pkg/logger"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App 是 Wails 应用的主结构体,绑定到前端
type App struct {
	ctx   context.Context
	cfg   *config.Config
	log   *logger.Logger
	bus   *event.Bus
	db    *sqlite.DB
	svc   *service.Container
}

// NewApp 创建新的 App 实例
func NewApp(cfg *config.Config, log *logger.Logger, bus *event.Bus, db *sqlite.DB, svc *service.Container) *App {
	return &App{
		cfg: cfg,
		log: log,
		bus: bus,
		db:  db,
		svc: svc,
	}
}

// Startup 在应用启动时调用
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.log.Info().Msg("Pulse started")

	// 注册默认用户和项目
	if err := a.svc.UserService.EnsureDefaultUserAndProject(ctx); err != nil {
		a.log.Error().Err(err).Msg("failed to ensure default user/project")
	}

	// 发送欢迎事件到前端
	a.emit("app:ready", map[string]interface{}{
		"version":     a.cfg.App.Version,
		"message":     "Welcome to Pulse!",
		"dataDir":     a.cfg.App.DataDir,
		"currentTime": time.Now().Format(time.RFC3339),
	})
}

// Shutdown 在应用关闭时调用
func (a *App) Shutdown(ctx context.Context) {
	a.log.Info().Msg("Pulse shutting down")
	if a.db != nil {
		_ = a.db.Close()
	}
}

// ============================================
// 通用绑定方法
// ============================================

// GetAppInfo 返回应用基本信息
func (a *App) GetAppInfo() map[string]string {
	return map[string]string{
		"name":    a.cfg.App.Name,
		"version": a.cfg.App.Version,
		"dataDir": a.cfg.App.DataDir,
	}
}

// Echo 测试方法 - 用于验证前后端通信
func (a *App) Echo(message string) string {
	a.log.Info().Str("input", message).Msg("Echo called")
	return fmt.Sprintf("Pulse received: %s", message)
}

// GetCurrentUser 获取当前用户
func (a *App) GetCurrentUser() *service.UserDTO {
	user, err := a.svc.UserService.GetCurrentUser(a.ctx)
	if err != nil {
		a.log.Error().Err(err).Msg("GetCurrentUser failed")
		return nil
	}
	return user
}

// ============================================
// Project API 绑定
// ============================================

// ListProjects 列出所有项目
func (a *App) ListProjects() ([]service.ProjectDTO, error) {
	return a.svc.ProjectService.List(a.ctx)
}

// GetOrCreateDefaultProject 获取或创建默认项目
func (a *App) GetOrCreateDefaultProject() (*service.ProjectDTO, error) {
	return a.svc.ProjectService.GetOrCreateDefault(a.ctx)
}

// CreateProject 创建项目
func (a *App) CreateProject(req service.CreateProjectRequest) (*service.ProjectDTO, error) {
	return a.svc.ProjectService.Create(a.ctx, req)
}

// ============================================
// Scenario API 绑定
// ============================================

// ListScenarios 列出场景
func (a *App) ListScenarios(req service.ListScenariosRequest) (*service.ListScenariosResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 50
	}
	return a.svc.ScenarioService.List(a.ctx, req)
}

// GetScenario 获取单个场景详情
func (a *App) GetScenario(id int64) (*service.ScenarioDTO, error) {
	return a.svc.ScenarioService.GetByID(a.ctx, id)
}

// GetScenarioByUUID 根据 UUID 获取场景
func (a *App) GetScenarioByUUID(uuid string) (*service.ScenarioDTO, error) {
	return a.svc.ScenarioService.GetByUUID(a.ctx, uuid)
}

// SaveScenario 保存场景(创建或更新)
func (a *App) SaveScenario(req service.SaveScenarioRequest) (*service.ScenarioDTO, error) {
	return a.svc.ScenarioService.Save(a.ctx, req)
}

// DeleteScenario 删除场景
func (a *App) DeleteScenario(id int64) error {
	return a.svc.ScenarioService.Delete(a.ctx, id)
}

// DeleteScenarios 批量删除
func (a *App) DeleteScenarios(ids []int64) (int, error) {
	return a.svc.ScenarioService.DeleteMany(a.ctx, ids)
}

// DuplicateScenario 复制场景
func (a *App) DuplicateScenario(id int64) (*service.ScenarioDTO, error) {
	return a.svc.ScenarioService.Duplicate(a.ctx, id)
}

// ============================================
// Agent API 绑定
// ============================================

// ListAgents 列出 Agent
func (a *App) ListAgents() ([]service.AgentDTO, error) {
	return a.svc.AgentService.List(a.ctx)
}

// RegisterLocalAgent 注册本地 Agent
func (a *App) RegisterLocalAgent() (*service.AgentDTO, error) {
	return a.svc.AgentService.RegisterLocal(a.ctx)
}

// ============================================
// TestRun API 绑定(预留)
// ============================================

// ListTestRunsByScenario 列出场景的所有测试运行
func (a *App) ListTestRunsByScenario(scenarioID int64, limit int) ([]service.TestRunDTO, error) {
	return a.svc.TestRunService.ListByScenario(a.ctx, scenarioID, limit)
}

// ListAllTestRuns 列出所有测试运行
func (a *App) ListAllTestRuns(limit int) ([]service.TestRunDTO, error) {
	return a.svc.TestRunService.ListAll(a.ctx, limit)
}

// StartRun 启动一次压测
func (a *App) StartRun(req service.StartRunRequest) (*service.StartRunResponse, error) {
	resp, err := a.svc.RunService.StartRun(a.ctx, req)
	if err != nil {
		return nil, err
	}

	// 启动事件桥接 goroutine:把引擎的 Events 和 Metrics 转发到 Wails Runtime
	go a.bridgeRunEvents(resp.RunID)

	return resp, nil
}

// bridgeRunEvents 桥接引擎事件到 Wails Runtime
func (a *App) bridgeRunEvents(runID int64) {
	eng, ok := a.svc.RunService.GetActiveEngine(runID)
	if !ok {
		return
	}

	// 转发生命周期事件
	go func() {
		for event := range eng.Events() {
			eventName := event.Type + ":" + fmt.Sprintf("%d", runID)
			wailsRuntime.EventsEmit(a.ctx, eventName, event.Payload)
			// 终止事件后停止监听
			if event.Type == "test:completed" || event.Type == "test:failed" || event.Type == "test:stopped" {
				return
			}
		}
	}()

	// 转发指标快照
	go func() {
		for snapshot := range eng.Metrics() {
			wailsRuntime.EventsEmit(a.ctx, fmt.Sprintf("metric:update:%d", runID), snapshot)
		}
	}()
}

// PauseRun 暂停压测
func (a *App) PauseRun(runID int64) error {
	return a.svc.RunService.PauseRun(a.ctx, runID)
}

// ResumeRun 恢复压测
func (a *App) ResumeRun(runID int64) error {
	return a.svc.RunService.ResumeRun(a.ctx, runID)
}

// StopRun 停止压测
func (a *App) StopRun(runID int64) error {
	return a.svc.RunService.StopRun(a.ctx, runID)
}

// GetRun 获取测试运行详情
func (a *App) GetRun(runID int64) (*service.TestRunDTO, error) {
	run, err := a.svc.RunService.GetRun(a.ctx, runID)
	if err != nil {
		return nil, err
	}
	scenario, _ := a.svc.ScenarioService.GetByID(a.ctx, run.ScenarioID)
	scenarioName := ""
	if scenario != nil {
		scenarioName = scenario.Name
	}
	return a.svc.TestRunService.toDTO(run, scenarioName), nil
}

// IsRunActive 检查 run 是否在活跃列表
func (a *App) IsRunActive(runID int64) bool {
	return a.svc.RunService.IsActive(runID)
}

// ============================================
// Report API 绑定(预留)
// ============================================

// ListReports 列出报告
func (a *App) ListReports(limit, offset int) ([]service.ReportDTO, int64, error) {
	return a.svc.ReportService.List(a.ctx, limit, offset)
}

// GetReportByTestRunID 根据测试运行 ID 获取报告
func (a *App) GetReportByTestRunID(testRunID int64) (*service.ReportDTO, error) {
	return a.svc.ReportService.GetByTestRunID(a.ctx, testRunID)
}

// ExportReportHTML 导出报告为 HTML
func (a *App) ExportReportHTML(reportID int64, outputDir string) (string, error) {
	return a.svc.ReportService.ExportHTML(a.ctx, reportID, outputDir)
}

// ============================================
// Import API 绑定
// ============================================

// ImportHAR 导入 HAR 文件内容生成场景
func (a *App) ImportHAR(req service.ImportHARRequest) (*service.ImportHARResponse, error) {
	return a.svc.ImportService.ImportHAR(a.ctx, req)
}

// ============================================
// 内部方法
// ============================================

// emit 推送事件到前端
func (a *App) emit(eventName string, data ...interface{}) {
	if a.ctx == nil {
		return
	}
	wailsRuntime.EventsEmit(a.ctx, eventName, data...)
}

// on 监听前端事件
func (a *App) on(eventName string, callback func(optionalData ...interface{})) {
	if a.ctx == nil {
		return
	}
	wailsRuntime.EventsOn(a.ctx, eventName, callback)
}