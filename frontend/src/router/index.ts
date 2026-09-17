import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'dashboard',
    component: () => import('@/views/Dashboard.vue'),
    meta: { title: 'Dashboard', icon: 'Home' },
  },
  {
    path: '/scenarios',
    name: 'scenarios',
    component: () => import('@/views/Scenarios.vue'),
    meta: { title: 'Scenarios', icon: 'Document' },
  },
  {
    path: '/scenarios/editor/:id?',
    name: 'scenario-editor',
    component: () => import('@/views/ScenarioEditor.vue'),
    meta: { title: 'Scenario Editor' },
  },
  {
    path: '/monitor/:id',
    name: 'live-monitor',
    component: () => import('@/views/LiveMonitor.vue'),
    meta: { title: 'Live Monitor' },
  },
  {
    path: '/reports',
    name: 'reports',
    component: () => import('@/views/Reports.vue'),
    meta: { title: 'Reports', icon: 'Chart' },
  },
  {
    path: '/reports/:id',
    name: 'report-detail',
    component: () => import('@/views/ReportDetail.vue'),
    meta: { title: 'Report Detail' },
  },
  {
    path: '/agents',
    name: 'agents',
    component: () => import('@/views/Agents.vue'),
    meta: { title: 'Agents', icon: 'Server' },
  },
  {
    path: '/settings',
    name: 'settings',
    component: () => import('@/views/Settings.vue'),
    meta: { title: 'Settings', icon: 'Settings' },
  },
  {
  path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router