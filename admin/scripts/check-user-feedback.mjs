import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { compileTemplate, parse as parseSFC } from '@vue/compiler-sfc'
import ts from 'typescript'
import { executeTypeScriptModule } from './lib/typescript-runtime.mjs'

const currentDir = dirname(fileURLToPath(import.meta.url))
const adminDir = resolve(currentDir, '..')
const paths = {
  types: resolve(adminDir, 'src/types/userFeedback.ts'),
  api: resolve(adminDir, 'src/api/index.ts'),
  routes: resolve(adminDir, 'src/router/adminRoutes.ts'),
  routeMeta: resolve(adminDir, 'src/router/meta.d.ts'),
  routeAccess: resolve(adminDir, 'src/router/adminAccess.ts'),
  status: resolve(adminDir, 'src/views/user-feedback/userFeedbackStatus.ts'),
  view: resolve(adminDir, 'src/views/user-feedback/index.vue'),
  overview: resolve(adminDir, 'src/views/user-feedback/components/UserFeedbackOverview.vue'),
  detailDrawer: resolve(adminDir, 'src/views/user-feedback/components/UserFeedbackDetailDrawer.vue'),
  timeline: resolve(adminDir, 'src/views/user-feedback/components/UserFeedbackTimeline.vue'),
  statusDialog: resolve(adminDir, 'src/views/user-feedback/components/UserFeedbackStatusDialog.vue'),
  viteConfig: resolve(adminDir, 'vite.config.ts'),
  package: resolve(adminDir, 'package.json'),
}

for (const [name, path] of Object.entries(paths)) {
  if (!existsSync(path)) throw new Error(`user feedback ${name} contract missing: ${path}`)
}

const sources = Object.fromEntries(
  Object.entries(paths).map(([name, path]) => [name, readFileSync(path, 'utf8')]),
)

function fail(message) {
  throw new Error(`user feedback contract failed: ${message}`)
}

function sourceFile(name) {
  return ts.createSourceFile(paths[name], sources[name], ts.ScriptTarget.Latest, true)
}

function componentContract(name) {
  const { descriptor, errors } = parseSFC(sources[name], { filename: paths[name] })
  assert.deepEqual(errors, [], `${name} must be a valid Vue SFC`)
  assert.ok(descriptor.template, `${name} must contain a template`)
  assert.ok(descriptor.scriptSetup, `${name} must use script setup`)
  const compiledTemplate = compileTemplate({
    source: descriptor.template.content,
    filename: paths[name],
    id: `user-feedback-${name}`,
  })
  assert.deepEqual(compiledTemplate.errors, [], `${name} template must compile`)
  const script = [descriptor.script?.content || '', descriptor.scriptSetup.content].join('\n')
  return {
    descriptor,
    template: descriptor.template.ast,
    script,
    scriptFile: ts.createSourceFile(`${paths[name]}.ts`, script, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS),
  }
}

function visit(node, callback) {
  callback(node)
  for (const child of node.children || []) visit(child, callback)
}

function elements(template, tag) {
  const found = []
  visit(template, (node) => {
    if (node.type === 1 && node.tag === tag) found.push(node)
  })
  return found
}

function templateProp(element, name, directiveName = '') {
  return element.props.find((prop) => {
    if (prop.type === 6) return !directiveName && prop.name === name
    if (prop.type !== 7 || prop.name !== directiveName) return false
    if (!prop.arg) return name === directiveName
    return prop.arg?.type === 4 && prop.arg.content === name
  })
}

function attributeValue(element, name) {
  const prop = templateProp(element, name)
  return prop?.type === 6 ? prop.value?.content ?? true : undefined
}

function directiveExpression(element, name, directiveName) {
  const prop = templateProp(element, name, directiveName)
  return prop?.type === 7 ? prop.exp?.content || '' : ''
}

function functionNode(file, name) {
  let found
  const find = (node) => {
    if (ts.isFunctionDeclaration(node) && node.name?.text === name) found = node
    if (ts.isVariableDeclaration(node) && ts.isIdentifier(node.name) && node.name.text === name
      && node.initializer && (ts.isArrowFunction(node.initializer) || ts.isFunctionExpression(node.initializer))) {
      found = node.initializer
    }
    ts.forEachChild(node, find)
  }
  find(file)
  assert.ok(found, `missing function ${name}`)
  return found
}

function callsIn(node, file) {
  const calls = []
  const find = (child) => {
    if (ts.isCallExpression(child)) calls.push(child.expression.getText(file))
    ts.forEachChild(child, find)
  }
  find(node)
  return calls
}

function assertCalls(file, functionName, expected) {
  const calls = callsIn(functionNode(file, functionName), file)
  for (const callee of expected) {
    assert.ok(calls.includes(callee), `${functionName} must call ${callee}; got ${calls.join(', ')}`)
  }
}

function firstStatement(file, functionName) {
  const fn = functionNode(file, functionName)
  assert.ok(fn.body && ts.isBlock(fn.body), `${functionName} must use a block body`)
  assert.ok(fn.body.statements.length > 0, `${functionName} must not be empty`)
  return fn.body.statements[0]
}

function catchClause(file, functionName) {
  const fn = functionNode(file, functionName)
  let found
  const find = (node) => {
    if (ts.isCatchClause(node)) found = node
    ts.forEachChild(node, find)
  }
  find(fn)
  assert.ok(found, `${functionName} must handle request failures`)
  return found
}

function guardedRequestPhases(file, functionName, guardName) {
  const fn = functionNode(file, functionName)
  let requestTry
  const find = (node) => {
    if (!requestTry && ts.isTryStatement(node)) requestTry = node
    ts.forEachChild(node, find)
  }
  find(fn)
  assert.ok(requestTry?.catchClause, `${functionName} must contain a request catch clause`)
  assert.ok(requestTry?.finallyBlock, `${functionName} must contain a request finally block`)

  const phaseStatements = {
    success: requestTry.tryBlock.statements,
    catch: requestTry.catchClause.block.statements,
    finally: requestTry.finallyBlock.statements,
  }
  assert.equal(phaseStatements.success.length, 2, `${functionName} success must only declare the response and apply guarded effects`)
  assert.ok(ts.isVariableStatement(phaseStatements.success[0]), `${functionName} success must first await its response`)
  assert.equal(phaseStatements.catch.length, 1, `${functionName} catch effects must use one guard call`)
  assert.equal(phaseStatements.finally.length, 1, `${functionName} finally effects must use one guard call`)

  const guardedEffects = {}
  for (const [phase, statement] of Object.entries({
    success: phaseStatements.success[1],
    catch: phaseStatements.catch[0],
    finally: phaseStatements.finally[0],
  })) {
    assert.ok(ts.isExpressionStatement(statement) && ts.isCallExpression(statement.expression), `${functionName} ${phase} must call ${guardName}`)
    const call = statement.expression
    assert.equal(call.expression.getText(file), guardName, `${functionName} ${phase} must use ${guardName}`)
    const callback = call.arguments.at(-1)
    assert.ok(callback && (ts.isArrowFunction(callback) || ts.isFunctionExpression(callback)), `${functionName} ${phase} guard must own its side effects`)
    guardedEffects[phase] = callback.body.getText(file)
  }
  return guardedEffects
}

function propertyName(node, file) {
  if (ts.isIdentifier(node) || ts.isStringLiteralLike(node)) return node.text
  return node.getText(file)
}

function property(object, name, file) {
  return object.properties.find(item => ts.isPropertyAssignment(item) && propertyName(item.name, file) === name)
}

function stringValue(node) {
  return node && ts.isStringLiteralLike(node) ? node.text : undefined
}

function numberValue(node) {
  return node && ts.isNumericLiteral(node) ? Number(node.text) : undefined
}

function interfaceContract(name) {
  const file = sourceFile('types')
  let declaration
  file.forEachChild((node) => {
    if (ts.isInterfaceDeclaration(node) && node.name.text === name) declaration = node
  })
  if (!declaration) fail(`types missing interface ${name}`)
  return {
    extends: declaration.heritageClauses?.flatMap(clause => clause.types.map(type => type.expression.getText(file))) || [],
    fields: Object.fromEntries(declaration.members.map((member) => {
      if (!ts.isPropertySignature(member) || !member.type) fail(`${name} contains unsupported member`)
      return [propertyName(member.name, file), {
        optional: Boolean(member.questionToken),
        type: member.type.getText(file),
      }]
    })),
  }
}

function assertInterface(name, expectedFields, expectedExtends = []) {
  assert.deepEqual(interfaceContract(name), { extends: expectedExtends, fields: expectedFields }, `${name} must match Backend DTO`)
}

function assertStringUnion(name, expected) {
  const file = sourceFile('types')
  let declaration
  file.forEachChild((node) => {
    if (ts.isTypeAliasDeclaration(node) && node.name.text === name) declaration = node
  })
  if (!declaration || !ts.isUnionTypeNode(declaration.type)) fail(`types missing union ${name}`)
  const actual = declaration.type.types.map((type) => {
    if (!ts.isLiteralTypeNode(type) || !ts.isStringLiteral(type.literal)) fail(`${name} must contain only string literals`)
    return type.literal.text
  })
  assert.deepEqual(actual, expected, `${name} values must match Backend enum`)
}

const required = type => ({ optional: false, type })
const optional = type => ({ optional: true, type })

assertStringUnion('UserFeedbackStatus', ['pending', 'processing', 'resolved', 'closed'])
assertStringUnion('UserFeedbackMessageType', ['initial', 'supplement', 'status'])
assertStringUnion('UserFeedbackAuthorType', ['user', 'admin'])
assertInterface('UserFeedbackOverview', {
  pending: required('number'), processing: required('number'), resolved: required('number'), closed: required('number'),
})
assertInterface('UserFeedbackAttachment', {
  id: required('string'), originalName: required('string'), contentType: optional('string'), sizeBytes: required('number'),
  url: required('string'), sortOrder: required('number'), createdAt: required('number'),
})
assertInterface('UserFeedbackMessage', {
  id: required('string'), messageType: required('UserFeedbackMessageType'), authorType: required('UserFeedbackAuthorType'),
  authorId: required('number'), authorName: optional('string'), content: required('string'),
  fromStatus: optional('UserFeedbackStatus'), toStatus: optional('UserFeedbackStatus'),
  attachments: required('UserFeedbackAttachment[]'), createdAt: required('number'),
})
assertInterface('UserFeedbackSummary', {
  id: required('string'), feedbackNo: required('string'), submitterId: required('number'), submitterName: optional('string'),
  summary: required('string'), imageCount: required('number'), firstImageUrl: optional('string'), status: required('UserFeedbackStatus'),
  handlerId: optional('number'), handlerName: optional('string'), version: required('number'),
  lastActivityAt: required('number'), resolvedAt: optional('number'), closedAt: optional('number'),
  createdAt: required('number'), updatedAt: required('number'),
})
assertInterface('UserFeedbackDetail', {
  messages: required('UserFeedbackMessage[]'), allowsSupplement: required('boolean'),
}, ['UserFeedbackSummary'])
assertInterface('UserFeedbackList', {
  list: required('UserFeedbackSummary[]'), total: required('number'), page: required('number'), pageSize: required('number'),
})
assertInterface('AdminUserFeedbackListQuery', {
  keyword: optional('string'), submitterId: optional('number'), handlerId: optional('number'), status: optional('UserFeedbackStatus'),
  submittedFrom: optional('number'), submittedTo: optional('number'), page: optional('number'), pageSize: optional('number'),
})
assertInterface('UpdateUserFeedbackStatusInput', {
  status: required('UserFeedbackStatus'), note: required('string'), notifyUser: required('boolean'),
  version: required('number'), requestId: required('string'),
})

function assertFeedbackRoute() {
  const file = sourceFile('routes')
  const matches = []
  const visit = (node) => {
    if (ts.isObjectLiteralExpression(node) && stringValue(property(node, 'path', file)?.initializer) === 'user-feedbacks') {
      matches.push(node)
    }
    ts.forEachChild(node, visit)
  }
  visit(file)
  assert.equal(matches.length, 1, 'exactly one user-feedbacks route object must exist')

  const route = matches[0]
  const component = property(route, 'component', file)?.initializer
  assert.ok(component && ts.isArrowFunction(component), 'feedback route component must be lazy')
  assert.ok(ts.isCallExpression(component.body), 'feedback route component must call import()')
  assert.equal(component.body.expression.kind, ts.SyntaxKind.ImportKeyword, 'feedback route must use dynamic import')
  assert.equal(stringValue(component.body.arguments[0]), '../views/user-feedback/index.vue')

  const meta = property(route, 'meta', file)?.initializer
  assert.ok(meta && ts.isObjectLiteralExpression(meta), 'feedback route meta must be an object')
  assert.equal(stringValue(property(meta, 'menuPath', file)?.initializer), '/user-feedbacks')
  assert.equal(stringValue(property(meta, 'requiredPermission', file)?.initializer), 'admin:menu:user-feedback:list')
  const adminUi = property(meta, 'adminUi', file)?.initializer
  assert.ok(adminUi && ts.isObjectLiteralExpression(adminUi), 'feedback route adminUi must be an object')
  assert.equal(numberValue(property(adminUi, 'version', file)?.initializer), 1)
  assert.equal(stringValue(property(adminUi, 'pattern', file)?.initializer), 'filter-list')
}

assertFeedbackRoute()

function assertRouteMetaType() {
  const file = sourceFile('routeMeta')
  let routeMeta
  const visit = (node) => {
    if (ts.isInterfaceDeclaration(node) && node.name.text === 'RouteMeta') routeMeta = node
    ts.forEachChild(node, visit)
  }
  visit(file)
  assert.ok(routeMeta, 'RouteMeta declaration must exist')
  const permission = routeMeta.members.find(member => ts.isPropertySignature(member) && propertyName(member.name, file) === 'requiredPermission')
  assert.equal(permission?.type?.getText(file), 'string')
  assert.ok(permission?.questionToken, 'requiredPermission must remain optional for existing routes')
}

assertRouteMetaType()

function assertAdminApiRequests() {
  const calls = []
  const request = {
    get: (...args) => calls.push({ verb: 'get', args }),
    patch: (...args) => calls.push({ verb: 'patch', args }),
  }
  const apiModule = executeTypeScriptModule(sources.api, paths.api, {
    '../utils/request': { __esModule: true, default: request },
  })
  const api = apiModule.adminApi
  const query = { status: 'processing', submitterId: 12, handlerId: 21, submittedFrom: 1000, submittedTo: 2000, page: 2, pageSize: 30 }
  const update = { status: 'resolved', note: '已处理', notifyUser: true, version: 3, requestId: 'request-1' }

  api.userFeedbackOverview(query)
  api.userFeedbackList(query)
  api.userFeedbackDetail('feedback/42')
  api.userFeedbackUpdateStatus(42, update)

  assert.equal(calls.length, 4, 'feedback API must issue exactly one request per method')
  assert.deepEqual(calls.map(call => call.verb), ['get', 'get', 'get', 'patch'])
  assert.equal(calls[0].args[0], '/api/v2/admin/user-feedbacks/overview')
  assert.equal(calls[0].args[1]?.params, query, 'overview query must be passed through params')
  assert.deepEqual(Object.keys(calls[0].args[1]), ['params'])
  assert.equal(calls[1].args[0], '/api/v2/admin/user-feedbacks')
  assert.equal(calls[1].args[1]?.params, query, 'list query must be passed through params')
  assert.deepEqual(Object.keys(calls[1].args[1]), ['params'])
  assert.equal(calls[2].args[0], '/api/v2/admin/user-feedbacks/feedback%2F42')
  assert.equal(calls[2].args.length, 1, 'detail request must not send an unrelated body or config')
  assert.equal(calls[3].args[0], '/api/v2/admin/user-feedbacks/42/status')
  assert.equal(calls[3].args[1], update, 'status payload must be the PATCH body')
  const jsonConfig = calls[3].args[2]
  assert.equal(jsonConfig?.headers?.['Content-Type'], 'application/json', 'status PATCH must use jsonConfig')
  assert.equal(jsonConfig?.transformRequest?.length, 1, 'jsonConfig must retain its serializer')
  assert.equal(jsonConfig.transformRequest[0](update), JSON.stringify(update))
}

assertAdminApiRequests()

function assertRouteAccessRuntime() {
  const accessModule = executeTypeScriptModule(sources.routeAccess, paths.routeAccess, {
    '../api': { adminApi: {} },
    '../utils/permission': { setPerms: () => {} },
  })
  const canAccess = accessModule.canAccessAdminRoute
  const menuPath = '/user-feedbacks'
  const permission = 'admin:menu:user-feedback:list'
  const meta = { menuPath, requiredPermission: permission }

  assert.equal(canAccess(meta, { menuPaths: new Set([menuPath]), permissions: [] }), false, 'menu without required permission must be denied')
  assert.equal(canAccess(meta, { menuPaths: new Set([menuPath]), permissions: [permission] }), true, 'menu and required permission must be allowed')
  assert.equal(canAccess(meta, { menuPaths: new Set(), permissions: [permission] }), false, 'permission without menu must be denied')
  assert.equal(canAccess({ allowWithoutMenu: true, requiredPermission: permission }, { menuPaths: new Set(), permissions: [] }), true, 'allowWithoutMenu must keep its existing bypass semantics')
}

assertRouteAccessRuntime()

function assertStatusRuntime() {
  let activePermissions = new Set()
  const statusModule = executeTypeScriptModule(sources.status, paths.status, {
    '@/utils/permission': { hasPerm: permission => activePermissions.has(permission) },
  })

  assert.equal(statusModule.USER_FEEDBACK_LIST_PERMISSION, 'admin:menu:user-feedback:list')
  assert.equal(statusModule.USER_FEEDBACK_HANDLE_PERMISSION, 'admin:menu:user-feedback:handle')
  assert.deepEqual({ ...statusModule.userFeedbackStatusMeta('pending') }, { label: '待处理', type: 'warning' })
  assert.deepEqual({ ...statusModule.userFeedbackStatusMeta('processing') }, { label: '处理中', type: 'warning' })
  assert.deepEqual({ ...statusModule.userFeedbackStatusMeta('resolved') }, { label: '已解决', type: 'success' })
  assert.deepEqual({ ...statusModule.userFeedbackStatusMeta('closed') }, { label: '已关闭', type: 'info' })
  assert.deepEqual([...statusModule.nextUserFeedbackStatuses('pending')], ['processing', 'closed'])
  assert.deepEqual([...statusModule.nextUserFeedbackStatuses('processing')], ['resolved'])
  assert.deepEqual([...statusModule.nextUserFeedbackStatuses('resolved')], ['closed', 'processing'])
  assert.deepEqual([...statusModule.nextUserFeedbackStatuses('closed')], ['processing'])

  assert.equal(statusModule.canHandleUserFeedback(), false, 'empty permission set must not handle feedback')
  activePermissions = new Set(['admin:menu:user-feedback:list'])
  assert.equal(statusModule.canHandleUserFeedback(), false, 'list permission must not grant handling')
  activePermissions = new Set(['admin:menu:user-feedback:handle'])
  assert.equal(statusModule.canHandleUserFeedback(), true, 'handle permission must grant handling')
}

assertStatusRuntime()

const view = componentContract('view')
const overview = componentContract('overview')
const detailDrawer = componentContract('detailDrawer')
const timeline = componentContract('timeline')
const statusDialog = componentContract('statusDialog')

function assertAdminUIImports() {
  const imports = new Map()
  for (const statement of view.scriptFile.statements) {
    if (!ts.isImportDeclaration(statement) || stringValue(statement.moduleSpecifier) !== '@/components/admin-ui') continue
    for (const item of statement.importClause?.namedBindings?.elements || []) imports.set(item.name.text, true)
  }
  for (const name of ['AdminPageShell', 'AdminPageHeader', 'AdminSearchBar', 'AdminTablePanel']) {
    assert.ok(imports.has(name), `view must import ${name} from the Admin UI component index`)
    assert.equal(elements(view.template, name).length, 1, `view must render exactly one ${name}`)
  }
}

assertAdminUIImports()

function assertPageTemplate() {
  assert.equal(elements(view.template, 'UserFeedbackOverview').length, 1)
  assert.equal(elements(view.template, 'UserFeedbackDetailDrawer').length, 1)
  assert.equal(elements(view.template, 'UserFeedbackStatusDialog').length, 1)

  const filterLabels = elements(view.template, 'el-form-item').map(item => attributeValue(item, 'label'))
  for (const label of ['编号 / 用户 / 消息', '状态', '处理人 ID', '提交时间']) {
    assert.ok(filterLabels.includes(label), `view filters must include ${label}`)
  }

  const columnLabels = elements(view.template, 'el-table-column').map(item => attributeValue(item, 'label'))
  for (const label of ['反馈编号', '内容摘要 / 首图', '提交人', '状态', '处理人', '图片数', '最近活动', '提交时间', '操作']) {
    assert.ok(columnLabels.includes(label), `feedback table must include ${label}`)
  }

  const panel = elements(view.template, 'AdminTablePanel')[0]
  for (const prop of ['loading', 'empty', 'error']) {
    assert.ok(directiveExpression(panel, prop, 'bind'), `table panel must bind ${prop}`)
  }
  const pagination = elements(view.template, 'el-pagination')[0]
  assert.ok(pagination, 'feedback table must render pagination')
  assert.ok(directiveExpression(pagination, 'current-page', 'model'), 'pagination must bind current page')
  assert.ok(directiveExpression(pagination, 'page-size', 'model'), 'pagination must bind page size')

  const thumbnail = elements(view.template, 'el-image')[0]
  assert.ok(thumbnail, 'feedback list must render the first image with el-image')
  assert.match(directiveExpression(thumbnail, 'if', 'if'), /row\.firstImageUrl/, 'thumbnail must only render when firstImageUrl exists')
  assert.equal(directiveExpression(thumbnail, 'src', 'bind'), 'row.firstImageUrl')
  assert.match(directiveExpression(thumbnail, 'preview-src-list', 'bind'), /row\.firstImageUrl/, 'thumbnail must support image preview')
  assert.equal(attributeValue(thumbnail, 'preview-teleported'), true, 'thumbnail preview must be teleported')
}

assertPageTemplate()

function assertPageRefreshRuntime() {
  assertCalls(view.scriptFile, 'loadOverview', ['adminApi.userFeedbackOverview'])
  assertCalls(view.scriptFile, 'loadList', ['adminApi.userFeedbackList'])
  assertCalls(view.scriptFile, 'refreshOverviewAndList', ['loadOverview', 'loadList'])
  assertCalls(view.scriptFile, 'handleOverviewSelect', ['refreshOverviewAndList'])
  assertCalls(view.scriptFile, 'handleVisibilityChange', ['refreshOverviewAndList'])
  assertCalls(view.scriptFile, 'mountPage', ['refreshOverviewAndList', 'document.addEventListener'])
  assertCalls(view.scriptFile, 'unmountPage', ['invalidateFeedbackRequestSequences', 'document.removeEventListener'])
  assertCalls(view.scriptFile, 'handleStatusSuccess', ['loadOverview', 'loadList'])
  assert.match(functionNode(view.scriptFile, 'handleStatusSuccess').getText(view.scriptFile), /detailDrawerRef\.value\?\.refresh\(\)/)
  const openStatusGuard = firstStatement(view.scriptFile, 'openStatusDialog')
  assert.ok(ts.isIfStatement(openStatusGuard), 'page status dialog handler must start with a permission guard')
  assert.match(openStatusGuard.getText(view.scriptFile), /canHandleUserFeedback\(\)/)
  assert.doesNotMatch(view.script, /\bsetInterval\s*\(/, 'feedback page must not poll on an interval')

  const overviewPhases = guardedRequestPhases(view.scriptFile, 'loadOverview', 'applyIfFeedbackRequestCurrent')
  assert.match(overviewPhases.success, /overview\.value\s*=\s*response\.data/, 'overview success update must be guarded')
  assert.match(overviewPhases.catch, /overviewError\.value\s*=/, 'overview error state must be guarded')
  assert.match(overviewPhases.catch, /showRequestError/, 'overview global error must be guarded')
  assert.match(overviewPhases.finally, /overviewLoading\.value\s*=\s*false/, 'overview loading cleanup must be guarded')

  const listPhases = guardedRequestPhases(view.scriptFile, 'loadList', 'applyIfFeedbackRequestCurrent')
  assert.match(listPhases.success, /rows\.value\s*=\s*response\.data\.list/, 'list success update must be guarded')
  assert.match(listPhases.catch, /listError\.value\s*=/, 'list error state must be guarded')
  assert.match(listPhases.catch, /showRequestError/, 'list global error must be guarded')
  assert.match(listPhases.finally, /listLoading\.value\s*=\s*false/, 'list loading cleanup must be guarded')

  const exported = executeTypeScriptModule(view.descriptor.script?.content || '', `${paths.view}.module.ts`)
  const sequences = { overview: 4, list: 9 }
  const staleOverview = sequences.overview
  const staleList = sequences.list
  exported.invalidateFeedbackRequestSequences(sequences)
  assert.deepEqual(sequences, { overview: 5, list: 10 }, 'unmount invalidation must advance both request sequences')
  let staleEffects = 0
  assert.equal(exported.applyIfFeedbackRequestCurrent(sequences, 'overview', staleOverview, () => { staleEffects += 1 }), false)
  assert.equal(exported.applyIfFeedbackRequestCurrent(sequences, 'list', staleList, () => { staleEffects += 1 }), false)
  assert.equal(staleEffects, 0, 'stale responses must not update state or show global errors after unmount')
  assert.equal(exported.applyIfFeedbackRequestCurrent(sequences, 'list', sequences.list, () => { staleEffects += 1 }), true)
  assert.equal(staleEffects, 1, 'the current request must still apply its side effects')
}

assertPageRefreshRuntime()

function assertFilterQueryRuntime() {
  const exported = executeTypeScriptModule(view.descriptor.script?.content || '', `${paths.view}.module.ts`)
  const filters = {
    keyword: '  FB-20260905 用户补充  ',
    status: 'processing',
    handlerId: 17,
    submittedDates: ['2026-09-01', '2026-09-05'],
  }
  const query = { ...exported.buildUserFeedbackQuery(filters, 3, 50) }
  assert.equal(query.keyword, 'FB-20260905 用户补充')
  assert.equal(query.status, 'processing')
  assert.equal(query.handlerId, 17)
  assert.equal(query.page, 3)
  assert.equal(query.pageSize, 50)
  assert.equal(query.submittedTo, new Date(2026, 8, 6).getTime() - 1)
  assert.equal(query.submittedFrom, new Date(2026, 8, 1).getTime())

  const originalTimeZone = process.env.TZ
  try {
    process.env.TZ = 'America/New_York'
    const dstQuery = exported.buildUserFeedbackQuery({
      keyword: '',
      status: '',
      handlerId: 0,
      submittedDates: ['2026-03-08', '2026-03-08'],
    }, 1, 20)
    assert.equal(dstQuery.submittedFrom, Date.parse('2026-03-08T05:00:00.000Z'))
    assert.equal(dstQuery.submittedTo, Date.parse('2026-03-09T04:00:00.000Z') - 1)
    assert.equal(dstQuery.submittedTo - dstQuery.submittedFrom, 23 * 60 * 60 * 1000 - 1, 'DST spring-forward date must end at next local midnight, not after a fixed 24 hours')
  } finally {
    if (originalTimeZone === undefined) delete process.env.TZ
    else process.env.TZ = originalTimeZone
  }
  assert.doesNotMatch(view.descriptor.script?.content || '', /86_?400_?000/, 'date range must not add a fixed 24-hour duration')

  const overviewQuery = { ...exported.buildUserFeedbackOverviewQuery(filters) }
  assert.equal('status' in overviewQuery, false, 'overview must aggregate every status under the other filters')
  assert.equal('page' in overviewQuery, false, 'overview must not send pagination')
  assert.equal('pageSize' in overviewQuery, false, 'overview must not send pagination')
  assert.equal(overviewQuery.keyword, query.keyword)
  assert.equal(overviewQuery.handlerId, query.handlerId)
  assert.equal(overviewQuery.submittedFrom, query.submittedFrom)
  assert.equal(overviewQuery.submittedTo, query.submittedTo)

  const emptyQuery = { ...exported.buildUserFeedbackQuery({
    keyword: '   ',
    status: '',
    handlerId: 0,
    submittedDates: null,
  }, 1, 20) }
  assert.deepEqual(emptyQuery, { page: 1, pageSize: 20 }, 'empty filters must not emit ambiguous query values')
}

assertFilterQueryRuntime()

function assertOverviewComponent() {
  const buttons = elements(overview.template, 'button')
  assert.equal(buttons.length, 1, 'overview must use one repeated semantic button')
  assert.match(directiveExpression(buttons[0], 'for', 'for'), /overviewItems/)
  assert.match(directiveExpression(buttons[0], 'click', 'on'), /emit\('select'/)
  for (const status of ['pending', 'processing', 'resolved', 'closed']) {
    assert.match(overview.script, new RegExp(`status:\\s*'${status}'`), `overview must render ${status}`)
  }
  assert.ok(elements(overview.template, 'el-alert').length > 0, 'overview must expose its error state')
  assert.ok(elements(overview.template, 'el-skeleton').length > 0, 'overview must expose its loading state')
}

assertOverviewComponent()

function assertDetailDrawerComponent() {
  const drawers = elements(detailDrawer.template, 'AdminDrawer')
  assert.equal(drawers.length, 1)
  assert.match(directiveExpression(drawers[0], 'show-footer', 'bind'), /canHandle/, 'detail drawer footer must not render without handle permission')
  assert.match(directiveExpression(drawers[0], 'confirm', 'on'), /requestStatusUpdate/, 'fixed drawer footer must enter the guarded update handler')
  assert.equal(elements(detailDrawer.template, 'UserFeedbackTimeline').length, 1)
  const first = firstStatement(detailDrawer.scriptFile, 'requestStatusUpdate')
  assert.ok(ts.isIfStatement(first), 'drawer update handler must start with a permission guard')
  assert.match(first.getText(detailDrawer.scriptFile), /canHandleUserFeedback\(\)/)
  assertCalls(detailDrawer.scriptFile, 'loadDetail', ['adminApi.userFeedbackDetail'])
  assertCalls(detailDrawer.scriptFile, 'unmountDrawer', ['invalidateDetailRequestSequence'])
  assert.match(detailDrawer.script, /onBeforeUnmount\s*\(\s*unmountDrawer\s*\)/, 'detail drawer must invalidate pending requests when unmounted')

  const phases = guardedRequestPhases(detailDrawer.scriptFile, 'loadDetail', 'applyIfDetailRequestCurrent')
  assert.match(phases.success, /detail\.value\s*=\s*response\.data/, 'detail success update must be guarded')
  assert.match(phases.success, /emit\(\s*'detail-change'/, 'detail success emit must be guarded')
  assert.match(phases.catch, /detail\.value\s*=\s*null/, 'detail error reset must be guarded')
  assert.match(phases.catch, /showRequestError/, 'detail global error must be guarded')
  assert.match(phases.finally, /loading\.value\s*=\s*false/, 'detail loading cleanup must be guarded')

  const exported = executeTypeScriptModule(detailDrawer.descriptor.script?.content || '', `${paths.detailDrawer}.module.ts`)
  const staleSequence = 7
  const currentSequence = exported.invalidateDetailRequestSequence(staleSequence)
  assert.equal(currentSequence, 8, 'detail unmount invalidation must advance its request sequence')
  let staleEffects = 0
  for (const phase of ['success', 'catch', 'finally']) {
    assert.equal(exported.applyIfDetailRequestCurrent(currentSequence, staleSequence, () => { staleEffects += 1 }), false, `stale detail ${phase} effect must be suppressed`)
  }
  assert.equal(staleEffects, 0, 'unmounted detail request must have no success, catch, or finally side effects')
  assert.equal(exported.applyIfDetailRequestCurrent(currentSequence, currentSequence, () => { staleEffects += 1 }), true)
  assert.equal(staleEffects, 1, 'current detail request must still apply effects')
  assert.match(detailDrawer.script, /defineExpose\s*\(\s*\{\s*refresh:\s*loadDetail\s*\}\s*\)/)
}

assertDetailDrawerComponent()

function assertTimelineComponent() {
  assert.equal(elements(timeline.template, 'el-timeline').length, 1)
  assert.equal(elements(timeline.template, 'el-timeline-item').length, 1, 'timeline must repeat one timeline item template')
  const images = elements(timeline.template, 'el-image')
  assert.equal(images.length, 1, 'timeline must use one repeated image preview')
  assert.ok(directiveExpression(images[0], 'preview-src-list', 'bind'), 'timeline images must expose preview sources')
  assert.equal(attributeValue(images[0], 'preview-teleported'), true, 'image preview must be teleported')
  assert.match(timeline.script, /\.sort\s*\(.*createdAt/s, 'timeline must sort all messages by createdAt')
  for (const messageType of ['initial', 'supplement', 'status']) {
    assert.match(timeline.script, new RegExp(`case\\s+'${messageType}'`), `timeline must label ${messageType} messages`)
  }
}

assertTimelineComponent()

function assertStatusDialogComponent() {
  const dialogs = elements(statusDialog.template, 'AdminDialog')
  assert.equal(dialogs.length, 1)
  assert.equal(attributeValue(dialogs[0], 'append-to-body'), true, 'status dialog must append to body')
  assert.equal(directiveExpression(dialogs[0], 'before-close', 'bind'), 'beforeClose', 'status dialog must bind a real before-close guard')
  assert.equal(elements(statusDialog.template, 'el-select').length, 1)
  assert.equal(elements(statusDialog.template, 'el-input').length, 1)
  assert.equal(elements(statusDialog.template, 'el-switch').length, 1)

  for (const functionName of ['open', 'submit']) {
    const first = firstStatement(statusDialog.scriptFile, functionName)
    assert.ok(ts.isIfStatement(first), `${functionName} must start with a permission guard`)
    assert.match(first.getText(statusDialog.scriptFile), /canHandleUserFeedback\(\)/)
  }
  assertCalls(statusDialog.scriptFile, 'submit', ['adminApi.userFeedbackUpdateStatus'])
  const submit = functionNode(statusDialog.scriptFile, 'submit').getText(statusDialog.scriptFile)
  assert.match(submit, /nextUserFeedbackStatuses/, 'submit must re-check the legal transition')
  assert.match(submit, /visible\.value\s*=\s*false/, 'dialog may close after a successful request')
  const failed = catchClause(statusDialog.scriptFile, 'submit').getText(statusDialog.scriptFile)
  assert.doesNotMatch(failed, /visible\.value\s*=\s*false/, 'request failure must keep the dialog open')
  assert.doesNotMatch(failed, /form\.(?:status|note|notifyUser)\s*=|requestId\.value\s*=/, 'request failure must retain the draft and request id')
  assert.match(failed, /反馈已更新，请刷新详情后再处理/, 'version conflict must prompt the user to refresh detail')
  const submitFunction = functionNode(statusDialog.scriptFile, 'submit')
  let finalizer
  const findFinalizer = (node) => {
    if (ts.isTryStatement(node) && node.finallyBlock) finalizer = node.finallyBlock
    ts.forEachChild(node, findFinalizer)
  }
  findFinalizer(submitFunction)
  assert.ok(finalizer, 'status submit must restore retry state in finally')
  assert.match(finalizer.getText(statusDialog.scriptFile), /submitting\.value\s*=\s*false/, 'failed status update must become retryable')
  assert.equal((statusDialog.script.match(/requestId\.value\s*=\s*newRequestId\(\)/g) || []).length, 1, 'request id must be created once when opening')

  const exported = executeTypeScriptModule(statusDialog.descriptor.script?.content || '', `${paths.statusDialog}.module.ts`)
  assert.equal(exported.requiresStatusNote('pending', 'processing'), false)
  for (const [from, to] of [['pending', 'closed'], ['processing', 'resolved'], ['resolved', 'closed'], ['resolved', 'processing'], ['closed', 'processing']]) {
    assert.equal(exported.requiresStatusNote(from, to), true, `${from} -> ${to} must require a note`)
  }
  assert.equal(exported.isVersionConflict({ msg: '反馈已更新，请刷新后重试' }), true)
  assert.equal(exported.isVersionConflict({ msg: '网络错误' }), false)

  let closeCalls = 0
  assert.equal(exported.closeStatusDialog(false, () => { closeCalls += 1 }), true)
  assert.equal(closeCalls, 1, 'idle dialog close must call Element Plus done callback')
  assert.equal(exported.closeStatusDialog(true, () => { closeCalls += 1 }), false)
  assert.equal(closeCalls, 1, 'submitting dialog close must not call Element Plus done callback')
  assertCalls(statusDialog.scriptFile, 'beforeClose', ['closeStatusDialog'])
}

assertStatusDialogComponent()

assert.match(sources.viteConfig, /['"]@['"]\s*:\s*fileURLToPath\(new URL\(['"]\.\/src['"],\s*import\.meta\.url\)\)/)

const packageJSON = JSON.parse(sources.package)
assert.equal(packageJSON.scripts?.['check:user-feedback'], 'node scripts/check-user-feedback.mjs')
assert.match(packageJSON.scripts?.['check:all'] || '', /(?:^|&&\s*)npm run check:user-feedback(?:\s*&&|$)/)

console.log('Admin user feedback page AST and runtime contract passed.')
