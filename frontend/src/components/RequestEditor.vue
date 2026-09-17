<script setup lang="ts">
// 单个请求的详细编辑器
import { computed, ref, watch } from 'vue'
import type { RequestEditor } from '@/types/scenario'
import KeyValueEditor from './KeyValueEditor.vue'

const props = defineProps<{
  modelValue: RequestEditor
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: RequestEditor): void
}>()

const methods = [
  { label: 'GET', value: 'GET' },
  { label: 'POST', value: 'POST' },
  { label: 'PUT', value: 'PUT' },
  { label: 'DELETE', value: 'DELETE' },
  { label: 'PATCH', value: 'PATCH' },
  { label: 'HEAD', value: 'HEAD' },
  { label: 'OPTIONS', value: 'OPTIONS' },
]

const bodyTypes = [
  { label: '无', value: 'none' },
  { label: 'JSON', value: 'json' },
  { label: 'Form', value: 'form' },
  { label: 'Raw', value: 'raw' },
]

// 局部编辑状态
const request = ref<RequestEditor>(JSON.parse(JSON.stringify(props.modelValue)))

watch(() => props.modelValue, (val) => {
  request.value = JSON.parse(JSON.stringify(val))
}, { deep: true })

function update() {
  emit('update:modelValue', JSON.parse(JSON.stringify(request.value)))
}

function onHeadersUpdate(headers: any[]) {
  request.value.headers = headers
  update()
}

function onExtractorsUpdate(extractors: any[]) {
  request.value.extractors = extractors
  update()
}

// JSON 校验状态
const jsonValid = computed(() => {
  if (request.value.bodyType !== 'json') return true
  if (!request.value.bodyJson.trim()) return true
  try {
    JSON.parse(request.value.bodyJson)
    return true
  } catch {
    return false
  }
})

const jsonError = computed(() => {
  if (jsonValid.value) return ''
  try {
    JSON.parse(request.value.bodyJson)
  } catch (e: any) {
    return e?.message || 'JSON 格式错误'
  }
  return ''
})

// 添加 status code
const newStatus = ref(200)

function addStatus() {
  const code = Number(newStatus.value)
  if (!isNaN(code) && code >= 100 && code < 600 && !request.value.assertions.status.includes(code)) {
    request.value.assertions.status.push(code)
    update()
  }
}

function removeStatus(idx: number) {
  request.value.assertions.status.splice(idx, 1)
  update()
}

// Body contains
function addBodyContains() {
  request.value.assertions.bodyContains.push('')
  update()
}

function removeBodyContains(idx: number) {
  request.value.assertions.bodyContains.splice(idx, 1)
  update()
}
</script>

<template>
  <div class="req-editor">
    <!-- Method + URL -->
    <div class="req-row">
      <label class="req-label">Method</label>
      <div style="display: flex; gap: 8px;">
        <select
          :value="request.method"
          @change="(e: any) => { request.method = e.target.value; update() }"
          class="req-select method-select"
        >
          <option v-for="m in methods" :key="m.value" :value="m.value">{{ m.label }}</option>
        </select>
        <input
          :value="request.url"
          @input="(e: any) => { request.url = e.target.value; update() }"
          placeholder="/api/..."
          class="req-input url-input"
        />
      </div>
    </div>

    <!-- Name -->
    <div class="req-row">
      <label class="req-label">名称</label>
      <input
        :value="request.name"
        @input="(e: any) => { request.name = e.target.value; update() }"
        placeholder="请求名称(如: 登录)"
        class="req-input"
      />
    </div>

    <!-- Headers -->
    <div class="req-section">
      <div class="req-section-header">
        <span>Headers</span>
        <span class="req-section-hint">HTTP 请求头</span>
      </div>
      <KeyValueEditor
        :model-value="request.headers"
        key-placeholder="Header Name"
        value-placeholder="Header Value"
        key-label="Name"
        value-label="Value"
        add-label="添加 Header"
        @update:model-value="onHeadersUpdate"
      />
    </div>

    <!-- Body -->
    <div class="req-section">
      <div class="req-section-header">
        <span>Body</span>
        <select
          :value="request.bodyType"
          @change="(e: any) => { request.bodyType = e.target.value; update() }"
          class="req-select-small"
        >
          <option v-for="b in bodyTypes" :key="b.value" :value="b.value">{{ b.label }}</option>
        </select>
      </div>

      <div v-if="request.bodyType === 'none'" class="req-body-empty">
        此请求不包含请求体
      </div>

      <div v-else-if="request.bodyType === 'json'">
        <textarea
          :value="request.bodyJson"
          @input="(e: any) => { request.bodyJson = e.target.value; update() }"
          rows="8"
          placeholder='{ "key": "value" }'
          class="req-textarea"
          :class="{ 'invalid': !jsonValid }"
        />
        <div v-if="!jsonValid" class="req-error">
          ⚠️ {{ jsonError }}
        </div>
        <div v-else class="req-hint">
          支持变量占位符: <code>&#123;&#123;user&#125;&#125;</code> <code>&#123;&#123;token&#125;&#125;</code>
        </div>
      </div>

      <div v-else-if="request.bodyType === 'raw'">
        <textarea
          :value="request.bodyRaw"
          @input="(e: any) => { request.bodyRaw = e.target.value; update() }"
          rows="6"
          placeholder="原始请求体(如 application/xml 或纯文本)"
          class="req-textarea"
        />
      </div>
    </div>

    <!-- Extractors -->
    <div class="req-section">
      <div class="req-section-header">
        <span>Extractors</span>
        <span class="req-section-hint">从响应中提取变量(后续请求可用 <code>&#123;&#123;name&#125;&#125;</code>)</span>
      </div>
      <KeyValueEditor
        :model-value="request.extractors"
        key-placeholder="变量名"
        value-placeholder="JSONPath,如 $.data.token"
        key-label="Name"
        value-label="JSONPath"
        key-width="30%"
        add-label="添加 Extractor"
        @update:model-value="onExtractorsUpdate"
      />
    </div>

    <!-- Assertions -->
    <div class="req-section">
      <div class="req-section-header">
        <label style="display: flex; align-items: center; gap: 8px; cursor: pointer;">
          <input
            type="checkbox"
            :checked="request.assertions.enabled"
            @change="(e: any) => { request.assertions.enabled = e.target.checked; update() }"
          />
          <span>Assertions</span>
        </label>
        <span class="req-section-hint">响应断言(失败计入错误)</span>
      </div>

      <div v-if="request.assertions.enabled" class="req-assertions">
        <!-- Status codes -->
        <div class="assertion-row">
          <label class="assertion-label">预期状态码:</label>
          <div class="assertion-values">
            <div
              v-for="(code, idx) in request.assertions.status"
              :key="idx"
              class="status-chip"
            >
              {{ code }}
              <button @click="removeStatus(idx)" class="chip-remove">✕</button>
            </div>
            <div style="display: flex; gap: 4px; align-items: center;">
              <input
                v-model.number="newStatus"
                type="number"
                min="100"
                max="599"
                placeholder="200"
                class="assertion-add-input"
              />
              <button @click="addStatus" class="assertion-add-btn">+ 添加</button>
          </div>
          </div>
        </div>

        <!-- Latency -->
        <div class="assertion-row">
          <label class="assertion-label">P95 延迟上限 (ms):</label>
          <input
            :value="request.assertions.maxLatencyMs"
            @input="(e: any) => { request.assertions.maxLatencyMs = Number(e.target.value); update() }"
            type="number"
            min="0"
            placeholder="0 = 不检查"
            class="req-input-inline"
          />
        </div>

        <!-- Body contains -->
        <div class="assertion-row">
          <label class="assertion-label">Body 必须包含:</label>
          <div style="flex: 1;">
            <div
              v-for="(s, idx) in request.assertions.bodyContains"
              :key="idx"
              style="display: flex; gap: 4px; margin-bottom: 4px;"
            >
              <input
                :value="s"
                @input="(e: any) => { request.assertions.bodyContains[idx] = e.target.value; update() }"
                placeholder="字符串片段"
                class="req-input"
                style="flex: 1;"
              />
              <button @click="removeBodyContains(idx)" class="chip-remove" style="width: 28px; height: 28px;">✕</button>
            </div>
            <button @click="addBodyContains" class="assertion-add-btn">+ 添加</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Think Time -->
    <div class="req-section">
      <div class="req-section-header">
        <span>Think Time</span>
        <span class="req-section-hint">请求之间的等待时间(毫秒)</span>
      </div>
      <div style="display: flex; align-items: center; gap: 12px;">
        <input
          :value="request.thinkTime"
          @input="(e: any) => { request.thinkTime = Number(e.target.value); update() }"
          type="number"
          min="0"
          step="100"
          placeholder="0"
          class="req-input-inline"
          style="width: 100px;"
        />
        <span style="color: var(--text-secondary); font-size: 13px;">ms</span>
        <span style="color: var(--text-tertiary); font-size: 12px;">(0 = 无)</span>
      </div>
    </div>

    <!-- Weight -->
    <div class="req-section">
      <div class="req-section-header">
        <span>Weight</span>
        <span class="req-section-hint">多请求场景下的执行权重(用于流量分配)</span>
      </div>
      <input
        :value="request.weight"
        @input="(e: any) => { request.weight = Number(e.target.value); update() }"
        type="number"
        min="1"
        class="req-input-inline"
        style="width: 100px;"
      />
    </div>
  </div>
</template>

<style scoped>
.req-editor {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.req-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.req-label {
  width: 80px;
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: 600;
}

.req-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  background: var(--bg-base);
  border: 1px solid var(--border);
  border-radius: 6px;
}

.req-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.req-section-hint {
  font-size: 11px;
  font-weight: 400;
  color: var(--text-tertiary);
}

.req-input {
  flex: 1;
  padding: 6px 10px;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 13px;
  font-family: var(--font-mono);
  outline: none;
}

.req-input:focus {
  border-color: var(--primary);
}

.req-input-inline {
  padding: 6px 10px;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
}

.req-select {
  padding: 6px 12px;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  min-width: 100px;
}

.req-select-small {
  padding: 4px 8px;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 12px;
  cursor: pointer;
}

.url-input {
  flex: 1;
  font-family: var(--font-mono);
}

.method-select {
  min-width: 110px;
}

.req-textarea {
  width: 100%;
  padding: 8px 12px;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 13px;
  font-family: var(--font-mono);
  outline: none;
  resize: vertical;
  min-height: 120px;
}

.req-textarea:focus {
  border-color: var(--primary);
}

.req-textarea.invalid {
  border-color: var(--error);
}

.req-body-empty {
  padding: 16px;
  text-align: center;
  color: var(--text-tertiary);
  font-size: 13px;
}

.req-error {
  margin-top: 4px;
  padding: 6px 10px;
  background: rgba(255, 77, 79, 0.1);
  color: var(--error);
  border-radius: 4px;
  font-size: 12px;
}

.req-hint {
  margin-top: 4px;
  color: var(--text-tertiary);
  font-size: 11px;
}

.req-hint code {
  background: var(--bg-hover);
  padding: 1px 4px;
  border-radius: 3px;
  margin-right: 4px;
  font-family: var(--font-mono);
  color: var(--primary);
}

.req-assertions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.assertion-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.assertion-label {
  width: 140px;
  font-size: 12px;
  color: var(--text-secondary);
  padding-top: 6px;
  flex-shrink: 0;
}

.assertion-values {
  flex: 1;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.status-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: var(--primary-bg);
  color: var(--primary);
  border-radius: 4px;
  font-family: var(--font-mono);
  font-size: 13px;
}

.chip-remove {
  background: transparent;
  color: inherit;
  border: none;
  cursor: pointer;
  padding: 0 4px;
  font-size: 11px;
  border-radius: 3px;
}

.chip-remove:hover {
  background: rgba(255, 77, 79, 0.2);
  color: var(--error);
}

.assertion-add-input {
  width: 80px;
  padding: 4px 8px;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 13px;
  font-family: var(--font-mono);
  outline: none;
}

.assertion-add-btn {
  padding: 4px 12px;
  background: transparent;
  color: var(--primary);
  border: 1px dashed var(--primary);
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
}

.assertion-add-btn:hover {
  background: var(--primary-bg);
}
</style>