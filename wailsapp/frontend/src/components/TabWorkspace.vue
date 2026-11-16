<script setup lang="ts">
import { ref } from 'vue'
import { useBrowser, type BrowserTab, type TabPane } from '../stores/browser'
import BucketList from './BucketList.vue'
import ObjectBrowser from './ObjectBrowser.vue'
import SharePointDriveList from './SharePointDriveList.vue'
import SharePointBrowser from './SharePointBrowser.vue'
import RWSStudyList from './RWSStudyList.vue'
import RWSStudyDetail from './RWSStudyDetail.vue'
import PropertiesDialog from './dialogs/PropertiesDialog.vue'

const browser = useBrowser()
const properties = ref<{ title: string; rows: [string, string][] } | null>(null)

interface MenuItem {
  label?: string
  action?: () => void
  disabled?: boolean
  separator?: boolean
}
const ctx = ref<{ x: number; y: number; items: MenuItem[] } | null>(null)

function onBucketProperties(title: string, rows: [string, string][]) {
  properties.value = { title, rows }
}

function tabOf(pane: TabPane): BrowserTab | null {
  return browser.paneActive(pane)
}

function paneStyle(index: number) {
  if (browser.panes.length < 2) return { flexGrow: 1, flexBasis: 0 }
  return {
    flexGrow: index === 0 ? browser.splitOffset : 1 - browser.splitOffset,
    flexBasis: 0,
  }
}

// 标签右键菜单：与旧版 tabs.go menuFor 的 JetBrains 全套一致。
function tabContext(e: MouseEvent, pane: TabPane, tab: BrowserTab) {
  e.preventDefault()
  const idx = pane.tabs.findIndex((t) => t.id === tab.id)
  const n = pane.tabs.length
  const isLeft = pane.id === browser.panes[0].id
  const splitLabel = isLeft
    ? browser.panes.length > 1
      ? '移动到右侧分割'
      : '向右分割并移动'
    : '移动到左侧分割'

  const items: MenuItem[] = [
    { label: '复制标签页', action: () => browser.duplicate(tab.id) },
    { label: '关闭', action: () => browser.closeTab(tab.id) },
    { label: '关闭其他', action: () => browser.closeOthers(tab.id), disabled: n <= 1 },
    { label: '关闭右侧所有标签页', action: () => browser.closeToRight(tab.id), disabled: idx >= n - 1 },
    { label: '关闭所有标签页', action: () => browser.closeAll() },
    { separator: true },
    {
      label: splitLabel,
      action: () => (isLeft ? browser.splitRight(tab.id) : browser.moveToLeft(tab.id)),
    },
    { separator: true },
    { label: '左移', action: () => browser.move(tab.id, -1), disabled: idx <= 0 },
    { label: '右移', action: () => browser.move(tab.id, 1), disabled: idx >= n - 1 },
  ]
  ctx.value = { x: e.clientX, y: e.clientY, items }
}

// 标签条空白处右键：关闭所有 / 取消分屏。
function barContext(e: MouseEvent) {
  e.preventDefault()
  const items: MenuItem[] = []
  if (browser.tabCount > 0) {
    items.push({ label: '关闭所有标签页', action: () => browser.closeAll() })
  }
  if (browser.panes.length > 1) {
    if (items.length) items.push({ separator: true })
    items.push({ label: '取消分屏（合并到左侧）', action: () => browser.unsplit() })
  }
  if (items.length) ctx.value = { x: e.clientX, y: e.clientY, items }
}

// 分隔条拖动：按鼠标位置调整左右窗格比例（限制 20%~80%）。
function startDrag(e: MouseEvent) {
  const row = (e.currentTarget as HTMLElement).parentElement
  if (!row) return
  const move = (ev: MouseEvent) => {
    const rect = row.getBoundingClientRect()
    if (rect.width > 0) browser.setSplitOffset((ev.clientX - rect.left) / rect.width)
  }
  const up = () => {
    window.removeEventListener('mousemove', move)
    window.removeEventListener('mouseup', up)
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
  }
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
  window.addEventListener('mousemove', move)
  window.addEventListener('mouseup', up)
}

function runItem(item: MenuItem) {
  ctx.value = null
  if (!item.disabled) item.action?.()
}
</script>

<template>
  <div class="tab-workspace" @click="ctx = null">
    <div class="panes-row">
      <template v-for="(pane, pi) in browser.panes" :key="pane.id">
        <div class="pane" :style="paneStyle(pi)" @click="browser.activePaneId = pane.id">
          <div class="tabbar" @contextmenu="barContext($event)">
            <div
              v-for="t in pane.tabs"
              :key="t.id"
              class="tab-item"
              :class="{ active: pane.activeId === t.id }"
              :title="t.title"
              @click.stop="browser.selectTab(pane.id, t.id)"
              @contextmenu.stop="tabContext($event, pane, t)"
              @mousedown.middle.prevent="browser.closeTab(t.id)"
            >
              <span class="tab-dot" :class="{ sp: t.account.Provider === 'sharepoint', rws: t.account.Provider === 'rws' }"></span>
              <span class="tab-title">{{ t.title }}</span>
              <button class="tab-close" title="关闭标签" @click.stop="browser.closeTab(t.id)">×</button>
            </div>
            <div class="topbar-spacer"></div>
            <span class="faint tab-hint">
              {{ browser.panes.length > 1 ? '双击分屏' : '右键标签可复制 / 分屏 / 关闭' }}
            </span>
          </div>

          <div class="tab-content">
            <template v-if="tabOf(pane)">
              <BucketList
                v-if="tabOf(pane)!.view === 'buckets'"
                :key="tabOf(pane)!.id"
                :tab="tabOf(pane)!"
                @properties="onBucketProperties"
              />
              <SharePointDriveList
                v-else-if="tabOf(pane)!.view === 'drives'"
                :key="tabOf(pane)!.id"
                :tab="tabOf(pane)!"
                @properties="onBucketProperties"
              />
              <SharePointBrowser
                v-else-if="tabOf(pane)!.view === 'sp-objects'"
                :key="tabOf(pane)!.id"
                :tab="tabOf(pane)!"
              />
              <RWSStudyList
                v-else-if="tabOf(pane)!.view === 'rws-studies'"
                :key="tabOf(pane)!.id"
                :tab="tabOf(pane)!"
              />
              <RWSStudyDetail
                v-else-if="tabOf(pane)!.view === 'rws-detail'"
                :key="tabOf(pane)!.id"
                :tab="tabOf(pane)!"
              />
              <ObjectBrowser v-else :key="tabOf(pane)!.id" :tab="tabOf(pane)!" />
            </template>
            <div v-else class="placeholder">
              <div class="placeholder-icon"></div>
              <div class="placeholder-title">双击左侧账号打开 Bucket 列表</div>
              <div class="faint">每个账号一个标签，切换标签不会重新加载内容</div>
            </div>
          </div>
        </div>
        <div
          v-if="pi === 0 && browser.panes.length > 1"
          class="split-handle"
          title="拖动调整分屏比例"
          @mousedown="startDrag"
        ></div>
      </template>
    </div>

    <div
      v-if="ctx"
      class="context-menu"
      :style="{ left: ctx.x + 'px', top: ctx.y + 'px' }"
      @click.stop
    >
      <template v-for="(it, i) in ctx.items" :key="i">
        <div v-if="it.separator" class="menu-separator"></div>
        <button v-else class="menu-item" :disabled="it.disabled" @click="runItem(it)">
          <span>{{ it.label }}</span>
        </button>
      </template>
    </div>

    <PropertiesDialog
      v-if="properties"
      :title="properties.title"
      :rows="properties.rows"
      @close="properties = null"
    />
  </div>
</template>
