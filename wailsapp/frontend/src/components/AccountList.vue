<script setup lang="ts">
import { computed, ref } from 'vue'
import { useSession } from '../stores/session'
import type { Account } from '../lib/api'

const emit = defineEmits<{
  action: [name: string, acc: Account | null]
  select: [acc: Account]
}>()

const session = useSession()
const selectedId = ref<number | null>(null)
const ctx = ref<{ x: number; y: number; acc: Account } | null>(null)

// 按 Provider 分组：折叠状态，默认全部折叠
const collapsed = ref<Record<string, boolean>>({})

interface ProviderGroup {
  key: string
  label: string
  accounts: Account[]
}

const groups = computed<ProviderGroup[]>(() => {
  const s3List = session.accounts.filter(a => a.Provider !== 'sharepoint' && a.Provider !== 'rws')
  const spList = session.accounts.filter(a => a.Provider === 'sharepoint')
  const rwsList = session.accounts.filter(a => a.Provider === 'rws')
  const result: ProviderGroup[] = []
  if (s3List.length > 0) {
    result.push({ key: 's3', label: 'S3', accounts: s3List })
  }
  if (spList.length > 0) {
    result.push({ key: 'sharepoint', label: 'SharePoint', accounts: spList })
  }
  if (rwsList.length > 0) {
    result.push({ key: 'rws', label: 'Rave (RWS)', accounts: rwsList })
  }
  return result
})

function isCollapsed(key: string): boolean {
  // 默认折叠：未设置时返回 true，用户手动展开后记住状态
  return collapsed.value[key] !== false
}

function toggleGroup(key: string) {
  collapsed.value[key] = !isCollapsed(key)
}

function select(acc: Account) {
  selectedId.value = acc.ID
  ctx.value = null
  emit('select', acc)
}

function contextMenu(e: MouseEvent, acc: Account) {
  e.preventDefault()
  selectedId.value = acc.ID
  emit('select', acc)
  ctx.value = { x: e.clientX, y: e.clientY, acc }
}

function fire(name: string, acc: Account | null) {
  ctx.value = null
  emit('action', name, acc)
}
</script>

<template>
  <aside class="sidebar" @click="ctx = null">
    <div class="sidebar-head">
      <span>账号</span>
      <span class="badge">{{ session.accounts.length }}</span>
      <div class="topbar-spacer"></div>
      <button class="btn btn-ghost btn-icon" title="新增账号" @click="fire('new-account', null)">＋</button>
    </div>

    <div class="account-scroll">
      <template v-if="groups.length > 0">
        <div v-for="g in groups" :key="g.key" class="account-group">
          <div class="group-header" @click="toggleGroup(g.key)">
            <span class="group-arrow" :class="{ open: !isCollapsed(g.key) }">▶</span>
            <span class="group-dot" :class="{ sp: g.key === 'sharepoint', rws: g.key === 'rws' }"></span>
            <span class="group-label">{{ g.label }}</span>
            <span class="badge">{{ g.accounts.length }}</span>
          </div>
          <div v-show="!isCollapsed(g.key)" class="group-body">
            <div
              v-for="acc in g.accounts"
              :key="acc.ID"
              class="list-item account-item"
              :class="{ active: selectedId === acc.ID }"
              @click="select(acc)"
              @dblclick="fire('open-account', acc)"
              @contextmenu="contextMenu($event, acc)"
            >
              <span class="account-dot" :class="{ sp: acc.Provider === 'sharepoint', rws: acc.Provider === 'rws' }"></span>
              <span class="account-name">{{ acc.Name }}</span>
            </div>
          </div>
        </div>
      </template>
      <p v-else class="empty-hint faint">
        还没有账号，点右上角 ＋ 新增
      </p>
    </div>

    <div
      v-if="ctx"
      class="context-menu"
      :style="{ left: ctx.x + 'px', top: ctx.y + 'px' }"
      @click.stop
    >
      <button class="menu-item" @click="fire('edit-account', ctx.acc)">编辑账号</button>
      <button class="menu-item" @click="fire('delete-account', ctx.acc)">删除账号</button>
      <button class="menu-item" @click="fire('test-account', ctx.acc)">测试连接</button>
    </div>
  </aside>
</template>
