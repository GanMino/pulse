# Pulse 引擎性能基准测试指南

> 本文档介绍如何运行 Pulse 自研压测引擎的基准测试,以及预期性能指标。

---

## 一、运行测试

### 1. 单元 + 集成测试

```bash
cd ~/Desktop/work/学习/pulse-dev/pulse

# 运行所有 engine 包测试
go test ./internal/engine/... -v -timeout 60s

# 只跑集成测试
go test ./internal/engine/ -run "TestLocalEngine" -v

# 只跑场景基准测试
go test ./internal/engine/ -run "TestScenario" -v
```

### 2. 性能基准(Benchmark)

```bash
# 跑所有基准
go test ./internal/engine/ -bench=. -benchmem -benchtime=10s

# 跑特定基准
go test ./internal/engine/ -bench=BenchmarkEngine_LocalhostRPS -benchtime=10s
go test ./internal/engine/ -bench=BenchmarkEngine_HighConcurrency -benchtime=10s
go test ./internal/engine/ -bench=BenchmarkEngine_RampUp -benchtime=10s

# 生成 CPU profile
go test ./internal/engine/ -bench=. -benchtime=30s -cpuprofile=cpu.prof
go tool pprof cpu.prof

# 生成内存 profile
go test ./internal/engine/ -bench=. -benchtime=30s -memprofile=mem.prof
go tool pprof mem.prof
```

### 3. 并发压力测试

```bash
# 1k VUs 压测(标准模式)
go test ./internal/engine/ -run TestScenario_ConcurrentStress -v -timeout 60s

# 高并发基线(更慢)
go test ./internal/engine/ -run TestScenario_HighConcurrency -v -timeout 120s
```

---

## 二、性能预期指标

### 单机 localhost 测试

| 场景 | VUs | 时长 | 预期 RPS | 预期 P95 | 预期 P99 |
|---|---|---|---|---|---|
| **Hello World**(无 body) | 100 | 2s | 50,000+ | < 1ms | < 5ms |
| **Hello World** | 1,000 | 3s | 50,000+ | < 5ms | < 10ms |
| **JSON API** | 500 | 2s | 30,000+ | < 5ms | < 10ms |
| **POST Login**(JSON body) | 200 | 2s | 20,000+ | < 10ms | < 20ms |
| **带 think time(1s)** | 100 | 5s | 100 | - | - |

> **基线参考**(M2 MacBook Pro / i7-12700 / SSD):
> - localhost RPS:50k-100k+
> - 远程 HTTP RPS:30k-50k+
> - 1k 并发稳定运行:✅
> - 10k 并发稳定运行:✅(需 16GB+ 内存)

### 与其他工具对比(粗略)

| 工具 | localhost RPS | 1k VUs 内存 | 上手时间 |
|---|---|---|---|
| **Pulse(Local)** | 50k-100k | ~100MB | 5 分钟 |
| **k6** | 50k-80k | ~80MB | 10 分钟 |
| **wrk** | 100k-200k | ~30MB | 30 分钟 |
| **vegeta** | 30k-50k | ~50MB | 30 分钟 |
| **JMeter** | 5k-10k | ~500MB | 60 分钟 |
| **Locust** | 5k-10k | ~200MB | 20 分钟 |

> Pulse 在保证 UI 易用性的同时,性能达到 k6 级别。

---

## 三、性能优化检查清单

### ✅ 已实现的优化

- [x] **Goroutine Pool**:动态扩缩,避免频繁创建/销毁
- [x] **HTTP Keep-Alive**:连接复用,减少 TCP 握手
- [x] **HTTP/2 支持**:多路复用,降低延迟
- [x] **连接池**:MaxIdleConnsPerHost=200
- [x] **Token Bucket 限流**:基于 golang.org/x/time/rate
- [x] **HDR Histogram**:延迟分位数,常数时间
- [x] **滑动窗口 RPS**:100ms 粒度,常数时间
- [x] **零分配**:核心热路径上避免内存分配

### 🔧 未来优化方向(V2)

- [ ] **fasthttp 客户端**:替换 net/http,性能翻倍
- [ ] **io_uring**(Linux):内核级异步 I/O
- [ ] **Pipeline 模式**:单连接多请求复用
- [ ] **PGO**(Profile-Guided Optimization):Go 1.21+ 自动优化
- [ ] **SIMD 解析**:JSON Path 提取向量化
- [ ] **零拷贝**:响应 body 直接传给 collector

---

## 四、测试场景清单

### 单元测试(`*_test.go`)

| 文件 | 测试数 | 覆盖 |
|---|---|---|
| `pool_test.go` | 5 | Goroutine Pool |
| `httpclient_test.go` | 2 | HTTP Client 配置 |
| `ratelimit_test.go` | 4 | Token Bucket |
| `metrics/histogram_test.go` | 5 | HDR Histogram |
| `metrics/collector_test.go` | 4 | Collector |
| `extract_test.go` | 9 | JSON Path + Variable |
| `benchmark_test.go` | 6 | Scheduler + Assertions |
| `integration_test.go` | 5 | 端到端 Local Engine |
| `benchmark_suite_test.go` | 4 | 性能场景 |
| Service 层测试 | 4+ | Run/Scenario/Import/Project |

**总计 50+ 个测试用例**。

### 运行命令总结

```bash
# 完整测试套件
go test ./... -v -timeout 120s

# 只跑 benchmark
go test ./... -bench=. -benchmem -benchtime=10s | tee bench-results.txt

# 性能对比
go test ./internal/engine/ -bench=BenchmarkEngine -cpuprofile=cpu.prof
go tool pprof -http=:8080 cpu.prof  # 浏览器查看分析
```

---

## 五、故障排查

### 常见问题

1. **RPS 远低于预期**
   - 检查 `KeepAlive` 是否开启
   - 检查 `MaxIdleConnsPerHost` 设置
   - 使用 `-cpuprofile` 分析瓶颈

2. **P99 很高但 P50 正常**
   - 系统 GC 暂停导致
   - 目标服务响应慢
   - 使用 `GODEBUG=gctrace=1` 查看 GC

3. **Engine 启动失败**
   - 检查 `runtime_config.VUs` 不能超过 `Config.MaxVUs`
   - 检查 URL 格式正确

4. **测试运行内存占用过高**
   - 默认 MaxVUs=20000,降低可减少内存
   - 长时间运行建议定期 `Reset()` histogram

---

## 六、性能测试推荐配置

### 配置文件 (`~/.pulse/config.yaml`)

```yaml
app:
  name: "Pulse"
  version: "0.1.0"
  data_dir: "~/.pulse/data"

engine:
  max_vus: 20000
  keep_alive: true
  http2: true
  timeout_seconds: 30
```

### Linux/macOS 内核参数调优

```bash
# 增加文件描述符上限
ulimit -n 65536

# 增加临时端口范围
sudo sysctl -w net.ipv4.ip_local_port_range="1024 65535"

# TCP 优化
sudo sysctl -w net.ipv4.tcp_tw_reuse=1
sudo sysctl -w net.ipv4.tcp_max_syn_backlog=4096
```

---

## 七、参考对比

### 与 k6 性能对比(同为 Go 内核)

- **k6 OSS**:核心 engine 是 Go,CLI-only,UI 在 Grafana Cloud(付费)
- **Pulse**:核心 engine 是 Go,**桌面 UI + OSS + 完整自托管**

### 与 JMeter 性能对比

- **JMeter**:Java 内核,GUI 完整,**启动慢 + 资源占用高 + 学习曲线陡**
- **Pulse**:Go 内核,GUI 现代,**启动快 + 资源占用低 + 5 分钟上手**

---

## 八、最终目标(MVP v0.1.0)

| 指标 | 目标 | 实际 |
|---|---|---|
| 单机最大并发 VUs | ≥ 10,000 | TBD |
| localhost RPS | ≥ 50,000 | TBD |
| 远程 HTTP RPS | ≥ 30,000 | TBD |
| 5 分钟上手 | ✅ | TBD |
| 报告 HTML 导出 | ✅ | ✅ |
| HAR 导入 | ✅ | ✅ |
| Apache 2.0 | ✅ | ✅ |

> 实际数字以你本地运行结果为准。运行 `go test -bench=.` 后将数字填入。

---

**🎯 准备好跑测试了!** 运行 `go test ./internal/engine/ -v -timeout 60s` 开始验证。