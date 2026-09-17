package sqlite

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pulse/pulse/internal/model"
	"gorm.io/gorm"
)

// TestRunRepository 是测试运行记录的数据访问层
type TestRunRepository struct {
	db *DB
}

func NewTestRunRepository(db *DB) *TestRunRepository {
	return &TestRunRepository{db: db}
}

// Create 创建测试运行记录
func (r *TestRunRepository) Create(ctx context.Context, run *model.TestRun) error {
	if run.UUID == "" {
		run.UUID = uuid.New().String()
	}
	if run.Status == "" {
		run.Status = model.TestRunStatusPending
	}
	if run.TriggerType == "" {
		run.TriggerType = model.TestRunTriggerManual
	}
	return r.db.WithContext(ctx).Create(run).Error
}

// GetByID 根据 ID 获取测试运行
func (r *TestRunRepository) GetByID(ctx context.Context, id int64) (*model.TestRun, error) {
	var run model.TestRun
	err := r.db.WithContext(ctx).First(&run, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &run, nil
}

// GetByUUID 根据 UUID 获取测试运行
func (r *TestRunRepository) GetByUUID(ctx context.Context, uuid string) (*model.TestRun, error) {
	var run model.TestRun
	err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&run).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &run, nil
}

// UpdateStatus 更新测试运行状态
func (r *TestRunRepository) UpdateStatus(ctx context.Context, id int64, status model.TestRunStatus) error {
	result := r.db.WithContext(ctx).
		Model(&model.TestRun{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("test run %d not found", id)
	}
	return nil
}

// UpdateResult 更新测试运行结果
func (r *TestRunRepository) UpdateResult(ctx context.Context, run *model.TestRun) error {
	return r.db.WithContext(ctx).
		Model(run).
		Updates(map[string]interface{}{
			"status":         run.Status,
			"started_at":     run.StartedAt,
			"finished_at":    run.FinishedAt,
			"duration_ms":    run.DurationMs,
			"summary":        run.Summary,
			"error_message":  run.ErrorMessage,
		}).Error
}

// ListByScenario 列出场景的所有测试运行
func (r *TestRunRepository) ListByScenario(ctx context.Context, scenarioID int64, limit int) ([]*model.TestRun, error) {
	if limit <= 0 {
		limit = 50
	}
	var runs []*model.TestRun
	err := r.db.WithContext(ctx).
		Where("scenario_id = ?", scenarioID).
		Order("created_at DESC").
		Limit(limit).
		Find(&runs).Error
	return runs, err
}

// ListOptions 测试运行列表选项
type RunListOptions struct {
	ProjectID  int64
	ScenarioID int64
	Status     string
	Limit      int
	Offset     int
}

// List 通用查询
func (r *TestRunRepository) List(ctx context.Context, opts RunListOptions) ([]*model.TestRun, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.TestRun{})

	if opts.ProjectID > 0 {
		q = q.Where("project_id = ?", opts.ProjectID)
	}
	if opts.ScenarioID > 0 {
		q = q.Where("scenario_id = ?", opts.ScenarioID)
	}
	if opts.Status != "" {
		q = q.Where("status = ?", opts.Status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if opts.Limit <= 0 {
		opts.Limit = 50
	}
	if opts.Limit > 200 {
		opts.Limit = 200
	}

	var runs []*model.TestRun
	err := q.Order("created_at DESC").
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&runs).Error
	return runs, total, err
}

// Delete 删除测试运行(MVP 硬删除)
func (r *TestRunRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.TestRun{}, id).Error
}