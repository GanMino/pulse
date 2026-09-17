# ==============================================
# Pulse - Makefile
# ==============================================
# 常用命令:
#   make dev          - 开发模式
#   make build        - 构建当前平台
#   make build-all    - 构建所有平台
#   make test         - 运行测试
#   make lint         - Lint 检查
#   make clean        - 清理构建产物
#   make release      - 发布新版本(打 tag + push)
# ==============================================

# 变量
BIN_NAME := pulse
VERSION := $(shell grep '"version"' frontend/package.json | head -1 | sed -E 's/.*"version":\s*"([^"]+)".*/\1/')
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -s -w \
	-X 'github.com/pulse/pulse/internal/config.App.Version=$(VERSION)' \
	-X 'github.com/pulse/pulse/internal/config.App.BuildTime=$(BUILD_TIME)' \
	-X 'github.com/pulse/pulse/internal/config.App.GitCommit=$(GIT_COMMIT)'

.PHONY: help install dev build build-mac build-win build-linux build-all \
	test test-backend test-frontend bench lint format clean release \
	release-tag run help

# ==============================================
# 默认目标
# ==============================================

help:  ## 显示帮助
	@echo ""
	@echo "  Pulse v$(VERSION) - Makefile"
	@echo ""
	@echo "  Usage:"
	@echo "    make <target>"
	@echo ""
	@echo "  Targets:"
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) { \
			printf "    \033[36m%-20s\033[0m %s\n", $$1, $$2 \
		} \
	}' $(MAKEFILE_LIST)
	@echo ""

# ==============================================
# 开发
# ==============================================

install:  ## 安装所有依赖
	@echo "→ 安装 Go 模块..."
	go mod download
	@echo "→ 安装前端依赖..."
	cd frontend && pnpm install
	@echo "✓ 所有依赖安装完成"

dev:  ## 开发模式(带热重载)
	wails dev

run:  ## 别名: 等同于 dev
	wails dev

# ==============================================
# 构建
# ==============================================

build:  ## 构建当前平台
	wails build -ldflags "$(LDFLAGS)"

build-mac:  ## 构建 macOS(通用二进制)
	wails build -platform darwin/universal -ldflags "$(LDFLAGS)"

build-mac-intel:  ## 构建 macOS Intel
	wails build -platform darwin/amd64 -ldflags "$(LDFLAGS)"

build-mac-arm:  ## 构建 macOS Apple Silicon
	wails build -platform darwin/arm64 -ldflags "$(LDFLAGS)"

build-win:  ## 构建 Windows
	wails build -platform windows/amd64 -ldflags "$(LDFLAGS)"

build-linux:  ## 构建 Linux
	wails build -platform linux/amd64 -ldflags "$(LDFLAGS)"

build-all: build-mac build-win build-linux  ## 构建所有平台
	@echo "✓ 所有平台构建完成"
	@ls -la build/bin/

# ==============================================
# 测试
# ==============================================

test: test-backend test-frontend  ## 运行所有测试

test-backend:  ## 运行后端测试
	go test -race -timeout 120s ./internal/...

test-frontend:  ## 运行前端测试
	cd frontend && pnpm test

test-coverage:  ## 生成测试覆盖率报告
	go test -race -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ 覆盖率报告已生成: coverage.html"

# ==============================================
# 性能基准
# ==============================================

bench:  ## 运行性能基准
	go test ./internal/engine/ -bench=. -benchmem -benchtime=10s

bench-cpu:  ## 生成 CPU profile
	go test ./internal/engine/ -bench=BenchmarkEngine -cpuprofile=cpu.prof -benchtime=30s
	go tool pprof -text cpu.prof | head -30

bench-mem:  ## 生成内存 profile
	go test ./internal/engine/ -bench=BenchmarkEngine -memprofile=mem.prof -benchtime=30s
	go tool pprof -text mem.prof | head -30

# ==============================================
# 代码质量
# ==============================================

lint:  ## 运行所有 Linter
	golangci-lint run ./...
	cd frontend && pnpm lint

lint-backend:  ## 仅后端 Lint
	golangci-lint run ./...

lint-frontend:  ## 仅前端 Lint
	cd frontend && pnpm lint

format:  ## 格式化代码
	gofmt -w .
	cd frontend && pnpm format

format-check:  ## 检查格式(不修改)
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "✗ Go 代码格式不正确"; \
		gofmt -l .; \
		exit 1; \
	fi
	cd frontend && pnpm format:check

# ==============================================
# 清理
# ==============================================

clean:  ## 清理构建产物
	rm -rf build/bin/
	rm -rf frontend/dist/
	rm -rf frontend/node_modules/
	rm -f coverage.out coverage.html
	rm -f cpu.prof mem.prof trace.out
	go clean -cache
	@echo "✓ 清理完成"

# ==============================================
# 发布
# ==============================================

release-tag:  ## 创建并推送 release tag(自动触发 GitHub Actions)
	@if [ -z "$(VERSION)" ]; then \
		echo "✗ 版本号为空"; \
		exit 1; \
	fi
	@echo "→ 创建 tag v$(VERSION)..."
	git tag -a "v$(VERSION)" -m "Release v$(VERSION)"
	@echo "→ 推送 tag 到 GitHub..."
	git push origin "v$(VERSION)"
	@echo "✓ Tag v$(VERSION) 已推送"
	@echo "→ GitHub Actions 会自动构建并发布"
	@echo "→ 查看: https://github.com/$(shell git remote get-url origin | sed -E 's|.*github.com[:/]([^/]+)/([^/.]+)(\.git)?|\1/\2|')/actions"

# ==============================================
# Wails 特定
# ==============================================

wails-doctor:  ## 检查 Wails 环境
	wails doctor

wails-generate:  ## 重新生成 Wails JS 绑定
	wails generate module

# ==============================================
# 开发辅助
# ==============================================

deps-update:  ## 更新所有依赖到最新
	go get -u ./...
	cd frontend && pnpm update

tree:  ## 显示项目目录结构
	@find . -type d -not -path '*/node_modules*' -not -path '*/.git*' -not -path '*/dist*' -not -path '*/build/bin*' | sort | head -30

stats:  ## 代码统计
	@echo "代码统计:"
	@echo "  Go 文件:    $$(find . -name '*.go' -not -path './frontend/*' | wc -l | tr -d ' ')"
	@echo "  Vue 文件:   $$(find . -name '*.vue' -not -path './frontend/wailsjs/*' | wc -l | tr -d ' ')"
	@echo "  TS 文件:    $$(find . -name '*.ts' -not -path './frontend/wailsjs/*' | wc -l | tr -d ' ')"
	@echo "  测试文件:    $$(find . -name '*_test.go' | wc -l | tr -d ' ')"
	@echo "  总代码行数:  $$(find . -type f \( -name '*.go' -o -name '*.vue' -o -name '*.ts' -o -name '*.md' \) -not -path './frontend/wailsjs/*' -exec cat {} \; | wc -l | tr -d ' ')"