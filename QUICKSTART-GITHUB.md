# 🚀 Pulse GitHub 推送快速指南

> 5 分钟将 Pulse 推送到 GitHub!

---

## 第一步:创建 GitHub 仓库(2 分钟)

1. 打开 https://github.com/new
2. 填写:
   - **Repository name**: `pulse`
   - **Description**: `💓 Modern open-source load testing tool — desktop app for developers`
   - **Public** 公开仓库
   - **不要**勾选任何初始化选项
3. 点击 **Create repository**
4. 复制仓库 URL(SSh:`git@github.com:YOUR_NAME/pulse.git`)

---

## 第二步:配置 SSH Key(如果还没有)

```bash
# 检查现有 key
ls -la ~/.ssh/id_*.pub 2>/dev/null

# 没有就生成
ssh-keygen -t ed25519 -C "your_email@example.com"

# 复制公钥
cat ~/.ssh/id_ed25519.pub
# 然后到 GitHub → Settings → SSH and GPG keys → New SSH key 粘贴
```

测试连接:
```bash
ssh -T git@github.com
# 输出: Hi YOUR_NAME! You've successfully authenticated...
```

---

## 第三步:运行推送脚本(1 分钟)

```bash
cd ~/Desktop/work/学习/pulse-dev/pulse

# 1. 编辑脚本,设置你的用户名和邮箱
nano scripts/setup-git.sh
# 修改以下变量:
#   GITHUB_USER="your-github-username"
#   GITHUB_EMAIL="your-email@example.com"

# 2. 运行
bash scripts/setup-git.sh
```

脚本会问你 `确认推送? (y/N)`,输入 `y`。

---

## 第四步:配置 GitHub 仓库(2 分钟)

### 4.1 About 描述

访问 https://github.com/YOUR_NAME/pulse

点击右上角 ⚙️ 设置图标:

- **Description**: `💓 Modern open-source load testing tool — desktop app for developers`
- **Website**: `https://pulse.dev`(可选)
- **Topics**(点击 + 添加):
  - `load-testing`
  - `performance-testing`
  - `wails`
  - `golang`
  - `vue3`
  - `desktop-app`
  - `developer-tools`
  - `apache-2-0`

### 4.2 启用 Features

Settings → General → Features:
- ☑️ Issues
- ☑️ Discussions
- ☐ Wiki(可选)
- ☑️ Sponsorships(可选)

Settings → General → Pull Requests:
- ☑️ Allow squash merging
- ☑️ Allow rebase merging
- ☐ Allow merge commits(推荐关闭)
- ☑️ Automatically delete head branches

### 4.3 保护 main 分支

Settings → Branches → Add rule:

- Branch name pattern: `main`
- ☑️ Require a pull request before merging
  - ☑️ Require approvals: 1
- ☑️ Require status checks to pass before merging
- ☑️ Require linear history
- (可选) ☑️ Include administrators

---

## 第五步:触发首次 Release(可选)

```bash
cd ~/Desktop/work/学习/pulse-dev/pulse

# 编辑版本号(已默认是 0.1.0,如果有改动就编辑)

# 提交并打 tag
make release-tag
# 等价于:
#   git tag -a v0.1.0 -m "Release v0.1.0"
#   git push origin v0.1.0
```

GitHub Actions 会自动构建三平台产物并发布。

---

## 第六步:推广(可选)

### 发布社区帖

中文:
- [V2EX 创造](https://www.v2ex.com/?tab=creative)
- [掘金沸点](https://juejin.im/pins)
- [TesterHome](https://testerhome.com/)

英文:
- [Hacker News Show HN](https://news.ycombinator.com/show)
- [Reddit r/golang](https://www.reddit.com/r/golang/)

参考模板:
```
🚀 Just released Pulse v0.1.0 — a modern open-source load testing tool!

💓 Modern UI · ⚡ Self-developed Go engine · 📥 HAR import · 🌐 Cross-platform · 📦 Apache 2.0

⭐ GitHub: https://github.com/YOUR_NAME/pulse

#opensource #golang #loadtesting #vue3 #wails
```

### 准备 Social Preview 图

1280x640 PNG,推荐内容:
- 左:Pulse Logo
- 中:标语 + "压力测试,理应如此。"
- 右:实时大屏截图

设置路径:Settings → Social preview → Upload an image

---

## 🎉 完成!

现在 Pulse 已经在 GitHub 上了,你可以:

- 🔗 分享链接:`https://github.com/YOUR_NAME/pulse`
- ⭐ 等用户 Star
- 🐛 收集 Issues 和 PR
- 📦 等 GitHub Actions 构建完成,Release 就有产物了

---

## 📚 完整文档

- [RELEASE.md](docs/RELEASE.md) - 完整发布流程
- [CONTRIBUTING.md](CONTRIBUTING.md) - 贡献指南
- [README.md](README.md) - 项目说明

---

## ❓ 遇到问题?

| 问题 | 解决 |
|---|---|
| `Permission denied` | 检查 SSH Key 配置,或用 HTTPS + Token |
| GitHub Actions 失败 | 查看 Actions 页面的日志 |
| Wails 构建报错 | `go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.1` |
| macOS Gatekeeper 警告 | 右键 → 打开(未签名) |

更多问题见 [RELEASE.md 常见问题](docs/RELEASE.md#七常见问题)。