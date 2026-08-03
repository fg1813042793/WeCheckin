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
  'components/performance/BindAccountView.vue',
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
if (!apiSource.includes('publicConfig()') || !apiSource.includes("get(`${DINGTALK_H5_API}/public-config`)")) {
  throw new Error('services/dingtalkH5Api.js 必须封装钉钉 H5 公开启动配置接口 /public-config')
}
if (!apiSource.includes('ssoLogin(data)') || !apiSource.includes("post(`${DINGTALK_H5_API}/sso-login`, data)")) {
  throw new Error('services/dingtalkH5Api.js 必须封装钉钉免登接口 /sso-login')
}
if (!apiSource.includes('bindSelf(data)') || !apiSource.includes("post(`${DINGTALK_H5_API}/bind-self`, data)")) {
  throw new Error('services/dingtalkH5Api.js 必须封装钉钉自助绑定接口 /bind-self')
}

const configSource = readFileSync(resolve(root, 'config/index.js'), 'utf8')
if (configSource.includes("BASE_URL: env.VITE_API_BASE_URL || 'http://localhost:8083'")) {
  throw new Error('config/index.js 不能默认直连 localhost:8083，H5 本地开发应默认走同源 /api/v2 代理')
}

const requestSource = readFileSync(resolve(root, 'utils/request.js'), 'utf8')
if (!requestSource.includes('if (!baseUrl) return url')) {
  throw new Error('utils/request.js 必须支持空 BASE_URL，以便请求保持同源路径')
}

const dingTalkSource = readFileSync(resolve(root, 'utils/dingtalk.js'), 'utf8')
if (!dingTalkSource.includes('dd?.getAuthCode') || !dingTalkSource.includes('requestAuthCodeApi.call')) {
  throw new Error('utils/dingtalk.js 必须同时兼容 dd.getAuthCode 和 dd.runtime.permission.requestAuthCode')
}
if (!dingTalkSource.includes('function loadDingTalkScript()') || !dingTalkSource.includes('document.head.appendChild(script)')) {
  throw new Error('utils/dingtalk.js 必须按钉钉端内环境动态加载钉钉 JSAPI')
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

const htmlSource = readFileSync(resolve(root, 'index.html'), 'utf8')
if (htmlSource.includes('dingtalk.open.js')) {
  throw new Error('index.html 不能全局引入钉钉 JSAPI，否则普通浏览器会触发 notInDingTalk 控制台错误')
}

const pageSource = readFileSync(resolve(root, 'pages/index/index.vue'), 'utf8')
if (pageSource.includes("../../api") || pageSource.includes("../api")) {
  throw new Error('pages/index/index.vue 不能从 api 目录导入模块，避免浏览器加载 /api/index.js')
}
for (const marker of [
  "import AppShell from '../../components/performance/AppShell.vue'",
  "import BindAccountView from '../../components/performance/BindAccountView.vue'",
  "import LoginView from '../../components/performance/LoginView.vue'",
  "import { isDingTalkRuntime, requestAuthCode, waitForDingTalkJSAPI } from '../../utils/dingtalk'",
  "import OrgView from '../../components/performance/OrgView.vue'",
  "import WorkbenchView from '../../components/performance/WorkbenchView'",
  'providePerformanceContext(',
  'async function tryDingTalkAutoLogin()',
  'const publicCorpId = ref',
  'async function ensurePublicConfig()',
  'dingTalkAuthApi.publicConfig()',
  'function autoLoginErrorMessage(error)',
  'async function shouldTryDingTalkAutoLogin()',
  'waitForDingTalkJSAPI(1200)',
  'if (!(await shouldTryDingTalkAutoLogin())) {',
  'const corpId = currentCorpId()',
  'const authCode = await requestAuthCode(corpId)',
  'dingTalkAuthApi.ssoLogin({ corpId, authCode })',
  'isBindRequiredError(error)',
  'async function bindDingTalkUser()',
  'dingTalkAuthApi.bindSelf({',
  'await tryDingTalkAutoLogin()',
  'function resetRouteTabState()',
  'routeTabs.value = []',
  "activePerformanceTab.value = 'mine'",
  "managerReviewTab.value = 'pending'",
  "hrbpReviewTab.value = 'pending'",
  'resetRouteTabState()',
  'const createReviewExpandedDeptKeys = ref(new Set())',
  'const createTargetUserTreeRows = computed(() => flattenCreateTargetTree(createTargetUserTree.value, createReviewExpandedDeptKeys.value))',
  '@click="toggleCreateReviewDept(row.key)"',
  '@click.stop="toggleCreateReviewDepartment(row)"',
  'createReviewExpandedDeptKeys.value = new Set()'
]) {
  if (!pageSource.includes(marker)) {
    throw new Error(`pages/index/index.vue 必须完成组件拆分: ${marker}`)
  }
}
if (pageSource.includes('const OrgView =') || pageSource.includes('const WorkbenchView =') || pageSource.includes('<style scoped>')) {
  throw new Error('pages/index/index.vue 不能继续内联大视图组件或页面样式')
}

const orgSource = readFileSync(resolve(root, 'components/performance/OrgView.vue'), 'utf8')
for (const marker of ['org-toolbar', 'org-tree-list', 'org-tree-row', 'org-employee-list', 'org-config-modal', '已按权限过滤']) {
  if (!orgSource.includes(marker)) {
    throw new Error(`OrgView.vue 必须包含组织架构优化样式结构: ${marker}`)
  }
}
for (const marker of [
  'orgFilters',
  'filteredUsers',
  'org-filter-bar',
  'placeholder="搜索员工姓名/账号"',
  'org-department-filter',
  'orgFilterDepartmentRows',
  'placeholder="搜索部门名称"',
  '全部上级状态',
  '全部HRBP状态',
  'resetOrgFilters'
]) {
  if (!orgSource.includes(marker)) {
    throw new Error(`OrgView.vue 必须包含流程执行搜索筛选能力: ${marker}`)
  }
}

const summarySource = readFileSync(resolve(root, 'components/performance/SummaryView.js'), 'utf8')
const workbenchSource = readFileSync(resolve(root, 'components/performance/WorkbenchView.js'), 'utf8')
const tablePaginationSource = readFileSync(resolve(root, 'components/performance/tablePagination.js'), 'utf8')
for (const forbidden of ['ctx.openCreateReviewDialog', 'ctx.canCreateReview()']) {
  if (summarySource.includes(forbidden)) {
    throw new Error(`HRBP汇总页面不应展示创建考评单入口: ${forbidden}`)
  }
}
for (const marker of [
  'ctx.summaryFilters.employeeName',
  'ctx.summaryFilters.departmentName',
  'ctx.summaryFilters.departmentNames',
  'renderDepartmentFilter()',
  'renderDepartmentRow(row)',
  'renderSummaryPeriodFilter()',
  "placeholder: '员工姓名'",
  "placeholder: '搜索部门名称'",
  "ctx.summaryFilters.period || '考评月份'",
  "mode: 'date'",
  "fields: 'month'",
  '全部状态'
]) {
  if (!summarySource.includes(marker)) {
    throw new Error(`HRBP汇总页面筛选项必须收敛为员工姓名、部门名称、考评月份、状态: ${marker}`)
  }
}
for (const forbidden of [
  'ctx.summaryFilters.keyword',
  'ctx.summaryFilters.nextPeriod',
  'ctx.summaryFilters.managerId',
  'ctx.summaryFilters.hrbpId',
  'ctx.summaryFilters.grade',
  '全部部门',
  '全部上级',
  '全部HRBP',
  '全部分档',
  "h('input', { class: 'field-input', type: 'month', value: ctx.summaryFilters.period"
]) {
  if (summarySource.includes(forbidden)) {
    throw new Error(`HRBP汇总页面不应继续展示旧筛选项: ${forbidden}`)
  }
}

for (const marker of [
  'class: \'table-panel-stack summary-table-panel-stack\'',
  'class: \'table-panel-stack hrbp-table-panel-stack\'',
  'class: \'table-panel-stack manager-table-panel-stack\'',
  'class: \'table-panel-stack history-table-panel-stack\''
]) {
  if (!`${summarySource}\n${workbenchSource}`.includes(marker)) {
    throw new Error(`钉钉 H5 表格分页必须放在表格框外侧，并保持统一外层结构: ${marker}`)
  }
}

if (!tablePaginationSource.includes("class: 'table-pagination table-pagination-outside'")) {
  throw new Error('钉钉 H5 表格分页必须使用 table-pagination-outside 样式放在框外')
}

if (/h\('section',\s*\{ class: 'panel' \}, \[[\s\S]{0,1800}renderTablePagination\(/.test(summarySource) ||
  /h\('section',\s*\{ class: 'panel' \}, \[[\s\S]{0,1800}renderTablePagination\(/.test(workbenchSource)) {
  throw new Error('钉钉 H5 表格分页不能继续放在 panel 框内部')
}

const componentSources = [
  pageSource,
  orgSource,
  readFileSync(resolve(root, 'components/performance/LoginView.vue'), 'utf8'),
  readFileSync(resolve(root, 'components/performance/AppShell.vue'), 'utf8'),
  readFileSync(resolve(root, 'components/performance/WorkbenchView.js'), 'utf8'),
  readFileSync(resolve(root, 'styles/performance.css'), 'utf8')
].join('\n')

for (const marker of [
  '.review-create-head {\n  position: relative;',
  'padding-right: 58px;',
  '.review-create-close {\n  position: absolute;',
  'top: 50%;',
  'right: 16px;',
  '.login-warning {'
]) {
  if (!componentSources.includes(marker)) {
    throw new Error(`新建考评单弹窗必须保持默认折叠和右上角关闭按钮样式: ${marker}`)
  }
}

for (const marker of ['OA管理', '绩效管理', '我的绩效', 'HRBP评价', '流程执行', '#1677ff']) {
  if (!componentSources.includes(marker)) {
    throw new Error(`钉钉 H5 组件缺少绩效工作台标识: ${marker}`)
  }
}

const appShellSource = readFileSync(resolve(root, 'components/performance/AppShell.vue'), 'utf8')
for (const marker of [
  'desktop-page-tabs',
  'desktop-page-tabs-nav desktop-page-tabs-nav-left',
  'desktop-page-tabs-nav desktop-page-tabs-nav-right',
  'aria-label="向左翻动页签"',
  'aria-label="向右翻动页签"',
  "scrollDesktopPageTabs('left')",
  "scrollDesktopPageTabs('right')",
  'desktopTabsOverflowing',
  'desktopTabsCanScrollLeft',
  'desktopTabsCanScrollRight',
  'desktopPageTabsInnerRef',
  'updateDesktopPageTabsScrollState',
  'scrollActiveDesktopPageTabIntoView',
  '@click.stop="handleTopNavClick(item)"',
  'topSubmenuOpenKey',
  'class="top-nav-group"',
  'class="top-submenu-backdrop"',
  'class="top-submenu-dropdown"',
  'isTopSubmenuOpen(item)',
  'closeTopSubmenu',
  'handleTopChildClick',
  'top-nav-caret',
  'routeTabs.length',
  'handleRouteTabClick',
  'handleRouteTabClose',
  'isPageTabActive'
]) {
  if (!appShellSource.includes(marker)) {
    throw new Error(`PC 端必须在顶部内容区展示页面 tab: ${marker}`)
  }
}
for (const forbidden of [
  'menuLayout === \'top\' && topSubmenuItems.length',
  'class="top-submenu">',
  'class="top-submenu "'
]) {
  if (appShellSource.includes(forbidden)) {
    throw new Error(`PC 顶部一级菜单子菜单必须使用下拉浮层，不能继续使用横向子菜单条: ${forbidden}`)
  }
}
for (const marker of [
  '.desktop-page-tabs {',
  '.desktop-page-tabs-inner {',
  '.desktop-page-tabs-nav {',
  '.desktop-page-tabs-nav:disabled',
  '.top-nav-group {',
  '.top-submenu-backdrop {',
  '.top-submenu-dropdown {',
  '.desktop-page-tab.active::after',
  '.desktop-page-tabs,\n  .top-submenu-dropdown {\n    display: none;'
]) {
  if (!componentSources.includes(marker)) {
    throw new Error(`PC 页面 tab 必须具备桌面样式并在移动端隐藏: ${marker}`)
  }
}

console.log('钉钉 H5 微应用脚手架检查通过')
