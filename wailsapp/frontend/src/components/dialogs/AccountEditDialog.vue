<script setup lang="ts">
import { reactive, ref } from 'vue'
import Modal from '../Modal.vue'
import { api, emptyAccount, errText, type Account } from '../../lib/api'

const props = defineProps<{ account: Account | null }>()
const emit = defineEmits<{ close: []; saved: [] }>()

const form = reactive<Account>({ ...(props.account ?? emptyAccount()) })
const saving = ref(false)
const error = ref('')

async function save() {
  error.value = ''
  if (!form.Name.trim()) {
    error.value = '账号名称不能为空'
    return
  }
  saving.value = true
  try {
    if (props.account) await api.updateAccount(form)
    else await api.addAccount(form)
    emit('saved')
  } catch (e) {
    error.value = errText(e)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Modal :title="account ? '编辑账号' : '新增账号'" width="520px" @close="emit('close')">
    <div class="form-row">
      <label>名称</label>
      <input v-model="form.Name" class="input" placeholder="账号名称" />
    </div>
    <div class="form-row">
      <label>类型</label>
      <select v-model="form.Provider" class="input">
        <option value="s3">S3</option>
        <option value="sharepoint">SharePoint</option>
        <option value="rws">Rave (RWS)</option>
      </select>
    </div>

    <template v-if="form.Provider === 'sharepoint'">
      <div class="form-row">
        <label>Tenant ID</label>
        <input v-model="form.TenantID" class="input" placeholder="Azure AD 租户 ID" />
      </div>
      <div class="form-row">
        <label>Client ID</label>
        <input v-model="form.ClientID" class="input" placeholder="应用注册的 Client ID" />
      </div>
      <div class="form-row">
        <label>Client Secret</label>
        <input v-model="form.ClientSecret" class="input" type="password" placeholder="客户端密钥" />
      </div>
      <div class="form-row">
        <label>Site URL</label>
        <input v-model="form.SiteURL" class="input" placeholder="如 https://contoso.sharepoint.com/sites/TeamA" />
      </div>
    </template>

    <template v-else-if="form.Provider === 'rws'">
      <div class="form-row">
        <label>Domain</label>
        <input v-model="form.Region" class="input" placeholder="如 janussandbox、innovate、rave" />
      </div>
      <div class="form-row">
        <label>Username</label>
        <input v-model="form.AccessKeyID" class="input" placeholder="RWS 用户名" />
      </div>
      <div class="form-row">
        <label>Password</label>
        <input v-model="form.SecretAccessKey" class="input" type="password" placeholder="RWS 密码" />
      </div>
      <div class="form-row">
        <label>Host（可选）</label>
        <input v-model="form.Endpoint" class="input" placeholder="留空使用默认 https://%s.mdsol.com" />
      </div>
      <div class="form-row">
        <label>Virtual Dir（可选）</label>
        <input v-model="form.SiteURL" class="input" placeholder="留空使用默认 RaveWebServices" />
      </div>
    </template>

    <template v-else>
      <div class="form-row">
        <label>Endpoint</label>
        <input v-model="form.Endpoint" class="input" placeholder="留空则使用 AWS 默认端点" />
      </div>
      <div class="form-row">
        <label>Region</label>
        <input v-model="form.Region" class="input" placeholder="如 us-east-1" />
      </div>
      <div class="form-row">
        <label>Access Key ID</label>
        <input v-model="form.AccessKeyID" class="input" />
      </div>
      <div class="form-row">
        <label>Secret Access Key</label>
        <input v-model="form.SecretAccessKey" class="input" type="password" />
      </div>
      <div class="form-row">
        <label>Bucket（可选）</label>
        <input v-model="form.Bucket" class="input" placeholder="留空列出全部 Bucket；填写则只显示指定桶" />
      </div>
      <label class="checkbox-row">
        <input v-model="form.UsePathStyle" type="checkbox" />
        使用 Path-Style（兼容 MinIO 等）
      </label>
    </template>

    <p v-if="error" class="error-text">{{ error }}</p>
    <div class="dialog-actions">
      <button class="btn" @click="emit('close')">取消</button>
      <button class="btn btn-primary" :disabled="saving" @click="save">
        {{ saving ? '保存中…' : '保存' }}
      </button>
    </div>
  </Modal>
</template>
