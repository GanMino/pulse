package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// AgentStatus 分布式 Agent 状态(预留)
type AgentStatus string

const (
	AgentStatusOnline    AgentStatus = "online"
	AgentStatusOffline   AgentStatus = "offline"
	AgentStatusBusy      AgentStatus = "busy"
)

// Agent 是分布式压测节点(MVP 仅预留,本地单机模式不实现具体逻辑)
type Agent struct {
	ID            int64       `json:"id" gorm:"primaryKey;autoIncrement"`
	UUID          string      `json:"uuid" gorm:"uniqueIndex;not null"`
	Name          string      `json:"name" gorm:"not null;size:128"`
	Address       string      `json:"address" gorm:"not null;size:128"` // host:port
	Status        AgentStatus `json:"status" gorm:"size:32;default:'offline';index"`
	LastHeartbeat *time.Time  `json:"last_heartbeat"`
	Tags          StringArray `json:"tags" gorm:"type:text"` // e.g. {"region":"us-east","max_vus":10000}
	MaxVUs        int         `json:"max_vus" gorm:"default:0"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func (Agent) TableName() string {
	return "agents"
}

// AgentTags 便捷方法:获取标签
func (a *Agent) GetTags() StringArray {
	return a.Tags
}

// SetTagsJSON 设置标签的 JSON 字符串
func (a *Agent) SetTagsJSON(tags string) error {
	var arr StringArray
	if err := json.Unmarshal([]byte(tags), &arr); err != nil {
		return err
	}
	a.Tags = arr
	return nil
}

// Value 暴露 Tags 的 driver.Valuer
func (a *Agent) Value() (driver.Value, error) {
	return json.Marshal(a)
}