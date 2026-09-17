// 场景编辑器用的内部数据结构(区别于后端 DTO)
// 编辑器需要保存完整的 Config,DTO 只是简化的列表展示
export interface RequestEditor {
  uid: string                  // 编辑器内部的唯一 ID,用于 React key
  name: string
  method: string
  url: string
  headers: HeaderEntry[]
  bodyType: 'json' | 'form' | 'raw' | 'none'
  bodyJson: string
  bodyRaw: string
  extractors: ExtractorEntry[]
  assertions: AssertionEntry
  thinkTime: number             // ms,0 = 无
  weight: number               // 用于多请求时的权重
}

export interface HeaderEntry {
  uid: string
  key: string
  value: string
  enabled: boolean
}

export interface ExtractorEntry {
  uid: string
  name: string                  // 变量名
  jsonPath: string              // $.data.token
  enabled: boolean
}

export interface AssertionEntry {
  status: number[]              // [200, 201]
  maxLatencyMs: number          // P95 上限, 0 = 不检查
  bodyContains: string[]        // body 必须包含的字符串
  enabled: boolean
}

export interface VariableEntry {
  uid: string
  name: string                  // 变量名(用作 {{name}} 占位符)
  values: string[]              // 多个值 → 数据驱动
  enabled: boolean
}

export interface ThresholdConfig {
  p95Ms: number                 // P95 延迟上限 (ms), 0 = 不检查
  p99Ms: number
  errorRate: number             // 0-1
  minRPS: number                // 0 = 不检查
}

export interface LoadConfig {
  vus: number
  duration: string              // "5m", "30s", "1h"
  rampUpType: 'linear' | 'step' | 'wave' | 'none'
  rampUpDuration: string
  targetRPS: number             // 0 = 不限
  thinkTimeMs: number           // 默认 think time (ms)
}

export interface ScenarioEditorData {
  uid: string                   // 编辑器 ID,= "new" 或 scenario id
  name: string
  description: string
  requests: RequestEditor[]
  variables: VariableEntry[]
  thresholds: ThresholdConfig
  load: LoadConfig
  tags: string[]
}

export function createEmptyRequest(name = ''): RequestEditor {
  return {
    uid: generateUid(),
    name: name || 'New Request',
    method: 'GET',
    url: '/api/...',
    headers: [],
    bodyType: 'none',
    bodyJson: '{}',
    bodyRaw: '',
    extractors: [],
    assertions: {
      status: [200],
      maxLatencyMs: 0,
      bodyContains: [],
      enabled: true,
    },
    thinkTime: 0,
    weight: 1,
  }
}

export function createEmptyVariable(): VariableEntry {
  return {
    uid: generateUid(),
    name: '',
    values: [''],
    enabled: true,
  }
}

let _uidCounter = 0
export function generateUid(): string {
  _uidCounter++
  return `e_${Date.now()}_${_uidCounter}_${Math.random().toString(36).slice(2, 7)}`
}

// 转换为后端 DTO 的 Config
export function toScenarioConfig(data: ScenarioEditorData): any {
  return {
    load: {
      vus: data.load.vus,
      duration: data.load.duration,
      targetRPS: data.load.targetRPS || 0,
      rampUp: data.load.rampUpType !== 'none' ? {
        type: data.load.rampUpType,
        duration: data.load.rampUpDuration || '30s',
      } : undefined,
      thinkTime: msToDuration(data.load.thinkTimeMs),
    },
    requests: data.requests.map(r => ({
      name: r.name,
      method: r.method,
      url: r.url,
      headers: headersToObject(r.headers),
      body: r.bodyType === 'json' ? safeParseJSON(r.bodyJson) : undefined,
      bodyRaw: r.bodyType === 'raw' ? r.bodyRaw : undefined,
      extractors: extractorsToObject(r.extractors),
      assertions: r.assertions.enabled ? {
        status: r.assertions.status,
        maxLatencyMs: r.assertions.maxLatencyMs,
        bodyContains: r.assertions.bodyContains.filter(s => s.trim()),
      } : undefined,
      thinkTime: msToDuration(r.thinkTime),
      weight: r.weight,
    })),
    variables: variablesToObject(data.variables),
    thresholds: {
      p95Ms: data.thresholds.p95Ms,
      p99Ms: data.thresholds.p99Ms,
      errorRate: data.thresholds.errorRate,
      minRPS: data.thresholds.minRPS,
    },
  }
}

function headersToObject(headers: HeaderEntry[]): Record<string, string> {
  const obj: Record<string, string> = {}
  for (const h of headers) {
    if (h.enabled && h.key.trim()) {
      obj[h.key.trim()] = h.value
    }
  }
  return obj
}

function extractorsToObject(extractors: ExtractorEntry[]): Record<string, string> {
  const obj: Record<string, string> = {}
  for (const e of extractors) {
    if (e.enabled && e.name.trim() && e.jsonPath.trim()) {
      obj[e.name.trim()] = e.jsonPath.trim()
    }
  }
  return obj
}

function variablesToObject(vars: VariableEntry[]): Record<string, any> {
  const obj: Record<string, any> = {}
  for (const v of vars) {
    if (v.enabled && v.name.trim()) {
      obj[v.name.trim()] = v.values.length > 1 ? v.values : v.values[0]
    }
  }
  return obj
}

function safeParseJSON(text: string): any {
  if (!text.trim()) return undefined
  try {
    return JSON.parse(text)
  } catch {
    return undefined
  }
}

function msToDuration(ms: number): string {
  if (ms <= 0) return ''
  if (ms < 1000) return `${ms}ms`
  if (ms < 60000) return `${Math.round(ms / 1000)}s`
  return `${Math.round(ms / 60000)}m`
}

// 从后端 Config 解析为编辑器数据
export function fromScenarioConfig(scenarioId: number | string, config: any, fallbackName = '未命名场景'): ScenarioEditorData {
  const data: ScenarioEditorData = {
    uid: String(scenarioId),
    name: fallbackName,
    description: '',
    requests: [],
    variables: [],
    thresholds: { p95Ms: 0, p99Ms: 0, errorRate: 0, minRPS: 0 },
    load: { vus: 100, duration: '5m', rampUpType: 'linear', rampUpDuration: '30s', targetRPS: 0, thinkTimeMs: 0 },
    tags: [],
  }

  if (!config) return data

  if (config.load) {
    data.load = {
      vus: config.load.vus || 100,
      duration: config.load.duration || '5m',
      rampUpType: config.load.rampUp?.type || 'linear',
      rampUpDuration: config.load.rampUp?.duration || '30s',
      targetRPS: config.load.targetRPS || 0,
      thinkTimeMs: durationToMs(config.load.thinkTime),
    }
  }

  if (config.requests && Array.isArray(config.requests)) {
    data.requests = config.requests.map((r: any) => ({
      uid: generateUid(),
      name: r.name || 'Request',
      method: r.method || 'GET',
      url: r.url || '',
      headers: Object.entries(r.headers || {}).map(([k, v]) => ({
        uid: generateUid(),
        key: k,
        value: String(v),
        enabled: true,
      })),
      bodyType: r.bodyRaw ? 'raw' : (r.body ? 'json' : 'none'),
      bodyJson: r.body ? JSON.stringify(r.body, null, 2) : '{}',
      bodyRaw: r.bodyRaw || '',
      extractors: Object.entries(r.extractors || {}).map(([k, v]) => ({
        uid: generateUid(),
        name: k,
        jsonPath: String(v),
        enabled: true,
      })),
      assertions: {
        status: r.assertions?.status || [200],
        maxLatencyMs: r.assertions?.maxLatencyMs || 0,
        bodyContains: r.assertions?.bodyContains || [],
        enabled: !!r.assertions,
      },
      thinkTime: durationToMs(r.thinkTime),
      weight: r.weight || 1,
    }))
  }

  if (config.variables) {
    data.variables = Object.entries(config.variables).map(([k, v]) => ({
      uid: generateUid(),
      name: k,
      values: Array.isArray(v) ? v.map(String) : [String(v)],
      enabled: true,
    }))
  }

  if (config.thresholds) {
    data.thresholds = {
      p95Ms: config.thresholds.p95Ms || 0,
      p99Ms: config.thresholds.p99Ms || 0,
      errorRate: config.thresholds.errorRate || 0,
      minRPS: config.thresholds.minRPS || 0,
    }
  }

  return data
}

function durationToMs(d: string | undefined): number {
  if (!d) return 0
  // 简单解析:支持 "5m", "30s", "1h", "500ms"
  const match = /^(\d+(?:\.\d+)?)(ms|s|m|h)?$/.exec(d.trim())
  if (!match) return 0
  const num = parseFloat(match[1])
  const unit = match[2] || 's'
  switch (unit) {
    case 'ms': return Math.round(num)
    case 's': return Math.round(num * 1000)
    case 'm': return Math.round(num * 60000)
    case 'h': return Math.round(num * 3600000)
    default: return 0
  }
}

export function createEmptyEditorData(scenarioId?: number | string): ScenarioEditorData {
  return {
    uid: scenarioId ? String(scenarioId) : 'new',
    name: '未命名场景',
    description: '',
    requests: [createEmptyRequest('First Request')],
    variables: [],
    thresholds: { p95Ms: 0, p99Ms: 0, errorRate: 0, minRPS: 0 },
    load: { vus: 100, duration: '5m', rampUpType: 'linear', rampUpDuration: '30s', targetRPS: 0, thinkTimeMs: 0 },
    tags: [],
  }
}