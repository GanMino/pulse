<script setup lang="ts">
// 通用键值对编辑器 - 用于 Headers / Extractors / Body Contains 等
import { ref } from 'vue'

interface KVEntry {
  uid: string
  key: string
  value: string
  enabled: boolean
}

const props = defineProps<{
  modelValue: KVEntry[]
  keyPlaceholder?: string
  valuePlaceholder?: string
  valueLabel?: string       // 值的列标题
  keyLabel?: string         // key 的列标题
  keyWidth?: string         // key 列宽度,如 "30%"
  addLabel?: string         // 添加按钮文字
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: KVEntry[]): void
}>()

function addEntry() {
  const newEntry: KVEntry = {
    uid: `kv_${Date.now()}_${Math.random().toString(36).slice(2, 7)}`,
    key: '',
    value: '',
    enabled: true,
  }
  emit('update:modelValue', [...props.modelValue, newEntry])
}

function removeEntry(uid: string) {
  emit('update:modelValue', props.modelValue.filter(e => e.uid !== uid))
}

function updateEntry(uid: string, field: 'key' | 'value' | 'enabled', val: string | boolean) {
  emit('update:modelValue', props.modelValue.map(e =>
    e.uid === uid ? { ...e, [field]: val } : e
  ))
}

function moveEntry(uid: string, direction: -1 | 1) {
  const idx = props.modelValue.findIndex(e => e.uid === uid)
  if (idx < 0) return
  const newIdx = idx + direction
  if (newIdx < 0 || newIdx >= props.modelValue.length) return

  const arr = [...props.modelValue]
  const [item] = arr.splice(idx, 1)
  arr.splice(newIdx, 0, item)
  emit('update:modelValue', arr)
}
</script>

<template>
  <div class="kv-editor">
    <!-- Header row -->
    <div class="kv-header">
      <div style="width: 32px;"></div>
        <div style="width: 32px; text-align: center; font-size: 11px; color: var(--text-secondary);">启用</div>
        <div :style="{ width: keyWidth || '35%', fontSize: '11px', color: 'var(--text-secondary)', paddingLeft: '8px' }">
          {{ keyLabel || 'Key' }}
        </div>
        <div style="flex: 1; font-size: 11px; color: var(--text-secondary); padding-left: 8px;">
          {{ valueLabel || 'Value' }}
        </div>
        <div style="width: 56px;"></div>
    </div>

    <!-- Empty state -->
    <div v-if="modelValue.length === 0" class="kv-empty">
      暂无条目,点击下方按钮添加
    </div>

    <!-- Entries -->
    <div v-for="(entry, idx) in modelValue" :key="entry.uid" class="kv-row">
      <div class="kv-controls-left">
        <button
          @click="moveEntry(entry.uid, -1)"
          :disabled="idx === 0"
          class="kv-icon-btn"
          title="上移"
        >↑</button>
        <button
          @click="moveEntry(entry.uid, 1)"
          :disabled="idx === modelValue.length - 1"
          class="kv-icon-btn"
          title="下移"
        >↓</button>
      </div>

      <input
        type="checkbox"
        :checked="entry.enabled"
        @change="(e: any) => updateEntry(entry.uid, 'enabled', e.target.checked)"
        style="width: 16px; height: 16px; cursor: pointer;"
      />

      <input
        type="text"
        :value="entry.key"
        @input="(e: any) => updateEntry(entry.uid, 'key', e.target.value)"
        :placeholder="keyPlaceholder || 'key'"
        class="kv-input"
        :style="{ width: keyWidth || '35%' }"
      />

      <input
        type="text"
        :value="entry.value"
        @input="(e: any) => updateEntry(entry.uid, 'value', e.target.value)"
        :placeholder="valuePlaceholder || 'value'"
        class="kv-input"
        style="flex: 1;"
      />

      <button @click="removeEntry(entry.uid)" class="kv-remove-btn" title="删除">✕</button>
    </div>

    <!-- Add button -->
    <button @click="addEntry" class="kv-add-btn">
      ➕ {{ addLabel || '添加' }}
    </button>
  </div>
</template>

<style scoped>
.kv-editor {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.kv-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 4px;
  font-weight: 600;
}

.kv-empty {
  padding: 16px;
  text-align: center;
  color: var(--text-tertiary);
  background: var(--bg-base);
  border: 1px dashed var(--border);
  border-radius: 6px;
  font-size: 13px;
}

.kv-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.kv-controls-left {
  display: flex;
  flex-direction: column;
  gap: 2px;
  width: 32px;
}

.kv-icon-btn {
  width: 20px;
  height: 16px;
  background: transparent;
  color: var(--text-tertiary);
  border: none;
  cursor: pointer;
  font-size: 10px;
  padding: 0;
  border-radius: 2px;
}

.kv-icon-btn:hover:not(:disabled) {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.kv-icon-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.kv-input {
  padding: 6px 10px;
  background: var(--bg-base);
  border: 1px solid var(--border);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 13px;
  font-family: var(--font-mono);
  outline: none;
}

.kv-input:focus {
  border-color: var(--primary);
}

.kv-remove-btn {
  width: 28px;
  height: 28px;
  background: transparent;
  color: var(--text-tertiary);
  border: 1px solid transparent;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  padding: 0;
}

.kv-remove-btn:hover {
  background: rgba(255, 77, 79, 0.1);
  color: var(--error);
  border-color: var(--error);
}

.kv-add-btn {
  padding: 8px 12px;
  background: transparent;
  color: var(--primary);
  border: 1px dashed var(--primary);
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  margin-top: 4px;
}

.kv-add-btn:hover {
  background: var(--primary-bg);
}
</style>