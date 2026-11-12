<script setup lang="ts">
import { ref } from 'vue'
import { useSession } from '../stores/session'
import type { Account } from '../lib/api'

const emit = defineEmits<{
  action: [name: string, acc: Account | null]
  select: [acc: Account]
}>()

const session = useSession()
const selectedId = ref<number | null>(null)
const ctx = ref<{ x: number; y: number; acc: Account } | null>(null)

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
      <div
        v-for="acc in session.accounts"
        :key="acc.ID"
        class="list-item account-item"
        :class="{ active: selectedId === acc.ID }"
        @click="select(acc)"
        @dblclick="fire('open-account', acc)"
        @contextmenu="contextMenu($event, acc)"
      >
        <span class="account-dot" :class="{ sp: acc.Provider === 'sharepoint' }"></span>
        <span class="account-name">{{ acc.Name }}</span>
        <span v-if="acc.Provider === 'sharepoint'" class="badge badge-accent">sp</span>
        <span v-else class="badge">s3</span>
      </div>
      <p v-if="session.accounts.length === 0" class="empty-hint faint">还没有账号，点右上角 ＋ 新增</p>
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
