import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = resolve(fileURLToPath(new URL('..', import.meta.url)))

const requiredFiles = [
  '.dockerignore',
  '.env.docker.example',
  '.env.example',
  'App.vue',
  'Dockerfile',
  'README.md',
  'config/index.js',
  'docker-compose.h5.yml',
  'docker/nginx.conf.template',
  'index.html',
  'main.js',
  'manifest.json',
  'package.json',
  'pages.json',
  'components/app/AppShell.vue',
  'components/app/useAppShellNavigation.js',
  'components/auth/AuthBrandHeader.vue',
  'components/auth/AuthGate.vue',
  'pages/index/index.vue',
  'pages/index/usePerformanceApp.js',
  'pages/index/composables/usePerformanceAuth.js',
  'pages/index/composables/usePerformanceData.js',
  'pages/index/composables/usePerformanceNavigation.js',
  'pages/index/composables/usePerformancePermissions.js',
  'router/index.js',
  'api/common/base.js',
  'api/auth/login/index.js',
  'api/auth/bind-account/index.js',
  'api/profile/index.js',
  'api/workbench/dashboard/index.js',
  'api/performance/common/reviews.js',
  'api/performance/mine/index.js',
  'api/performance/parameters/index.js',
  'api/performance/flow-config/index.js',
  'styles/performance.css',
  'utils/dingtalk.js',
  'utils/request.js',
  'views/auth/login/index.vue',
  'views/auth/bind-account/index.vue',
  'views/profile/index.vue',
  'views/profile/composables/usePerformanceProfile.js',
  'views/workbench/dashboard/index.vue',
  'views/performance/mine/index.vue',
  'views/performance/mine/composables/usePerformanceReviewCreation.js',
  'views/performance/history/index.vue',
  'views/performance/manager-review/index.vue',
  'views/performance/hrbp-review/index.vue',
  'views/performance/mine/components/CreateReviewDialog.vue',
  'views/performance/flow-config/components/OrgUserConfigDialog.vue',
  'views/performance/flow-config/composables/useOrgDirectory.js',
  'components/app/AppContentOutlet.js',
  'pages/index/components/ReviewActionDialog.vue',
  'views/performance/common/context.js',
  'views/performance/common/constants.js',
  'views/performance/common/helpers.js',
  'views/performance/common/reviewTabs.js',
  'views/performance/common/reviewFlow.js',
  'views/performance/common/composables/usePerformanceReviewActions.js',
  'views/performance/common/composables/usePerformanceReviewList.js',
  'views/performance/common/components/ProcessProgressModal.js',
  'views/performance/common/components/ReviewDetailModal.js',
  'views/performance/common/components/ReviewForm.js',
  'views/performance/common/components/ReviewFormSections.js',
  'views/performance/common/components/reviewFormHelpers.js',
  'views/performance/common/tables.js',
  'views/performance/common/tablePagination.js',
  'views/performance/common/workbenchPage.js',
  'views/performance/summary/index.vue',
  'views/performance/flow-config/index.vue',
  'views/performance/parameters/index.vue',
  'vite.config.js'
]

function readJson(path) {
  return JSON.parse(readFileSync(resolve(root, path), 'utf8'))
}

const missing = requiredFiles.filter((file) => !existsSync(resolve(root, file)))
if (missing.length) {
  throw new Error(`钉钉 H5 微应用脚手架文件缺失: ${missing.join(', ')}`)
}

const forbiddenFlatViewFiles = [
  'components/performance/LoginView.vue',
  'components/performance/BindAccountView.vue',
  'components/performance/usePerformanceApp.js',
  'components/performance/WorkbenchView.js',
  'components/performance/OrgView.vue',
  'components/performance/SummaryView.js',
  'components/performance/TemplateView.js',
  'components/performance/PerformanceContentRouter.js',
  'components/performance/WorkbenchTables.js',
  'components/performance/tablePagination.js',
  'components/performance/AccountView.vue',
  'components/performance/CreateReviewDialog.vue',
  'components/performance/OrgUserConfigDialog.vue',
  'components/performance/ProcessProgressModal.js',
  'components/performance/ReviewActionDialog.vue',
  'components/performance/ReviewDetailModal.js',
  'components/performance/ReviewForm.js',
  'components/performance/ReviewFormSections.js',
  'components/performance/context.js',
  'views/performance/router/ContentRouter.js',
  'composables/performance/appConstants.js',
  'composables/performance/appHelpers.js',
  'composables/performance/reviewFormHelpers.js',
  'composables/performance/useAppShellNavigation.js',
  'composables/performance/useOrgDirectory.js',
  'composables/performance/usePerformanceApp.js',
  'composables/performance/usePerformanceAuth.js',
  'composables/performance/usePerformanceData.js',
  'composables/performance/usePerformanceNavigation.js',
  'composables/performance/usePerformancePermissions.js',
  'composables/performance/usePerformanceProfile.js',
  'composables/performance/usePerformanceReviewActions.js',
  'composables/performance/usePerformanceReviewCreation.js',
  'composables/performance/usePerformanceReviewList.js',
  'components/performance/reviewFlow.js',
  'components/performance/ProfileDialog.vue',
  'components/performance/AuthGate.vue',
  'components/performance/AppShell.vue',
  'components/performance/AuthBrandHeader.vue',
  'views/performance/DashboardPage.js',
  'views/performance/MinePage.js',
  'views/performance/HistoryPage.js',
  'views/performance/ManagerReviewPage.js',
  'views/performance/HrbpReviewPage.js',
  'views/performance/SummaryPage.js',
  'views/performance/FlowConfigPage.vue',
  'views/performance/TemplatePage.js',
  'views/workbench/dashboard/index.js',
  'views/performance/mine/index.js',
  'views/performance/history/index.js',
  'views/performance/manager-review/index.js',
  'views/performance/hrbp-review/index.js',
  'views/performance/summary/index.js',
  'views/performance/summary/SummaryPage.js',
  'views/performance/flow-config/OrgPage.vue',
  'views/performance/parameters/index.js',
  'views/performance/shared/context.js',
  'views/performance/shared/constants.js',
  'views/performance/shared/helpers.js',
  'views/performance/shared/reviewFlow.js',
  'views/performance/shared/composables/usePerformanceReviewActions.js',
  'views/performance/shared/composables/usePerformanceReviewList.js',
  'views/performance/shared/components/ProcessProgressModal.js',
  'views/performance/shared/components/ReviewActionDialog.vue',
  'views/performance/shared/components/ReviewDetailModal.js',
  'views/performance/shared/components/ReviewForm.js',
  'views/performance/shared/components/ReviewFormSections.js',
  'views/performance/shared/components/reviewFormHelpers.js',
  'views/performance/shared/tables.js',
  'views/performance/shared/tablePagination.js',
  'views/performance/shared/workbenchPage.js',
  'views/performance/common/components/ReviewActionDialog.vue',
  'views/performance/parameters/TemplatePage.js'
]
const existingFlatViewFiles = forbiddenFlatViewFiles.filter((file) => existsSync(resolve(root, file)))
if (existingFlatViewFiles.length) {
  throw new Error(`钉钉 H5 页面文件必须按 views/一级菜单/二级菜单/index 组织，不能继续使用扁平页面文件: ${existingFlatViewFiles.join(', ')}`)
}

const forbiddenRouterFiles = [
  'router/menuRoutes.js',
  'router/deepLink.js'
]
const existingRouterFiles = forbiddenRouterFiles.filter((file) => existsSync(resolve(root, file)))
if (existingRouterFiles.length) {
  throw new Error(`钉钉 H5 router 应收敛为单入口 index.js，不能继续拆出过薄路由文件: ${existingRouterFiles.join(', ')}`)
}

const pkg = readJson('package.json')
if (pkg.name !== 'wecheckin-dingtalk-h5') {
  throw new Error('package.json name 必须是 wecheckin-dingtalk-h5')
}
if (!pkg.scripts?.['dev:h5'] || !pkg.scripts?.['build:h5']) {
  throw new Error('package.json 必须提供 dev:h5 和 build:h5 脚本')
}

const manifest = readJson('manifest.json')
if (manifest.name !== '钉钉H5应用' || manifest.vueVersion !== '3') {
  throw new Error('manifest.json 必须声明钉钉 H5 中性默认名称和 Vue 3')
}

const pages = readJson('pages.json')
if (!pages.pages?.some((page) => page.path === 'pages/index/index')) {
  throw new Error('pages.json 必须包含 pages/index/index')
}

if (existsSync(resolve(root, 'api/index.js'))) {
  throw new Error('钉钉 H5 前端 API 不能放在根 api/index.js，必须按 views 层级拆分，避免和后端 /api 接口前缀冲突')
}

if (existsSync(resolve(root, 'services/dingtalkH5Api.js'))) {
  throw new Error('钉钉 H5 API 已迁移到 dingtalk-h5/api，不能继续保留 services/dingtalkH5Api.js')
}

const apiBaseSource = readFileSync(resolve(root, 'api/common/base.js'), 'utf8')
const loginApiSource = readFileSync(resolve(root, 'api/auth/login/index.js'), 'utf8')
const bindApiSource = readFileSync(resolve(root, 'api/auth/bind-account/index.js'), 'utf8')
const profileApiSource = readFileSync(resolve(root, 'api/profile/index.js'), 'utf8')
const dashboardApiSource = readFileSync(resolve(root, 'api/workbench/dashboard/index.js'), 'utf8')
const reviewApiSource = readFileSync(resolve(root, 'api/performance/common/reviews.js'), 'utf8')
const mineApiSource = readFileSync(resolve(root, 'api/performance/mine/index.js'), 'utf8')
const templateApiSource = readFileSync(resolve(root, 'api/performance/parameters/index.js'), 'utf8')
const flowConfigApiSource = readFileSync(resolve(root, 'api/performance/flow-config/index.js'), 'utf8')
if (!apiBaseSource.includes("const API_V2 = '/api/v2'") || !apiBaseSource.includes('/dingtalk/h5')) {
  throw new Error('api/common/base.js 必须统一声明 /api/v2/dingtalk/h5 接口前缀')
}
for (const [apiPath, source, markers] of [
  ['api/auth/login/index.js', loginApiSource, ['publicConfig()', "get(`${DINGTALK_H5_API}/public-config`)", 'login(data)', "post(`${DINGTALK_H5_API}/login`, data)", 'ssoLogin(data)', "post(`${DINGTALK_H5_API}/sso-login`, data)", 'logout()']],
  ['api/auth/bind-account/index.js', bindApiSource, ['bindSelf(data)', "post(`${DINGTALK_H5_API}/bind-self`, data)"]],
  ['api/profile/index.js', profileApiSource, ['uploadAvatar(filePath)', "uploadFile(`${DINGTALK_H5_API}/account/avatar`, filePath)", 'updateProfile(data)', 'changePassword(data)']],
  ['api/workbench/dashboard/index.js', dashboardApiSource, ['bootstrap()', "get(`${DINGTALK_H5_API}/bootstrap`)", 'workbench()', "get(`${DINGTALK_H5_API}/workbench`)"]],
  ['api/performance/common/reviews.js', reviewApiSource, ['listReviews(params = {})', "get(`${DINGTALK_H5_API}/reviews`, params)", 'reviewDetail(id)', 'reviewAction(id, action, data = {})', 'deleteReview(id)', 'exportReviewsUrl(params = {})']],
  ['api/performance/mine/index.js', mineApiSource, ['createReview(data)', "post(`${DINGTALK_H5_API}/reviews`, data)"]],
  ['api/performance/parameters/index.js', templateApiSource, ['getTemplate()', "get(`${DINGTALK_H5_API}/template`)", 'saveTemplate(data)', "put(`${DINGTALK_H5_API}/template`, data)"]],
  ['api/performance/flow-config/index.js', flowConfigApiSource, ['listUsers()', "get(`${DINGTALK_H5_API}/users`)", 'createUser(data)', 'updateUser(id, data)', 'deleteUser(id)']]
]) {
  for (const marker of markers) {
    if (!source.includes(marker)) {
      throw new Error(`${apiPath} 缺少页面级接口封装: ${marker}`)
    }
  }
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

const dockerEnvExample = readFileSync(resolve(root, '.env.docker.example'), 'utf8')
for (const marker of [
  'DINGTALK_H5_HTTP_PORT=8086',
  'VITE_API_BASE_URL=',
  'VITE_DINGTALK_CORP_ID=',
  'NGINX_API_PROXY_TARGET=http://host.docker.internal:8083'
]) {
  if (!dockerEnvExample.includes(marker)) {
    throw new Error(`.env.docker.example 缺少 Docker 单独部署变量: ${marker}`)
  }
}

const dockerfileSource = readFileSync(resolve(root, 'Dockerfile'), 'utf8')
for (const marker of [
  'FROM node:20-alpine AS builder',
  'npm run build:h5',
  'FROM nginx:1.27-alpine',
  'dist/build/h5'
]) {
  if (!dockerfileSource.includes(marker)) {
    throw new Error(`Dockerfile 必须支持钉钉 H5 静态构建部署: ${marker}`)
  }
}

const composeSource = readFileSync(resolve(root, 'docker-compose.h5.yml'), 'utf8')
for (const marker of [
  'DingTalk H5 standalone Docker Compose configuration',
  'dingtalk-h5:',
  'logging: *default-logging',
  '${DINGTALK_H5_HTTP_PORT:-8086}:80',
  'host.docker.internal:host-gateway',
  'NGINX_API_PROXY_TARGET'
]) {
  if (!composeSource.includes(marker)) {
    throw new Error(`docker-compose.h5.yml 缺少钉钉 H5 单独部署配置: ${marker}`)
  }
}
for (const forbidden of ['mysql:', 'redis:', 'backend:']) {
  if (composeSource.includes(forbidden)) {
    throw new Error(`docker-compose.h5.yml 只能部署钉钉 H5 静态站点，不能包含 ${forbidden}`)
  }
}

const nginxTemplateSource = readFileSync(resolve(root, 'docker/nginx.conf.template'), 'utf8')
for (const marker of [
  'try_files $uri $uri/ /index.html',
  'location /api/v2/',
  'proxy_pass ${NGINX_API_PROXY_TARGET}',
  'location /uploads/',
  'location = /health'
]) {
  if (!nginxTemplateSource.includes(marker)) {
    throw new Error(`docker/nginx.conf.template 缺少 H5 Nginx 配置: ${marker}`)
  }
}

const htmlSource = readFileSync(resolve(root, 'index.html'), 'utf8')
if (htmlSource.includes('dingtalk.open.js')) {
  throw new Error('index.html 不能全局引入钉钉 JSAPI，否则普通浏览器会触发 notInDingTalk 控制台错误')
}

const pageSource = readFileSync(resolve(root, 'pages/index/index.vue'), 'utf8')
const authGateSource = readFileSync(resolve(root, 'components/auth/AuthGate.vue'), 'utf8')
const appContentOutletSource = readFileSync(resolve(root, 'components/app/AppContentOutlet.js'), 'utf8')
const routeIndexSource = readFileSync(resolve(root, 'router/index.js'), 'utf8')
const createReviewDialogSource = readFileSync(resolve(root, 'views/performance/mine/components/CreateReviewDialog.vue'), 'utf8')
const dashboardPageSource = readFileSync(resolve(root, 'views/workbench/dashboard/index.vue'), 'utf8')
const minePageSource = readFileSync(resolve(root, 'views/performance/mine/index.vue'), 'utf8')
const historyPageSource = readFileSync(resolve(root, 'views/performance/history/index.vue'), 'utf8')
const managerPageSource = readFileSync(resolve(root, 'views/performance/manager-review/index.vue'), 'utf8')
const hrbpPageSource = readFileSync(resolve(root, 'views/performance/hrbp-review/index.vue'), 'utf8')
const workbenchSharedSource = readFileSync(resolve(root, 'views/performance/common/workbenchPage.js'), 'utf8')
const performanceAppSource = readFileSync(resolve(root, 'pages/index/usePerformanceApp.js'), 'utf8')
const performanceConstantsSource = readFileSync(resolve(root, 'views/performance/common/constants.js'), 'utf8')
const performanceReviewTabsSource = readFileSync(resolve(root, 'views/performance/common/reviewTabs.js'), 'utf8')
const performanceAppHelpersSource = readFileSync(resolve(root, 'views/performance/common/helpers.js'), 'utf8')
const appShellNavigationSource = readFileSync(resolve(root, 'components/app/useAppShellNavigation.js'), 'utf8')
const performanceNavigationSource = readFileSync(resolve(root, 'pages/index/composables/usePerformanceNavigation.js'), 'utf8')
const performanceProfileSource = readFileSync(resolve(root, 'views/profile/composables/usePerformanceProfile.js'), 'utf8')
const performanceAuthSource = readFileSync(resolve(root, 'pages/index/composables/usePerformanceAuth.js'), 'utf8')
const performanceDataSource = readFileSync(resolve(root, 'pages/index/composables/usePerformanceData.js'), 'utf8')
const performancePermissionsSource = readFileSync(resolve(root, 'pages/index/composables/usePerformancePermissions.js'), 'utf8')
const performanceReviewActionsSource = readFileSync(resolve(root, 'views/performance/common/composables/usePerformanceReviewActions.js'), 'utf8')
const performanceReviewCreationSource = readFileSync(resolve(root, 'views/performance/mine/composables/usePerformanceReviewCreation.js'), 'utf8')
const performanceReviewListSource = readFileSync(resolve(root, 'views/performance/common/composables/usePerformanceReviewList.js'), 'utf8')
for (const [sourcePath, source] of Object.entries({
  'pages/index/usePerformanceApp.js': performanceAppSource,
  'pages/index/composables/usePerformanceAuth.js': performanceAuthSource,
  'pages/index/composables/usePerformanceData.js': performanceDataSource,
  'views/profile/composables/usePerformanceProfile.js': performanceProfileSource,
  'views/performance/common/composables/usePerformanceReviewActions.js': performanceReviewActionsSource,
  'views/performance/common/composables/usePerformanceReviewList.js': performanceReviewListSource,
  'views/performance/mine/composables/usePerformanceReviewCreation.js': performanceReviewCreationSource
})) {
  for (const forbidden of ['services/dingtalkH5Api', 'dingTalkAuthApi.', 'dingTalkPerformanceApi.']) {
    if (source.includes(forbidden)) {
      throw new Error(`${sourcePath} 必须改为按页面目录导入 dingtalk-h5/api 模块，不能继续使用旧 API 对象: ${forbidden}`)
    }
  }
}
if (pageSource.includes("../../api") || pageSource.includes("../api")) {
  throw new Error('pages/index/index.vue 不能从 api 目录导入模块，避免浏览器加载 /api/index.js')
}
for (const marker of [
  "import AppShell from '../../components/app/AppShell.vue'",
  "import AppContentOutlet from '../../components/app/AppContentOutlet'",
  "import AuthGate from '../../components/auth/AuthGate.vue'",
  "import BindAccountView from '../../views/auth/bind-account/index.vue'",
  "import LoginView from '../../views/auth/login/index.vue'",
  "import ProfileDialog from '../../views/profile/index.vue'",
  "import CreateReviewDialog from '../../views/performance/mine/components/CreateReviewDialog.vue'",
  "import ReviewActionDialog from './components/ReviewActionDialog.vue'",
  "import { usePerformanceApp } from './usePerformanceApp'",
  '<template #bind>',
  '<BindAccountView',
  '<template #login>',
  '<LoginView',
  'const {',
  '} = usePerformanceApp()'
]) {
  if (!pageSource.includes(marker)) {
    throw new Error(`pages/index/index.vue 必须只保留页面装配层: ${marker}`)
  }
}
for (const forbidden of [
  "import BindAccountView from '../../components/performance/BindAccountView.vue'",
  "import LoginView from '../../components/performance/LoginView.vue'",
  "import OrgView from '../../components/performance/OrgView.vue'",
  "import WorkbenchView from '../../components/performance/WorkbenchView'",
  'providePerformanceContext(',
  'async function tryDingTalkAutoLogin()',
  'const OrgView =',
  'const WorkbenchView =',
  '<style scoped>'
]) {
  if (pageSource.includes(forbidden)) {
    throw new Error(`pages/index/index.vue 不能继续包含已拆出的组件或业务逻辑: ${forbidden}`)
  }
}
for (const marker of [
  '<slot v-if="!ready" name="loading"',
  'v-else-if="!user && bindState.visible"',
  'name="bind"',
  'v-else-if="!user && sessionAccessDenied"',
  'name="denied"',
  'v-else-if="!user"',
  'name="login"',
  '<slot v-else />',
  'sessionAccessDeniedMessage'
]) {
  if (!authGateSource.includes(marker)) {
    throw new Error(`AuthGate.vue 必须只负责鉴权状态分发并暴露插槽: ${marker}`)
  }
}
for (const forbidden of [
  '../../views/auth',
  'BindAccountView',
  'LoginView',
  'defineEmits'
]) {
  if (authGateSource.includes(forbidden)) {
    throw new Error(`AuthGate.vue 不能继续耦合具体页面或业务事件: ${forbidden}`)
  }
}
for (const marker of [
  "import { resolveMenuPageComponent } from '../../router/index'",
  "name: 'AppContentOutlet'",
  'resolveMenuPageComponent(props.contentView)'
]) {
  if (!appContentOutletSource.includes(marker)) {
    throw new Error(`components/app/AppContentOutlet.js 必须只做应用内容出口渲染分发，菜单页面映射应放到 router/index.js: ${marker}`)
  }
}
for (const forbidden of [
  "import DashboardPage from '../../workbench/dashboard/index.vue'",
  "import FlowConfigPage from '../flow-config/index.vue'",
  "import HrbpReviewPage from '../hrbp-review/index.vue'",
  "import HistoryPage from '../history/index.vue'",
  "import ManagerReviewPage from '../manager-review/index.vue'",
  "import MinePage from '../mine/index.vue'",
  "import SummaryPage from '../summary/index.vue'",
  "import TemplatePage from '../parameters/index.vue'",
  'const pageMap = {'
]) {
  if (appContentOutletSource.includes(forbidden)) {
    throw new Error(`components/app/AppContentOutlet.js 不能继续内置页面映射，必须迁移到 router/index.js: ${forbidden}`)
  }
}
for (const marker of [
  "import { menuPageKeys } from '../views/performance/common/constants'",
  "import { reviewFormTabKeys } from '../views/performance/common/reviewTabs'",
  "import DashboardPage from '../views/workbench/dashboard/index.vue'",
  "import FlowConfigPage from '../views/performance/flow-config/index.vue'",
  "import HrbpReviewPage from '../views/performance/hrbp-review/index.vue'",
  "import HistoryPage from '../views/performance/history/index.vue'",
  "import ManagerReviewPage from '../views/performance/manager-review/index.vue'",
  "import MinePage from '../views/performance/mine/index.vue'",
  "import SummaryPage from '../views/performance/summary/index.vue'",
  "import TemplatePage from '../views/performance/parameters/index.vue'",
  'const menuPageComponents = {',
  'export function resolveMenuPageComponent(contentView)',
  'export function normalizeReviewDeepLinkView(value)',
  'export function normalizeReviewDeepLinkTab(value)'
]) {
  if (!routeIndexSource.includes(marker)) {
    throw new Error(`router/index.js 必须集中维护菜单页面映射和通知/URL 深链解析: ${marker}`)
  }
}
if (routeIndexSource.includes('const reviewFormTabKeys = new Set([')) {
  throw new Error('router/index.js 不能继续定义绩效表单 tab 白名单，必须从 views/performance/common/reviewTabs.js 导入')
}
for (const marker of [
  'export const reviewFormTabKeys = new Set([',
  "'currentTargets'",
  "'reflection'",
  "'values'",
  "'manager'",
  "'hrbp'",
  "'nextTargets'"
]) {
  if (!performanceReviewTabsSource.includes(marker)) {
    throw new Error(`views/performance/common/reviewTabs.js 必须维护绩效表单 tab 白名单: ${marker}`)
  }
}
for (const [pagePath, source] of Object.entries({
  'views/workbench/dashboard/index.vue': dashboardPageSource,
  'views/performance/mine/index.vue': minePageSource,
  'views/performance/history/index.vue': historyPageSource,
  'views/performance/manager-review/index.vue': managerPageSource,
  'views/performance/hrbp-review/index.vue': hrbpPageSource
})) {
  if (source.includes('WorkbenchView')) {
    throw new Error(`${pagePath} 不能继续导入万能 WorkbenchView，必须承载对应菜单页面自己的渲染逻辑`)
  }
}
for (const marker of [
  'renderWorkbenchTodoList(',
  'workbenchTodoMeta(',
  'openWorkbenchTodo(review)'
]) {
  if (!dashboardPageSource.includes(marker)) {
    throw new Error(`工作台页面必须独立承载待办卡片逻辑: ${marker}`)
  }
}
for (const marker of [
  'renderMyPerformanceSwitcher(',
  'h(ReviewForm',
  'ctx.openCreateReviewDialog'
]) {
  if (!minePageSource.includes(marker)) {
    throw new Error(`本月绩效页面必须独立承载个人绩效表单逻辑: ${marker}`)
  }
}
for (const marker of [
  'renderHistoryFilters(',
  'renderHistoryTable(',
  'HistoryReviewDetailModal'
]) {
  if (!historyPageSource.includes(marker)) {
    throw new Error(`历史绩效页面必须独立承载历史表格和详情逻辑: ${marker}`)
  }
}
for (const marker of [
  'renderManagerReviewTable(',
  'switchManagerReviewTab',
  'h(ReviewForm'
]) {
  if (!managerPageSource.includes(marker)) {
    throw new Error(`上级审批页面必须独立承载审批列表和详情逻辑: ${marker}`)
  }
}
for (const marker of [
  'renderHrbpReviewTable(',
  'switchHrbpReviewTab',
  'h(ReviewForm'
]) {
  if (!hrbpPageSource.includes(marker)) {
    throw new Error(`HRBP评价页面必须独立承载评价列表和详情逻辑: ${marker}`)
  }
}
for (const marker of [
  "import {",
  "from '../../views/performance/common/constants'",
  "from '../../views/performance/common/helpers'",
  "from './composables/usePerformanceAuth'",
  "from './composables/usePerformanceData'",
  "from './composables/usePerformanceNavigation'",
  "from './composables/usePerformancePermissions'",
  "from '../../views/profile/composables/usePerformanceProfile'",
  "from '../../views/performance/common/composables/usePerformanceReviewActions'",
  "from '../../views/performance/mine/composables/usePerformanceReviewCreation'",
  "from '../../views/performance/common/composables/usePerformanceReviewList'",
  "import { setNavigationTitle } from '../../utils/dingtalk'",
  'providePerformanceContext(',
  'usePerformanceAuth({',
  'usePerformanceData({',
  'usePerformancePermissions({',
  'await tryDingTalkAutoLogin()',
  'replaceRouteTabs: true',
  'usePerformanceNavigation({',
  'resetRouteTabState()',
  'usePerformanceReviewCreation({',
  'usePerformanceReviewList({',
  'usePerformanceReviewActions({'
]) {
  if (!performanceAppSource.includes(marker)) {
    throw new Error(`pages/index/usePerformanceApp.js 必须承载钉钉 H5 页面业务逻辑: ${marker}`)
  }
}
for (const forbidden of [
  'const statusMeta = {',
  'const myPerformanceStatuses =',
  'const reviewActionConfirmCopy = {',
  'const menuPageKeys = new Set',
  'function defaultAppConfig()',
  'function normalizeAppConfig(',
  'function currentMonth()',
  'function buildCreateTargetUserTree(',
  'async function closeRouteTab(',
  'function ensureActiveMenu(',
  'function resetRouteTabState(',
  'function syncRouteTabLabels(',
  'function submitProfileDialog(',
  'function uploadProfileAvatar(',
  'function replaceLocalAccountReferences(',
  'async function createReview()',
  'async function loadReviews(',
  'function reviewQueryParamsForContentView(',
  'async function tryDingTalkAutoLogin()',
  'const publicCorpId = ref',
  'async function ensurePublicConfig()',
  'dingTalkAuthApi.publicConfig()',
  'function autoLoginErrorMessage(error)',
  'async function shouldTryDingTalkAutoLogin()',
  'async function loadUsers()',
  'async function saveTemplate(data)',
  'async function saveUser(user)',
  'function hasApiPermission(key)',
  'function canPerformReviewAction(action)',
  'async function performReviewAction(',
  'function validateSelfSubmitReview(',
  'async function submitWithdrawReview('
]) {
  if (performanceAppSource.includes(forbidden)) {
    throw new Error(`pages/index/usePerformanceApp.js 不能继续承载纯常量或纯工具函数: ${forbidden}`)
  }
}
const performanceAppLineCount = performanceAppSource.split(/\r?\n/).length
if (performanceAppLineCount > 900) {
  throw new Error(`pages/index/usePerformanceApp.js 行数过大：当前 ${performanceAppLineCount} 行，目标不超过 900 行，请继续按业务域拆分`)
}
for (const marker of [
  'export function usePerformanceAuth(',
  'async function tryDingTalkAutoLogin()',
  'async function bindDingTalkUser()',
  'async function login()',
  'function resetSessionState()',
  'function applySessionAuthPayload('
]) {
  if (!performanceAuthSource.includes(marker)) {
    throw new Error(`usePerformanceAuth.js 必须承载登录、免登、绑定和会话状态逻辑: ${marker}`)
  }
}
for (const marker of [
  'export function usePerformanceData(',
  'async function refreshData(',
  'async function loadUsers()',
  'async function loadTemplate()',
  'async function saveTemplate(data)',
  'async function saveUser(user)',
  'async function deleteReview(id)'
]) {
  if (!performanceDataSource.includes(marker)) {
    throw new Error(`usePerformanceData.js 必须承载人员、模板、刷新和数据变更逻辑: ${marker}`)
  }
}
for (const marker of [
  'export function usePerformancePermissions(',
  'function hasApiPermission(key)',
  'function hasButtonPermission(key)',
  'function canCreateReview()',
  'function canPerformReviewAction(action)',
  'function canWithdraw(review)'
]) {
  if (!performancePermissionsSource.includes(marker)) {
    throw new Error(`usePerformancePermissions.js 必须承载菜单、按钮、接口和流转动作权限判断: ${marker}`)
  }
}
for (const marker of [
  'export function usePerformanceNavigation(',
  "import { normalizeReviewDeepLinkTab, normalizeReviewDeepLinkView } from '../../../router/index'",
  "const activePerformanceTab = ref('mine')",
  'async function switchView(',
  'async function closeRouteTab(',
  'function ensureActiveMenu(',
  'function resetRouteTabState(',
  'function workbenchTodoTarget(',
  'replaceRouteTabs',
  "managerReviewTab.value = 'pending'",
  "hrbpReviewTab.value = 'pending'"
]) {
  if (!performanceNavigationSource.includes(marker)) {
    throw new Error(`usePerformanceNavigation.js 必须承载绩效应用菜单、页签和待办跳转逻辑: ${marker}`)
  }
}
for (const forbidden of [
  'export function normalizeReviewDeepLinkView(',
  'export function normalizeReviewDeepLinkTab('
]) {
  if (performanceNavigationSource.includes(forbidden)) {
    throw new Error(`usePerformanceNavigation.js 不能继续定义深链解析函数，必须迁移到 router/index.js: ${forbidden}`)
  }
}
if (!performanceAppSource.includes("from '../../router/index'")) {
  throw new Error('pages/index/usePerformanceApp.js 必须从 router/index.js 使用通知/URL 深链解析')
}
for (const marker of [
  'export function usePerformanceProfile(',
  'const profileDialog = reactive({',
  'const profileAvatarPreview = computed(',
  'function openProfileDialog(',
  'function chooseProfileAvatar(',
  'async function uploadProfileAvatar(',
  'async function submitProfileDialog(',
  'function replaceLocalAccountReferences('
]) {
  if (!performanceProfileSource.includes(marker)) {
    throw new Error(`usePerformanceProfile.js 必须承载个人资料、头像和密码编辑逻辑: ${marker}`)
  }
}
for (const marker of [
  'export function usePerformanceReviewCreation(',
  'const createReviewDialog = reactive({',
  'const createReviewExpandedDeptKeys = ref(new Set())',
  'function openCreateReviewDialog(',
  'function toggleCreateReviewDepartment(',
  'async function createReview('
]) {
  if (!performanceReviewCreationSource.includes(marker)) {
    throw new Error(`usePerformanceReviewCreation.js 必须承载创建考评单业务逻辑: ${marker}`)
  }
}
for (const marker of [
  'export function usePerformanceReviewList(',
  'async function loadReviews(',
  'function reviewQueryParamsForContentView(',
  'async function switchManagerReviewTab(',
  'async function switchHrbpReviewTab('
]) {
  if (!performanceReviewListSource.includes(marker)) {
    throw new Error(`usePerformanceReviewList.js 必须承载考评列表查询业务逻辑: ${marker}`)
  }
}
for (const marker of [
  'export function usePerformanceReviewActions(',
  'async function performReviewAction(',
  'function validateHrbpSubmitReview(',
  'async function submitWithdrawReview(',
  'async function submitReturnReview(',
  'async function submitDisputeReview('
]) {
  if (!performanceReviewActionsSource.includes(marker)) {
    throw new Error(`usePerformanceReviewActions.js 必须承载表单校验与提交流转逻辑: ${marker}`)
  }
}
for (const marker of [
  'export const statusMeta',
  'export const myPerformanceStatuses',
  'export const myPerformanceStatusSet',
  'export const reviewActionApiPermissions',
  'export const reviewActionButtonPermissions',
  'export const reviewActionConfirmCopy',
  'export const menuPageKeys',
  'export const historyMonthOptions'
]) {
  if (!performanceConstantsSource.includes(marker)) {
    throw new Error(`views/performance/common/constants.js 必须集中维护绩效页面常量: ${marker}`)
  }
}
for (const marker of [
  'export function defaultAppConfig()',
  'export function normalizeAppConfig(',
  'export function normalizeMenuTree(',
  'export function flattenMenuTree(',
  'export function buildCreateTargetUserTree(',
  'export function flattenCreateTargetTree(',
  'export function currentMonth()',
  'export function nextMonthFromPeriod(',
  'export function firstText('
]) {
  if (!performanceAppHelpersSource.includes(marker)) {
    throw new Error(`appHelpers.js 必须集中维护绩效应用纯工具函数: ${marker}`)
  }
}
for (const marker of [
  '@click="$emit(\'toggle-dept\', row.key)"',
  '@click.stop="$emit(\'toggle-department\', row)"'
]) {
  if (!createReviewDialogSource.includes(marker)) {
    throw new Error(`CreateReviewDialog.vue 必须保留部门树展开和部门多选交互: ${marker}`)
  }
}

const orgSource = readFileSync(resolve(root, 'views/performance/flow-config/index.vue'), 'utf8')
const orgUserConfigDialogSource = readFileSync(resolve(root, 'views/performance/flow-config/components/OrgUserConfigDialog.vue'), 'utf8')
const orgDirectorySource = readFileSync(resolve(root, 'views/performance/flow-config/composables/useOrgDirectory.js'), 'utf8')
for (const marker of ['org-toolbar', 'org-tree-list', 'org-tree-row', 'org-employee-list', '已按权限过滤']) {
  if (!orgSource.includes(marker)) {
    throw new Error(`views/performance/flow-config/index.vue 必须包含组织架构优化样式结构: ${marker}`)
  }
}
for (const marker of [
  "import OrgUserConfigDialog from './components/OrgUserConfigDialog.vue'",
  '<OrgUserConfigDialog',
  '@toggle-relation-picker="toggleRelationPicker"',
  '@toggle-relation-department="toggleRelationDepartment"',
  '@select-relation-user="selectRelationUser"'
]) {
  if (!orgSource.includes(marker)) {
    throw new Error(`views/performance/flow-config/index.vue 必须复用独立配置审批人弹窗组件: ${marker}`)
  }
}
for (const forbidden of [
  'class="org-config-modal-mask"',
  'class="org-config-modal"',
  'function relationPickerText('
]) {
  if (orgSource.includes(forbidden)) {
    throw new Error(`views/performance/flow-config/index.vue 不能继续承载配置审批人弹窗实现: ${forbidden}`)
  }
}
for (const marker of [
  'org-config-modal-mask',
  'org-config-modal',
  'relation-picker-field',
  'managerPickerRows',
  'hrbpPickerRows',
  'relationPickerText(',
  "emit('toggle-relation-picker'",
  "emit('toggle-relation-department'",
  "emit('select-relation-user'"
]) {
  if (!orgUserConfigDialogSource.includes(marker)) {
    throw new Error(`OrgUserConfigDialog.vue 必须承载配置审批人弹窗和树形单选交互: ${marker}`)
  }
}
for (const marker of [
  "from './composables/useOrgDirectory'",
  'buildDepartmentTree',
  'flattenDepartmentTree',
  'filterDepartmentTree',
  'departmentText'
]) {
  if (!orgSource.includes(marker)) {
    throw new Error(`views/performance/flow-config/index.vue 必须复用组织树组合式工具，保留页面交互层: ${marker}`)
  }
}
for (const forbidden of [
  'function buildDepartmentTree(',
  'function flattenDepartmentTree(',
  'function filterDepartmentTree(',
  'function normalizeText(',
  'function realDepartmentLevels(',
  'function sortUsers('
]) {
  if (orgSource.includes(forbidden)) {
    throw new Error(`views/performance/flow-config/index.vue 不能继续承载组织树纯工具函数: ${forbidden}`)
  }
}
for (const marker of [
  'export function buildDepartmentTree(',
  'export function flattenDepartmentTree(',
  'export function filterDepartmentTree(',
  'export function flattenDepartmentSelectionTree(',
  'export function normalizeText(',
  'export function departmentText(',
  'export function userOptionText('
]) {
  if (!orgDirectorySource.includes(marker)) {
    throw new Error(`views/performance/flow-config/composables/useOrgDirectory.js 必须集中维护流程执行组织树工具: ${marker}`)
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
    throw new Error(`views/performance/flow-config/index.vue 必须包含流程执行搜索筛选能力: ${marker}`)
  }
}

const summarySource = readFileSync(resolve(root, 'views/performance/summary/index.vue'), 'utf8')
const templateSource = readFileSync(resolve(root, 'views/performance/parameters/index.vue'), 'utf8')
const workbenchSource = [
  dashboardPageSource,
  minePageSource,
  historyPageSource,
  managerPageSource,
  hrbpPageSource,
  workbenchSharedSource
].join('\n')
const workbenchTablesSource = readFileSync(resolve(root, 'views/performance/common/tables.js'), 'utf8')
const reviewDetailModalSource = readFileSync(resolve(root, 'views/performance/common/components/ReviewDetailModal.js'), 'utf8')
const processProgressModalSource = readFileSync(resolve(root, 'views/performance/common/components/ProcessProgressModal.js'), 'utf8')
const reviewFormSource = readFileSync(resolve(root, 'views/performance/common/components/ReviewForm.js'), 'utf8')
const reviewFormSectionsSource = readFileSync(resolve(root, 'views/performance/common/components/ReviewFormSections.js'), 'utf8')
const reviewFormHelpersSource = readFileSync(resolve(root, 'views/performance/common/components/reviewFormHelpers.js'), 'utf8')
const tablePaginationSource = readFileSync(resolve(root, 'views/performance/common/tablePagination.js'), 'utf8')
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
  "fields: 'month'"
]) {
  if (!summarySource.includes(marker)) {
    throw new Error(`HRBP汇总页面筛选项必须收敛为员工姓名、部门名称、考评月份: ${marker}`)
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
  "import { HistoryReviewDetailModal } from '../common/components/ReviewDetailModal'",
  'h(HistoryReviewDetailModal'
]) {
  if (!summarySource.includes(marker)) {
    throw new Error(`views/performance/summary/index.vue 必须复用独立的绩效详情弹窗组件: ${marker}`)
  }
}
for (const marker of [
  "import { HistoryReviewDetailModal } from '../common/components/ReviewDetailModal'",
  'h(HistoryReviewDetailModal'
]) {
  if (!workbenchSource.includes(marker)) {
    throw new Error(`绩效列表页面必须复用独立的绩效详情弹窗组件: ${marker}`)
  }
}
if (summarySource.includes("from './WorkbenchView'")) {
  throw new Error('views/performance/summary/index.vue 不能再从 WorkbenchView.js 引用绩效详情弹窗')
}
if (workbenchSource.includes('export const HistoryReviewDetailModal')) {
  throw new Error('HistoryReviewDetailModal 必须保持在 ReviewDetailModal.js，不能回流到页面中')
}
for (const marker of [
  'export const HistoryReviewDetailModal',
  "import { usePerformanceContext } from '../context'",
  "import { ReviewForm } from './ReviewForm'"
]) {
  if (!reviewDetailModalSource.includes(marker)) {
    throw new Error(`ReviewDetailModal.js 必须承载绩效详情弹窗: ${marker}`)
  }
}
for (const marker of [
  "import { ReviewForm } from '../common/components/ReviewForm'",
  "from '../common/tables'",
  'renderHistoryTable(',
  'renderHrbpReviewTable(',
  'renderManagerReviewTable('
]) {
  if (!workbenchSource.includes(marker)) {
    throw new Error(`绩效菜单页面必须复用独立表单、流程和表格工具: ${marker}`)
  }
}
if (!workbenchTablesSource.includes("import { currentAssignee } from './reviewFlow'")) {
  throw new Error('views/performance/common/tables.js 必须复用共享流程状态工具 currentAssignee')
}
if (workbenchSource.includes('const ReviewForm = {') || workbenchSource.includes('const ProcessModal = {')) {
  throw new Error('绩效菜单页面不能继续承载绩效详情表单和流程弹窗')
}
for (const marker of [
  'export function renderHistoryTable(',
  'export function renderHrbpReviewTable(',
  'export function renderManagerReviewTable(',
  'export function renderHistoryFilters(',
  'export function renderHrbpReviewedFilters(',
  'export function renderManagerReviewedFilters(',
  'export function reviewMatchesHrbpReviewedFilters(',
  'export function reviewMatchesManagerReviewedFilters('
]) {
  if (!workbenchTablesSource.includes(marker)) {
    throw new Error(`views/performance/common/tables.js 必须独立承载工作台表格和筛选渲染: ${marker}`)
  }
}
for (const forbidden of [
  'const HISTORY_TABLE_COLUMNS =',
  'const HRBP_REVIEW_TABLE_COLUMNS =',
  'const MANAGER_REVIEW_TABLE_COLUMNS =',
  'function renderHistoryCell(',
  'function renderHrbpReviewCell(',
  'function renderManagerReviewCell(',
  'function renderHistoryFilters(',
  'function renderHrbpReviewedFilters(',
  'function renderManagerReviewedFilters(',
  'function reviewMatchesHrbpReviewedFilters(',
  'function reviewMatchesManagerReviewedFilters('
]) {
  if (workbenchSource.includes(forbidden)) {
    throw new Error(`绩效菜单页面不能继续承载工作台表格、筛选和单元格实现: ${forbidden}`)
  }
}
for (const forbidden of [
  'const ProcessModal =',
  'const ValueStandardModal =',
  'const CurrentObjectiveSection =',
  'const SelfSummarySection =',
  'const NextObjectiveSection =',
  'const ValueSection =',
  'const ManagerSection =',
  'const HrbpSection ='
]) {
  if (reviewFormSource.includes(forbidden)) {
    throw new Error(`ReviewForm.js 不能继续承载弹窗或表单分区实现: ${forbidden}`)
  }
}
for (const marker of [
  'export const ProcessProgressModal',
  'flowProgressRows(ctx, review)',
  'currentAssignee(ctx, review)'
]) {
  if (!processProgressModalSource.includes(marker)) {
    throw new Error(`ProcessProgressModal.js 必须独立承载流程进度弹窗: ${marker}`)
  }
}
for (const marker of [
  "from './reviewFormHelpers'",
  'export const CurrentObjectiveSection',
  'export const SelfSummarySection',
  'export const NextObjectiveSection',
  'export const ValueSection',
  'export const ManagerSection',
  'export const HrbpSection',
  'export const EmployeeConfirmSection',
  'export const FinalSection',
  'export const HistorySection',
  'const ValueStandardModal'
]) {
  if (!reviewFormSectionsSource.includes(marker)) {
    throw new Error(`ReviewFormSections.js 必须独立承载绩效表单分区: ${marker}`)
  }
}
for (const forbidden of [
  'function estimateTextareaRows(',
  'function textareaAutoHeightStyle(',
  'function createObjective(',
  'function confirmObjectiveDelete(',
  'function hasManagerEvaluation(',
  'function hasHrbpEvaluation(',
  'function valueRubricItems('
]) {
  if (reviewFormSectionsSource.includes(forbidden)) {
    throw new Error(`ReviewFormSections.js 不能继续承载绩效表单辅助函数: ${forbidden}`)
  }
}
for (const marker of [
  'export function readInputValue(',
  'export function textareaAutoHeightStyle(',
  'export function addCurrentObjective(',
  'export function addNextObjective(',
  'export async function confirmRemoveCurrentObjective(',
  'export async function confirmRemoveNextObjective(',
  'export function hasManagerEvaluation(',
  'export function hasHrbpEvaluation(',
  'export function reviewGradeBadge(',
  'export function valueRubricItems(',
  'export function renderValueRubricList('
]) {
  if (!reviewFormHelpersSource.includes(marker)) {
    throw new Error(`views/performance/common/components/reviewFormHelpers.js 必须集中维护绩效表单辅助函数: ${marker}`)
  }
}
for (const forbidden of [
  'const reviewFormTabs =',
  'function normalizeReviewFormTab(',
  'function renderReviewFormPane('
]) {
  if (workbenchSource.includes(forbidden)) {
    throw new Error(`绩效菜单页面不能继续承载绩效表单 tab 分发逻辑: ${forbidden}`)
  }
}
for (const marker of [
  'const reviewFormTabs =',
  'function normalizeReviewFormTab(',
  'function renderReviewFormPane(',
  'h(CurrentObjectiveSection',
  'h(SelfSummarySection',
  'h(ValueSection',
  'h(ManagerSection',
  'h(HrbpSection',
  'h(NextObjectiveSection'
]) {
  if (!reviewFormSource.includes(marker)) {
    throw new Error(`ReviewForm.js 必须独立承载绩效表单 tab 分发逻辑: ${marker}`)
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
  summarySource,
  templateSource,
  workbenchSource,
  authGateSource,
  createReviewDialogSource,
  performanceAppSource,
  readFileSync(resolve(root, 'views/auth/login/index.vue'), 'utf8'),
  readFileSync(resolve(root, 'views/auth/bind-account/index.vue'), 'utf8'),
  readFileSync(resolve(root, 'components/auth/AuthBrandHeader.vue'), 'utf8'),
  readFileSync(resolve(root, 'components/app/AppShell.vue'), 'utf8'),
  performanceConstantsSource,
  performanceAppHelpersSource,
  readFileSync(resolve(root, 'styles/performance.css'), 'utf8')
].join('\n')

const performanceCssSource = readFileSync(resolve(root, 'styles/performance.css'), 'utf8')
for (const forbidden of [
  '.workbench-todo-',
  '.workbench-loading',
  '.profile-center-',
  '.review-action-',
  '.withdraw-confirm-card',
  '.withdraw-reason-textarea',
  '.review-create-',
  '.create-target-',
  '.template-',
  '.org-',
  '.summary-page',
  '.summary-department-',
  '.summary-month-',
  '.summary-action-buttons',
  '.summary-view-btn'
]) {
  if (performanceCssSource.includes(forbidden)) {
    throw new Error(`页面/弹窗专属样式不能继续放在 styles/performance.css: ${forbidden}`)
  }
}

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

for (const marker of ['钉钉H5应用', '绩效管理', '我的绩效', 'HRBP评价', '流程执行', '#1677ff']) {
  if (!componentSources.includes(marker)) {
    throw new Error(`钉钉 H5 组件缺少绩效工作台标识: ${marker}`)
  }
}

const appShellSource = readFileSync(resolve(root, 'components/app/AppShell.vue'), 'utf8')
for (const marker of [
  "import { useAppShellNavigation } from './useAppShellNavigation'",
  'useAppShellNavigation(props, emit)'
]) {
  if (!appShellSource.includes(marker)) {
    throw new Error(`AppShell.vue 必须复用应用壳导航组合式逻辑: ${marker}`)
  }
}
for (const marker of [
  'export function useAppShellNavigation(props, emit)',
  'const navIconMap =',
  'function readMenuLayout()',
  'function setMenuLayout(layout)',
  'function scrollDesktopPageTabs(direction)',
  'function scrollActiveDesktopPageTabIntoView()',
  'function setupDesktopPageTabsResizeObserver()',
  'function navIcon(name)',
  'const appDisplayTitle = computed',
  'const mobileSubmenuOpenItem = computed'
]) {
  if (!appShellNavigationSource.includes(marker)) {
    throw new Error(`components/app/useAppShellNavigation.js 必须承载菜单、页签和头像状态逻辑: ${marker}`)
  }
}
for (const forbidden of [
  'const navIconMap =',
  'function readMenuLayout()',
  'function setMenuLayout(',
  'function scrollDesktopPageTabs(direction)',
  'function setupDesktopPageTabsResizeObserver()',
  'function navIcon(name)',
  'function firstText('
]) {
  if (appShellSource.includes(forbidden)) {
    throw new Error(`AppShell.vue 不能继续承载导航纯逻辑和浏览器交互细节: ${forbidden}`)
  }
}
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
  '.desktop-page-tab.active::after'
]) {
  if (!componentSources.includes(marker)) {
    throw new Error(`PC 页面 tab 必须具备桌面样式并在移动端隐藏: ${marker}`)
  }
}
if (!/\.desktop-page-tabs\s*,\s*\.top-submenu-dropdown\s*\{[\s\S]*?display\s*:\s*none\s*;[\s\S]*?\}/.test(componentSources)) {
  throw new Error('PC 页面 tab 必须在移动端隐藏 desktop-page-tabs 和 top-submenu-dropdown')
}

console.log('钉钉 H5 微应用脚手架检查通过')
