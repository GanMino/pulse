<script setup lang="ts">
import { ref, computed, h } from 'vue'
import { darkTheme, NConfigProvider, NMessageProvider, NDialogProvider } from 'naive-ui'
import { RouterView } from 'vue-router'
import AppShell from '@/components/AppShell.vue'
import CommandPalette from '@/components/CommandPalette.vue'
import { useUIStore } from '@/stores/ui'

const uiStore = useUIStore()

const theme = computed(() => {
  if (uiStore.theme === 'dark') return darkTheme
  return null
})
</script>

<template>
  <NConfigProvider :theme="theme">
    <NMessageProvider>
      <NDialogProvider>
        <AppShell>
          <RouterView />
        </AppShell>
        <CommandPalette />
      </NDialogProvider>
    </NMessageProvider>
  </NConfigProvider>
</template>

<style>
html, body, #app {
  margin: 0;
  padding: 0;
  height: 100%;
  width: 100%;
  font-family: 'Inter', -apple-system, 'PingFang SC', 'Microsoft YaHei', sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
</style>