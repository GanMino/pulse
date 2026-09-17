package model

import "time"

// ReportStatus 报告状态
type ReportStatus string

const (
	ReportStatusGenerating ReportStatus = "generating"
	ReportStatusReady      ReportStatus = "ready"
	ReportStatusFailed     ReportStatus = "failed"
)

// Report 是测试运行的完整报告
type Report struct {
	ID         int64        `json:"id" gorm:"primaryKey;autoIncrement"`
	UUID       string       `json:"uuid" gorm:"uniqueIndex;not null"`
	TestRunID  int64        `json:"test_run_id" gorm:"uniqueIndex;not null"`
	Summary    *RunSummary  `json:"summary" gorm:"type:text"`
	HTMLPath   string       `json:"html_path" gorm:"size:512"` // 本地文件路径
	Status     ReportStatus `json:"status" gorm:"size:32;default:'generating'"`
	CreatedAt  time.Time    `json:"created_at"`
}

func (Report) TableName() string {
	return "reports"
}