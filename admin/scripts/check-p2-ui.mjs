import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const srcDir = resolve(currentDir, '../src')

function read(path) {
  return readFileSync(resolve(srcDir, path), 'utf8')
}

function requireSnippet(source, snippet, label) {
  if (!source.includes(snippet)) {
    throw new Error(`${label} missing required snippet: ${snippet}`)
  }
}

function forbidSnippet(source, snippet, label) {
  if (source.includes(snippet)) {
    throw new Error(`${label} still contains old snippet: ${snippet}`)
  }
}

const dashboard = read('views/dashboard/index.vue')
for (const snippet of [
  'class="admin-page dashboard-page"',
  'dashboard-toolbar',
  'metric-grid',
  'quick-actions',
  'dashboard-trend',
  'dashboard-todos',
  'dashboard-health',
  'DashboardTrendChart',
  'hasPerm(',
  'updatedAt',
  'goRoute(',
]) {
  requireSnippet(dashboard, snippet, 'admin dashboard')
}
forbidSnippet(dashboard, '<template #header>控制台</template>', 'admin dashboard')
forbidSnippet(dashboard, 'const pendingItems = [', 'admin dashboard')
forbidSnippet(dashboard, 'const moduleHealth = [', 'admin dashboard')
forbidSnippet(dashboard, 'Record<string, any>', 'admin dashboard')
forbidSnippet(dashboard, "from 'vue-echarts'", 'admin dashboard')

const login = read('views/login/index.vue')
for (const snippet of [
  'login-shell',
  'login-brand-panel',
  'login-form-panel',
  'login-feature-list',
  'prefix-icon="User"',
  'prefix-icon="Lock"',
  '@media (max-width: 768px)',
]) {
  requireSnippet(login, snippet, 'admin login page')
}
forbidSnippet(login, 'login-container', 'admin login page')

const routes = read('router/adminRoutes.ts')
for (const snippet of [
  "path: 'exam/responses'",
  "path: 'exam/statistic'",
  "path: 'exam/formkit'",
  "path: 'exam/formkit/report'",
]) {
  requireSnippet(routes, snippet, 'admin route config')
}

const surveyList = read('views/survey/SurveyList.vue')
const examList = read('views/exam/ExamList.vue')
const surveyListStyles = read('styles/survey-list.css')
for (const [source, label] of [[surveyList, 'survey list'], [examList, 'exam list']]) {
  for (const snippet of [':show-actions="false"', 'class="survey-search-inline-actions"', 'native-type="submit"']) {
    requireSnippet(source, snippet, label)
  }
}
for (const snippet of ['.survey-search-inline-actions', 'justify-content:flex-start;', 'flex:0 0 auto;']) {
  requireSnippet(surveyListStyles, snippet, 'survey and exam list styles')
}
