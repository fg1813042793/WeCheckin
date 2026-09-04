import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const rootDir = resolve(currentDir, '..')

function read(relativePath) {
  return readFileSync(resolve(rootDir, relativePath), 'utf8')
}

function requireSnippet(source, snippet, description) {
  if (!source.includes(snippet)) {
    throw new Error(`${description}: missing ${snippet}`)
  }
}

function forbidSnippet(source, snippet, description) {
  if (source.includes(snippet)) {
    throw new Error(`${description}: forbidden ${snippet}`)
  }
}

const requestSource = read('src/utils/request.ts')
requireSnippet(requestSource, 'export interface ApiResponse<T = unknown>', 'request response data must default to unknown')
requireSnippet(requestSource, 'get<T = unknown>', 'GET response data must default to unknown')
requireSnippet(requestSource, 'post<T = unknown, D = unknown>', 'POST request and response data must default to unknown')
requireSnippet(requestSource, 'isRequestErrorNotified', 'request errors must expose the notification marker')
requireSnippet(requestSource, 'showRequestError', 'pages must be able to avoid duplicate request notifications')

const uploadApiSource = read('src/api/upload.ts')
requireSnippet(uploadApiSource, "'/api/v2/admin/uploads'", 'admin uploads must use the authenticated v2 endpoint')
requireSnippet(uploadApiSource, 'adminUploadRequest', 'Element Plus uploads must share one request adapter')
requireSnippet(uploadApiSource, 'ADMIN_UPLOAD_PERMISSION', 'admin uploads must enforce the matching frontend permission')
requireSnippet(uploadApiSource, 'canUploadAdminFile()', 'admin uploads must guard direct helper calls')

const uploadConsumers = [
  'src/views/event/index.vue',
  'src/views/user/index.vue',
  'src/views/mgr/index.vue',
  'src/views/news/index.vue',
  'src/views/enroll/index.vue',
  'src/views/survey/formkit/QuestionPreview.vue',
  'src/views/exam/formkit/QuestionPreview.vue',
]
for (const file of uploadConsumers) {
  const source = read(file)
  forbidSnippet(source, 'action="/upload"', `${file} must not use the unauthenticated legacy upload route`)
  forbidSnippet(source, "fetch('/upload'", `${file} must not bypass the shared request layer`)
}

for (const file of uploadConsumers.slice(0, 5)) {
  const source = read(file)
  requireSnippet(source, 'canUploadAdminFile', `${file} must disable uploads without upload:create`)
}

const publicFillFiles = [
  'src/views/survey/SurveyFillPC.vue',
  'src/views/survey/SurveyFillPC1.vue',
  'src/views/exam/ExamFillPC.vue',
  'src/views/exam/ExamFillPC1.vue',
]
const legacyPublicPaths = [
  '/passport/login_pwd',
  '/survey/view',
  '/survey/validate',
  '/survey/submit',
  '/exam/view',
  '/exam/result',
  '/exam/validate',
  '/exam/submit',
]
for (const file of publicFillFiles) {
  const source = read(file)
  for (const path of legacyPublicPaths) {
    forbidSnippet(source, path, `${file} must use the v2 public API contract`)
  }
  forbidSnippet(source, 'async function apiGet(', `${file} must use src/api/public.ts`)
  forbidSnippet(source, 'async function apiPost(', `${file} must use src/api/public.ts`)
}

const publicApiSource = read('src/api/public.ts')
for (const path of [
  '/api/v2/auth/password-login',
  '/api/v2/surveys/',
  '/api/v2/survey/validate',
  '/api/v2/exams/',
  '/api/v2/exam-results',
]) {
  requireSnippet(publicApiSource, path, 'public fill pages must centralize v2 endpoints')
}

const routesSource = read('src/router/adminRoutes.ts')
for (const line of routesSource.split(/\r?\n/)) {
  if (!line.includes("path: '") || !line.includes('component:')) continue
  if (!line.includes('menuPath:') && !line.includes('allowWithoutMenu: true')) {
    throw new Error(`admin route must declare its menu authorization source: ${line.trim()}`)
  }
}

const routerSource = read('src/router/index.ts')
for (const snippet of ['loadAdminAccessSnapshot', 'canAccessAdminRoute', "{ path: '/forbidden'"]) {
  requireSnippet(routerSource, snippet, 'router must enforce menu authorization for direct URLs')
}
forbidSnippet(routerSource, "redirect: '/dashboard'", 'the root route must not force users into an unauthorized dashboard')
requireSnippet(routerSource, "if (to.path === '/')", 'the root route must select an authorized destination')

const loginSource = read('src/views/login/index.vue')
forbidSnippet(loginSource, ": '/dashboard'", 'login must not default to an unauthorized dashboard')

const forbiddenSource = read('src/views/errors/ForbiddenView.vue')
forbidSnippet(forbiddenSource, "router.replace('/dashboard')", 'the forbidden page must return through authorized root routing')

const layoutSource = read('src/views/layout/index.vue')
requireSnippet(layoutSource, 'loadAdminAccessSnapshot', 'layout and router must share one access snapshot')
forbidSnippet(layoutSource, "request.get('/api/v2/home/setup'", 'admin settings must use an authenticated admin endpoint')
requireSnippet(layoutSource, "permissions.includes('admin:api:setup:list')", 'layout settings reads must honor setup:list')
forbidSnippet(layoutSource, ': [defaultVisitedTab, ...tabs]', 'stored tabs must not restore an unauthorized dashboard')

const permissionRequirements = {
  'src/views/news/index.vue': [
    'admin:menu:news:status',
    'admin:menu:news:vouch',
  ],
  'src/views/enroll/index.vue': [
    'admin:menu:enroll:status',
    'admin:menu:enroll:vouch',
    'admin:menu:enroll:export',
    'admin:menu:enroll:users',
  ],
  'src/views/event/index.vue': [
    'admin:menu:event:status',
    'admin:menu:event:vouch',
    'admin:menu:event:top',
    'admin:menu:event:users',
  ],
}
for (const [file, snippets] of Object.entries(permissionRequirements)) {
  const source = read(file)
  for (const snippet of snippets) {
    requireSnippet(source, snippet, `${file} must use the matching action permission`)
  }
}
forbidSnippet(read('src/views/news/index.vue'), "row.status === '1'", 'news status must use the numeric API contract')
