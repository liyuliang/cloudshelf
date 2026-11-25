<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'

const props = withDefaults(
  defineProps<{
    title: string
    width?: string
    closeOnBackdrop?: boolean
    showClose?: boolean
  }>(),
  {
    width: '440px',
    closeOnBackdrop: false,
    showClose: true,
  },
)

const emit = defineEmits<{ close: [] }>()

function onBackdropClick() {
  if (props.closeOnBackdrop) {
    emit('close')
  }
}

function onKeyDown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    emit('close')
  }
}

onMounted(() => {
  window.addEventListener('keydown', onKeyDown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeyDown)
})
</script>

<template>
  <div class="dialog-backdrop" @click.self="onBackdropClick">
    <div class="dialog" :style="{ width }">
      <div class="dialog-header">
        <h2>{{ title }}</h2>
        <button v-if="showClose" class="dialog-close-btn" title="关闭" @click="emit('close')">✕</button>
      </div>
      <slot />
    </div>
  </div>
</template>
