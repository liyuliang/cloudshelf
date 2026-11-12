<script setup lang="ts">
withDefaults(defineProps<{ title: string; width?: string }>(), { width: '440px' })
const emit = defineEmits<{ close: [] }>()
</script>

<template>
  <div class="dialog-backdrop" @click.self="emit('close')">
    <div class="dialog" :style="{ width }">
      <h2>{{ title }}</h2>
      <slot />
    </div>
  </div>
</template>
