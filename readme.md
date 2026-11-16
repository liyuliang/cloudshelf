# CloudShelf

一个桌面端多存储文件管理器：集中管理 S3 兼容存储（AWS S3、MinIO 等）与 SharePoint 文档库。使用 Go + Wails v2 + Vue 构建，提供 GUI 和 CLI 两种入口。

> 项目前身为 `s3client`，现更名为 **CloudShelf**：go module、二进制名、界面品牌、数据目录（`~/.s3client` → `~/.cloudshelf`，首次启动自动迁移）全部随之更新；主密码、账号数据、已有的 `.s3share` 分享文件保持兼容。

## S3 存储功能清单

> **设计原则**：以下功能为每个数据源分类的**基准能力**，后续新增的数据源分类（如 SharePoint）应保持与 S3 一致的功能数量和能力覆盖。

### 账号与安全
| # | 功能 | 说明 |
|---|------|------|
| 1 | **账号管理** | 新增、编辑、删除账号，支持自定义 Endpoint、Region、Path-Style |
| 2 | **密码安全存储** | Secret Access Key 经 AES-256-GCM 加密后存入 SQLite（`~/.cloudshelf/data.db`），密钥由主密码经 scrypt 派生，**绝不明文存储** |
| 3 | **主密码保护** | 首次使用设置主密码，后续启动需输入主密码解锁 |
| 4 | **加密导出/导入账号** | Argon2id + AES-256-GCM，可跨机器迁移 |
| 5 | **测试连接** | 账号区右键菜单一键测试 S3 凭证是否有效（ListBuckets） |

### Bucket 浏览
| # | 功能 | 说明 |
|---|------|------|
| 6 | **Bucket 列表** | 双击账号查看其下所有 Bucket；每个账号一个 tab，已打开的账号再次双击只是切回该 tab |
| 7 | **Bucket 权限检查** | 双击账号后用 HeadBucket（回退 ListObjectsV2）并发检查每个 Bucket 的访问权限；可访问的桶名绿色可双击进入，无权限的灰色禁止双击 |
| 8 | **右键查看属性** | Bucket / 文件夹 / 文件均可右键查看属性，属性值可 Ctrl+C 复制 |

### 对象浏览
| # | 功能 | 说明 |
|---|------|------|
| 9 | **对象浏览器** | 双击 Bucket 列出文件和文件夹，文件夹可继续下钻 |
| 10 | **记住上次目录** | 按「账号 + Bucket」记住上次打开的文件夹路径，再次点击该 Bucket 直接进入上次所在目录 |
| 11 | **可点击面包屑路径** | 路径每一段可点击，直接跳转到对应目录 |
| 12 | **本地模糊过滤** | 输入关键字实时筛选当前目录下的文件和文件夹 |
| 13 | **服务端前缀搜索** | 按名称前缀在服务端搜索，支持跨子目录命中（S3 只支持前缀匹配） |
| 14 | **文件元数据** | 文件列表显示大小、类型、最后修改时间，并带列头排序 |

### 文件操作
| # | 功能 | 说明 |
|---|------|------|
| 15 | **上传文件** | 原生对话框一次多选多个文件上传 |
| 16 | **下载文件** | 文件行左侧勾选复选框多选，再选目标目录批量下载 |
| 17 | **递归下载文件夹** | 勾选文件夹则递归下载其下全部文件并保留子目录结构 |
| 18 | **多文件并行上传/下载** | 信号量限流 4 并发，布局底部每个文件一行实时进度（文件名 + 百分比 + 进度条） |
| 19 | **删除文件/文件夹** | 连续两次确认，防止误删 |
| 20 | **新建文件夹** | 在当前目录下创建空文件夹（S3 以 / 结尾的空对象表示） |
| 21 | **系统原生文件对话框** | 上传选文件、下载选保存路径均使用操作系统原生对话框 |
| 22 | **下载后打开文件夹** | 下载完成自动在系统文件管理器中打开并选中文件 |
| 23 | **可调过期的预签名 URL** | 自定义过期分钟数（默认 10 分钟），URL 随输入实时刷新 |

### 标签与分屏
| # | 功能 | 说明 |
|---|------|------|
| 24 | **标签页右键菜单（类 JetBrains）** | 复制标签页 / 关闭 / 关闭其他 / 关闭右侧所有 / 关闭所有 / 向右分割并移动 / 移动到左侧分割 / 左移·右移调整顺序 |
| 25 | **标签关闭方式** | × 按钮或鼠标中键关闭 |
| 26 | **标签条右键菜单** | 右键标签条空白处可「关闭所有标签页 / 取消分屏」 |
| 27 | **双栏分屏** | 分屏中间的分隔条可拖动调整宽度 |
| 28 | **标签名省略** | 标签过长自动省略号截断 |

### 分享与协作
| # | 功能 | 说明 |
|---|------|------|
| 29 | **只读分享** | 生成 `.cloudshelf` 分享配置（兼容旧版 `.s3share`）；口令自动生成 8 位随机字符、可改、留空即免口令；有效期按小时设置，0 为不限期 |
| 30 | **分享加密** | 桶名与访问凭据用分享口令经 Argon2id + AES-256-GCM 加密 |
| 31 | **加载分享配置** | 解锁界面可直接「加载分享配置（无需解锁）」只读浏览；解锁后也可从菜单加载 |
| 32 | **只读保护** | 界面不提供上传/删除/新建文件夹入口，`awss3` 层对只读账号另有写入拦截兜底，过期后拒绝打开与下载 |

### 其他
| # | 功能 | 说明 |
|---|------|------|
| 33 | **账号类型图标** | 账号列表每行名称后方用小图标标注类型（S3 / SharePoint） |
| 34 | **网络代理设置** | 支持跟随系统环境变量 / 不使用代理 / HTTP·HTTPS 代理 / SOCKS5 代理，含测试连接 |
| 35 | **菜单栏（按功能分类）** | 解锁后菜单栏分为五组——「账号」「分享」「设置」「安全」「存储」 |
| 36 | **关闭最小化到系统托盘** | 点关闭按钮不退出，窗口隐藏缩到托盘；Windows / macOS / Linux 三平台行为一致 |
| 37 | **CLI 入口** | 纯 Go 命令行工具，无需 CGO，支持 list/add/del/test/buckets 命令 |

---

## 技术栈

| 组件 | 技术 |
|------|------|
| GUI | [Wails v2](https://wails.io/) + [Vue 3](https://vuejs.org/) |
| 系统托盘 | [fyne.io/systray](https://github.com/fyne-io/systray)（Windows / macOS / Linux） |
| S3 | [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2) |
| 数据库 | [modernc.org/sqlite](https://modernc.org/sqlite)（纯 Go，无需 CGO） |
| 加密 | AES-256-GCM + [scrypt](https://pkg.go.dev/golang.org/x/crypto/scrypt)（内部）/ [Argon2id](https://pkg.go.dev/golang.org/x/crypto/argon2)（导出文件）密钥派生 |
| CLI 密码输入 | [golang.org/x/term](https://pkg.go.dev/golang.org/x/term) |

## 目录结构

```
wailsapp/
  main.go                   Wails GUI 入口
  app.go                    Go 绑定层（facade）
  tray.go                   系统托盘（关闭最小化到托盘、恢复窗口与退出）
  facade_*.go               各业务模块的前端绑定
  frontend/                 Vue 3 前端
  build/                    构建资源（图标、Windows 清单等）
cmd/cli/main.go             CLI 入口（纯 Go，无需 CGO）
internal/
  crypto/crypto.go          AES-256-GCM 加解密 + scrypt 密钥派生
  model/account.go          账号数据模型
  store/store.go            SQLite 存储层（主密码管理 + 账号 CRUD）
  appdir/appdir.go          数据目录 ~/.cloudshelf（旧 ~/.s3client 首次启动自动迁移）
  awss3/client.go           S3 操作封装（列桶/列对象/上传/下载/删除/预签名/新建文件夹）
  proxy/proxy.go            网络代理配置（HTTP/SOCKS5/环境变量）
  sharepoint/client.go      SharePoint Graph REST 客户端
  share/share.go            只读分享配置（.cloudshelf，兼容旧 .s3share）的生成与口令解密加载
```

## 构建与运行

需要安装 [Wails CLI](https://wails.io/docs/gettingstarted/installation)：`go install github.com/wailsapp/wails/v2/cmd/wails@latest`

```bash
make build       # 构建 Wails 桌面应用（产出 wailsapp/build/bin/cloudshelf.exe）
make build-cli   # 构建纯 Go CLI（跨平台，无需 CGO）
make dev         # 启动开发模式（Go + Vue 热重载）
make run         # 运行已构建的桌面应用
```

### 手动构建（不使用 make）

```bash
# GUI
cd wailsapp && wails build

# CLI —— 任意平台
CGO_ENABLED=0 go build -o cloudshelf-cli ./cmd/cli
```

CLI 命令：`list` | `add` | `del <id>` | `test <id>` | `buckets <id>` | `help` | `exit`（交互提示符 `cloudshelf>`）

## 安全设计

- Secret Access Key **不以明文存储**
- 主密码经 scrypt（N=32768, r=8, p=1）派生 32 字节密钥
- 每次加密使用随机 12 字节 nonce + AES-256-GCM
- 主密码本身不落库，仅存加密校验器用于解锁验证
- 数据库文件位于 `~/.cloudshelf/data.db`，权限 0700（由旧版 `~/.s3client` 首次启动自动迁移而来）
- **代理配置**存于 `~/.cloudshelf/config.json`（权限 0600），代理账号/密码为**明文**
- **账号导出文件**：用主密码经 Argon2id 派生独立密钥，AES-256-GCM 加密；文件自包含
