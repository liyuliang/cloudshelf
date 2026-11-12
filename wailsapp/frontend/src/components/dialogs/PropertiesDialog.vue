<script setup lang="ts">
import Modal from '../Modal.vue'
import { api } from '../../lib/api'

const props = defineProps<{ title: string; rows: [string, string][] }>()
const emit = defineEmits<{ close: [] }>()

async function copyValue(value: string) {
  await api.copyText(value)
}

async function copyAll() {
  await api.copyText(props.rows.map(([k, v]) => `${k}: ${v}`).join('\n'))
}
</script>

<template>
  <Modal :title="title" width="440px" @close="emit('close')">
    <div class="prop-table">
      <div v-for="[k, v] in rows" :key="k" class="prop-row">
        <span class="prop-key">{{ k }}</span>
        <span class="prop-value mono" title="点击复制" @click="copyValue(v)">{{ v || '—' }}</span>
      </div>
    </div>
    <p class="faint hint">点击属性值可复制，或用下方按钮复制全部。</p>
    <div class="dialog-actions">
      <button class="btn" @click="copyAll">复制全部</button>
      <button class="btn btn-primary" @click="emit('close')">关闭</button>
    </div>
  </Modal>
</template>
