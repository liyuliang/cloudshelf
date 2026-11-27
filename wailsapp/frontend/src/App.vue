<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import MenuBar from './components/MenuBar.vue'
import AccountList from './components/AccountList.vue'
import UnlockGate from './components/UnlockGate.vue'
import TabWorkspace from './components/TabWorkspace.vue'
import TransferBar from './components/TransferBar.vue'
import LogPanel from './components/LogPanel.vue'
import LANSharePanel from './components/LANSharePanel.vue'
import AccountEditDialog from './components/dialogs/AccountEditDialog.vue'
import ProxyDialog from './components/dialogs/ProxyDialog.vue'
import SimpleDialogs from './components/dialogs/SimpleDialogs.vue'
import Modal from './components/Modal.vue'
import { useSession } from './stores/session'
import { useBrowser } from './stores/browser'
import { api, errText, type Account, type ShareSession } from './lib/api'

type Dialog = 'account-edit' | 'proxy' | 'change-pw' | 'change-db' | 'export' | 'import' | null
type Kind = 'change-pw' | 'change-db' | 'export' | 'import'

const session = useSession()
const browser = useBrowser()
const dialog = ref<Dialog>(null)
const editTarget = ref<Account | null>(null)
const selected = ref<Account | null>(null)
const toast = ref('')
let toastTimer: number | undefined

const simpleKind = computed<Kind | null>(() => {
  const d = dialog.value
  return d === 'change-pw' || d === 'change-db' || d === 'export' || d === 'import'
    ? (d as Kind)
    : null
})

const confirmState = ref<{ message: string; yes: () => void } | null>(null)
const lanPanelOpen = ref(false)

const AUTO_LOCK_MS = 10 * 60 * 1000
const ACTIVITY_EVENTS = ['mousemove', 'mousedown', 'keydown', 'wheel', 'touchstart'] as const
let idleTimer: number | undefined
let lastActivityReset = 0

function clearIdleTimer() {
  if (idleTimer !== undefined) {
    window.clearTimeout(idleTimer)
    idleTimer = undefined
  }
}

async function autoLock() {
  clearIdleTimer()
  if (!session.unlocked) return
  dialog.value = null
  confirmState.value = null
  editTarget.value = null
  selected.value = null
  // 只锁定账号库并回到主密码界面，不关闭后台局域网分享服务（lanServer）
  await session.lock()
  notify('超过 10 分钟未操作，已自动锁定')
}

function resetIdleTimer(force = false) {
  if (!session.unlocked) {
    clearIdleTimer()
    return
  }
  const now = Date.now()
  if (!force && now - lastActivityReset < 1000) return
  lastActivityReset = now
  clearIdleTimer()
  idleTimer = window.setTimeout(() => {
    void autoLock()
  }, AUTO_LOCK_MS)
}

function onUserActivity() {
  resetIdleTimer(false)
}

watch(
  () => session.unlocked,
  (unlocked) => {
    if (unlocked) {
      resetIdleTimer(true)
    } else {
      clearIdleTimer()
    }
  },
)

onMounted(() => {
  void session.init()
  for (const ev of ACTIVITY_EVENTS) {
    window.addEventListener(ev, onUserActivity, { passive: true })
  }
})

onUnmounted(() => {
  clearIdleTimer()
  for (const ev of ACTIVITY_EVENTS) {
    window.removeEventListener(ev, onUserActivity)
  }
})

function notify(msg: string) {
  toast.value = msg
  if (toastTimer) window.clearTimeout(toastTimer)
  toastTimer = window.setTimeout(() => (toast.value = ''), 3500)
}

async function handleMenu(name: string) {
  switch (name) {
    case 'new-account':
      editTarget.value = null
      dialog.value = 'account-edit'
      break
    case 'edit-account':
      if (!selected.value) return notify('请先选择一个账号')
      editTarget.value = selected.value
      dialog.value = 'account-edit'
      break
    case 'delete-account':
      askDelete()
      break
    case 'test-account':
      await testSelected()
      break
    case 'export':
      dialog.value = 'export'
      break
    case 'import':
      dialog.value = 'import'
      break
    case 'lock':
      await session.lock()
      selected.value = null
      notify('已加锁')
      break
    case 'change-pw':
      dialog.value = 'change-pw'
      break
    case 'change-db':
      dialog.value = 'change-db'
      break
    case 'proxy':
      dialog.value = 'proxy'
      break
    case 'open-account':
      if (selected.value) browser.openAccount(selected.value)
      break
    case 'load-share':
      void loadShareFromMenu()
      break
    case 'lan-shares':
      lanPanelOpen.value = !lanPanelOpen.value
      break
    default:
      break
  }
}

// 从菜单加载分享配置
async function loadShareFromMenu() {
  try {
    const path = await api.pickFile('选择分享配置文件（.cloudshelf，兼容旧版 .s3share）')
    if (!path) return
    const b64 = await api.readFileB64(path)
    const needPw = await api.needSharePassword(b64)
    if (needPw) {
      // 需要口令：简单用 prompt（后续可改为 Modal）
      const pw = window.prompt('请输入分享口令')
      if (!pw) return
      const sess = await api.openShareFile(b64, pw)
      openShareTab(sess)
    } else {
      const sess = await api.openShareFile(b64, '')
      openShareTab(sess)
    }
  } catch (e) {
    notify(errText(e))
  }
}

function openShareTab(sess: ShareSession) {
  if (sess.Expired) {
    notify(`该分享的有效期已于 ${sess.ExpiresAt} 结束，无法使用。`)
    return
  }
  browser.openShareSession(sess)
}

// 从解锁界面加载的分享（未解锁账号库也可用）
function onShareLoaded(sess: ShareSession) {
  // 标记为分享模式（不需要解锁也能浏览）
  session.unlocked = true
  browser.openShareSession(sess)
}

function handleAccountAction(name: string, acc: Account | null) {
  if (acc) selected.value = acc
  void handleMenu(name)
}

function askDelete() {
  const acc = selected.value
  if (!acc) return notify('请先选择一个账号')
  confirmState.value = {
    message: `确定删除账号「${acc.Name}」吗？此操作不可撤销。`,
    yes: async () => {
      confirmState.value = null
      try {
        await api.deleteAccount(acc.ID)
        await session.reloadAccounts()
        selected.value = null
        notify('账号已删除')
      } catch (e) {
        notify(errText(e))
      }
    },
  }
}

async function testSelected() {
  const acc = selected.value
  if (!acc) return notify('请先选择一个账号')
  notify('正在测试连接…')
  try {
    await api.testAccount(acc)
    notify('连接成功')
  } catch (e) {
    notify('连接失败：' + errText(e))
  }
}

async function onSaved() {
  dialog.value = null
  try {
    await session.reloadAccounts()
  } catch (e) {
    notify(errText(e))
  }
  notify('账号已保存')
}

async function onChanged() {
  await session.reloadAccounts()
}

async function onRelock() {
  selected.value = null
  await session.lock()
  notify('存储位置已修改，请重新输入主密码解锁')
}
</script>

<template>
  <div class="shell">
    <MenuBar v-if="session.unlocked" @action="handleMenu" />

    <UnlockGate v-if="!session.unlocked" @open-proxy="dialog = 'proxy'" @share-loaded="onShareLoaded" />

    <div v-else class="body-row">
      <AccountList
        @action="handleAccountAction"
        @select="(acc) => (selected = acc)"
      />
      <main class="content">
        <TabWorkspace />
      </main>
    </div>

    <TransferBar />
    <LANSharePanel v-if="session.unlocked && lanPanelOpen" @notify="notify" />
    <LogPanel />

    <AccountEditDialog
      v-if="dialog === 'account-edit'"
      :account="editTarget"
      @close="dialog = null"
      @saved="onSaved"
    />
    <ProxyDialog v-if="dialog === 'proxy'" @close="dialog = null" />
    <SimpleDialogs
      v-if="simpleKind"
      :kind="simpleKind"
      @close="dialog = null"
      @changed="onChanged"
      @relock="onRelock"
    />

    <Modal v-if="confirmState" title="确认删除" width="360px" @close="confirmState = null">
      <p>{{ confirmState.message }}</p>
      <div class="dialog-actions">
        <button class="btn" @click="confirmState = null">取消</button>
        <button class="btn btn-danger" @click="confirmState?.yes()">删除</button>
      </div>
    </Modal>

    <div v-if="toast" class="toast">{{ toast }}</div>
  </div>
</template>
