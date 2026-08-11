import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const root = process.cwd()
const catalogPath = resolve(root, 'src/constants/dingtalkH5Icons.ts')
const iconfontPath = resolve(root, 'src/styles/uview-iconfont.css')
const menuPath = resolve(root, 'src/views/menu/index.vue')

function fail(message) {
  console.error(`[check-dingtalk-h5-icons] ${message}`)
  process.exit(1)
}

if (!existsSync(catalogPath)) {
  fail('缺少钉钉 H5 uView 图标目录 src/constants/dingtalkH5Icons.ts')
}

if (!existsSync(iconfontPath)) {
  fail('缺少 uView 图标字体样式 src/styles/uview-iconfont.css')
}

const catalogSource = readFileSync(catalogPath, 'utf8')
const iconfontSource = readFileSync(iconfontPath, 'utf8')
const menuSource = readFileSync(menuPath, 'utf8')

const namesBlock = catalogSource.match(/DINGTALK_H5_ICON_NAMES\s*=\s*\[([\s\S]*?)\]\s*as const/)
if (!namesBlock) {
  fail('图标目录需要导出 DINGTALK_H5_ICON_NAMES')
}

const iconNames = [...namesBlock[1].matchAll(/'([^']+)'/g)].map(match => match[1])
if (iconNames.length < 120) {
  fail(`图标目录数量过少，当前 ${iconNames.length} 个，期望至少 120 个`)
}

for (const required of ['grid', 'list', 'file-text', 'order', 'account-fill', 'man-add', 'setting', 'home', 'search', 'download', 'trash']) {
  if (!iconNames.includes(required)) {
    fail(`图标目录缺少 ${required}`)
  }
}

if (!iconfontSource.includes("font-family: 'uicon-iconfont'") || !iconfontSource.includes('.uicon-grid:before')) {
  fail('uView 图标字体样式不完整，无法真实预览图标')
}

if (menuSource.includes('dingtalkH5IconShort')) {
  fail('权限管理页面仍在使用文字短标预览钉钉 H5 图标')
}

for (const snippet of ['DINGTALK_H5_ICON_OPTIONS', 'dingtalkH5FilteredIconOptions', 'permission-h5-real-icon']) {
  if (!menuSource.includes(snippet)) {
    fail(`权限管理页面缺少 ${snippet}`)
  }
}

console.log(`[check-dingtalk-h5-icons] ok: ${iconNames.length} icons`)
