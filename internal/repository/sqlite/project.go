package sqlite

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pulse/pulse/internal/model"
	"gorm.io/gorm"
)

// ProjectRepository 是项目的数据访问层
type ProjectRepository struct {
	db *DB
}

func NewProjectRepository(db *DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Create 创建项目
func (r *ProjectRepository) Create(ctx context.Context, p *model.Project) error {
	if p.UUID == "" {
		p.UUID = uuid.New().String()
	}
	err := r.db.WithContext(ctx).Create(p).Error
	if err != nil {
		if isUniqueConstraintError(err) {
			return fmt.Errorf("%w: project uuid=%s", ErrDuplicate, p.UUID)
		}
		return err
	}
	return nil
}

// GetByID 根据 ID 获取项目
func (r *ProjectRepository) GetByID(ctx context.Context, id int64) (*model.Project, error) {
	var p model.Project
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

// GetByUUID 根据 UUID 获取项目
func (r *ProjectRepository) GetByUUID(ctx context.Context, uuid string) (*model.Project, error) {
	var p model.Project
	err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

// Update 更新项目
func (r *ProjectRepository) Update(ctx context.Context, p *model.Project) error {
	return r.db.WithContext(ctx).Save(p).Error
}

// Delete 删除项目(MVP 硬删除;TODO 级联删除场景)
func (r *ProjectRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Project{}, id).Error
}

// List 列出所有项目
func (r *ProjectRepository) List(ctx context.Context) ([]*model.Project, error) {
	var projects []*model.Project
	err := r.db.WithContext(ctx).Order("updated_at DESC").Find(&projects).Error
	return projects, err
}

// GetOrCreateDefault 获取或创建默认项目(用于首次启动)
func (r *ProjectRepository) GetOrCreateDefault(ctx context.Context, ownerID int64) (*model.Project, error) {
	var project model.Project
	err := r.db.WithContext(ctx).Where("name = ?", "Default").First(&project).Error
	if err == nil {
		return &project, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	project = model.Project{
		UUID:        uuid.New().String(),
		Name:        "Default",
		Description: "默认项目",
		Color:       "#3370FF",
		OwnerID:     ownerID,
	}
	if err := r.Create(ctx, &project); err != nil {
		return nil, err
	}
	return &project, nil
}