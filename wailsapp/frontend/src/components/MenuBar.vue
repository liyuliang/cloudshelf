<script setup lang="ts">
import { ref } from 'vue'

const emit = defineEmits<{ action: [name: string] }>()
const open = ref<string | null>(null)

type Item = { key: string; label: string; disabled?: boolean }
type Group = { name: string; items: Item[] }

// 菜单结构与旧版 showMain 一致（账号/安全/存储/分享/设置）。
// 标注「后续阶段」的项在 P1 未接入，禁用以免误导；旧 Fyne 版仍可完整使用。
const groups: Group[] = [
  {
    name: '账号',
    items: [
      { key: 'new-account', label: '新增账号' },
      { key: 'edit-account', label: '编辑账号' },
      { key: 'delete-account', label: '删除账号' },
      { key: 'test-account', label: '测试连接' },
      { key: 'export', label: '导出账号' },
      { key: 'import', label: '导入账号' },
    ],
  },
  {
    name: '安全',
    items: [
      { key: 'lock', label: '加锁' },
      { key: 'change-pw', label: '修改主密码' },
    ],
  },
  {
    name: '存储',
    items: [{ key: 'change-db', label: '修改存储位置' }],
  },
  {
    name: '分享',
    items: [
      { key: 'load-share', label: '加载分享配置' },
      { key: 'lan-shares', label: '局域网分享管理' },
    ],
  },
  {
    name: '设置',
    items: [
      { key: 'proxy', label: '代理设置' },
      { key: 'check-update', label: '检查更新' },
    ],
  },
]

function pick(key: string, disabled?: boolean) {
  open.value = null
  if (!disabled) emit('action', key)
}
</script>

<template>
  <header class="topbar" @mouseleave="open = null">
    <div class="brand"><span class="brand-mark"></span>CloudShelf</div>
    <nav class="menubar">
      <div v-for="g in groups" :key="g.name" class="menu-group">
        <button
          class="menu-trigger"
          :class="{ active: open === g.name }"
          @click="open = open === g.name ? null : g.name"
        >
          {{ g.name }}
        </button>
        <div v-if="open === g.name" class="menu-dropdown">
          <button
            v-for="it in g.items"
            :key="it.key"
            class="menu-item"
            :disabled="it.disabled"
            @click="pick(it.key, it.disabled)"
          >
            <span>{{ it.label }}</span>
            <span v-if="it.disabled" class="badge">后续阶段</span>
          </button>
        </div>
      </div>
    </nav>
    <div class="topbar-spacer"></div>
  </header>
</template>
