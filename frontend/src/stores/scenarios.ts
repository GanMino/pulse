import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as api from '@/api/scenarios'
import type { ScenarioDTO, ListScenariosRequest } from '@/api/scenarios'

export const useScenariosStore = defineStore('scenarios', () => {
  // 状态
  const items = ref<ScenarioDTO[]>([])
  const total = ref(0)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 过滤器状态
  const searchQuery = ref('')
  const projectIdFilter = ref<number | null>(null)

  // 计算属性
  const hasItems = computed(() => items.value.length > 0)
  const totalCount = computed(() => total.value)

  // 操作
  async function fetchList(opts: Partial<ListScenariosRequest> = {}) {
    loading.value = true
    error.value = null
    try {
      const req: ListScenariosRequest = {
        limit: 100,
        offset: 0,
        search: searchQuery.value || undefined,
        projectId: projectIdFilter.value || undefined,
        ...opts,
      }
      const response = await api.listScenarios(req)
      items.value = response.items
      total.value = response.total
    } catch (e: any) {
      error.value = e.message || 'Failed to fetch scenarios'
      console.error('fetchList error:', e)
    } finally {
      loading.value = false
    }
  }

  async function fetchById(id: number): Promise<ScenarioDTO | null> {
    try {
      return await api.getScenario(id)
    } catch (e: any) {
      error.value = e.message || 'Failed to fetch scenario'
      return null
    }
  }

  async function save(req: api.SaveScenarioRequest): Promise<ScenarioDTO | null> {
    try {
      const dto = await api.saveScenario(req)
      // 更新本地列表
      const idx = items.value.findIndex(s => s.id === dto.id)
      if (idx >= 0) {
        items.value[idx] = dto
      } else {
        items.value.unshift(dto)
        total.value++
      }
      return dto
    } catch (e: any) {
      error.value = e.message || 'Failed to save scenario'
      return null
    }
  }

  async function remove(id: number): Promise<boolean> {
    try {
      await api.deleteScenario(id)
      items.value = items.value.filter(s => s.id !== id)
      total.value = Math.max(0, total.value - 1)
      return true
    } catch (e: any) {
      error.value = e.message || 'Failed to delete scenario'
      return false
    }
  }

  async function removeMany(ids: number[]): Promise<number> {
    try {
      const count = await api.deleteScenarios(ids)
      items.value = items.value.filter(s => !ids.includes(s.id))
      total.value = Math.max(0, total.value - count)
      return count
    } catch (e: any) {
      error.value = e.message || 'Failed to delete scenarios'
      return 0
    }
  }

  async function duplicate(id: number): Promise<ScenarioDTO | null> {
    try {
      const dto = await api.duplicateScenario(id)
      items.value.unshift(dto)
      total.value++
      return dto
    } catch (e: any) {
      error.value = e.message || 'Failed to duplicate scenario'
      return null
    }
  }

  function clearError() {
    error.value = null
  }

  return {
    // state
    items,
    total,
    loading,
    error,
    searchQuery,
    projectIdFilter,
    // getters
    hasItems,
    totalCount,
    // actions
    fetchList,
    fetchById,
    save,
    remove,
    removeMany,
    duplicate,
    clearError,
  }
})