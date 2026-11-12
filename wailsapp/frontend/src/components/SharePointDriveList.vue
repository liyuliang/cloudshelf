<script setup lang="ts">
import { useBrowser, type BrowserTab } from '../stores/browser'
import type { Drive } from '../lib/api'

const props = defineProps<{ tab: BrowserTab }>()
const emit = defineEmits<{ properties: [title: string, rows: [string, string][]] }>()
const browser = useBrowser()

function open(d: Drive) {
  void browser.openDrive(props.tab, d.ID, d.Name)
}

function showProps(d: Drive) {
  emit('properties', '文档库属性', [
    ['名称', d.Name],
    ['ID', d.ID],
  ])
}
</script>

<template>
  <div class="bucket-view">
    <div class="content-head">
      <div>
        <div class="content-title">{{ tab.title }} — 文档库列表</div>
        <div class="faint">共 {{ tab.drives.length }} 个文档库，双击进入</div>
      </div>
      <button class="btn" :disabled="tab.loading" @click="browser.loadDrives(tab)">刷新</button>
    </div>

    <div class="table-wrap">
      <div class="obj-grid bucket-grid table-header">
        <span>名称</span><span>操作</span>
      </div>
      <div class="table-scroll">
        <div
          v-for="d in tab.drives"
          :key="d.ID"
          class="obj-grid bucket-grid table-row"
          @dblclick="open(d)"
          @contextmenu.prevent="showProps(d)"
        >
          <span class="cell-main"><span class="row-icon">📚</span>{{ d.Name }}</span>
          <span class="perm-ok">双击进入</span>
        </div>
        <p v-if="!tab.loading && tab.drives.length === 0" class="empty-hint faint">没有文档库</p>
        <p v-if="tab.loading" class="empty-hint faint">正在加载…</p>
      </div>
    </div>

    <div v-if="tab.error" class="msg-area error-text">⚠ {{ tab.error }}</div>
  </div>
</template>
