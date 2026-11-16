<script setup lang="ts">
import { ref } from 'vue'
import type { BrowserTab } from '../stores/browser'
import { useBrowser } from '../stores/browser'

const props = defineProps<{ tab: BrowserTab }>()
const browser = useBrowser()

const activeSection = ref<'subjects' | 'sites' | 'datasets'>('subjects')
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
      <div v-if="activeSection === 'datasets'" class="rws-table-wrap">
        <div class="rws-dataset-toolbar">
          <select v-model="tab.rwsDataType" class="input input-sm" @change="browser.loadRWSStudyDetail(tab, tab.rwsStudyOID)">
            <option value="regular">Regular</option>
            <option value="raw">Raw</option>
          </select>
        </div>
        <table class="rws-table">
          <thead>
            <tr><th>OID</th><th>名称</th><th>Domain</th><th>字段数</th><th></th></tr>
          </thead>
          <tbody>
            <tr v-for="ds in tab.rwsDatasets" :key="ds.OID" class="rws-row">
              <td class="rws-oid">{{ ds.OID }}</td>
              <td>{{ ds.Name }}</td>
              <td>{{ ds.Domain }}</td>
              <td>{{ ds.Fields?.length || 0 }}</td>
              <td>
                <button
                  class="btn btn-sm"
                  :disabled="tab.loading"
                  @click="browser.loadRWSDatasetCSV(tab, ds.Name)"
                >
                  查看数据
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <p v-if="tab.rwsDatasets.length === 0" class="faint center-hint">暂无数据集定义</p>

        <!-- Dataset CSV Data -->
        <div v-if="tab.rwsDatasetData" class="rws-csv-panel">
          <div class="rws-csv-header">
            <strong>{{ tab.rwsActiveDataset }}</strong>
            <span class="faint">（{{ tab.rwsDatasetData.Rows?.length || 0 }} 行）</span>
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
