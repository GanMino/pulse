# Pulse 发布指南

本文档介绍如何将 Pulse 项目发布到 GitHub,包括仓库初始化、打包、Release 等完整流程。

---

## 🎯 流程概览

```
本地代码 → git init → GitHub 创建空仓库 → git push
                                         ↓
                                    GitHub Actions 自动构建
                                         ↓
                                    生成 3 个平台产物
                                         ↓
                                    自动创建 GitHub Release
                                         ↓
                                    用户下载预编译版本
```

---

## 一、第一次发布(从 0 到 1)

### Step 1:在 GitHub 创建空仓库

1. 访问 https://github.com/new
2. 填写:
   - **Repository name**: `pulse`
   - **Description**: `💓 Modern open-source load testing tool — desktop app for developers`
   - **Visibility**: Public(开源)
   - **不要**勾选 "Add a README file" / "Add .gitignore" / "Choose a license"
3. 点击 "Create repository"

### Step 2:配置 SSH Key(如果还没有)

```bash
# 检查是否已有
ls -la ~/.ssh/id_*.pub

# 没有就生成
ssh-keygen -t ed25519 -C "your_email@example.com"
cat ~/.ssh/id_ed25519.pub
# 复制输出,在 GitHub → Settings → SSH and GPG keys → New SSH key 添加
```

### Step 3:运行推送脚本

```bash
cd ~/Desktop/work/学习/pulse-dev/pulse

# 编辑脚本,设置你的 GitHub 用户名和邮箱
vim scripts/setup-git.sh
# 修改:
#   GITHUB_USER="your-username"
#   GITHUB_EMAIL="your-email@example.com"
#   GITHUB_REPO="pulse"
#   GIT_USER_NAME="Your Name"

# 运行
bash scripts/setup-git.sh
```

脚本会自动:
- 初始化 git 仓库
- 配置用户
- 添加所有文件
- 创建初始提交
- 添加远程仓库
- 推送到 GitHub

### Step 4:配置 GitHub 仓库

访问 https://github.com/YOUR_USERNAME/pulse/settings

#### 4.1 About(描述)
- **Description**: `💓 Modern open-source load testing tool — desktop app for developers`
- **Website**: `https://pulse.dev`(可选)
- **Topics**:`load-testing` `performance-testing` `desktop-app` `wails` `go` `vue3` `developer-tools`

#### 4.2 Features(启用功能)

Settings → General → Features:
- ✅ Issues
- ✅ Sponsorships
- ✅ Preserve this repository
- ✅ Discussions
- ✅ Wiki(可选)

Settings → General → Pull Requests:
- ✅ Allow squash merging
- ✅ Allow rebase merging
- ❌ Allow merge commit(推荐关闭,保持历史清晰)
- ✅ Automatically delete head branches

#### 4.3 Branch Protection

Settings → Branches → Add rule:
- **Branch name pattern**: `main`
- ✅ Require pull request reviews before merging(至少 1 个)
- ✅ Require status checks to pass before merging
- ✅ Require branches to be up to date before merging
- ✅ Require linear history

#### 4.4 Secrets(用于未来 CI/CD)

Settings → Secrets and variables → Actions:
- `CODECOV_TOKEN` - Codecov token(可选,代码覆盖率)

---

## 二、发布新版本

### 流程图

```
更新版本号 → 更新 CHANGELOG → git commit → 打 tag → git push --tags
                                                          ↓
                                                  GitHub Actions 自动触发
                                                          ↓
                                                  构建 3 平台产物
                                                          ↓
                                                  创建 GitHub Release
```

### Step 1:更新版本号

需要修改的地方:

```bash
# 1. frontend/package.json
{
  "name": "pulse-frontend",
  "version": "0.2.0",  # ← 改这里
  ...
}

# 2. internal/config/config.go
var (
    Version = "0.2.0"  // ← 改这里
)

# 3. CHANGELOG.md
## [Unreleased]
### Added
- 新功能

## [0.2.0] - 2025-03-01
### Added
- ...

# 4. wails.json
{
  ...
  "info": {
    "productVersion": "0.2.0",  // ← 改这里
    ...
  }
}
```

### Step 2:提交

```bash
git add .
git commit -m "chore(release): v0.2.0

- 添加 gRPC 协议支持
- 添加 Chrome 录制插件
- 修复 ... bug
- 优化 ... 性能"
```

### Step 3:打 Tag 并推送

**方法 1:使用 Makefile(推荐)**

```bash
# VERSION 会自动从 frontend/package.json 读取
make release-tag
```

**方法 2:手动**

```bash
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
```

### Step 4:等待 GitHub Actions 完成

访问 https://github.com/YOUR_USERNAME/pulse/actions

- 等待 release workflow 完成(约 10-20 分钟)
- 三个 job 会并行构建(macOS / Windows / Linux)

### Step 5:完善 Release 说明

访问 https://github.com/YOUR_USERNAME/pulse/releases/tag/v0.2.0

- 检查产物是否齐全:
  - `pulse-macos-universal`(DMG + APP)
  - `pulse-windows-amd64`(MSI + EXE)
  - `pulse-linux-amd64`(AppImage)
  - `SHA256SUMS.txt`(校验和)
- 编辑 Release Notes(可选,补充重要信息)
- 标记为 Latest release

---

## 三、本地构建产物验证

在推送 tag 前,建议本地先验证构建:

### macOS

```bash
make build-mac
open build/bin/pulse.app
```

### Windows(在 Windows 上)

```cmd
make build-win
build\bin\pulse.exe
```

### Linux

```bash
make build-linux
chmod +x build/bin/pulse
./build/bin/pulse
```

### 验证清单

- [ ] 应用能正常启动
- [ ] 默认用户和项目创建成功
- [ ] 能创建并运行一个简单场景
- [ ] 实时大屏数据正常显示
- [ ] 报告能正确生成
- [ ] 系统托盘 / 通知正常
- [ ] 关闭后重新打开数据持久化

---

## 四、发布后的推广

### 1. 完善 GitHub 仓库页面

- 添加 Topics:`load-testing`, `performance-testing`, `wails`, `go`, `vue3`
- 添加 Description
- 添加 Website(可选)
- 启用 Discussions

### 2. 创建 Social Preview 图

1280x640 PNG,在 Settings → Social preview → Upload:

推荐设计:
- 左侧:Pulse Logo
- 中间:标语"压力测试,理应如此"
- 右侧:实时大屏截图

### 3. 发布到社区

#### 中文社区
- [V2EX](https://www.v2ex.com/?tab=creative) - 创意 / 程序员板块
- [掘金](https://juejin.im/) - 技术文章
- [TesterHome](https://testerhome.com/) - 测试社区
- [知乎](https://www.zhihu.com/) - 技术专栏
- [思否](https://segmentfault.com/) - SegmentFault

#### 英文社区
- [Hacker News](https://news.ycombinator.com/show) - Show HN
- [Reddit](https://www.reddit.com/r/golang/) - r/golang, r/programming
- [Product Hunt](https://www.producthunt.com/) - 提交产品
- [Dev.to](https://dev.to/) - 技术文章
- [IndieHackers](https://www.indiehackers.com/)

#### 发帖模板

```markdown
🚀 Just released Pulse v0.1.0 — a modern open-source load testing tool!

💓 Modern UI (Vue 3 + Naive UI)
⚡ Self-developed Go engine (10k+ VUs, 50k+ RPS)
📥 HAR import (Chrome/Firefox/Edge → test scenario in 1 click)
🌐 Cross-platform (macOS/Windows/Linux)
📦 Apache 2.0

⭐ GitHub: https://github.com/YOUR_USERNAME/pulse
📖 Docs: https://github.com/YOUR_USERNAME/pulse#readme

Would love your feedback! #opensource #golang #loadtesting #vue3
```

### 4. 技术博客

考虑写一篇技术博客介绍 Pulse:
- 项目动机与设计思路
- Go 自研引擎的实现细节
- Wails + Vue 3 桌面应用开发经验
- 性能优化经验

发布到:
- 掘金 / 思否 / 知乎(中文)
- Medium / Dev.to(英文)

### 5. 提交到 awesome 列表

- [awesome-go](https://github.com/avelino/awesome-go)
- [awesome-vue](https://github.com/vuejs/awesome-vue)
- [awesome-performance](https://github.com/awesome-foss/awesome-performance)
- [awesome-testing](https://github.com/awesome-lists/awesome-testing)

---

## 五、版本号规则

我们遵循 [Semantic Versioning](https://semver.org/):

```
MAJOR.MINOR.PATCH

MAJOR: 不兼容的 API 变更
MINOR: 向下兼容的新功能
PATCH: 向下兼容的 bug 修复
```

预发布版本:
- `0.1.0-alpha` - 内部测试
- `0.1.0-beta` - 公开测试
- `0.1.0-rc.1` - 发布候选
- `0.1.0` - 正式发布

---

## 六、紧急修复流程

发现严重 bug,需要快速发布补丁版本:

```bash
# 1. 从 main 创建 hotfix 分支
git checkout main
git checkout -b hotfix/0.1.1

# 2. 修复 bug
# ... 修改代码 ...

# 3. 更新版本号到 0.1.1
# 修改 frontend/package.json, internal/config/config.go, wails.json

# 4. 更新 CHANGELOG
# 在 [Unreleased] 下添加修复说明
# 同时在 [0.1.1] 下同步

# 5. 提交并合并
git add .
git commit -m "fix: 修复严重 bug XXX"
git push origin hotfix/0.1.1

# 6. 合并到 main,打 tag
git checkout main
git merge hotfix/0.1.1
git push origin main

git tag -a v0.1.1 -m "Hotfix v0.1.1"
git push origin v0.1.1
```

---

## 七、常见问题

### Q: 推送失败 "Permission denied"
A: 检查 SSH Key 是否正确配置,或使用 HTTPS + Personal Access Token。

### Q: GitHub Actions 失败
A: 查看 https://github.com/YOUR_USERNAME/pulse/actions 的详细日志。

### Q: 产物缺失某个平台
A: 检查 release.yml 中的 matrix 配置,确认所有平台都包含。

### Q: Wails 构建报错
A: 升级 Wails CLI 到最新版本:`go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.1`

### Q: macOS Gatekeeper 警告
A: MVP 阶段未做代码签名。用户需要:
1. 右键点击 pulse.app → "打开"
2. 在弹出窗口点击 "打开"
3. 之后就可以正常打开了

或使用 `xattr -dr com.apple.quarantine pulse.app`

### Q: Windows SmartScreen 警告
A: 同上,未做代码签名。用户需要:
1. 点击"更多信息"
2. 选择"仍要运行"

---

## 八、发布检查清单

发布前:

- [ ] 所有测试通过(`make test`)
- [ ] Lint 无错误(`make lint`)
- [ ] 版本号在多处一致
- [ ] CHANGELOG.md 已更新
- [ ] README.md 是最新的
- [ ] 本地构建成功并测试过

发布时:

- [ ] Tag 命名正确(`vX.Y.Z`)
- [ ] Tag 推送成功
- [ ] GitHub Actions 全部成功

发布后:

- [ ] Release 页面显示所有产物
- [ ] SHA256SUMS.txt 正确
- [ ] 社区发帖推广
- [ ] 更新文档站(如适用)

---

**🎉 祝发布顺利!**