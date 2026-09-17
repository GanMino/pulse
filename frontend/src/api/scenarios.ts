// 类型定义(镜像后端 DTO)
export interface ScenarioDTO {
  id: number
  uuid: string
  projectId: number
  name: string
  description: string
  vus: number
  duration: string
  requestCount: number
  source: string
  status: string
  tags: string[]
  version: number
  createdAt: string
  updatedAt: string
  requests?: RequestPreviewDTO[]
}

export interface RequestPreviewDTO {
  name: string
  method: string
  url: string
}

export interface ListScenariosRequest {
  projectId?: number
  search?: string
  source?: string
  status?: string
  tag?: string
  limit?: number
  offset?: number
}

export interface ListScenariosResponse {
  items: ScenarioDTO[]
  total: number
  limit: number
  offset: number
}

export interface SaveScenarioRequest {
  id?: number
  projectId?: number
  name: string
  description?: string
  config?: any
  tags?: string[]
  source?: string
}

export interface ProjectDTO {
  id: number
  uuid: string
  name: string
  description: string
  color: string
  scenarioCount: number
  createdAt: string
  updatedAt: string
}

export interface CreateProjectRequest {
  name: string
  description?: string
  color?: string
}

export interface AgentDTO {
  id: number
  uuid: string
  name: string
  address: string
  status: string
  lastHeartbeat: string
  tags: string[]
  maxVUs: number
  createdAt: string
}

export interface UserDTO {
  id: number
  uuid: string
  username: string
  email: string
  displayName: string
  role: string
}

// 通用响应包装
export type WailsResult<T> = Promise<T>

// ============================================
// 错误处理
// ============================================

export class PulseError extends Error {
  constructor(message: string, public code?: string) {
    super(message)
    this.name = 'PulseError'
  }
}

// 统一的 API 调用包装
async function call<T>(fn: () => Promise<T>, errorMsg: string): Promise<T> {
  try {
    return await fn()
  } catch (e: any) {
    console.error(`[Pulse API] ${errorMsg}:`, e)
    throw new PulseError(e?.message || String(e))
  }
}

// ============================================
// API 调用方法(通过 Wails 绑定)
// ============================================

import {
  ListProjects as WailsListProjects,
  GetOrCreateDefaultProject as WailsGetOrCreateDefaultProject,
  CreateProject as WailsCreateProject,
  ListScenarios as WailsListScenarios,
  GetScenario as WailsGetScenario,
  GetScenarioByUUID as WailsGetScenarioByUUID,
  SaveScenario as WailsSaveScenario,
  DeleteScenario as WailsDeleteScenario,
  DeleteScenarios as WailsDeleteScenarios,
  DuplicateScenario as WailsDuplicateScenario,
  ListAgents as WailsListAgents,
  RegisterLocalAgent as WailsRegisterLocalAgent,
  GetCurrentUser as WailsGetCurrentUser,
  ListAllTestRuns as WailsListAllTestRuns,
  ListReports as WailsListReports,
  GetAppInfo as WailsGetAppInfo,
  StartRun as WailsStartRun,
  PauseRun as WailsPauseRun,
  ResumeRun as WailsResumeRun,
  StopRun as WailsStopRun,
  GetRun as WailsGetRun,
  IsRunActive as WailsIsRunActive,
} from '../../wailsjs/go/main/App'

// 应用信息
export function getAppInfo() {
  return call(() => WailsGetAppInfo(), 'GetAppInfo')
}

// ============================================
// HAR 导入
// ============================================

export interface ImportHARRequest {
  name?: string
  description?: string
  baseUrl?: string
  harData: number[]        // 字节数组(JS 传给 Go)
  vus?: number
  duration?: string
}

export interface ImportHARResponse {
  scenario: ScenarioDTO
  stats: {
    totalEntries: number
    importedRequests: number
    duplicatesSkipped: number
  }
}

import { ImportHAR as WailsImportHAR } from '../../wailsjs/go/main/App'

export function importHar(req: ImportHARRequest) {
  return call(() => WailsImportHAR(req), 'ImportHAR')
}

// 用户
export function getCurrentUser() {
  return call(() => WailsGetCurrentUser(), 'GetCurrentUser')
}

// 项目
export function listProjects() {
  return call(() => WailsListProjects(), 'ListProjects')
}

export function getOrCreateDefaultProject() {
  return call(() => WailsGetOrCreateDefaultProject(), 'GetOrCreateDefaultProject')
}

export function createProject(req: CreateProjectRequest) {
  return call(() => WailsCreateProject(req), 'CreateProject')
}

// 场景
export function listScenarios(req: ListScenariosRequest = {}) {
  return call(() => WailsListScenarios(req), 'ListScenarios')
}

export function getScenario(id: number) {
  return call(() => WailsGetScenario(id), 'GetScenario')
}

export function saveScenario(req: SaveScenarioRequest) {
  return call(() => WailsSaveScenario(req), 'SaveScenario')
}

export function deleteScenario(id: number) {
  return call(() => WailsDeleteScenario(id), 'DeleteScenario')
}

export function deleteScenarios(ids: number[]) {
  return call(() => WailsDeleteScenarios(ids), 'DeleteScenarios')
}

export function duplicateScenario(id: number) {
  return call(() => WailsDuplicateScenario(id), 'DuplicateScenario')
}

// Agent
export function listAgents() {
  return call(() => WailsListAgents(), 'ListAgents')
}

export function registerLocalAgent() {
  return call(() => WailsRegisterLocalAgent(), 'RegisterLocalAgent')
}

// TestRun
export function listAllTestRuns(limit = 50) {
  return call(() => WailsListAllTestRuns(limit), 'ListAllTestRuns')
}

// Report
export function listReports(limit = 50, offset = 0) {
  return call(() => WailsListReports(limit, offset), 'ListReports')
}

// ============================================
// Test Run API
// ============================================

export interface StartRunRequest {
  scenarioId: number
  runtime?: {
    vus?: number
    duration?: string
    targetRPS?: number
  }
}

export interface StartRunResponse {
  runId: number
  uuid: string
  startedAt: string
  scenarioName: string
}

export interface TestRunDTO {
  id: number
  uuid: string
  scenarioId: number
  scenarioName: string
  projectId: number
  status: string
  startedAt: string
  finishedAt: string
  durationMs: number
  triggerType: string
  errorMessage: string
  summary?: any
  createdAt: string
}

export interface MetricSnapshot {
  timestamp: string
  elapsedMs: number
  vusActive: number
  vusTarget: number
  rps: number
  totalRequests: number
  totalErrors: number
  errorRate: number
  latencyMs: {
    min: number
    avg: number
    max: number
    p50: number
    p90: number
    p95: number
    p99: number
  }
  statusCodes: Record<string, number>
  perRequest: Record<string, RequestSnapshot>
}

export interface RequestSnapshot {
  name: string
  count: number
  errors: number
  rps: number
  latencyMs: {
    min: number
    avg: number
    max: number
    p50: number
    p90: number
    p95: number
    p99: number
  }
}

export function startRun(req: StartRunRequest) {
  return call(() => WailsStartRun(req), 'StartRun')
}

export function pauseRun(runId: number) {
  return call(() => WailsPauseRun(runId), 'PauseRun')
}

export function resumeRun(runId: number) {
  return call(() => WailsResumeRun(runId), 'ResumeRun')
}

export function stopRun(runId: number) {
  return call(() => WailsStopRun(runId), 'StopRun')
}

export function getRun(runId: number) {
  return call(() => WailsGetRun(runId), 'GetRun')
}

export function isRunActive(runId: number) {
  return call(() => WailsIsRunActive(runId), 'IsRunActive')
}

// 监听后端事件(由 Wails Runtime 自动桥接)
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

export interface RunEvent {
  type: string
  payload: Record<string, any>
}

/**
 * 订阅 metric:update 事件
 * @param runId 测试运行 ID
 * @param callback 收到指标快照时触发
 * @returns 取消订阅函数
 */
export function subscribeMetric(runId: number, callback: (snapshot: MetricSnapshot) => void): () => void {
  const eventName = `metric:update:${runId}`
  EventsOn(eventName, (data: any) => {
    callback(data as MetricSnapshot)
  })
  return () => EventsOff(eventName)
}

/**
 * 订阅 test:xxx 生命周期事件
 */
export function subscribeRunEvents(runId: number, callbacks: {
  onStarted?: () => void
  onPaused?: () => void
  onResumed?: () => void
  onStopped?: () => void
  onCompleted?: () => void
  onFailed?: (err: string) => void
}): () => void {
  const handlers: Array<[string, (...args: any[]) => void]> = []

  if (callbacks.onStarted) {
    const h = (_: any) => callbacks.onStarted!()
    EventsOn(`test:started:${runId}`, h)
    handlers.push([`test:started:${runId}`, h])
  }
  if (callbacks.onCompleted) {
    const h = (_: any) => callbacks.onCompleted!()
    EventsOn(`test:completed:${runId}`, h)
    handlers.push([`test:completed:${runId}`, h])
  }
  if (callbacks.onFailed) {
    const h = (data: any) => callbacks.onFailed!(data?.error || 'Unknown error')
    EventsOn(`test:failed:${runId}`, h)
    handlers.push([`test:failed:${runId}`, h])
  }

  return () => {
    for (const [name, h] of handlers) {
      EventsOff(name, h as any)
    }
  }
}