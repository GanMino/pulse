import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUIStore = defineStore('ui', () => {
  // 主题:dark / light / system
  const theme = ref<'dark' | 'light' | 'system'>('dark')

  // 语言:zh-CN / en-US
  const language = ref<'zh-CN' | 'en-US'>('zh-CN')

  // 命令面板打开状态
  const commandPaletteOpen = ref(false)

  // 侧边栏折叠状态
  const sidebarCollapsed = ref(false)

  // 当前测试运行状态
  const isRunning = ref(false)
  const currentRunId = ref<string | null>(null)

  function toggleTheme() {
    const themes: Array<'dark' | 'light' | 'system'> = ['dark', 'light', 'system']
    const idx = themes.indexOf(theme.value)
    theme.value = themes[(idx + 1) % themes.length]
  }

  return {
    theme,
    language,
    commandPaletteOpen,
    sidebarCollapsed,
    isRunning,
    currentRunId,
    toggleTheme,
  }
})