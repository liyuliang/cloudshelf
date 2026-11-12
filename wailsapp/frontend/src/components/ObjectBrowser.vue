<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useBrowser, type BrowserTab } from '../stores/browser'
import { api, errText, type S3Object, type TransferItem } from '../lib/api'
import { breadcrumbs, humanSize, iconFor, parentPrefix } from '../lib/path'
import Modal from './Modal.vue'
import PropertiesDialog from './dialogs/PropertiesDialog.vue'
import PresignDialog from './dialogs/PresignDialog.vue'
import CreateShareDialog from './dialogs/CreateShareDialog.vue'

const props = defineProps<{ tab: BrowserTab }>()
const browser = useBrowser()

const searchText = ref('')
const selectedKey = ref<string | null>(null)
const checked = ref<Record<string, boolean>>({})
const properties = ref<{ title: string; rows: [string, string][] } | null>(null)
const presignTarget = ref<S3Object | null>(null)
const shareDialog = ref(false)
const folderDialog = ref(false)
const folderName = ref('')
const deleteStage = ref<0 | 1 | 2>(0)
const busyText = ref('')
const notice = ref('')
let searchTimer: number | undefined
let noticeTimer: number | undefined

const selectedObj = computed(() => props.tab.objects.find((o) => o.Key === selectedKey.value) ?? null)
const crumbs = computed(() => breadcrumbs(props.tab.prefix))
const checkedCount = computed(() => Object.values(checked.value).filter(Boolean).length)

watch(
  () => props.tab.bucket,
  () => {
    searchText.value = ''
  },
)
watch(
  () => props.tab.objects,
  () => {
    selectedKey.value = null
    checked.value = {}
  },
)
onBeforeUnmount(() => {
  if (searchTimer) window.clearTimeout(searchTimer)
  if (noticeTimer) window.clearTimeout(noticeTimer)
})

function notify(msg: string) {
  notice.value = msg
  if (noticeTimer) window.clearTimeout(noticeTimer)
  noticeTimer = window.setTimeout(() => (notice.value = ''), 3500)
}

function onSearchInput() {
  if (searchTimer) window.clearTimeout(searchTimer)
  searchTimer = window.setTimeout(() => {
    if (!searchText.value.trim()) browser.clearSearch(props.tab)
    else void browser.doSearch(props.tab, searchText.value)
  }, 150)
}

function select(o: S3Object) {
  selectedKey.value = o.Key
  // 旧版行为：文件行左键点中即自动勾选
  if (!o.IsDir) checked.value[o.Key] = true
}

function toggleCheck(o: S3Object) {
  checked.value[o.Key] = !checked.value[o.Key]
}

function open(o: S3Object) {
  if (o.IsDir) void browser.navigate(props.tab, o.Key)
}

function goUp() {
  void browser.navigate(props.tab, parentPrefix(props.tab.prefix))
}

function goCrumb(prefix: string) {
  void browser.navigate(props.tab, prefix)
}

function showProps(o: S3Object) {
  const rows: [string, string][] = o.IsDir
    ? [['名称', o.Name], ['Key', o.Key], ['类型', '文件夹']]
    : [
        ['名称', o.Name],
        ['Key', o.Key],
        ['大小', humanSize(o.Size)],
        ['类型', o.Type],
        ['最后修改', o.LastModified],
      ]
  properties.value = { title: o.IsDir ? '文件夹属性' : '文件属性', rows }
}

function openPresign() {
  const o = selectedObj.value
  if (!o || o.IsDir) return notify('请先选择一个文件')
  presignTarget.value = o
}

function askDelete() {
  const o = selectedObj.value
  if (!o) return notify('请先选择一个文件或文件夹')
  deleteStage.value = 1
}

async function confirmDelete() {
  const o = selectedObj.value
  if (!o) return
  deleteStage.value = 0
  busyText.value = '删除中…'
  try {
    await api.deleteObject(props.tab.account, props.tab.bucket, o.Key)
    await browser.navigate(props.tab, props.tab.prefix)
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
    await api.createFolder(props.tab.account, props.tab.bucket, props.tab.prefix, name)
    await browser.navigate(props.tab, props.tab.prefix)
    notify('文件夹已创建')
  } catch (e) {
    props.tab.error = errText(e)
  } finally {
    busyText.value = ''
  }
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
  const items: TransferItem[] = paths.map((p) => {
    const name = p.split(/[\\/]/).pop() || p
    return { ID: makeTaskId(), Name: name, Key: '', LocalPath: p }
  })
  try {
    const res = await api.uploadFiles(props.tab.account, props.tab.bucket, props.tab.prefix, items)
    await browser.navigate(props.tab, props.tab.prefix)
    reportResult(res, '上传')
  } catch (e) {
    props.tab.error = errText(e)
  } finally {
    busyText.value = ''
  }
}

async function download() {
  const files = props.tab.objects.filter((o) => !o.IsDir && checked.value[o.Key])
  if (files.length === 0) return notify('请先勾选要下载的文件（行左侧复选框）')
  const dir = await api.pickDirectory('选择下载到的目录')
  if (!dir) return
  busyText.value = '下载中…'
  const items: TransferItem[] = files.map((o) => {
    const name = o.Name.split('/').pop() || o.Name
    return { ID: makeTaskId(), Name: name, Key: o.Key, LocalPath: '' }
  })
  try {
    const res = await api.downloadFiles(props.tab.account, props.tab.bucket, dir, items)
    await api.revealInExplorer(dir.replace(/[\\/]$/, '') + '/' + items[0].Name)
    reportResult(res, '下载')
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
        <button class="crumb-link" @click="browser.loadBuckets(tab)">← Bucket 列表</button>
        <button class="crumb-link crumb-root" @click="goCrumb('')">{{ tab.bucket }}</button>
        <template v-for="c in crumbs" :key="c.prefix">
          <span class="crumb-sep">/</span>
          <button class="crumb-link" @click="goCrumb(c.prefix)">{{ c.label }}</button>
        </template>
      </div>
      <div class="search-row">
        <input
          v-model="searchText"
          class="input"
          placeholder="按名称前缀搜索，或粘贴完整路径自动进入该目录…"
          @input="onSearchInput"
        />
        <button
          v-if="searchText || tab.searchActive"
          class="btn btn-ghost"
          title="清空搜索"
          @click="searchText = ''; browser.clearSearch(tab)"
        >
          ✕
        </button>
      </div>
      <div v-if="tab.searchActive" class="search-status">
        服务端前缀搜索「{{ tab.searchKeyword }}」· 找到 {{ tab.objects.length }} 项
      </div>
    </div>

    <div class="table-wrap">
      <div class="obj-grid table-header">
        <span>名称</span><span>大小</span><span>类型</span><span>修改时间</span>
      </div>
      <div class="table-scroll">
        <div v-if="tab.prefix !== ''" class="obj-grid table-row back-row" @dblclick="goUp">
          <span class="cell-main">⬆ 返回上一级</span><span></span><span></span><span></span>
        </div>

        <div
          v-for="o in tab.objects"
          :key="o.Key"
          class="obj-grid table-row"
          :class="{ active: selectedKey === o.Key }"
          @click="select(o)"
          @dblclick="open(o)"
          @contextmenu.prevent="showProps(o)"
        >
          <span class="cell-main">
            <input
              v-if="!o.IsDir"
              class="row-check"
              type="checkbox"
              :checked="!!checked[o.Key]"
              @click.stop
              @change="toggleCheck(o)"
            />
            <span v-else class="row-check-placeholder"></span>
            <span class="row-icon">{{ iconFor(o) }}</span>
            <span class="cell-name" :title="o.Name">{{ o.Name }}</span>
          </span>
          <span class="cell-meta">{{ o.IsDir ? '' : humanSize(o.Size) }}</span>
          <span class="cell-meta">{{ o.IsDir ? '文件夹' : o.Type }}</span>
          <span class="cell-meta">{{ o.LastModified || '' }}</span>
        </div>

        <p v-if="tab.loading" class="empty-hint faint">正在加载…</p>
        <p v-else-if="tab.objects.length === 0" class="empty-hint faint">该目录下没有条目</p>
      </div>
    </div>

    <div class="msg-area">
      <p v-if="tab.error" class="error-text">⚠ 加载失败: {{ tab.error }}</p>
      <p v-if="tab.truncated" class="truncate-hint">
        ⚠ 该目录条目超过 S3 单页上限（1000），仅显示第一页；请用上方搜索框按名称前缀查找其它文件。
      </p>
      <p v-if="notice" class="success-text">{{ notice }}</p>
    </div>

    <div class="action-bar">
      <button class="btn" :disabled="tab.loading" @click="browser.refresh(tab)">刷新</button>
      <button class="btn" :disabled="!!busyText" @click="upload">上传文件</button>
      <button class="btn" :disabled="!!busyText" @click="download">
        下载文件<span v-if="checkedCount"> ({{ checkedCount }})</span>
      </button>
      <button class="btn btn-danger" :disabled="!!busyText" @click="askDelete">删除</button>
      <button class="btn" @click="openPresign">Web URL</button>
      <button class="btn" @click="folderDialog = true">新建文件夹</button>
      <button class="btn" @click="shareDialog = true">分享只读</button>
      <div class="topbar-spacer"></div>
      <span v-if="busyText" class="busy-text">{{ busyText }}</span>
    </div>

    <PropertiesDialog
      v-if="properties"
      :title="properties.title"
      :rows="properties.rows"
      @close="properties = null"
    />
    <PresignDialog
      v-if="presignTarget"
      :key="presignTarget.Key"
      :account="tab.account"
      :bucket="tab.bucket"
      :objectKey="presignTarget.Key"
      @close="presignTarget = null"
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
      <p>确定删除「{{ selectedObj?.Name }}」吗？</p>
      <div class="dialog-actions">
        <button class="btn" @click="deleteStage = 0">取消</button>
        <button class="btn btn-danger" @click="deleteStage = 2">继续</button>
      </div>
    </Modal>
    <Modal v-if="deleteStage === 2" title="再次确认" width="380px" @close="deleteStage = 0">
      <p>此操作不可撤销，确定要删除「{{ selectedObj?.Name }}」吗？</p>
      <div class="dialog-actions">
        <button class="btn" @click="deleteStage = 0">取消</button>
        <button class="btn btn-danger" @click="confirmDelete">确定删除</button>
      </div>
    </Modal>

    <CreateShareDialog
      v-if="shareDialog"
      :account="tab.account"
      :bucket="tab.bucket"
      @close="shareDialog = false"
    />
  </div>
</template>
