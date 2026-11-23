<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import Modal from '../Modal.vue'
import { api, errText, type Account, type LANShare } from '../../lib/api'

const props = defineProps<{
  account: Account
  bucket: string
  prefix: string
}>()
const emit = defineEmits<{ close: []; notify: [msg: string] }>()

const name = ref(props.prefix
  ? `${props.bucket}/${props.prefix.replace(/\/$/, '')}`
  : props.bucket)
const hours = ref(4)
const busy = ref(false)
const error = ref('')
const shares = ref<LANShare[]>([])
const tab = ref<'create' | 'manage'>('create')
const firewallShown = ref(false)

let refreshTimer: number | undefined

onMounted(async () => {
  await loadShares()
  // Auto-refresh every 30s to update remaining time
  refreshTimer = window.setInterval(loadShares, 30_000)
})

onUnmounted(() => {
  if (refreshTimer) window.clearInterval(refreshTimer)
})

async function loadShares() {
  try {
    shares.value = await api.listLANShares() || []
  } catch {
    // ignore
  }
}

async function create() {
  busy.value = true
  error.value = ''
  try {
    const share = await api.createLANShare(props.account, props.bucket, props.prefix, name.value, hours.value)
    if (!firewallShown.value) {
      firewallShown.value = true
    }
    await loadShares()
    tab.value = 'manage'
    // Auto-copy first URL
    if (share.url) {
      await api.copyToClipboard(share.url)
      emit('notify', `分享链接已复制到剪贴板`)
    }
  } catch (e) {
    error.value = errText(e)
  } finally {
    busy.value = false
  }
}

async function remove(token: string) {
  try {
    await api.removeLANShare(token)
    await loadShares()
  } catch (e) {
    emit('notify', '删除失败: ' + errText(e))
  }
}

async function copyURL(url: string) {
  try {
    await api.copyToClipboard(url)
    emit('notify', '链接已复制')
  } catch {
    // fallback
  }
}
</script>

<template>
  <Modal title="局域网分享" width="560px" @close="emit('close')">
    <!-- Tab switcher -->
    <div class="lan-tabs">
      <button class="lan-tab" :class="{ active: tab === 'create' }" @click="tab = 'create'">
        创建分享
      </button>
      <button class="lan-tab" :class="{ active: tab === 'manage' }" @click="tab = 'manage'">
        管理分享
        <span v-if="shares.length" class="badge badge-accent">{{ shares.length }}</span>
      </button>
    </div>

    <!-- Create tab -->
    <div v-if="tab === 'create'" class="lan-create">
      <div class="lan-info-card">
        <div class="lan-info-row">
          <span class="lan-info-label">账号</span>
          <span>{{ account.Name }}</span>
        </div>
        <div class="lan-info-row">
          <span class="lan-info-label">Bucket</span>
          <span class="mono">{{ bucket }}</span>
        </div>
        <div v-if="prefix" class="lan-info-row">
          <span class="lan-info-label">目录</span>
          <span class="mono">{{ prefix }}</span>
        </div>
      </div>

      <div class="form-row">
        <label>分享名称</label>
        <input v-model="name" class="input" placeholder="给分享起个名字" />
      </div>
      <div class="form-row">
        <label>有效期（小时）</label>
        <input v-model.number="hours" class="input" type="number" min="1" max="168" />
        <span class="hint faint">默认 4 小时，最长 168 小时（7 天）</span>
      </div>

      <div class="lan-firewall-hint">
        ⚠ 提示：首次使用局域网分享时，系统可能弹出防火墙对话框，请选择「允许」以便同局域网设备访问。
      </div>

      <p v-if="error" class="error-text">{{ error }}</p>
      <div class="dialog-actions">
        <button class="btn" @click="emit('close')">取消</button>
        <button class="btn btn-primary" :disabled="busy || !name.trim()" @click="create">
          {{ busy ? '创建中…' : '创建分享' }}
        </button>
      </div>
    </div>

    <!-- Manage tab -->
    <div v-if="tab === 'manage'" class="lan-manage">
      <div v-if="shares.length === 0" class="empty-hint faint">
        暂无活跃的局域网分享
      </div>
      <div v-for="s in shares" :key="s.token" class="lan-share-card">
        <div class="lan-share-head">
          <span class="lan-share-name">{{ s.name }}</span>
          <span class="badge" :class="s.remaining === '已过期' ? 'badge-err' : 'badge-accent'">
            {{ s.remaining }}
          </span>
        </div>
        <div class="lan-share-meta">
          <span class="faint">{{ s.bucket }}{{ s.prefix ? '/' + s.prefix : '' }}</span>
          <span class="faint">创建于 {{ s.createdAt }}</span>
        </div>
        <div class="lan-share-urls">
          <div v-for="url in s.urls" :key="url" class="lan-url-row">
            <code class="lan-url-text">{{ url }}</code>
            <button class="btn btn-ghost lan-copy-btn" @click="copyURL(url)" title="复制链接">📋</button>
          </div>
        </div>
        <div class="lan-share-actions">
          <button class="btn btn-danger btn-sm" @click="remove(s.token)">停止分享</button>
        </div>
      </div>

      <div class="dialog-actions">
        <button class="btn" @click="emit('close')">关闭</button>
      </div>
    </div>
  </Modal>
</template>
