package sqlite

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/pulse/pulse/internal/model"
	"gorm.io/gorm"
)

// ErrNotFound 通用未找到错误
var ErrNotFound = errors.New("record not found")

// ErrDuplicate 重复创建错误
var ErrDuplicate = errors.New("duplicate record")

// ScenarioRepository 是场景的数据访问层
type ScenarioRepository struct {
	db *DB
}

// NewScenarioRepository 创建 ScenarioRepository
func NewScenarioRepository(db *DB) *ScenarioRepository {
	return &ScenarioRepository{db: db}
}

// Create 创建新场景
func (r *ScenarioRepository) Create(ctx context.Context, s *model.Scenario) error {
	if s.UUID == "" {
		s.UUID = uuid.New().String()
	}
	if s.Version == 0 {
		s.Version = 1
	}
	if s.Status == "" {
		s.Status = model.ScenarioStatusDraft
	}
	if s.Source == "" {
		s.Source = model.ScenarioSourceManual
	}

	err := r.db.WithContext(ctx).Create(s).Error
	if err != nil {
		if isUniqueConstraintError(err) {
			return fmt.Errorf("%w: scenario uuid=%s", ErrDuplicate, s.UUID)
		}
		return err
	}
	return nil
}

// GetByID 根据 ID 获取场景
func (r *ScenarioRepository) GetByID(ctx context.Context, id int64) (*model.Scenario, error) {
	var s model.Scenario
	err := r.db.WithContext(ctx).First(&s, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

// GetByUUID 根据 UUID 获取场景
func (r *ScenarioRepository) GetByUUID(ctx context.Context, uuid string) (*model.Scenario, error) {
	var s model.Scenario
	err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

// Update 更新场景(自动增加版本号)
func (r *ScenarioRepository) Update(ctx context.Context, s *model.Scenario) error {
	// 使用乐观锁:version 字段
	result := r.db.WithContext(ctx).
		Model(s).
		Where("id = ? AND version = ?", s.ID, s.Version).
		Updates(map[string]interface{}{
			"name":        s.Name,
			"description": s.Description,
			"config":      s.Config,
			"status":      s.Status,
			"tags":        s.Tags,
			"version":     s.Version + 1,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("optimistic lock failed: scenario id=%d version=%d", s.ID, s.Version)
	}
	s.Version++
	return nil
}

// UpdateConfig 仅更新场景配置(不触发版本号变化)
// 用于自动化场景下不计数版本
func (r *ScenarioRepository) UpdateConfig(ctx context.Context, id int64, cfg *model.ScenarioConfig) error {
	return r.db.WithContext(ctx).
		Model(&model.Scenario{}).
		Where("id = ?", id).
		Update("config", cfg).Error
}

// Delete 删除场景(软删除可用 GORM DeletedAt,这里采用硬删除)
func (r *ScenarioRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Scenario{}, id).Error
}

// ListOptions 列表查询选项
type ListOptions struct {
	ProjectID int64
	Source    string
	Status    string
	Tag       string
	Search    string
	Limit     int
	Offset    int
	OrderBy   string
	OrderDesc bool
}

// List 分页查询场景列表
func (r *ScenarioRepository) List(ctx context.Context, opts ListOptions) ([]*model.Scenario, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Scenario{})

	if opts.ProjectID > 0 {
		q = q.Where("project_id = ?", opts.ProjectID)
	}
	if opts.Source != "" {
		q = q.Where("source = ?", opts.Source)
	}
	if opts.Status != "" {
		q = q.Where("status = ?", opts.Status)
	}
	if opts.Search != "" {
		like := "%" + strings.TrimSpace(opts.Search) + "%"
		q = q.Where("name LIKE ? OR description LIKE ?", like, like)
	}

	// 标签过滤(tags 是 JSON 数组,需要 LIKE 简单处理)
	// 注:更精确的可以用 json_each() 扩展
	if opts.Tag != "" {
		q = q.Where("tags LIKE ?", "%\""+opts.Tag+"\"%")
	}

	// 总数
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	orderBy := "updated_at"
	if opts.OrderBy != "" {
		orderBy = opts.OrderBy
	}
	if opts.OrderDesc {
		orderBy += " DESC"
	} else {
		orderBy += " ASC"
	}
	q = q.Order(orderBy)

	// 分页
	if opts.Limit <= 0 {
		opts.Limit = 50
	}
	if opts.Limit > 500 {
		opts.Limit = 500
	}
	q = q.Limit(opts.Limit).Offset(opts.Offset)

	var scenarios []*model.Scenario
	if err := q.Find(&scenarios).Error; err != nil {
		return nil, 0, err
	}

	return scenarios, total, nil
}

// ListAll 列出所有场景(用于备份导出)
func (r *ScenarioRepository) ListAll(ctx context.Context) ([]*model.Scenario, error) {
	var scenarios []*model.Scenario
	err := r.db.WithContext(ctx).
		Order("updated_at DESC").
		Find(&scenarios).Error
	return scenarios, err
}

// CountByProject 统计项目下的场景数
func (r *ScenarioRepository) CountByProject(ctx context.Context, projectID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Scenario{}).
		Where("project_id = ?", projectID).
		Count(&count).Error
	return count, err
}

// isUniqueConstraintError 判断是否为唯一约束错误
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "duplicate key")
}