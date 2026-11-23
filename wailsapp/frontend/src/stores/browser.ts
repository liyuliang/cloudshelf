import { defineStore } from 'pinia'
import { computed, reactive, ref } from 'vue'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import {
  api, errText,
  type Account, type Bucket, type S3Object,
  type Drive, type SPItem, type ShareSession,
  type RWSStudy, type RWSSubject, type RWSSite,
  type RWSDataset, type RWSDatasetData,
} from '../lib/api'

export type ViewKind = 'buckets' | 'objects' | 'drives' | 'sp-objects' | 'rws-studies' | 'rws-detail'

export interface BrowserTab {
  id: string
  key: string
  title: string
  account: Account
  view: ViewKind
  bucket: string
  prefix: string
  buckets: Bucket[]
  objects: S3Object[]
  truncated: boolean
  loading: boolean
  error: string
  searchKeyword: string
  searchActive: boolean
  // SharePoint 专用
  drives: Drive[]
  driveID: string
  driveName: string
  itemPath: string
  spItems: SPItem[]
  spFilterText: string
  // 只读分享
  readOnly: boolean
  shareRootPath: string // 分享根目录，只读分享不允许回到该目录之上
  // 分享会话元数据（用于信息展示）
  shareSession?: ShareSession
  // RWS 专用
  rwsStudies: RWSStudy[]
  rwsStudyOID: string
  rwsStudyName: string
  rwsSubjects: RWSSubject[]
  rwsSites: RWSSite[]
  rwsDatasets: RWSDataset[]
  rwsDatasetData: RWSDatasetData | null
  rwsActiveDataset: string
  rwsDataType: string // 'regular' | 'raw'
}


// P3：最多两个窗格，各自维护标签列表与选中项，对应旧版 tabs.go 的 tabWorkspace/tabPane。
export interface TabPane {
  id: string
  tabs: BrowserTab[]
  activeId: string | null
}

export const useBrowser = defineStore('browser', () => {
  const panes = ref<TabPane[]>([{ id: 'pane-0', tabs: [], activeId: null }])
  const activePaneId = ref('pane-0')
  const splitOffset = ref(0.5)

  // 监听后台逐个桶权限探测事件，实时更新 bucket 列表状态
  EventsOn('bucket:access', (...args: unknown[]) => {
    const first = args.length === 1 ? args[0] : args
    const e = (Array.isArray(first) ? first[0] : first) as
      { AccountID: number; Bucket: string; Accessible: boolean } | undefined
    if (!e || !e.Bucket) return
    for (const pane of panes.value) {
      for (const tab of pane.tabs) {
        if (tab.account.ID === e.AccountID && tab.view === 'buckets') {
          const b = tab.buckets.find(bk => bk.Name === e.Bucket)
          if (b) {
            b.Accessible = e.Accessible
            b.Checked = true
          }
        }
      }
    }
  })

  const activePane = computed(
    () => panes.value.find((p) => p.id === activePaneId.value) ?? panes.value[0],
  )
  const active = computed(() => {
    const p = activePane.value
    return p.tabs.find((t) => t.id === p.activeId) ?? null
  })
  const tabCount = computed(() => panes.value.reduce((n, p) => n + p.tabs.length, 0))
  const split = computed(() => panes.value.length > 1)

  function paneActive(p: TabPane): BrowserTab | null {
    return p.tabs.find((t) => t.id === p.activeId) ?? null
  }

  // 必须返回 reactive 对象：makeTab 的结果会被 push 进 panes（深层响应式数组），
  // 组件读到的是 push 后由 Vue 生成的响应式代理。若这里返回普通对象，loadBuckets 等
  // 拿到的将是原始对象，对它的写入不会触发重渲染——界面永远停在「正在加载…」，
  // 桶列表、错误信息都不会显示。reactive 保证 push 的就是同一个代理。
  function makeTab(acc: Account): BrowserTab {
    const isSP = acc.Provider === 'sharepoint'
    const isRWS = acc.Provider === 'rws'
    return reactive({
      id: `tab-${acc.ID}-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
      key: `acc:${acc.ID}`,
      title: acc.Name,
      account: acc,
      view: isRWS ? 'rws-studies' : isSP ? 'drives' : 'buckets',
      bucket: '',
      prefix: '',
      buckets: [],
      objects: [],
      truncated: false,
      loading: false,
      error: '',
      searchKeyword: '',
      searchActive: false,
      drives: [],
      driveID: '',
      driveName: '',
      itemPath: '',
      spItems: [],
      spFilterText: '',
      readOnly: false,
      shareRootPath: '',
      // RWS
      rwsStudies: [],
      rwsStudyOID: '',
      rwsStudyName: '',
      rwsSubjects: [],
      rwsSites: [],
      rwsDatasets: [],
      rwsDatasetData: null,
      rwsActiveDataset: '',
      rwsDataType: 'regular',
    })
  }


  function locate(tabId: string): { pane: TabPane; tab: BrowserTab; index: number } | null {
    for (const pane of panes.value) {
      const index = pane.tabs.findIndex((t) => t.id === tabId)
      if (index >= 0) return { pane, tab: pane.tabs[index], index }
    }
    return null
  }

  // 空窗格自动收起；若全空则回到单栏。
  function cleanupPanes() {
    if (panes.value.length <= 1) return
    const kept = panes.value.filter((p) => p.tabs.length > 0)
    if (kept.length === 0) {
      panes.value = [{ id: 'pane-0', tabs: [], activeId: null }]
      activePaneId.value = 'pane-0'
      return
    }
    panes.value = kept
    if (!kept.some((p) => p.id === activePaneId.value)) activePaneId.value = kept[0].id
  }

  // ---- 标签基本操作 ----

  function openAccount(acc: Account): BrowserTab {
    for (const p of panes.value) {
      const existing = p.tabs.find((t) => t.key === `acc:${acc.ID}`)
      if (existing) {
        p.activeId = existing.id
        activePaneId.value = p.id
        return existing
      }
    }
    const pane = activePane.value
    const tab = makeTab(acc)
    pane.tabs.push(tab)
    pane.activeId = tab.id
    if (acc.Provider === 'sharepoint') {
      void loadDrives(tab)
    } else if (acc.Provider === 'rws') {
      void loadRWSStudies(tab)
    } else {
      void loadBuckets(tab)
    }
    return tab
  }

  // 打开分享会话为一个标签（不需要解锁账号库）
  function openShareSession(sess: ShareSession): BrowserTab {
    const key = `share:${sess.Provider}:${sess.Name}:${Date.now()}`
    const pane = activePane.value
    const tab = makeTab(sess.Account)
    tab.id = `share-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`
    tab.key = key
    tab.title = `🔒 ${sess.Name}（只读）`
    tab.readOnly = true
    tab.shareSession = sess
    tab.shareRootPath = sess.ItemPath || ''
    pane.tabs.push(tab)
    pane.activeId = tab.id
    // 直接进入目标：S3 进入 bucket 对象列表，SharePoint 进入文档库内容
    if (sess.Provider === 'sharepoint') {
      tab.driveID = sess.DriveID
      tab.driveName = sess.DriveName
      tab.itemPath = sess.ItemPath
      void loadSPItems(tab)
    } else {
      tab.bucket = sess.Bucket
      tab.view = 'objects'
      void navigate(tab, '')
    }
    return tab
  }


  function selectTab(paneId: string, tabId: string) {
    activePaneId.value = paneId
    const p = panes.value.find((x) => x.id === paneId)
    if (p) p.activeId = tabId
  }

  function closeTab(tabId: string) {
    const loc = locate(tabId)
    if (!loc) return
    loc.pane.tabs.splice(loc.index, 1)
    if (loc.pane.activeId === tabId) {
      loc.pane.activeId = loc.pane.tabs.length
        ? loc.pane.tabs[Math.max(0, loc.index - 1)].id
        : null
    }
    cleanupPanes()
  }

  // duplicate 在原标签右侧复制一个内容相同的新标签并选中（重新加载，等价旧版 addTab(open)）。
  function duplicate(tabId: string) {
    const loc = locate(tabId)
    if (!loc) return
    const copy = makeTab(loc.tab.account)
    copy.title = loc.tab.title
    loc.pane.tabs.splice(loc.index + 1, 0, copy)
    loc.pane.activeId = copy.id
    activePaneId.value = loc.pane.id
    if (loc.tab.account.Provider === 'sharepoint') {
      void loadDrives(copy)
    } else if (loc.tab.account.Provider === 'rws') {
      void loadRWSStudies(copy)
    } else {
      void loadBuckets(copy)
    }
  }

  function closeOthers(tabId: string) {
    const loc = locate(tabId)
    if (!loc) return
    loc.pane.tabs = [loc.tab]
    loc.pane.activeId = tabId
    cleanupPanes()
  }

  function closeToRight(tabId: string) {
    const loc = locate(tabId)
    if (!loc) return
    loc.pane.tabs = loc.pane.tabs.slice(0, loc.index + 1)
    if (!loc.pane.tabs.some((t) => t.id === loc.pane.activeId)) loc.pane.activeId = tabId
  }

  function closeAll() {
    panes.value = [{ id: 'pane-0', tabs: [], activeId: null }]
    activePaneId.value = 'pane-0'
  }

  function move(tabId: string, delta: number) {
    const loc = locate(tabId)
    if (!loc) return
    const target = loc.index + delta
    if (target < 0 || target >= loc.pane.tabs.length) return
    const [t] = loc.pane.tabs.splice(loc.index, 1)
    loc.pane.tabs.splice(target, 0, t)
  }

  // splitRight：只有一个标签时复制（避免空窗格），否则移动；最多两栏。
  function splitRight(tabId: string) {
    const loc = locate(tabId)
    if (!loc) return
    const left = panes.value[0]
    let right = panes.value[1]
    if (!right) {
      right = { id: 'pane-1', tabs: [], activeId: null }
      panes.value = [left, right]
    }
    if (loc.pane.tabs.length <= 1) {
      const copy = makeTab(loc.tab.account)
      copy.title = loc.tab.title
      right.tabs.push(copy)
      right.activeId = copy.id
      void loadBuckets(copy)
    } else {
      loc.pane.tabs.splice(loc.index, 1)
      if (loc.pane.activeId === tabId) loc.pane.activeId = loc.pane.tabs[0]?.id ?? null
      right.tabs.push(loc.tab)
      right.activeId = loc.tab.id
    }
    activePaneId.value = right.id
  }

  // moveToLeft 把右侧窗格的标签移回左侧；右侧空了会自动收起分屏。
  function moveToLeft(tabId: string) {
    const loc = locate(tabId)
    if (!loc || loc.pane.id === panes.value[0].id) return
    loc.pane.tabs.splice(loc.index, 1)
    const left = panes.value[0]
    left.tabs.push(loc.tab)
    left.activeId = loc.tab.id
    activePaneId.value = left.id
    cleanupPanes()
  }

  // unsplit 取消分屏：把右侧窗格的所有标签合并到左侧。
  function unsplit() {
    if (panes.value.length < 2) return
    const left = panes.value[0]
    const right = panes.value[1]
    for (const t of right.tabs) left.tabs.push(t)
    left.activeId = right.activeId ?? left.activeId
    panes.value = [left]
    activePaneId.value = left.id
  }

  function setSplitOffset(v: number) {
    splitOffset.value = Math.min(0.8, Math.max(0.2, v))
  }

  // ---- 浏览动作（与 P2 一致，操作单个 BrowserTab）----

  async function loadBuckets(tab: BrowserTab) {
    tab.view = 'buckets'
    tab.loading = true
    tab.error = ''
    try {
      tab.buckets = await api.listBuckets(tab.account)
    } catch (e) {
      tab.error = errText(e)
    } finally {
      tab.loading = false
    }
  }

  async function openBucket(tab: BrowserTab, bucket: string) {
    tab.bucket = bucket
    tab.view = 'objects'
    tab.searchKeyword = ''
    tab.searchActive = false
    let saved = ''
    try {
      saved = await api.getBucketPrefix(tab.account, bucket)
    } catch {
      // 读不到上次目录就回退到根
    }
    await navigate(tab, saved || '')
  }

  async function navigate(tab: BrowserTab, prefix: string) {
    tab.view = 'objects'
    tab.prefix = prefix
    tab.searchKeyword = ''
    tab.searchActive = false
    tab.loading = true
    tab.error = ''
    try {
      const res = await api.listObjects(tab.account, tab.bucket, prefix)
      tab.objects = res.Items
      tab.truncated = res.Truncated
      try {
        await api.setBucketPrefix(tab.account, tab.bucket, prefix)
      } catch {
        // 记住目录不是关键路径，失败不影响浏览
      }
    } catch (e) {
      tab.error = errText(e)
    } finally {
      tab.loading = false
    }
  }

  async function refresh(tab: BrowserTab) {
    if (tab.view === 'buckets') return loadBuckets(tab)
    if (tab.view === 'drives') return loadDrives(tab)
    if (tab.view === 'sp-objects') return loadSPItems(tab)
    if (tab.view === 'rws-studies') return loadRWSStudies(tab)
    if (tab.view === 'rws-detail') return loadRWSStudyDetail(tab, tab.rwsStudyOID)
    if (tab.searchActive && tab.searchKeyword) return doSearch(tab, tab.searchKeyword)
    return navigate(tab, tab.prefix)
  }

  // doSearch 复刻旧版：完整路径先逐级下钻停在最后找到的目录，再用末段文件名做服务端前缀搜索。
  async function doSearch(tab: BrowserTab, raw: string) {
    tab.loading = true
    tab.error = ''
    try {
      const target = await api.resolveSearchInput(tab.account, tab.bucket, tab.prefix, raw)
      const changed = target.Bucket !== tab.bucket || target.Dir !== tab.prefix
      tab.bucket = target.Bucket
      tab.prefix = target.Dir
      tab.view = 'objects'
      tab.searchKeyword = target.Keyword
      tab.searchActive = target.IsSearch
      const res = target.IsSearch
        ? await api.searchObjects(tab.account, target.Bucket, target.Dir, target.Keyword)
        : await api.listObjects(tab.account, target.Bucket, target.Dir)
      tab.objects = res.Items
      tab.truncated = res.Truncated
      if (changed) {
        try {
          await api.setBucketPrefix(tab.account, target.Bucket, target.Dir)
        } catch {
          // 非关键
        }
      }
    } catch (e) {
      tab.error = errText(e)
    } finally {
      tab.loading = false
    }
  }

  function clearSearch(tab: BrowserTab) {
    if (!tab.searchActive && !tab.searchKeyword) return
    void navigate(tab, tab.prefix)
  }

  // ---- SharePoint 浏览动作 ----

  async function loadDrives(tab: BrowserTab) {
    tab.view = 'drives'
    tab.loading = true
    tab.error = ''
    try {
      tab.drives = await api.spListDrives(tab.account)
    } catch (e) {
      tab.error = errText(e)
    } finally {
      tab.loading = false
    }
  }

  async function openDrive(tab: BrowserTab, driveID: string, driveName: string) {
    tab.driveID = driveID
    tab.driveName = driveName
    tab.itemPath = ''
    tab.spFilterText = ''
    await loadSPItems(tab)
  }

  async function loadSPItems(tab: BrowserTab) {
    tab.view = 'sp-objects'
    tab.loading = true
    tab.error = ''
    tab.spFilterText = ''
    try {
      tab.spItems = await api.spListChildren(tab.account, tab.driveID, tab.itemPath)
    } catch (e) {
      tab.error = errText(e)
    } finally {
      tab.loading = false
    }
  }

  async function spNavigate(tab: BrowserTab, itemPath: string) {
    tab.itemPath = itemPath
    tab.spFilterText = ''
    await loadSPItems(tab)
  }

  function spGoUp(tab: BrowserTab) {
    if (tab.itemPath === '') {
      // 回到文档库列表（只读分享不允许）
      if (!tab.readOnly) void loadDrives(tab)
    } else {
      const parent = tab.itemPath.replace(/\/$/, '')
      const idx = parent.lastIndexOf('/')
      void spNavigate(tab, idx >= 0 ? parent.slice(0, idx) : '')
    }
  }

  // ---- RWS 浏览动作 ----

  async function loadRWSStudies(tab: BrowserTab) {
    tab.view = 'rws-studies'
    tab.loading = true
    tab.error = ''
    try {
      tab.rwsStudies = await api.rwsListStudies(tab.account)
    } catch (e) {
      tab.error = errText(e)
    } finally {
      tab.loading = false
    }
  }

  async function loadRWSStudyDetail(tab: BrowserTab, studyOID: string) {
    tab.view = 'rws-detail'
    tab.rwsStudyOID = studyOID
    const study = tab.rwsStudies.find(s => s.OID === studyOID)
    tab.rwsStudyName = study?.Name || studyOID
    tab.loading = true
    tab.error = ''
    tab.rwsDatasetData = null
    tab.rwsActiveDataset = ''
    try {
      const [subjects, sites, datasets] = await Promise.all([
        api.rwsListSubjects(tab.account, studyOID),
        api.rwsListSites(tab.account, studyOID),
        api.rwsListDatasets(tab.account, studyOID, tab.rwsDataType),
      ])
      tab.rwsSubjects = subjects
      tab.rwsSites = sites
      tab.rwsDatasets = datasets
    } catch (e) {
      tab.error = errText(e)
    } finally {
      tab.loading = false
    }
  }

  async function loadRWSDatasetCSV(tab: BrowserTab, dataset: string) {
    tab.rwsActiveDataset = dataset
    tab.loading = true
    tab.error = ''
    try {
      tab.rwsDatasetData = await api.rwsGetDatasetCSV(
        tab.account, tab.rwsStudyOID, tab.rwsDataType, dataset,
      )
    } catch (e) {
      tab.error = errText(e)
    } finally {
      tab.loading = false
    }
  }

  function rwsGoBack(tab: BrowserTab) {
    if (tab.view === 'rws-detail') {
      void loadRWSStudies(tab)
    }
  }

  return {
    panes,
    activePaneId,
    splitOffset,
    activePane,
    active,
    tabCount,
    split,
    paneActive,
    openAccount,
    openShareSession,
    selectTab,
    closeTab,
    duplicate,
    closeOthers,
    closeToRight,
    closeAll,
    move,
    splitRight,
    moveToLeft,
    unsplit,
    setSplitOffset,
    loadBuckets,
    openBucket,
    navigate,
    refresh,
    doSearch,
    clearSearch,
    // SharePoint
    loadDrives,
    openDrive,
    loadSPItems,
    spNavigate,
    spGoUp,
    // RWS
    loadRWSStudies,
    loadRWSStudyDetail,
    loadRWSDatasetCSV,
    rwsGoBack,
  }
})
