<script setup lang="ts">
import { useBrowser, type BrowserTab } from '../stores/browser'
import type { Bucket } from '../lib/api'

const props = defineProps<{ tab: BrowserTab }>()
const emit = defineEmits<{ properties: [title: string, rows: [string, string][]] }>()
const browser = useBrowser()

function open(b: Bucket) {
  if (!b.Accessible) return
  void browser.openBucket(props.tab, b.Name)
}

function showProps(b: Bucket) {
  emit('properties', 'Bucket 属性', [
    ['名称', b.Name],
    ['访问权限', b.Accessible ? '可访问' : '无权限'],
  ])
}
</script>

<template>
  <div class="bucket-view">
    <div class="content-head">
      <div>
        <div class="content-title">{{ tab.title }} — Bucket 列表</div>
        <div class="faint">共 {{ tab.buckets.length }} 个 Bucket，双击可访问的桶进入</div>
      </div>
      <button class="btn" :disabled="tab.loading" @click="browser.loadBuckets(tab)">刷新</button>
    </div>

    <div class="table-wrap">
      <div class="obj-grid bucket-grid table-header">
        <span>名称</span><span>权限</span>
      </div>
      <div class="table-scroll">
        <div
          v-for="b in tab.buckets"
          :key="b.Name"
          class="obj-grid bucket-grid table-row"
          :class="{ disabled: !b.Accessible }"
          @dblclick="open(b)"
          @contextmenu.prevent="showProps(b)"
        >
          <span class="cell-main"><span class="row-icon">🪣</span>{{ b.Name }}</span>
          <span :class="b.Accessible ? 'perm-ok' : 'perm-no'">
            {{ b.Accessible ? '✓ 可访问' : '✗ 无权限' }}
          </span>
        </div>
        <p v-if="!tab.loading && tab.buckets.length === 0" class="empty-hint faint">没有 Bucket</p>
        <p v-if="tab.loading" class="empty-hint faint">正在加载…</p>
      </div>
    </div>

    <div v-if="tab.error" class="msg-area error-text">⚠ {{ tab.error }}</div>
  </div>
</template>
