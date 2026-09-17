package service

import (
	"context"
	"fmt"

	"github.com/pulse/pulse/pkg/har"
)

// ImportService 提供 HAR 文件导入等高级导入功能
type ImportService struct {
	scenarioService *ScenarioService
	projectService  *ProjectService
}

// NewImportService 创建 ImportService
func NewImportService(s *ScenarioService, p *ProjectService) *ImportService {
	return &ImportService{
		scenarioService: s,
		projectService: p,
	}
}

// ImportHARRequest HAR 导入请求
type ImportHARRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	BaseURL     string `json:"baseUrl,omitempty"` // 替换 URL 的 base(如 "https://api.example.com")
	HARData     []byte `json:"harData"`          // HAR 文件内容
	VUs         int    `json:"vus,omitempty"`
	Duration    string `json:"duration,omitempty"`
}

// ImportHARResponse HAR 导入响应
type ImportHARResponse struct {
	Scenario *ScenarioDTO `json:"scenario"`
	Stats    ImportStats  `json:"stats"`
}

// ImportStats 导入统计
type ImportStats struct {
	TotalEntries   int `json:"totalEntries"`
	ImportedRequests int `json:"importedRequests"`
	DuplicatesSkipped int `json:"duplicatesSkipped"`
}

// ImportHAR 从 HAR 数据导入场景
func (s *ImportService) ImportHAR(ctx context.Context, req ImportHARRequest) (*ImportHARResponse, error) {
	// 解析 HAR
	parsed, err := har.Parse(req.HARData)
	if err != nil {
		return nil, fmt.Errorf("parse HAR: %w", err)
	}

	totalEntries := len(parsed.Log.Entries)

	// 转换选项
	opts := har.DefaultOptions()
	if req.BaseURL != "" {
		opts.BaseURL = req.BaseURL
	}
	if req.VUs > 0 {
		opts.DefaultVUs = req.VUs
	}
	if req.Duration != "" {
		opts.DefaultDuration = req.Duration
	}

	scenarioCfg := parsed.ToScenario(opts)

	// 默认场景名
	name := req.Name
	if name == "" {
		name = "Imported from HAR"
	}

	// 创建场景
	dto, err := s.scenarioService.Save(ctx, SaveScenarioRequest{
		Name:        name,
		Description: req.Description,
		Config:      scenarioCfg,
		Source:      "har_import",
	})
	if err != nil {
		return nil, fmt.Errorf("save imported scenario: %w", err)
	}

	duplicatesSkipped := totalEntries - len(scenarioCfg.Requests)

	return &ImportHARResponse{
		Scenario: dto,
		Stats: ImportStats{
			TotalEntries:      totalEntries,
			ImportedRequests:  len(scenarioCfg.Requests),
			DuplicatesSkipped: duplicatesSkipped,
		},
	}, nil
}