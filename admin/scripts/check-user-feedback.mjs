import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
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
  id: required('number'), originalName: required('string'), contentType: optional('string'), sizeBytes: required('number'),
  url: required('string'), sortOrder: required('number'), createdAt: required('number'),
})
assertInterface('UserFeedbackMessage', {
  id: required('number'), messageType: required('UserFeedbackMessageType'), authorType: required('UserFeedbackAuthorType'),
  authorId: required('number'), authorName: optional('string'), content: required('string'),
  fromStatus: optional('UserFeedbackStatus'), toStatus: optional('UserFeedbackStatus'),
  attachments: required('UserFeedbackAttachment[]'), createdAt: required('number'),
})
assertInterface('UserFeedbackSummary', {
  id: required('number'), feedbackNo: required('string'), submitterId: required('number'), submitterName: optional('string'),
  summary: required('string'), imageCount: required('number'), status: required('UserFeedbackStatus'),
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

assert.match(sources.view, /from\s+['"]@\/components\/admin-ui['"]/)
assert.match(sources.view, /<AdminPageShell(?:\s|>)/)
assert.match(sources.viteConfig, /['"]@['"]\s*:\s*fileURLToPath\(new URL\(['"]\.\/src['"],\s*import\.meta\.url\)\)/)

const packageJSON = JSON.parse(sources.package)
assert.equal(packageJSON.scripts?.['check:user-feedback'], 'node scripts/check-user-feedback.mjs')
assert.match(packageJSON.scripts?.['check:all'] || '', /(?:^|&&\s*)npm run check:user-feedback(?:\s*&&|$)/)

console.log('Admin user feedback AST and runtime contract passed.')
