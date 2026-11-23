<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { EventsOn } from '../../wailsjs/runtime/runtime'

interface LogEntry {
  id: number
  time: string
  method: string
  url: string
  status: number
  duration: number
  error: string
  kind: string // 'probe' = 权限探测，失败是预期结果
}

const entries = ref<LogEntry[]>([])
const expanded = ref(false)
const autoScroll = ref(true)
const scrollEl = ref<HTMLElement | null>(null)
let off: (() => void) | undefined

const MAX_ENTRIES = 500

onMounted(() => {
  off = EventsOn('request:log', (...args: unknown[]) => {
    const first = args.length === 1 ? args[0] : args
    const e = (Array.isArray(first) ? first[0] : first) as LogEntry | undefined
    if (!e || !e.id) return
    entries.value.push(e)
    if (entries.value.length > MAX_ENTRIES) {
      entries.value = entries.value.slice(-MAX_ENTRIES)
    }
    if (autoScroll.value) {
      nextTick(() => {
        const el = scrollEl.value
        if (el) el.scrollTop = el.scrollHeight
      })
    }
  })
})
onBeforeUnmount(() => off?.())

const count = computed(() => entries.value.length)
// 权限探测的 4xx 是预期结果，不计入错误数；网络错误仍然是真实故障。
const errorCount = computed(
  () => entries.value.filter(e => e.error || (e.status >= 400 && e.kind !== 'probe')).length,
)
const probeCount = computed(() => entries.value.filter(e => e.kind === 'probe').length)

function statusClass(e: LogEntry): string {
  if (e.error) return 'log-err'
  if (e.kind === 'probe' && e.status >= 400) return 'log-probe'
  if (e.status >= 400) return 'log-err'
  if (e.status >= 300) return 'log-warn'
  return 'log-ok'
}

function statusText(e: LogEntry): string {
  if (e.error) return 'ERR'
  return String(e.status || '---')
}

// 请求失败时 method 徽标不再上色，避免 403 旁边显示绿色 GET 造成误解。
// 探测请求的 4xx 是预期结果，仍保留方法颜色。
function methodClass(e: LogEntry): string {
  if (e.error || (e.status >= 400 && e.kind !== 'probe')) return ''
  switch (e.method) {
    case 'GET': return 'method-get'
    case 'PUT': case 'POST': return 'method-put'
    case 'DELETE': return 'method-del'
    case 'HEAD': return 'method-head'
    default: return ''
  }
}

function shortUrl(url: string): string {
  try {
    const u = new URL(url)
    // 只显示 host + path（去掉签名等长查询参数）
    let display = u.host + u.pathname
    if (display.length > 120) display = display.slice(0, 117) + '…'
    return display
  } catch {
    return url.length > 120 ? url.slice(0, 117) + '…' : url
  }
}

function clear() {
  entries.value = []
}
</script>

<template>
  <div class="log-panel" :class="{ expanded }">
    <div class="log-head" @click="expanded = !expanded">
      <span class="log-toggle">{{ expanded ? '▼' : '▶' }}</span>
      <span class="log-title">请求日志</span>
      <span v-if="count" class="badge">{{ count }}</span>
      <span v-if="errorCount" class="badge badge-err">{{ errorCount }} 错误</span>
      <span v-if="probeCount" class="badge badge-probe" title="权限探测：403 等否定结果属于预期">{{ probeCount }} 探测</span>
      <div class="topbar-spacer"></div>
      <label v-if="expanded" class="log-auto-scroll" @click.stop>
        <input type="checkbox" v-model="autoScroll" /> 自动滚动
      </label>
      <button v-if="expanded && count" class="btn btn-ghost btn-icon log-clear-btn" title="清空日志" @click.stop="clear">清空</button>
    </div>
    <div v-if="expanded" ref="scrollEl" class="log-scroll">
      <div v-if="!entries.length" class="log-empty faint">暂无请求记录</div>
      <div v-for="e in entries" :key="e.id" class="log-row" :class="statusClass(e)">
        <span class="log-time mono">{{ e.time }}</span>
        <span class="log-method mono" :class="methodClass(e)">{{ e.method }}</span>
        <span class="log-status mono" :class="statusClass(e)">{{ statusText(e) }}</span>
        <span v-if="e.kind === 'probe'" class="log-tag" title="权限探测：403 等否定结果属于预期">探测</span>
        <span class="log-dur mono">{{ e.duration }}ms</span>
        <span class="log-url" :title="e.url">{{ shortUrl(e.url) }}</span>
        <span v-if="e.error" class="log-error" :title="e.error">{{ e.error }}</span>
      </div>
    </div>
  </div>
</template>
