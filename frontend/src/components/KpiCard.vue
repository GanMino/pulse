<script setup lang="ts">
// KPI 卡片组件
import { computed } from 'vue'

const props = defineProps<{
  label: string
  value: number | string
  unit?: string
  trend?: number          // 变化率(%):正数=上升,负数=下降
  status?: 'success' | 'warning' | 'error' | 'normal'
  threshold?: { warning?: number; error?: number; reverse?: boolean }
  icon?: string
}>()

const formattedValue = computed(() => {
  if (typeof props.value === 'number') {
    if (props.value >= 10000) {
      return (props.value / 1000).toFixed(1) + 'k'
    }
    if (props.value >= 1000) {
      return props.value.toLocaleString()
    }
    if (Number.isInteger(props.value)) {
      return String(props.value)
    }
    return props.value.toFixed(2)
  }
  return props.value
})

const trendArrow = computed(() => {
  if (props.trend === undefined) return ''
  if (props.trend > 0) return '↑'
  if (props.trend < 0) return '↓'
  return '→'
})

const trendColor = computed(() => {
  if (props.trend === undefined) return 'var(--text-tertiary)'
  if (props.trend > 0) return 'var(--success)'
  if (props.trend < 0) return 'var(--error)'
  return 'var(--text-tertiary)'
})

const actualStatus = computed(() => {
  if (props.status) return props.status
  if (!props.threshold) return 'normal'

  const v = typeof props.value === 'number' ? props.value : 0
  const reverse = props.threshold.reverse || false

  if (props.threshold.error !== undefined) {
    if (reverse && v < props.threshold.error) return 'error'
    if (!reverse && v > props.threshold.error) return 'error'
  }
  if (props.threshold.warning !== undefined) {
    if (reverse && v < props.threshold.warning) return 'warning'
    if (!reverse && v > props.threshold.warning) return 'warning'
  }
  return 'success'
})

const statusColor = computed(() => {
  switch (actualStatus.value) {
    case 'success': return 'var(--success)'
    case 'warning': return 'var(--warning)'
    case 'error': return 'var(--error)'
    default: return 'var(--text-secondary)'
  }
})
</script>

<template>
  <div :class="['kpi-card', `status-${actualStatus}`]">
    <div class="kpi-header">
      <span v-if="icon" class="kpi-icon">{{ icon }}</span>
      <span class="kpi-label">{{ label }}</span>
    </div>
    <div class="kpi-value-row">
      <div class="kpi-value" :style="{ color: statusColor }">
        {{ formattedValue }}
      </div>
      <span v-if="unit" class="kpi-unit">{{ unit }}</span>
    </div>
    <div v-if="trend !== undefined" class="kpi-trend" :style="{ color: trendColor }">
      {{ trendArrow }} {{ Math.abs(trend).toFixed(1) }}% vs 1m ago
    </div>
  </div>
</template>

<style scoped>
.kpi-card {
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  transition: border-color 0.2s;
}

.kpi-card.status-warning {
  border-color: var(--warning);
  background: rgba(250, 173, 20, 0.04);
}

.kpi-card.status-error {
  border-color: var(--error);
  background: rgba(255, 77, 79, 0.04);
}

.kpi-card.status-success {
  border-color: var(--success);
}

.kpi-header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.kpi-icon {
  font-size: 14px;
}

.kpi-label {
  flex: 1;
}

.kpi-value-row {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.kpi-value {
  font-size: 32px;
  font-weight: 600;
  font-family: var(--font-mono);
  line-height: 1.1;
}

.kpi-unit {
  font-size: 14px;
  color: var(--text-secondary);
  font-weight: 500;
}

.kpi-trend {
  font-size: 11px;
  font-weight: 500;
  margin-top: 2px;
}
</style>