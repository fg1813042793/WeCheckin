import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
import { executeTypeScriptModule } from './lib/typescript-runtime.mjs'

const read = path => readFileSync(new URL(`../${path}`, import.meta.url), 'utf8')
const pages = ['user', 'mgr', 'role']
function exportFunction(page) {
  const source = read(`src/views/${page}/index.vue`).split(/<script[^>]*>/)[1].split('</script>')[0]
  const ast = ts.createSourceFile(`${page}.ts`, source, ts.ScriptTarget.Latest, true)
  const fn = ast.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === 'exportData')
  assert.ok(fn?.modifiers?.some(modifier => modifier.kind === ts.SyntaxKind.AsyncKeyword), `${page}: export must await the audited endpoint`)
  return fn.getText(ast)
}
const functions = Object.fromEntries(pages.map(page => [page, exportFunction(page)]))
const blobs = []
const clicks = []
const revoked = []
const browser = {
  Blob,
  URL: {
    createObjectURL(blob) { blobs.push(blob); return `blob:test-${blobs.length}` },
    revokeObjectURL(url) { revoked.push(url) },
  },
  document: {
    body: { appendChild() {} },
    createElement() { return { href: '', download: '', click() { clicks.push(this.download) }, remove() {} } },
  },
  setTimeout(fn) { fn() },
}
const { buildCsv, exportCsvPage } = executeTypeScriptModule(
  `import { Blob, URL, document, setTimeout } from 'test-browser'\n${read('src/utils/csvExport.ts')}`,
  'csvExport.ts', { 'test-browser': browser },
)

assert.equal(buildCsv(['列'], [['中文,"换行\n下一行'], [null], [false], [0]]), '\uFEFF"列"\r\n"中文,""换行\n下一行"\r\n""\r\n"false"\r\n"0"')
for (const value of ['=1+1', '+SUM(1,2)', '-1+2', '@SUM(1)', '\t=1', '\r=1', '\n=1', '  =1', '\uFEFF=1', '\u0000=1', '\u200b=1']) {
  assert.ok(buildCsv(['列'], [[value]]).split('\r\n')[1].startsWith('"\''), `unescaped formula prefix: ${JSON.stringify(value)}`)
}
assert.throws(() => buildCsv(['列'], [[{ private: 'do not serialize objects' }]]), /CSV/)
await assert.rejects(exportCsvPage('invalid.csv', ['列'], async () => null, row => [row]), /列表/)
await assert.rejects(exportCsvPage('large.csv', ['列'], async () => Array(1001).fill('x'), row => [row]), /1000/)
assert.equal(clicks.length, 0)

const requests = []
let responseRows = [], failure, release
const { adminApi } = executeTypeScriptModule(read('src/api/index.ts'), 'api.ts', {
  '../utils/request': { get: async (url, { params }) => {
    requests.push({ url, params: { ...params } })
    if (release) await release.promise
    if (failure) throw failure
    return { code: 0, data: { list: responseRows, total: 200 } }
  } },
})
const notices = []
const { showRequestError } = executeTypeScriptModule(read('src/utils/request.ts'), 'request.ts', {
  axios: { create: () => ({ interceptors: { request: { use() {} }, response: { use() {} } } }) },
  'element-plus': { ElMessage: { error: message => notices.push(message) } },
  './adminSession': { clearAdminSession() {} },
})
const filters = { keyword: '测试', mobile: '130', status: 0, positionId: 2, managerUserId: 3, roleId: 4 }
const cases = {
  user: { method: 'userExport', resource: 'users', filename: '用户列表.csv', row: { name: 'fresh-user', mobile: '13000000000', deptIds: [8], status: 1 } },
  mgr: { method: 'mgrExport', resource: 'managers', filename: '管理员列表.csv', row: { name: 'fresh-manager', status: 0 } },
  role: { method: 'roleExport', resource: 'roles', filename: '角色列表.csv', row: { name: 'fresh-role', status: 1, allowAdminLogin: 0, remark: '=formula' } },
}
for (const page of pages) {
  const item = cases[page]
  assert.equal(typeof adminApi[item.method], 'function')
  const bindings = {
    adminApi, exportCsvPage, showRequestError,
    page: { value: 3 }, pageSize: { value: 50 }, keyword: { value: '测试' },
    sortRules: { value: [{ field: 'name', order: 'desc' }] },
    buildUserListQuery: (page, pageSize) => ({ ...filters, page, pageSize }),
    deptPathText: ids => ids.join('、'),
    list: new Proxy({}, { get() { throw new Error('cached rows must never be accessed') } }),
  }
  const { exportData } = executeTypeScriptModule(
    `import { ${Object.keys(bindings).join(', ')} } from 'bindings'\nexport ${functions[page]}`,
    `${page}-export.ts`, { bindings },
  )
  responseRows = [item.row]; failure = undefined
  await exportData()
  assert.equal(requests.at(-1).url, `/api/v2/admin/${item.resource}/export`)
  assert.deepEqual(requests.at(-1).params, page === 'user'
    ? { ...filters, page: 3, pageSize: 50, sort: 'name:desc' }
    : { page: 3, pageSize: 50, keyword: '测试' })
  assert.equal(clicks.at(-1), item.filename)
  const csv = await blobs.at(-1).text()
  assert.ok(csv.includes(item.row.name))
  if (page === 'user') assert.ok(csv.includes('13000000000'), 'mobile field is missing')
  if (page === 'role') assert.ok(csv.includes("'=formula"), 'role remark formula is unsafe')

  let downloads = clicks.length
  failure = { msg: '请求失败', __adminRequestErrorNotified: true }
  await exportData()
  assert.equal(clicks.length, downloads, 'failed fetch must not download cached rows')
  assert.equal(notices.length, 0, 'request layer already notified; no duplicate toast')
  failure = new Error('unnotified failure')
  await exportData()
  assert.equal(notices.pop(), '导出失败')
  assert.equal(clicks.length, downloads)

  failure = undefined
  let resume
  release = { promise: new Promise(resolve => { resume = resolve }) }
  const before = requests.length
  const first = exportData(), second = exportData()
  assert.equal(requests.length, before + 1, 'double click should fetch only once')
  resume(); await Promise.all([first, second]); release = undefined
  assert.equal(clicks.length, downloads + 1)
}
assert.equal(revoked.length, blobs.length, 'object URLs must be released')
console.log('CSV export checks passed: current-page API queries, no stale fallback, single error notice, duplicate-click guard, formula escaping and URL cleanup')
