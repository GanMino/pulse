<div align="center">

# 💓 Pulse

**Modern open-source load testing tool — desktop app for developers.**

压力测试,理应如此。

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![Vue](https://img.shields.io/badge/Vue-3.4+-4FC08D?logo=vue.js)](https://vuejs.org/)
[![Wails](https://img.shields.io/badge/Wails-v2-red)](https://wails.io/)
[![GitHub Stars](https://img.shields.io/github/stars/pulse/pulse?style=social)](https://github.com/pulse/pulse)

[**快速开始**](#-快速开始) · [**功能特性**](#-功能特性) · [**架构设计**](#-架构设计) · [**路线图**](#-路线图) · [**贡献**](#-贡献)

</div>

---

## 📑 目录

- [项目简介](#-项目简介)
- [快速开始](#-快速开始)
- [功能特性](#-功能特性)
- [架构设计](#-架构设计)
- [使用示例](#-使用示例)
- [性能基准](#-性能基准)
- [路线图](#-路线图)
- [贡献](#-贡献)
- [社区](#-社区)
- [许可证](#-许可证)

---

## ✨ 项目简介

**Pulse** 是一款现代化的开源桌面压测工具,致力于让"打开应用 → 配置场景 → 启动压测 → 查看报告"这件事在 5 分钟内完成。

我们用 Wails + Vue 3 + 自研 Go 引擎重新定义了压测工具的体验:

- 🎨 **现代 UI**:Vue 3 + Naive UI,深色主题,媲美 Apifox / Postman
- ⚡ **极致性能**:自研 Go 引擎,单机支持 10,000+ VUs,50k+ RPS
- 📊 **实时大屏**:2 秒刷新率的实时指标大屏,KPI / 折线图 / 状态码
- 📥 **HAR 导入**:从浏览器录制一键生成压测场景(Chrome / Firefox / Edge)
- 🚀 **5 分钟上手**:不需要 Java,不需要写代码,可视化拖拽配置
- 🌐 **跨平台**:macOS / Windows / Linux 三端,单一二进制
- 🔒 **本地优先**:所有数据本地存储,无需注册,无需云服务
- 📦 **开源**:Apache 2.0 协议,商用友好

### 对比传统工具

| 特性 | Pulse | JMeter | k6 | Locust |
|---|---|---|---|---|
| **上手时间** | **5 分钟** | 60 分钟 | 10 分钟 | 20 分钟 |
| **UI 体验** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐(CLI) | ⭐⭐⭐ |
| **单机并发** | 10,000+ VUs | 5,000 VUs | 50,000+ VUs | 5,000 VUs |
| **协议支持** | HTTP/HTTPS | 全协议 | HTTP/gRPC | 自定义 |
| **HAR 导入** | ✅ | ⚠️ 需代理 | ❌ | ❌ |
| **实时大屏** | ✅ 现代 | ⚠️ 老式 | ⚠️ Cloud 付费 | ⚠️ 基础 |
| **部署** | 单二进制 | Java + GUI | CLI + Docker | Python + Web |
| **跨平台** | macOS/Win/Linux | ✅ | ✅ | ✅ |
| **开源协议** | Apache 2.0 | Apache 2.0 | AGPL | MIT |

---

## 🚀 快速开始

### 前置要求

- **Go** 1.22+ [下载](https://go.dev/dl/)
- **Node.js** 20+ [下载](https://nodejs.org/)
- **pnpm** 9+ (`npm install -g pnpm`)
- **Wails CLI** v2.10+:
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.1
  ```
- **平台依赖**:
  - **macOS**: Xcode Command Line Tools (`xcode-select --install`)
  - **Linux**: `gcc`, `libgtk-3-dev`, `libwebkit2gtk-4.0-dev` 或 `libwebkit2gtk-4.1-dev`
  - **Windows**: WebView2 Runtime(Win11 自带,Win10 需安装)

### 安装

```bash
# 1. 克隆仓库
git clone https://github.com/pulse/pulse.git
cd pulse

# 2. 安装依赖
go mod download
cd frontend && pnpm install && cd ..

# 3. 验证环境
wails doctor

# 4. 开发模式启动(带热重载)
wails dev
```

应用窗口会自动打开。开始你的第一次压测!

### 构建发布版

```bash
# 当前平台
wails build

# macOS 通用二进制
wails build -platform darwin/universal

# Windows
wails build -platform windows/amd64

# Linux
wails build -platform linux/amd64

# 或使用 Makefile
make build         # 当前平台
make build-mac     # macOS
make build-win     # Windows
make build-linux   # Linux
```

构建产物在 `build/bin/` 目录:
- `pulse.app` / `pulse.exe` / `pulse`(二进制)
- `pulse.dmg` / `pulse.msi` / `pulse.AppImage`(安装包)

### 下载预编译版本

访问 [Releases 页面](https://github.com/pulse/pulse/releases) 下载对应平台的最新版本。

| 平台 | 下载 |
|---|---|
| macOS (Apple Silicon + Intel) | [pulse-macos-universal.dmg](https://github.com/pulse/pulse/releases/latest) |
| Windows (x64) | [pulse-windows-amd64.msi](https://github.com/pulse/pulse/releases/latest) |
| Linux (x64) | [pulse-linux-amd64.AppImage](https://github.com/pulse/pulse/releases/latest) |

---

## 🎯 功能特性

### 场景管理
- ✅ 可视化场景编辑器(三栏布局:请求列表 / 请求详情 / 压测配置)
- ✅ 多请求场景支持(增 / 删 / 复制 / 排序)
- ✅ Headers 键值对编辑器(启用开关 + 排序)
- ✅ Body 编辑器(JSON / Raw / Form 三种类型)
- ✅ Extractors 响应提取(JSON Path → 变量)
- ✅ Assertions 断言(状态码 / 延迟 / Body 包含)
- ✅ Variables 变量定义(支持数据驱动测试)
- ✅ Thresholds 阈值检查(P95 / P99 / 错误率 / 最小 RPS)
- ✅ 实时 JSON 校验
- ✅ 多场景管理(列表 / 卡片双视图)
- ✅ 搜索 / 标签筛选 / 批量操作
- ✅ HAR 文件导入(Chrome / Firefox / Edge DevTools 导出)
- ✅ 场景复制 / 版本管理

### 压测执行
- ✅ 自研 Go 压测引擎(无外部依赖)
- ✅ HTTP / HTTPS / HTTP/2 支持
- ✅ 连接池 + Keep-Alive 优化
- ✅ Ramp-up 模式:Linear / Step / Wave
- ✅ 目标 RPS 限流(Token Bucket)
- ✅ Think Time 等待
- ✅ 变量提取 + 跨请求传递
- ✅ 断言失败计入错误
- ✅ 单机 20,000 VUs 上限
- ✅ 暂停 / 恢复 / 提前停止

### 实时监控
- ✅ 5 个 KPI 卡片(RPS / P95 / 错误率 / 总请求 / VUs)
- ✅ 实时折线图(RPS + P50 + P95 曲线)
- ✅ 状态码分布(2xx / 3xx / 4xx / 5xx 颜色编码)
- ✅ Per-Request 明细表(每个请求的 RPS / 延迟 / 错误率)
- ✅ 进度条 + 实时计时
- ✅ 全屏模式(F 键)
- ✅ 阈值告警(超过阈值时变红 / 黄)
- ✅ 断线检测

### 报告与历史
- ✅ 自动生成报告(测试完成时)
- ✅ 5 个 KPI 汇总卡
- ✅ 阈值检查清单(✅ / ❌)
- ✅ Per-Request 详细统计
- ✅ 历史报告列表(搜索 + 状态筛选)
- ✅ 报告对比(同一场景多次运行的对比)

### 桌面应用特性
- ✅ 系统托盘(显示运行状态)
- ✅ 桌面通知(测试完成 / 失败时)
- ✅ 暗色 / 亮色主题切换
- ✅ 中英双语切换
- ✅ 命令面板(⌘K / Ctrl+K)
- ✅ 快捷键(⌘S 保存 / ⌘⇧R 运行 / F 全屏)

### 数据管理
- ✅ 本地 SQLite + DuckDB 双数据库
- ✅ 自动迁移 / 数据备份
- ✅ 旧数据清理(>30天)

---

## 🏗 架构设计

### 三层架构

```
┌──────────────────────────────────────────────────────────┐
│                  Pulse 桌面应用                          │
│                                                          │
│  ┌────────────────────────────────────┐                  │
│  │  WebView (Chromium WebKit)         │                  │
│  │  ├── Vue 3 + Vite + TypeScript     │                  │
│  │  ├── Naive UI(组件库)              │                  │
│  │  ├── ECharts(图表)                  │                  │
│  │  └── Pinia(状态管理)                │                  │
│  └────────────────────────────────────┘                  │
│                       ↕ (Wails IPC)                      │
│  ┌────────────────────────────────────┐                  │
│  │  Go Runtime                         │                  │
│  │  ├── App struct (绑定方法)          │                  │
│  │  ├── Service Layer                  │                  │
│  │  │   ├── ScenarioService            │                  │
│  │  │   ├── ProjectService             │                  │
│  │  │   ├── RunService ⭐              │                  │
│  │  │   ├── ReportService              │                  │
│  │  │   ├── ImportService              │                  │
│  │  │   └── AgentService               │                  │
│  │  ├── Repository Layer               │                  │
│  │  │   ├── SQLite (元数据)            │                  │
│  │  │   └── DuckDB (时序指标,预留)     │                  │
│  │  └── ⭐ 自研压测引擎                │                  │
│  │      ├── Goroutine Pool             │                  │
│  │      ├── HTTP Client (优化)         │                  │
│  │      ├── Token Bucket (限流)        │                  │
│  │      ├── HDR Histogram (延迟)       │                  │
│  │      ├── Scheduler (Ramp-up)        │                  │
│  │      └── VU Execution Loop          │                  │
│  └────────────────────────────────────┘                  │
│                       ↕                                   │
│  ┌────────────────────────────────────┐                  │
│  │  Storage (本地)                     │                  │
│  │  ├── SQLite: ~/.pulse/data/pulse.db│                  │
│  │  ├── DuckDB: ~/.pulse/data/        │                  │
│  │  └── Reports: ~/.pulse/reports/    │                  │
│  └────────────────────────────────────┘                  │
└──────────────────────────────────────────────────────────┘
```

### 模块划分

```
internal/
├── config/      配置管理(Viper)
├── event/       事件总线抽象
├── model/       领域模型(GORM)
├── repository/  数据访问层
│   └── sqlite/  SQLite + GORM 实现
├── service/     业务逻辑层
│   ├── container.go      依赖注入容器
│   ├── scenario_service.go
│   ├── project_service.go
│   ├── run_service.go    ⭐ 压测执行服务
│   ├── report_service.go
│   ├── import_service.go
│   ├── agent_service.go
│   ├── user_service.go
│   └── test_run_service.go
└── engine/      ⭐ 自研压测引擎
    ├── engine.go        接口定义
    ├── scheduler.go      Ramp-up 调度
    ├── vu.go            VU 执行循环
    ├── local.go         Local Engine 主实现
    ├── pool.go          Goroutine Pool
    ├── httpclient.go    优化 HTTP 客户端
    ├── ratelimit.go     Token Bucket
    ├── variables.go     变量替换器
    ├── extract.go       JSON Path 提取
    ├── assert.go        断言执行
    └── metrics/         指标采集
        ├── histogram.go HDR Histogram
        ├── collector.go 聚合收集器
        └── window.go    滑动窗口
```

---

## 📸 使用示例

### 方式 A:手动创建场景

```
1. 点击 "+ New Scenario"
2. 配置场景:
   - Method: POST
   - URL: https://api.example.com/login
   - Headers: Content-Type: application/json
   - Body: { "username": "alice", "password": "123456" }
3. 配置压测参数:
   - VUs: 100
   - Duration: 5m
   - Ramp-up: Linear 30s
4. 点击 "▶ Save & Run" (⌘⇧R)
5. 自动跳转到实时大屏
```

### 方式 B:从 HAR 导入(推荐!)

```
1. Chrome DevTools → Network 标签
2. 录制你的用户操作
3. 右键 → "Save all as HAR with content"
4. Pulse → "📥 导入 HAR" → 选择文件
5. 自动生成多请求场景
6. 调整 VUs / Duration
7. 一键运行
```

### 方式 C:配置变量(数据驱动测试)

```json
{
  "variables": {
    "user": ["alice", "bob", "charlie", "diana"],
    "pass": "123456"
  },
  "requests": [
    {
      "name": "登录",
      "method": "POST",
      "url": "/api/login",
      "body": {
        "username": "{{user}}",
        "password": "{{pass}}"
      }
    }
  ]
}
```

### 方式 D:配置断言

```json
{
  "requests": [
    {
      "name": "下单",
      "url": "/api/order",
      "assertions": {
        "status": [200, 201],
        "maxLatencyMs": 500,
        "bodyContains": ["success"]
      },
      "extractors": {
        "orderId": "$.data.id"
    }
  ]
}
```

---

## 📊 性能基准

### localhost 测试(M2 MacBook Pro 基线)

| 场景 | VUs | 预期 RPS | 预期 P95 | 预期 P99 |
|---|---|---|---|---|
| **Hello World** | 100 | 50,000+ | < 1ms | < 5ms |
| **JSON API** | 500 | 30,000+ | < 5ms | < 10ms |
| **POST Login** | 200 | 20,000+ | < 10ms | < 20ms |
| **高并发** | 1,000 | 50,000+ | < 5ms | < 15ms |

### 跑测试

```bash
# 单元 + 集成测试
go test ./internal/engine/... -v -timeout 60s

# 性能基准
go test ./internal/engine/ -bench=. -benchmem -benchtime=10s

# CPU profile
go test ./internal/engine/ -bench=BenchmarkEngine_HighConcurrency -cpuprofile=cpu.prof
go tool pprof -http=:8080 cpu.prof
```

详细性能分析见 [BENCHMARK.md](docs/BENCHMARK.md) 和 [PERFORMANCE.md](docs/PERFORMANCE.md)。

---

## 🛣 路线图

### v0.1.0 — MVP (当前)
- [x] 场景管理 CRUD + 编辑器
- [x] HAR 文件导入
- [x] 自研 Go 压测引擎
- [x] 实时监控大屏
- [x] 报告生成与导出
- [x] 跨平台打包
- [x] Apache 2.0 开源

### v0.2.0 — 协议扩展(3 个月后)
- [ ] gRPC 协议支持
- [ ] WebSocket 协议支持
- [ ] Chrome 浏览器录制插件
- [ ] 性能基线对比
- [ ] 命令行工具(CLI)
- [ ] 模板市场

### v0.3.0 — 分布式(6 个月后)
- [ ] 分布式 Agent(基于预留接口)
- [ ] macOS / Windows 代码签名
- [ ] AI 辅助生成压测脚本(GPT 接入)
- [ ] Pulse Cloud(SaaS)

### v1.0.0 — 完整版(12 个月后)
- [ ] 生产环境全链路压测
- [ ] 完整企业版功能(SSO / RBAC / 审计)
- [ ] 性能基线 + 自动告警
- [ ] 插件市场 + 模板社区

---

## 🤝 贡献

我们欢迎所有形式的贡献!详情见 [CONTRIBUTING.md](CONTRIBUTING.md)。

### 贡献方式

| 方式 | 说明 |
|---|---|
| 🐛 **报告 Bug** | 在 [Issues](https://github.com/pulse/pulse/issues) 提交 |
| 💡 **提出功能** | 在 [Discussions](https://github.com/pulse/pulse/discussions) 讨论 |
| 📝 **改进文档** | 直接 PR |
| 🔨 **提交代码** | Fork → 修改 → PR |
| 🌍 **翻译** | 帮助翻译界面文案 |
| ⭐ **点亮 Star** | 你的支持是我们最大的动力 |

### 开发流程

```bash
# 1. Fork 仓库

# 2. 克隆你的 fork
git clone https://github.com/YOUR_USERNAME/pulse.git
cd pulse

# 3. 创建功能分支
git checkout -b feature/amazing-feature

# 4. 开发(开发模式带热重载)
wails dev

# 5. 运行测试
go test ./...
cd frontend && pnpm lint

# 6. 提交(推荐使用 Conventional Commits)
git commit -m "feat: 添加 HAR 导入功能"

# 7. 推送到你的 fork
git push origin feature/amazing-feature

# 8. 在 GitHub 创建 Pull Request
```

### 代码风格

- **Go**: `gofmt` + `golangci-lint`
- **TypeScript**: ESLint + Prettier
- **Vue**: Composition API + `<script setup>`
- **提交规范**: [Conventional Commits](https://www.conventionalcommits.org/)

---

## 💬 社区

- **GitHub Issues**: [问题反馈](https://github.com/pulse/pulse/issues)
- **GitHub Discussions**: [功能讨论](https://github.com/pulse/pulse/discussions)
- **Discord**: [加入社区](https://discord.gg/pulse)
- **Twitter**: [@pulse_dev](https://twitter.com/pulse_dev)

---

## 🙏 致谢

Pulse 的诞生离不开以下优秀的开源项目:

- [Wails](https://wails.io) — Go + Web 桌面应用框架
- [Vue 3](https://vuejs.org) — 前端框架
- [Naive UI](https://www.naiveui.com) — Vue 3 UI 组件库
- [ECharts](https://echarts.apache.org) — 数据可视化
- [Vite](https://vitejs.dev) — 前端构建工具
- [Pinia](https://pinia.vuejs.org) — Vue 状态管理
- [GORM](https://gorm.io) — Go ORM
- [HDR Histogram](https://github.com/HdrHistogram/HdrHistogram) — 延迟分位数
- [SQLite](https://www.sqlite.org) — 嵌入式数据库
- [DuckDB](https://duckdb.org) — 嵌入式 OLAP 数据库

特别感谢 [k6](https://k6.io)、[Locust](https://locust.io)、[JMeter](https://jmeter.apache.org)、[nGrinder](https://naver.github.io/ngrinder/) 等前辈工具的启发。

---

## 📄 许可证

本项目采用 [Apache License 2.0](LICENSE) 协议。

```
Copyright 2024 Pulse Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
```

商用、二次开发、私有化部署都可以,只需要保留版权声明。

---

## 📊 项目状态

```
✅ MVP v0.1.0 - 功能完成,准备发布
⏳ v0.2.0 - 协议扩展(规划中)
⏳ v0.3.0 - 分布式(规划中)
⏳ v1.0.0 - 完整版(规划中)
```

**当前代码量**:约 15,500 行(Go + Vue + TS + Markdown)
**测试覆盖**:60+ 个测试用例
**支持平台**:macOS / Windows / Linux

---

<div align="center">

**如果这个项目对你有帮助,请给我们一个 ⭐️ !**

[⬆ 回到顶部](#-pulse)

</div>