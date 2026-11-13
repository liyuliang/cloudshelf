<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import Modal from '../Modal.vue'
import { api, emptyProxy, errText, type ProxyConfig, type ProxyMode } from '../../lib/api'

const emit = defineEmits<{ close: [] }>()
const modes = ref<ProxyMode[]>([])
const form = reactive<ProxyConfig>(emptyProxy())
const testURL = ref('https://s3.amazonaws.com/')
const testing = ref(false)
const saving = ref(false)
const error = ref('')
const notice = ref('')

const custom = computed(() => form.Mode === 'http' || form.Mode === 'socks5')

onMounted(async () => {
  try {
    modes.value = await api.proxyModes()
    Object.assign(form, await api.getProxy())
  } catch (e) {
    error.value = errText(e)
  }
})

function normalize() {
  form.Port = Number(form.Port) || 0
}

async function testConn() {
  error.value = ''
  notice.value = ''
  if (!testURL.value.trim()) {
    error.value = '请填写测试地址'
    return
  }
  normalize()
  try {
    await api.validateProxy(form)
  } catch (e) {
    error.value = errText(e)
    return
  }
  testing.value = true
  try {
    notice.value = await api.checkProxy(form, testURL.value.trim(), 12)
  } catch (e) {
    error.value = errText(e)
  } finally {
    testing.value = false
  }
}

async function save() {
  error.value = ''
  notice.value = ''
  normalize()
  saving.value = true
  try {
    await api.validateProxy(form)
    await api.setProxy(form)
    emit('close')
  } catch (e) {
    error.value = errText(e)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Modal title="代理设置" width="520px" @close="emit('close')">
    <div class="form-row">
      <label>模式</label>
      <select v-model="form.Mode" class="input">
        <option v-for="m in modes" :key="m.Mode" :value="m.Mode">{{ m.Label }}</option>
      </select>
    </div>
    <div class="form-row">
      <label>服务器</label>
      <input v-model="form.Host" class="input" :disabled="!custom" placeholder="如 127.0.0.1，或粘贴完整代理 URL" />
    </div>
    <div class="form-row">
      <label>端口</label>
      <input v-model.number="form.Port" class="input" type="number" :disabled="!custom" placeholder="如 7890" />
    </div>
    <div class="form-row">
      <label>用户名</label>
      <input v-model="form.Username" class="input" :disabled="!custom" />
    </div>
    <div class="form-row">
      <label>密码</label>
      <input v-model="form.Password" class="input" type="password" :disabled="!custom" />
    </div>
    <div class="form-row">
      <label>例外（NO_PROXY）</label>
      <input v-model="form.NoProxy" class="input" :disabled="!custom" placeholder="如 localhost,127.0.0.1,.internal" />
    </div>
    <div class="form-row">
      <label>测试地址</label>
      <div class="inline-row">
        <input v-model="testURL" class="input" placeholder="如 https://s3.amazonaws.com/" />
        <button class="btn" :disabled="testing" @click="testConn">
          {{ testing ? '测试中…' : '测试连接' }}
        </button>
      </div>
    </div>

    <p class="faint hint">配置保存在 ~/.cloudshelf/config.json，GUI 与 CLI 共用；保存后立即生效，进行中的传输不受影响。</p>
    <p v-if="error" class="error-text">{{ error }}</p>
    <p v-if="notice" class="success-text">{{ notice }}</p>
    <div class="dialog-actions">
      <button class="btn" @click="emit('close')">取消</button>
      <button class="btn btn-primary" :disabled="saving" @click="save">
        {{ saving ? '保存中…' : '保存' }}
      </button>
    </div>
  </Modal>
</template>
