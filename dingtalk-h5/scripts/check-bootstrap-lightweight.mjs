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

if (componentSources.includes('ctx.loadBootstrap')) {
  throw new Error('组件刷新按钮不能调用 bootstrap，应改为 ctx.refreshData')
}

console.log('钉钉 H5 bootstrap 轻量化检查通过')
