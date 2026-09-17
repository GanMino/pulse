package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pulse/pulse/internal/model"
	"github.com/pulse/pulse/internal/repository/sqlite"
)

// TestRunService 测试运行服务(简化版,MVP 阶段)
type TestRunService struct {
	testRunRepo  *sqlite.TestRunRepository
	scenarioRepo *sqlite.ScenarioRepository
	projectRepo  *sqlite.ProjectRepository
	userRepo     *sqlite.UserRepository
}

// NewTestRunService 创建 TestRunService
func NewTestRunService(
	testRunRepo *sqlite.TestRunRepository,
	scenarioRepo *sqlite.ScenarioRepository,
	projectRepo *sqlite.ProjectRepository,
	userRepo *sqlite.UserRepository,
) *TestRunService {
	return &TestRunService{
		testRunRepo:  testRunRepo,
		scenarioRepo: scenarioRepo,
		projectRepo:  projectRepo,
		userRepo:     userRepo,
	}
}

// TestRunDTO 测试运行数据传输对象
type TestRunDTO struct {
	ID          int64                  `json:"id"`
	UUID        string                 `json:"uuid"`
	ScenarioID  int64                  `json:"scenarioId"`
	ScenarioName string                `json:"scenarioName"`
	ProjectID   int64                  `json:"projectId"`
	Status      string                 `json:"status"`
	StartedAt   string                 `json:"startedAt"`
	FinishedAt  string                 `json:"finishedAt"`
	DurationMs  int64                  `json:"durationMs"`
	TriggerType string                 `json:"triggerType"`
	ErrorMessage string                `json:"errorMessage"`
	Summary     *model.RunSummary      `json:"summary"`
	CreatedAt   string                 `json:"createdAt"`
}

// ListByScenario 列出场景的所有测试运行
func (s *TestRunService) ListByScenario(ctx context.Context, scenarioID int64, limit int) ([]TestRunDTO, error) {
	runs, err := s.testRunRepo.ListByScenario(ctx, scenarioID, limit)
	if err != nil {
		return nil, fmt.Errorf("list runs: %w", err)
	}

	// 获取场景名称
	scenario, _ := s.scenarioRepo.GetByID(ctx, scenarioID)
	scenarioName := ""
	if scenario != nil {
		scenarioName = scenario.Name
	}

	items := make([]TestRunDTO, 0, len(runs))
	for _, r := range runs {
		items = append(items, s.toDTO(r, scenarioName))
	}
	return items, nil
}

// ListAll 列出所有测试运行
func (s *TestRunService) ListAll(ctx context.Context, limit int) ([]TestRunDTO, error) {
	if limit <= 0 {
		limit = 50
	}
	runs, total, err := s.testRunRepo.List(ctx, RunListOptions{Limit: limit})
	if err != nil {
		return nil, err
	}

	items := make([]TestRunDTO, 0, len(runs))
	for _, r := range runs {
		scenarioName := ""
		if sc, _ := s.scenarioRepo.GetByID(ctx, r.ScenarioID); sc != nil {
			scenarioName = sc.Name
		}
		items = append(items, s.toDTO(r, scenarioName))
	}
	_ = total
	return items, nil
}

// RunListOptions 列表选项(简化版)
type RunListOptions struct {
	ProjectID  int64
	ScenarioID int64
	Status     string
	Limit      int
	Offset     int
}

// toDTO 转换为 DTO
func (s *TestRunService) toDTO(r *model.TestRun, scenarioName string) TestRunDTO {
	dto := TestRunDTO{
		ID:           r.ID,
		UUID:         r.UUID,
		ScenarioID:   r.ScenarioID,
		ScenarioName: scenarioName,
		ProjectID:    r.ProjectID,
		Status:       string(r.Status),
		DurationMs:   r.DurationMs,
		TriggerType:  string(r.TriggerType),
		ErrorMessage: r.ErrorMessage,
		Summary:      r.Summary,
		CreatedAt:    r.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if r.StartedAt != nil {
		dto.StartedAt = r.StartedAt.Format("2006-01-02 15:04:05")
	}
	if r.FinishedAt != nil {
		dto.FinishedAt = r.FinishedAt.Format("2006-01-02 15:04:05")
	}
	return dto
}

// 防止 uuid 未使用告警
var _ = uuid.New