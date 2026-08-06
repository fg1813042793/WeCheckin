import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = resolve(fileURLToPath(new URL('..', import.meta.url)))
const pageSource = readFileSync(resolve(root, 'pages/index/index.vue'), 'utf8')
const performanceAppSource = readFileSync(resolve(root, 'pages/index/usePerformanceApp.js'), 'utf8')
const performanceAuthSource = readFileSync(resolve(root, 'pages/index/composables/usePerformanceAuth.js'), 'utf8')
const performanceDataSource = readFileSync(resolve(root, 'pages/index/composables/usePerformanceData.js'), 'utf8')
const performanceReviewListSource = readFileSync(resolve(root, 'views/performance/common/composables/usePerformanceReviewList.js'), 'utf8')
const logicSource = `${pageSource}\n${performanceAppSource}\n${performanceAuthSource}\n${performanceDataSource}\n${performanceReviewListSource}`
const componentSources = [
  'views/profile/index.vue',
  'views/workbench/dashboard/index.vue',
  'views/performance/mine/index.vue',
  'views/performance/history/index.vue',
  'views/performance/manager-review/index.vue',
  'views/performance/hrbp-review/index.vue',
  'views/performance/flow-config/index.vue',
  'views/performance/summary/index.vue',
  'views/performance/parameters/index.vue',
  'pages/index/components/ReviewActionDialog.vue',
  'views/performance/mine/components/CreateReviewDialog.vue',
  'views/performance/flow-config/components/OrgUserConfigDialog.vue',
  'components/app/AppContentOutlet.js',
  'views/performance/common/context.js',
  'views/performance/common/reviewFlow.js',
  'views/performance/common/components/ProcessProgressModal.js',
  'views/performance/common/components/ReviewDetailModal.js',
  'views/performance/common/components/ReviewForm.js',
  'views/performance/common/components/ReviewFormSections.js',
  'views/performance/common/tables.js',
  'views/performance/common/tablePagination.js',
  'views/performance/common/workbenchPage.js'
].map((path) => readFileSync(resolve(root, path), 'utf8')).join('\n')

const match = logicSource.match(/async function loadBootstrap\(\) \{([\s\S]*?)\n\s*\}/)
if (!match) {
  throw new Error('钉钉 H5 必须保留 loadBootstrap 启动函数')
}

const body = match[1]
for (const snippet of ['res.data.users', 'res.data.reviews', 'res.data.template', 'listUsers()', 'listReviews(', 'getTemplate()']) {
  if (body.includes(snippet)) {
    throw new Error(`loadBootstrap 只能获取当前用户启动态，不能读取全量数据: ${snippet}`)
  }
}

for (const marker of ['async function refreshData', 'async function loadUsers', 'async function loadReviews', 'async function loadTemplate']) {
  if (!logicSource.includes(marker)) {
    throw new Error(`钉钉 H5 必须使用按需刷新函数: ${marker}`)
  }
}

for (const marker of ['permissionVersion: 0', 'apiPermissionKeys: []', 'apiPermissionReady: false', 'applySessionAuthPayload(res.data || {})']) {
  if (!logicSource.includes(marker)) {
    throw new Error(`usePerformanceApp.js 必须保存登录后权限状态: ${marker}`)
  }
}

for (const marker of ['applySessionAuthPayload(payload)', 'payloadHasSessionPermissions(payload)', 'await loadBootstrap()']) {
  if (!logicSource.includes(marker)) {
    throw new Error(`login 必须优先使用登录响应权限并仅在缺失时兜底 bootstrap: ${marker}`)
  }
}

const refreshMatch = logicSource.match(/async function refreshData\([^)]*\) \{([\s\S]*?)\n\s*\}/)
if (!refreshMatch) {
  throw new Error('钉钉 H5 必须保留 refreshData 刷新函数')
}
const refreshBody = refreshMatch[1]
for (const snippet of ['loadBootstrap()', 'Promise.all([loadReviews(), loadUsers(), loadTemplate()])']) {
  if (refreshBody.includes(snippet)) {
    throw new Error(`refreshData 只能刷新当前视图需要的数据，不能无条件拉取: ${snippet}`)
  }
}
for (const marker of ['async function ensureReferenceData', 'ensureReferenceData(']) {
  if (!logicSource.includes(marker)) {
    throw new Error(`钉钉 H5 必须缓存用户/模板等引用数据: ${marker}`)
  }
}

if (componentSources.includes('ctx.loadBootstrap')) {
  throw new Error('组件刷新按钮不能调用 bootstrap，应改为 ctx.refreshData')
}

console.log('钉钉 H5 bootstrap 轻量化检查通过')
