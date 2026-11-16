<script setup lang="ts">
import type { BrowserTab } from '../stores/browser'
import { useBrowser } from '../stores/browser'

const props = defineProps<{ tab: BrowserTab }>()
const browser = useBrowser()
</script>

<template>
  <div class="rws-study-list">
    <div class="browser-toolbar">
      <span class="faint">Study 列表</span>
      <button class="btn btn-sm" @click="browser.refresh(props.tab)">刷新</button>
    </div>
    <p v-if="tab.loading" class="faint center-hint">加载中…</p>
    <p v-else-if="tab.error" class="error-text center-hint">{{ tab.error }}</p>
    <p v-else-if="tab.rwsStudies.length === 0" class="faint center-hint">未找到任何 Study</p>
    <div v-else class="rws-table-wrap">
      <table class="rws-table">
        <thead>
          <tr>
            <th>Study OID</th>
            <th>名称</th>
            <th>描述</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="s in tab.rwsStudies"
            :key="s.OID"
            class="rws-row"
            @dblclick="browser.loadRWSStudyDetail(tab, s.OID)"
          >
            <td class="rws-oid">{{ s.OID }}</td>
            <td>{{ s.Name }}</td>
            <td class="faint">{{ s.Description }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
