<script setup lang="ts">
import { NCard, NText, NSpace, NAlert } from 'naive-ui'
</script>

<template>
  <div>
    <NSpace align="center" style="margin-bottom: 16px;">
      <NText style="font-size: 24px; font-weight: 600;">Agents</NText>
    </NSpace>

    <NAlert type="info" title="分布式能力规划中" style="margin-bottom: 16px;">
      Pulse 当前为单机模式,支持最高 20,000 VUs 并发。
      分布式 Agent(多机压测)将在 v0.3.0 中实现。
    </NAlert>

    <NCard :bordered="false" embedded>
      <NSpace vertical>
        <NSpace align="center">
          <span style="color: var(--success);">●</span>
          <NText strong>local-1</NText>
          <NText depth="3" style="font-size: 12px;">localhost</NText>
          <NText depth="3" style="font-size: 12px;">Status: Online</NText>
          <NText depth="3" style="font-size: 12px;">Load: 0 / 20,000</NText>
        </NSpace>
      </NSpace>
    </NCard>
  </div>
</template>