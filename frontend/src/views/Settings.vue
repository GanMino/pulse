<script setup lang="ts">
import { ref } from 'vue'
import {
  NCard,
  NSpace,
  NRadioGroup,
  NRadio,
  NText,
  NDivider,
  NSwitch,
  NInputNumber,
  NButton,
} from 'naive-ui'
import { useUIStore } from '@/stores/ui'

const uiStore = useUIStore()

const theme = ref<'dark' | 'light' | 'system'>('dark')
const language = ref<'zh-CN' | 'en-US'>('zh-CN')
const notifyOnComplete = ref(true)
const notifyOnFailure = ref(true)

const maxVUs = ref(20000)
const keepAlive = ref(true)
const http2 = ref(true)
const timeout = ref(30)
</script>

<template>
  <div>
    <NText style="font-size: 24px; font-weight: 600; display: block; margin-bottom: 16px;">
      Settings
    </NText>

    <NGrid :cols="2" :x-gap="16">
      <NGi>
        <NCard title="通用" :bordered="false" embedded>
          <NSpace vertical size="large">
            <div>
              <NText strong style="display: block; margin-bottom: 8px;">主题</NText>
              <NRadioGroup v-model:value="theme">
                <NRadio value="dark">Dark</NRadio>
                <NRadio value="light">Light</NRadio>
                <NRadio value="system">System</NRadio>
              </NRadioGroup>
            </div>

            <div>
              <NText strong style="display: block; margin-bottom: 8px;">语言</NText>
              <NRadioGroup v-model:value="language">
                <NRadio value="zh-CN">中文</NRadio>
                <NRadio value="en-US">English</NRadio>
              </NRadioGroup>
            </div>
          </NSpace>
        </NCard>
      </NGi>

      <NGi>
        <NCard title="引擎" :bordered="false" embedded>
          <NSpace vertical size="large">
            <div>
              <NText strong>最大 VUs</NText>
              <NInputNumber v-model:value="maxVUs" :min="100" :max="100000" style="width: 100%;" />
            </div>
            <div>
              <NSpace align="center">
                <NSwitch v-model:value="keepAlive" />
                <NText>Keep-Alive</NText>
              </NSpace>
            </div>
            <div>
              <NSpace align="center">
                <NSwitch v-model:value="http2" />
                <NText>HTTP/2</NText>
              </NSpace>
            </div>
            <div>
              <NText strong>超时(秒)</NText>
              <NInputNumber v-model:value="timeout" :min="1" :max="300" style="width: 100%;" />
            </div>
          </NSpace>
        </NCard>
      </NGi>

      <NGi :span="2">
        <NCard title="通知" :bordered="false" embedded>
          <NSpace vertical>
            <NSpace align="center">
              <NSwitch v-model:value="notifyOnComplete" />
              <NText>测试完成时桌面通知</NText>
            </NSpace>
            <NSpace align="center">
              <NSwitch v-model:value="notifyOnFailure" />
              <NText>测试失败时通知</NText>
            </NSpace>
          </NSpace>
        </NCard>
      </NGi>

      <NGi :span="2">
        <NCard title="关于" :bordered="false" embedded>
          <NSpace vertical>
            <NText>Pulse v0.1.0</NText>
            <NText depth="3">© 2024 Pulse Contributors</NText>
            <NText depth="3">Apache 2.0 License</NText>
            <NSpace>
              <NButton tag="a" href="https://github.com/pulse/pulse" target="_blank">
                🐙 GitHub
              </NButton>
              <NButton tag="a" href="https://pulse.dev" target="_blank">
                🌐 Website
              </NButton>
            </NSpace>
          </NSpace>
        </NCard>
      </NGi>
    </NGrid>
  </div>
</template>