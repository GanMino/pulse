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

command -v go &> /dev/null || error "Go 未安装。访问 https://go.dev/dl/ 安装"
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
success "Go $GO_VERSION ✓"

command -v node &> /dev/null || error "Node.js 未安装"
NODE_VERSION=$(node --version)
success "Node.js $NODE_VERSION ✓"

command -v pnpm &> /dev/null || {
    warn "pnpm 未安装,尝试自动安装..."
    npm install -g pnpm@9 || error "pnpm 安装失败"
}
PNPM_VERSION=$(pnpm --version)
success "pnpm $PNPM_VERSION ✓"

command -v wails &> /dev/null || {
    warn "Wails CLI 未安装,尝试自动安装..."
    go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.1 || error "Wails 安装失败"
    success "Wails 已安装"
}
WAILS_VERSION=$(wails version 2>/dev/null | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | head -1)
success "Wails $WAILS_VERSION ✓"

# ========================================
# Step 2: 依赖安装
# ========================================
info ""
info "Step 2/5:安装依赖..."

go mod download
success "Go 依赖 ✓"

cd frontend
if [ ! -d "node_modules" ]; then
    pnpm install
fi
success "前端依赖 ✓"
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
    # 数据库尚未创建,提示用户先启动一次
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
wails dev