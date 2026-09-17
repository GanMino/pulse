package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pulse/pulse/internal/model"
	"github.com/pulse/pulse/internal/repository/sqlite"
)

// UserService 用户服务(简化,本地单机模式)
type UserService struct {
	userRepo    *sqlite.UserRepository
	projectRepo *sqlite.ProjectRepository
}

// NewUserService 创建 UserService
func NewUserService(userRepo *sqlite.UserRepository, projectRepo *sqlite.ProjectRepository) *UserService {
	return &UserService{
		userRepo:    userRepo,
		projectRepo: projectRepo,
	}
}

// UserDTO 用户数据传输对象
type UserDTO struct {
	ID          int64  `json:"id"`
	UUID        string `json:"uuid"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

// GetCurrentUser 获取当前用户(本地默认用户)
func (s *UserService) GetCurrentUser(ctx context.Context) (*UserDTO, error) {
	user, err := s.userRepo.GetOrCreateDefault(ctx)
	if err != nil {
		return nil, fmt.Errorf("get default user: %w", err)
	}
	return &UserDTO{
		ID:          user.ID,
		UUID:        user.UUID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        user.Role,
	}, nil
}

// EnsureDefaultUserAndProject 确保默认用户和项目存在
// 在应用启动时调用
func (s *UserService) EnsureDefaultUserAndProject(ctx context.Context) error {
	user, err := s.userRepo.GetOrCreateDefault(ctx)
	if err != nil {
		return err
	}
	_, err = s.projectRepo.GetOrCreateDefault(ctx, user.ID)
	if err != nil {
		return err
	}
	return nil
}

// 防止 uuid 未使用告警
var _ = uuid.New
var _ = errors.New
var _ = model.User{}