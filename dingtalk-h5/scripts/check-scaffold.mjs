import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = resolve(fileURLToPath(new URL('..', import.meta.url)))

const requiredFiles = [
  '.env.example',
  'App.vue',
  'README.md',
  'config/index.js',
  'index.html',
  'main.js',
  'manifest.json',
  'package.json',
  'pages.json',
  'components/performance/AppShell.vue',
  'components/performance/AccountView.vue',
  'components/performance/LoginView.vue',
  'components/performance/OrgView.vue',
  'components/performance/SummaryView.js',
  'components/performance/TemplateView.js',
  'components/performance/WorkbenchView.js',
  'components/performance/context.js',
  'pages/index/index.vue',
  'services/dingtalkH5Api.js',
  'styles/performance.css',
  'utils/dingtalk.js',
  'utils/request.js',
  'vite.config.js'
]

function readJson(path) {
  return JSON.parse(readFileSync(resolve(root, path), 'utf8'))
}

const missing = requiredFiles.filter((file) => !existsSync(resolve(root, file)))
if (missing.length) {
  throw new Error(`钉钉 H5 微应用脚手架文件缺失: ${missing.join(', ')}`)
}

const pkg = readJson('package.json')
if (pkg.name !== 'wecheckin-dingtalk-h5') {
  throw new Error('package.json name 必须是 wecheckin-dingtalk-h5')
}
if (!pkg.scripts?.['dev:h5'] || !pkg.scripts?.['build:h5']) {
  throw new Error('package.json 必须提供 dev:h5 和 build:h5 脚本')
}

const manifest = readJson('manifest.json')
if (manifest.name !== 'OA管理' || manifest.vueVersion !== '3') {
  throw new Error('manifest.json 必须声明 OA管理 和 Vue 3')
}

const pages = readJson('pages.json')
if (!pages.pages?.some((page) => page.path === 'pages/index/index')) {
  throw new Error('pages.json 必须包含 pages/index/index')
}

if (existsSync(resolve(root, 'api/index.js'))) {
  throw new Error('钉钉 H5 前端源码不能放在 /api/index.js，避免和后端 /api 接口前缀冲突')
}

const apiSource = readFileSync(resolve(root, 'services/dingtalkH5Api.js'), 'utf8')
if (!apiSource.includes("const API_V2 = '/api/v2'") || !apiSource.includes('/dingtalk/h5')) {
  throw new Error('services/dingtalkH5Api.js 必须使用 /api/v2/dingtalk/h5 接口前缀')
}

const configSource = readFileSync(resolve(root, 'config/index.js'), 'utf8')
if (configSource.includes("BASE_URL: env.VITE_API_BASE_URL || 'http://localhost:8083'")) {
  throw new Error('config/index.js 不能默认直连 localhost:8083，H5 本地开发应默认走同源 /api/v2 代理')
}

const requestSource = readFileSync(resolve(root, 'utils/request.js'), 'utf8')
if (!requestSource.includes('if (!baseUrl) return url')) {
  throw new Error('utils/request.js 必须支持空 BASE_URL，以便请求保持同源路径')
}

const viteSource = readFileSync(resolve(root, 'vite.config.js'), 'utf8')
if (!viteSource.includes("env.VITE_DEV_PROXY_TARGET || 'http://localhost:8083'")) {
  throw new Error('vite.config.js 必须默认将 /api/v2 代理到本地后端 localhost:8083')
}
if (!viteSource.includes("'/api/v2'")) {
  throw new Error('vite.config.js 必须只代理真实接口前缀 /api/v2')
}
if (viteSource.includes("'/api':")) {
  throw new Error('vite.config.js 不能代理整个 /api，否则会拦截前端源码目录 api/index.js')
}

const envExample = readFileSync(resolve(root, '.env.example'), 'utf8')
if (!envExample.includes('VITE_API_BASE_URL=') || envExample.includes('VITE_API_BASE_URL=http://localhost:8083')) {
  throw new Error('.env.example 中 VITE_API_BASE_URL 应默认为空，避免本地 H5 跨域直连后端')
}

const pageSource = readFileSync(resolve(root, 'pages/index/index.vue'), 'utf8')
if (pageSource.includes("../../api") || pageSource.includes("../api")) {
  throw new Error('pages/index/index.vue 不能从 api 目录导入模块，避免浏览器加载 /api/index.js')
}
for (const marker of [
  "import AppShell from '../../components/performance/AppShell.vue'",
  "import LoginView from '../../components/performance/LoginView.vue'",
  "import OrgView from '../../components/performance/OrgView.vue'",
  "import WorkbenchView from '../../components/performance/WorkbenchView'",
  'providePerformanceContext('
]) {
  if (!pageSource.includes(marker)) {
    throw new Error(`pages/index/index.vue 必须完成组件拆分: ${marker}`)
  }
}
if (pageSource.includes('const OrgView =') || pageSource.includes('const WorkbenchView =') || pageSource.includes('<style scoped>')) {
  throw new Error('pages/index/index.vue 不能继续内联大视图组件或页面样式')
}

const orgSource = readFileSync(resolve(root, 'components/performance/OrgView.vue'), 'utf8')
for (const marker of ['org-toolbar', 'org-table-shell', 'org-table', 'user-card-list']) {
  if (!orgSource.includes(marker)) {
    throw new Error(`OrgView.vue 必须包含组织架构优化样式结构: ${marker}`)
  }
}

const componentSources = [
  pageSource,
  orgSource,
  readFileSync(resolve(root, 'components/performance/LoginView.vue'), 'utf8'),
  readFileSync(resolve(root, 'components/performance/AppShell.vue'), 'utf8'),
  readFileSync(resolve(root, 'components/performance/WorkbenchView.js'), 'utf8'),
  readFileSync(resolve(root, 'styles/performance.css'), 'utf8')
].join('\n')

for (const marker of ['OA管理', '绩效管理', '本月绩效', 'HRBP评价', '流程执行', '#1677ff']) {
  if (!componentSources.includes(marker)) {
    throw new Error(`钉钉 H5 组件缺少绩效工作台标识: ${marker}`)
  }
}

console.log('钉钉 H5 微应用脚手架检查通过')
