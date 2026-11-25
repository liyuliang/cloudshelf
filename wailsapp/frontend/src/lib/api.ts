// Go 绑定的薄封装：集中一处，方便后续阶段扩展与替换。
// 全部方法由 wails generate module 自动生成，请勿手改 wailsjs/ 目录。
import * as App from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'

export type Account = main.AccountDTO
export type ProxyConfig = main.ProxyDTO
export type ProxyMode = main.ProxyModeDTO
export type Bucket = main.BucketDTO
export type S3Object = main.ObjectDTO
export type ObjectList = main.ObjectListDTO
export type SearchTarget = main.SearchTargetDTO
export type TransferItem = main.TransferItemDTO
export type TransferResult = main.TransferResultDTO
export type Drive = main.DriveDTO
export type SPItem = main.SPItemDTO
export type SPTransferItem = main.SPTransferItemDTO
export type ShareSession = main.ShareSessionDTO
export type RWSStudy = main.RWSStudyDTO
export type RWSSubject = main.RWSSubjectDTO
export type RWSSite = main.RWSSiteDTO
export type RWSDataset = main.RWSDatasetDTO
export type RWSDatasetData = main.RWSDatasetDataDTO

// 进度事件载荷：只通过 EventsEmit 推送，不出现在任何绑定方法签名里，
// 所以 Wails 不会为它生成 TS 类，需要手工定义（字段与 Go 的 TransferProgressDTO 一致）。
// LAN 分享 DTO（手工定义，字段与 Go 的 LANShareDTO JSON tag 一致）
export interface LANShare {
  token: string
  name: string
  provider?: string
  bucket: string
  prefix: string
  driveId?: string
  studyOid?: string
  url: string
  urls: string[]
  createdAt: string
  expiresAt: string
  remaining: string
}

export interface TransferProgress {
  ID: string
  Name: string
  Done: number
  Total: number
  Percent: number
  Status: string
  Error: string
}

export function emptyAccount(): Account {
  return {
    ID: 0,
    Name: '',
    Provider: 's3',
    Endpoint: '',
    Region: '',
    AccessKeyID: '',
    SecretAccessKey: '',
    Bucket: '',
    UsePathStyle: false,
    TenantID: '',
    ClientID: '',
    ClientSecret: '',
    SiteURL: '',
  }
}

export function emptyProxy(): ProxyConfig {
  return { Mode: 'env', Host: '', Port: 0, Username: '', Password: '', NoProxy: '' }
}

export const api = {
  // 会话门
  isInitialized: () => App.IsInitialized(),
  unlock: (pw: string) => App.Unlock(pw),
  lock: () => App.Lock(),
  verifyMasterPassword: (pw: string) => App.VerifyMasterPassword(pw),
  changeMasterPassword: (pw: string) => App.ChangeMasterPassword(pw),
  currentDBPath: () => App.CurrentDBPath(),
  changeDBLocation: (p: string) => App.ChangeDBLocation(p),

  // 加密导入导出 + 文件对话框
  exportAccounts: (pw: string) => App.ExportAccounts(pw),
  importAccounts: (pw: string, b64: string) => App.ImportAccounts(pw, b64),
  pickSaveFile: (title: string, name: string) => App.PickSaveFile(title, name),
  pickFile: (title: string) => App.PickFile(title),
  readFileB64: (p: string) => App.ReadFileB64(p),
  writeFileB64: (p: string, b64: string) => App.WriteFileB64(p, b64),
  copyText: (t: string) => App.CopyText(t),

  // 账号
  listAccounts: () => App.ListAccounts(),
  addAccount: (a: Account) => App.AddAccount(a),
  updateAccount: (a: Account) => App.UpdateAccount(a),
  deleteAccount: (id: number) => App.DeleteAccount(id),
  testAccount: (a: Account) => App.TestAccount(a),

  // S3 浏览（P2）
  listBuckets: (a: Account) => App.ListBuckets(a),
  listObjects: (a: Account, bucket: string, prefix: string) => App.ListObjects(a, bucket, prefix),
  searchObjects: (a: Account, bucket: string, prefix: string, kw: string) =>
    App.SearchObjects(a, bucket, prefix, kw),
  resolveSearchInput: (a: Account, bucket: string, prefix: string, raw: string) =>
    App.ResolveSearchInput(a, bucket, prefix, raw),
  getBucketPrefix: (a: Account, bucket: string) => App.GetBucketPrefix(a, bucket),
  setBucketPrefix: (a: Account, bucket: string, prefix: string) =>
    App.SetBucketPrefix(a, bucket, prefix),
  presignURL: (a: Account, bucket: string, key: string, minutes: number) =>
    App.PresignURL(a, bucket, key, minutes),
  deleteObject: (a: Account, bucket: string, key: string) => App.DeleteObject(a, bucket, key),
  createFolder: (a: Account, bucket: string, prefix: string, name: string) =>
    App.CreateFolder(a, bucket, prefix, name),
  pickDirectory: (title: string) => App.PickDirectory(title),
  pickMultipleFiles: (title: string) => App.PickMultipleFiles(title),
  uploadFile: (a: Account, bucket: string, key: string, localPath: string) =>
    App.UploadFile(a, bucket, key, localPath),
  downloadFile: (a: Account, bucket: string, key: string, localPath: string) =>
    App.DownloadFile(a, bucket, key, localPath),
  // P3 四路并发（替代上面两个单文件方法）
  uploadFiles: (a: Account, bucket: string, prefix: string, items: TransferItem[]) =>
    App.UploadFiles(a, bucket, prefix, items),
  downloadFiles: (a: Account, bucket: string, dir: string, items: TransferItem[]) =>
    App.DownloadFiles(a, bucket, dir, items),
  revealInExplorer: (p: string) => App.RevealInExplorer(p),

  // SharePoint 浏览（P4）
  spListDrives: (a: Account) => App.SPListDrives(a),
  spListChildren: (a: Account, driveID: string, itemPath: string) =>
    App.SPListChildren(a, driveID, itemPath),
  spCreateFolder: (a: Account, driveID: string, parentPath: string, name: string) =>
    App.SPCreateFolder(a, driveID, parentPath, name),
  spDeleteItem: (a: Account, driveID: string, itemID: string) =>
    App.SPDeleteItem(a, driveID, itemID),
  spUploadFiles: (a: Account, driveID: string, itemPath: string, items: SPTransferItem[]) =>
    App.SPUploadFiles(a, driveID, itemPath, items),
  spDownloadFiles: (a: Account, driveID: string, dir: string, items: SPTransferItem[]) =>
    App.SPDownloadFiles(a, driveID, dir, items),

  // 分享配置（P4）
  openShareFile: (dataB64: string, password: string) => App.OpenShareFile(dataB64, password),
  needSharePassword: (dataB64: string) => App.NeedSharePassword(dataB64),
  createS3Share: (a: Account, bucket: string, name: string, hours: number, password: string) =>
    App.CreateS3Share(a, bucket, name, hours, password),
  createSharePointShare: (
    a: Account, driveID: string, driveName: string, itemPath: string,
    name: string, hours: number, password: string,
  ) => App.CreateSharePointShare(a, driveID, driveName, itemPath, name, hours, password),

  // RWS 浏览
  rwsListStudies: (a: Account) => App.RWSListStudies(a),
  rwsListSubjects: (a: Account, studyOID: string) => App.RWSListSubjects(a, studyOID),
  rwsListSites: (a: Account, studyOID: string) => App.RWSListSites(a, studyOID),
  rwsListDatasets: (a: Account, studyOID: string) => App.RWSListDatasets(a, studyOID),
  rwsGetDatasetCSV: (a: Account, studyOID: string, typ: string, dataset: string) =>
    App.RWSGetDatasetCSV(a, studyOID, typ, dataset),
  rwsProbeDatasetTypes: (a: Account, studyOID: string, datasets: string[]) =>
    App.RWSProbeDatasetTypes(a, studyOID, datasets),

  // 局域网分享
  createLANShare: (a: Account, bucket: string, prefix: string, name: string, hours: number) =>
    App.CreateLANShare(a, bucket, prefix, name, hours),
  createSharePointLANShare: (a: Account, driveID: string, driveName: string, itemPath: string, name: string, hours: number) =>
    App.CreateSharePointLANShare(a, driveID, driveName, itemPath, name, hours),
  createRWSLANShare: (a: Account, studyOID: string, studyName: string, subPath: string, name: string, hours: number) =>
    App.CreateRWSLANShare(a, studyOID, studyName, subPath, name, hours),
  listLANShares: () => App.ListLANShares(),
  removeLANShare: (token: string) => App.RemoveLANShare(token),
  getLANAddresses: () => App.GetLANAddresses(),
  getLANSharePort: () => App.GetLANSharePort(),
  copyToClipboard: (text: string) => App.CopyToClipboard(text),

  // 代理
  proxyModes: () => App.ProxyModes(),
  getProxy: () => App.GetProxy(),
  setProxy: (c: ProxyConfig) => App.SetProxy(c),
  validateProxy: (c: ProxyConfig) => App.ValidateProxy(c),
  checkProxy: (c: ProxyConfig, url: string, timeoutSec: number) =>
    App.CheckProxy(c, url, timeoutSec),
}

export function errText(e: unknown): string {
  if (typeof e === 'string') return e
  if (e instanceof Error) return e.message
  return String(e)
}
