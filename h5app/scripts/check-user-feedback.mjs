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

async function loadBaseUploadContractModule() {
  const file = 'src/api/dingtalk-h5/base.ts'
  const content = source(file)
  const sourceFile = ts.createSourceFile(file, content, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS)
  const selectedNames = new Set([
    'MultipartUploadFile',
    'UploadMultipartFilesOptions',
    'parseUploadResponseData',
    'SAFE_IMAGE_MIME_TYPES',
    'safeImageMimeType',
    'imageExtension',
    'safeUploadFilename',
    'uploadMultipartFilesInH5',
  ])
  const selectedStatements = sourceFile.statements.filter((statement) => {
    if (ts.isFunctionDeclaration(statement) || ts.isInterfaceDeclaration(statement))
      return Boolean(statement.name && selectedNames.has(statement.name.text))
    if (!ts.isVariableStatement(statement))
      return false
    return statement.declarationList.declarations.some((declaration) => {
      return ts.isIdentifier(declaration.name) && selectedNames.has(declaration.name.text)
    })
  })
  assert.equal(selectedStatements.length, selectedNames.size, 'base upload contract declarations are incomplete')

  const moduleSource = `
type UploadFormValue = string | number | boolean
interface ApiEnvelope<T = unknown> { code?: number, data?: T }
const DINGTALK_H5_CONFIG = { CLIENT_PLATFORM: 'dingtalk-h5' }
function authToken() { return 'test-token' }
function buildApiUrl(url: string) { return url }
${selectedStatements.map(statement => statement.getText(sourceFile)).join('\n')}
export { safeImageMimeType, safeUploadFilename, uploadMultipartFilesInH5 }
`
  const output = ts.transpileModule(moduleSource, {
    compilerOptions: {
      module: ts.ModuleKind.ESNext,
      target: ts.ScriptTarget.ES2022,
    },
  })
  const moduleUrl = `data:text/javascript;base64,${Buffer.from(output.outputText).toString('base64')}`
  return import(moduleUrl)
}

async function assertMultipartUploadContract(uploadContract) {
  assert.equal(
    uploadContract.safeUploadFilename({ filePath: 'blob:https://example.test/temporary-id' }, 0, 'image/png'),
    'image-1.png',
  )
  assert.equal(
    uploadContract.safeUploadFilename({ filePath: 'blob:test', name: 'evidence.jpg' }, 0, 'image/png'),
    'evidence.png',
  )
  assert.equal(
    uploadContract.safeImageMimeType({ filePath: 'blob:test', mimeType: 'image/png' }, 'image/jpeg'),
    'image/jpeg',
  )
  assert.equal(
    uploadContract.safeImageMimeType({ filePath: 'blob:test', mimeType: 'image/webp' }, ''),
    'image/webp',
  )
  assert.throws(
    () => uploadContract.safeImageMimeType({ filePath: 'blob:test', mimeType: 'image/png' }, 'text/plain'),
    /不支持的图片类型/,
  )

  const originalFetch = globalThis.fetch
  const originalSetTimeout = globalThis.setTimeout
  const originalClearTimeout = globalThis.clearTimeout
  const timerToken = { kind: 'feedback-upload-timeout' }
  let timerStarted = false
  let timerCleared = false
  let timeoutCallback
  let localFetchSignal
  let postFetchSignal
  try {
    globalThis.setTimeout = (callback, delay) => {
      assert.equal(typeof callback, 'function')
      assert.equal(delay, 30000)
      timerStarted = true
      timeoutCallback = callback
      return timerToken
    }
    globalThis.clearTimeout = (token) => {
      assert.equal(token, timerToken)
      timerCleared = true
    }
    globalThis.fetch = async (input, init) => {
      if (String(input) === 'blob:test-image') {
        assert.equal(timerStarted, true, 'upload timeout must start before reading the first image')
        localFetchSignal = init?.signal
        return {
          ok: true,
          blob: async () => new Blob(['image'], { type: 'image/png' }),
        }
      }
      postFetchSignal = init?.signal
      const uploadedImage = init?.body?.get('images')
      assert.equal(uploadedImage?.name, 'image-1.png')
      assert.equal(uploadedImage?.type, 'image/png')
      return {
        ok: true,
        text: async () => JSON.stringify({ code: 0, data: { id: 1 } }),
      }
    }

    const response = await uploadContract.uploadMultipartFilesInH5(
      '/api/v2/dingtalk/h5/user-feedbacks',
      [{ filePath: 'blob:test-image' }],
      {},
    )
    assert.equal(response.data.id, 1)
    assert.ok(localFetchSignal instanceof AbortSignal)
    assert.equal(localFetchSignal, postFetchSignal, 'local image and POST fetch must share one abort signal')
    assert.equal(timerCleared, true, 'successful upload must clear its total timeout')
    timeoutCallback()
    assert.equal(localFetchSignal.aborted, true, 'total timeout callback must abort the shared signal')

    timerStarted = false
    timerCleared = false
    globalThis.fetch = async (_input, init) => {
      assert.equal(timerStarted, true, 'failed local image reads must still be covered by the total timeout')
      assert.ok(init?.signal instanceof AbortSignal)
      throw new Error('local image read failed')
    }
    await assert.rejects(
      uploadContract.uploadMultipartFilesInH5(
        '/api/v2/dingtalk/h5/user-feedbacks',
        [{ filePath: 'blob:failed-image', mimeType: 'image/png' }],
        {},
      ),
      /local image read failed/,
    )
    assert.equal(timerCleared, true, 'failed local image reads must clear the total timeout')
  }
  finally {
    globalThis.fetch = originalFetch
    globalThis.setTimeout = originalSetTimeout
    globalThis.clearTimeout = originalClearTimeout
  }
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
  `label: '我的反馈'`,
  `icon: 'chat'`,
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

const uploadContract = await loadBaseUploadContractModule()
await assertMultipartUploadContract(uploadContract)

console.log('user feedback contracts ok')
