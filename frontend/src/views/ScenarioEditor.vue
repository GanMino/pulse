<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage, NModal } from 'naive-ui'
import { useScenariosStore } from '@/stores/scenarios'
import { useProjectsStore } from '@/stores/projects'
import * as api from '@/api/scenarios'
import {
  createEmptyRequest,
  createEmptyVariable,
  createEmptyEditorData,
  fromScenarioConfig,
  toScenarioConfig,
  type ScenarioEditorData,
  type RequestEditor,
  type VariableEntry,
} from '@/types/scenario'
import RequestEditor from '@/components/RequestEditor.vue'
import KeyValueEditor from '@/components/KeyValueEditor.vue'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const store = useScenariosStore()
const projectsStore = useProjectsStore()

const scenarioId = computed(() => {
  const id = route.params.id
  return id ? Number(id) : 0
})

const isEdit = computed(() => scenarioId.value > 0)
const isDirty = ref(false)
const saving = ref(false)
const loaded = ref(false)

// 编辑器数据
const data = ref<ScenarioEditorData>(createEmptyEditorData())

// 当前选中的请求索引
const activeRequestIdx = ref(0)

// HAR 导入对话框
const harImportDialog = ref(false)
const harFileData = ref<string>('')
const harFileName = ref('')
const harBaseURL = ref('')
const harImporting = ref(false)
const harImportResult = ref<api.ImportHARResponse | null>(null)

// ============================================
// 加载
// ============================================

onMounted(async () => {
  await projectsStore.fetchList()

  if (isEdit.value) {
    // 编辑模式:从后端加载
    try {
      const dto = await api.getScenario(scenarioId.value)
      if (dto && dto.id > 0) {
        data.value = fromScenarioConfig(dto.id, (dto as any).config, dto.name)
        data.value.name = dto.name
        data.value.description = dto.description || ''
      } else {
        message.error('加载场景失败')
        router.push({ name: 'scenarios' })
        return
      }
    } catch (e: any) {
      message.error(`加载失败: ${e.message}`)
      router.push({ name: 'scenarios' })
      return
    }
  } else {
    // 新建模式:确保默认项目存在
    await projectsStore.ensureDefault()
  }

  setTimeout(() => { loaded.value = true }, 100)
})

// 标记为已修改
watch(data, () => {
  if (loaded.value) isDirty.value = true
}, { deep: true })

// ============================================
// 多请求操作
// ============================================

function addRequest() {
  data.value.requests.push(createEmptyRequest(`Request ${data.value.requests.length + 1}`))
  activeRequestIdx.value = data.value.requests.length - 1
  isDirty.value = true
}

function removeRequest(idx: number) {
  if (data.value.requests.length <= 1) {
    message.warning('至少需要保留一个请求')
    return
  }
  if (!confirm(`确定删除请求 "${data.value.requests[idx].name}"?`)) return
  data.value.requests.splice(idx, 1)
  if (activeRequestIdx.value >= data.value.requests.length) {
    activeRequestIdx.value = data.value.requests.length - 1
  }
  isDirty.value = true
}

function duplicateRequest(idx: number) {
  const src = data.value.requests[idx]
  const dup: RequestEditor = JSON.parse(JSON.stringify(src))
  dup.uid = `req_${Date.now()}_${Math.random().toString(36).slice(2, 7)}`
  dup.name = src.name + ' (副本)'
  data.value.requests.splice(idx + 1, 0, dup)
  activeRequestIdx.value = idx + 1
  isDirty.value = true
}

function moveRequest(idx: number, direction: -1 | 1) {
  const newIdx = idx + direction
  if (newIdx < 0 || newIdx >= data.value.requests.length) return
  const arr = data.value.requests
  ;[arr[idx], arr[newIdx]] = [arr[newIdx], arr[idx]]
  if (activeRequestIdx.value === idx) activeRequestIdx.value = newIdx
  else if (activeRequestIdx.value === newIdx) activeRequestIdx.value = idx
  isDirty.value = true
}

function selectRequest(idx: number) {
  activeRequestIdx.value = idx
}

// ============================================
// 变量操作
// ============================================

function addVariable() {
  data.value.variables.push(createEmptyVariable())
  isDirty.value = true
}

function removeVariable(uid: string) {
  data.value.variables = data.value.variables.filter(v => v.uid !== uid)
  isDirty.value = true
}

function updateVariables(vars: any[]) {
  data.value.variables = vars
  isDirty.value = true
}

// ============================================
// 保存 / 运行
// ============================================

async function save(opts: { skipRun?: boolean } = {}) {
  if (saving.value) return
  saving.value = true

  try {
    const cfg = toScenarioConfig(data.value)

    const dto = await store.save({
      id: scenarioId.value || undefined,
      name: data.value.name,
      description: data.value.description,
      config: cfg,
      tags: data.value.tags,
    })

    if (dto) {
      isDirty.value = false
      message.success(`保存成功 (v${dto.version})`)
      if (!isEdit.value) {
        router.replace({ name: 'scenario-editor', params: { id: String(dto.id) } })
      }
    } else {
      message.error(store.error || '保存失败')
    }
  } catch (e: any) {
    message.error(e?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function run() {
  // 先保存
  await save({ skipRun: true })

  // 然后启动运行
  try {
    const resp = await api.startRun({ scenarioId: scenarioId.value })
    if (resp && resp.runId > 0) {
      message.success(`已启动: ${resp.scenarioName}`)
      router.push({ name: 'live-monitor', params: { id: String(resp.runId) } })
    } else {
      message.error('启动失败')
    }
  } catch (e: any) {
    message.error(`启动失败: ${e.message}`)
  }
}

function back() {
  if (isDirty.value) {
    if (!confirm('有未保存的修改,确定离开吗?')) return
  }
  router.push({ name: 'scenarios' })
}

// ============================================
// 快捷键
// ============================================

function handleKeydown(e: KeyboardEvent) {
  // 跳过在输入框中的按键
  const target = e.target as HTMLElement
  const isInInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable

  const meta = e.metaKey || e.ctrlKey

  if (meta && e.key === 's' && !e.shiftKey) {
    e.preventDefault()
    save()
  } else if (meta && e.shiftKey && (e.key === 'R' || e.key === 'r')) {
    e.preventDefault()
    run()
  } else if (meta && e.key === 'Enter') {
    e.preventDefault()
    save()
  } else if (!isInInput && meta && e.key === 'i' && !e.shiftKey) {
    e.preventDefault()
    openHarImportDialog()
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})

// ============================================
// HAR 导入
// ============================================

function openHarImportDialog() {
  harImportDialog.value = true
  harFileData.value = ''
  harFileName.value = ''
  harBaseURL.value = ''
  harImportResult.value = null
}

function closeHarImportDialog() {
  harImportDialog.value = false
}

async function onHarFileSelect(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  harFileName.value = file.name
  harBaseURL.value = ''

  // 读取为 base64(传给后端)
  const reader = new FileReader()
  reader.onload = () => {
    const result = reader.result as string
    // 去掉 "data:application/json;base64," 前缀
    const base64 = result.split(',')[1]
    harFileData.value = base64
  }
  reader.readAsDataURL(file)
}

async function importHar() {
  if (!harFileData.value) {
    message.warning('请先选择 HAR 文件')
    return
  }

  harImporting.value = true
  try {
    // base64 → 字节
    const binary = atob(harFileData.value)
    const bytes = new Uint8Array(binary.length)
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i)
    }

    const req: api.ImportHARRequest = {
      name: `导入自 ${harFileName.value}`,
      harData: Array.from(bytes),
      baseUrl: harBaseURL.value || undefined,
      vus: data.value.load.vus,
      duration: data.value.load.duration,
    }

    const result = await api.importHar(req)
    harImportResult.value = result

    // 应用到编辑器
    if (result.scenario) {
      const newData = fromScenarioConfig(0, (result.scenario as any).config, result.scenario.name)
      newData.requests = newData.requests
      data.value = newData
      activeRequestIdx.value = 0
      isDirty.value = true
      message.success(`已导入 ${result.stats.importedRequests} 个请求(去重 ${result.stats.duplicatesSkipped})`)
    }
  } catch (e: any) {
    message.error(`导入失败: ${e.message}`)
  } finally {
    harImporting.value = false
  }
}

const activeRequest = computed(() => data.value.requests[activeRequestIdx.value])
</script>

<template>
  <div class="editor-container">
    <!-- 顶部 -->
    <div class="editor-topbar">
      <div class="editor-topbar-left">
        <button @click="back" class="btn-secondary">← 返回</button>
        <input
          v-model="data.name"
          placeholder="场景名称"
          class="editor-name-input"
        />
        <span class="editor-badge">
          {{ isEdit ? `v${store.items.find(s => s.id === scenarioId)?.version || '?'}` : '新建' }}
        </span>
        <span v-if="isDirty" class="editor-dirty">● 未保存</span>
      </div>

      <div class="editor-topbar-right">
        <button @click="openHarImportDialog" class="btn-secondary">📥 导入 HAR</button>
        <button @click="save" :disabled="saving" class="btn-secondary">💾 保存 (⌘S)</button>
        <button @click="run" :disabled="saving" class="btn-primary">▶ 运行 (⌘⇧R)</button>
      </div>
    </div>

    <!-- 三栏布局 -->
    <div class="editor-grid">
      <!-- 左:请求列表 -->
      <div class="editor-sidebar">
        <div class="sidebar-header">
          <span>REQUESTS</span>
          <button @click="addRequest" class="sidebar-add-btn">+ 添加</button>
        </div>
        <div class="request-list">
          <div
            v-for="(req, idx) in data.requests"
            :key="req.uid"
            :class="['request-item', { active: idx === activeRequestIdx }]"
            @click="selectRequest(idx)"
          >
            <div class="request-item-main">
              <div class="request-item-name">{{ req.name }}</div>
              <div class="request-item-meta">
                <span :class="['method-badge', `method-${req.method.toLowerCase()}`]">{{ req.method }}</span>
                <span class="request-item-url">{{ req.url }}</span>
              </div>
            </div>
            <div class="request-item-actions" @click.stop>
              <button @click="moveRequest(idx, -1)" :disabled="idx === 0" title="上移">↑</button>
              <button @click="moveRequest(idx, 1)" :disabled="idx === data.requests.length - 1" title="下移">↓</button>
              <button @click="duplicateRequest(idx)" title="复制">⎘</button>
              <button @click="removeRequest(idx)" title="删除" class="danger">🗑</button>
            </div>
          </div>
        </div>

        <!-- 变量定义面板 -->
        <div class="sidebar-section">
          <div class="sidebar-header">
            <span>VARIABLES</span>
            <button @click="addVariable" class="sidebar-add-btn">+ 添加</button>
          </div>
          <div v-if="data.variables.length === 0" class="sidebar-empty">
            暂无变量。在请求中使用 <code>{{ '{{name}}' }}</code> 占位符。
          </div>
          <KeyValueEditor
            v-else
            :model-value="data.variables as any"
            key-placeholder="变量名"
            value-placeholder="值(多个值用逗号分隔)"
            key-label="Name"
            value-label="Values"
            key-width="30%"
            add-label="添加变量"
            @update:model-value="updateVariables"
          />
        </div>
      </div>

      <!-- 中:请求详情 -->
      <div class="editor-main">
        <RequestEditor
          v-if="activeRequest"
          v-model="data.requests[activeRequestIdx]"
        />
      </div>

      <!-- 右:压测配置 + 阈值 -->
      <div class="editor-config">
        <!-- 压测配置 -->
        <div class="config-section">
          <div class="config-section-title">压测配置</div>

          <div class="config-row">
            <label class="config-label">VUs</label>
            <div class="config-input-group">
              <input
                v-model.number="data.load.vus"
                type="number"
                min="1"
                max="100000"
                class="config-input"
              />
              <input
                v-model.number="data.load.vus"
                type="range"
                min="1"
                max="1000"
                class="config-slider"
              />
            </div>
          </div>

          <div class="config-row">
            <label class="config-label">Duration</label>
            <select v-model="data.load.duration" class="config-select">
              <option value="30s">30 秒</option>
              <option value="1m">1 分钟</option>
              <option value="5m">5 分钟</option>
              <option value="10m">10 分钟</option>
              <option value="30m">30 分钟</option>
              <option value="1h">1 小时</option>
            </select>
          </div>

          <div class="config-row">
            <label class="config-label">Ramp-up</label>
            <select v-model="data.load.rampUpType" class="config-select">
              <option value="linear">Linear(线性)</option>
              <option value="step">Step(阶梯)</option>
              <option value="wave">Wave(波浪)</option>
              <option value="none">None(无)</option>
            </select>
          </div>

          <div v-if="data.load.rampUpType !== 'none'" class="config-row">
            <label class="config-label">Ramp 时长</label>
            <select v-model="data.load.rampUpDuration" class="config-select">
              <option value="10s">10 秒</option>
              <option value="30s">30 秒</option>
              <option value="1m">1 分钟</option>
              <option value="2m">2 分钟</option>
              <option value="5m">5 分钟</option>
            </select>
          </div>

          <div class="config-row">
            <label class="config-label">Target RPS</label>
            <input
              v-model.number="data.load.targetRPS"
              type="number"
              min="0"
              placeholder="不限"
              class="config-input"
            />
          </div>

          <div class="config-row">
            <label class="config-label">Think Time (ms)</label>
            <input
              v-model.number="data.load.thinkTimeMs"
              type="number"
              min="0"
              step="100"
              placeholder="0"
              class="config-input"
            />
          </div>
        </div>

        <!-- 阈值 -->
        <div class="config-section">
          <div class="config-section-title">阈值(达到即视为失败)</div>

          <div class="config-row">
            <label class="config-label">P95 延迟 (ms)</label>
            <input
              v-model.number="data.thresholds.p95Ms"
              type="number"
              min="0"
              placeholder="0 = 不检查"
              class="config-input"
            />
          </div>

          <div class="config-row">
            <label class="config-label">P99 延迟 (ms)</label>
            <input
              v-model.number="data.thresholds.p99Ms"
              type="number"
              min="0"
              placeholder="0 = 不检查"
              class="config-input"
            />
          </div>

          <div class="config-row">
            <label class="config-label">错误率上限</label>
            <input
              v-model.number="data.thresholds.errorRate"
              type="number"
              min="0"
              max="1"
              step="0.01"
              placeholder="0.01"
              class="config-input"
            />
          </div>

          <div class="config-row">
            <label class="config-label">最小 RPS</label>
            <input
              v-model.number="data.thresholds.minRPS"
              type="number"
              min="0"
              placeholder="0 = 不检查"
              class="config-input"
            />
          </div>
        </div>

        <!-- 描述 -->
        <div class="config-section">
          <div class="config-section-title">描述</div>
          <textarea
            v-model="data.description"
            rows="3"
            class="config-textarea"
            placeholder="场景说明..."
          />
        </div>

        <!-- 快捷键提示 -->
        <div class="config-shortcuts">
          <div class="shortcuts-title">快捷键</div>
          <div class="shortcut-row"><kbd>⌘S</kbd> 保存</div>
          <div class="shortcut-row"><kbd>⌘⇧R</kbd> 保存并运行</div>
          <div class="shortcut-row"><kbd>⌘I</kbd> 导入 HAR</div>
        </div>
      </div>
    </div>

    <!-- HAR 导入对话框 -->
    <NModal
      v-model:show="harImportDialog"
      preset="card"
      title="📥 导入 HAR 录制"
      style="max-width: 600px;"
      :bordered="false"
    >
      <div style="display: flex; flex-direction: column; gap: 16px;">
        <div>
          <label class="har-label">HAR 文件</label>
          <input
            type="file"
            accept=".har,application/json,text/json"
            @change="onHarFileSelect"
            class="har-file-input"
          />
          <div v-if="harFileName" class="har-file-name">已选择: {{ harFileName }}</div>
        </div>

        <div>
          <label class="har-label">Base URL(可选)</label>
          <input
            v-model="harBaseURL"
            placeholder="如 https://api.example.com,留空则使用原始 URL"
            class="har-input"
          />
          <div class="har-hint">
            指定后,所有以该 URL 开头的请求会被替换为相对路径(如 /api/users)
          </div>
        </div>

        <div v-if="harImportResult" class="har-stats">
          ✅ 导入完成:
          <ul style="margin: 8px 0 0 20px;">
            <li>总条目数: {{ harImportResult.stats.totalEntries }}</li>
            <li>导入请求: {{ harImportResult.stats.importedRequests }}</li>
            <li>去重跳过: {{ harImportResult.stats.duplicatesSkipped }}</li>
          </ul>
        </div>
      </div>

      <template #footer>
        <div style="display: flex; gap: 8px; justify-content: flex-end;">
          <button @click="closeHarImportDialog" class="btn-secondary">关闭</button>
          <button @click="importHar" :disabled="!harFileData || harImporting" class="btn-primary">
            {{ harImporting ? '导入中...' : '导入' }}
          </button>
        </div>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.editor-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.editor-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 16px;
  margin-bottom: 16px;
  border-bottom: 1px solid var(--border);
}

.editor-topbar-left,
.editor-topbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.editor-name-input {
  background: transparent;
  border: 1px solid transparent;
  color: var(--text-primary);
  font-size: 18px;
  font-weight: 600;
  padding: 6px 10px;
  border-radius: 6px;
  outline: none;
  width: 300px;
  transition: all 0.2s;
}

.editor-name-input:hover,
.editor-name-input:focus {
  border-color: var(--primary);
  background: var(--bg-elevated);
}

.editor-badge {
  padding: 2px 8px;
  background: var(--bg-hover);
  border-radius: 4px;
  font-size: 12px;
  color: var(--text-secondary);
  font-family: var(--font-mono);
}

.editor-dirty {
  padding: 2px 8px;
  background: rgba(250, 173, 20, 0.15);
  color: #FAAD14;
  border-radius: 4px;
  font-size: 12px;
}

.btn-primary,
.btn-secondary {
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.15s;
}

.btn-primary {
  background: var(--primary);
  color: #fff;
}

.btn-primary:hover:not(:disabled) {
  background: var(--primary-hover);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: var(--bg-elevated);
  color: var(--text-primary);
  border-color: var(--border);
}

.btn-secondary:hover:not(:disabled) {
  background: var(--bg-hover);
}

.editor-grid {
  display: grid;
  grid-template-columns: 240px 1fr 320px;
  gap: 16px;
  flex: 1;
  min-height: 0;
}

.editor-sidebar {
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow: auto;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  padding: 0 4px;
}

.sidebar-add-btn {
  padding: 4px 8px;
  background: transparent;
  color: var(--primary);
  border: 1px dashed var(--primary);
  border-radius: 4px;
  cursor: pointer;
  font-size: 11px;
}

.sidebar-add-btn:hover {
  background: var(--primary-bg);
}

.sidebar-section {
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sidebar-empty {
  padding: 12px;
  text-align: center;
  color: var(--text-tertiary);
  font-size: 12px;
  background: var(--bg-base);
  border-radius: 4px;
}

.sidebar-empty code {
  background: var(--bg-hover);
  padding: 1px 4px;
  border-radius: 3px;
  color: var(--primary);
  font-family: var(--font-mono);
}

.request-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.request-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 10px;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
}

.request-item:hover {
  border-color: var(--primary);
}

.request-item.active {
  background: var(--primary-bg);
  border-color: var(--primary);
}

.request-item-main {
  flex: 1;
  min-width: 0;
}

.request-item-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.request-item-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--text-secondary);
}

.request-item-url {
  font-family: var(--font-mono);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.method-badge {
  padding: 1px 6px;
  border-radius: 3px;
  font-size: 10px;
  font-weight: 600;
  font-family: var(--font-mono);
}

.method-get { background: rgba(82, 196, 26, 0.15); color: #52C41A; }
.method-post { background: rgba(51, 112, 255, 0.15); color: #3370FF; }
.method-put { background: rgba(250, 173, 20, 0.15); color: #FAAD14; }
.method-delete { background: rgba(255, 77, 79, 0.15); color: #FF4D4F; }
.method-patch { background: rgba(19, 194, 194, 0.15); color: #13C2C2; }
.method-head, .method-options { background: var(--bg-hover); color: var(--text-secondary); }

.request-item-actions {
  display: none;
  gap: 2px;
}

.request-item:hover .request-item-actions,
.request-item.active .request-item-actions {
  display: flex;
}

.request-item-actions button {
  width: 22px;
  height: 22px;
  background: transparent;
  color: var(--text-tertiary);
  border: none;
  border-radius: 3px;
  cursor: pointer;
  font-size: 11px;
}

.request-item-actions button:hover:not(:disabled) {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.request-item-actions button.danger:hover {
  background: rgba(255, 77, 79, 0.1);
  color: var(--error);
}

.request-item-actions button:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.editor-main {
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 24px;
  overflow: auto;
}

.editor-config {
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow: auto;
}

.config-section {
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.config-section-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  padding-bottom: 6px;
  border-bottom: 1px solid var(--border);
}

.config-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.config-label {
  width: 100px;
  font-size: 12px;
  color: var(--text-secondary);
}

.config-input,
.config-select {
  flex: 1;
  padding: 6px 10px;
  background: var(--bg-base);
  border: 1px solid var(--border);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
  min-width: 0;
}

.config-input:focus,
.config-select:focus {
  border-color: var(--primary);
}

.config-input-group {
  flex: 1;
  display: flex;
  gap: 4px;
  align-items: center;
}

.config-slider {
  flex: 1;
  accent-color: var(--primary);
}

.config-textarea {
  width: 100%;
  padding: 6px 10px;
  background: var(--bg-base);
  border: 1px solid var(--border);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
  resize: vertical;
  font-family: inherit;
}

.config-shortcuts {
  padding: 12px;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 8px;
}

.shortcuts-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.shortcut-row {
  font-size: 12px;
  color: var(--text-secondary);
  padding: 2px 0;
}

.shortcut-row kbd {
  display: inline-block;
  padding: 1px 6px;
  background: var(--bg-base);
  border: 1px solid var(--border);
  border-radius: 3px;
  font-family: var(--font-mono);
  font-size: 11px;
  margin-right: 8px;
}

.har-label {
  display: block;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.har-file-input {
  display: block;
  width: 100%;
  padding: 8px;
  background: var(--bg-base);
  border: 1px solid var(--border);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 13px;
  cursor: pointer;
}

.har-file-name {
  margin-top: 6px;
  font-size: 12px;
  color: var(--text-secondary);
}

.har-input {
  width: 100%;
  padding: 6px 10px;
  background: var(--bg-base);
  border: 1px solid var(--border);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
}

.har-hint {
  margin-top: 4px;
  font-size: 11px;
  color: var(--text-tertiary);
}

.har-stats {
  padding: 12px;
  background: rgba(82, 196, 26, 0.1);
  border: 1px solid var(--success);
  border-radius: 6px;
  font-size: 13px;
  color: var(--success);
}

.har-stats ul {
  margin: 8px 0 0 20px;
  color: var(--text-primary);
}
</style>