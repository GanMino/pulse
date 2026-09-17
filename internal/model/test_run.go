package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// TestRunStatus 测试运行状态
type TestRunStatus string

const (
	TestRunStatusPending   TestRunStatus = "pending"
	TestRunStatusRunning   TestRunStatus = "running"
	TestRunStatusPaused    TestRunStatus = "paused"
	TestRunStatusCompleted  TestRunStatus = "completed"
	TestRunStatusFailed    TestRunStatus = "failed"
	TestRunStatusAborted   TestRunStatus = "aborted"
)

// TestRunTriggerType 触发类型
type TestRunTriggerType string

const (
	TestRunTriggerManual    TestRunTriggerType = "manual"
	TestRunTriggerScheduled TestRunTriggerType = "scheduled"
	TestRunTriggerCI        TestRunTriggerType = "ci"
)

// TestRun 是一次压测执行的记录
type TestRun struct {
	ID            int64               `json:"id" gorm:"primaryKey;autoIncrement"`
	UUID          string              `json:"uuid" gorm:"uniqueIndex;not null"`
	ScenarioID    int64               `json:"scenario_id" gorm:"index;not null"`
	ProjectID     int64               `json:"project_id" gorm:"index;not null"`
	Status        TestRunStatus       `json:"status" gorm:"size:32;index;not null"`
	StartedAt     *time.Time          `json:"started_at"`
	FinishedAt    *time.Time          `json:"finished_at"`
	DurationMs    int64               `json:"duration_ms"` // 实际运行时长(毫秒)
	RuntimeConfig *RuntimeConfig      `json:"runtime_config" gorm:"type:text"`
	Summary       *RunSummary         `json:"summary" gorm:"type:text"`
	TriggerType   TestRunTriggerType  `json:"trigger_type" gorm:"size:32;default:'manual'"`
	TriggeredBy   int64               `json:"triggered_by" gorm:"index"`
	ErrorMessage  string              `json:"error_message" gorm:"type:text"`
	CreatedAt     time.Time           `json:"created_at"`
}

func (TestRun) TableName() string {
	return "test_runs"
}

// RuntimeConfig 运行时配置(可在运行时覆盖场景配置)
type RuntimeConfig struct {
	VUs       int    `json:"vus,omitempty"`
	Duration  string `json:"duration,omitempty"`
	TargetRPS int    `json:"targetRPS,omitempty"`
}

// RunSummary 是测试运行的汇总指标
type RunSummary struct {
	TotalRequests   int64                  `json:"totalRequests"`
	TotalErrors     int64                  `json:"totalErrors"`
	ErrorRate       float64                `json:"errorRate"`
	AvgRPS          float64                `json:"avgRPS"`
	PeakRPS         float64                `json:"peakRPS"`
	P50Ms           float64                `json:"p50Ms"`
	P90Ms           float64                `json:"p90Ms"`
	P95Ms           float64                `json:"p95Ms"`
	P99Ms           float64                `json:"p99Ms"`
	MaxMs           float64                `json:"maxMs"`
	AvgMs           float64                `json:"avgMs"`
	StatusCodes     map[string]int64       `json:"statusCodes"`
	RequestStats    map[string]*RequestStat `json:"requestStats,omitempty"`
	ThresholdResults []ThresholdResult     `json:"thresholdResults,omitempty"`
}

// RequestStat 单个请求的统计
type RequestStat struct {
	Name       string  `json:"name"`
	Count      int64   `json:"count"`
	Errors     int64   `json:"errors"`
	AvgRPS     float64 `json:"avgRPS"`
	P95Ms      float64 `json:"p95Ms"`
	P99Ms      float64 `json:"p99Ms"`
}

// ThresholdResult 阈值检查结果
type ThresholdResult struct {
	Name      string  `json:"name"`      // p95Ms / errorRate / minRPS
	Operator  string  `json:"operator"`  // < / > / <= / >=
	Expected  float64 `json:"expected"`
	Actual    float64 `json:"actual"`
	Passed    bool    `json:"passed"`
}

// ============================================
// JSON 序列化
// ============================================

func (c *RuntimeConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

func (c *RuntimeConfig) Scan(value interface{}) error {
	if value == nil {
		*c = RuntimeConfig{}
		return nil
	}
	return unmarshalJSON(value, c)
}

func (s *RunSummary) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

func (s *RunSummary) Scan(value interface{}) error {
	if value == nil {
		*s = RunSummary{}
		return nil
	}
	return unmarshalJSON(value, s)
}

func unmarshalJSON(value interface{}, target interface{}) error {
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("unmarshalJSON: unsupported scan type")
	}
	return json.Unmarshal(data, target)
}