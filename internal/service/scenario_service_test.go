package service

import (
	"context"
	"strings"
	"testing"

	"github.com/pulse/pulse/internal/config"
	"github.com/pulse/pulse/internal/model"
	"github.com/pulse/pulse/internal/repository/sqlite"
)

func setupTestContainer(t *testing.T) *Container {
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

func TestScenarioService_SaveNewScenario(t *testing.T) {
	c := setupTestContainer(t)
	ctx := context.Background()

	req := SaveScenarioRequest{
		Name:        "测试场景",
		Description: "登录压测",
		Config: &model.ScenarioConfig{
			Load: model.LoadConfig{
				VUs:      50,
				Duration: "2m",
			},
			Requests: []model.RequestConfig{
				{Name: "Login", Method: "POST", URL: "/api/login"},
			},
		},
	}

	dto, err := c.ScenarioService.Save(ctx, req)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if dto.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if dto.UUID == "" {
		t.Error("expected UUID")
	}
	if dto.Name != "测试场景" {
		t.Errorf("name = %q", dto.Name)
	}
	if dto.VUs != 50 {
		t.Errorf("VUs = %d, want 50", dto.VUs)
	}
	if dto.RequestCount != 1 {
		t.Errorf("request count = %d, want 1", dto.RequestCount)
	}
}

func TestScenarioService_SaveExistingUpdates(t *testing.T) {
	c := setupTestContainer(t)
	ctx := context.Background()

	// 先创建
	created, err := c.ScenarioService.Save(ctx, SaveScenarioRequest{
		Name: "原",
		Config: &model.ScenarioConfig{
			Load: model.LoadConfig{VUs: 10, Duration: "30s"},
		},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// 更新
	updated, err := c.ScenarioService.Save(ctx, SaveScenarioRequest{
		ID:   created.ID,
		Name: "改",
		Config: &model.ScenarioConfig{
			Load: model.LoadConfig{VUs: 200, Duration: "10m"},
		},
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.ID != created.ID {
		t.Errorf("ID changed: %d vs %d", updated.ID, created.ID)
	}
	if updated.Version != created.Version+1 {
		t.Errorf("version = %d, want %d", updated.Version, created.Version+1)
	}
	if updated.VUs != 200 {
		t.Errorf("VUs = %d, want 200", updated.VUs)
	}
}

func TestScenarioService_ListWithSearch(t *testing.T) {
	c := setupTestContainer(t)
	ctx := context.Background()

	// 创建多个场景
	for _, name := range []string{"登录", "注册", "支付", "查询"} {
		_, err := c.ScenarioService.Save(ctx, SaveScenarioRequest{
			Name: name,
			Config: &model.ScenarioConfig{
				Load: model.LoadConfig{VUs: 10, Duration: "30s"},
			},
		})
		if err != nil {
			t.Fatalf("save %s: %v", name, err)
		}
	}

	t.Run("ListAll", func(t *testing.T) {
		resp, err := c.ScenarioService.List(ctx, ListScenariosRequest{})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if resp.Total != 4 {
			t.Errorf("total = %d, want 4", resp.Total)
		}
	})

	t.Run("SearchByName", func(t *testing.T) {
		resp, err := c.ScenarioService.List(ctx, ListScenariosRequest{
			Search: "登录",
		})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if resp.Total != 1 {
			t.Errorf("total = %d, want 1", resp.Total)
		}
		if resp.Items[0].Name != "登录" {
			t.Errorf("name = %q", resp.Items[0].Name)
		}
	})

	t.Run("ListWithPagination", func(t *testing.T) {
		resp, err := c.ScenarioService.List(ctx, ListScenariosRequest{
			Limit:  2,
			Offset: 0,
		})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(resp.Items) != 2 {
			t.Errorf("len = %d, want 2", len(resp.Items))
		}
		if resp.Total != 4 {
			t.Errorf("total = %d, want 4", resp.Total)
		}
	})
}

func TestScenarioService_Delete(t *testing.T) {
	c := setupTestContainer(t)
	ctx := context.Background()

	created, err := c.ScenarioService.Save(ctx, SaveScenarioRequest{
		Name: "To Delete",
		Config: &model.ScenarioConfig{
			Load: model.LoadConfig{VUs: 10, Duration: "30s"},
		},
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}

	err = c.ScenarioService.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err = c.ScenarioService.GetByID(ctx, created.ID)
	if err != ErrScenarioNotFound {
		t.Errorf("expected ErrScenarioNotFound, got %v", err)
	}
}

func TestScenarioService_Duplicate(t *testing.T) {
	c := setupTestContainer(t)
	ctx := context.Background()

	original, err := c.ScenarioService.Save(ctx, SaveScenarioRequest{
		Name: "Original",
		Config: &model.ScenarioConfig{
			Load: model.LoadConfig{VUs: 100, Duration: "5m"},
			Requests: []model.RequestConfig{
				{Name: "R1", Method: "GET", URL: "/api/test"},
			},
		},
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}

	dup, err := c.ScenarioService.Duplicate(ctx, original.ID)
	if err != nil {
		t.Fatalf("duplicate: %v", err)
	}
	if dup.ID == original.ID {
		t.Error("expected different ID")
	}
	if dup.Name != "Original (副本)" {
		t.Errorf("name = %q", dup.Name)
	}
	if dup.VUs != original.VUs {
		t.Errorf("VUs not copied: %d vs %d", dup.VUs, original.VUs)
	}
	if dup.Version != 1 {
		t.Errorf("version = %d, want 1", dup.Version)
	}
}

func TestScenarioService_Validation(t *testing.T) {
	c := setupTestContainer(t)
	ctx := context.Background()

	t.Run("EmptyName", func(t *testing.T) {
		_, err := c.ScenarioService.Save(ctx, SaveScenarioRequest{
			Name: "",
			Config: &model.ScenarioConfig{
				Load: model.LoadConfig{VUs: 10, Duration: "30s"},
			},
		})
		if err == nil {
			t.Error("expected error for empty name")
		} else if !strings.Contains(err.Error(), "name is required") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("NilConfig", func(t *testing.T) {
		_, err := c.ScenarioService.Save(ctx, SaveScenarioRequest{
			Name: "test",
		})
		if err == nil {
			t.Error("expected error for nil config")
		}
	})
}