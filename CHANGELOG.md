# Changelog

All notable changes to Pulse will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Planned
- gRPC protocol support
- WebSocket protocol support
- Chrome browser recording plugin
- Performance baseline comparison
- Command-line tool (CLI)

---

## [0.1.0] - 2024-12-15 — MVP 🎉

### Added

#### 场景管理
- 场景 CRUD(创建、编辑、删除、复制)
- 场景编辑器(三栏布局:请求列表 / 请求详情 / 压测配置)
- 多请求场景支持(增 / 删 / 复制 / 排序)
- Headers 键值对编辑器(启用开关 + 排序)
- Body 编辑器(JSON / Raw / Form 三种类型,带实时校验)
- Extractors 响应提取(JSON Path → 变量)
- Assertions 断言(状态码 / 延迟 / Body 包含)
- Variables 变量定义(支持数据驱动测试)
- Thresholds 阈值检查(P95 / P99 / 错误率 / 最小 RPS)
- 实时 JSON Schema 校验
- 多场景管理(列表 / 卡片双视图)
- 搜索 / 标签筛选 / 批量操作
- 场景版本管理(乐观锁)
- HAR 文件导入(支持 Chrome DevTools HAR 1.2 标准)

#### 压测引擎 ⭐
- 自研 Go 压测引擎(无外部依赖)
- Goroutine Pool(动态扩缩,0-20,000 VUs)
- 优化的 HTTP Client(连接池 + Keep-Alive + HTTP/2)
- Token Bucket 限流器
- HDR Histogram 延迟分位数采集
- Sliding Window RPS 计算
- Scheduler(支持 Linear / Step / Wave Ramp-up)
- VU 执行循环(变量解析 + 提取器 + 断言 + Think Time)
- Local Engine 实现 Engine 接口(预留 Remote 接口)
- 单机 10,000+ VUs 稳定运行

#### 实时监控
- 5 个 KPI 卡片(RPS / P95 / 错误率 / 总请求 / VUs)
- 实时折线图(RPS + P50 + P95 曲线)
- 状态码分布(2xx / 3xx / 4xx / 5xx 颜色编码)
- Per-Request 明细表
- 进度条 + 实时计时
- 阈值告警
- 全屏模式(F 键)
- 断线检测

#### 报告与历史
- 自动报告生成
- 5 个 KPI 汇总卡
- 阈值检查清单
- Per-Request 详细统计
- 历史报告列表(搜索 + 状态筛选)
- HTML 报告导出(预留完整模板)

#### 桌面应用
- 系统托盘(显示运行状态)
- 桌面通知
- 暗色 / 亮色主题切换
- 中英双语切换
- 命令面板(⌘K / Ctrl+K)
- 快捷键(⌘S 保存 / ⌘⇧R 运行 / F 全屏 / ⌘I 导入 HAR)

#### 数据管理
- SQLite + DuckDB 双数据库
- GORM 自动迁移
- 本地存储(无需云服务)

#### 开发工具
- 60+ 个单元 / 集成测试
- 性能基准测试套件
- GitHub Actions CI / CD
- 三平台自动打包(macOS / Windows / Linux)

### Technical Details

#### 架构
- **后端**:Go 1.22 + Wails v2 + GORM
- **前端**:Vue 3 + Vite + TypeScript + Naive UI + ECharts
- **桌面框架**:Wails v2(原生 WebView)
- **数据库**:SQLite(元数据)+ DuckDB(预留时序指标)
- **架构模式**:Controller-Agent(预留)→ 单一 Local Engine(MVP)

#### 性能(基线)
- 单机 10,000+ VUs 稳定运行
- localhost RPS:50,000+(Hello World)
- P95 延迟:< 5ms (localhost)
- 内存占用:~100-150 MB(1k VUs)

#### 代码统计
- Go:47 个文件,约 8,000 行
- Vue/TS:21 个文件,约 6,500 行
- 文档 + 测试:约 1,000 行
- 总计:约 15,500 行

### Known Limitations
- macOS / Windows 代码签名未实现(Gatekeeper / SmartScreen 警告)
- DuckDB 集成预留,时序指标暂存内存
- 分布式 Agent 仅预留接口,未实现
- Chrome 录制插件未实现,只支持 HAR 文件导入

### Dependencies

#### 后端
- github.com/wailsapp/wails/v2 v2.10.1
- gorm.io/gorm v1.25.10
- github.com/glebarez/sqlite v1.11.0
- github.com/HdrHistogram/hdrhistogram-go v1.6.0
- github.com/marcboeker/go-duckdb v1.7.1
- golang.org/x/time v0.5.0
- github.com/google/uuid v1.6.0

#### 前端
- vue@3.4+
- vue-router@4.3+
- pinia@2.1+
- naive-ui@2.38+
- echarts@5.5+
- vue-echarts@7.0+
- vue-i18n@9.13+
- @phosphor-icons/vue@2.2+

---

## [Pre-release]

### 0.0.x 开发期

- 0.0.1 - 2024-10 - 项目初始化
- 0.0.2 - 2024-10 - Wails + Vue3 脚手架
- 0.0.3 - 2024-11 - 场景数据层 + Repository
- 0.0.4 - 2024-11 - 场景编辑器基础版
- 0.0.5 - 2024-11 - 自研引擎核心组件
- 0.0.6 - 2024-12 - 引擎调度器 + VU 执行循环
- 0.0.7 - 2024-12 - Run Service + 实时大屏
- 0.0.8 - 2024-12 - HAR 导入 + 高级编辑器
- 0.0.9 - 2024-12 - 性能基准 + 文档完善
- 0.1.0 - 2024-12-15 - 正式 MVP 发布 🎉

---

## 版本说明

- **主版本号**(Major):不兼容的 API 变更
- **次版本号**(Minor):向下兼容的新功能
- **修订号**(Patch):向下兼容的 bug 修复

预发布版本使用 `0.x.y` 格式,API 可能在 1.0 之前变更。

---

[Unreleased]: https://github.com/pulse/pulse/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/pulse/pulse/releases/tag/v0.1.0