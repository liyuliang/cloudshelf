<script setup lang="ts">
import { onMounted, ref } from 'vue'
import Modal from '../Modal.vue'
import { api, errText } from '../../lib/api'

type Kind = 'change-pw' | 'change-db' | 'export' | 'import'
const props = defineProps<{ kind: Kind }>()
const emit = defineEmits<{ close: []; relock: []; changed: [] }>()

const meta: Record<Kind, { title: string; ok: string }> = {
  'change-pw': { title: '修改主密码', ok: '确定' },
  'change-db': { title: '修改存储位置', ok: '选择位置并移动' },
  export: { title: '导出账号', ok: '导出' },
  import: { title: '导入账号', ok: '选择文件并导入' },
}

const pw = ref('')
const pw2 = ref('')
const curPath = ref('')
const busy = ref(false)
const error = ref('')
const notice = ref('')

onMounted(async () => {
  if (props.kind === 'change-db') {
    try {
      curPath.value = await api.currentDBPath()
    } catch (e) {
      error.value = errText(e)
    }
  }
})

async function run() {
  error.value = ''
  notice.value = ''
  busy.value = true
  try {
    switch (props.kind) {
      case 'change-pw': {
        if (!pw.value) throw new Error('新主密码不能为空')
        if (pw.value !== pw2.value) throw new Error('两次输入的主密码不一致')
        await api.changeMasterPassword(pw.value)
        notice.value = '主密码已更新'
        break
      }
      case 'change-db': {
        const p = await api.pickSaveFile('选择新的数据库文件位置', curPath.value)
        if (!p) return
        await api.changeDBLocation(p)
        emit('relock')
        emit('close')
        break
      }
      case 'export': {
        if (!pw.value) throw new Error('请输入主密码')
        const b64 = await api.exportAccounts(pw.value)
        const p = await api.pickSaveFile('保存加密的账号导出文件', 'cloudshelf-accounts.enc')
        if (!p) return
        await api.writeFileB64(p, b64)
        notice.value = '账号已加密导出'
        break
      }
      case 'import': {
        if (!pw.value) throw new Error('请输入主密码')
        const p = await api.pickFile('选择账号导出文件')
        if (!p) return
        const b64 = await api.readFileB64(p)
        const n = await api.importAccounts(pw.value, b64)
        emit('changed')
        notice.value = `成功导入 ${n} 个账号`
        break
      }
    }
  } catch (e) {
    error.value = errText(e)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <Modal :title="meta[kind].title" width="440px" @close="emit('close')">
    <template v-if="kind === 'change-pw'">
      <div class="form-row">
        <label>新主密码</label>
        <input v-model="pw" class="input" type="password" placeholder="新主密码" />
      </div>
      <div class="form-row">
        <label>确认新主密码</label>
        <input v-model="pw2" class="input" type="password" placeholder="再次输入" />
      </div>
    </template>

    <template v-else-if="kind === 'change-db'">
      <div class="form-row">
        <label>当前位置</label>
        <div class="readonly-path mono">{{ curPath || '（读取中…）' }}</div>
      </div>
      <p class="faint hint">
        将把数据库移动到所选位置，之后需要重新输入主密码解锁。此操作会关闭当前账号库。
      </p>
    </template>

    <template v-else>
      <div class="form-row">
        <label>主密码</label>
        <input v-model="pw" class="input" type="password" placeholder="主密码" />
      </div>
      <p class="faint hint">
        {{ kind === 'export' ? '导出文件使用 Argon2id + AES-256-GCM 加密。' : '选择此前导出的 .enc 文件，导入后账号列表会刷新。' }}
      </p>
    </template>

    <p v-if="error" class="error-text">{{ error }}</p>
    <p v-if="notice" class="success-text">{{ notice }}</p>
    <div class="dialog-actions">
      <button class="btn" @click="emit('close')">{{ notice ? '关闭' : '取消' }}</button>
      <button v-if="!notice" class="btn btn-primary" :disabled="busy" @click="run">
        {{ busy ? '处理中…' : meta[kind].ok }}
      </button>
    </div>
  </Modal>
</template>
