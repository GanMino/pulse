<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useUIStore } from '@/stores/ui'
import {
  NModal,
  NCard,
  NInput,
  NList,
  NListItem,
  NThing,
  NText,
  NEmpty,
} from 'naive-ui'

const router = useRouter()
const { t } = useI18n()
const uiStore = useUIStore()

const searchQuery = ref('')
const inputRef = ref<InstanceType<typeof NInput> | null>(null)

// 命令列表(MVP 简化版)
const commands = computed(() => [
  { id: 'new-scenario', label: t('dashboard.newScenario'), icon: '➕', shortcut: '⌘N', action: () => router.push({ name: 'scenario-editor' }) },
  { id: 'import-har', label: t('dashboard.importHar'), icon: '📥', shortcut: '⌘I', action: () => alert('Import HAR - TODO') },
  { id: 'export-report', label: '导出报告', icon: '📤', shortcut: '⌘E', action: () => alert('Export Report - TODO') },
  { id: 'toggle-theme', label: '切换主题', icon: '🎨', shortcut: '⌘⇧T', action: () => uiStore.toggleTheme() },
  { id: 'goto-dashboard', label: t('nav.dashboard'), icon: '🏠', action: () => router.push({ name: 'dashboard' }) },
  { id: 'goto-scenarios', label: t('nav.scenarios'), icon: '📋', action: () => router.push({ name: 'scenarios' }) },
  { id: 'goto-reports', label: t('nav.reports'), icon: '📊', action: () => router.push({ name: 'reports' }) },
  { id: 'goto-settings', label: t('nav.settings'), icon: '⚙️', action: () => router.push({ name: 'settings' }) },
])

const filteredCommands = computed(() => {
  if (!searchQuery.value) return commands.value
  const q = searchQuery.value.toLowerCase()
  return commands.value.filter(c => c.label.toLowerCase().includes(q))
})

watch(() => uiStore.commandPaletteOpen, async (open) => {
  if (open) {
    await nextTick()
    inputRef.value?.focus()
  } else {
    searchQuery.value = ''
  }
})

function executeCommand(cmd: typeof commands.value[0]) {
  cmd.action()
  uiStore.commandPaletteOpen = false
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    uiStore.commandPaletteOpen = false
  }
}
</script>

<template>
  <NModal
    v-model:show="uiStore.commandPaletteOpen"
    preset="card"
    :style="{ width: '600px' }"
    :bordered="false"
    :show-close="false"
    role="dialog"
    aria-modal="true"
    @keydown="handleKeydown"
  >
    <template #header>
      <NInput
        ref="inputRef"
        v-model:value="searchQuery"
        placeholder="搜索命令、场景、报告..."
        size="large"
        :bordered="false"
      />
    </template>

    <div style="max-height: 400px; overflow: auto;">
      <NList v-if="filteredCommands.length > 0" hoverable>
        <NListItem
          v-for="cmd in filteredCommands"
          :key="cmd.id"
          @click="executeCommand(cmd)"
          style="cursor: pointer;"
        >
          <NThing>
            <template #avatar>
              <span style="font-size: 20px;">{{ cmd.icon }}</span>
            </template>
            <NText>{{ cmd.label }}</NText>
            <template #suffix>
              <NText v-if="cmd.shortcut" depth="3" style="font-family: var(--font-mono); font-size: 12px;">
                {{ cmd.shortcut }}
              </NText>
            </template>
          </NThing>
        </NListItem>
      </NList>

      <NEmpty v-else description="未找到匹配项" />
    </div>

    <template #footer>
      <NText depth="3" style="font-size: 12px;">
        ↑↓ 选择  ⏎ 执行  ESC 关闭
      </NText>
    </template>
  </NModal>
</template>