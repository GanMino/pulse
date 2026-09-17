package service

import (
	"context"
	"testing"
	"time"

	"github.com/pulse/pulse/internal/config"
	"github.com/pulse/pulse/internal/model"
	"github.com/pulse/pulse/internal/repository/sqlite"
)

func setupRunTestContainer(t *testing.T) *Container {
	t.Helper()
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			SQLitePath: tmpDir + "/test.db",
		},
	}

	db, err := sqlite.Open(cfg)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return NewContainer(db)
}

func TestRunService_StartRun(t *testing.T) {
	c := setupRunTestContainer(t)
	ctx := context.Background()

	// 准备场景
	dto, err := c.ScenarioService.Save(ctx, SaveScenarioRequest{
		Name: "Test Scenario",
		Config: &model.ScenarioConfig{
			Load: model.LoadConfig{
				VUs:      10,
				Duration: "5s",
			},
			Requests: []model.RequestConfig{
				{
					Name:   "Test",
					Method: "GET",
					URL:    "http://127.0.0.1:1/test", // 故意指向无效地址以快速失败
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("save scenario: %v", err)
	}

	// 启动运行
	resp, err := c.RunService.StartRun(ctx, StartRunRequest{
		ScenarioID: dto.ID,
	})
	if err != nil {
		t.Fatalf("start run: %v", err)
	}
	if resp.RunID == 0 {
		t.Error("expected non-zero run ID")
	}
	if resp.ScenarioName != "Test Scenario" {
		t.Errorf("name = %q", resp.ScenarioName)
	}

	// 检查 active 状态
	if !c.RunService.IsActive(resp.RunID) {
		t.Error("run should be active")
	}

	// 等待完成
	time.Sleep(7 * time.Second) // 等压测结束

	// 此时应该不再活跃
	// if c.RunService.IsActive(resp.RunID) {
	// 	t.Error("run should not be active after completion")
	// }

	// 获取 run 详情
	run, err := c.RunService.GetRun(ctx, resp.RunID)
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	if run.Status == model.TestRunStatusRunning {
		t.Error("status should not be running")
	}
}

func TestRunService_PauseResumeStop(t *testing.T) {
	c := setupRunTestContainer(t)
	ctx := context.Background()

	dto, err := c.ScenarioService.Save(ctx, SaveScenarioRequest{
		Name: "Pause Test",
		Config: &model.ScenarioConfig{
			Load: model.LoadConfig{
				VUs:      5,
				Duration: "10s",
			},
			Requests: []model.RequestConfig{
				{
					Name:   "Test",
					Method: "GET",
					URL:    "http://127.0.0.1:1/test",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}

	resp, err := c.RunService.StartRun(ctx, StartRunRequest{ScenarioID: dto.ID})
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	// 等启动
	time.Sleep(500 * time.Millisecond)

	// 暂停
	if err := c.RunService.PauseRun(ctx, resp.RunID); err != nil {
		t.Fatalf("pause: %v", err)
	}

	// 状态检查
	run, _ := c.RunService.GetRun(ctx, resp.RunID)
	if run.Status != model.TestRunStatusPaused {
		t.Errorf("status = %q, want paused", run.Status)
	}

	// 恢复
	if err := c.RunService.ResumeRun(ctx, resp.RunID); err != nil {
		t.Fatalf("resume: %v", err)
	}

	run, _ = c.RunService.GetRun(ctx, resp.RunID)
	if run.Status != model.TestRunStatusRunning {
		t.Errorf("status = %q, want running", run.Status)
	}

	// 停止
	if err := c.RunService.StopRun(ctx, resp.RunID); err != nil {
		t.Fatalf("stop: %v", err)
	}

	run, _ = c.RunService.GetRun(ctx, resp.RunID)
	if run.Status != model.TestRunStatusAborted {
		t.Errorf("status = %q, want aborted", run.Status)
	}
}

func TestRunService_GetRunStatus(t *testing.T) {
	c := setupRunTestContainer(t)
	ctx := context.Background()

	// 不存在的 run
	_, err := c.RunService.GetRunStatus(ctx, 999)
	if err == nil {
		t.Error("expected error for non-existent run")
	}
}

func TestRunService_ConvertScenario(t *testing.T) {
	c := setupRunTestContainer(t)

	// 模拟 model.Scenario 转换
	scenario := &model.Scenario{
		ID:   1,
		Name: "Convert Test",
		Config: model.ScenarioConfig{
			Load: model.LoadConfig{
				VUs:      50,
				Duration: "3m",
				RampUp: &model.RampUpConfig{
					Type:     "linear",
					Duration: "30s",
				},
				ThinkTime: "100ms",
			},
			Requests: []model.RequestConfig{
				{
					Name:   "Request1",
					Method: "POST",
					URL:    "/api/test",
					Body:   map[string]interface{}{"key": "value"},
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Extractors: map[string]string{
						"token": "$.data.token",
					},
					Assertions: model.AssertionConfig{
						Status:       []int{200, 201},
						MaxLatencyMs: 500,
					},
				},
			},
		},
	}

	engineScn := c.RunService.convertScenario(scenario)
	if engineScn.Name != "Convert Test" {
		t.Errorf("name = %q", engineScn.Name)
	}
	if engineScn.Load.VUs != 50 {
		t.Errorf("VUs = %d, want 50", engineScn.Load.VUs)
	}
	if engineScn.Load.Duration != 3*time.Minute {
		t.Errorf("duration = %v, want 3m", engineScn.Load.Duration)
	}
	if len(engineScn.Requests) != 1 {
		t.Errorf("requests = %d, want 1", len(engineScn.Requests))
	}
	if engineScn.Requests[0].Method != "POST" {
		t.Errorf("method = %q", engineScn.Requests[0].Method)
	}
	if engineScn.Load.RampUp == nil {
		t.Error("rampUp should not be nil")
	}
}