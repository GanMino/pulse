// Package sqlite 提供基于 SQLite + GORM 的数据访问实现
package sqlite

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite" // 纯 Go SQLite 驱动(避免 CGO 依赖)
	"github.com/pulse/pulse/internal/config"
	"github.com/pulse/pulse/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 是 Pulse 的数据库句柄
type DB struct {
	*gorm.DB
}

// Open 打开/创建 SQLite 数据库
func Open(cfg *config.Config) (*DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	dbPath := cfg.Database.SQLitePath
	if dbPath == "" {
		return nil, fmt.Errorf("sqlite path is empty")
	}

	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	// GORM 配置
	gormConfig := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Warn),
		DisableForeignKeyConstraintWhenMigrating: false,
		PrepareStmt:                              true, // 预编译 SQL,提升性能
	}

	// DSN:启用 WAL + 性能优化
	dsn := fmt.Sprintf(
		"file:%s?cache=shared&_journal_mode=WAL&_synchronous=NORMAL&_busy_timeout=5000&_foreign_keys=on",
		dbPath,
	)

	gdb, err := gorm.Open(sqlite.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// 设置连接池
	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &DB{DB: gdb}, nil
}

// Migrate 自动迁移所有表结构
// 保留向后兼容的能力:每个新版本手动追加迁移逻辑
func (db *DB) Migrate() error {
	// 1. 自动迁移基础模型
	if err := db.AutoMigrate(
		&model.User{},
		&model.Project{},
		&model.Scenario{},
		&model.TestRun{},
		&model.Report{},
		&model.Agent{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	// 2. 应用自定义迁移(版本管理)
	if err := db.applyCustomMigrations(); err != nil {
		return fmt.Errorf("custom migrations: %w", err)
	}

	return nil
}

// applyCustomMigrations 自定义迁移(版本号管理,避免 AutoMigrate 局限性)
func (db *DB) applyCustomMigrations() error {
	// 0. 创建 schema_versions 表(如果不存在)
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_versions (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			description TEXT
		)
	`).Error; err != nil {
		return err
	}

	migrations := []Migration{
		// {Version: 1, Description: "init schema", Up: ...},
	}

	for _, m := range migrations {
		var count int64
		if err := db.Raw("SELECT COUNT(*) FROM schema_versions WHERE version = ?", m.Version).Scan(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue // 已应用
		}

		if err := m.Up(db); err != nil {
			return fmt.Errorf("migration %d failed: %w", m.Version, err)
		}

		if err := db.Exec(
			"INSERT INTO schema_versions (version, description) VALUES (?, ?)",
			m.Version, m.Description,
		).Error; err != nil {
			return err
		}
	}

	return nil
}

// Migration 是单个迁移定义
type Migration struct {
	Version     int
	Description string
	Up          func(db *DB) error
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}