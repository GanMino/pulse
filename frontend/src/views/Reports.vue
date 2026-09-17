<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  NCard,
  NButton,
  NSpace,
  NText,
  NEmpty,
  NTag,
  NInput,
  useMessage,
} from 'naive-ui'
import * as api from '@/api/scenarios'

const router = useRouter()
const message = useMessage()

const loading = ref(true)
const runs = ref<api.TestRunDTO[]>([])
const searchQuery = ref('')
const statusFilter = ref<string>('') // '' = all

// 过滤
const filteredRuns = computed(() => {
  let result = runs.value

  if (statusFilter.value) {
    result = result.filter(r => r.status === statusFilter.value)
  }

  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter(r =>
      r.scenarioName.toLowerCase().includes(q) ||
      String(r.id).includes(q)
    )
  }

  return result
})

// 统计
const stats = computed(() => {
  return {
    total: runs.value.length,
    completed: runs.value.filter(r => r.status === 'completed').length,
    failed: runs.value.filter(r => r.status === 'failed' || r.status === 'aborted').length,
    running: runs.value.filter(r => r.status === 'running' || r.status === 'paused').length,
  }
})

onMounted(async () => {
  loading.value = true
  try {
    runs.value = await api.listAllTestRuns(100)
  } catch (e: any) {
    message.error(`加载报告失败: ${e.message}`)
  } finally {
    loading.value = false
  }
})

function viewReport(runId: number, status: string) {
  if (status === 'running' || status === 'paused') {
    router.push({ name: 'live-monitor', params: { id: String(runId) } })
  } else {
    router.push({ name: 'report-detail', params: { id: String(runId) } })
  }
}

function statusColor(status: string): string {
  if (status === 'completed') return 'success'
  if (status === 'running') return 'info'
  if (status === 'paused') return 'warning'
  if (status === 'failed' || status === 'aborted') return 'error'
  return 'default'
}

function statusLabel(status: string): string {
  if (status === 'completed') return '✅ 通过'
  if (status === 'running') return '🔄 运行中'
  if (status === 'paused') return '⏸ 暂停'
  if (status === 'failed') return '❌ 失败'
  if (status === 'aborted') return '⏹ 停止'
  return status
}

function formatNumber(n: number): string {
  if (n >= 10000) return (n / 1000).toFixed(1) + 'k'
  return n.toLocaleString()
}

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
  <div>
    <!-- 顶部 -->
    <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px;">
      <h2 style="margin: 0; font-size: 20px; font-weight: 600;">Reports</h2>
    </div>

    <!-- 统计卡 -->
    <div style="display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-bottom: 16px;">
      <div style="padding: 12px 16px; background: var(--bg-elevated); border: 1px solid var(--border); border-radius: 8px;">
        <NText depth="3" style="font-size: 12px;">总测试</NText>
        <div style="font-size: 24px; font-weight: 600; font-family: var(--font-mono);">{{ stats.total }}</div>
      </div>
      <div style="padding: 12px 16px; background: rgba(82, 196, 26, 0.05); border: 1px solid var(--success); border-radius: 8px;">
        <NText depth="3" style="font-size: 12px;">✅ 通过</NText>
        <div style="font-size: 24px; font-weight: 600; font-family: var(--font-mono); color: var(--success);">{{ stats.completed }}</div>
      </div>
      <div style="padding: 12px 16px; background: rgba(255, 77, 79, 0.05); border: 1px solid var(--error); border-radius: 8px;">
        <NText depth="3" style="font-size: 12px;">❌ 失败</NText>
        <div style="font-size: 24px; font-weight: 600; font-family: var(--font-mono); color: var(--error);">{{ stats.failed }}</div>
      </div>
      <div style="padding: 12px 16px; background: rgba(19, 194, 194, 0.05); border: 1px solid var(--info); border-radius: 8px;">
        <NText depth="3" style="font-size: 12px;">🔄 运行中</NText>
        <div style="font-size: 24px; font-weight: 600; font-family: var(--font-mono); color: var(--info);">{{ stats.running }}</div>
      </div>
    </div>

    <!-- 筛选 -->
    <div style="display: flex; gap: 12px; margin-bottom: 16px;">
      <NInput
        v-model:value="searchQuery"
        placeholder="🔍 搜索场景名或 ID..."
        style="flex: 1;"
        clearable
      />
      <select
        v-model="statusFilter"
        style="
          padding: 8px 12px;
          background: var(--bg-elevated);
          border: 1px solid var(--border);
          border-radius: 6px;
          color: var(--text-primary);
        "
      >
        <option value="">所有状态</option>
        <option value="completed">✅ 通过</option>
        <option value="failed">❌ 失败</option>
        <option value="aborted">⏹ 停止</option>
        <option value="running">🔄 运行中</option>
        <option value="paused">⏸ 暂停</option>
      </select>
    </div>

    <!-- 加载 -->
    <div v-if="loading" style="padding: 40px; text-align: center; color: var(--text-secondary);">
      加载中...
    </div>

    <!-- 空状态 -->
    <NCard v-else-if="filteredRuns.length === 0" :bordered="false" embedded>
      <NEmpty
        v-if="runs.length === 0"
        description="还没有测试报告。运行一次压测后会显示在这里。"
      >
        <NButton type="primary" @click="router.push({ name: 'scenarios' })">
          📋 查看场景列表
        </NButton>
      </NEmpty>
      <NEmpty
        v-else
        :description="`没有匹配「${searchQuery || statusFilter}」的报告`"
      />
    </NCard>

    <!-- 列表 -->
    <NCard v-else :bordered="false" embedded>
      <table style="width: 100%; border-collapse: collapse;">
        <thead>
          <tr style="background: var(--bg-hover); border-bottom: 1px solid var(--border);">
            <th style="padding: 12px; text-align: left; font-weight: 600;">场景</th>
            <th style="padding: 12px; text-align: left; font-weight: 600;">状态</th>
            <th style="padding: 12px; text-align: right; font-weight: 600;">总请求</th>
            <th style="padding: 12px; text-align: right; font-weight: 600;">P95</th>
            <th style="padding: 12px; text-align: right; font-weight: 600;">错误率</th>
            <th style="padding: 12px; text-align: right; font-weight: 600;">耗时</th>
            <th style="padding: 12px; text-align: right; font-weight: 600;">时间</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="run in filteredRuns"
            :key="run.id"
            style="border-bottom: 1px solid var(--border); cursor: pointer; transition: background 0.15s;"
            @click="viewReport(run.id, run.status)"
            @mouseover="(e: any) => e.currentTarget.style.background = 'var(--bg-hover)'"
            @mouseleave="(e: any) => e.currentTarget.style.background = 'transparent'"
          >
            <td style="padding: 12px;">
              <div style="font-weight: 500;">{{ run.scenarioName }}</div>
              <div style="font-size: 11px; color: var(--text-tertiary);">#{{ run.id }}</div>
            </td>
            <td style="padding: 12px;">
              <NTag :type="statusColor(run.status) as any" size="small" round>
                {{ statusLabel(run.status) }}
              </NTag>
            </td>
            <td style="padding: 12px; text-align: right; font-family: var(--font-mono);">
              {{ formatNumber(run.summary?.TotalRequests || 0) }}
            </td>
            <td style="padding: 12px; text-align: right; font-family: var(--font-mono);">
              {{ Math.round(run.summary?.P95Ms || 0) }} ms
            </td>
            <td style="padding: 12px; text-align: right; font-family: var(--font-mono);"
                :style="{ color: (run.summary?.ErrorRate || 0) > 0.01 ? 'var(--error)' : 'var(--text-secondary)' }">
              {{ ((run.summary?.ErrorRate || 0) * 100).toFixed(2) }}%
            </td>
            <td style="padding: 12px; text-align: right; font-family: var(--font-mono);">
              {{ run.durationMs ? Math.round(run.durationMs / 1000) + 's' : '-' }}
            </td>
            <td style="padding: 12px; text-align: right; font-size: 12px; color: var(--text-tertiary);">
              {{ timeAgo(run.createdAt) }}
            </td>
          </tr>
        </tbody>
      </table>
    </NCard>
  </div>
</template>