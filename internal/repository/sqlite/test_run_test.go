package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/pulse/pulse/internal/model"
)

func TestTestRunRepository_FullLifecycle(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	scenarioRepo := NewScenarioRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	runRepo := NewTestRunRepository(db)
	ctx := context.Background()

	user, _ := userRepo.GetOrCreateDefault(ctx)
	project, _ := projectRepo.GetOrCreateDefault(ctx, user.ID)
	scenario := &model.Scenario{
		ProjectID: project.ID,
		Name:      "Test Run Scenario",
		Config:    model.ScenarioConfig{Load: model.LoadConfig{VUs: 10, Duration: "30s"}},
		CreatedBy: user.ID,
	}
	_ = scenarioRepo.Create(ctx, scenario)

	t.Run("Create Run", func(t *testing.T) {
		run := &model.TestRun{
			ScenarioID: scenario.ID,
			ProjectID:  project.ID,
			Status:     model.TestRunStatusPending,
		}
		err := runRepo.Create(ctx, run)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if run.ID == 0 {
			t.Error("expected non-zero ID")
		}
	})

	t.Run("Update Status", func(t *testing.T) {
		run := &model.TestRun{
			ScenarioID: scenario.ID,
			ProjectID:  project.ID,
			Status:     model.TestRunStatusPending,
		}
		_ = runRepo.Create(ctx, run)

		err := runRepo.UpdateStatus(ctx, run.ID, model.TestRunStatusRunning)
		if err != nil {
			t.Fatalf("update status: %v", err)
		}

		got, _ := runRepo.GetByID(ctx, run.ID)
		if got.Status != model.TestRunStatusRunning {
			t.Errorf("status = %q, want running", got.Status)
		}
	})

	t.Run("Update Result with Summary", func(t *testing.T) {
		run := &model.TestRun{
			ScenarioID: scenario.ID,
			ProjectID:  project.ID,
		}
		_ = runRepo.Create(ctx, run)

		now := time.Now()
		run.StartedAt = &now
		finished := now.Add(5 * time.Minute)
		run.FinishedAt = &finished
		run.DurationMs = 300000
		run.Status = model.TestRunStatusCompleted
		run.Summary = &model.RunSummary{
			TotalRequests: 10000,
			TotalErrors:   50,
			ErrorRate:     0.005,
			AvgRPS:        33.3,
			P95Ms:         156.7,
			StatusCodes:   map[string]int64{"200": 9900, "500": 50},
		}

		err := runRepo.UpdateResult(ctx, run)
		if err != nil {
			t.Fatalf("update result: %v", err)
		}

		got, _ := runRepo.GetByID(ctx, run.ID)
		if got.Summary == nil {
			t.Fatal("summary should not be nil")
		}
		if got.Summary.TotalRequests != 10000 {
			t.Errorf("TotalRequests = %d, want 10000", got.Summary.TotalRequests)
		}
		if got.Status != model.TestRunStatusCompleted {
			t.Errorf("status = %q, want completed", got.Status)
		}
	})

	t.Run("ListByScenario", func(t *testing.T) {
		// 创建 3 个 runs
		for i := 0; i < 3; i++ {
			run := &model.TestRun{
				ScenarioID: scenario.ID,
				ProjectID:  project.ID,
			}
			_ = runRepo.Create(ctx, run)
		}

		runs, err := runRepo.ListByScenario(ctx, scenario.ID, 10)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(runs) < 3 {
			t.Errorf("len = %d, want >= 3", len(runs))
		}
	})
}