# 功能对照表：Fyne 版 → Wails 版

**验收原则**：「全部功能保持不变」逐行打勾，任何一行从 `✓` 退化都要先修再推进阶段。
旧 Fyne 版（根目录 `s3client.exe`）全程保留、可同时运行，用作 A/B 对照：同一 `~/.s3client/data.db` 与 `.s3share` 夹具在两个版本上执行相同操作，比对结果。

> 更名注：项目现为 **CloudShelf**（原 s3client）。新版数据目录为 `~/.cloudshelf`（首次启动由 `~/.s3client` 自动迁移），分享文件扩展名改为 `.cloudshelf`（旧 `.s3share` 仍可打开）；下文历史对照中的 `s3client.exe` / `~/.s3client` 指旧版产物与旧版目录。

状态：`✓` 已接入（标注阶段） · `P2/P3/P4/P5` 计划阶段 · `保留` 不受 UI 迁移影响

| # | 功能 | 旧版位置 | Wails facade | 前端组件 | 阶段 | 状态 |
|---|---|---|---|---|---|---|
| 1 | Windows 桌面窗口 | `ui.Run` | `main.go` wails.Run | — | P0 | ✓ |
| 2 | S3 账号管理，SQLite 加密存储、密码不明文 | `store.*` | `ListAccounts/AddAccount/UpdateAccount/DeleteAccount` | `AccountList/AccountEditDialog` | P1 | ✓ |
| 3 | 纯 Go CLI 入口（cmd/cli） | `cmd/cli` | 不适用 | — | — | 保留 |
| 4 | 双击账号左右分栏，右侧 Bucket 列表 | `showMain/showBuckets` | `ListBucketsWithAccess`(P2) | `TabWorkspace/BucketList` | P2 | ✓ |
| 5 | 双击 Bucket 列出文件与文件夹、可下钻 | `buildObjectView` | `ListObjects/ResolvePathDir`(P2) | `ObjectBrowser` | P2 | ✓ |
| 6 | 底部操作栏：刷新/上传/下载/删除/预签名/新建文件夹 | `buildObjectView` 操作栏 | `PresignURL/DeleteObject/CreateFolder` + `UploadFiles/DownloadFiles` | `ObjectBrowser` | P2/P3 | ✓ |
| 7 | 面包屑点击任意一段跳转 | `buildBreadcrumb` | `SetBucketPrefix`(P2) | `ObjectBrowser` | P2 | ✓ |
| 8 | 名称模糊筛选（S3 服务端前缀 / SharePoint 本地过滤） | `buildObjectView`/`buildSharePointView` | `SearchObjects` / SP 前端本地过滤 | `ObjectBrowser/SharePointBrowser` | P2/P4 | ✓ |
| 9 | 文件列表 大小/类型/修改时间 + 列头 | `newMetaCols` | 复用 `awss3.Object` | `ObjectBrowser` | P2 | ✓ |
| 10 | 文件列表按修改时间倒序 | `sortObjects/sortItems` | 后端已具备，直接复用 | `ObjectBrowser` | P2 | ✓ |
| 11 | 删除文件二次确认 | `buildObjectView` 双 confirm | — | `ObjectBrowser` | P2 | ✓ |
| 12 | 预签名 URL 自定义过期分钟数、随输入实时刷新 | `urlBtn` | `PresignURL(minutes)` | `PresignDialog` | P2 | ✓ |
| 13 | 界面美化：右对齐元数据、圆角选中高亮、按钮图标 | `tappableRow` | — | `style.css` 设计系统 | P1 | ✓ |
| 14 | Bucket 列表 HeadBucket 权限检查，可访问/无权限 | `ListBucketsWithAccess` | 同名方法(P2) | `BucketList` | P2 | ✓ |
| 15 | 可访问绿色可双击 / 无权限灰色禁止 | `showBuckets` | — | `BucketList` | P2 | ✓ |
| 16 | 对象浏览器底部可滚动错误消息区 | `msgScroll` | — | `ObjectBrowser` | P2 | ✓ |
| 17 | 系统原生文件对话框（zenity → Wails runtime） | `ui.go` zenity 9 处 | `PickFile/PickSaveFile/PickDirectory/PickMultipleFiles` | `SimpleDialogs/ObjectBrowser` | P1/P2 | ✓ |
| 18 | 下载完成打开系统文件管理器并选中 | `revealInExplorer` | `RevealInExplorer` | — | P2 | ✓ |
| 19 | 右键查看 Bucket/文件夹/文件 属性，值可复制 | `showProperties` | `CopyText` | `PropertiesDialog` | P2 | ✓ |
| 20 | 解锁后菜单栏：加锁、修改 SQLite 存储位置 | `showMain` 菜单 | `Lock/ChangeDBLocation` | `MenuBar/SimpleDialogs` | P1 | ✓ |
| 21 | 主密码回车解锁、密码框与按钮并列居中 | `showUnlock` | `Unlock/IsInitialized` | `UnlockGate` | P1 | ✓ |
| 22 | 深色主题 | `appTheme` | — | `style.css` | P1 | ✓ |
| 23 | exe 文件图标 + 窗口图标 | `resource_windows.syso`/`icon.ico` | `build/windows/icon.ico` | — | P0 | ✓ |
| 24 | 修改主密码（Argon2id 派生 + 事务重加密） | `ChangeMasterPassword` | 同名方法 | `SimpleDialogs` | P1 | ✓ |
| 25 | 账号区左侧约 30% 宽度 | `HSplit(0.3)` | — | `.sidebar` | P1 | ✓ |
| 26 | 信息提示弹窗约 300px 长方形 | `showInfo` | — | `Modal`/`.toast` | P1 | ✓ |
| 27 | GUI 通过 CGO 编译验证 | Makefile `build` | Wails Windows 纯 Go，CGO 非必需 | — | P0 | ✓（`wails doctor`） |
| 28 | 跨平台 Win/Mac/Linux | Makefile 三平台 | Wails 原生支持（`wails build` 按宿主 OS 产出） | — | P5 | ✓ |
| 29 | 加密导出/导入账号（Argon2id + AES-GCM 自包含文件） | `ExportAccounts/ImportAccounts` | 同名方法 + base64 | `SimpleDialogs` | P1 | ✓ |
| 30 | 记住每个 bucket 上次目录（meta 表，按账号+bucket 隔离） | `Get/SetBucketPrefix` | 同名方法(P2) | `ObjectBrowser` | P2 | ✓ |
| 31 | 多文件并行上传/下载（多选 + 4 并发限流） | `runParallelTransfers` | `UploadFiles/DownloadFiles`（`parallelRun` 限流 4） | `ObjectBrowser/TransferBar` | P3 | ✓ |
| 32 | 进度条在底部区域，每文件一行 文件名+百分比+条 | `transferRow` | `transfer:progress` 事件（节流 1%/100ms） | `TransferBar` | P3 | ✓ |
| 33 | 文件行左键点中自动勾选复选框 | `Tapped` | — | `ObjectBrowser` | P2 | ✓ |
| 34 | 菜单栏按 账号/安全/存储 分组 | `showMain` 菜单 | — | `MenuBar` | P1 | ✓ |
| 35 | 账号行右键上下文菜单（编辑/删除/测试连接） | `withSecondaryMenu` | `UpdateAccount/DeleteAccount/TestAccount` | `AccountList` | P1 | ✓ |
| 36 | 账号类型图标 / SharePoint 账号 `sp` 标识 | `providerIcon`/`accountDisplayName` | `AccountDTO.Provider` | `AccountList` | P1 | ✓ |
| 37 | 账号编辑表单 S3 / SharePoint 双类型切换 | `editAccount` | `AddAccount/UpdateAccount` | `AccountEditDialog` | P1 | ✓ |
| 38 | 测试连接按 Provider 分流 | `doTest` | `TestAccount` | `MenuBar/AccountList` | P1 | ✓ |
| 39 | SharePoint Graph 文档库访问（OAuth2/ListDrives/ListChildren/DownloadItem） | `internal/sharepoint` | `SPListDrives/SPListChildren/SPUploadFiles/SPDownloadFiles/SPCreateFolder/SPDeleteItem` | `SharePointDriveList/SharePointBrowser` | P4 | ✓ |
| 40 | SharePoint 文件浏览器（面包屑/搜索/下载进度/建夹/删） | `buildSharePointView` | 同上 + 前端本地名称过滤 | `SharePointBrowser` | P4 | ✓ |
| 41 | 标签页 JetBrains 全套（复制/关闭/关闭其他/关闭右侧/关闭所有/分屏/拖动） | `tabs.go` | — | `stores/browser.ts` 双栏 panes + `TabWorkspace` | P3 | ✓ |
| 42 | 网络代理设置（4 模式/地址端口/账号密码/NO_PROXY/测试连接） | `settings.go` | `Get/Set/Validate/CheckProxy/ProxyModes` | `ProxyDialog` | P1 | ✓ |
| 43 | 配置存 `~/.cloudshelf/config.json`(0600)，GUI/CLI 共用，保存立即生效 | `internal/proxy` | 原样复用 | — | P1 | ✓ |
| 44 | S3 列表截断提示 + 服务端前缀搜索（不分页） | `SearchObjects/ListObjects` | 同名方法(P2) | `ObjectBrowser` | P2 | ✓ |
| 45 | 完整路径解析：按 `/` 数量逐级下钻，停在最后找到的目录 | `parseObjectPath/ResolvePathDir` | `ResolvePathDir`(P2) | `ObjectBrowser` | P2 | ✓ |
| 46 | 只读分享生成（S3/SharePoint，口令/免口令/有效期） | `share.Create/CreateSharePoint` | `CreateS3Share/CreateSharePointShare` | `CreateShareDialog` | P4 | ✓ |
| 47 | 解锁前加载分享配置、只读会话、有效期回显、写操作拦截 | `loadShare/openShare` | `OpenShareFile/NeedSharePassword` | `UnlockGate`（免口令/口令弹窗）、`SharePointBrowser`/`ObjectBrowser`（只读 + 过期回显） | P4 | ✓ |
| 48 | 关闭窗口二次确认 | `SetCloseIntercept` | `OnBeforeClose`（关闭→最小化到托盘；托盘「退出」二次确认） | — | P5 | ✓ |

## 阶段汇总

| 阶段 | 覆盖行 | 状态 |
|---|---|---|
| P0 脚手架 | 1, 23, 27 | ✓ 已交付 |
| P1 门 + 账号 + 设置 | 2, 13, 20–22, 24–26, 29, 34–38, 42, 43 | ✓ 已交付 |
| P2 S3 浏览 | 4–19, 30, 33, 44, 45 | ✓ 本次交付 |
| P3 传输 + 分屏 | 6, 31, 32, 41 | ✓ 本次交付 |
| P4 SharePoint + 分享 | 8(SP), 39, 40, 46, 47 | ✓ 本次交付 |
| P5 收尾 | 28, 48 | ✓ 本次交付 |

## 数据兼容闸门（每阶段必跑）

`wailsapp/backend_compat_test.go`（6 项，全绿）固定以下不可变格式：
- `accounts` 表 12 列及顺序、`meta` 表
- scrypt `N=32768,r=8,p=1`，salt/key 32B，verifier 明文
- AES-256-GCM `nonce(12)||ct||tag(16)` + 标准 base64
- 导出 / `.cloudshelf`（旧 `.s3share`）外层 JSON：`version=1, kdf=argon2id, argon_time=3, argon_memory=65536, argon_threads=4, salt, payload`
- `config.json` JSON 键与非 Windows 0600
- facade 的 `AccountDTO`/`ProxyDTO` 往返不丢字段

E2E：用现有 `~/.cloudshelf/data.db`（由 `~/.s3client` 自动迁移而来）以旧主密码在新版解锁 → 账号逐项比对；导出→导入数量一致；现有 `.cloudshelf`/`.s3share`（需口令/免口令）打开结果一致；旧 exe 同夹具 A/B 对照。
