import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const root = resolve(currentDir, '..')
const workspaceRoot = resolve(root, '..')
const readAdmin = path => readFileSync(resolve(root, path), 'utf8')
const readWorkspace = path => readFileSync(resolve(workspaceRoot, path), 'utf8')

const routes = readAdmin('src/router/adminRoutes.ts')
for (const snippet of ["path: 'swagger-docs'", "name: 'SwaggerDocs'", "views/swagger-docs/index.vue", "title: '接口文档'"]) {
  if (!routes.includes(snippet)) throw new Error(`swagger docs route missing ${snippet}`)
}

const page = readAdmin('src/views/swagger-docs/index.vue')
for (const snippet of [
  '/swagger/doc.json',
  '/swagger/index.html',
  '<iframe',
  'fetch(',
  'window.open(',
  'RefreshRight',
  'TopRight',
  '<el-result',
  'aria-label="刷新接口文档"',
  'aria-label="在新窗口打开接口文档"',
]) {
  if (!page.includes(snippet)) throw new Error(`swagger docs page missing ${snippet}`)
}

const vite = readAdmin('vite.config.ts')
if (!vite.includes("'/swagger'")) throw new Error('Vite must proxy /swagger to the backend')

for (const path of ['nginx/default.conf.template', 'admin/nginx/default.conf.template']) {
  const source = readWorkspace(path)
  if (!source.includes('location = /swagger')) throw new Error(`${path} must proxy the exact /swagger path`)
  if (!source.includes('location /swagger/')) throw new Error(`${path} must proxy /swagger/`)
  if (!source.includes('proxy_pass ${NGINX_API_UPSTREAM};')) throw new Error(`${path} swagger proxy must use NGINX_API_UPSTREAM`)
}

const packageJSON = JSON.parse(readAdmin('package.json'))
if (packageJSON.scripts?.['check:swagger-docs'] !== 'node scripts/check-swagger-docs.mjs') {
  throw new Error('package.json must expose check:swagger-docs')
}
if (!packageJSON.scripts?.['check:all']?.includes('npm run check:swagger-docs')) {
  throw new Error('check:all must include check:swagger-docs')
}
