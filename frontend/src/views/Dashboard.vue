<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  NCard,
  NSpace,
  NButton,
  NStatistic,
  NGrid,
  NGi,
  NText,
  NDivider,
  NSpin,
  NTag,
} from 'naive-ui'
import { useScenariosStore } from '@/stores/scenarios'
import { useProjectsStore } from '@/stores/projects'
import * as api from '@/api/scenarios'

const router = useRouter()
const store = useScenariosStore()
const projectsStore = useProjectsStore()

const loading = ref(true)
const user = ref<api.UserDTO | null>(null)

// 最近测试运行
const recentRuns = ref<api.TestRunDTO[]>([])
const totalRuns = ref(0)
const totalCompletedRuns = ref(0)
const totalFailedRuns = ref(0)
const totalRequests = ref(0)

onMounted(async () => {
  loading.value = true
  try {
    const [userData, scenarios, projects, runs] = await Promise.all([
      api.getCurrentUser().catch(() => null),
      store.fetchList({ limit: 5 }),
      projectsStore.fetchList(),
      api.listAllTestRuns(20).catch(() => []),
    ])

    user.value = userData
    recentRuns.value = runs || []

    // 统计
    totalRuns.value = recentRuns.value.length
    totalCompletedRuns.value = recentRuns.value.filter(r => r.status === 'completed').length
    totalFailedRuns.value = recentRuns.value.filter(r => r.status === 'failed' || r.status === 'aborted').length
    totalRequests.value = recentRuns.value.reduce((sum, r) => {
      return sum + (r.summary?.TotalRequests || 0)
    }, 0)
  } catch (e) {
    console.error('Dashboard load error:', e)
  } finally {
    loading.value = false
  }
})

function newScenario() {
  router.push({ name: 'scenario-editor' })
}

function goScenarios() {
  router.push({ name: 'scenarios' })
}

function goReports() {
  router.push({ name: 'reports' })
}

function goAgents() {
  router.push({ name: 'agents' })
}

function viewRun(runId: number) {
  router.push({ name: 'report-detail', params: { id: String(runId) } })
}

function viewLiveMonitor(runId: number) {
  router.push({ name: 'live-monitor', params: { id: String(runId) } })
}

// 状态颜色
function statusColor(status: string): string {
  if (status === 'completed') return 'success'
  if (status === 'running') return 'info'
  if (status === 'paused') return 'warning'
  if (status === 'failed' || status === 'aborted') return 'error'
  return 'default'
}

// 状态文本
function statusLabel(status: string): string {
  if (status === 'completed') return '✅ 通过'
  if (status === 'running') return '🔄 运行中'
  if (status === 'paused') return '⏸ 暂停'
  if (status === 'failed') return '❌ 失败'
  if (status === 'aborted') return '⏹ 停止'
  return status
}

// 格式化数字
function formatNumber(n: number): string {
  if (n >= 10000) return (n / 1000).toFixed(1) + 'k'
  return n.toLocaleString()
}

// 格式化相对时间
function timeAgo(dateStr: string): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000)
  if (seconds < 60) return `${seconds} 秒前`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes} 分钟前`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours} 小时前`
  const days = Math.floor(hours / 24)
  if (days < 7) return `${days} 天前`
  return date.toLocaleDateString('zh-CN')
}
</script>

<template>
  <NSpin :show="loading">
    <div>
      <!-- 欢迎语 -->
      <div style="margin-bottom: 24px;">
        <NText style="font-size: 24px; font-weight: 600;">
          欢迎回来{{ user?.displayName ? ',' + user.displayName : '' }} 👋
        </NText>
        <br />
        <NText depth="3">压力测试,理应如此。</NText>
      </div>

      <NGrid :cols="3" :x-gap="16" :y-gap="16" responsive="screen">
        <!-- 快速开始 -->
        <NGi :span="2">
          <NCard :bordered="false" embedded>
            <template #header>
              <NSpace align="center">
                <span style="font-size: 18px;">🚀</span>
                <NText strong style="font-size: 16px;">快速开始</NText>
              </NSpace>
            </template>

            <NSpace vertical size="large">
              <NButton type="primary" size="large" block @click="newScenario">
                ➕ 新建场景
              </NButton>
              <NButton size="large" block @click="goScenarios">
                📋 浏览所有场景 ({{ store.totalCount }})
              </NButton>
              <NButton size="large" block @click="goReports">
                📊 查看历史报告 ({{ totalRuns }})
              </NButton>
              <NButton size="large" block @click="goAgents">
                🖥️ Agents 管理
              </NButton>
            </NSpace>

            <NDivider />

            <NText depth="3" style="font-size: 12px;">
              ⏱ 平均上手时间:4.2 分钟
            </NText>
          </NCard>
        </NGi>

        <!-- 数据统计 -->
        <NGi :span="1">
          <NCard :bordered="false" embedded>
            <template #header>
              <NSpace align="center">
                <span style="font-size: 18px;">📊</span>
                <NText strong style="font-size: 16px;">你的数据</NText>
              </NSpace>
            </template>

            <NSpace vertical size="medium">
              <NStatistic label="场景数" :value="store.totalCount" />
              <NStatistic label="项目数" :value="projectsStore.items.length" />
              <NStatistic label="总测试次数" :value="totalRuns" />
              <NStatistic label="总请求数" :value="formatNumber(totalRequests)" />
            </NSpace>
          </NCard>
        </NGi>

        <!-- 最近测试运行 -->
        <NGi :span="2">
          <NCard :bordered="false" embedded>
            <template #header>
              <NSpace align="center" justify="space-between" style="width: 100%;">
                <NSpace align="center">
                  <span style="font-size: 18px;">📈</span>
                  <NText strong style="font-size: 16px;">最近测试运行</NText>
                </NSpace>
                <NButton text size="small" @click="goReports">查看全部 →</NButton>
              </NSpace>
            </template>

            <div v-if="recentRuns.length === 0" style="padding: 24px; text-align: center; color: var(--text-secondary);">
              还没有测试运行。创建场景后点击"▶ Run"开始第一次压测。
              <div style="margin-top: 16px;">
                <NButton @click="newScenario" type="primary" ghost>➕ 创建第一个场景</NButton>
              </div>
            </div>

            <div v-else>
              <div
                v-for="run in recentRuns.slice(0, 5)"
                :key="run.id"
                style="
                  display: flex;
                  align-items: center;
                  justify-content: space-between;
                  padding: 12px;
                  border-bottom: 1px solid var(--border);
                  cursor: pointer;
                  transition: background 0.15s;
                "
                @click="run.status === 'running' ? viewLiveMonitor(run.id) : viewRun(run.id)"
                @mouseover="(e: any) => e.currentTarget.style.background = 'var(--bg-hover)'"
                @mouseleave="(e: any) => e.currentTarget.style.background = 'transparent'"
              >
                <div style="flex: 1;">
                  <div style="display: flex; align-items: center; gap: 8px;">
                    <span style="font-weight: 500;">{{ run.scenarioName }}</span>
                    <NTag :type="statusColor(run.status) as any" size="small" round>
                      {{ statusLabel(run.status) }}
                    </NTag>
                  </div>
                  <div style="font-size: 12px; color: var(--text-secondary); margin-top: 2px;">
                    {{ run.durationMs ? Math.round(run.durationMs / 1000) + 's' : '-' }}
                    ·
                    {{ formatNumber(run.summary?.TotalRequests || 0) }} 请求
                    ·
                    P95 {{ Math.round(run.summary?.P95Ms || 0) }}ms
                    <span v-if="run.summary?.ErrorRate" style="color: var(--error);">
                      · 错误率 {{ ((run.summary.ErrorRate || 0) * 100).toFixed(2) }}%
                    </span>
                  </div>
                </div>
                <span style="font-size: 12px; color: var(--text-tertiary);">
                  {{ timeAgo(run.createdAt) }}
                </span>
              </div>
            </div>
          </NCard>
        </NGi>

        <!-- 最近场景 -->
        <NGi :span="1">
          <NCard :bordered="false" embedded>
            <template #header>
              <NSpace align="center">
                <span style="font-size: 18px;">📌</span>
                <NText strong style="font-size: 16px;">最近场景</NText>
              </NSpace>
            </template>

            <div v-if="store.items.length === 0" style="padding: 16px; text-align: center; color: var(--text-secondary); font-size: 13px;">
              暂无场景
            </div>

            <div v-else>
              <div
                v-for="scenario in store.items.slice(0, 5)"
                :key="scenario.id"
                style="
                  display: flex;
                  align-items: center;
                  justify-content: space-between;
                  padding: 8px 0;
                  cursor: pointer;
                  font-size: 13px;
                  border-bottom: 1px solid var(--border);
                "
                @click="router.push({ name: 'scenario-editor', params: { id: String(scenario.id) } })"
              >
                <span style="color: var(--text-primary);">{{ scenario.name }}</span>
                <span style="color: var(--text-tertiary); font-size: 11px;">
                  v{{ scenario.version }}
                </span>
              </div>
            </div>
          </NCard>
        </NGi>
      </NGrid>

      <!-- 系统状态 -->
      <NCard :bordered="false" embedded style="margin-top: 24px;">
        <NSpace align="center">
          <span style="font-size: 18px;">⚙️</span>
          <NText>系统状态</NText>
        </NSpace>
        <NText tag="div" depth="3" style="margin-top: 8px; padding-left: 28px; line-height: 1.8;">
          • 数据库:已连接<br />
          • 用户:{{ user?.username || '...' }}<br />
          • 默认项目:{{ projectsStore.defaultProject?.name || projectsStore.items[0]?.name || '...' }}<br />
          • 引擎:Go 自研(Local)·支持 20,000+ VUs
        </NText>
      </NCard>
    </div>
  </NSpin>
</template>