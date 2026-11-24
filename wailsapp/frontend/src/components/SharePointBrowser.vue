<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useBrowser, type BrowserTab } from '../stores/browser'
import { api, errText, type SPItem, type SPTransferItem } from '../lib/api'
import { humanSize, spIconFor, spBreadcrumbs, spParentPath, spWithinRoot } from '../lib/path'
import Modal from './Modal.vue'
import PropertiesDialog from './dialogs/PropertiesDialog.vue'
import CreateShareDialog from './dialogs/CreateShareDialog.vue'

const props = defineProps<{ tab: BrowserTab }>()
const browser = useBrowser()

const filterText = ref('')
const selectedId = ref<string | null>(null)
const checked = ref<Record<string, boolean>>({})
const properties = ref<{ title: string; rows: [string, string][] } | null>(null)
const shareDialog = ref(false)
const folderDialog = ref(false)
const folderName = ref('')
const deleteStage = ref<0 | 1 | 2>(0)
const busyText = ref('')
const notice = ref('')
let filterTimer: number | undefined
let noticeTimer: number | undefined

const selectedItem = computed(() => props.tab.spItems.find((it) => it.ID === selectedId.value) ?? null)
const crumbs = computed(() => spBreadcrumbs(props.tab.itemPath))
const checkedCount = computed(() => Object.values(checked.value).filter(Boolean).length)

// 本地名称过滤（与旧版一致：输入关键字 → 切换行的显隐）
const filteredItems = computed(() => {
  const lf = filterText.value.toLowerCase().trim()
  if (!lf) return props.tab.spItems
  return props.tab.spItems.filter((it) => it.Name.toLowerCase().includes(lf))
})

watch(
  () => props.tab.driveID,
  () => { filterText.value = '' },
)
watch(
  () => props.tab.spItems,
  () => {
    selectedId.value = null
    checked.value = {}
  },
)
onBeforeUnmount(() => {
  if (filterTimer) window.clearTimeout(filterTimer)
  if (noticeTimer) window.clearTimeout(noticeTimer)
})

function notify(msg: string) {
  notice.value = msg
  if (noticeTimer) window.clearTimeout(noticeTimer)
  noticeTimer = window.setTimeout(() => (notice.value = ''), 3500)
}

function select(it: SPItem) {
  selectedId.value = it.ID
  if (!it.IsFolder) checked.value[it.ID] = true
}

function toggleCheck(it: SPItem) {
  checked.value[it.ID] = !checked.value[it.ID]
}

function open(it: SPItem) {
  if (it.IsFolder) {
    const sub = props.tab.itemPath
      ? props.tab.itemPath + '/' + it.Name
      : it.Name
    void browser.spNavigate(props.tab, sub)
  }
}

function goUp() {
  const readOnly = props.tab.readOnly
  const cur = props.tab.itemPath
  if (cur === '') {
    if (!readOnly) void browser.loadDrives(props.tab)
    return
  }
  const parent = spParentPath(cur)
  // 只读分享不允许回到分享根目录之上
  if (readOnly && !spWithinRoot(parent, props.tab.shareRootPath)) return
  void browser.spNavigate(props.tab, parent)
}

// 是否显示返回按钮
const showBack = computed(() => {
  if (!props.tab.readOnly) return true
  const cur = props.tab.itemPath
  if (cur === '') return false
  const parent = spParentPath(cur)
  return spWithinRoot(parent, props.tab.shareRootPath)
})

function goCrumb(path: string) {
  void browser.spNavigate(props.tab, path)
}

function showProps(it: SPItem) {
  const rows: [string, string][] = it.IsFolder
    ? [
        ['名称', it.Name],
        ['ID', it.ID],
        ['类型', '文件夹'],
        ['创建时间', it.CreatedTime || '-'],
        ['修改时间', it.LastModified || '-'],
        ['修改者', it.LastModifiedBy || '-'],
      ]
    : [
        ['名称', it.Name],
        ['ID', it.ID],
        ['大小', humanSize(it.Size)],
        ['创建时间', it.CreatedTime || '-'],
        ['修改时间', it.LastModified || '-'],
        ['创建者', it.CreatedBy || '-'],
        ['修改者', it.LastModifiedBy || '-'],
      ]
  properties.value = { title: it.IsFolder ? '文件夹属性' : '文件属性', rows }
}

function makeTaskId(): string {
  return `t-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`
}

function reportResult(res: { Success: number; Failed: number; Errors?: string[] }, verb: string) {
  notify(`${verb}完成：${res.Success} 成功${res.Failed ? `，${res.Failed} 失败` : ''}`)
  if (res.Errors && res.Errors.length) props.tab.error = res.Errors.join('；')
}

async function upload() {
  const paths = await api.pickMultipleFiles('选择要上传的文件（可多选）')
  if (!paths || paths.length === 0) return
  busyText.value = '上传中…'
  const items: SPTransferItem[] = paths.map((p) => {
    const name = p.split(/[\\/]/).pop() || p
    return { ID: makeTaskId(), Name: name, ItemID: '', LocalPath: p }
  })
  try {
    const res = await api.spUploadFiles(
      props.tab.account, props.tab.driveID, props.tab.itemPath, items,
    )
    await browser.loadSPItems(props.tab)
    reportResult(res, '上传')
  } catch (e) {
    props.tab.error = errText(e)
  } finally {
    busyText.value = ''
  }
}

async function download() {
  const files = props.tab.spItems.filter((it) => !it.IsFolder && checked.value[it.ID])
  if (files.length === 0) return notify('请先勾选要下载的文件（行左侧复选框）')
  const dir = await api.pickDirectory('选择下载到的目录')
  if (!dir) return
  busyText.value = '下载中…'
  const items: SPTransferItem[] = files.map((it) => ({
    ID: makeTaskId(), Name: it.Name, ItemID: it.ID, LocalPath: '',
  }))
  try {
    const res = await api.spDownloadFiles(
      props.tab.account, props.tab.driveID, dir, items,
    )
    await api.revealInExplorer(dir.replace(/[\\/]$/, '') + '/' + items[0].Name)
    reportResult(res, '下载')
  } catch (e) {
    props.tab.error = errText(e)
  } finally {
    busyText.value = ''
  }
}

async function askDelete() {
  const it = selectedItem.value
  if (!it) return notify('请先选择一个文件或文件夹')
  deleteStage.value = 1
}

async function confirmDelete() {
  const it = selectedItem.value
  if (!it) return
  deleteStage.value = 0
  busyText.value = '删除中…'
  try {
    await api.spDeleteItem(props.tab.account, props.tab.driveID, it.ID)
    await browser.loadSPItems(props.tab)
    notify('已删除')
  } catch (e) {
    props.tab.error = errText(e)
  } finally {
    busyText.value = ''
  }
}

async function createFolder() {
  const name = folderName.value.trim()
  if (!name) return
  folderDialog.value = false
  folderName.value = ''
  busyText.value = '创建中…'
  try {
    await api.spCreateFolder(props.tab.account, props.tab.driveID, props.tab.itemPath, name)
    await browser.loadSPItems(props.tab)
    notify('文件夹已创建')
  } catch (e) {
    props.tab.error = errText(e)
  } finally {
    busyText.value = ''
  }
}
</script>

<template>
  <div class="object-browser">
    <div class="browser-top">
      <div class="crumbs">
        <button v-if="!tab.readOnly" class="crumb-link" @click="browser.loadDrives(tab)">← 文档库列表</button>
        <button class="crumb-link crumb-root" @click="goCrumb('')">{{ tab.driveName }}</button>
        <template v-for="c in crumbs" :key="c.path">
          <span class="crumb-sep">/</span>
          <button class="crumb-link" @click="goCrumb(c.path)">{{ c.label }}</button>
        </template>
      </div>
      <div class="search-row">
        <input
          v-model="filterText"
          class="input"
          placeholder="输入关键字过滤文件名…"
        />
        <button
          v-if="filterText"
          class="btn btn-ghost"
          title="清空搜索"
          @click="filterText = ''"
        >
          ✕
        </button>
      </div>
      <div v-if="filterText" class="search-status">
        本地过滤「{{ filterText }}」· 显示 {{ filteredItems.length }} / {{ tab.spItems.length }} 项
      </div>
      <div v-if="tab.readOnly && tab.shareSession" class="search-status">
        🔒 只读分享 · {{ tab.shareSession.Remaining }}
        <template v-if="tab.shareSession.ExpiresAt"> · 有效期至 {{ tab.shareSession.ExpiresAt }}</template>
      </div>
    </div>

    <div class="table-wrap">
      <div class="obj-grid-sp table-header">
        <span>名称</span><span>大小</span><span>类型</span><span>创建时间</span><span>修改时间</span><span>修改者</span>
      </div>
      <div class="table-scroll">
        <div v-if="showBack" class="obj-grid-sp table-row back-row" @dblclick="goUp">
          <span class="cell-main">⬆ 返回上一级</span><span></span><span></span><span></span><span></span><span></span>
        </div>

        <div
          v-for="it in filteredItems"
          :key="it.ID"
          class="obj-grid-sp table-row"
          :class="{ active: selectedId === it.ID }"
          @click="select(it)"
          @dblclick="open(it)"
          @contextmenu.prevent="showProps(it)"
        >
          <span class="cell-main">
            <input
              v-if="!it.IsFolder"
              class="row-check"
              type="checkbox"
              :checked="!!checked[it.ID]"
              @click.stop
              @change="toggleCheck(it)"
            />
            <span v-else class="row-check-placeholder"></span>
            <span class="row-icon">{{ spIconFor(it) }}</span>
            <span class="cell-name" :title="it.Name">{{ it.Name }}</span>
          </span>
          <span class="cell-meta">{{ it.IsFolder ? '' : humanSize(it.Size) }}</span>
          <span class="cell-meta">{{ it.IsFolder ? '文件夹' : '' }}</span>
          <span class="cell-meta">{{ it.CreatedTime || '' }}</span>
          <span class="cell-meta">{{ it.LastModified || '' }}</span>
          <span class="cell-meta cell-modifier" :title="it.LastModifiedBy || ''">{{ it.LastModifiedBy || '' }}</span>
        </div>

        <p v-if="tab.loading" class="empty-hint faint">正在加载…</p>
        <p v-else-if="filteredItems.length === 0" class="empty-hint faint">该目录下没有条目</p>
      </div>
    </div>

    <div class="msg-area">
      <p v-if="tab.error" class="error-text">⚠ 加载失败: {{ tab.error }}</p>
      <p v-if="notice" class="success-text">{{ notice }}</p>
    </div>

    <div class="action-bar">
      <button class="btn" :disabled="tab.loading" @click="browser.refresh(tab)">刷新</button>
      <template v-if="!tab.readOnly">
        <button class="btn" :disabled="!!busyText" @click="upload">上传文件</button>
      </template>
      <button class="btn" :disabled="!!busyText" @click="download">
        下载文件<span v-if="checkedCount"> ({{ checkedCount }})</span>
      </button>
      <template v-if="!tab.readOnly">
        <button class="btn btn-danger" :disabled="!!busyText" @click="askDelete">删除</button>
        <button class="btn" @click="folderDialog = true">新建文件夹</button>
        <button class="btn" @click="shareDialog = true">分享只读</button>
      </template>
      <div class="topbar-spacer"></div>
      <span v-if="busyText" class="busy-text">{{ busyText }}</span>
    </div>

    <PropertiesDialog
      v-if="properties"
      :title="properties.title"
      :rows="properties.rows"
      @close="properties = null"
    />

    <Modal v-if="folderDialog" title="新建文件夹" width="360px" @close="folderDialog = false">
      <div class="form-row">
        <label>名称</label>
        <input v-model="folderName" class="input" placeholder="文件夹名" @keyup.enter="createFolder" />
      </div>
      <div class="dialog-actions">
        <button class="btn" @click="folderDialog = false">取消</button>
        <button class="btn btn-primary" @click="createFolder">创建</button>
      </div>
    </Modal>

    <Modal v-if="deleteStage === 1" title="确认删除" width="380px" @close="deleteStage = 0">
      <p>确定删除「{{ selectedItem?.Name }}」吗？</p>
      <div class="dialog-actions">
        <button class="btn" @click="deleteStage = 0">取消</button>
        <button class="btn btn-danger" @click="deleteStage = 2">继续</button>
      </div>
    </Modal>
    <Modal v-if="deleteStage === 2" title="再次确认" width="380px" @close="deleteStage = 0">
      <p>此操作不可撤销，确定要删除「{{ selectedItem?.Name }}」吗？</p>
      <div class="dialog-actions">
        <button class="btn" @click="deleteStage = 0">取消</button>
        <button class="btn btn-danger" @click="confirmDelete">确定删除</button>
      </div>
    </Modal>

    <CreateShareDialog
      v-if="shareDialog"
      :account="tab.account"
      :driveID="tab.driveID"
      :driveName="tab.driveName"
      :itemPath="tab.itemPath"
      @close="shareDialog = false"
    />
  </div>
</template>
