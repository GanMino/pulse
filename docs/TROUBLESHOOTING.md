# 🔧 Pulse 故障排查指南

> 启动 Pulse 过程中遇到的问题,这里都有答案。

---

## 📋 常见问题速查

| 问题 | 解决方案 |
|---|---|
| ❌ Go 模块下载失败 | 设置 `GOPROXY=https://goproxy.cn,direct` |
| ❌ pnpm 下载慢/超时 | 设置 `pnpm config set registry https://registry.npmmirror.com` |
| ❌ 前端依赖下载卡住 | 跳过 release 预下载,或重试 |
| ❌ wails dev dyld 错误 | 重新安装 Wails CLI |
| ❌ sqlite native binding 错误 | 升级 glebarez/sqlite 版本 |
| ❌ 数据库初始化失败 | 删除 `~/.pulse/data/pulse.db` 重试 |

---

## 🟢 场景 1:Go 模块下载超时

**症状**:
```
go: github.com/HdrHistogram/hdrhistogram-go@v1.3.0: 
  Get "https://proxy.golang.org/...": dial tcp: i/o timeout
```

**原因**:国际代理 `proxy.golang.org` 在大陆访问慢

**解决**:

```bash
# 一次性(当前终端)
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=off

# 永久(写到配置)
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=off
```

**进一步排查**:

```bash
# 测试哪个代理可用
curl -I https://goproxy.cn/      # 国内
curl -I https://proxy.golang.org/  # 国际
```

---

## 🟢 场景 2:pnpm 下载卡住

**症状**:
```
WARN GET https://registry.npmmirror.com/echarts error (ERR_SOCKET_TIMEOUT)
```

**原因**:当前 registry 镜像不稳定

**解决**(已自动写入 `scripts/demo.sh`):

```bash
# 切到国内镜像
pnpm config set registry https://registry.npmmirror.com

# 增加超时(默认 30s 太短)
pnpm config set network-timeout 300000

# 多次重试
pnpm config set fetch-retries 3

# 验证
pnpm config get registry
```

**手动快速安装**(如 demo.sh 卡住):

```bash
cd ~/Desktop/work/学习/pulse-dev/pulse/frontend
pnpm install --reporter=append-only 2>&1 | tee /tmp/pnpm.log
```

---

## 🟢 场景 3:快速跳过下载(我的 sandbox 已装好)

如果实在下载不下来,可以直接用 release 里的预装包:

```bash
# 1. 下载(替换 GanMino 为你的用户名)
curl -L -o /tmp/node_modules.tar.gz \
  "https://github.com/GanMino/pulse/releases/download/v0.1.0/pulse-node_modules.tar.gz"

# 2. 解压
cd ~/Desktop/work/学习/pulse-dev/pulse/frontend
rm -rf node_modules  # 如果有部分下载的
tar -xzf /tmp/node_modules.tar.gz

# 3. 重新跑 demo.sh(会跳过 node_modules 安装)
cd ..
bash scripts/demo.sh
```

> ⚠️ **注意**:node_modules 是 sandbox 内的 macOS arm64 版本,可能含本地缓存。解压后 macOS x86 用户的 esbuild 等 native 模块需要重新编译。建议仅在 arm64 Mac 试用。

---

## 🟢 场景 4:Wails dyld 错误

**症状**:
```
dyld[123]: missing LC_UUID load command
Abort trap: 6
```

**原因**:Go 编译时未生成 LC_UUID,macOS dyld 拒绝加载

**解决**:

```bash
# 1. 清理 Go 缓存
go clean -cache

# 2. 重新安装 Wails
go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.1

# 3. 验证
wails doctor
```

如果还失败,可能是 Go 版本与 Wails 不兼容,考虑升级到 Go 1.23+:

```bash
brew install go@1.23
export PATH="/usr/local/opt/go@1.23/bin:$PATH"
```

---

## 🟢 场景 5:`wails dev` 启动失败(应用窗口打不开)

**症状**:
```
panic: runtime error: invalid memory address or nil pointer dereference
```

**原因**:通常是 native 模块(WebView)加载问题

**排查**:

```bash
# 1. 查看 macOS 系统日志
log show --predicate 'process == "wails"' --last 5m

# 2. 检查 WebView2(Windows)
# Windows 11 自带,Windows 10 需要安装:
# https://developer.microsoft.com/en-us/microsoft-edge/webview2/

# 3. macOS:确认 Xcode CLT
xcode-select --install
```

---

## 🟢 场景 6:SQLite 错误

**症状**:
```
sql: database is locked
```

**解决**:

```bash
# 1. 关闭所有 Pulse 进程
pkill -f pulse

# 2. 删除数据库(会丢失数据!)
rm ~/.pulse/data/pulse.db

# 3. 重启(自动重新创建)
wails dev
```

**预防**:不要同时运行多个 Pulse 实例

---

## 🟢 场景 7:端口冲突

**症状**:
```
bind: address already in use
```

**说明**:Pulse 桌面应用通常不占端口,可能是浏览器工具冲突。

**解决**:
- 关闭 VSCode / IntelliJ 等占用 5173 端口的工具
- 或者在 `vite.config.ts` 中修改 `server.port`

---

## 🟢 场景 8:如何查看详细日志

```bash
# Go 后端日志
~/.pulse/logs/pulse.log  # 如果配置了文件日志
# 或者 wails dev 的终端输出

# 前端日志
# wails dev 终端的 [Vite] 输出

# macOS 系统日志
log show --predicate 'process == "Pulse"' --last 5m

# 启用 Pulse 调试日志
export PULSE_LOG_LEVEL=debug
wails dev
```

---

## 🟢 场景 9:重新从零开始

如果一切都乱了,完全重置:

```bash
# 1. 删除构建产物和缓存
cd ~/Desktop/work/学习/pulse-dev/pulse
rm -rf frontend/node_modules frontend/dist build/bin
go clean -cache

# 2. 删除应用数据(慎用!会丢失所有场景和报告)
rm -rf ~/.pulse

# 3. 重新拉代码
git pull origin main

# 4. 重新跑
bash scripts/demo.sh
```

---

## 🟢 场景 10:GitHub Actions 失败

如果 CI / Release 失败:

1. 打开 https://github.com/GanMino/pulse/actions
2. 查看失败 job 的日志
3. 常见原因:
   - macOS: 需要 Xcode CLT(`xcode-select --install`)
   - Linux: 缺少系统库(`libgtk-3-dev libwebkit2gtk-4.0-dev`)
   - Windows: 需要 WebView2

---

## 📊 性能调优

如果遇到性能问题:

```bash
# 1. 关闭 Spotlight 索引 Pulse 数据目录
mdutil -i off ~/.pulse

# 2. 禁用 Wails dev 工具
export PULSE_DEVTOOLS=false

# 3. 增加 macOS 文件描述符上限
ulimit -n 65536
```

---

## 🆘 完全卡死怎么办?

1. **查看 Issues**:https://github.com/GanMino/pulse/issues
2. **搜索错误信息**:在 Google 搜 "[错误信息] wails"
3. **重新克隆**:可能是当前仓库问题,删了重 clone
4. **查看类似项目**:k6 / locust 的 issue 区可能有相同问题的解决方案

---

## ✅ 验证安装成功

启动后,应用窗口打开,你应该看到:

```
┌────────────────────────────────────────┐
│  💓 Pulse  v0.1.0                     │
│  ──────────────────────────────────────│
│  Dashboard  Scenarios  Reports  Settings│
│  ──────────────────────────────────────│
│  欢迎使用 Pulse 👋                       │
│  压力测试,理应如此。                    │
│  ──────────────────────────────────────│
│  ● Ready · Agent: local-1 · DB: 0 MB    │
└────────────────────────────────────────┘
```

**🎉 看到这个窗口就成功了!** 进入 Scenarios 看到 5 个示例场景,说明 demo 数据也填充好了。

---

## 📞 仍然无法解决?

提供以下信息给我:
1. 操作系统和版本(`uname -a`)
2. Go / Node / pnpm 版本
3. `wails doctor` 输出
4. 完整的错误信息
5. 截图(如果有 GUI 错误)

我帮你具体分析!💪