package sqlite

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/pulse/pulse/internal/model"
	"gorm.io/gorm"
)

// AgentRepository 是分布式 Agent 的数据访问层(预留)
type AgentRepository struct {
	db *DB
}

func NewAgentRepository(db *DB) *AgentRepository {
	return &AgentRepository{db: db}
}

// Register 注册一个 Agent
func (r *AgentRepository) Register(ctx context.Context, agent *model.Agent) error {
	if agent.UUID == "" {
		agent.UUID = uuid.New().String()
	}
	if agent.Status == "" {
		agent.Status = model.AgentStatusOnline
	}
	now := time.Now()
	agent.LastHeartbeat = &now

	return r.db.WithContext(ctx).Create(agent).Error
}

// UpdateHeartbeat 更新心跳时间
func (r *AgentRepository) UpdateHeartbeat(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.Agent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":         model.AgentStatusOnline,
			"last_heartbeat": now,
		}).Error
}

// GetByID 根据 ID 获取 Agent
func (r *AgentRepository) GetByID(ctx context.Context, id int64) (*model.Agent, error) {
	var a model.Agent
	err := r.db.WithContext(ctx).First(&a, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

// List 列出所有 Agent
func (r *AgentRepository) List(ctx context.Context) ([]*model.Agent, error) {
	var agents []*model.Agent
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&agents).Error
	return agents, err
}

// Delete 删除 Agent
func (r *AgentRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Agent{}, id).Error
}

// MarkStaleOffline 标记心跳超时(>30s)的 Agent 为离线
func (r *AgentRepository) MarkStaleOffline(ctx context.Context, timeout time.Duration) (int64, error) {
	threshold := time.Now().Add(-timeout)
	result := r.db.WithContext(ctx).
		Model(&model.Agent{}).
		Where("last_heartbeat < ? AND status = ?", threshold, model.AgentStatusOnline).
		Update("status", model.AgentStatusOffline)
	return result.RowsAffected, result.Error
}