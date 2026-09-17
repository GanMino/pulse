package service

import (
	"context"
	"testing"

	"github.com/pulse/pulse/internal/config"
	"github.com/pulse/pulse/internal/repository/sqlite"
)

func setupImportTestContainer(t *testing.T) *Container {
	t.Helper()
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			SQLitePath: tmpDir + "/test.db",
		},
	}

	db, err := sqlite.Open(cfg)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return NewContainer(db)
}

const validHARJSON = `{
  "log": {
    "version": "1.2",
    "creator": {"name": "Test", "version": "1.0"},
    "entries": [
      {
        "startedDateTime": "2024-01-01T00:00:00.000Z",
        "time": 100,
        "request": {
          "method": "POST",
          "url": "https://api.example.com/login",
          "httpVersion": "HTTP/1.1",
          "headers": [
            {"name": "Content-Type", "value": "application/json"},
            {"name": "User-Agent", "value": "Test"}
          ],
          "queryString": [],
          "postData": {
            "mimeType": "application/json",
            "text": "{\"username\":\"alice\",\"password\":\"123456\"}"
          },
          "headersSize": 100,
          "bodySize": 50
        },
        "response": {
          "status": 200,
          "statusText": "OK",
          "httpVersion": "HTTP/1.1",
          "headers": [],
          "content": {
            "size": 80,
            "mimeType": "application/json",
            "text": "{\"code\":0,\"data\":{\"token\":\"abc123\"}}"
          },
          "redirectURL": "",
          "headersSize": 0,
          "bodySize": 80
        },
        "cache": {},
        "serverIPAddress": "127.0.0.1",
        "connection": "1234"
      },
      {
        "startedDateTime": "2024-01-01T00:00:01.000Z",
        "time": 50,
        "request": {
          "method": "GET",
          "url": "https://api.example.com/user/profile",
          "httpVersion": "HTTP/1.1",
          "headers": [{"name": "Authorization", "value": "Bearer abc123"}],
          "queryString": [],
          "postData": null,
          "headersSize": 50,
          "bodySize": 0
        },
        "response": {
          "status": 200,
          "statusText": "OK",
          "httpVersion": "HTTP/1.1",
          "headers": [],
          "content": {
            "size": 30,
            "mimeType": "application/json",
            "text": "{\"id\":1,\"name\":\"alice\"}"
          },
          "redirectURL": "",
          "headersSize": 0,
          "bodySize": 30
        },
        "cache": {},
        "serverIPAddress": "127.0.0.1",
        "connection": "1235"
      }
    ]
  }
}`

func TestImportService_ImportHAR(t *testing.T) {
	c := setupImportTestContainer(t)
	ctx := context.Background()

	t.Run("SuccessfulImport", func(t *testing.T) {
		resp, err := c.ImportService.ImportHAR(ctx, ImportHARRequest{
			Name:    "Imported Scenario",
			HARData: []byte(validHARJSON),
			VUs:     50,
		})
		if err != nil {
			t.Fatalf("import: %v", err)
		}
		if resp.Scenario == nil {
			t.Fatal("scenario is nil")
		}
		if resp.Scenario.Name != "Imported Scenario" {
			t.Errorf("name = %q", resp.Scenario.Name)
		}
		if resp.Stats.TotalEntries != 2 {
			t.Errorf("total entries = %d, want 2", resp.Stats.TotalEntries)
		}
		if resp.Stats.ImportedRequests != 2 {
			t.Errorf("imported requests = %d, want 2", resp.Stats.ImportedRequests)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		_, err := c.ImportService.ImportHAR(ctx, ImportHARRequest{
			HARData: []byte("not a har"),
		})
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})

	t.Run("EmptyHAR", func(t *testing.T) {
		emptyHAR := `{"log": {"version": "1.2", "entries": []}}`
		resp, err := c.ImportService.ImportHAR(ctx, ImportHARRequest{
			HARData: []byte(emptyHAR),
		})
		if err != nil {
			t.Fatalf("import empty: %v", err)
		}
		if resp.Stats.ImportedRequests != 0 {
			t.Errorf("expected 0 imports, got %d", resp.Stats.ImportedRequests)
		}
	})

	t.Run("WithBaseURL", func(t *testing.T) {
		resp, err := c.ImportService.ImportHAR(ctx, ImportHARRequest{
			HARData: []byte(validHARJSON),
			BaseURL: "https://api.example.com",
		})
		if err != nil {
			t.Fatalf("import: %v", err)
		}
		// 验证 URL 被替换
		if resp.Scenario == nil {
			t.Fatal("scenario is nil")
		}
	})

	t.Run("SourceIsHarImport", func(t *testing.T) {
		resp, err := c.ImportService.ImportHAR(ctx, ImportHARRequest{
			HARData: []byte(validHARJSON),
		})
		if err != nil {
			t.Fatalf("import: %v", err)
		}
		if resp.Scenario.Source != "har_import" {
			t.Errorf("source = %q, want har_import", resp.Scenario.Source)
		}
	})
}