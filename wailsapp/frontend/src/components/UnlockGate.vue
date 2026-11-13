<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useSession } from '../stores/session'
import { api, errText, type ShareSession } from '../lib/api'
import Modal from './Modal.vue'

const emit = defineEmits<{ 'open-proxy': []; 'share-loaded': [sess: ShareSession] }>()
const session = useSession()
const pw = ref('')
const localError = ref('')
const input = ref<HTMLInputElement | null>(null)

// 分享加载状态
const shareLoading = ref(false)
const shareNeedPw = ref(false)
const sharePw = ref('')
const shareData = ref('')
const shareError = ref('')

onMounted(() => input.value?.focus())

async function submit() {
  localError.value = ''
  if (!pw.value) {
    localError.value = '主密码不能为空'
    return
  }
  try {
    await session.unlock(pw.value)
  } catch (e) {
    localError.value = errText(e)
  }
}

async function loadShare() {
  shareError.value = ''
  shareLoading.value = true
  try {
    const path = await api.pickFile('选择分享配置文件（.cloudshelf，兼容旧版 .s3share）')
    if (!path) { shareLoading.value = false; return }
    const b64 = await api.readFileB64(path)
    shareData.value = b64
    // 先尝试免口令打开
    const needPw = await api.needSharePassword(b64)
    if (needPw) {
      shareNeedPw.value = true
      shareLoading.value = false
      return
    }
    // 免口令 → 直接打开
    const sess = await api.openShareFile(b64, '')
    shareLoading.value = false
    if (sess.Expired) {
      shareError.value = `该分享的有效期已于 ${sess.ExpiresAt} 结束，无法使用。`
      return
    }
    emit('share-loaded', sess)
  } catch (e) {
    shareError.value = errText(e)
    shareLoading.value = false
  }
}

async function submitSharePw() {
  shareError.value = ''
  if (!sharePw.value) { shareError.value = '请输入分享口令'; return }
  try {
    const sess = await api.openShareFile(shareData.value, sharePw.value)
    shareNeedPw.value = false
    if (sess.Expired) {
      shareError.value = `该分享的有效期已于 ${sess.ExpiresAt} 结束，无法使用。`
      return
    }
    emit('share-loaded', sess)
  } catch (e) {
    shareError.value = errText(e)
  }
}
</script>

<template>
  <div class="unlock-wrap">
    <div class="unlock-card card">
      <div class="unlock-mark"></div>
      <h1>CloudShelf</h1>
      <p class="muted">{{ session.initialized ? '输入主密码解锁' : '首次使用，请设置主密码' }}</p>
      <input
        ref="input"
        v-model="pw"
        class="input"
        type="password"
        placeholder="主密码"
        @keyup.enter="submit"
      />
      <p v-if="localError || session.error" class="error-text">{{ localError || session.error }}</p>
      <button class="btn btn-primary unlock-btn" :disabled="session.busy" @click="submit">
        {{ session.busy ? '解锁中…' : (session.initialized ? '解锁' : '设置并进入') }}
      </button>
      <div class="unlock-links">
        <button class="btn btn-ghost" @click="emit('open-proxy')">代理设置</button>
        <button class="btn btn-ghost" :disabled="shareLoading" @click="loadShare">
          {{ shareLoading ? '加载中…' : '加载分享配置' }}
        </button>
      </div>
      <p v-if="shareError" class="error-text" style="margin-top: 8px">{{ shareError }}</p>
    </div>

    <!-- 分享口令输入弹窗 -->
    <Modal v-if="shareNeedPw" title="加载分享配置" width="360px" @close="shareNeedPw = false">
      <div class="form-row">
        <label>分享口令</label>
        <input v-model="sharePw" class="input" type="password" placeholder="分享口令" @keyup.enter="submitSharePw" />
      </div>
      <p v-if="shareError" class="error-text">{{ shareError }}</p>
      <div class="dialog-actions">
        <button class="btn" @click="shareNeedPw = false">取消</button>
        <button class="btn btn-primary" @click="submitSharePw">确定</button>
      </div>
    </Modal>
  </div>
</template>
