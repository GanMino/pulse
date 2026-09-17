#!/bin/bash
# ==========================================
# Pulse 一键演示启动器
# ==========================================
# 完整流程:环境检查 → 依赖安装 → 数据初始化 →
#          演示数据填充 → 启动应用
# ==========================================

set -e

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

info() { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[✓]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
error() { echo -e "${RED}[✗]${NC} $1"; exit 1; }
header() { echo -e "${CYAN}$1${NC}"; }

# ==========================================
# 网络优化:配置 Go 模块代理
# 国内用户: goproxy.cn 可用
# 海外用户: proxy.golang.org 可用
# 自动检测: 先尝试 goproxy.cn,失败则降级
# ==========================================
setup_go_proxy() {
    info "配置 Go 模块代理..."

    # 测试 goproxy.cn 可用性
    if curl -sS --max-time 3 -o /dev/null -w "%{http_code}" https://goproxy.cn/github.com/!wailsapp/wails/v2/@v/list 2>/dev/null | grep -q "200"; then
        export GOPROXY=https://goproxy.cn,direct
        export GOSUMDB=off  # 关闭 sumdb(Google 服务从大陆访问可能超时)
        success "使用国内代理 (goproxy.cn, 关闭 sumdb)"
    else
        export GOPROXY=https://proxy.golang.org,direct
        export GOSUMDB=off  # 关闭 sumdb(部分环境访问 Google 服务慢)
        success "使用国际代理 (proxy.golang.org, 关闭 sumdb)"
    fi
}

# 立即配置
setup_go_proxy

# 路径
PULSE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PULSE_DIR"

header ""
header "╔════════════════════════════════════════╗"
header "║   💓 Pulse 一键演示启动器              ║"
header "║   5 分钟内看效果                       ║"
header "╚════════════════════════════════════════╝"
header ""

# ========================================
# Step 1: 环境检查
# ========================================
info "Step 1/5:检查环境..."

# 配置 Go PATH
if ! command -v go &> /dev/null; then
    if [ -d "/usr/local/go/bin" ]; then
        export PATH=$PATH:/usr/local/go/bin
        if [ -d "$HOME/go/bin" ]; then
            export PATH=$PATH:$HOME/go/bin
        fi
    fi
fi

command -v go &> /dev/null || error "Go 未安装。访问 https://go.dev/dl/ 安装"
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
success "Go $GO_VERSION ✓"

# 检查 Go 版本是否满足最低要求
GO_MAJOR=$(echo $GO_VERSION | cut -d. -f1)
GO_MINOR=$(echo $GO_VERSION | cut -d. -f2)
if [ "$GO_MAJOR" -lt 1 ] || ([ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 22 ]); then
    error "Go 版本过低,需要 1.22+ (当前 $GO_VERSION)"
fi

command -v node &> /dev/null || error "Node.js 未安装"
NODE_VERSION=$(node --version)
success "Node.js $NODE_VERSION ✓"

command -v pnpm &> /dev/null || {
    warn "pnpm 未安装,尝试自动安装..."
    npm install -g pnpm@9 || warn "pnpm 安装失败,可手动安装: npm i -g pnpm"
}
if command -v pnpm &> /dev/null; then
    PNPM_VERSION=$(pnpm --version)
    success "pnpm $PNPM_VERSION ✓"
fi

# 配置 Go env(持久)
go env -w GOPROXY="$GOPROXY" 2>/dev/null || true
go env -w GOSUMDB="$GOSUMDB" 2>/dev/null || true

# 检查 Wails
if ! command -v wails &> /dev/null; then
    if [ -f "$HOME/go/bin/wails" ]; then
        export PATH=$PATH:$HOME/go/bin
    fi
fi

if ! command -v wails &> /dev/null; then
    warn "Wails CLI 未安装,尝试自动安装..."
    # 关键:-linkmode=external 使用 clang 链接,生成 LC_UUID
    # 否则 macOS 会报 "missing LC_UUID load command" 错误
    CGO_ENABLED=1 go install -trimpath -ldflags="-s -w -linkmode=external" github.com/wailsapp/wails/v2/cmd/wails@v2.10.1 || error "Wails 安装失败"
    export PATH=$PATH:$(go env GOPATH)/bin
    success "Wails 已安装"
fi

# 修复已安装但缺 LC_UUID 的 wails(旧编译产物)
if command -v wails &> /dev/null; then
    if ! wails version &> /dev/null 2>&1; then
        warn "检测到 wails 缺少 LC_UUID,重新编译..."
        CGO_ENABLED=1 go install -trimpath -ldflags="-s -w -linkmode=external" github.com/wailsapp/wails/v2/cmd/wails@v2.10.1 || error "Wails 重新编译失败"
        export PATH=$PATH:$(go env GOPATH)/bin
        success "Wails 已修复"
    fi
fi
WAILS_VERSION=$(wails version 2>/dev/null | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | head -1)
success "Wails $WAILS_VERSION ✓"

# ========================================
# Step 2: 依赖安装
# ========================================
info ""
info "Step 2/5:安装依赖..."

# 先尝试 go mod tidy(更智能)
if go mod tidy 2>&1 | tee /tmp/pulse-mod-tidy.log; then
    success "Go 依赖 (go mod tidy) ✓"
else
    # 退化到 go mod download
    warn "go mod tidy 失败,尝试 go mod download..."
    if go mod download 2>&1 | tee /tmp/pulse-mod-download.log; then
        success "Go 依赖 (go mod download) ✓"
    else
        error "Go 依赖安装失败。请检查网络,或将 GOPROXY 设置为 https://goproxy.cn,direct"
    fi
fi

cd frontend
if [ ! -d "node_modules" ] || [ ! -f "node_modules/.package-lock.json" ]; then
    # 配置 pnpm registry(国内优先,加快下载)
    pnpm config set registry https://registry.npmmirror.com 2>/dev/null || true
    pnpm config set network-timeout 300000 2>/dev/null || true  # 5 分钟超时
    pnpm config set fetch-retries 3 2>/dev/null || true

    info "使用 registry: $(pnpm config get registry)"

    if pnpm install --reporter=default 2>&1 | tee /tmp/pulse-pnpm-install.log; then
        success "前端依赖 ✓"
    else
        error "前端依赖安装失败。日志:/tmp/pulse-pnpm-install.log\n试运行: cd frontend && pnpm install"
    fi
else
    success "前端依赖已安装(跳过)"
fi
cd ..

# ========================================
# Step 3: 数据库初始化
# ========================================
info ""
info "Step 3/5:初始化数据库..."

DB_DIR="${HOME}/.pulse/data"
mkdir -p "$DB_DIR"
DB_PATH="${DB_DIR}/pulse.db"

if [ ! -f "$DB_PATH" ]; then
    info "数据库未初始化,启动后将自动创建"
fi

# ========================================
# Step 4: 填充演示数据
# ========================================
info ""
info "Step 4/5:填充演示数据..."

if [ -f "$DB_PATH" ]; then
    bash scripts/seed-demo-data.sh
else
    warn "数据库尚未创建,先启动应用让它自动创建,然后再次运行本脚本填充演示数据"
fi

# ========================================
# Step 5: 启动应用
# ========================================
info ""
info "Step 5/5:启动 Pulse..."

header ""
header "═══════════════════════════════════════"
header "  🚀 正在启动 Pulse 开发模式..."
header "═══════════════════════════════════════"
header ""
header "  💡 提示:"
header "     • 应用窗口会自动打开"
header "     • 修改 Go / Vue 代码会自动重载"
header "     • 按 ⌘K 试试命令面板"
header "     • 进入 Scenarios 查看 5 个示例场景"
header "     • 进入 Dashboard 查看统计数据"
header "     • 进入 Reports 查看 7 条历史测试"
header ""
header "  ⏹  停止: Ctrl+C"
header ""

# 启动(会打开窗口)
# 关键:设置 CGO + external linker,确保 wails 内部编译的二进制
# (wailsbindings 和应用本身)都生成 LC_UUID
# Go 1.22 的 internal linker 在 macOS 上不生成 LC_UUID,导致 dyld 报错
export CGO_ENABLED=1
export GOFLAGS="-ldflags=-linkmode=external"
wails dev