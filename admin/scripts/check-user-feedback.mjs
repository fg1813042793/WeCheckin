import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import ts from 'typescript'

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
}

for (const [name, path] of Object.entries(paths)) {
  if (!existsSync(path)) {
    throw new Error(`user feedback ${name} contract missing: ${path}`)
  }
}

const sources = Object.fromEntries(
  Object.entries(paths).map(([name, path]) => [name, readFileSync(path, 'utf8')]),
)

function fail(message) {
  throw new Error(`user feedback contract failed: ${message}`)
}

function assertIncludes(sourceName, snippets) {
  for (const snippet of snippets) {
    if (!sources[sourceName].includes(snippet)) {
      fail(`${sourceName} missing ${snippet}`)
    }
  }
}

function propertyName(node) {
  return ts.isIdentifier(node) || ts.isStringLiteralLike(node) ? node.text : node.getText()
}

function interfaceContract(name) {
  const sourceFile = ts.createSourceFile(paths.types, sources.types, ts.ScriptTarget.Latest, true)
  let declaration
  sourceFile.forEachChild((node) => {
    if (ts.isInterfaceDeclaration(node) && node.name.text === name) declaration = node
  })
  if (!declaration) fail(`types missing interface ${name}`)
  return {
    extends: declaration.heritageClauses?.flatMap(clause => clause.types.map(type => type.expression.getText(sourceFile))) || [],
    fields: Object.fromEntries(declaration.members.map((member) => {
      if (!ts.isPropertySignature(member) || !member.type) fail(`${name} contains unsupported member`)
      return [propertyName(member.name), {
        optional: Boolean(member.questionToken),
        type: member.type.getText(sourceFile),
      }]
    })),
  }
}

function assertInterface(name, expectedFields, expectedExtends = []) {
  const actual = interfaceContract(name)
  if (JSON.stringify(actual.extends) !== JSON.stringify(expectedExtends)) {
    fail(`${name} extends ${JSON.stringify(actual.extends)}, expected ${JSON.stringify(expectedExtends)}`)
  }
  if (JSON.stringify(actual.fields) !== JSON.stringify(expectedFields)) {
    fail(`${name} fields do not match Backend DTO: ${JSON.stringify(actual.fields)}`)
  }
}

const required = type => ({ optional: false, type })
const optional = type => ({ optional: true, type })

assertIncludes('types', [
  "export type UserFeedbackStatus = 'pending' | 'processing' | 'resolved' | 'closed'",
  "export type UserFeedbackMessageType = 'initial' | 'supplement' | 'status'",
  "export type UserFeedbackAuthorType = 'user' | 'admin'",
])

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

assertIncludes('api', [
  "from '../types/userFeedback'",
  'userFeedbackOverview(params: AdminUserFeedbackListQuery = {})',
  'request.get<UserFeedbackOverview>(`${ADMIN_V2}/user-feedbacks/overview`, { params })',
  'userFeedbackList(params: AdminUserFeedbackListQuery = {})',
  'request.get<UserFeedbackList>(`${ADMIN_V2}/user-feedbacks`, { params })',
  'userFeedbackDetail(id: ID)',
  'request.get<UserFeedbackDetail>(`${ADMIN_V2}/user-feedbacks/${encodePath(id)}`)',
  'userFeedbackUpdateStatus(id: ID, data: UpdateUserFeedbackStatusInput)',
  'request.patch<UserFeedbackDetail, UpdateUserFeedbackStatusInput>(`${ADMIN_V2}/user-feedbacks/${encodePath(id)}/status`, data, jsonConfig)',
])

assertIncludes('routes', [
  "path: 'user-feedbacks'",
  "component: () => import('../views/user-feedback/index.vue')",
  "menuPath: '/user-feedbacks'",
  "requiredPermission: 'admin:menu:user-feedback:list'",
  "adminUi: { version: 1, pattern: 'filter-list' }",
])
assertIncludes('routeMeta', ['requiredPermission?: string'])
assertIncludes('routeAccess', [
  'meta.requiredPermission',
  'snapshot.permissions.includes(meta.requiredPermission)',
])

assertIncludes('status', [
  "export const USER_FEEDBACK_LIST_PERMISSION = 'admin:menu:user-feedback:list'",
  "export const USER_FEEDBACK_HANDLE_PERMISSION = 'admin:menu:user-feedback:handle'",
  'export function canHandleUserFeedback',
  'hasPerm(USER_FEEDBACK_HANDLE_PERMISSION)',
  "pending: { label: '待处理', type: 'warning' }",
  "processing: { label: '处理中', type: 'warning' }",
  "resolved: { label: '已解决', type: 'success' }",
  "closed: { label: '已关闭', type: 'info' }",
  "pending: ['processing', 'closed']",
  "processing: ['resolved']",
  "resolved: ['closed', 'processing']",
  "closed: ['processing']",
  'export function nextUserFeedbackStatuses',
])
assertIncludes('view', [
  "from '@/components/admin-ui'",
  '<AdminPageShell',
])
assertIncludes('viteConfig', [
  "'@': fileURLToPath(new URL('./src', import.meta.url))",
])

console.log('Admin user feedback contract passed.')
