#!/bin/bash
# ==========================================
# Pulse 项目 Git 初始化 + 推送脚本
# ==========================================
# 用法:
#   1. 在 GitHub 创建空仓库(pulse/pulse),不要初始化任何文件
#   2. 编辑下面的 GITHUB_REPO 变量
#   3. 运行: bash scripts/setup-git.sh
# ==========================================

set -e  # 出错时退出

# ====== 配置 ======
GITHUB_USER="pulse"           # 你的 GitHub 用户名
GITHUB_REPO="pulse"            # 仓库名
GITHUB_EMAIL="team@pulse.dev"  # 提交使用的邮箱
GIT_USER_NAME="Pulse Contributors"  # 提交使用的名字

REPO_URL="git@github.com:${GITHUB_USER}/${GITHUB_REPO}.git"
# 如果用 HTTPS,改为:
# REPO_URL="https://github.com/${GITHUB_USER}/${GITHUB_REPO}.git"

# ====== 颜色输出 ======
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[!]${NC} $1"
}

error() {
    echo -e "${RED}[✗]${NC} $1"
    exit 1
}

# ====== 前置检查 ======
info "检查前置条件..."

# 检查 git
if ! command -v git &> /dev/null; then
    error "git 未安装"
fi

# 检查是否在 pulse 项目根目录
if [ ! -f "go.mod" ] || [ ! -f "wails.json" ]; then
    error "请在 pulse 项目根目录运行此脚本"
fi

# 检查是否已经初始化
if [ -d ".git" ]; then
    warn ".git 目录已存在"
    read -p "是否继续? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 0
    fi
fi

# ====== 配置 Git 用户 ======
info "配置 Git 用户..."
git config user.name "$GIT_USER_NAME"
git config user.email "$GITHUB_EMAIL"
success "Git 用户配置完成"

# ====== 初始化仓库 ======
info "初始化 Git 仓库..."
git init
success "Git 仓库初始化完成"

# ====== 创建 main 分支 ======
git checkout -b main 2>/dev/null || git checkout main

# ====== 添加 .gitignore 中已包含的文件 ======
info "添加文件到暂存区..."

# 校验必要文件存在
required_files=(
    "README.md"
    "LICENSE"
    "go.mod"
    "main.go"
    "app.go"
    "wails.json"
    "Makefile"
    "CONTRIBUTING.md"
    "CHANGELOG.md"
    ".gitignore"
    "frontend/package.json"
)

for f in "${required_files[@]}"; do
    if [ ! -f "$f" ]; then
        warn "文件不存在: $f"
    fi
done

git add .
success "文件已添加到暂存区"

# ====== 显示状态 ======
echo ""
info "当前 Git 状态:"
git status --short | head -30
echo "..."

# ====== 提交 ======
info "创建初始提交..."

git commit -m "feat: Pulse v0.1.0 MVP 🎉

- 自研 Go 压测引擎(支持 20,000 VUs)
- 现代 Vue 3 + Naive UI 桌面应用
- 场景编辑器(HAR 导入 / 多请求 / 提取器 / 断言)
- 实时监控大屏(KPI / 折线图 / 状态码)
- 报告生成与导出
- 跨平台支持(macOS / Windows / Linux)
- Apache 2.0 开源协议"

success "初始提交完成"

# ====== 添加远程仓库 ======
info "添加远程仓库: $REPO_URL"
git remote add origin "$REPO_URL" 2>/dev/null || git remote set-url origin "$REPO_URL"
success "远程仓库已添加"

# ====== 推送 ======
echo ""
warn "即将推送到 GitHub..."
echo ""
info "如果使用 SSH,请确保已配置 SSH Key"
info "如果使用 HTTPS,推送时会要求输入用户名密码(或 Token)"
echo ""
read -p "确认推送? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    warn "已取消"
    echo ""
    info "如需稍后推送,运行:"
    echo "  git push -u origin main"
    exit 0
fi

info "推送到 main 分支..."
git push -u origin main
success "推送完成!"

# ====== 设置默认分支保护(可选) ======
echo ""
info "🎉 仓库已成功推送到 GitHub!"
echo ""
info "下一步建议:"
echo "  1. 在 GitHub 仓库页面设置 About 描述和 Topics"
echo "  2. 启用 GitHub Actions / Issues / Discussions"
echo "  3. 配置分支保护规则(保护 main)"
echo "  4. 创建第一个 Release(v0.1.0)"
echo "  5. 邀请协作者"
echo ""
info "查看仓库: https://github.com/${GITHUB_USER}/${GITHUB_REPO}"