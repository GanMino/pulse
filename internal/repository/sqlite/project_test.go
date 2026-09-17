package sqlite

import (
	"context"
	"testing"

	"github.com/pulse/pulse/internal/model"
)

func TestProjectRepository_GetOrCreateDefault(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	repo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	user, err := userRepo.GetOrCreateDefault(ctx)
	if err != nil {
		t.Fatalf("get default user: %v", err)
	}

	t.Run("First call creates", func(t *testing.T) {
		project, err := repo.GetOrCreateDefault(ctx, user.ID)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if project.Name != "Default" {
			t.Errorf("name = %q, want Default", project.Name)
		}
		if project.UUID == "" {
			t.Error("UUID should be set")
		}
	})

	t.Run("Second call returns existing", func(t *testing.T) {
		p1, _ := repo.GetOrCreateDefault(ctx, user.ID)
		p2, _ := repo.GetOrCreateDefault(ctx, user.ID)
		if p1.ID != p2.ID {
			t.Errorf("expected same ID, got %d vs %d", p1.ID, p2.ID)
		}
	})
}

func TestProjectRepository_List(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	repo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()
	user, _ := userRepo.GetOrCreateDefault(ctx)

	// 创建 3 个项目
	for i := 0; i < 3; i++ {
		p := &model.Project{
			Name:    "Project " + string(rune('A'+i)),
			OwnerID: user.ID,
		}
		_ = repo.Create(ctx, p)
	}

	projects, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(projects) != 3 {
		t.Errorf("len = %d, want 3", len(projects))
	}
}