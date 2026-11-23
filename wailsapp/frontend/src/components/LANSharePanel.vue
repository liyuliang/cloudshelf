<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed } from 'vue'
import { api, errText, type LANShare } from '../lib/api'

const emit = defineEmits<{ notify: [msg: string] }>()

const shares = ref<LANShare[]>([])
const expanded = ref(false)
let refreshTimer: number | undefined

onMounted(async () => {
  await loadShares()
  refreshTimer = window.setInterval(loadShares, 30_000)
})

onUnmounted(() => {
  if (refreshTimer) window.clearInterval(refreshTimer)
})

async function loadShares() {
  try {
    shares.value = (await api.listLANShares()) || []
  } catch {
    // ignore
  }
}

const count = computed(() => shares.value.length)
const expiredCount = computed(
  () => shares.value.filter((s) => s.remaining === '已过期').length,
)

async function copyURL(url: string) {
  try {
    await api.copyToClipboard(url)
    emit('notify', '链接已复制')
  } catch {
    // fallback
  }
}

async function remove(token: string) {
  try {
    await api.removeLANShare(token)
    await loadShares()
    emit('notify', '已停止分享')
  } catch (e) {
    emit('notify', '删除失败: ' + errText(e))
  }
}

async function removeAll() {
  for (const s of shares.value) {
    try {
      await api.removeLANShare(s.token)
    } catch {
      // ignore individual failures
    }
  }
  await loadShares()
  emit('notify', '已停止所有分享')
}
</script>

<template>
  <div class="lanpanel" :class="{ expanded }">
    <div class="lanpanel-head" @click="expanded = !expanded">
      <span class="lanpanel-toggle">{{ expanded ? '▼' : '▶' }}</span>
      <span class="lanpanel-title">局域网分享</span>
      <span v-if="count" class="badge badge-accent">{{ count }} 活跃</span>
      <div class="topbar-spacer"></div>
      <button
        v-if="expanded && count > 1"
        class="btn btn-ghost btn-icon lanpanel-clear-btn"
        title="停止所有分享"
        @click.stop="removeAll"
      >
        全部停止
      </button>
    </div>
    <div v-if="expanded" class="lanpanel-scroll">
      <div v-if="!shares.length" class="lanpanel-empty faint">
        暂无活跃的局域网分享
      </div>
      <div v-for="s in shares" :key="s.token" class="lanpanel-card">
        <div class="lanpanel-card-head">
          <span class="lanpanel-card-name">{{ s.name }}</span>
          <span
            class="badge"
            :class="
              s.remaining === '已过期' ? 'badge-err' : 'badge-accent'
            "
          >
            {{ s.remaining }}
          </span>
        </div>
        <div class="lanpanel-card-meta">
          <span class="faint">{{ s.bucket }}{{ s.prefix ? '/' + s.prefix : '' }}</span>
          <span class="faint">创建于 {{ s.createdAt }}</span>
        </div>
        <div class="lanpanel-card-urls">
          <div v-for="url in s.urls" :key="url" class="lanpanel-url-row">
            <code class="lanpanel-url-text">{{ url }}</code>
            <button
              class="btn btn-ghost lanpanel-copy-btn"
              @click="copyURL(url)"
              title="复制链接"
            >
              📋
            </button>
          </div>
        </div>
        <div class="lanpanel-card-actions">
          <button class="btn btn-danger btn-sm" @click="remove(s.token)">
            停止分享
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
