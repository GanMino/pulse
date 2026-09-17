// Package model 定义 Pulse 的领域模型
package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// ScenarioSource 场景来源
type ScenarioSource string

const (
	ScenarioSourceManual    ScenarioSource = "manual"
	ScenarioSourceHARImport ScenarioSource = "har_import"
	ScenarioSourceTemplate  ScenarioSource = "template"
)

// ScenarioStatus 场景状态
type ScenarioStatus string

const (
	ScenarioStatusDraft    ScenarioStatus = "draft"
	ScenarioStatusActive   ScenarioStatus = "active"
	ScenarioStatusArchived ScenarioStatus = "archived"
)

// Scenario 是压测场景的领域模型
type Scenario struct {
	ID          int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	UUID        string          `json:"uuid" gorm:"uniqueIndex;not null"`
	ProjectID   int64           `json:"project_id" gorm:"index;not null"`
	Name        string          `json:"name" gorm:"not null;size:128"`
	Description string          `json:"description" gorm:"type:text"`
	Config      ScenarioConfig  `json:"config" gorm:"type:text;not null"` // JSON 序列化
	Source      ScenarioSource  `json:"source" gorm:"size:32;default:'manual'"`
	Status      ScenarioStatus  `json:"status" gorm:"size:32;default:'draft'"`
	Tags        StringArray     `json:"tags" gorm:"type:text"` // JSON 数组
	Version     int             `json:"version" gorm:"default:1"`
	CreatedBy   int64           `json:"created_by" gorm:"index"`
	CreatedAt   time.Time       `json:"created_at" gorm:"autoIncrement"`
	UpdatedAt   time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定 GORM 表名
func (Scenario) TableName() string {
	return "scenarios"
}

// ScenarioConfig 是场景配置的完整结构
// 存储为 JSON,可灵活支持未来的扩展
type ScenarioConfig struct {
	// 压测配置
	Load LoadConfig `json:"load"`

	// 请求列表
	Requests []RequestConfig `json:"requests"`

	// 变量定义(参数化)
	Variables map[string]VariableValue `json:"variables,omitempty"`

	// 阈值定义
	Thresholds *ThresholdConfig `json:"thresholds,omitempty"`
}

// LoadConfig 压测负载配置
type LoadConfig struct {
	VUs          int          `json:"vus" binding:"required,min=1"`
	Duration     string       `json:"duration" binding:"required"` // e.g. "5m"
	RampUp       *RampUpConfig `json:"rampUp,omitempty"`
	TargetRPS    int          `json:"targetRPS,omitempty"`    // 0 = 不限
	ThinkTime    string       `json:"thinkTime,omitempty"`    // 默认 think time
	Iterations   int          `json:"iterations,omitempty"`  // 0 = 直到 duration 结束
}

// RampUpConfig 渐进式启动配置
type RampUpConfig struct {
	Type     string `json:"type"`     // linear / step / wave
	Duration string `json:"duration"` // e.g. "30s"
	Steps    int    `json:"steps,omitempty"` // for step type
}

// RequestConfig 单个 HTTP 请求配置
type RequestConfig struct {
	Name       string                 `json:"name" binding:"required"`
	Method     string                 `json:"method" binding:"required,oneof=GET POST PUT DELETE PATCH"`
	URL        string                 `json:"url" binding:"required"`
	Headers    map[string]string      `json:"headers,omitempty"`
	Body       map[string]interface{} `json:"body,omitempty"`
	BodyRaw    string                 `json:"bodyRaw,omitempty"` // 用于非 JSON body
	Extractors map[string]string      `json:"extractors,omitempty"` // JSONPath
	Assertions *AssertionConfig       `json:"assertions,omitempty"`
	ThinkTime  string                 `json:"thinkTime,omitempty"`
	Weight     int                    `json:"weight,omitempty"` // 多请求场景下的权重
}

// AssertionConfig 断言配置
type AssertionConfig struct {
	Status       []int    `json:"status,omitempty"`        // 期望的状态码列表
	MaxLatencyMs int      `json:"maxLatencyMs,omitempty"`  // 最大 P95 延迟
	BodyContains []string `json:"bodyContains,omitempty"`  // body 必须包含的字符串
	HeaderCheck  map[string]string `json:"headerCheck,omitempty"` // header 必须包含的键值
}

// VariableValue 是单个变量的定义
// 可以是单一值(标量)或多个值(数组,数据驱动测试)
type VariableValue struct {
	Value interface{} `json:"value"`
}

// ThresholdConfig 阈值配置
type ThresholdConfig struct {
	P95Ms        int     `json:"p95Ms,omitempty"`
	P99Ms        int     `json:"p99Ms,omitempty"`
	ErrorRate    float64 `json:"errorRate,omitempty"` // 0.01 = 1%
	MinRPS       int     `json:"minRPS,omitempty"`
	MaxAvgLatencyMs int  `json:"maxAvgLatencyMs,omitempty"`
}

// ============================================
// 自定义 GORM 数据类型
// ============================================

// ScenarioConfig 实现 driver.Valuer 和 sql.Scanner
// 用于在 SQLite 中存储为 TEXT(JSON)
func (c *ScenarioConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner
func (c *ScenarioConfig) Scan(value interface{}) error {
	if value == nil {
		*c = ScenarioConfig{}
		return nil
	}
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("ScenarioConfig: unsupported scan type")
	}
	return json.Unmarshal(data, c)
}

// StringArray 是 JSON 序列化的字符串数组
type StringArray []string

func (a *StringArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("StringArray: unsupported scan type")
	}
	return json.Unmarshal(data, a)
}