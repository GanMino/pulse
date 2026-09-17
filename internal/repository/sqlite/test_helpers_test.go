package sqlite

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pulse/pulse/internal/config"
)

// newTestDB 创建临时测试数据库
func newTestDB(t *testing.T) (*DB, func()) {
	t.Helper()

	// 临时目录
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			SQLitePath: dbPath,
		},
	}

	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if err := db.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}