import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import i18n from './i18n'

// Naive UI
import naive from 'naive-ui'

// 样式
import './styles/global.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(i18n)
app.use(naive)

// 监听后端事件(由 Wails 自动生成 wailsjs)
// 这个 hooks 是从 Wails 自动生成的 /wailsjs/runtime/runtime.js
import { EventsOn } from '../wailsjs/runtime/runtime'

EventsOn('app:ready', (data: any) => {
  console.log('[Pulse] App ready:', data)
})

app.mount('#app')