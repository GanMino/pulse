package sqlite

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/pulse/pulse/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserRepository 是用户的数据访问层
// MVP 桌面应用通常只有一个本地用户,实现简化
type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户(密码会被 bcrypt 哈希)
func (r *UserRepository) Create(ctx context.Context, u *model.User, plainPassword string) error {
	if u.UUID == "" {
		u.UUID = uuid.New().String()
	}
	if plainPassword != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.PasswordHash = string(hash)
	}
	return r.db.WithContext(ctx).Create(u).Error
}

// GetByID 根据 ID 获取用户
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).First(&u, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

// VerifyPassword 校验密码
func (r *UserRepository) VerifyPassword(u *model.User, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plainPassword))
	return err == nil
}

// GetOrCreateDefault 获取或创建默认本地用户
func (r *UserRepository) GetOrCreateDefault(ctx context.Context) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("username = ?", "local").First(&u).Error
	if err == nil {
		return &u, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	u = model.User{
		UUID:        uuid.New().String(),
		Username:    "local",
		Email:       "local@pulse.dev",
		DisplayName: "Local User",
		Role:        "admin",
	}
	// 本地用户无密码(单机模式)
	if err := r.Create(ctx, &u, ""); err != nil {
		return nil, err
	}
	return &u, nil
}