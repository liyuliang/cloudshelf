<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import type { TransferProgress } from '../lib/api'

interface Row {
  id: string
  name: string
  percent: number
  status: string
  error: string
}

const rows = ref<Row[]>([])
let off: (() => void) | undefined

onMounted(() => {
  off = EventsOn('transfer:progress', (...args: unknown[]) => {
    // Wails 把 EventsEmit 的 optionalData 展开成回调参数；兼容单对象与包成数组的两种形式。
    const first = args.length === 1 ? args[0] : args
    const p = (Array.isArray(first) ? first[0] : first) as TransferProgress | undefined
    if (!p || !p.ID) return
    const row: Row = {
      id: p.ID,
      name: p.Name,
      percent: Math.max(0, Math.min(100, p.Percent || 0)),
      status: p.Status,
      error: p.Error || '',
    }
    const idx = rows.value.findIndex((r) => r.id === p.ID)
    if (idx < 0) rows.value.push(row)
    else rows.value[idx] = row
  })
})
onBeforeUnmount(() => off?.())

const running = computed(() => rows.value.filter((r) => r.status === 'running').length)
const finished = computed(() => rows.value.filter((r) => r.status !== 'running').length)

function clearFinished() {
  rows.value = rows.value.filter((r) => r.status === 'running')
}

function statusText(r: Row): string {
  if (r.status === 'done') return '✓ 完成'
  if (r.status === 'error') return '✗ 失败'
  return '传输中'
}
</script>

<template>
  <div v-if="rows.length" class="transfer-area">
    <div class="transfer-head">
      <span class="transfer-title">传输进度</span>
      <span v-if="running" class="badge badge-accent">{{ running }} 进行中</span>
      <span v-if="finished" class="badge badge-success">{{ finished }} 已结束</span>
      <div class="topbar-spacer"></div>
      <button class="btn btn-ghost btn-icon" title="清除已结束的任务" @click="clearFinished">清除已完成</button>
    </div>
    <div class="transfer-scroll">
      <div v-for="r in rows" :key="r.id" class="transfer-row">
        <span class="transfer-name" :title="r.name">{{ r.name }}</span>
        <span
          class="transfer-state"
          :class="{ ok: r.status === 'done', bad: r.status === 'error' }"
        >{{ statusText(r) }}</span>
        <div class="progress-track">
          <div
            class="progress-fill"
            :class="{ done: r.status === 'done', bad: r.status === 'error' }"
            :style="{ width: r.percent + '%' }"
          ></div>
        </div>
        <span class="transfer-pct mono">{{ r.percent }}%</span>
      </div>
      <div v-if="rows.some((r) => r.status === 'error' && r.error)" class="transfer-errors">
        <p v-for="r in rows.filter((x) => x.status === 'error' && x.error)" :key="r.id" class="error-text">
          {{ r.name }}：{{ r.error }}
        </p>
      </div>
    </div>
  </div>
</template>
