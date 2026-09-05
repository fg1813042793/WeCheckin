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
  'src/pages/feedback/components/feedback-editor-state.ts',
  'src/pages/feedback/components/FeedbackImagePicker.vue',
  'src/pages/feedback/components/FeedbackCreatePage.vue',
  'src/pages/feedback/components/FeedbackDetailPage.vue',
  'src/pages/feedback/components/FeedbackTimeline.vue',
  'src/pages/feedback/components/FeedbackStatusOverview.vue',
  'src/pages/feedback/components/FeedbackList.vue',
  'src/pages/notifications/feedback-notification-open.ts',
  'src/pages/notifications/notification-mark-read.ts',
  'src/components/app-notification-panel/app-notification-panel.vue',
  'src/components/app-shell/app-shell-navigation-guard.ts',
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

function functionIdentifiers(sourceFile, functionName) {
  const identifiers = []
  const declaration = findFunction(sourceFile, functionName)
  function visit(node) {
    if (ts.isIdentifier(node))
      identifiers.push(node.text)
    ts.forEachChild(node, visit)
  }
  visit(declaration)
  return identifiers
}

function functionCallStringArguments(sourceFile, functionName, calleeName) {
  const values = []
  const declaration = findFunction(sourceFile, functionName)
  function visit(node) {
    if (ts.isCallExpression(node) && node.expression.getText(sourceFile) === calleeName) {
      const argument = node.arguments[0]
      if (argument && ts.isStringLiteral(argument))
        values.push(argument.text)
    }
    ts.forEachChild(node, visit)
  }
  visit(declaration)
  return values
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

async function loadNotificationPanelMetaModule() {
  const sourceFile = vueScriptSourceFile('src/components/app-notification-panel/app-notification-panel.vue')
  const variableNames = new Set([
    'notificationTypeMetas',
    'defaultNotificationTypeMeta',
    'notificationToneColors',
  ])
  const statements = sourceFile.statements.filter((statement) => {
    if (ts.isFunctionDeclaration(statement))
      return statement.name?.text === 'notificationTypeMeta'
    if (!ts.isVariableStatement(statement))
      return false
    return statement.declarationList.declarations.some(declaration => (
      ts.isIdentifier(declaration.name) && variableNames.has(declaration.name.text)
    ))
  })
  assert.equal(statements.length, 4, 'notification type meta declarations are incomplete')
  const moduleSource = `${statements.map(statement => statement.getText(sourceFile)).join('\n')}\nexport { notificationTypeMeta }`
  const output = ts.transpileModule(moduleSource, {
    compilerOptions: {
      module: ts.ModuleKind.ESNext,
      target: ts.ScriptTarget.ES2022,
    },
  })
  const moduleUrl = `data:text/javascript;base64,${Buffer.from(output.outputText).toString('base64')}`
  return import(moduleUrl)
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
  'getUserFeedbackDetail(id: string)',
  'supplementUserFeedback(id: string',
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
  types: { id: 'string', contentType: 'string', sizeBytes: 'number' },
})
assertInterfaceShape(feedbackApiFile, 'UserFeedbackMessage', {
  required: ['id', 'messageType', 'authorType', 'authorId', 'content', 'attachments', 'createdAt'],
  optional: ['authorName', 'fromStatus', 'toStatus'],
  types: {
    id: 'string',
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
  optional: ['submitterName', 'handlerId', 'handlerName', 'firstImageUrl', 'resolvedAt', 'closedAt'],
  types: { id: 'string', firstImageUrl: 'string', status: 'UserFeedbackStatus', version: 'number' },
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

const notificationMeta = await loadNotificationPanelMetaModule()
assert.deepEqual(notificationMeta.notificationTypeMeta('feedback_status'), {
  label: '用户反馈',
  icon: 'chat',
  color: '#2563eb',
  tone: 'primary',
})
assert.deepEqual(notificationMeta.notificationTypeMeta({
  type: 'feedback_status',
  style: { label: '反馈进度', icon: 'info-circle', tone: 'success' },
}), {
  label: '反馈进度',
  icon: 'info-circle',
  color: '#00875a',
  tone: 'success',
}, 'valid backend style must override feedback defaults')
assert.deepEqual(notificationMeta.notificationTypeMeta({
  type: 'feedback_status',
  style: { label: '', icon: '', tone: 'unsupported' },
}), {
  label: '用户反馈',
  icon: 'chat',
  color: '#2563eb',
  tone: 'primary',
}, 'invalid backend style values must fall back to feedback defaults')
assert.deepEqual(notificationMeta.notificationTypeMeta('unknown_notification'), {
  label: '系统消息',
  icon: 'email',
  color: '#475569',
  tone: 'info',
}, 'unknown notification types must keep the existing fallback')

const feedbackNotificationOpen = await loadTypeScriptModule('src/pages/notifications/feedback-notification-open.ts')
const notificationMarkRead = await loadTypeScriptModule('src/pages/notifications/notification-mark-read.ts')
const workflowMarkEvents = []
let resolveWorkflowMark
const workflowMarkResult = notificationMarkRead.runNotificationMarkRead({
  request() {
    workflowMarkEvents.push('request')
    return new Promise((resolve) => {
      resolveWorkflowMark = resolve
    })
  },
  commit() {
    workflowMarkEvents.push('commit')
  },
  fail() {
    workflowMarkEvents.push('fail')
  },
})
await Promise.resolve()
assert.deepEqual(workflowMarkEvents, ['request'], 'local workflow read state must wait for the API')
resolveWorkflowMark(false)
assert.equal(await workflowMarkResult, true)
assert.deepEqual(
  workflowMarkEvents,
  ['request', 'commit'],
  'workflow must preserve the legacy resolved-false behavior and commit local read state',
)

const strictFeedbackMarkEvents = []
assert.equal(await notificationMarkRead.runNotificationMarkRead({
  requireExplicitSuccess: true,
  async request() {
    strictFeedbackMarkEvents.push('request')
    return false
  },
  commit() {
    strictFeedbackMarkEvents.push('commit')
  },
  fail() {
    strictFeedbackMarkEvents.push('fail')
  },
}), false)
assert.deepEqual(
  strictFeedbackMarkEvents,
  ['request', 'fail'],
  'feedback must reject an ambiguous resolved-false mark and keep local state unchanged',
)

const rejectedWorkflowMarkEvents = []
assert.equal(await notificationMarkRead.runNotificationMarkRead({
  async request() {
    rejectedWorkflowMarkEvents.push('request')
    throw new Error('request failed')
  },
  commit() {
    rejectedWorkflowMarkEvents.push('commit')
  },
  fail() {
    rejectedWorkflowMarkEvents.push('fail')
  },
}), false)
assert.deepEqual(rejectedWorkflowMarkEvents, ['request', 'fail'], 'workflow API rejection must keep failing')

const openEvents = []
let openedFeedbackTab
let resolveMarkRead
const markReadGate = new Promise((resolve) => {
  resolveMarkRead = resolve
})
let notifyNavigationStarted
let resolveFeedbackNavigation
const feedbackNavigationStarted = new Promise((resolve) => {
  notifyNavigationStarted = resolve
})
const feedbackNavigationGate = new Promise((resolve) => {
  resolveFeedbackNavigation = resolve
})
const openingFeedback = feedbackNotificationOpen.openFeedbackNotification({
  sourceType: 'user_feedback',
  sourceId: '18446744073709551615',
  title: '反馈 FB-20260905 已解决',
}, {
  async markRead() {
    openEvents.push('mark:start')
    const marked = await markReadGate
    openEvents.push('mark:end')
    return marked
  },
  feedbackDetailContentKey(sourceID) {
    openEvents.push(`key:${sourceID}`)
    return routeKeys.feedbackDetailContentKey(sourceID)
  },
  openDynamicTab(tab) {
    openEvents.push('open:start')
    openedFeedbackTab = tab
    notifyNavigationStarted()
    return feedbackNavigationGate.then((opened) => {
      openEvents.push('open:end')
      return opened
    })
  },
  closePanel() {
    openEvents.push('close')
  },
  fallbackLabel: '反馈详情',
})
await Promise.resolve()
assert.deepEqual(openEvents, ['mark:start'], 'feedback deep link must wait for mark-read completion')
resolveMarkRead(true)
await feedbackNavigationStarted
assert.deepEqual(openEvents, [
  'mark:start',
  'mark:end',
  'key:18446744073709551615',
  'open:start',
], 'feedback deep link must wait for guarded navigation before closing the panel')
resolveFeedbackNavigation(true)
assert.equal(await openingFeedback, true)
assert.deepEqual(openEvents, [
  'mark:start',
  'mark:end',
  'key:18446744073709551615',
  'open:start',
  'open:end',
  'close',
])
assert.deepEqual(openedFeedbackTab, {
  key: 'feedback:detail:18446744073709551615',
  label: '反馈 FB-20260905 已解决',
  icon: 'chat',
  path: '/pages/index/index?view=feedback%3Adetail%3A18446744073709551615',
})

let fallbackFeedbackTab
assert.equal(await feedbackNotificationOpen.openFeedbackNotification({
  sourceType: 'user_feedback',
  sourceId: '42',
  title: '  ',
}, {
  async markRead() {
    return true
  },
  feedbackDetailContentKey: routeKeys.feedbackDetailContentKey,
  openDynamicTab(tab) {
    fallbackFeedbackTab = tab
    return true
  },
  closePanel() {},
  fallbackLabel: '反馈详情',
}), true)
assert.equal(fallbackFeedbackTab.label, '反馈详情')

const cancelledFeedbackNavigationEvents = []
assert.equal(await feedbackNotificationOpen.openFeedbackNotification({
  sourceType: 'user_feedback',
  sourceId: '42',
  title: '反馈详情',
}, {
  async markRead() {
    cancelledFeedbackNavigationEvents.push('mark')
    return true
  },
  feedbackDetailContentKey(sourceID) {
    cancelledFeedbackNavigationEvents.push(`key:${sourceID}`)
    return routeKeys.feedbackDetailContentKey(sourceID)
  },
  async openDynamicTab() {
    cancelledFeedbackNavigationEvents.push('open')
    return false
  },
  closePanel() {
    cancelledFeedbackNavigationEvents.push('close')
  },
  fallbackLabel: '反馈详情',
}), true)
assert.deepEqual(
  cancelledFeedbackNavigationEvents,
  ['mark', 'key:42', 'open'],
  'cancelled feedback navigation may stay read but must keep the notification panel open',
)

const failedMarkEvents = []
assert.equal(await feedbackNotificationOpen.openFeedbackNotification({
  sourceType: 'user_feedback',
  sourceId: '42',
  title: '',
}, {
  async markRead() {
    failedMarkEvents.push('mark')
    return false
  },
  feedbackDetailContentKey() {
    failedMarkEvents.push('key')
    return 'feedback:detail:42'
  },
  openDynamicTab() {
    failedMarkEvents.push('open')
  },
  closePanel() {
    failedMarkEvents.push('close')
  },
  fallbackLabel: '反馈详情',
}), true)
assert.deepEqual(failedMarkEvents, ['mark'], 'mark-read failure must not build, open, or close the feedback tab')

for (const invalidSourceID of ['', '0', '-1', '1.5', 'a', '18446744073709551616']) {
  const events = []
  assert.equal(await feedbackNotificationOpen.openFeedbackNotification({
    sourceType: 'user_feedback',
    sourceId: invalidSourceID,
    title: '',
  }, {
    async markRead() {
      events.push('mark')
      return true
    },
    feedbackDetailContentKey(sourceID) {
      events.push(`key:${sourceID}`)
      return routeKeys.feedbackDetailContentKey(sourceID)
    },
    openDynamicTab() {
      events.push('open')
    },
    closePanel() {
      events.push('close')
    },
    fallbackLabel: '反馈详情',
  }), true)
  assert.deepEqual(events, ['mark', `key:${invalidSourceID}`], `invalid source id must not navigate: ${invalidSourceID}`)
}

const workflowEvents = []
assert.equal(await feedbackNotificationOpen.openFeedbackNotification({
  sourceType: 'workflow_instance',
  sourceId: '91',
  title: '流程详情',
}, {
  async markRead() {
    workflowEvents.push('mark')
    return true
  },
  feedbackDetailContentKey() {
    workflowEvents.push('key')
    return ''
  },
  openDynamicTab() {
    workflowEvents.push('open')
  },
  closePanel() {
    workflowEvents.push('close')
  },
  fallbackLabel: '反馈详情',
}), false)
assert.deepEqual(workflowEvents, [], 'feedback helper must not consume workflow notifications')

const notificationPanelScript = vueScriptSourceFile('src/components/app-notification-panel/app-notification-panel.vue')
assert.ok(functionCalls(notificationPanelScript, 'handleFeedbackNotification').includes('openFeedbackNotification'))
assert.ok(functionCalls(notificationPanelScript, 'handleFeedbackNotification').includes('openNotificationTab'))
assert.ok(functionCalls(notificationPanelScript, 'markRead').includes('runNotificationMarkRead'))
assert.match(
  findFunction(notificationPanelScript, 'handleFeedbackNotification').getText(notificationPanelScript),
  /markRead\(notification,\s*\{\s*requireExplicitSuccess:\s*true\s*\}\)/,
  'only feedback notifications must request explicit mark-read success',
)
assert.doesNotMatch(
  findFunction(notificationPanelScript, 'openNotification').getText(notificationPanelScript),
  /requireExplicitSuccess/,
  'workflow notification opening must keep the default mark-read semantics',
)
assert.deepEqual(
  functionCalls(notificationPanelScript, 'openNotification')
    .filter(call => ['markRead', 'openWorkflowNotification'].includes(call)),
  ['markRead', 'openWorkflowNotification'],
  'workflow notification must keep mark-read dispatch before guarded navigation',
)
assert.deepEqual(
  functionCalls(notificationPanelScript, 'openNotificationTab')
    .filter(call => ['navigateWithUnsavedGuard', 'appContent.hasUnsavedTabChanges', 'confirmNavigationDiscard', 'appContent.openDynamicTab'].includes(call)),
  ['navigateWithUnsavedGuard', 'appContent.hasUnsavedTabChanges', 'confirmNavigationDiscard', 'appContent.openDynamicTab'],
  'notification tabs must use the shared unsaved-navigation guard',
)
for (const entry of ['openWorkflowNotification', 'openNotificationHistory']) {
  assert.ok(
    functionCalls(notificationPanelScript, entry).includes('openNotificationTabAndClose'),
    `${entry} must use guarded navigation and close only after success`,
  )
}
assert.deepEqual(
  functionCalls(notificationPanelScript, 'openNotificationTabAndClose')
    .filter(call => ['openNotificationTab', 'closePanel'].includes(call)),
  ['openNotificationTab', 'closePanel'],
  'workflow and history panels must close only after guarded navigation succeeds',
)
assert.deepEqual(
  functionCallStringArguments(notificationPanelScript, 'confirmNavigationDiscard', 't'),
  ['title', 'content', 'confirm', 'cancel'],
  'notification navigation confirmation must use the shared locale namespace',
)
assertContains('src/components/app-notification-panel/app-notification-panel.vue', [
  `notification.sourceType === 'user_feedback'`,
  'handleFeedbackNotification(notification)',
  'void markRead(notification)',
  `notification.sourceType === 'workflow_instance'`,
  `label: notification.title || '流程详情'`,
  `icon: 'file-text'`,
  'appContent.currentKey',
  'runNotificationMarkRead',
  'requireExplicitSuccess: true',
])

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
  'src/pages/feedback/components/FeedbackImagePicker.vue',
  'src/pages/feedback/components/FeedbackCreatePage.vue',
  'src/pages/feedback/components/FeedbackDetailPage.vue',
  'src/pages/feedback/components/FeedbackTimeline.vue',
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

const feedbackEditorState = await loadTypeScriptModule('src/pages/feedback/components/feedback-editor-state.ts')
const generatedRequestIds = ['request-first', 'request-next']
const requestIdState = feedbackEditorState.createStableFeedbackRequestIdState(() => generatedRequestIds.shift())
assert.equal(requestIdState.current(), 'request-first')
assert.equal(requestIdState.current(), 'request-first', 'request id must remain stable before success')
assert.equal(requestIdState.rotate(), 'request-next')
assert.equal(requestIdState.current(), 'request-next', 'request id must only change after rotate')
assert.match(feedbackEditorState.createFeedbackRequestId(), /^feedback-[a-z0-9-]{10,63}$/)

const validDraftImage = {
  clientId: 'image-valid',
  filePath: 'blob:valid',
  name: 'valid.jpg',
  mimeType: 'image/jpeg',
  size: 1024,
}
assert.equal(feedbackEditorState.validateFeedbackDraft('create', '', []), 'content_required')
assert.equal(feedbackEditorState.validateFeedbackDraft('create', 'x'.repeat(5000), []), '')
assert.equal(feedbackEditorState.validateFeedbackDraft('create', 'x'.repeat(5001), []), 'content_too_long')
assert.equal(feedbackEditorState.validateFeedbackDraft('supplement', '', []), 'content_or_image_required')
assert.equal(feedbackEditorState.validateFeedbackDraft('supplement', '', [validDraftImage]), '')
assert.equal(feedbackEditorState.validateFeedbackDraft('supplement', 'more', [
  { ...validDraftImage, error: 'too_large' },
]), 'image_invalid')
assert.equal(feedbackEditorState.validateFeedbackDraft('create', 'content', Array.from({ length: 7 }, (_, index) => ({
  ...validDraftImage,
  clientId: `image-${index}`,
}))), 'images_too_many')
assert.equal(feedbackEditorState.feedbackDraftHasContent('  ', []), false)
assert.equal(feedbackEditorState.feedbackDraftHasContent('  ', [validDraftImage]), true)

const chosenImages = feedbackEditorState.normalizeFeedbackChosenImages({
  tempFilePaths: ['blob:valid', 'blob:large', 'blob:text'],
  tempFiles: [
    { name: 'valid.jpg', size: 1024, type: 'image/jpeg' },
    { name: 'large.png', size: 10 * 1024 * 1024 + 1, type: 'image/png' },
    { name: 'text.txt', size: 10, type: 'text/plain' },
  ],
}, 'batch')
assert.equal(chosenImages.length, 3)
assert.equal(chosenImages[0].error, undefined)
assert.equal(chosenImages[1].error, 'too_large')
assert.equal(chosenImages[2].error, 'unsupported_type')
assert.deepEqual(feedbackEditorState.validFeedbackDraftImages(chosenImages).map(image => image.clientId), ['batch-1'])

assert.equal(feedbackEditorState.canSupplementFeedback('pending', true), true)
assert.equal(feedbackEditorState.canSupplementFeedback('processing', true), true)
assert.equal(feedbackEditorState.canSupplementFeedback('resolved', true), false)
assert.equal(feedbackEditorState.canSupplementFeedback('closed', true), false)
assert.equal(feedbackEditorState.canSupplementFeedback('pending', false), false)
assert.equal(feedbackEditorState.feedbackDetailIsInaccessible({ statusCode: 404 }), true)
assert.equal(feedbackEditorState.feedbackDetailIsInaccessible({ status: 403 }), true)
assert.equal(feedbackEditorState.feedbackDetailIsInaccessible({ response: { status: 404 } }), true)
assert.equal(feedbackEditorState.feedbackDetailIsInaccessible({
  data: { code: 43004, msg: '反馈不存在' },
}), true, 'uView originalData feedback-not-found envelope must be inaccessible')
assert.equal(feedbackEditorState.feedbackDetailIsInaccessible({
  data: { code: 0, msg: '反馈不存在' },
}), false, 'successful envelopes must not be treated as inaccessible')
assert.equal(feedbackEditorState.feedbackDetailIsInaccessible({
  data: { code: 43001, msg: '反馈状态不允许当前操作' },
}), false, 'ordinary business failures must not be treated as inaccessible')
assert.equal(feedbackEditorState.feedbackDetailIsInaccessible({ statusCode: 500 }), false)
assert.equal(feedbackEditorState.feedbackDetailIsInaccessible(new Error('network failed')), false)
assert.deepEqual(feedbackEditorState.feedbackImageSelectionCount(2, 6), {
  selectedCount: 2,
  maxCount: 6,
})
assert.deepEqual(feedbackEditorState.feedbackImageSelectionCount(1, 3), {
  selectedCount: 1,
  maxCount: 3,
}, 'image selection count must use the configured maxCount')

const sortedMessages = feedbackEditorState.sortFeedbackMessages([
  { id: '18446744073709551615', createdAt: 30 },
  { id: '2', createdAt: 10 },
  { id: '3', createdAt: 20 },
])
assert.deepEqual(sortedMessages.map(message => message.id), ['2', '3', '18446744073709551615'])

assertContains('src/pages/feedback/feedback.routes.ts', [
  `import FeedbackCreatePage from './components/FeedbackCreatePage.vue'`,
  `import FeedbackDetailPage from './components/FeedbackDetailPage.vue'`,
  'key === FEEDBACK_CREATE_CONTENT_KEY',
  'feedbackDetailIdFromContentKey(key)',
])

const feedbackCreateScript = vueScriptSourceFile('src/pages/feedback/components/FeedbackCreatePage.vue')
for (const call of ['validateFeedbackDraft', 'validFeedbackDraftImages', 'createUserFeedback'])
  assert.ok(functionCalls(feedbackCreateScript, 'submitFeedback').includes(call), `submitFeedback must call ${call}`)
for (const call of ['requestIdState.rotate', 'unregisterCloseGuard', 'appContent.requestRefresh', 'appContent.removeDynamicTab', 'feedbackDetailContentKey', 'appContent.openDynamicTab'])
  assert.ok(functionCalls(feedbackCreateScript, 'handleCreateSuccess').includes(call), `create success must call ${call}`)
assert.ok(lifecycleCallbackCalls(feedbackCreateScript, 'onMounted').includes('registerCloseGuard'), 'create page must register close guard')
assert.ok(lifecycleCallbackCalls(feedbackCreateScript, 'onBeforeUnmount').includes('unregisterCloseGuard'), 'create page must unregister close guard')

const feedbackDetailScript = vueScriptSourceFile('src/pages/feedback/components/FeedbackDetailPage.vue')
for (const call of ['feedbackDetailIdFromContentKey', 'getUserFeedbackDetail', 'feedbackDetailIsInaccessible'])
  assert.ok(functionCalls(feedbackDetailScript, 'loadDetail').includes(call), `loadDetail must call ${call}`)
for (const call of ['validateFeedbackDraft', 'validFeedbackDraftImages', 'supplementUserFeedback'])
  assert.ok(functionCalls(feedbackDetailScript, 'submitSupplement').includes(call), `submitSupplement must call ${call}`)
for (const call of ['requestIdState.rotate', 'unregisterCloseGuard', 'registerCloseGuard', 'appContent.requestRefresh'])
  assert.ok(functionCalls(feedbackDetailScript, 'handleSupplementSuccess').includes(call), `supplement success must call ${call}`)
assert.equal(functionCalls(feedbackDetailScript, 'loadDetail').includes('Number'), false, 'detail id must remain a decimal string')
assert.equal(
  functionIdentifiers(feedbackDetailScript, 'hasUnsavedChanges').includes('canSupplement'),
  false,
  'existing supplement drafts must remain guarded after the feedback becomes read-only',
)
assert.ok(lifecycleCallbackCalls(feedbackDetailScript, 'onMounted').includes('registerCloseGuard'), 'detail page must register close guard')
assert.ok(lifecycleCallbackCalls(feedbackDetailScript, 'onBeforeUnmount').includes('unregisterCloseGuard'), 'detail page must unregister close guard')
assertContains('src/pages/feedback/components/FeedbackDetailPage.vue', [
  'detail.lastActivityAt || detail.updatedAt',
  `t('detailPage.lastUpdated')`,
])

const feedbackImagePickerScript = vueScriptSourceFile('src/pages/feedback/components/FeedbackImagePicker.vue')
assert.ok(functionCalls(feedbackImagePickerScript, 'chooseImages').includes('uni.chooseImage'))
assert.ok(functionCalls(feedbackImagePickerScript, 'previewImage').includes('uni.previewImage'))
assert.ok(functionCalls(feedbackImagePickerScript, 'removeImage').includes('emit'))
assert.ok(
  feedbackImagePickerScript.statements.some(statement => statement.getText(feedbackImagePickerScript).includes('feedbackImageSelectionCount')),
  'image picker must derive its visible count from executable selection-count logic',
)

const navigationGuard = await loadTypeScriptModule('src/components/app-shell/app-shell-navigation-guard.ts')
const originalUni = globalThis.uni
const zhNavigationCopy = JSON.parse(source('src/locale/lang/zh-CN.json')).appShell?.unsavedNavigation
let receivedNavigationModal
globalThis.uni = {
  showModal(options) {
    receivedNavigationModal = options
    options.success({ confirm: false })
  },
}
try {
  assert.equal(await navigationGuard.confirmUnsavedNavigation(zhNavigationCopy), false)
  assert.deepEqual({
    title: receivedNavigationModal.title,
    content: receivedNavigationModal.content,
    confirmText: receivedNavigationModal.confirmText,
    cancelText: receivedNavigationModal.cancelText,
  }, {
    title: zhNavigationCopy.title,
    content: zhNavigationCopy.content,
    confirmText: zhNavigationCopy.confirm,
    cancelText: zhNavigationCopy.cancel,
  }, 'shared confirmation must pass all locale copy to uni.showModal')
}
finally {
  if (originalUni === undefined)
    delete globalThis.uni
  else
    globalThis.uni = originalUni
}

const cancelledNavigationEvents = []
let resolveCancelledNavigation
const cancelledNavigation = navigationGuard.navigateWithUnsavedGuard({
  activeKey: 'feedback:detail:42',
  targetKey: 'dashboard',
  hasUnsavedChanges(key) {
    cancelledNavigationEvents.push(`guard:${key}`)
    return true
  },
  confirmLeave() {
    cancelledNavigationEvents.push('confirm')
    return new Promise((resolve) => {
      resolveCancelledNavigation = resolve
    })
  },
  navigate() {
    cancelledNavigationEvents.push('navigate')
  },
})
await Promise.resolve()
assert.deepEqual(cancelledNavigationEvents, ['guard:feedback:detail:42', 'confirm'])
resolveCancelledNavigation(false)
assert.equal(await cancelledNavigation, false)
assert.deepEqual(cancelledNavigationEvents, ['guard:feedback:detail:42', 'confirm'], 'cancel must not navigate')

const confirmedNavigationEvents = []
assert.equal(await navigationGuard.navigateWithUnsavedGuard({
  activeKey: 'feedback:detail:42',
  targetKey: 'workflow',
  hasUnsavedChanges(key) {
    confirmedNavigationEvents.push(`guard:${key}`)
    return true
  },
  async confirmLeave() {
    confirmedNavigationEvents.push('confirm')
    return true
  },
  navigate() {
    confirmedNavigationEvents.push('navigate')
  },
}), true)
assert.deepEqual(
  confirmedNavigationEvents,
  ['guard:feedback:detail:42', 'confirm', 'navigate'],
  'navigation must happen only after confirmation succeeds',
)

const unguardedNavigationEvents = []
assert.equal(await navigationGuard.navigateWithUnsavedGuard({
  activeKey: 'feedback:detail:42',
  targetKey: 'dashboard',
  hasUnsavedChanges(key) {
    unguardedNavigationEvents.push(`guard:${key}`)
    return false
  },
  async confirmLeave() {
    unguardedNavigationEvents.push('confirm')
    return false
  },
  navigate() {
    unguardedNavigationEvents.push('navigate')
  },
}), true)
assert.deepEqual(unguardedNavigationEvents, ['guard:feedback:detail:42', 'navigate'])

const sameTabNavigationEvents = []
assert.equal(await navigationGuard.navigateWithUnsavedGuard({
  activeKey: 'feedback:detail:42',
  targetKey: 'feedback:detail:42',
  hasUnsavedChanges() {
    sameTabNavigationEvents.push('guard')
    return true
  },
  async confirmLeave() {
    sameTabNavigationEvents.push('confirm')
    return false
  },
  navigate() {
    sameTabNavigationEvents.push('navigate')
  },
}), true)
assert.deepEqual(sameTabNavigationEvents, ['navigate'], 'same-tab focus must not prompt')

const closeNavigationEvents = []
assert.equal(await navigationGuard.navigateWithUnsavedGuard({
  activeKey: 'feedback:detail:42',
  targetKey: 'dashboard',
  guardUnsavedChanges: false,
  hasUnsavedChanges() {
    closeNavigationEvents.push('guard')
    return true
  },
  async confirmLeave() {
    closeNavigationEvents.push('confirm')
    return false
  },
  navigate() {
    closeNavigationEvents.push('navigate')
  },
}), true)
assert.deepEqual(closeNavigationEvents, ['navigate'], 'the existing close-tab decision must not prompt twice')

const appShellScript = vueScriptSourceFile('src/components/app-shell/app-shell.vue')
for (const entry of ['navigateByKey', 'handleNavItemClick', 'handleTopNavItemClick', 'completeTabClose']) {
  assert.ok(functionCalls(appShellScript, entry).includes('navigateToItem'), `${entry} must keep routing through navigateToItem`)
}
for (const call of ['navigateWithUnsavedGuard', 'appContent.hasUnsavedTabChanges', 'confirmNavigationDiscard']) {
  assert.ok(functionCalls(appShellScript, 'navigateToItem').includes(call), `navigateToItem must call ${call}`)
}
assert.ok(
  functionCalls(appShellScript, 'confirmNavigationDiscard').includes('confirmUnsavedNavigation'),
  'app shell must use the shared confirmation adapter',
)
assert.deepEqual(
  functionCallStringArguments(appShellScript, 'confirmNavigationDiscard', 't'),
  ['title', 'content', 'confirm', 'cancel'],
  'app shell navigation confirmation must use locale for all visible copy',
)
assert.doesNotMatch(
  findFunction(appShellScript, 'confirmNavigationDiscard').getText(appShellScript),
  /[\u3400-\u9FFF]/,
  'app shell navigation confirmation must not hardcode Chinese copy',
)
assert.match(
  findFunction(appShellScript, 'completeTabClose').getText(appShellScript),
  /navigateToItem\(nextItem,\s*\{\s*guardUnsavedChanges:\s*false\s*\}\)/,
  'closing a tab must preserve the existing popup decision without a second navigation prompt',
)

const feedbackTimelineScript = vueScriptSourceFile('src/pages/feedback/components/FeedbackTimeline.vue')
assert.ok(feedbackTimelineScript.statements.some(statement => statement.getText(feedbackTimelineScript).includes('sortFeedbackMessages')))
assert.ok(functionCalls(feedbackTimelineScript, 'previewAttachments').includes('uni.previewImage'))

const indexScript = vueScriptSourceFile('src/pages/index/index.vue')
assert.ok(functionCalls(indexScript, 'routeViewKey').includes('normalizeFeedbackDynamicContentKey'), 'URL view must normalize feedback dynamic keys')
assert.ok(functionCalls(indexScript, 'openFeedbackRouteTab').includes('normalizeFeedbackDynamicContentKey'))
assert.ok(functionCalls(indexScript, 'openFeedbackRouteTab').includes('appContent.openDynamicTab'))
assert.equal(functionCalls(indexScript, 'openFeedbackRouteTab').includes('Number'), false, 'feedback URL ids must remain strings')
assert.ok(functionCalls(indexScript, 'applyRouteQuery').includes('openFeedbackRouteTab'))

for (const localeFile of ['src/locale/lang/zh-CN.json', 'src/locale/lang/en-US.json']) {
  const locale = JSON.parse(source(localeFile))
  assert.deepEqual(
    Object.keys(locale.appShell?.unsavedNavigation || {}).sort(),
    ['cancel', 'confirm', 'content', 'title'],
    `${localeFile} missing appShell.unsavedNavigation copy`,
  )
  for (const value of Object.values(locale.appShell.unsavedNavigation))
    assert.equal(typeof value, 'string', `${localeFile} appShell.unsavedNavigation copy must be a string`)
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
  for (const namespace of ['createPage', 'detailPage', 'timeline', 'imagePicker'])
    assert.equal(typeof locale.feedback[namespace], 'object', `${localeFile} missing feedback.${namespace}`)
  for (const key of ['title', 'description', 'contentLabel', 'contentPlaceholder', 'submit', 'cancel', 'success', 'failed', 'contentRequired', 'contentTooLong', 'imagesTooMany', 'imageInvalid'])
    assert.equal(typeof locale.feedback.createPage[key], 'string', `${localeFile} missing feedback.createPage.${key}`)
  for (const key of ['title', 'unavailable', 'loadFailed', 'retry', 'supplementTitle', 'supplementPlaceholder', 'submitSupplement', 'supplementSuccess', 'supplementFailed', 'readOnly', 'handler', 'createdAt', 'lastUpdated'])
    assert.equal(typeof locale.feedback.detailPage[key], 'string', `${localeFile} missing feedback.detailPage.${key}`)
  for (const key of ['initial', 'supplement', 'status', 'statusChange', 'admin', 'user', 'empty'])
    assert.equal(typeof locale.feedback.timeline[key], 'string', `${localeFile} missing feedback.timeline.${key}`)
  for (const key of ['add', 'preview', 'remove', 'chooseFailed', 'tooLarge', 'unsupportedType', 'selectedCount', 'limitHint'])
    assert.equal(typeof locale.feedback.imagePicker[key], 'string', `${localeFile} missing feedback.imagePicker.${key}`)
}

console.log('user feedback contracts ok')
