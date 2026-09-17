<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useUIStore } from '@/stores/ui'
import {
  NLayout,
  NLayoutHeader,
  NLayoutContent,
  NLayoutFooter,
  NMenu,
  NIcon,
  NText,
  NSpace,
  NButton,
} from 'naive-ui'

// Wails 生成的 API
import { GetAppInfo } from '../../wailsjs/go/main/App'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const uiStore = useUIStore()

const appInfo = ref<{ name: string; version: string }>({ name: 'Pulse', version: '0.1.0' })

onMounted(async () => {
  try {
    appInfo.value = await GetAppInfo()
  } catch (e) {
    console.error('Failed to get app info:', e)
  }

  // 监听 ⌘K 打开命令面板
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})

function handleKeydown(e: KeyboardEvent) {
  // ⌘K (macOS) or Ctrl+K (Windows/Linux)
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault()
    uiStore.commandPaletteOpen = true
  }
}

const menuOptions = computed(() => [
  {
    key: 'dashboard',
    label: t('nav.dashboard'),
    icon: () => h(NIcon, null, { default: () => '🏠' }),
  },
  {
    key: 'scenarios',
    label: t('nav.scenarios'),
    icon: () => h(NIcon, null, { default: () => '📋' }),
  },
  {
    key: 'reports',
    label: t('nav.reports'),
    icon: () => h(NIcon, null, { default: () => '📊' }),
  },
  {
    key: 'agents',
    label: t('nav.agents'),
    icon: () => h(NIcon, null, { default: () => '🖥️' }),
  },
  {
    key: 'settings',
    label: t('nav.settings'),
    icon: () => h(NIcon, null, { default: () => '⚙️' }),
  },
])

function navigate(key: string) {
  router.push({ name: key })
}

// 监听路由变化,高亮当前项
const activeKey = computed(() => {
  if (route.name === 'dashboard') return 'dashboard'
  if (route.name && ['scenarios', 'scenario-editor'].includes(route.name as string)) return 'scenarios'
  if (route.name && ['reports', 'report-detail'].includes(route.name as string)) return 'reports'
  if (route.name === 'agents') return 'agents'
  if (route.name === 'settings') return 'settings'
  return ''
})
</script>

<template>
  <NLayout style="height: 100vh">
    <!-- 顶部导航 -->
    <NLayoutHeader bordered style="height: 56px; padding: 0 24px; display: flex; align-items: center;">
      <NSpace align="center" :wrap="false">
        <div style="display: flex; align-items: center; gap: 8px;">
          <span style="font-size: 20px;">💓</span>
          <NText strong style="font-size: 18px;">{{ appInfo.name }}</NText>
          <NText depth="3" style="font-size: 12px;">v{{ appInfo.version }}</NText>
        </div>

        <div style="width: 1px; height: 24px; background: var(--border); margin: 0 16px;"></div>

        <NMenu
          mode="horizontal"
          :options="menuOptions"
          :value="activeKey"
          @update:value="navigate"
          style="min-width: 400px;"
        />
      </NSpace>

      <div style="flex: 1;"></div>

      <NSpace align="center">
        <NButton quaternary @click="uiStore.commandPaletteOpen = true">
          🔍 ⌘K
        </NButton>
      </NSpace>
    </NLayoutHeader>

    <!-- 主内容 -->
    <NLayoutContent
      content-style="padding: 24px; height: calc(100vh - 56px - 32px); overflow: auto;"
      :native-scrollbar="false"
    >
      <slot />
    </NLayoutContent>

    <!-- 底部状态栏 -->
    <NLayoutFooter
      bordered
      style="height: 32px; padding: 0 24px; display: flex; align-items: center; font-size: 12px;"
    >
      <NSpace align="center" :wrap="false">
        <span style="color: var(--success);">●</span>
        <NText depth="3">Ready</NText>
        <NText depth="3">·</NText>
        <NText depth="3">Agent: local-1</NText>
        <NText depth="3">·</NText>
        <NText depth="3">DB: 0 MB</NText>
      </NSpace>

      <div style="flex: 1;"></div>

      <NText depth="3">v{{ appInfo.version }} · Apache 2.0</NText>
    </NLayoutFooter>
  </NLayout>
</template>