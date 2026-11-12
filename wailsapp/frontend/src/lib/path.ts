// 与 internal/ui/ui.go 的纯函数保持一致（旧版为行为真源）。
import type { S3Object, SPItem } from './api'

export function humanSize(n: number): string {
  const unit = 1024
  if (n < unit) return `${n} B`
  let div = unit
  let exp = 0
  for (let x = Math.floor(n / unit); x >= unit; x = Math.floor(x / unit)) {
    div *= unit
    exp++
  }
  return `${(n / div).toFixed(1)} ${'KMGTPE'[exp]}B`
}

export function parentPrefix(prefix: string): string {
  let trimmed = prefix
  if (trimmed.length > 0 && trimmed.endsWith('/')) trimmed = trimmed.slice(0, -1)
  for (let i = trimmed.length - 1; i >= 0; i--) {
    if (trimmed[i] === '/') return trimmed.slice(0, i + 1)
  }
  return ''
}

export interface Crumb {
  label: string
  prefix: string
}

export function breadcrumbs(prefix: string): Crumb[] {
  const parts = prefix.replace(/\/$/, '').split('/').filter(Boolean)
  return parts.map((label, i) => ({ label, prefix: parts.slice(0, i + 1).join('/') + '/' }))
}

export function isFolder(o: S3Object): boolean {
  return o.IsDir
}

export function iconFor(o: S3Object): string {
  if (o.IsDir) return '📁'
  const ext = (o.Name.split('.').pop() || '').toLowerCase()
  if (['tgz', 'gz', 'zip', 'rar', '7z', 'tar'].includes(ext)) return '🗜'
  if (['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'bmp'].includes(ext)) return '🖼'
  if (['txt', 'md', 'json', 'yaml', 'yml', 'log', 'csv'].includes(ext)) return '📄'
  return '📦'
}

export function spIconFor(it: SPItem): string {
  if (it.IsFolder) return '📁'
  const ext = (it.Name.split('.').pop() || '').toLowerCase()
  if (['tgz', 'gz', 'zip', 'rar', '7z', 'tar'].includes(ext)) return '🗜'
  if (['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'bmp'].includes(ext)) return '🖼'
  if (['txt', 'md', 'json', 'yaml', 'yml', 'log', 'csv'].includes(ext)) return '📄'
  if (['docx', 'doc', 'xlsx', 'xls', 'pptx', 'ppt', 'pdf'].includes(ext)) return '📄'
  return '📦'
}

// SharePoint 路径工具（复刻 internal/ui/ui.go 的 spParentPath / spWithinRoot）

export function spParentPath(p: string): string {
  const trimmed = p.replace(/\/$/, '')
  const idx = trimmed.lastIndexOf('/')
  if (idx >= 0) return trimmed.slice(0, idx)
  return ''
}

export function spWithinRoot(path: string, root: string): boolean {
  if (!root) return true
  return path === root || path.startsWith(root + '/')
}

export interface SPCrumb {
  label: string
  path: string
}

export function spBreadcrumbs(itemPath: string): SPCrumb[] {
  const parts = itemPath.replace(/\/$/, '').split('/').filter(Boolean)
  return parts.map((label, i) => ({ label, path: parts.slice(0, i + 1).join('/') }))
}

