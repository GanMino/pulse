import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as api from '@/api/scenarios'
import type { ProjectDTO } from '@/api/scenarios'

export const useProjectsStore = defineStore('projects', () => {
  const items = ref<ProjectDTO[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const defaultProject = ref<ProjectDTO | null>(null)

  async function fetchList() {
    loading.value = true
    error.value = null
    try {
      const list = await api.listProjects()
      items.value = list
    } catch (e: any) {
      error.value = e?.message || 'Failed to fetch projects'
    } finally {
      loading.value = false
    }
  }

  async function ensureDefault() {
    try {
      defaultProject.value = await api.getOrCreateDefaultProject()
      return defaultProject.value
    } catch (e: any) {
      error.value = e?.message || 'Failed to get default project'
      return null
    }
  }

  async function create(name: string, description = '', color = '#3370FF') {
    try {
      const dto = await api.createProject({ name, description, color })
      items.value.push(dto)
      return dto
    } catch (e: any) {
      error.value = e?.message || 'Failed to create project'
      return null
    }
  }

  function getById(id: number): ProjectDTO | undefined {
    return items.value.find(p => p.id === id)
  }

  const hasItems = computed(() => items.value.length > 0)

  return {
    items,
    loading,
    error,
    defaultProject,
    hasItems,
    fetchList,
    ensureDefault,
    create,
    getById,
  }
})