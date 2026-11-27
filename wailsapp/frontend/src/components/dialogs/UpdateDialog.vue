<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import Modal from '../Modal.vue'
import { api, errText, type ReleaseInfo, type UpdateProgress } from '../../lib/api'
import { EventsOn } from '../../../wailsjs/runtime/runtime'

const props = defineProps<{
  info: ReleaseInfo | null
  checking?: boolean
}>()

const emit = defineEmits<{
  close: []
  notify: [msg: string]
}>()

const updating = ref(false)
const progress = ref<UpdateProgress>({ done: 0, total: 0, percent: 0 })
const errorMsg = ref('')
let offProgress: (() => void) | undefined

onMounted(() => {
  offProgress = EventsOn('update:progress', (p: UpdateProgress) => {
    progress.value = p
  })
})

onUnmounted(() => {
  offProgress?.()
})

function formatBytes(n: number): string {
  if (!n || n <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let v = n
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024
    i++
  }
  return `${v.toFixed(i === 0 ? 0 : 1)} ${units[i]}`
}

async function startUpdate() {
  if (!props.info?.assetUrl || updating.value) return
  updating.value = true
  errorMsg.value = ''
  progress.value = { done: 0, total: props.info.assetSize || 0, percent: 0 }
  try {
    await api.applyUpdate(props.info.assetUrl, props.info.assetName)
  } catch (e) {
    updating.value = false
    errorMsg.value = errText(e)
  }
}

async function openRelease() {
  await api.openReleasePage(props.info?.htmlUrl || '')
}
</script>

<template>
  <Modal title="检查更新" width="480px" :show-close="!updating" @close="emit('close')">
    <div v-if="checking" class="update-box">
      <p>正在检查 GitHub Release 最新版本…</p>
    </div>

    <div v-else-if="info" class="update-box">
      <div class="version-row">
        <span>当前版本：<strong>{{ info.currentVersion || 'dev' }}</strong></span>
        <span v-if="info.latestVersion">最新版本：<strong>{{ info.latestVersion }}</strong></span>
      </div>

      <p v-if="!info.latestVersion" class="muted">
        仓库暂未发布正式 Release 版本。
      </p>
      <p v-else-if="!info.hasUpdate && !info.isDev" class="success-text">
        当前已是最新版本。
      </p>
      <p v-else-if="info.isDev" class="muted">
        当前为开发构建（{{ info.currentVersion }}），GitHub 最新正式版为 {{ info.latestVersion }}。
      </p>
      <p v-else class="highlight-text">
        发现新版本 {{ info.latestVersion }}，建议更新！
      </p>

      <div v-if="info.releaseNotes" class="release-notes">
        <div class="notes-title">更新说明</div>
        <pre class="notes-body">{{ info.releaseNotes }}</pre>
      </div>

      <div v-if="updating" class="progress-wrap">
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: `${progress.percent}%` }"></div>
        </div>
        <div class="progress-text">
          正在下载更新包… {{ progress.percent }}%（{{ formatBytes(progress.done) }} / {{ formatBytes(progress.total) }}）
        </div>
      </div>

      <p v-if="errorMsg" class="error-text">{{ errorMsg }}</p>

      <div class="dialog-actions">
        <button class="btn" :disabled="updating" @click="emit('close')">
          {{ info.hasUpdate ? '暂不更新' : '关闭' }}
        </button>
        <button v-if="info.htmlUrl" class="btn" :disabled="updating" @click="openRelease">
          前往 GitHub 下载
        </button>
        <button
          v-if="info.canSelfUpdate && (info.hasUpdate || info.isDev)"
          class="btn btn-primary"
          :disabled="updating"
          @click="startUpdate"
        >
          {{ updating ? '正在更新…' : '立即更新并重启' }}
        </button>
      </div>
    </div>
  </Modal>
</template>

<style scoped>
.update-box {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.version-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 12px;
  background: var(--bg-elev, #1c2030);
  border-radius: 6px;
  font-size: 13px;
}
.muted {
  color: var(--text-dim, #99a1b7);
  font-size: 13px;
  margin: 0;
}
.success-text {
  color: #3fb950;
  font-size: 13px;
  margin: 0;
}
.highlight-text {
  color: #5b8cff;
  font-weight: 500;
  font-size: 13px;
  margin: 0;
}
.error-text {
  color: #f85149;
  font-size: 12px;
  margin: 0;
}
.release-notes {
  border: 1px solid var(--border, #272c3c);
  border-radius: 6px;
  padding: 8px 10px;
  background: var(--bg-base, #0f1117);
}
.notes-title {
  font-size: 12px;
  color: var(--text-dim, #99a1b7);
  margin-bottom: 4px;
}
.notes-body {
  margin: 0;
  max-height: 160px;
  overflow: auto;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: inherit;
}
.progress-wrap {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.progress-bar {
  height: 8px;
  background: var(--bg-elev, #1c2030);
  border-radius: 4px;
  overflow: hidden;
}
.progress-fill {
  height: 100%;
  background: #5b8cff;
  transition: width 0.15s ease;
}
.progress-text {
  font-size: 12px;
  color: var(--text-dim, #99a1b7);
}
</style>
