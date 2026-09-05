import assert from 'node:assert/strict'
import { Buffer } from 'node:buffer'
import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import process from 'node:process'
import ts from 'typescript'

const root = process.cwd()

const requiredFiles = [
  'src/api/user-feedback.ts',
  'src/pages/feedback/feedback-route-keys.ts',
  'src/pages/feedback/feedback.routes.ts',
  'src/pages/feedback/feedback.menu.ts',
  'src/pages/feedback/feedback-status.ts',
]

for (const file of requiredFiles) {
  assert.equal(existsSync(resolve(root, file)), true, `missing user feedback file: ${file}`)
}

function source(file) {
  return readFileSync(resolve(root, file), 'utf8')
}

function assertContains(file, patterns) {
  const content = source(file)
  for (const pattern of patterns) {
    assert.equal(content.includes(pattern), true, `${file} missing contract: ${pattern}`)
  }
}

async function loadTypeScriptModule(file) {
  const output = ts.transpileModule(source(file), {
    compilerOptions: {
      module: ts.ModuleKind.ESNext,
      target: ts.ScriptTarget.ES2022,
    },
  })
  const moduleUrl = `data:text/javascript;base64,${Buffer.from(output.outputText).toString('base64')}`
  return import(moduleUrl)
}

assertContains('src/api/dingtalk-h5/base.ts', [
  'uploadMultipartFiles',
  'FormData',
  'fetch(',
  'response.blob()',
  'AbortController',
  `form.append(options.fileFieldName || 'images'`,
  '当前上传仅支持 H5',
])

assertContains('src/api/user-feedback.ts', [
  '/api/v2/dingtalk/h5/user-feedbacks',
  '/overview',
  '/supplements',
  'dingtalk_h5:api:feedback:list',
  'dingtalk_h5:api:feedback:create',
  'dingtalk_h5:api:feedback:detail',
  'dingtalk_h5:api:feedback:supplement',
  'getUserFeedbackOverview',
  'listUserFeedbacks',
  'createUserFeedback',
  'getUserFeedbackDetail',
  'supplementUserFeedback',
])

assertContains('src/pages/feedback/feedback.menu.ts', [
  `key: 'feedback'`,
  'contentKey: FEEDBACK_CONTENT_KEY',
  `permissionKey: 'dingtalk_h5:menu:feedback'`,
])

assertContains('src/config/app-navigation.ts', [
  'feedbackMenuPages',
  'feedbackRootNavItem',
  'workflowRootNavItem,\n  feedbackRootNavItem',
  '...workflowMenuPages,\n  ...feedbackMenuPages',
])

assertContains('src/config/app-content-routes.ts', [
  'feedbackContentRoutes',
  'resolveFeedbackContentComponent',
  'resolveWorkflowContentComponent',
  'appContentRoutes.dashboard.component',
])

assertContains('src/config/app-icons.ts', [`feedback: 'chat'`])

const routeKeys = await loadTypeScriptModule('src/pages/feedback/feedback-route-keys.ts')
assert.equal(routeKeys.FEEDBACK_CONTENT_KEY, 'feedback')
assert.equal(routeKeys.FEEDBACK_CREATE_CONTENT_KEY, 'feedback:create')
assert.equal(routeKeys.feedbackDetailContentKey(' 123 '), 'feedback:detail:123')
assert.equal(routeKeys.feedbackDetailContentKey('a/b'), 'feedback:detail:a%2Fb')
assert.equal(routeKeys.feedbackDetailContentKey(''), '')
assert.equal(routeKeys.feedbackDetailIdFromContentKey('feedback:detail:a%2Fb'), 'a/b')
assert.equal(routeKeys.feedbackDetailIdFromContentKey('feedback:detail:'), '')
assert.equal(routeKeys.feedbackDetailIdFromContentKey('feedback:detail:%E0%A4%A'), '')
assert.equal(routeKeys.normalizeFeedbackDynamicContentKey('feedback:detail:%E0%A4%A'), '')
assert.equal(routeKeys.normalizeFeedbackDynamicContentKey('feedback:detail:%31'), '')
assert.equal(routeKeys.normalizeFeedbackDynamicContentKey('feedback:create'), 'feedback:create')

const statuses = await loadTypeScriptModule('src/pages/feedback/feedback-status.ts')
assert.equal(statuses.feedbackStatusMeta('pending').type, 'warning')
assert.equal(statuses.feedbackStatusMeta('processing').type, 'warning')
assert.equal(statuses.feedbackStatusMeta('resolved').type, 'success')
assert.equal(statuses.feedbackStatusMeta('closed').type, 'info')

console.log('user feedback contracts ok')
