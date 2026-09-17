// Package service 实现 Pulse 的业务逻辑层
// Service 层封装 Repository,提供更高级别的 API 给 Wails 绑定层
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pulse/pulse/internal/model"
	"github.com/pulse/pulse/internal/repository/sqlite"
)

// ErrScenarioNotFound 场景未找到
var ErrScenarioNotFound = errors.New("scenario not found")

// ErrProjectNotFound 项目未找到
var ErrProjectNotFound = errors.New("project not found")

// ScenarioService 场景服务
type ScenarioService struct {
	scenarioRepo *sqlite.ScenarioRepository
	projectRepo  *sqlite.ProjectRepository
	userRepo     *sqlite.UserRepository
}

// NewScenarioService 创建 ScenarioService
func NewScenarioService(
	scenarioRepo *sqlite.ScenarioRepository,
	projectRepo *sqlite.ProjectRepository,
	userRepo *sqlite.UserRepository,
) *ScenarioService {
	return &ScenarioService{
		scenarioRepo: scenarioRepo,
		projectRepo:  projectRepo,
		userRepo:     userRepo,
	}
}

// ListScenariosRequest 场景列表请求(给前端用)
type ListScenariosRequest struct {
	ProjectID int64  `json:"projectId,omitempty"`
	Search    string `json:"search,omitempty"`
	Source    string `json:"source,omitempty"`
	Status    string `json:"status,omitempty"`
	Tag       string `json:"tag,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

// ScenarioDTO 场景数据传输对象(给前端用,扁平化字段)
type ScenarioDTO struct {
	ID          int64                 `json:"id"`
	UUID        string                `json:"uuid"`
	ProjectID   int64                 `json:"projectId"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	VUs         int                   `json:"vus"`
	Duration    string                `json:"duration"`
	RequestCount int                  `json:"requestCount"`
	Source      string                `json:"source"`
	Status      string                `json:"status"`
	Tags        []string               `json:"tags"`
	Version     int                   `json:"version"`
	CreatedAt   string                `json:"createdAt"`
	UpdatedAt   string                `json:"updatedAt"`
	// 简化后的请求列表(用于列表展示)
	Requests    []RequestPreviewDTO   `json:"requests,omitempty"`
}

// RequestPreviewDTO 请求预览(列表中显示)
type RequestPreviewDTO struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	URL    string `json:"url"`
}

// ListScenariosResponse 列表响应
type ListScenariosResponse struct {
	Items  []ScenarioDTO `json:"items"`
	Total  int64         `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

// List 列出场景(支持搜索、过滤、分页)
func (s *ScenarioService) List(ctx context.Context, req ListScenariosRequest) (*ListScenariosResponse, error) {
	opts := sqlite.ListOptions{
		ProjectID: req.ProjectID,
		Search:    req.Search,
		Source:    req.Source,
		Status:    req.Status,
		Tag:       req.Tag,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}

	scenarios, total, err := s.scenarioRepo.List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("list scenarios: %w", err)
	}

	items := make([]ScenarioDTO, 0, len(scenarios))
	for _, sc := range scenarios {
		items = append(items, toScenarioDTO(sc))
	}

	if opts.Limit <= 0 {
		opts.Limit = 50
	}

	return &ListScenariosResponse{
		Items:  items,
		Total:  total,
		Limit:  opts.Limit,
		Offset: opts.Offset,
	}, nil
}

// GetByID 获取场景详情(含完整 Config)
func (s *ScenarioService) GetByID(ctx context.Context, id int64) (*ScenarioDTO, error) {
	sc, err := s.scenarioRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			return nil, ErrScenarioNotFound
		}
		return nil, err
	}
	dto := toScenarioDTO(sc)
	return &dto, nil
}

// GetByUUID 根据 UUID 获取
func (s *ScenarioService) GetByUUID(ctx context.Context, uuid string) (*ScenarioDTO, error) {
	sc, err := s.scenarioRepo.GetByUUID(ctx, uuid)
	if err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			return nil, ErrScenarioNotFound
		}
		return nil, err
	}
	dto := toScenarioDTO(sc)
	return &dto, nil
}

// SaveScenarioRequest 保存场景请求
type SaveScenarioRequest struct {
	ID          int64                  `json:"id,omitempty"` // 0 = 新建
	ProjectID   int64                  `json:"projectId"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      *model.ScenarioConfig  `json:"config"`
	Tags        []string               `json:"tags"`
	Source      string                 `json:"source,omitempty"`
}

// Save 保存场景(创建或更新)
func (s *ScenarioService) Save(ctx context.Context, req SaveScenarioRequest) (*ScenarioDTO, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Config == nil {
		return nil, fmt.Errorf("config is required")
	}

	// 获取当前用户(MVP 单机模式)
	user, err := s.userRepo.GetOrCreateDefault(ctx)
	if err != nil {
		return nil, fmt.Errorf("get default user: %w", err)
	}

	// 如果没有指定项目,使用默认项目
	if req.ProjectID == 0 {
		project, err := s.projectRepo.GetOrCreateDefault(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("get default project: %w", err)
		}
		req.ProjectID = project.ID
	}

	// 准备 tags
	var tags model.StringArray
	if len(req.Tags) > 0 {
		tags = model.StringArray(req.Tags)
	}

	source := model.ScenarioSource(req.Source)
	if source == "" {
		source = model.ScenarioSourceManual
	}

	if req.ID == 0 {
		// 新建
		sc := &model.Scenario{
			ProjectID:   req.ProjectID,
			Name:        req.Name,
			Description: req.Description,
			Config:      *req.Config,
			Tags:       tags,
			Source:     source,
			Status:     model.ScenarioStatusActive,
			CreatedBy:  user.ID,
		}
		if err := s.scenarioRepo.Create(ctx, sc); err != nil {
			return nil, fmt.Errorf("create scenario: %w", err)
		}
		dto := toScenarioDTO(sc)
		return &dto, nil
	}

	// 更新
	sc, err := s.scenarioRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			return nil, ErrScenarioNotFound
		}
		return nil, err
	}
	sc.Name = req.Name
	sc.Description = req.Description
	sc.Config = *req.Config
	sc.Tags = tags
	sc.Source = source

	if err := s.scenarioRepo.Update(ctx, sc); err != nil {
		return nil, fmt.Errorf("update scenario: %w", err)
	}
	dto := toScenarioDTO(sc)
	return &dto, nil
}

// Delete 删除场景
func (s *ScenarioService) Delete(ctx context.Context, id int64) error {
	sc, err := s.scenarioRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			return ErrScenarioNotFound
		}
		return err
	}
	return s.scenarioRepo.Delete(ctx, sc.ID)
}

// DeleteMany 批量删除
func (s *ScenarioService) DeleteMany(ctx context.Context, ids []int64) (int, error) {
	count := 0
	for _, id := range ids {
		if err := s.scenarioRepo.Delete(ctx, id); err == nil {
			count++
		}
	}
	return count, nil
}

// Duplicate 复制场景
func (s *ScenarioService) Duplicate(ctx context.Context, id int64) (*ScenarioDTO, error) {
	src, err := s.scenarioRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			return nil, ErrScenarioNotFound
		}
		return nil, err
	}

	// 创建副本
	dup := *src
	dup.ID = 0
	dup.Name = src.Name + " (副本)"
	dup.Version = 1
	dup.Status = model.ScenarioStatusDraft

	if err := s.scenarioRepo.Create(ctx, &dup); err != nil {
		return nil, fmt.Errorf("duplicate scenario: %w", err)
	}

	dto := toScenarioDTO(&dup)
	return &dto, nil
}

// toScenarioDTO 将 model 转换为 DTO
func toScenarioDTO(sc *model.Scenario) ScenarioDTO {
	dto := ScenarioDTO{
		ID:          sc.ID,
		UUID:        sc.UUID,
		ProjectID:   sc.ProjectID,
		Name:        sc.Name,
		Description: sc.Description,
		VUs:         sc.Config.Load.VUs,
		Duration:    sc.Config.Load.Duration,
		RequestCount: len(sc.Config.Requests),
		Source:      string(sc.Source),
		Status:      string(sc.Status),
		Tags:        []string(sc.Tags),
		Version:     sc.Version,
		CreatedAt:   sc.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   sc.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	// 简化的请求预览
	if len(sc.Config.Requests) > 0 {
		dto.Requests = make([]RequestPreviewDTO, 0, len(sc.Config.Requests))
		for _, r := range sc.Config.Requests {
			dto.Requests = append(dto.Requests, RequestPreviewDTO{
				Name:   r.Name,
				Method: r.Method,
				URL:    r.URL,
			})
		}
	}

	return dto
}