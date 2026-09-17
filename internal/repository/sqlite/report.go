package sqlite

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pulse/pulse/internal/model"
	"gorm.io/gorm"
)

// ReportRepository 是报告的数据访问层
type ReportRepository struct {
	db *DB
}

func NewReportRepository(db *DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// Create 创建报告记录
func (r *ReportRepository) Create(ctx context.Context, rep *model.Report) error {
	if rep.UUID == "" {
		rep.UUID = uuid.New().String()
	}
	if rep.Status == "" {
		rep.Status = model.ReportStatusGenerating
	}
	return r.db.WithContext(ctx).Create(rep).Error
}

// GetByTestRunID 根据测试运行 ID 获取报告
func (r *ReportRepository) GetByTestRunID(ctx context.Context, testRunID int64) (*model.Report, error) {
	var rep model.Report
	err := r.db.WithContext(ctx).Where("test_run_id = ?", testRunID).First(&rep).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &rep, nil
}

// GetByID 根据 ID 获取报告
func (r *ReportRepository) GetByID(ctx context.Context, id int64) (*model.Report, error) {
	var rep model.Report
	err := r.db.WithContext(ctx).First(&rep, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &rep, nil
}

// UpdateStatus 更新报告状态
func (r *ReportRepository) UpdateStatus(ctx context.Context, id int64, status model.ReportStatus) error {
	return r.db.WithContext(ctx).
		Model(&model.Report{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// Update 更新报告
func (r *ReportRepository) Update(ctx context.Context, rep *model.Report) error {
	return r.db.WithContext(ctx).Save(rep).Error
}

// List 列出所有报告
func (r *ReportRepository) List(ctx context.Context, limit, offset int) ([]*model.Report, int64, error) {
	if limit <= 0 {
		limit = 50
	}

	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Report{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var reports []*model.Report
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&reports).Error
	return reports, total, err
}

// Delete 删除报告
func (r *ReportRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Report{}, id).Error
}

// DeleteOlderThan 删除指定天数之前的报告
func (r *ReportRepository) DeleteOlderThan(ctx context.Context, days int) (int64, error) {
	result := r.db.WithContext(ctx).
		Where(fmt.Sprintf("created_at < datetime('now', '-%d days')", days)).
		Delete(&model.Report{})
	return result.RowsAffected, result.Error
}