package service

import (
	"context"
	"fmt"
	"time"

	"github.com/pulse/pulse/internal/model"
	"github.com/pulse/pulse/internal/repository/sqlite"
)

// AgentService Agent 服务(分布式预留)
type AgentService struct {
	agentRepo *sqlite.AgentRepository
}

// NewAgentService 创建 AgentService
func NewAgentService(agentRepo *sqlite.AgentRepository) *AgentService {
	return &AgentService{agentRepo: agentRepo}
}

// AgentDTO Agent DTO
type AgentDTO struct {
	ID            int64     `json:"id"`
	UUID          string    `json:"uuid"`
	Name          string    `json:"name"`
	Address       string    `json:"address"`
	Status        string    `json:"status"`
	LastHeartbeat string    `json:"lastHeartbeat"`
	Tags          []string  `json:"tags"`
	MaxVUs        int       `json:"maxVUs"`
	CreatedAt     time.Time `json:"createdAt"`
}

// List 列出所有 Agent
func (s *AgentService) List(ctx context.Context) ([]AgentDTO, error) {
	agents, err := s.agentRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list agents: %w", err)
	}

	items := make([]AgentDTO, 0, len(agents))
	for _, a := range agents {
		items = append(items, AgentDTO{
			ID:            a.ID,
			UUID:          a.UUID,
			Name:          a.Name,
			Address:       a.Address,
			Status:        string(a.Status),
			LastHeartbeat: formatTime(a.LastHeartbeat),
			Tags:          []string(a.Tags),
			MaxVUs:        a.MaxVUs,
			CreatedAt:     a.CreatedAt,
		})
	}
	return items, nil
}

// RegisterLocal 注册一个本地 Agent(MVP 阶段占位)
func (s *AgentService) RegisterLocal(ctx context.Context) (*AgentDTO, error) {
	agent := &model.Agent{
		Name:    "local-1",
		Address: "localhost",
		Status:   model.AgentStatusOnline,
		MaxVUs:   20000,
		Tags:     model.StringArray{"local", "mvp"},
	}
	if err := s.agentRepo.Register(ctx, agent); err != nil {
		return nil, err
	}

	return &AgentDTO{
		ID:            agent.ID,
		UUID:          agent.UUID,
		Name:          agent.Name,
		Address:       agent.Address,
		Status:        string(agent.Status),
		LastHeartbeat: formatTime(agent.LastHeartbeat),
		Tags:          []string(agent.Tags),
		MaxVUs:        agent.MaxVUs,
		CreatedAt:     agent.CreatedAt,
	}, nil
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}