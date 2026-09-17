<script setup lang="ts">
import { ref, computed, h, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useMessage, useDialog, NButton, NSpace, NTag, NIcon } from 'naive-ui'
import { useScenariosStore } from '@/stores/scenarios'
import { useProjectsStore } from '@/stores/projects'
import type { ScenarioDTO } from '@/api/scenarios'

const router = useRouter()
const { t } = useI18n()
const message = useMessage()
const dialog = useDialog()
const store = useScenariosStore()
const projectsStore = useProjectsStore()

const searchInput = ref('')
const selectedIds = ref<number[]>([])
const viewMode = ref<'list' | 'card'>('list')

// 防抖搜索
let searchTimer: number | null = null
watch(searchInput, (val) => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = window.setTimeout(() => {
    store.searchQuery = val
    store.fetchList()
  }, 300)
})

watch(() => store.projectIdFilter, () => {
  store.fetchList()
})

onMounted(async () => {
  await projectsStore.fetchList()
  await store.fetchList()
})

// ============================================
// 操作方法
// ============================================

function newScenario() {
  router.push({ name: 'scenario-editor' })
}

function editScenario(id: number) {
  router.push({ name: 'scenario-editor', params: { id: String(id) } })
}

async function runScenario(scenario: ScenarioDTO) {
  // 直接运行,使用场景的默认配置
  try {
    message.info(`正在启动 "${scenario.name}"...`)
    const resp = await api.startRun({ scenarioId: scenario.id })
    if (resp && resp.runId > 0) {
      message.success(`已启动:${resp.scenarioName}`)
      router.push({ name: 'live-monitor', params: { id: String(resp.runId) } })
    } else {
      message.error('启动失败')
    }
  } catch (e: any) {
    message.error(`启动失败: ${e.message}`)
  }
}

async function deleteScenario(scenario: ScenarioDTO) {
  dialog.warning({
    title: '确认删除',
    content: `确定删除场景 "${scenario.name}" 吗?此操作不可撤销。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      const ok = await store.remove(scenario.id)
      if (ok) {
        message.success('已删除')
      } else {
        message.error(store.error || '删除失败')
      }
    },
  })
}

async function duplicateScenario(scenario: ScenarioDTO) {
  const dto = await store.duplicate(scenario.id)
  if (dto) {
    message.success(`已复制为 "${dto.name}"`)
  } else {
    message.error(store.error || '复制失败')
  }
}

async function batchDelete() {
  if (selectedIds.value.length === 0) return
  dialog.warning({
    title: '批量删除',
    content: `确定删除选中的 ${selectedIds.value.length} 个场景吗?`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      const count = await store.removeMany([...selectedIds.value])
      selectedIds.value = []
      message.success(`已删除 ${count} 个场景`)
    },
  })
}

function toggleSelect(id: number) {
  const idx = selectedIds.value.indexOf(id)
  if (idx >= 0) {
    selectedIds.value.splice(idx, 1)
  } else {
    selectedIds.value.push(id)
  }
}

function isSelected(id: number) {
  return selectedIds.value.includes(id)
}

function statusColor(status: string): string {
  if (status === 'active') return 'success'
  if (status === 'draft') return 'default'
  if (status === 'archived') return 'warning'
  return 'default'
}

// ============================================
// 列表列定义
// ============================================

const columns = computed(() => [
  {
    title: '',
    key: 'select',
    width: 40,
    render(row: ScenarioDTO) {
      return h('input', {
        type: 'checkbox',
        checked: isSelected(row.id),
        onChange: () => toggleSelect(row.id),
      })
    },
  },
  {
    title: '名称',
    key: 'name',
    render(row: ScenarioDTO) {
      return h('div', { style: 'display: flex; align-items: center; gap: 8px;' }, [
        row.tags?.includes('pinned') && h('span', { style: 'color: #FAAD14;' }, '⭐'),
        h('span', { style: 'font-weight: 500;' }, row.name),
        h(NTag, { size: 'small', type: 'default', round: true }, () => `v${row.version}`),
      ])
    },
  },
  {
    title: '请求',
    key: 'requestCount',
    render: (row: ScenarioDTO) => `${row.requestCount} 个`,
  },
  { title: 'VUs', key: 'vus' },
  { title: '时长', key: 'duration' },
  {
    title: '来源',
    key: 'source',
    render: (row: ScenarioDTO) =>
      h(NTag, {
        size: 'small',
        type: row.source === 'manual' ? 'info' : 'warning',
        round: true,
      }, () => row.source),
  },
  {
    title: '状态',
    key: 'status',
    render: (row: ScenarioDTO) =>
      h(NTag, { size: 'small', type: statusColor(row.status) as any }, () => row.status),
  },
  { title: '更新时间', key: 'updatedAt' },
  {
    title: '操作',
    key: 'actions',
    width: 240,
    render(row: ScenarioDTO) {
      return h(NSpace, { size: 'small' }, () => [
        h(NButton, {
          size: 'tiny',
          type: 'primary',
          ghost: true,
          onClick: () => runScenario(row),
        }, () => '▶ Run'),
        h(NButton, {
          size: 'tiny',
          onClick: () => editScenario(row.id),
        }, () => '✎ Edit'),
        h(NButton, {
          size: 'tiny',
          onClick: () => duplicateScenario(row),
        }, () => '⎘ Copy'),
        h(NButton, {
          size: 'tiny',
          type: 'error',
          ghost: true,
          onClick: () => deleteScenario(row),
        }, () => '🗑'),
      ])
    },
  },
])
</script>

<template>
  <div>
    <!-- 顶部 -->
    <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px;">
      <h2 style="margin: 0; font-size: 20px; font-weight: 600;">Scenarios</h2>
      <NButton type="primary" @click="newScenario">
        ➕ {{ t('scenario.newScenario') }}
      </NButton>
    </div>

    <!-- 筛选栏 -->
    <div style="display: flex; gap: 12px; margin-bottom: 16px;">
      <input
        v-model="searchInput"
        type="text"
        placeholder="🔍 搜索场景名称..."
        style="
          flex: 1;
          padding: 8px 12px;
          background: var(--bg-elevated);
          border: 1px solid var(--border);
          border-radius: 6px;
          color: var(--text-primary);
          font-size: 14px;
          outline: none;
        "
      />

      <select
        :value="store.projectIdFilter ?? ''"
        @change="(e: any) => store.projectIdFilter = e.target.value ? Number(e.target.value) : null"
        style="
          padding: 8px 12px;
          background: var(--bg-elevated);
          border: 1px solid var(--border);
          border-radius: 6px;
          color: var(--text-primary);
        "
      >
        <option value="">所有项目</option>
        <option v-for="p in projectsStore.items" :key="p.id" :value="p.id">
          {{ p.name }} ({{ p.scenarioCount }})
        </option>
      </select>

      <div style="display: flex; gap: 4px;">
        <button
          @click="viewMode = 'list'"
          :style="{
            padding: '8px 12px',
              background: viewMode === 'list' ? 'var(--primary)' : 'var(--bg-elevated)',
              border: '1px solid var(--border)',
              borderRadius: '6px',
              color: viewMode === 'list' ? '#fff' : 'var(--text-primary)',
              cursor: 'pointer',
            }"
        >≡ 列表</button>
        <button
          @click="viewMode = 'card'"
          :style="{
            padding: '8px 12px',
              background: viewMode === 'card' ? 'var(--primary)' : 'var(--bg-elevated)',
              border: '1px solid var(--border)',
              borderRadius: '6px',
              color: viewMode === 'card' ? '#fff' : 'var(--text-primary)',
              cursor: 'pointer',
            }"
        >⊞ 卡片</button>
      </div>
    </div>

    <!-- 批量操作栏 -->
    <div
      v-if="selectedIds.length > 0"
      style="
        padding: 8px 16px;
        background: var(--primary-bg);
        border: 1px solid var(--primary);
        border-radius: 6px;
        margin-bottom: 16px;
        display: flex;
        align-items: center;
        gap: 12px;
      "
    >
      <span style="color: var(--primary); font-weight: 500;">
        已选中 {{ selectedIds.length }} 个场景
      </span>
      <NButton size="small" type="error" @click="batchDelete">批量删除</NButton>
      <NButton size="small" @click="selectedIds = []">取消选择</NButton>
    </div>

    <!-- 加载状态 -->
    <div
      v-if="store.loading"
      style="padding: 40px; text-align: center; color: var(--text-secondary);"
    >
      加载中...
    </div>

    <!-- 错误状态 -->
    <div
      v-else-if="store.error"
      style="
        padding: 16px;
        background: rgba(255, 77, 79, 0.1);
        border: 1px solid var(--error);
        border-radius: 6px;
        color: var(--error);
        margin-bottom: 16px;
      "
    >
      错误:{{ store.error }}
      <NButton size="tiny" @click="store.clearError()">关闭</NButton>
    </div>

    <!-- 空状态 -->
    <div
      v-else-if="!store.hasItems && !searchInput"
      style="
        padding: 60px 20px;
        text-align: center;
        background: var(--bg-elevated);
        border: 1px dashed var(--border);
        border-radius: 8px;
      "
    >
      <div style="font-size: 48px; margin-bottom: 16px;">📋</div>
      <h3 style="margin: 0 0 8px 0;">还没有场景</h3>
      <p style="color: var(--text-secondary); margin: 0 0 16px 0;">
        点击下方按钮创建你的第一个压测场景
      </p>
      <NButton type="primary" size="large" @click="newScenario">
        ➕ 创建第一个场景
      </NButton>
    </div>

    <!-- 无搜索结果 -->
    <div
      v-else-if="!store.hasItems && searchInput"
      style="padding: 40px; text-align: center; color: var(--text-secondary);"
    >
      没有匹配 "{{ searchInput }}" 的场景
    </div>

    <!-- 列表视图 -->
    <div
      v-else-if="viewMode === 'list'"
      style="
        background: var(--bg-elevated);
        border: 1px solid var(--border);
        border-radius: 8px;
        overflow: hidden;
      "
    >
      <table style="width: 100%; border-collapse: collapse;">
        <thead>
          <tr style="background: var(--bg-hover); border-bottom: 1px solid var(--border);">
            <th style="padding: 12px; width: 40px;"></th>
            <th style="padding: 12px; text-align: left; font-weight: 600;">名称</th>
            <th style="padding: 12px; text-align: left;">请求数</th>
            <th style="padding: 12px; text-align: left;">VUs</th>
            <th style="padding: 12px; text-align: left;">时长</th>
            <th style="padding: 12px; text-align: left;">来源</th>
            <th style="padding: 12px; text-align: left;">状态</th>
            <th style="padding: 12px; text-align: left;">更新时间</th>
            <th style="padding: 12px; text-align: left;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="scenario in store.items"
            :key="scenario.id"
            style="border-bottom: 1px solid var(--border);"
          >
            <td style="padding: 12px;">
              <input
                type="checkbox"
                :checked="isSelected(scenario.id)"
                @change="toggleSelect(scenario.id)"
              />
            </td>
            <td style="padding: 12px;">
              <div style="display: flex; align-items: center; gap: 8px;">
                <span v-if="scenario.tags?.includes('pinned')" style="color: #FAAD14;">⭐</span>
                <span
                  style="font-weight: 500; cursor: pointer; color: var(--primary);"
                  @click="editScenario(scenario.id)"
                >{{ scenario.name }}</span>
                <span
                  style="
                    padding: 2px 6px;
                    background: var(--bg-hover);
                    border-radius: 4px;
                    font-size: 11px;
                    color: var(--text-secondary);
                  "
                >v{{ scenario.version }}</span>
              </div>
            </td>
            <td style="padding: 12px;">{{ scenario.requestCount }} 个</td>
            <td style="padding: 12px;">{{ scenario.vus }}</td>
            <td style="padding: 12px;">{{ scenario.duration }}</td>
            <td style="padding: 12px;">
              <span
                :style="{
                  padding: '2px 8px',
                  borderRadius: '4px',
                  fontSize: '12px',
                  background: scenario.source === 'manual' ? 'rgba(19, 194, 194, 0.1)' : 'rgba(250, 173, 20, 0.1)',
                  color: scenario.source === 'manual' ? '#13C2C2' : '#FAAD14',
                }"
              >{{ scenario.source }}</span>
            </td>
            <td style="padding: 12px;">
              <span
                :style="{
                  padding: '2px 8px',
                  borderRadius: '4px',
                  fontSize: '12px',
                  background: scenario.status === 'active' ? 'rgba(82, 196, 26, 0.1)' : 'var(--bg-hover)',
                  color: scenario.status === 'active' ? '#52C41A' : 'var(--text-secondary)',
                }"
              >{{ scenario.status }}</span>
            </td>
            <td style="padding: 12px; color: var(--text-secondary); font-size: 13px;">
              {{ scenario.updatedAt }}
            </td>
            <td style="padding: 12px;">
              <div style="display: flex; gap: 4px;">
                <button
                  @click="runScenario(scenario)"
                  style="
                    padding: 4px 10px;
                    background: var(--primary);
                    color: #fff;
                    border: none;
                    border-radius: 4px;
                    cursor: pointer;
                    font-size: 12px;
                  "
                >▶ Run</button>
                <button
                  @click="editScenario(scenario.id)"
                  style="
                    padding: 4px 10px;
                    background: var(--bg-hover);
                    color: var(--text-primary);
                    border: none;
                    border-radius: 4px;
                    cursor: pointer;
                    font-size: 12px;
                  "
                >✎ Edit</button>
                <button
                  @click="duplicateScenario(scenario)"
                  style="
                    padding: 4px 10px;
                    background: var(--bg-hover);
                    color: var(--text-primary);
                    border: none;
                    border-radius: 4px;
                    cursor: pointer;
                    font-size: 12px;
                  "
                >⎘ Copy</button>
                <button
                  @click="deleteScenario(scenario)"
                  style="
                    padding: 4px 10px;
                    background: transparent;
                    color: var(--error);
                    border: 1px solid var(--error);
                    border-radius: 4px;
                    cursor: pointer;
                    font-size: 12px;
                  "
                >🗑</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 卡片视图 -->
    <div
      v-else
      style="
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
        gap: 16px;
      "
    >
      <div
        v-for="scenario in store.items"
        :key="scenario.id"
        style="
          background: var(--bg-elevated);
          border: 1px solid var(--border);
          border-radius: 8px;
          padding: 16px;
          transition: all 0.2s;
          cursor: pointer;
        "
        @click="editScenario(scenario.id)"
        @mouseover="(e: any) => e.currentTarget.style.borderColor = 'var(--primary)'"
        @mouseleave="(e: any) => e.currentTarget.style.borderColor = 'var(--border)'"
      >
        <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px;">
          <div style="display: flex; align-items: center; gap: 8px;">
            <span v-if="scenario.tags?.includes('pinned')" style="color: #FAAD14;">⭐</span>
            <h3 style="margin: 0; font-size: 16px; font-weight: 600;">{{ scenario.name }}</h3>
          </div>
          <span
            style="
              padding: 2px 6px;
              background: var(--bg-hover);
              border-radius: 4px;
              font-size: 11px;
              color: var(--text-secondary);
            "
          >v{{ scenario.version }}</span>
        </div>

        <p
          v-if="scenario.description"
          style="
            margin: 0 0 12px 0;
            color: var(--text-secondary);
            font-size: 13px;
            line-height: 1.5;
            min-height: 20px;
          "
        >{{ scenario.description }}</p>

        <div style="display: flex; gap: 16px; margin-bottom: 12px; font-size: 13px; color: var(--text-secondary);">
          <span>📋 {{ scenario.requestCount }} 请求</span>
          <span>👥 {{ scenario.vus }} VUs</span>
          <span>⏱ {{ scenario.duration }}</span>
        </div>

        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span
            :style="{
              padding: '2px 8px',
              borderRadius: '4px',
              fontSize: '11px',
              background: scenario.status === 'active' ? 'rgba(82, 196, 26, 0.1)' : 'var(--bg-hover)',
              color: scenario.status === 'active' ? '#52C41A' : 'var(--text-secondary)',
            }"
          >{{ scenario.status }}</span>
          <span style="font-size: 12px; color: var(--text-tertiary);">
            {{ scenario.updatedAt }}
          </span>
        </div>
      </div>
    </div>

    <!-- 底部统计 -->
    <div
      v-if="store.hasItems"
      style="
        margin-top: 16px;
        padding: 8px 16px;
        color: var(--text-secondary);
        font-size: 13px;
        text-align: center;
      "
    >
      共 {{ store.total }} 个场景 · 显示 {{ store.items.length }} 个
    </div>
  </div>
</template>

<style scoped>
input[type="::checkbox"] {
  cursor: pointer;
  accent-color: var(--primary);
}

button:hover {
  filter: brightness(1.1);
}
</style>