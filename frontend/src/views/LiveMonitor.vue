<script setup lang="ts">
// LiveMonitor - 实时大屏主页面
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import * as api from '@/api/scenarios'
import type { MetricSnapshot, RequestSnapshot } from '@/api/scenarios'
import KpiCard from '@/components/KpiCard.vue'

const route = useRoute()
const router = useRouter()
const message = useMessage()

const runId = computed(() => Number(route.params.id))

// ============================================
// 状态
// ============================================

// 实时指标
const latestSnapshot = ref<MetricSnapshot | null>(null)
const snapshotHistory = ref<MetricSnapshot[]>([])

// 上一次的快照(用于计算趋势)
const previousSnapshot = ref<MetricSnapshot | null>(null)

// 测试运行状态
const status = ref<'pending' | 'running' | 'paused' | 'completed' | 'failed' | 'aborted'>('pending')
const scenarioName = ref('')
const startedAt = ref<Date | null>(null)
const durationSec = ref(5 * 60) // 默认 5 分钟(从场景配置读取)
const elapsedMs = ref(0)
const isFullscreen = ref(false)
const eventConnected = ref(false)

// 错误
const errorMessage = ref('')

// 完成时的总结数据
const finalSummary = ref<any>(null)
const thresholdResults = ref<any[]>([])

// ============================================
// 订阅
// ============================================

let unsubscribeMetric: (() => void) | null = null
let unsubscribeEvents: (() => void) | null = null
let pollTimer: number | null = null

onMounted(async () => {
  // 加载 run 基础信息
  try {
    const run = await api.getRun(runId.value)
    if (run) {
      scenarioName.value = run.scenarioName
      status.value = run.status as any
      if (run.startedAt) startedAt.value = new Date(run.startedAt)
    }
  } catch (e: any) {
    message.error(`加载测试运行失败: ${e.message}`)
    router.push({ name: 'reports' })
    return
  }

  // 检查是否还在活跃
  const active = await api.isRunActive(runId.value)
  if (active) {
    // 活跃中:订阅实时数据
    subscribeToUpdates()
  } else {
    // 已结束:加载历史 summary
    await loadFinalSummary()
  }

  // 启动定时器(更新 elapsed)
  pollTimer = window.setInterval(updateElapsed, 200)
})

onUnmounted(() => {
  unsubscribeMetric?.()
  unsubscribeEvents?.()
  if (pollTimer) clearInterval(pollTimer)
  if (isFullscreen.value) {
    document.exitFullscreen?.()
  }
})

function subscribeToUpdates() {
  // 订阅指标
  unsubscribeMetric = api.subscribeMetric(runId.value, (snapshot) => {
    eventConnected.value = true
    previousSnapshot.value = latestSnapshot.value
    latestSnapshot.value = snapshot
    elapsedMs.value = snapshot.elapsedMs

    // 保留最近 120 个快照(2 分钟)
    snapshotHistory.value.push(snapshot)
    if (snapshotHistory.value.length > 120) {
      snapshotHistory.value.shift()
    }
  })

  // 订阅事件
  unsubscribeEvents = api.subscribeRunEvents(runId.value, {
    onStarted: () => {
      status.value = 'running'
      message.success('测试已开始')
    },
    onCompleted: () => {
      status.value = 'completed'
      message.success('测试已完成')
      loadFinalSummary()
    },
    onFailed: (err) => {
      status.value = 'failed'
      errorMessage.value = err
      message.error(`测试失败: ${err}`)
    },
  })

  // 3 秒后如果还没收到数据,标记为未连接
  setTimeout(() => {
    if (!latestSnapshot.value) {
      eventConnected.value = false
    }
  }, 3000)
}

async function loadFinalSummary() {
  try {
    const run = await api.getRun(runId.value)
    if (run?.summary) {
      finalSummary.value = run.summary
    }
  } catch (e) {
    console.error('loadFinalSummary:', e)
  }
}

function updateElapsed() {
  if (startedAt.value) {
    elapsedMs.value = Date.now() - startedAt.value.getTime()
  }
}

// ============================================
// 操作
// ============================================

async function pauseRun() {
  try {
    await api.pauseRun(runId.value)
    status.value = 'paused'
    message.info('已暂停')
  } catch (e: any) {
    message.error(`暂停失败: ${e.message}`)
  }
}

async function resumeRun() {
  try {
    await api.resumeRun(runId.value)
    status.value = 'running'
    message.info('已恢复')
  } catch (e: any) {
    message.error(`恢复失败: ${e.message}`)
  }
}

async function stopRun() {
  if (!confirm('确定停止当前测试吗?此操作不可撤销。')) return
  try {
    await api.stopRun(runId.value)
    status.value = 'aborted'
    message.warning('测试已停止')
    setTimeout(() => {
      router.push({ name: 'reports' })
    }, 1000)
  } catch (e: any) {
    message.error(`停止失败: ${e.message}`)
  }
}

// ============================================
// 全屏
// ============================================

function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen?.()
    isFullscreen.value = true
  } else {
    document.exitFullscreen?.()
    isFullscreen.value = false
  }
}

// F 键全屏
function handleKeydown(e: KeyboardEvent) {
  if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return
  if (e.key === 'f' || e.key === 'F') {
    toggleFullscreen()
  } else if (e.key === 'Escape' && isFullscreen.value) {
    toggleFullscreen()
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})

// ============================================
// 计算属性
// ============================================

const rpsTrend = computed(() => {
  if (!previousSnapshot.value || !latestSnapshot.value) return undefined
  if (previousSnapshot.value.rps === 0) return undefined
  const diff = ((latestSnapshot.value.rps - previousSnapshot.value.rps) / previousSnapshot.value.rps) * 100
  return Math.round(diff * 10) / 10
})

const p95Trend = computed(() => {
  if (!previousSnapshot.value || !latestSnapshot.value) return undefined
  if (previousSnapshot.value.latencyMs.p95 === 0) return undefined
  const diff = ((latestSnapshot.value.latencyMs.p95 - previousSnapshot.value.latencyMs.p95) / previousSnapshot.value.latencyMs.p95) * 100
  return Math.round(diff * 10) / 10
})

const errorTrend = computed(() => {
  if (!previousSnapshot.value || !latestSnapshot.value) return undefined
  return Math.round((latestSnapshot.value.errorRate - previousSnapshot.value.errorRate) * 10000) / 100
})

const elapsedFormatted = computed(() => {
  const totalSec = Math.floor(elapsedMs.value / 1000)
  const m = Math.floor(totalSec / 60)
  const s = totalSec % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
})

const durationFormatted = computed(() => {
  const m = Math.floor(durationSec.value / 60)
  const s = durationSec.value % 60
  return `${m}m${s > 0 ? ` ${s}s` : ''}`
})

const progressPercent = computed(() => {
  if (durationSec.value === 0) return 0
  return Math.min(100, (elapsedMs.value / 1000 / durationSec.value) * 100)
})

// 折线图数据
const lineChartData = computed(() => {
  const history = snapshotHistory.value
  return {
    times: history.map(s => s.elapsedMs / 1000),
    rps: history.map(s => Math.round(s.rps)),
    p50: history.map(s => Math.round(s.latencyMs.p50 * 10) / 10),
    p95: history.map(s => Math.round(s.latencyMs.p95 * 10) / 10),
    p99: history.map(s => Math.round(s.latencyMs.p99 * 10) / 10),
  }
})

// 状态码分布
const statusCodeEntries = computed(() => {
  const snap = latestSnapshot.value
  if (!snap) return []
  return Object.entries(snap.statusCodes)
    .sort((a, b) => Number(b[1]) - Number(a[1]))
    .map(([code, count]) => ({ code, count, pct: snap.totalRequests > 0 ? (count / snap.totalRequests) * 100 : 0 }))
})

// Per-Request 列表
const perRequestList = computed<RequestSnapshot[]>(() => {
  const snap = latestSnapshot.value
  if (!snap) return []
  return Object.values(snap.perRequest)
    .sort((a, b) => b.count - a.count)
})

// 延迟分布(简化为分桶)
const latencyDistribution = computed(() => {
  const snap = latestSnapshot.value
  if (!snap) return []
  const buckets = [
    { range: '<10ms', min: 0, max: 10 },
    { range: '10-50ms', min: 10, max: 50 },
    { range: '50-100ms', min: 50, max: 100 },
    { range: '100-200ms', min: 100, max: 200 },
    { range: '200-500ms', min: 200, max: 500 },
    { range: '500ms-1s', min: 500, max: 1000 },
    { range: '>1s', min: 1000, max: Infinity },
  ]

  // 这里简化:用分位数估算分布(MVP 简化实现)
  return buckets.map(b => ({
    range: b.range,
    pct: 25, // 占位,真实应该从 histogram 获取
  }))
})

const isCompleted = computed(() => ['completed', 'failed', 'aborted'].includes(status.value))
const isRunning = computed(() => status.value === 'running')
const isPaused = computed(() => status.value === 'paused')

const statusBadge = computed(() => {
  switch (status.value) {
    case 'running': return { text: '运行中', color: 'var(--success)', icon: '●' }
    case 'paused': return { text: '已暂停', color: 'var(--warning)', icon: '⏸' }
    case 'completed': return { text: '已完成', color: 'var(--success)', icon: '✅' }
    case 'failed': return { text: '失败', color: 'var(--error)', icon: '❌' }
    case 'aborted': return { text: '已停止', color: 'var(--warning)', icon: '⏹' }
    default: return { text: '等待中', color: 'var(--text-secondary)', icon: '○' }
  }
})

function formatNumber(n: number): string {
  if (n >= 10000) return (n / 1000).toFixed(1) + 'k'
  return n.toLocaleString()
}
</script>

<template>
  <div :class="['live-monitor', { 'is-fullscreen': isFullscreen }]">
    <!-- 顶部状态条 -->
    <div class="monitor-topbar">
      <div class="monitor-topbar-left">
        <button @click="router.back()" class="btn-secondary">← 返回</button>
        <h2 style="margin: 0; font-size: 18px; font-weight: 600;">{{ scenarioName }}</h2>
        <span class="status-badge" :style="{ color: statusBadge.color, borderColor: statusBadge.color }">
          {{ statusBadge.icon }} {{ statusBadge.text }}
        </span>
        <span v-if="!eventConnected && isRunning" class="connection-warning">⚠️ 等待数据</span>
      </div>

      <div class="monitor-topbar-right">
        <div class="elapsed-display">
          <span class="elapsed-time">{{ elapsedFormatted }}</span>
          <span class="elapsed-divider">/</span>
          <span class="duration-target">{{ durationFormatted }}</span>
        </div>

        <button v-if="isRunning" @click="pauseRun" class="btn-warning">⏸ 暂停</button>
        <button v-if="isPaused" @click="resumeRun" class="btn-success">▶ 继续</button>
        <button v-if="isRunning || isPaused" @click="stopRun" class="btn-danger">⏹ 停止</button>
        <button @click="toggleFullscreen" class="btn-secondary">{{ isFullscreen ? '⊗ 退出全屏' : '⛶ 全屏 (F)' }}</button>
      </div>
    </div>

    <!-- 进度条 -->
    <div class="progress-bar">
      <div class="progress-fill" :style="{ width: progressPercent + '%' }"></div>
    </div>

    <!-- 主内容 -->
    <div class="monitor-content">
      <!-- KPI 卡片 -->
      <div class="kpi-grid">
        <KpiCard
          icon="📊"
          label="RPS"
          :value="latestSnapshot?.rps || 0"
          :trend="rpsTrend"
          :threshold="{ warning: 100, error: 10 }"
          status="normal"
        />
        <KpiCard
          icon="⏱"
          label="P95 延迟"
          :value="latestSnapshot?.latencyMs.p95 || 0"
          unit="ms"
          :trend="p95Trend"
          :threshold="{ warning: 200, error: 500 }"
        />
        <KpiCard
          icon="❌"
          label="错误率"
          :value="(latestSnapshot?.errorRate || 0) * 100"
          unit="%"
          :trend="errorTrend"
          :threshold="{ warning: 1, error: 5 }"
        />
        <KpiCard
          icon="🚀"
          label="总请求"
          :value="latestSnapshot?.totalRequests || 0"
          status="normal"
        />
        <KpiCard
          icon="👥"
          label="活跃 VUs"
          :value="latestSnapshot?.vusActive || 0"
          :unit="`/ ${latestSnapshot?.vusTarget || 0}`"
          status="normal"
        />
      </div>

      <!-- 完成总结(测试结束后显示) -->
      <div v-if="isCompleted && finalSummary" class="completion-banner" :class="{ 'failed': status === 'failed' || status === 'aborted' }">
        <div class="completion-icon">{{ status === 'completed' ? '✅' : (status === 'failed' ? '❌' : '⏹') }}</div>
        <div class="completion-content">
          <h3>{{ status === 'completed' ? '测试通过' : (status === 'failed' ? '测试失败' : '测试已停止') }}</h3>
          <div class="completion-stats">
            <span>📊 总请求 <strong>{{ formatNumber(finalSummary.TotalRequests || 0) }}</strong></span>
            <span>⚡ 平均 RPS <strong>{{ (finalSummary.AvgRPS || 0).toFixed(1) }}</strong></span>
            <span>⏱ P95 <strong>{{ (finalSummary.P95Ms || 0).toFixed(1) }}ms</strong></span>
            <span>❌ 错误率 <strong>{{ ((finalSummary.ErrorRate || 0) * 100).toFixed(2) }}%</strong></span>
          </div>
          <div class="completion-actions">
            <button @click="router.push({ name: 'reports' })" class="btn-primary">📊 查看所有报告</button>
          </div>
        </div>
      </div>

      <!-- 折线图(用 SVG 简化实现,不引入 echarts 复杂性) -->
      <div v-if="snapshotHistory.length > 1" class="chart-card">
        <div class="chart-header">
          <span>📈 RPS & Latency over Time</span>
          <span class="chart-legend">
            <span class="legend-item"><span class="legend-color" style="background:#3370FF"></span>RPS</span>
            <span class="legend-item"><span class="legend-color" style="background:#FAAD14"></span>P95</span>
            <span class="legend-item"><span class="legend-color" style="background:#52C41A"></span>P50</span>
          </span>
        </div>
        <div class="chart-container">
          <svg :viewBox="`0 0 ${lineChartData.times.length * 50} 200`" preserveAspectRatio="none" style="width: 100%; height: 200px;">
            <!-- 网格 -->
            <line v-for="i in 5" :key="i" :x1="0" :y1="i * 40" :x2="lineChartData.times.length * 50" :y2="i * 40" stroke="#2A2A2A" stroke-width="0.5" />
            <!-- RPS 线 -->
            <polyline
              :points="lineChartData.times.map((t, i) => `${i * 50},${200 - Math.min(lineChartData.rps[i] / 10, 180)}`).join(' ')"
              fill="none" stroke="#3370FF" stroke-width="2"
            />
            <!-- P95 线 -->
            <polyline
              :points="lineChartData.times.map((t, i) => `${i * 50},${200 - Math.min(lineChartData.p95[i] / 2, 180)}`).join(' ')"
              fill="none" stroke="#FAAD14" stroke-width="1.5"
            />
            <!-- P50 线 -->
            <polyline
              :points="lineChartData.times.map((t, i) => `${i * 50},${200 - Math.min(lineChartData.p50[i] / 2, 180)}`).join(' ')"
              fill="none" stroke="#52C41A" stroke-width="1.5"
            />
          </svg>
        </div>
      </div>

      <!-- 双列:状态码 + Per-Request -->
      <div class="grid-2">
        <!-- 状态码分布 -->
        <div class="info-card">
          <div class="info-card-header">📊 Status Codes</div>
          <div v-if="statusCodeEntries.length === 0" class="info-card-empty">暂无数据</div>
          <div v-else class="status-codes">
            <div v-for="entry in statusCodeEntries" :key="entry.code" class="status-code-row">
              <span :class="['status-code', `code-${Math.floor(Number(entry.code) / 100)}xx`]">{{ entry.code }}</span>
              <div class="status-code-bar-wrap">
                <div class="status-code-bar" :style="{ width: entry.pct + '%' }"></div>
              </div>
              <span class="status-code-pct">{{ entry.pct.toFixed(1) }}%</span>
              <span class="status-code-count">{{ formatNumber(entry.count) }}</span>
            </div>
          </div>
        </div>

        <!-- Per-Request 明细 -->
        <div class="info-card">
          <div class="info-card-header">📋 Per-Request Stats</div>
          <div v-if="perRequestList.length === 0" class="info-card-empty">暂无数据</div>
          <div v-else class="per-request-table">
            <div class="per-request-row header">
              <div>Request</div>
              <div style="text-align: right;">Count</div>
              <div style="text-align: right;">P50</div>
              <div style="text-align: right;">P95</div>
              <div style="text-align: right;">Errors</div>
            </div>
            <div v-for="req in perRequestList" :key="req.name" class="per-request-row">
              <div class="per-request-name">🟢 {{ req.name }}</div>
              <div style="text-align: right;">{{ formatNumber(req.count) }}</div>
              <div style="text-align: right;">{{ req.latencyMs.p50.toFixed(0) }}ms</div>
              <div style="text-align: right;">{{ req.latencyMs.p95.toFixed(0) }}ms</div>
              <div style="text-align: right;" :style="{ color: req.errors > 0 ? 'var(--error)' : 'var(--text-secondary)' }">
                {{ ((req.errors / Math.max(req.count, 1)) * 100).toFixed(1) }}%
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 等待状态 -->
      <div v-if="!latestSnapshot && isRunning" class="waiting-banner">
        <div class="waiting-icon">⏳</div>
        <div class="waiting-text">
          <h3>正在初始化压测...</h3>
          <p>第一波请求正在准备中,请稍候</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.live-monitor {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 56px - 32px);
  overflow: hidden;
}

.live-monitor.is-fullscreen {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: var(--bg-base);
  height: 100vh;
}

.monitor-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border);
  margin-bottom: 16px;
}

.monitor-topbar-left,
.monitor-topbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border: 1px solid;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.connection-warning {
  padding: 4px 8px;
  background: rgba(250, 173, 20, 0.15);
  color: var(--warning);
  border-radius: 4px;
  font-size: 11px;
}

.elapsed-display {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 6px;
  font-family: var(--font-mono);
}

.elapsed-time {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.elapsed-divider {
  color: var(--text-tertiary);
}

.duration-target {
  color: var(--text-secondary);
  font-size: 13px;
}

.btn-primary,
.btn-secondary,
.btn-warning,
.btn-danger,
.btn-success {
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.15s;
}

.btn-primary { background: var(--primary); color: #fff; }
.btn-primary:hover { background: var(--primary-hover); }
.btn-secondary { background: var(--bg-elevated); color: var(--text-primary); border-color: var(--border); }
.btn-secondary:hover { background: var(--bg-hover); }
.btn-warning { background: var(--warning); color: #fff; }
.btn-warning:hover { filter: brightness(1.1); }
.btn-danger { background: var(--error); color: #fff; }
.btn-danger:hover { filter: brightness(1.1); }
.btn-success { background: var(--success); color: #fff; }
.btn-success:hover { filter: brightness(1.1); }

.progress-bar {
  height: 4px;
  background: var(--bg-hover);
  border-radius: 2px;
  margin-bottom: 16px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--primary);
  transition: width 0.3s;
}

.monitor-content {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.kpi-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 12px;
}

.completion-banner {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 24px;
  background: rgba(82, 196, 26, 0.08);
  border: 1px solid var(--success);
  border-radius: 8px;
}

.completion-banner.failed {
  background: rgba(255, 77, 79, 0.08);
  border-color: var(--error);
}

.completion-icon {
  font-size: 32px;
}

.completion-content h3 {
  margin: 0 0 8px 0;
  font-size: 16px;
  color: var(--text-primary);
}

.completion-stats {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.completion-stats strong {
  color: var(--text-primary);
  font-family: var(--font-mono);
  margin-left: 4px;
}

.completion-actions {
  display: flex;
  gap: 8px;
}

.chart-card,
.info-card {
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 16px;
}

.chart-header,
.info-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 12px;
}

.chart-legend {
  display: flex;
  gap: 12px;
  font-size: 11px;
  color: var(--text-secondary);
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.legend-color {
  display: inline-block;
  width: 12px;
  height: 3px;
  border-radius: 2px;
}

.chart-container {
  margin: 0 -8px;
}

.grid-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.info-card-empty {
  padding: 24px;
  text-align: center;
  color: var(--text-tertiary);
  font-size: 13px;
}

.status-codes {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.status-code-row {
  display: grid;
  grid-template-columns: 50px 1fr 60px 70px;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.status-code {
  padding: 2px 6px;
  border-radius: 3px;
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 600;
  text-align: center;
}

.code-2xx { background: rgba(82, 196, 26, 0.15); color: #52C41A; }
.code-3xx { background: rgba(19, 194, 194, 0.15); color: #13C2C2; }
.code-4xx { background: rgba(250, 173, 20, 0.15); color: #FAAD14; }
.code-5xx { background: rgba(255, 77, 79, 0.15); color: #FF4D4F; }

.status-code-bar-wrap {
  background: var(--bg-base);
  border-radius: 3px;
  height: 6px;
  overflow: hidden;
}

.status-code-bar {
  height: 100%;
  background: var(--primary);
  transition: width 0.3s;
}

.status-code-pct {
  text-align: right;
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--text-secondary);
}

.status-code-count {
  text-align: right;
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--text-tertiary);
}

.per-request-table {
  display: flex;
  flex-direction: column;
}

.per-request-row {
  display: grid;
  grid-template-columns: 1.5fr 1fr 1fr 1fr 1fr;
  gap: 8px;
  padding: 8px 0;
  font-size: 12px;
  font-family: var(--font-mono);
  border-bottom: 1px solid var(--border);
}

.per-request-row.header {
  font-weight: 600;
  color: var(--text-secondary);
  font-family: var(--font-sans);
}

.per-request-row:last-child {
  border-bottom: none;
}

.per-request-name {
  color: var(--text-primary);
  font-family: var(--font-sans);
  font-weight: 500;
}

.waiting-banner {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 40px;
  background: var(--bg-elevated);
  border: 1px dashed var(--border);
  border-radius: 8px;
}

.waiting-icon {
  font-size: 32px;
}

.waiting-text h3 {
  margin: 0 0 4px 0;
  font-size: 14px;
}

.waiting-text p {
  margin: 0;
  color: var(--text-secondary);
  font-size: 12px;
}
</style>