import assert from 'node:assert/strict'
import { Buffer } from 'node:buffer'
import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import process from 'node:process'
import ts from 'typescript'

const root = process.cwd()

const requiredFiles = [
  'src/api/user-feedback.ts',
  'src/common/dingtalk-h5-auth-expiry.ts',
  'src/pages/feedback/feedback-route-keys.ts',
  'src/pages/feedback/feedback.routes.ts',
  'src/pages/feedback/feedback.menu.ts',
  'src/pages/feedback/feedback-status.ts',
  'src/pages/feedback/components/FeedbackCenter.vue',
  'src/pages/feedback/components/feedback-center-state.ts',
  'src/pages/feedback/components/FeedbackStatusOverview.vue',
  'src/pages/feedback/components/FeedbackList.vue',
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

function assertInterfaceShape(file, interfaceName, contract) {
  const sourceFile = ts.createSourceFile(file, source(file), ts.ScriptTarget.Latest, true, ts.ScriptKind.TS)
  const declaration = sourceFile.statements.find(statement => (
    ts.isInterfaceDeclaration(statement) && statement.name.text === interfaceName
  ))
  assert.ok(declaration, `${file} missing interface: ${interfaceName}`)

  const properties = new Map()
  for (const member of declaration.members) {
    if (!ts.isPropertySignature(member) || !member.type || !member.name)
      continue
    const name = member.name.getText(sourceFile).replace(/^['"]|['"]$/g, '')
    properties.set(name, {
      optional: Boolean(member.questionToken),
      type: member.type.getText(sourceFile).replace(/\s+/g, ' ').trim(),
    })
  }
  const expectedNames = [...contract.required, ...contract.optional].sort()
  assert.deepEqual([...properties.keys()].sort(), expectedNames, `${interfaceName} fields drifted`)
  for (const name of contract.required) {
    assert.equal(properties.get(name)?.optional, false, `${interfaceName}.${name} must be required`)
  }
  for (const name of contract.optional) {
    assert.equal(properties.get(name)?.optional, true, `${interfaceName}.${name} must be optional`)
  }
  for (const [name, expectedType] of Object.entries(contract.types || {})) {
    assert.equal(properties.get(name)?.type, expectedType, `${interfaceName}.${name} type drifted`)
  }
  assert.deepEqual(
    (declaration.heritageClauses || []).flatMap(clause => clause.types.map(type => type.expression.getText(sourceFile))).sort(),
    [...(contract.extends || [])].sort(),
    `${interfaceName} inheritance drifted`,
  )
}

function vueScriptSourceFile(file) {
  const content = source(file)
  const match = content.match(/<script setup lang="ts">([\s\S]*?)<\/script>/)
  assert.ok(match, `${file} must contain a TypeScript script setup block`)
  return ts.createSourceFile(file, match[1], ts.ScriptTarget.Latest, true, ts.ScriptKind.TS)
}

function findFunction(sourceFile, functionName) {
  const declaration = sourceFile.statements.find(statement => (
    ts.isFunctionDeclaration(statement) && statement.name?.text === functionName
  ))
  assert.ok(declaration, `${sourceFile.fileName} missing function: ${functionName}`)
  return declaration
}

function functionCalls(sourceFile, functionName) {
  const calls = []
  const declaration = findFunction(sourceFile, functionName)
  function visit(node) {
    if (ts.isCallExpression(node))
      calls.push(node.expression.getText(sourceFile))
    ts.forEachChild(node, visit)
  }
  visit(declaration)
  return calls
}

function lifecycleCallbackCalls(sourceFile, lifecycleName) {
  const lifecycleCall = sourceFile.statements
    .filter(ts.isExpressionStatement)
    .map(statement => statement.expression)
    .find(expression => (
      ts.isCallExpression(expression)
      && expression.expression.getText(sourceFile) === lifecycleName
    ))
  assert.ok(lifecycleCall && ts.isCallExpression(lifecycleCall), `${sourceFile.fileName} missing ${lifecycleName}`)
  const callback = lifecycleCall.arguments[0]
  assert.ok(callback && (ts.isArrowFunction(callback) || ts.isFunctionExpression(callback)), `${lifecycleName} must receive a callback`)
  const calls = []
  function visit(node) {
    if (ts.isCallExpression(node))
      calls.push(node.expression.getText(sourceFile))
    ts.forEachChild(node, visit)
  }
  visit(callback)
  return calls
}

async function loadBaseUploadContractModule() {
  const baseFile = 'src/api/dingtalk-h5/base.ts'
  const baseSourceFile = ts.createSourceFile(baseFile, source(baseFile), ts.ScriptTarget.Latest, true, ts.ScriptKind.TS)
  const baseNames = new Set([
    'MultipartUploadFile',
    'UploadMultipartFilesOptions',
    'trimRightSlash',
    'isAbsoluteUrl',
    'withBaseUrl',
    'authToken',
    'buildApiUrl',
    'parseUploadResponseData',
    'SAFE_IMAGE_MIME_TYPES',
    'safeImageMimeType',
    'imageExtension',
    'safeUploadFilename',
    'isAbortError',
    'uploadMultipartFilesInH5',
  ])
  const baseStatements = baseSourceFile.statements.filter((statement) => {
    if (ts.isFunctionDeclaration(statement) || ts.isInterfaceDeclaration(statement))
      return Boolean(statement.name && baseNames.has(statement.name.text))
    if (!ts.isVariableStatement(statement))
      return false
    return statement.declarationList.declarations.some((declaration) => {
      return ts.isIdentifier(declaration.name) && baseNames.has(declaration.name.text)
    })
  })
  assert.equal(baseStatements.length, baseNames.size, 'base upload contract declarations are incomplete')

  const authFile = 'src/common/dingtalk-h5-auth-expiry.ts'
  const authSourceFile = ts.createSourceFile(authFile, source(authFile), ts.ScriptTarget.Latest, true, ts.ScriptKind.TS)
  const authNames = new Set([
    'LOGIN_EXPIRED_MESSAGES',
    'isApiEnvelope',
    'isDingTalkH5AuthExpired',
    'handleDingTalkH5AuthExpired',
  ])
  const authStatements = authSourceFile.statements.filter((statement) => {
    if (ts.isFunctionDeclaration(statement))
      return Boolean(statement.name && authNames.has(statement.name.text))
    if (!ts.isVariableStatement(statement))
      return false
    return statement.declarationList.declarations.some((declaration) => {
      return ts.isIdentifier(declaration.name) && authNames.has(declaration.name.text)
    })
  })
  assert.equal(authStatements.length, authNames.size, 'auth expiry contract declarations are incomplete')

  const moduleSource = `
type UploadFormValue = string | number | boolean
type QueryValue = string | number | boolean | null | undefined
interface ApiEnvelope<T = unknown> { code?: number, msg?: string, message?: string, data?: T }
const DINGTALK_H5_CONFIG = {
  BASE_URL: 'https://api.example.test/root/',
  CLIENT_PLATFORM: 'dingtalk-h5',
  TOKEN_KEY: 'DT_H5_TOKEN',
}
const DINGTALK_H5_AUTH_EXPIRED_EVENT = 'dingtalk-h5-auth-expired'
const uni = globalThis.__feedbackContractUni
${authStatements.map(statement => statement.getText(authSourceFile)).join('\n')}
${baseStatements.map(statement => statement.getText(baseSourceFile)).join('\n')}
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

async function assertMultipartUploadContract(uploadContract, authHarness) {
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
    authHarness.storage.set('DT_H5_TOKEN', 'Bearer contract-token')
    authHarness.events.length = 0
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
      assert.equal(String(input), 'https://api.example.test/root/api/v2/dingtalk/h5/user-feedbacks')
      assert.equal(init?.headers?.Authorization, 'Bearer contract-token')
      assert.equal(init?.headers?.['X-Client-Platform'], 'dingtalk-h5')
      assert.equal(init?.headers?.['X-Contract-Test'], 'kept')
      assert.equal(
        Object.keys(init?.headers || {}).some(key => key.toLowerCase() === 'content-type'),
        false,
        'fetch must let FormData set its own multipart boundary',
      )
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
      {
        header: {
          'Content-Type': 'application/json',
          'content-type': 'text/plain',
          'X-Contract-Test': 'kept',
        },
      },
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

    timerCleared = false
    let pendingFetchSignal
    globalThis.fetch = async (_input, init) => {
      pendingFetchSignal = init?.signal
      return new Promise((_resolve, reject) => {
        init?.signal?.addEventListener('abort', () => {
          const error = new Error('fetch aborted')
          error.name = 'AbortError'
          reject(error)
        }, { once: true })
      })
    }
    const timedOutUpload = uploadContract.uploadMultipartFilesInH5(
      '/api/v2/dingtalk/h5/user-feedbacks',
      [],
      {},
    )
    await Promise.resolve()
    timeoutCallback()
    await assert.rejects(timedOutUpload, error => (
      error instanceof Error && error.message === '上传超时，请检查网络后重试'
    ))
    assert.equal(pendingFetchSignal?.aborted, true, 'an in-flight upload fetch must be aborted by the total timeout')
    assert.equal(timerCleared, true, 'timed out uploads must clear the total timeout')

    const uploadEndpoint = '/api/v2/dingtalk/h5/user-feedbacks'
    const responseWith = (status, payload) => ({
      ok: status >= 200 && status < 300,
      status,
      text: async () => payload,
    })

    const businessFailure = { code: 43001, msg: '反馈状态不允许当前操作' }
    authHarness.storage.set('DT_H5_TOKEN', 'Bearer business-token')
    authHarness.events.length = 0
    globalThis.fetch = async () => responseWith(200, JSON.stringify(businessFailure))
    await assert.rejects(
      uploadContract.uploadMultipartFilesInH5(uploadEndpoint, [], {}),
      (error) => {
        assert.deepEqual(error, businessFailure)
        return true
      },
    )
    assert.equal(authHarness.storage.get('DT_H5_TOKEN'), 'Bearer business-token')
    assert.deepEqual(authHarness.events, [])

    const httpFailure = { code: 403, msg: '无权上传反馈' }
    globalThis.fetch = async () => responseWith(403, JSON.stringify(httpFailure))
    await assert.rejects(
      uploadContract.uploadMultipartFilesInH5(uploadEndpoint, [], {}),
      (error) => {
        assert.deepEqual(error, httpFailure)
        return true
      },
    )

    globalThis.fetch = async () => responseWith(200, '{invalid-json')
    await assert.rejects(
      uploadContract.uploadMultipartFilesInH5(uploadEndpoint, [], {}),
      error => error instanceof Error && error.message === '上传响应异常',
    )

    const expiredEnvelope = { code: 401, msg: '登录已过期' }
    authHarness.storage.set('DT_H5_TOKEN', 'Bearer expired-token')
    authHarness.events.length = 0
    globalThis.fetch = async () => responseWith(200, JSON.stringify(expiredEnvelope))
    await assert.rejects(
      uploadContract.uploadMultipartFilesInH5(uploadEndpoint, [], {}),
      (error) => {
        assert.deepEqual(error, expiredEnvelope)
        return true
      },
    )
    assert.equal(authHarness.storage.has('DT_H5_TOKEN'), false, 'login expiry must clear the H5 token')
    assert.deepEqual(authHarness.events, [['dingtalk-h5-auth-expired', expiredEnvelope]])

    const networkError = new Error('socket closed')
    globalThis.fetch = async () => {
      throw networkError
    }
    await assert.rejects(
      uploadContract.uploadMultipartFilesInH5(uploadEndpoint, [], {}),
      error => error === networkError,
    )

    const externalAbort = new Error('request cancelled outside upload timer')
    externalAbort.name = 'AbortError'
    globalThis.fetch = async () => {
      throw externalAbort
    }
    await assert.rejects(
      uploadContract.uploadMultipartFilesInH5(uploadEndpoint, [], {}),
      error => error === externalAbort,
    )
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
assertContains('src/common/dingtalk-h5-auth-expiry.ts', [
  'isApiEnvelope',
  'isDingTalkH5AuthExpired',
  'handleDingTalkH5AuthExpired',
  'DINGTALK_H5_AUTH_EXPIRED_EVENT',
  'removeStorageSync',
  'uni.$emit',
])
assertContains('src/common/http.interceptor.ts', [
  `from '@/common/dingtalk-h5-auth-expiry'`,
  'isApiEnvelope(rawData)',
  'handleDingTalkH5AuthExpired(rawData)',
])
assert.equal(
  source('src/common/http.interceptor.ts').includes('const LOGIN_EXPIRED_MESSAGES'),
  false,
  'http interceptor must reuse the shared auth expiry helper',
)

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
assertContains('src/pages/feedback/feedback-route-keys.ts', [
  'feedbackDetailContentKey(id: string)',
  'BigInt(value)',
])

const feedbackApiFile = 'src/api/user-feedback.ts'
assertInterfaceShape(feedbackApiFile, 'UserFeedbackOverview', {
  required: ['pending', 'processing', 'resolved', 'closed'],
  optional: [],
  types: { pending: 'number', processing: 'number', resolved: 'number', closed: 'number' },
})
assertInterfaceShape(feedbackApiFile, 'UserFeedbackAttachment', {
  required: ['id', 'originalName', 'sizeBytes', 'url', 'sortOrder', 'createdAt'],
  optional: ['contentType'],
  types: { id: 'number', contentType: 'string', sizeBytes: 'number' },
})
assertInterfaceShape(feedbackApiFile, 'UserFeedbackMessage', {
  required: ['id', 'messageType', 'authorType', 'authorId', 'content', 'attachments', 'createdAt'],
  optional: ['authorName', 'fromStatus', 'toStatus'],
  types: {
    messageType: 'UserFeedbackMessageType',
    authorType: 'UserFeedbackAuthorType',
    attachments: 'UserFeedbackAttachment[]',
  },
})
assertInterfaceShape(feedbackApiFile, 'UserFeedbackSummary', {
  required: [
    'id',
    'feedbackNo',
    'submitterId',
    'summary',
    'imageCount',
    'status',
    'version',
    'lastActivityAt',
    'createdAt',
    'updatedAt',
  ],
  optional: ['submitterName', 'handlerId', 'handlerName', 'resolvedAt', 'closedAt'],
  types: { id: 'number', status: 'UserFeedbackStatus', version: 'number' },
})
assertInterfaceShape(feedbackApiFile, 'UserFeedbackDetail', {
  required: ['messages', 'allowsSupplement'],
  optional: [],
  types: { messages: 'UserFeedbackMessage[]', allowsSupplement: 'boolean' },
  extends: ['UserFeedbackSummary'],
})
assertInterfaceShape(feedbackApiFile, 'UserFeedbackList', {
  required: ['list', 'total', 'page', 'pageSize'],
  optional: [],
  types: { list: 'UserFeedbackSummary[]', total: 'number', page: 'number', pageSize: 'number' },
})
assertInterfaceShape(feedbackApiFile, 'UserFeedbackListQuery', {
  required: [],
  optional: ['keyword', 'status', 'page', 'pageSize'],
  types: { keyword: 'string', status: 'UserFeedbackStatus', page: 'number', pageSize: 'number' },
})
assertInterfaceShape(feedbackApiFile, 'CreateUserFeedbackInput', {
  required: ['content', 'requestId'],
  optional: ['images'],
  types: { content: 'string', requestId: 'string', images: 'readonly MultipartUploadFile[]' },
})
assertInterfaceShape(feedbackApiFile, 'SupplementUserFeedbackInput', {
  required: ['content', 'requestId', 'version'],
  optional: ['images'],
  types: {
    content: 'string',
    requestId: 'string',
    version: 'number',
    images: 'readonly MultipartUploadFile[]',
  },
})

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
assert.equal(routeKeys.feedbackDetailContentKey('18446744073709551615'), 'feedback:detail:18446744073709551615')
assert.equal(routeKeys.feedbackDetailContentKey('a/b'), '')
assert.equal(routeKeys.feedbackDetailContentKey(''), '')
assert.equal(routeKeys.feedbackDetailIdFromContentKey('feedback:detail:18446744073709551615'), '18446744073709551615')
for (const invalidID of ['0', '-1', '1.5', 'a', '18446744073709551616']) {
  assert.equal(routeKeys.feedbackDetailContentKey(invalidID), '', `invalid detail id must be rejected: ${invalidID}`)
  assert.equal(routeKeys.feedbackDetailIdFromContentKey(`feedback:detail:${invalidID}`), '', `invalid detail key must be rejected: ${invalidID}`)
}
assert.equal(routeKeys.feedbackDetailIdFromContentKey('feedback:detail:a%2Fb'), '')
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

const authHarness = {
  storage: new Map(),
  events: [],
}
globalThis.__feedbackContractUni = {
  getStorageSync(key) {
    return authHarness.storage.get(key) || ''
  },
  removeStorageSync(key) {
    authHarness.storage.delete(key)
  },
  $emit(event, payload) {
    authHarness.events.push([event, payload])
  },
}
try {
  const uploadContract = await loadBaseUploadContractModule()
  await assertMultipartUploadContract(uploadContract, authHarness)
}
finally {
  delete globalThis.__feedbackContractUni
}

assertContains('src/pages/feedback/feedback.routes.ts', [
  `import FeedbackCenter from './components/FeedbackCenter.vue'`,
  '[FEEDBACK_CONTENT_KEY]: FeedbackCenter',
])

assertContains('src/pages/feedback/components/FeedbackCenter.vue', [
  'useLocale(\'feedback\')',
  'getUserFeedbackOverview',
  'listUserFeedbacks',
  'FeedbackStatusOverview',
  'FeedbackList',
  'FEEDBACK_CREATE_CONTENT_KEY',
  'feedbackDetailContentKey',
  'appContent.openDynamicTab',
  'appContent.refreshTick',
  `appContent.currentKey === FEEDBACK_CONTENT_KEY`,
  'registerFeedbackVisibilityRefresh',
  'loadOverview()',
  'loadFeedbacks()',
  '@select="selectStatus"',
  '@retry="loadOverview"',
  '@retry="loadFeedbacks"',
])

assertContains('src/pages/feedback/components/FeedbackStatusOverview.vue', [
  'pending: \'warning\'',
  'processing: \'warning\'',
  'resolved: \'success\'',
  'closed: \'info\'',
  'emit(\'select\', item.status)',
  'u-loading',
  'u-icon name="reload"',
  't(\'overviewFailed\')',
])

assertContains('src/pages/feedback/components/FeedbackList.vue', [
  'feedbackNo',
  'summary',
  'imageCount',
  'lastActivityAt',
  'feedbackStatusMeta',
  'u-pagination',
  'emit(\'retry\')',
  'emit(\'open\', item)',
  'u-loading',
])

for (const componentFile of [
  'src/pages/feedback/components/FeedbackCenter.vue',
  'src/pages/feedback/components/FeedbackStatusOverview.vue',
  'src/pages/feedback/components/FeedbackList.vue',
]) {
  assert.doesNotMatch(source(componentFile), /[\u3400-\u9FFF]/, `${componentFile} must read visible copy from locale`)
}

const feedbackCenterSource = source('src/pages/feedback/components/FeedbackCenter.vue')
assert.equal(/setInterval|setTimeout\s*\([^,]+,\s*\d+\s*\)/.test(feedbackCenterSource), false, 'feedback center must not poll')
const feedbackCenterScript = vueScriptSourceFile('src/pages/feedback/components/FeedbackCenter.vue')
assert.deepEqual(
  functionCalls(feedbackCenterScript, 'openCreate').filter(call => ['createFeedbackDynamicTab', 'appContent.openDynamicTab'].includes(call)),
  ['createFeedbackDynamicTab', 'appContent.openDynamicTab'],
  'openCreate must build and open one stable dynamic tab',
)
assert.deepEqual(
  functionCalls(feedbackCenterScript, 'openFeedback').filter(call => ['feedbackDetailContentKey', 'createFeedbackDynamicTab', 'appContent.openDynamicTab'].includes(call)),
  ['feedbackDetailContentKey', 'createFeedbackDynamicTab', 'appContent.openDynamicTab'],
  'openFeedback must build and open the canonical detail tab',
)
assert.ok(functionCalls(feedbackCenterScript, 'loadOverview').includes('requestDeduper.run'), 'overview must reuse an in-flight request')
assert.ok(functionCalls(feedbackCenterScript, 'loadFeedbacks').includes('requestDeduper.run'), 'list must reuse an in-flight request')
assert.ok(functionCalls(feedbackCenterScript, 'loadFeedbacks').includes('resolveFeedbackListPage'), 'list must correct an out-of-range page')
assert.ok(functionCalls(feedbackCenterScript, 'loadFeedbacks').filter(call => call === 'loadFeedbacks').length >= 1, 'corrected page must be queried again')
assert.ok(lifecycleCallbackCalls(feedbackCenterScript, 'onMounted').includes('registerFeedbackVisibilityRefresh'), 'mount must register visibility refresh')
assert.ok(lifecycleCallbackCalls(feedbackCenterScript, 'onBeforeUnmount').includes('removeVisibilityListener'), 'unmount must remove visibility refresh')
assert.ok(lifecycleCallbackCalls(feedbackCenterScript, 'onBeforeUnmount').includes('requestDeduper.clear'), 'unmount must clear request dedupe state')

const feedbackCenterState = await loadTypeScriptModule('src/pages/feedback/components/feedback-center-state.ts')
assert.deepEqual(feedbackCenterState.resolveFeedbackListPage(3, 12, 12), { page: 1, shouldReload: true })
assert.deepEqual(feedbackCenterState.resolveFeedbackListPage(3, 25, 12), { page: 3, shouldReload: false })
assert.deepEqual(feedbackCenterState.resolveFeedbackListPage(2, 0, 12), { page: 1, shouldReload: true })
assert.deepEqual(feedbackCenterState.resolveFeedbackListPage(1, 0, 12), { page: 1, shouldReload: false })
const listRequestParameters = { keyword: 'network', status: 'pending', page: 2, pageSize: 12 }
assert.equal(
  feedbackCenterState.feedbackListRequestKey(listRequestParameters),
  feedbackCenterState.feedbackListRequestKey({ ...listRequestParameters }),
  'same list parameters must produce one request key',
)
assert.notEqual(
  feedbackCenterState.feedbackListRequestKey(listRequestParameters),
  feedbackCenterState.feedbackListRequestKey({ ...listRequestParameters, page: 1 }),
  'different pages must not share an in-flight request',
)

const requestDeduper = feedbackCenterState.createInFlightRequestDeduper()
let requestCount = 0
let resolveFirstRequest
const firstRequestResult = new Promise((resolve) => {
  resolveFirstRequest = resolve
})
const firstRequest = requestDeduper.run('list:1', () => {
  requestCount += 1
  return firstRequestResult
})
const duplicateRequest = requestDeduper.run('list:1', () => {
  requestCount += 1
  return Promise.resolve('duplicate')
})
assert.equal(firstRequest, duplicateRequest, 'same request key must return the existing promise')
assert.equal(requestCount, 1, 'same request key must only execute once while in flight')
resolveFirstRequest('first')
assert.equal(await duplicateRequest, 'first')
await requestDeduper.run('list:1', async () => {
  requestCount += 1
  return 'after-settle'
})
assert.equal(requestCount, 2, 'settled request keys must be reusable')

const visibilityDocument = {
  visibilityState: 'hidden',
  listeners: new Set(),
  addEventListener(event, listener) {
    assert.equal(event, 'visibilitychange')
    this.listeners.add(listener)
  },
  removeEventListener(event, listener) {
    assert.equal(event, 'visibilitychange')
    this.listeners.delete(listener)
  },
  dispatch() {
    for (const listener of this.listeners)
      listener()
  },
}
let feedbackActive = true
let visibilityRefreshCount = 0
const removeVisibilityListener = feedbackCenterState.registerFeedbackVisibilityRefresh(
  visibilityDocument,
  () => feedbackActive,
  () => { visibilityRefreshCount += 1 },
)
visibilityDocument.dispatch()
assert.equal(visibilityRefreshCount, 0, 'hidden document must not refresh feedback')
visibilityDocument.visibilityState = 'visible'
feedbackActive = false
visibilityDocument.dispatch()
assert.equal(visibilityRefreshCount, 0, 'inactive feedback page must not refresh')
feedbackActive = true
visibilityDocument.dispatch()
assert.equal(visibilityRefreshCount, 1, 'visible active feedback page must refresh once')
removeVisibilityListener()
visibilityDocument.dispatch()
assert.equal(visibilityRefreshCount, 1, 'removed listener must not refresh')

const createTab = feedbackCenterState.createFeedbackDynamicTab('feedback:create', 'New Feedback', 'plus-circle')
assert.deepEqual(createTab, {
  key: 'feedback:create',
  label: 'New Feedback',
  icon: 'plus-circle',
  path: '/pages/index/index?view=feedback%3Acreate',
})
assert.deepEqual(
  feedbackCenterState.createFeedbackDynamicTab('feedback:detail:42', 'FB-42', 'chat'),
  {
    key: 'feedback:detail:42',
    label: 'FB-42',
    icon: 'chat',
    path: '/pages/index/index?view=feedback%3Adetail%3A42',
  },
)
assert.equal(feedbackCenterState.createFeedbackDynamicTab('', 'Invalid', 'chat'), null)

for (const localeFile of ['src/locale/lang/zh-CN.json', 'src/locale/lang/en-US.json']) {
  const locale = JSON.parse(source(localeFile))
  assert.deepEqual(Object.keys(locale.feedback.statuses).sort(), ['closed', 'pending', 'processing', 'resolved'])
  for (const key of [
    'title',
    'description',
    'create',
    'keyword',
    'keywordPlaceholder',
    'status',
    'allStatuses',
    'search',
    'reset',
    'loading',
    'empty',
    'loadFailed',
    'overviewFailed',
    'retry',
    'feedbackNo',
    'summary',
    'images',
    'lastActivity',
    'view',
    'previousPage',
    'nextPage',
  ]) {
    assert.equal(typeof locale.feedback[key], 'string', `${localeFile} missing feedback.${key}`)
  }
}

console.log('user feedback contracts ok')
