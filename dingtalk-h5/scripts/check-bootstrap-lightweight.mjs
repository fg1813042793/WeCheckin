import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = resolve(fileURLToPath(new URL('..', import.meta.url)))
const pageSource = readFileSync(resolve(root, 'pages/index/index.vue'), 'utf8')
const componentSources = [
  'components/performance/OrgView.vue',
  'components/performance/SummaryView.js',
  'components/performance/TemplateView.js',
  'components/performance/WorkbenchView.js'
].map((path) => readFileSync(resolve(root, path), 'utf8')).join('\n')

const match = pageSource.match(/async function loadBootstrap\(\) \{([\s\S]*?)\n\}/)
if (!match) {
  throw new Error('pages/index/index.vue 必须保留 loadBootstrap 启动函数')
}

const body = match[1]
for (const snippet of ['res.data.users', 'res.data.reviews', 'res.data.template', 'dingTalkPerformanceApi.users', 'dingTalkPerformanceApi.reviews', 'dingTalkPerformanceApi.template']) {
  if (body.includes(snippet)) {
    throw new Error(`loadBootstrap 只能获取当前用户启动态，不能读取全量数据: ${snippet}`)
  }
}

for (const marker of ['async function refreshData', 'async function loadUsers', 'async function loadReviews', 'async function loadTemplate']) {
  if (!pageSource.includes(marker)) {
    throw new Error(`pages/index/index.vue 必须使用按需刷新函数: ${marker}`)
  }
}

for (const marker of ['permissionVersion: 0', 'apiPermissionKeys: []', 'apiPermissionReady: false', 'applySessionAuthPayload(res.data || {})']) {
  if (!pageSource.includes(marker)) {
    throw new Error(`pages/index/index.vue 必须保存登录后权限状态: ${marker}`)
  }
}

for (const marker of ['applySessionAuthPayload(payload)', 'payloadHasSessionPermissions(payload)', 'await loadBootstrap()']) {
  if (!pageSource.includes(marker)) {
    throw new Error(`login 必须优先使用登录响应权限并仅在缺失时兜底 bootstrap: ${marker}`)
  }
}

const refreshMatch = pageSource.match(/async function refreshData\([^)]*\) \{([\s\S]*?)\n\}/)
if (!refreshMatch) {
  throw new Error('pages/index/index.vue 必须保留 refreshData 刷新函数')
}
const refreshBody = refreshMatch[1]
for (const snippet of ['loadBootstrap()', 'Promise.all([loadReviews(), loadUsers(), loadTemplate()])']) {
  if (refreshBody.includes(snippet)) {
    throw new Error(`refreshData 只能刷新当前视图需要的数据，不能无条件拉取: ${snippet}`)
  }
}
for (const marker of ['async function ensureReferenceData', 'ensureReferenceData(']) {
  if (!pageSource.includes(marker)) {
    throw new Error(`pages/index/index.vue 必须缓存用户/模板等引用数据: ${marker}`)
  }
}

if (componentSources.includes('ctx.loadBootstrap')) {
  throw new Error('组件刷新按钮不能调用 bootstrap，应改为 ctx.refreshData')
}

console.log('钉钉 H5 bootstrap 轻量化检查通过')
