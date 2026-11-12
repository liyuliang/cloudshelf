<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import Modal from '../Modal.vue'
import { api, errText, type Account } from '../../lib/api'

const props = defineProps<{ account: Account; bucket: string; objectKey: string }>()
const emit = defineEmits<{ close: [] }>()

const minutes = ref(10)
const url = ref('')
const error = ref('')
const busy = ref(false)
const copied = ref(false)

async function regenerate() {
  busy.value = true
  error.value = ''
  copied.value = false
  try {
    url.value = await api.presignURL(props.account, props.bucket, props.objectKey, Number(minutes.value) || 10)
  } catch (e) {
    error.value = errText(e)
  } finally {
    busy.value = false
  }
}

watch(minutes, () => void regenerate())
onMounted(() => void regenerate())

async function copy() {
  if (!url.value) return
  await api.copyText(url.value)
  copied.value = true
}
</script>

<template>
  <Modal title="创建预签名 Web URL" width="560px" @close="emit('close')">
    <div class="form-row">
      <label>对象 Key</label>
      <div class="readonly-path mono">{{ objectKey }}</div>
    </div>
    <div class="form-row">
      <label>有效期（分钟，默认 10）</label>
      <input v-model.number="minutes" class="input" type="number" min="1" />
    </div>
    <div class="form-row">
      <label>URL（随输入实时刷新）</label>
      <textarea class="input mono" rows="4" readonly :value="url"></textarea>
    </div>
    <p v-if="error" class="error-text">{{ error }}</p>
    <p v-if="copied" class="success-text">已复制到剪贴板</p>
    <div class="dialog-actions">
      <button class="btn" :disabled="!url" @click="copy">复制 URL</button>
      <button class="btn btn-primary" :disabled="busy" @click="emit('close')">关闭</button>
    </div>
  </Modal>
</template>
