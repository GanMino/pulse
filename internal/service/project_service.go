package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pulse/pulse/internal/model"
	"github.com/pulse/pulse/internal/repository/sqlite"
)

// ProjectService 项目服务
type ProjectService struct {
	projectRepo  *sqlite.ProjectRepository
	scenarioRepo *sqlite.ScenarioRepository
	userRepo     *sqlite.UserRepository
}

// NewProjectService 创建 ProjectService
func NewProjectService(
	projectRepo *sqlite.ProjectRepository,
	scenarioRepo *sqlite.ScenarioRepository,
	userRepo *sqlite.UserRepository,
) *ProjectService {
	return &ProjectService{
		projectRepo:  projectRepo,
		scenarioRepo: scenarioRepo,
		userRepo:     userRepo,
	}
}

// ProjectDTO 项目数据传输对象
type ProjectDTO struct {
	ID            int64  `json:"id"`
	UUID          string `json:"uuid"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Color         string `json:"color"`
	ScenarioCount int64  `json:"scenarioCount"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// List 列出所有项目
func (s *ProjectService) List(ctx context.Context) ([]ProjectDTO, error) {
	projects, err := s.projectRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	items := make([]ProjectDTO, 0, len(projects))
	for _, p := range projects {
		count, _ := s.scenarioRepo.CountByProject(ctx, p.ID)
		items = append(items, ProjectDTO{
			ID:            p.ID,
			UUID:          p.UUID,
			Name:          p.Name,
			Description:   p.Description,
			Color:         p.Color,
			ScenarioCount: count,
			CreatedAt:     p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     p.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return items, nil
}

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

// Create 创建项目
func (s *ProjectService) Create(ctx context.Context, req CreateProjectRequest) (*ProjectDTO, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	user, err := s.userRepo.GetOrCreateDefault(ctx)
	if err != nil {
		return nil, fmt.Errorf("get default user: %w", err)
	}

	if req.Color == "" {
		req.Color = "#3370FF"
	}

	p := &model.Project{
		UUID:        uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		OwnerID:     user.ID,
	}

	if err := s.projectRepo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}

	return &ProjectDTO{
		ID:            p.ID,
		UUID:          p.UUID,
		Name:          p.Name,
		Description:   p.Description,
		Color:         p.Color,
		ScenarioCount: 0,
		CreatedAt:     p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// GetOrCreateDefault 获取或创建默认项目
func (s *ProjectService) GetOrCreateDefault(ctx context.Context) (*ProjectDTO, error) {
	user, err := s.userRepo.GetOrCreateDefault(ctx)
	if err != nil {
		return nil, fmt.Errorf("get default user: %w", err)
	}

	project, err := s.projectRepo.GetOrCreateDefault(ctx, user.ID)
	if err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	return &ProjectDTO{
		ID:            project.ID,
		UUID:          project.UUID,
		Name:          project.Name,
		Description:   project.Description,
		Color:         project.Color,
		ScenarioCount: 0,
		CreatedAt:     project.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     project.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}