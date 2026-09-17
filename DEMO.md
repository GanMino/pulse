# 🎬 Pulse 演示指南

> 没有云端"在线演示",但你可以**5 分钟内**在自己的机器上看到 Pulse 的完整效果!

---

## 方式 1:一键演示(最快,5 分钟)

### 前置条件

- macOS / Windows / Linux
- Go 1.22+,Node.js 20+,pnpm 9+
- Wails CLI v2.10+
- ~ 500MB 磁盘空间

### 一行命令

```bash
cd ~/Desktop/work/学习/pulse-dev/pulse
bash scripts/demo.sh
```

脚本会自动:
1. ✅ 检查 Go / Node / pnpm / Wails 环境
2. ✅ 安装依赖
3. ✅ 初始化数据库
4. ✅ 填充演示数据(5 场景 + 7 历史测试)
5. ✅ 启动应用窗口

### 你会看到

应用窗口打开后,你会看到丰富的数据:

#### 📊 Dashboard
- 欢迎语 + 你的数据
- 5 个场景 / 7 次测试 / 100万+ 总请求数
- 最近测试列表(成功/失败状态清晰)
- 场景列表

#### 📋 Scenarios
- ⭐ 登录流程压测(v3)
- 商品列表查询(v1)
- 大促容量测试(v5)
- 支付链路(草稿)
- 从 Chrome HAR 录制导入

#### 📈 Reports
- 7 条历史测试(7 天前 / 3 天前 / 1 天前 / 12 小时前)
- 不同状态:✅ 通过 / ❌ 失败
- 总请求数从 21,600 到 312,450
- P95 从 78ms 到 1,245ms(覆盖各种场景)

---

## 方式 2:从零开始(更深入,15 分钟)

### Step 1:克隆 + 安装

```bash
git clone https://github.com/GanMino/pulse.git
cd pulse
make install
```

### Step 2:启动应用

```bash
wails dev
```

### Step 3:首次启动 - 看到空状态

首次启动会显示:
- 空的 Dashboard(0 场景 / 0 测试)
- Scenarios 列表空,显示"创建第一个场景"按钮
- Reports 空

### Step 4:创建你的第一个场景

1. 点 **➕ New Scenario**
2. 配置请求(POST /api/login + JSON body)
3. 配置压测参数(VUs: 100 / Duration: 1m)
4. 点 **💾 Save** (⌘S)
5. 点 **▶ Save & Run** (⌘⇧R)

### Step 5:运行真实压测

实时大屏打开,会看到:
- KPI 卡片实时刷新
- 折线图动态滚动
- 状态码分布

### Step 6:填充演示数据(可选)

如果想看更丰富的数据:

```bash
bash scripts/seed-demo-data.sh
# 关闭 Pulse 后重启,数据就有了
```

---

## 方式 3:只看设计稿(30 秒)

打开 [`UI-UX详细设计.md`](UI-UX详细设计.md),里面有:
- 9 个页面的 ASCII mockup
- 7 种状态的实时大屏设计
- 完整设计系统(配色 / 字体 / 间距)

---

## 🎬 制作 Demo 视频(社区宣传用)

### 5 分钟录屏脚本

```
[0:00 - 0:15] 开场
"💓 这是 Pulse,一款让你 5 分钟上手的开源压测工具"

[0:15 - 0:30] 启动应用
打开 Pulse,展示干净的 Dashboard 界面
"零配置启动,默认用户自动创建"

[0:30 - 1:00] 创建场景
点击"新建场景" → 输入 URL → 配置 VUs 100 → Duration 1m
"三步配置,无需写代码"

[1:00 - 1:30] HAR 导入演示
演示 Chrome DevTools 录制 → 导出 HAR → 拖入 Pulse
"录制浏览器操作,一键生成压测场景"

[1:30 - 2:30] 运行 + 实时大屏
点击"运行" → 自动跳转到实时大屏
展示 KPI 卡片、折线图、状态码分布
"2 秒刷新一次的实时监控"

[2:30 - 3:00] 报告 + 历史
测试完成后展示自动生成的报告
进入 Reports → 展示历史对比
"所有数据本地保存,可导出 HTML"

[3:00 - 3:30] 命令面板 + 快捷键
按 ⌘K 演示命令面板
展示暗色主题、快捷键
"键盘友好,效率工具"

[3:30 - 4:00] 性能数据
切到终端,展示:
go test -bench=. -benchtime=10s
"Go 自研引擎,单机能跑 50k+ RPS"

[4:00 - 5:00] 结尾
- GitHub 链接
- 截图最后的关键页面
- ⭐ Star 邀请
```

### 推荐工具

- **macOS**: QuickTime Player → 新建屏幕录制
- **Windows**: OBS Studio
- **跨平台**: OBS Studio(免费)

### 视频上传

- **B 站**: https://member.bilibili.com/vupload
- **YouTube**: https://studio.youtube.com
- **Twitter / X**: 直接发短视频

---

## 📸 截图指南(用于 README / 博客)

### 推荐截图清单

1. **Dashboard**(暗色主题,展示统计)
2. **Scenarios 列表**(展示卡片视图)
3. **Scenario Editor**(三栏布局)
4. **LiveMonitor**(实时大屏,KPI 卡片 + 折线图)
5. **Report Detail**(报告详情)
6. **RunService 输出**(终端展示 go test -bench)

### 截图工具

- **macOS**: `Cmd+Shift+4`(区域)/ `Cmd+Shift+5`(全屏)
- **Windows**: `Win+Shift+S` (Snipping Tool)
- **Linux**: `gnome-screenshot` / `flameshot`

### 截图优化

```bash
# macOS:关闭阴影,只截窗口
defaults write com.apple.screencapture disable-shadow -bool true

# 截取特定窗口(优雅)
Cmd+Shift+4 → 空格 → 点击窗口
```

---

## 🐳 未来:在线 Demo(规划中)

如果你想要"在线 demo"看效果,有几种方案(MVP 不做):

### 方案 A:Web 版本(推荐)

把 Pulse 改成 Web 应用,部署到云端:
- 用户无需下载
- 提供公共账号试用
- 但需要后端服务器(成本)

### 方案 B:Docker 镜像

```bash
docker run -it --rm \
  -e DISPLAY=$DISPLAY \
  -v /tmp/.X11-unix:/tmp/.X11-unix \
  ganmino/pulse:latest
```

需要在支持 X11 的 Linux 主机上运行。

### 方案 C:录屏 GIF

最简单:
- 用 `peek` / `LICEcap` 录 30 秒 GIF
- 放在 README 里
- 用户点开就能看

### 方案 D:静态截图

最低成本:
- 上面提供的截图指南
- 放在 README / 博客
- 写文字介绍

---

## 🎯 推荐路径

如果你的目标是**自己体验**:用**方式 1**(一键演示)
如果你的目标是**理解代码**:用**方式 2**(从零开始)
如果你的目标是**快速看设计**:用**方式 3**(看 mockup)
如果你的目标是**社区宣传**:制作**5 分钟录屏 + 截图**

---

**🚀 5 分钟启动看效果,就是这么简单!**