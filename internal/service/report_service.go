package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pulse/pulse/internal/model"
	"github.com/pulse/pulse/internal/repository/sqlite"
)

// ReportService 报告服务(简化版,MVP)
type ReportService struct {
	reportRepo  *sqlite.ReportRepository
	testRunRepo *sqlite.TestRunRepository
}

// NewReportService 创建 ReportService
func NewReportService(
	reportRepo *sqlite.ReportRepository,
	testRunRepo *sqlite.TestRunRepository,
) *ReportService {
	return &ReportService{
		reportRepo:  reportRepo,
		testRunRepo: testRunRepo,
	}
}

// ReportDTO 报告数据传输对象
type ReportDTO struct {
	ID         int64              `json:"id"`
	UUID       string             `json:"uuid"`
	TestRunID  int64              `json:"testRunId"`
	Summary    *model.RunSummary  `json:"summary"`
	HTMLPath   string             `json:"htmlPath"`
	Status     string             `json:"status"`
	CreatedAt  string             `json:"createdAt"`
}

// List 列出所有报告
func (s *ReportService) List(ctx context.Context, limit, offset int) ([]ReportDTO, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	reports, total, err := s.reportRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	items := make([]ReportDTO, 0, len(reports))
	for _, r := range reports {
		items = append(items, ReportDTO{
			ID:        r.ID,
			UUID:      r.UUID,
			TestRunID: r.TestRunID,
			Summary:   r.Summary,
			HTMLPath:  r.HTMLPath,
			Status:    string(r.Status),
			CreatedAt: r.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return items, total, nil
}

// GetByTestRunID 根据测试运行 ID 获取报告
func (s *ReportService) GetByTestRunID(ctx context.Context, testRunID int64) (*ReportDTO, error) {
	rep, err := s.reportRepo.GetByTestRunID(ctx, testRunID)
	if err != nil {
		return nil, fmt.Errorf("get report: %w", err)
	}
	return &ReportDTO{
		ID:        rep.ID,
		UUID:      rep.UUID,
		TestRunID: rep.TestRunID,
		Summary:   rep.Summary,
		HTMLPath:  rep.HTMLPath,
		Status:    string(rep.Status),
		CreatedAt: rep.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// ExportHTML 导出报告为 HTML 文件
func (s *ReportService) ExportHTML(ctx context.Context, reportID int64, outputDir string) (string, error) {
	rep, err := s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return "", fmt.Errorf("get report: %w", err)
	}

	if outputDir == "" {
		home, _ := os.UserHomeDir()
		outputDir = filepath.Join(home, ".pulse", "reports")
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("create dir: %w", err)
	}

	filename := fmt.Sprintf("pulse-report-%s.html", rep.UUID)
	fullPath := filepath.Join(outputDir, filename)

	// 生成 HTML 内容(MVP 简化)
	html := generateSimpleHTML(rep)

	if err := os.WriteFile(fullPath, []byte(html), 0o644); err != nil {
		return "", fmt.Errorf("write html: %w", err)
	}

	// 更新 DB 中的路径
	rep.HTMLPath = fullPath
	if err := s.reportRepo.Update(ctx, rep); err != nil {
		return "", fmt.Errorf("update report path: %w", err)
	}

	return fullPath, nil
}

// generateSimpleHTML 生成简单的 HTML 报告(MVP 占位)
func generateSimpleHTML(rep *model.Report) string {
	summary := rep.Summary
	if summary == nil {
		summary = &model.RunSummary{}
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<title>Pulse Report - %s</title>
<style>
  body { font-family: -apple-system, sans-serif; max-width: 800px; margin: 40px auto; padding: 0 20px; color: #1a1a1a; }
  h1 { color: #3370FF; }
  .kpi { display: inline-block; margin: 10px 20px 10px 0; padding: 16px; background: #f5f5f5; border-radius: 8px; }
  .kpi-label { font-size: 12px; color: #666; }
  .kpi-value { font-size: 28px; font-weight: 600; color: #1a1a1a; }
</style>
</head>
<body>
<h1>📊 Pulse Report</h1>
<p>Generated: %s</p>

<div>
  <div class="kpi"><div class="kpi-label">Total Requests</div><div class="kpi-value">%d</div></div>
  <div class="kpi"><div class="kpi-label">Total Errors</div><div class="kpi-value">%d</div></div>
  <div class="kpi"><div class="kpi-label">Error Rate</div><div class="kpi-value">%.2f%%</div></div>
  <div class="kpi"><div class="kpi-label">Avg RPS</div><div class="kpi-value">%.1f</div></div>
  <div class="kpi"><div class="kpi-label">P95 Latency</div><div class="kpi-value">%.0f ms</div></div>
  <div class="kpi"><div class="kpi-label">Duration</div><div class="kpi-value">%d ms</div></div>
</div>

<p style="margin-top: 40px; color: #666; font-size: 12px;">
  Pulse v0.1.0 · Apache 2.0 License
</p>
</body>
</html>`,
		rep.UUID,
		rep.CreatedAt.Format("2006-01-02 15:04:05"),
		summary.TotalRequests,
		summary.TotalErrors,
		summary.ErrorRate*100,
		summary.AvgRPS,
		summary.P95Ms,
		summary.DurationMs,
	)
}