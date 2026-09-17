# Contributing to Pulse

首先,感谢你考虑为 Pulse 做贡献! 🎉

Pulse 是一个开源项目,我们欢迎任何形式的贡献,无论是代码、文档、bug 报告还是功能建议。

## 📋 目录

- [行为准则](#-行为准则)
- [我能贡献什么?](#-我能贡献什么)
- [开发流程](#-开发流程)
- [环境准备](#-环境准备)
- [代码风格](#-代码风格)
- [提交规范](#-提交规范)
- [Pull Request 流程](#-pull-request-流程)
- [测试要求](#-测试要求)
- [发布流程](#-发布流程)

---

## 🤝 行为准则

### 我们的承诺

为了营造一个开放和友好的环境,我们承诺:

- 🤗 欢迎所有贡献者,无论经验水平、性别、性取向、残疾、外貌、身材、种族或国籍
- 💬 使用友好和包容的语言
- 🎯 尊重不同的观点和经验
- 🙏 优雅地接受建设性批评
- 🌟 关注对社区最有利的事情

### 不可接受的行为

- ❌ 使用性化的语言或图像
- ❌ 人身攻击、侮辱/贬损评论
- ❌ 公开或私下的骚扰
- ❌ 未经许可发布他人的私人信息
- ❌ 其他不道德或不专业的行为

---

## 🎁 我能贡献什么?

| 类型 | 说明 | 入口 |
|---|---|---|
| 🐛 **报告 Bug** | 发现问题时提交 issue | [New Issue](https://github.com/pulse/pulse/issues/new?template=bug.md) |
| 💡 **功能请求** | 提议新功能或改进 | [New Issue](https://github.com/pulse/pulse/issues/new?template=feature.md) |
| ❓ **提问** | 使用问题、配置问题 | [Discussions](https://github.com/pulse/pulse/discussions) |
| 📝 **改进文档** | 修正错别字、补充说明 | 直接 PR |
| 🔨 **修复 Bug** | 解决 issue 中的 bug | Fork → 修改 → PR |
| ⚡ **新功能** | 实现 feature request | 先在 issue 中讨论 |
| 🌐 **翻译** | 翻译界面 / 文档 | 直接 PR |
| ⭐ **Star / 分享** | 让更多人知道 Pulse | - |
| 💰 **赞助** | 资助项目发展 | [GitHub Sponsors](https://github.com/sponsors/pulse) |

---

## 🔧 开发流程

### 推荐流程

```
1. Issue → 检查是否已有相关 issue
   ↓
2. Discussion → 与维护者讨论方案(对大改动)
   ↓
3. Fork → 创建你的分支
   ↓
4. Develop → wails dev(热重载)
   ↓
5. Test → go test + pnpm test
   ↓
6. Lint → gofmt + golangci-lint + ESLint
   ↓
7. Commit → 遵循 Conventional Commits
   ↓
8. Push → 推送到你的 fork
   ↓
9. PR → 创建 Pull Request
   ↓
10. Review → 与维护者讨论修改
    ↓
11. Merge → 🎉 你的代码上线了!
```

---

## 🛠 环境准备

### 前置依赖

- **Go**: 1.22+
- **Node.js**: 20+
- **pnpm**: 9+ (`npm install -g pnpm`)
- **Wails CLI**: v2.10+
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.1
  ```
- **Git**: 任意版本
- **平台依赖**:
  - **macOS**: `xcode-select --install`
  - **Linux**: `gcc`, `libgtk-3-dev`, `libwebkit2gtk-4.0-dev` 或 `libwebkit2gtk-4.1-dev`
  - **Windows**: WebView2 Runtime

### 验证环境

```bash
go version      # go version go1.22+
node --version  # v20+
pnpm --version  # 9+
wails version   # v2.10.x
wails doctor    # 检查所有依赖
```

### 本地启动

```bash
git clone https://github.com/YOUR_USERNAME/pulse.git
cd pulse
go mod download
cd frontend && pnpm install && cd ..

# 启动开发模式(带热重载)
wails dev
```

应用窗口会自动打开。修改 Go / Vue 代码会自动重载。

---

## 📝 代码风格

### Go 代码

- 使用 `gofmt` 格式化
- 使用 `golangci-lint` 检查(配置见 `.golangci.yml`)
- 遵循 [Effective Go](https://go.dev/doc/effective_go)
- 函数和变量使用驼峰命名,导出符号大写开头
- 添加清晰的注释(中文 / 英文均可,保持一致)
- 公开 API 添加 godoc 注释

```go
// CalculateSum 计算两个整数的和
func CalculateSum(a, b int) int {
    return a + b
}
```

### TypeScript / Vue 代码

- 使用 ESLint + Prettier(项目已配置)
- Vue 使用 Composition API + `<script setup>`
- 组件名使用 PascalCase
- 文件名使用 kebab-case 或 PascalCase(组件文件)
- Props / Events 添加类型定义

```vue
<script setup lang="ts">
import { ref } from 'vue'

interface Props {
  title: string
  count?: number
}

const props = withDefaults(defineProps<Props>(), {
  count: 0,
})
</script>
```

### 命名约定

| 类型 | 风格 | 示例 |
|---|---|---|
| Go 包名 | 小写单词 | `engine`, `service` |
| Go 文件名 | 小写下划线 | `scenario_service.go` |
| Go 结构体 | PascalCase | `ScenarioRepository` |
| Go 方法 | PascalCase | `GetByID` |
| Go 私有 | camelCase | `parseRequest` |
| Vue 组件 | PascalCase | `KpiCard.vue` |
| Vue 路由 | kebab-case | `scenario-editor` |
| TS 函数 | camelCase | `loadList` |
| 常量 | UPPER_SNAKE | `MAX_VUS` |

---

## 📜 提交规范

我们使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范。

### 格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type 类型

| Type | 说明 |
|---|---|
| `feat` | 新功能 |
| `fix` | 修复 bug |
| `docs` | 仅文档变更 |
| `style` | 不影响代码含义的格式变更(空格、格式化等) |
| `refactor` | 重构代码(既不修复 bug 也不添加功能) |
| `perf` | 性能优化 |
| `test` | 添加或修正测试 |
| `chore` | 构建过程或辅助工具变更 |
| `ci` | CI 配置变更 |

### 示例

```bash
feat(scenario): 添加 HAR 文件导入功能

- 支持 Chrome DevTools HAR 1.2 格式
- 自动跳过浏览器特定 header
- 自动 JSON Path 提取器

Closes #123

fix(engine): 修复 VU goroutine 泄漏问题

VU goroutine 在暂停时没有正确退出,导致资源占用持续增长。

Fixes #456

docs(readme): 更新快速开始说明

补充 Linux 平台依赖安装命令。
```

---

## 🔄 Pull Request 流程

### 1. Fork 仓库

点击 GitHub 页面右上角的 "Fork" 按钮。

### 2. Clone 你的 fork

```bash
git clone https://github.com/YOUR_USERNAME/pulse.git
cd pulse
git remote add upstream https://github.com/pulse/pulse.git
```

### 3. 创建分支

```bash
git checkout -b feature/amazing-feature
# 或
git checkout -b fix/annoying-bug
```

### 4. 开发 & 测试

```bash
# 启动开发模式
wails dev

# 后端测试
go test ./...

# 前端测试
cd frontend && pnpm test

# Lint
gofmt -l .
cd frontend && pnpm lint
```

### 5. 提交

```bash
git add .
git commit -m "feat: 添加 amazing feature"
```

### 6. 同步上游(可选)

```bash
git fetch upstream
git rebase upstream/main
```

### 7. 推送到你的 fork

```bash
git push origin feature/amazing-feature
```

### 8. 创建 Pull Request

1. 访问 https://github.com/pulse/pulse
2. 点击 "New Pull Request"
3. 选择你的分支
4. 填写 PR 模板:
   - **标题**: 简洁描述
   - **描述**: 关联的 issue、改动说明、截图
5. 等待 CI 通过
6. 与 reviewer 讨论修改
7. 合并 🎉

### PR 检查清单

提交 PR 前,请确认:

- [ ] 代码遵循项目的代码风格
- [ ] 添加了必要的测试(并全部通过)
- [ ] 更新了相关文档
- [ ] 提交信息遵循 Conventional Commits
- [ ] 没有引入新的 lint 警告
- [ ] 没有遗留的调试代码(console.log 等)
- [ ] 没有未解决的合并冲突

---

## 🧪 测试要求

### 后端测试

```bash
# 所有测试
go test ./... -v

# 特定包
go test ./internal/engine/... -v

# 性能基准
go test ./internal/engine/ -bench=. -benchmem
```

新功能必须包含单元测试。Bug 修复必须包含回归测试。

### 前端测试

```bash
cd frontend

# Lint
pnpm lint

# Type check
pnpm type-check

# Build
pnpm build
```

### 测试覆盖率目标

| 模块 | 目标覆盖率 |
|---|---|
| `internal/engine/` | ≥ 70% |
| `internal/service/` | ≥ 60% |
| `internal/repository/` | ≥ 70% |
| 前端组件 | ≥ 50% (核心组件) |

---

## 📦 发布流程

(仅维护者)

```bash
# 1. 更新版本号(在多个文件中)
# - frontend/package.json: version
# - internal/config/config.go: App.Version
# - README.md / CHANGELOG.md

# 2. 更新 CHANGELOG.md

# 3. 提交并打 tag
git add .
git commit -m "chore(release): v0.2.0"
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin main --tags

# 4. GitHub Actions 自动构建并发布
# 5. 在 GitHub Release 页面编辑发布说明
```

---

## 🏷 Issue 模板

### Bug 报告

```markdown
## Bug 描述
简洁清晰地描述 bug。

## 复现步骤
1. 打开场景编辑器
2. 配置场景
3. 点击运行
4. 看到错误

## 期望行为
应该发生什么。

## 实际行为
实际发生了什么。

## 截图
(如有)

## 环境
- Pulse 版本:v0.1.0
- 操作系统:macOS 14.1
- Wails 版本:v2.10.1

## 其他
任何其他相关信息。
```

### 功能请求

```markdown
## 功能描述
清晰描述希望添加的功能。

## 解决的问题
为什么需要这个功能?解决了什么问题?

## 替代方案
考虑过的其他方案。

## 优先级
低 / 中 / 高
```

---

## ❓ 问题?

如果有任何问题:

- 📖 查看 [README.md](README.md)
- 💬 在 [Discussions](https://github.com/pulse/pulse/discussions) 中提问
- 🐛 在 [Issues](https://github.com/pulse/pulse/issues) 中搜索
- 📧 邮件:[team@pulse.dev](mailto:team@pulse.dev)

---

## 🙏 致谢

感谢所有为 Pulse 做出贡献的人!

<a href="https://github.com/pulse/pulse/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=pulse/pulse" alt="Contributors" />
</a>

---

**再次感谢你的贡献!** 💓