<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { NCard, NButton, NStatistic, NSpace, NText, NGrid, NGi } from 'naive-ui'

const route = useRoute()
const router = useRouter()
const reportId = route.params.id
</script>

<template>
  <div>
    <NSpace align="center" justify="space-between" style="margin-bottom: 16px;">
      <NSpace align="center">
        <NButton quaternary @click="router.push({ name: 'reports' })">← Reports</NButton>
        <NText style="font-size: 18px;">Report #{{ reportId }}</NText>
      </NSpace>
    </NSpace>

    <NGrid :cols="5" :x-gap="12">
      <NGi>
        <NCard :bordered="false" embedded>
          <NStatistic label="Total Requests" :value="412341" />
        </NCard>
      </NGi>
      <NGi>
        <NCard :bordered="false" embedded>
          <NStatistic label="Avg RPS" :value="1374" />
        </NCard>
      </NGi>
      <NGi>
        <NCard :bordered="false" embedded>
          <NStatistic label="P95 Latency" :value="156" />
        </NCard>
      </NGi>
      <NGi>
        <NCard :bordered="false" embedded>
          <NStatistic label="Total Errors" :value="1732" />
        </NCard>
      </NGi>
      <NGi>
        <NCard :bordered="false" embedded>
          <NStatistic label="Duration" value="5m" />
        </NCard>
      </NGi>
    </NGrid>

    <NCard :bordered="false" embedded style="margin-top: 16px;">
      <NText depth="3">报告详情将在测试运行后填充。</NText>
    </NCard>
  </div>
</template>