package sqlite

import (
	"context"
	"testing"

	"github.com/pulse/pulse/internal/model"
)

func TestScenarioRepository_CRUD(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	repo := NewScenarioRepository(db)
	ctx := context.Background()

	// 准备项目
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	user, err := userRepo.GetOrCreateDefault(ctx)
	if err != nil {
		t.Fatalf("get default user: %v", err)
	}
	project, err := projectRepo.GetOrCreateDefault(ctx, user.ID)
	if err != nil {
		t.Fatalf("get default project: %v", err)
	}

	t.Run("Create", func(t *testing.T) {
		scenario := &model.Scenario{
			ProjectID:   project.ID,
			Name:        "登录流程压测",
			Description: "测试登录和获取用户",
			Config: model.ScenarioConfig{
				Load: model.LoadConfig{
					VUs:      100,
					Duration: "5m",
				},
				Requests: []model.RequestConfig{
					{
						Name:   "Login",
						Method: "POST",
						URL:    "/api/login",
					},
				},
			},
			CreatedBy: user.ID,
		}

		err := repo.Create(ctx, scenario)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if scenario.ID == 0 {
			t.Fatal("expected non-zero ID")
		}
		if scenario.UUID == "" {
			t.Fatal("expected UUID to be set")
		}
		if scenario.Version != 1 {
			t.Errorf("expected version 1, got %d", scenario.Version)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		scenario := &model.Scenario{
			ProjectID: project.ID,
			Name:      "GetByID Test",
			Config:    model.ScenarioConfig{Load: model.LoadConfig{VUs: 50, Duration: "1m"}},
			CreatedBy: user.ID,
		}
		_ = repo.Create(ctx, scenario)

		got, err := repo.GetByID(ctx, scenario.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Name != scenario.Name {
			t.Errorf("name = %q, want %q", got.Name, scenario.Name)
		}
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 99999)
		if err != ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("GetByUUID", func(t *testing.T) {
		scenario := &model.Scenario{
			ProjectID: project.ID,
			Name:      "GetByUUID Test",
			Config:    model.ScenarioConfig{Load: model.LoadConfig{VUs: 10, Duration: "30s"}},
			CreatedBy: user.ID,
		}
		_ = repo.Create(ctx, scenario)

		got, err := repo.GetByUUID(ctx, scenario.UUID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.ID != scenario.ID {
			t.Errorf("ID = %d, want %d", got.ID, scenario.ID)
		}
	})

	t.Run("Update_IncrementsVersion", func(t *testing.T) {
		scenario := &model.Scenario{
			ProjectID: project.ID,
			Name:      "Update Test",
			Config:    model.ScenarioConfig{Load: model.LoadConfig{VUs: 10, Duration: "30s"}},
			CreatedBy: user.ID,
		}
		_ = repo.Create(ctx, scenario)

		originalVersion := scenario.Version
		scenario.Name = "Updated Name"
		err := repo.Update(ctx, scenario)
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if scenario.Version != originalVersion+1 {
			t.Errorf("version = %d, want %d", scenario.Version, originalVersion+1)
		}

		// 验证保存
		got, _ := repo.GetByID(ctx, scenario.ID)
		if got.Name != "Updated Name" {
			t.Errorf("name = %q, want %q", got.Name, "Updated Name")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		scenario := &model.Scenario{
			ProjectID: project.ID,
			Name:      "Delete Test",
			Config:    model.ScenarioConfig{Load: model.LoadConfig{VUs: 10, Duration: "30s"}},
			CreatedBy: user.ID,
		}
		_ = repo.Create(ctx, scenario)

		err := repo.Delete(ctx, scenario.ID)
		if err != nil {
			t.Fatalf("delete: %v", err)
		}

		_, err = repo.GetByID(ctx, scenario.ID)
		if err != ErrNotFound {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})
}

func TestScenarioRepository_List(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	repo := NewScenarioRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	user, _ := userRepo.GetOrCreateDefault(ctx)
	project, _ := projectRepo.GetOrCreateDefault(ctx, user.ID)

	// 创建多个场景
	for i := 0; i < 5; i++ {
		scenario := &model.Scenario{
			ProjectID: project.ID,
			Name:      "Scenario " + string(rune('A'+i)),
			Config:    model.ScenarioConfig{Load: model.LoadConfig{VUs: 10, Duration: "30s"}},
			Tags:      StringArray{"tag1"},
			CreatedBy: user.ID,
		}
		_ = repo.Create(ctx, scenario)
	}

	t.Run("ListAll", func(t *testing.T) {
		scenarios, total, err := repo.List(ctx, ListOptions{ProjectID: project.ID})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if total != 5 {
			t.Errorf("total = %d, want 5", total)
		}
		if len(scenarios) != 5 {
			t.Errorf("len = %d, want 5", len(scenarios))
		}
	})

	t.Run("ListWithSearch", func(t *testing.T) {
		scenarios, _, err := repo.List(ctx, ListOptions{
			ProjectID: project.ID,
			Search:    "Scenario B",
		})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(scenarios) != 1 {
			t.Errorf("len = %d, want 1", len(scenarios))
		}
		if scenarios[0].Name != "Scenario B" {
			t.Errorf("name = %q, want %q", scenarios[0].Name, "Scenario B")
		}
	})

	t.Run("ListWithPagination", func(t *testing.T) {
		scenarios, _, err := repo.List(ctx, ListOptions{
			ProjectID: project.ID,
			Limit:     2,
			Offset:    0,
		})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(scenarios) != 2 {
			t.Errorf("len = %d, want 2", len(scenarios))
		}
	})

	t.Run("ListWithTag", func(t *testing.T) {
		scenarios, _, err := repo.List(ctx, ListOptions{
			ProjectID: project.ID,
			Tag:       "tag1",
		})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(scenarios) != 5 {
			t.Errorf("len = %d, want 5", len(scenarios))
		}
	})
}

func TestScenarioConfig_JSONRoundtrip(t *testing.T) {
	original := model.ScenarioConfig{
		Load: model.LoadConfig{
			VUs:       100,
			Duration:  "5m",
			TargetRPS: 1000,
			RampUp: &model.RampUpConfig{
				Type:     "linear",
				Duration: "30s",
			},
		},
		Requests: []model.RequestConfig{
			{
				Name:   "Login",
				Method: "POST",
				URL:    "/api/login",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: map[string]interface{}{
					"username": "{{user}}",
					"password": "{{pass}}",
				},
				Extractors: map[string]string{
					"token": "$.data.token",
				},
			},
		},
		Variables: map[string]model.VariableValue{
			"user": {Value: "alice"},
			"pass": {Value: "secret123"},
		},
		Thresholds: &model.ThresholdConfig{
			P95Ms:     200,
			ErrorRate: 0.01,
		},
	}

	// 序列化为 JSON
	value2, err := original.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}

	// 反序列化
	var decoded model.ScenarioConfig
	err = decoded.Scan(value2)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	if decoded.Load.VUs != original.Load.VUs {
		t.Errorf("VUs = %d, want %d", decoded.Load.VUs, original.Load.VUs)
	}
	if len(decoded.Requests) != 1 {
		t.Errorf("len(Requests) = %d, want 1", len(decoded.Requests))
	}
	if decoded.Requests[0].URL != "/api/login" {
		t.Errorf("URL = %q", decoded.Requests[0].URL)
	}
}