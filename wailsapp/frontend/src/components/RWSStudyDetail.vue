<script setup lang="ts">
import { ref, computed } from 'vue'
import type { BrowserTab } from '../stores/browser'
import { useBrowser } from '../stores/browser'

const props = defineProps<{ tab: BrowserTab }>()
const browser = useBrowser()

const activeSection = ref<'subjects' | 'sites' | 'datasets'>('subjects')

/** 从合并类型字符串（如 "raw+regular"）取第一个类型作为 CSV 请求提示 */
function firstType(typ: string | undefined): string {
  if (!typ) return ''
  return typ.split('+')[0]
}

/** 为类型标签生成 CSS class（支持 raw+regular 组合） */
function typeBadgeClass(typ: string | undefined): string {
  if (!typ) return 'rws-type-unknown'
  if (typ.includes('+')) return 'rws-type-both'
  return 'rws-type-' + typ
}
</script>

<template>
  <div class="rws-detail">
    <div class="browser-toolbar">
      <button class="btn btn-sm" @click="browser.rwsGoBack(tab)">← 返回</button>
      <span class="breadcrumb-text">{{ tab.rwsStudyName }}</span>
      <button class="btn btn-sm" @click="browser.loadRWSStudyDetail(tab, tab.rwsStudyOID)">刷新</button>
    </div>

    <p v-if="tab.loading && !tab.rwsSubjects.length" class="faint center-hint">加载中…</p>
    <p v-else-if="tab.error" class="error-text center-hint">{{ tab.error }}</p>

    <template v-else>
      <div class="rws-section-tabs">
        <button
          class="rws-section-btn"
          :class="{ active: activeSection === 'subjects' }"
          @click="activeSection = 'subjects'"
        >
          受试者 ({{ tab.rwsSubjects.length }})
        </button>
        <button
          class="rws-section-btn"
          :class="{ active: activeSection === 'sites' }"
          @click="activeSection = 'sites'"
        >
          中心 ({{ tab.rwsSites.length }})
        </button>
        <button
          class="rws-section-btn"
          :class="{ active: activeSection === 'datasets' }"
          @click="activeSection = 'datasets'"
        >
          数据集 ({{ tab.rwsDatasets.length }})
        </button>
      </div>

      <!-- Subjects -->
      <div v-if="activeSection === 'subjects'" class="rws-table-wrap">
        <table class="rws-table">
          <thead>
            <tr><th>Subject Key</th><th>Site</th><th>Status</th></tr>
          </thead>
          <tbody>
            <tr v-for="s in tab.rwsSubjects" :key="s.SubjectKey" class="rws-row">
              <td>{{ s.SubjectKey }}</td>
              <td>{{ s.SiteRef }}</td>
              <td>{{ s.Status }}</td>
            </tr>
          </tbody>
        </table>
        <p v-if="tab.rwsSubjects.length === 0" class="faint center-hint">暂无受试者数据</p>
      </div>

      <!-- Sites -->
      <div v-if="activeSection === 'sites'" class="rws-table-wrap">
        <table class="rws-table">
          <thead>
            <tr><th>OID</th><th>名称</th><th>中心编号</th></tr>
          </thead>
          <tbody>
            <tr v-for="s in tab.rwsSites" :key="s.OID" class="rws-row">
              <td>{{ s.OID }}</td>
              <td>{{ s.Name }}</td>
              <td>{{ s.StudySiteNumber }}</td>
            </tr>
          </tbody>
        </table>
        <p v-if="tab.rwsSites.length === 0" class="faint center-hint">暂无中心数据</p>
      </div>

      <!-- Datasets -->
      <div v-if="activeSection === 'datasets'" class="rws-datasets-layout" :class="{ 'has-csv': tab.rwsDatasetData }">
        <!-- 数据集列表（CSV 展开时收缩为侧边） -->
        <div class="rws-dataset-list-pane">
          <table class="rws-table">
            <thead>
              <tr><th>名称</th><th>OID</th><th>类型</th><th></th></tr>
            </thead>
            <tbody>
              <tr
                v-for="ds in tab.rwsDatasets" :key="ds.OID"
                class="rws-row"
                :class="{ 'rws-row-active': tab.rwsActiveDataset === ds.Name }"
              >
                <td><strong>{{ ds.Name }}</strong></td>
                <td class="rws-oid">{{ ds.OID }}</td>
                <td><span class="rws-type-badge" :class="typeBadgeClass(ds.Type)">{{ ds.Type || '—' }}</span></td>
                <td>
                  <button
                    class="btn btn-sm"
                    :disabled="tab.loading"
                    @click="browser.loadRWSDatasetCSV(tab, ds.Name, firstType(ds.Type))"
                  >
                    查看
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
          <p v-if="tab.rwsDatasets.length === 0" class="faint center-hint">暂无数据集定义</p>
        </div>

        <!-- Dataset CSV Data（展开时占据主要空间） -->
        <div v-if="tab.rwsDatasetData" class="rws-csv-pane">
          <div class="rws-csv-header">
            <strong>{{ tab.rwsActiveDataset }}</strong>
            <span v-if="tab.rwsDatasetData.Type" class="rws-type-badge" :class="'rws-type-' + tab.rwsDatasetData.Type">{{ tab.rwsDatasetData.Type }}</span>
            <span class="faint">{{ tab.rwsDatasetData.Headers?.length || 0 }} 列 · {{ tab.rwsDatasetData.Rows?.length || 0 }} 行</span>
            <button class="btn btn-sm rws-csv-close" @click="tab.rwsDatasetData = null; tab.rwsActiveDataset = ''">✕ 关闭</button>
          </div>
          <div class="rws-csv-scroll">
            <table class="rws-table rws-csv-table">
              <thead>
                <tr>
                  <th v-for="h in tab.rwsDatasetData.Headers" :key="h">{{ h }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(row, ri) in tab.rwsDatasetData.Rows" :key="ri" class="rws-row">
                  <td v-for="(cell, ci) in row" :key="ci">{{ cell }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
