import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const adminDir = resolve(currentDir, '..')
const pagePath = resolve(adminDir, 'src/views/survey/template/index.vue')
const composablePath = resolve(adminDir, 'src/views/survey/template/useSurveyNotificationTemplates.ts')
const examPagePath = resolve(adminDir, 'src/views/exam/template/index.vue')
const examComposablePath = resolve(adminDir, 'src/views/exam/template/useExamNotificationTemplates.ts')
const routeSource = readFileSync(resolve(adminDir, 'src/router/adminRoutes.ts'), 'utf8')
const setupSource = readFileSync(resolve(adminDir, 'src/views/setup/index.vue'), 'utf8')
const apiSource = readFileSync(resolve(adminDir, 'src/api/index.ts'), 'utf8')

for (const path of [pagePath, composablePath]) {
  if (!existsSync(path)) throw new Error(`survey template page file missing: ${path}`)
}
for (const path of [examPagePath, examComposablePath]) {
  if (!existsSync(path)) throw new Error(`exam template page file missing: ${path}`)
}

const pageSource = readFileSync(pagePath, 'utf8')
const composableSource = readFileSync(composablePath, 'utf8')
const examPageSource = readFileSync(examPagePath, 'utf8')
const examComposableSource = readFileSync(examComposablePath, 'utf8')
for (const snippet of [
  '<AdminPageShell',
  '<AdminPageHeader',
  '问卷通知模版',
  '内置模版',
  '自定义模版',
  "hasPerm('admin:menu:survey:templates:edit')",
  'openAdd',
  'openEdit',
  'deletePreset',
  '<el-button circle :icon="Refresh" title="刷新" :loading="loading" @click="loadTemplates" />',
  'label="操作" width="170" fixed="right"',
  'class="admin-table-actions"',
]) {
  if (!pageSource.includes(snippet)) throw new Error(`survey template page missing: ${snippet}`)
}
for (const snippet of ['loadTemplates', 'saveTemplates', 'surveyNotificationTemplatesGet', 'surveyNotificationTemplatesSave']) {
  if (!composableSource.includes(snippet)) throw new Error(`survey template composable missing: ${snippet}`)
}
for (const snippet of ['<AdminPageShell', '<AdminPageHeader', '考试通知模版', '{score}', '{fullScore}', "hasPerm('admin:menu:exam:templates:edit')", 'openAdd', 'openEdit', 'deletePreset']) {
  if (!examPageSource.includes(snippet)) throw new Error(`exam template page missing: ${snippet}`)
}
for (const snippet of ['loadTemplates', 'saveTemplates', 'examNotificationTemplatesGet', 'examNotificationTemplatesSave']) {
  if (!examComposableSource.includes(snippet)) throw new Error(`exam template composable missing: ${snippet}`)
}
for (const snippet of [
  "path: 'survey/templates'",
  "menuPath: '/survey/templates'",
  "requiredPermission: 'admin:menu:survey:templates'",
  "pattern: 'list'",
]) {
  if (!routeSource.includes(snippet)) throw new Error(`survey template route missing: ${snippet}`)
}
for (const snippet of ["path: 'exam/templates'", "menuPath: '/exam/templates'", "requiredPermission: 'admin:menu:exam:templates'", "pattern: 'list'"]) {
  if (!routeSource.includes(snippet)) throw new Error(`exam template route missing: ${snippet}`)
}
for (const snippet of ['surveyNotificationTemplatesGet()', 'surveyNotificationTemplatesSave(data:']) {
  if (!apiSource.includes(snippet)) throw new Error(`survey template API missing: ${snippet}`)
}
for (const snippet of ['examNotificationTemplatesGet()', 'examNotificationTemplatesSave(data:']) {
  if (!apiSource.includes(snippet)) throw new Error(`exam template API missing: ${snippet}`)
}
for (const forbidden of ['问卷通知模版', 'builtinPresets', 'customPresets', 'survey-template-presets']) {
  if (setupSource.includes(forbidden)) throw new Error(`system setup still owns survey templates: ${forbidden}`)
}
