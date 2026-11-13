<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import Modal from '../Modal.vue'
import { api, errText, type Account } from '../../lib/api'

const props = defineProps<{
  account: Account
  // S3
  bucket?: string
  // SharePoint
  driveID?: string
  driveName?: string
  itemPath?: string
}>()
const emit = defineEmits<{ close: [] }>()

const isSP = computed(() => props.account.Provider === 'sharepoint')
const defaultName = computed(() => {
  if (isSP.value) {
    return props.itemPath ? `${props.driveName}/${props.itemPath}` : (props.driveName || '')
  }
  return props.bucket || ''
})

const name = ref(defaultName.value)
const hours = ref('24')
const password = ref(randomCode())
const busy = ref(false)
const error = ref('')

// 实时回显有效期
const expiryHint = computed(() => {
  const h = parseInt(hours.value, 10)
  if (!h || h <= 0) return '有效期至：不限期'
  const d = new Date(Date.now() + h * 3600_000)
  return `有效期至：${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
})

function pad(n: number) { return n < 10 ? '0' + n : String(n) }

function randomCode(): string {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789'
  let s = ''
  for (let i = 0; i < 8; i++) s += chars[Math.floor(Math.random() * chars.length)]
  return s
}

async function generate() {
  busy.value = true
  error.value = ''
  try {
    const h = parseInt(hours.value, 10) || 0
    let dataB64: string
    if (isSP.value) {
      dataB64 = await api.createSharePointShare(
        props.account, props.driveID!, props.driveName!, props.itemPath || '',
        name.value, h, password.value,
      )
    } else {
      dataB64 = await api.createS3Share(props.account, props.bucket!, name.value, h, password.value)
    }
    // 弹出保存对话框
    const defaultFile = isSP.value ? 'sharepoint-readonly.cloudshelf' : `${props.bucket}-readonly.cloudshelf`
    const savePath = await api.pickSaveFile('保存分享配置文件', defaultFile)
    if (!savePath) {
      busy.value = false
      return
    }
    await api.writeFileB64(savePath, dataB64)
    // 显示结果
    const pwText = password.value
      ? `分享口令：${password.value}`
      : '免口令（口令随文件保存，对方无需输入）'
    const expText = h > 0 ? `${h} 小时` : '不限期'
    alert(`已生成只读分享配置：\n${savePath}\n\n有效期：${expText}\n${pwText}`)
    emit('close')
  } catch (e) {
    error.value = errText(e)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <Modal title="生成只读分享配置" width="440px" @close="emit('close')">
    <div class="form-row">
      <label>分享名称</label>
      <input v-model="name" class="input" />
    </div>
    <div class="form-row">
      <label>有效期（小时，0 = 不限）</label>
      <input v-model="hours" class="input" type="number" min="0" />
    </div>
    <div class="form-row">
      <label>分享口令（留空 = 免口令）</label>
      <input v-model="password" class="input" placeholder="自动生成，可修改或留空" />
    </div>
    <p class="faint" style="margin: 4px 0 8px">{{ expiryHint }}</p>
    <p v-if="error" class="error-text">{{ error }}</p>
    <div class="dialog-actions">
      <button class="btn" @click="emit('close')">取消</button>
      <button class="btn btn-primary" :disabled="busy" @click="generate">
        {{ busy ? '生成中…' : '生成并保存' }}
      </button>
    </div>
  </Modal>
</template>
