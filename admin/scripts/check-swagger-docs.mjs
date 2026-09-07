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
const swaggerProxyMatch = vite.match(/'([^']*swagger[^']*)'\s*:\s*\{/)
if (!swaggerProxyMatch) throw new Error('Vite must proxy /swagger to the backend')

const swaggerProxyPattern = new RegExp(swaggerProxyMatch[1])
for (const path of ['/swagger', '/swagger?url=/swagger/doc.json', '/swagger/index.html']) {
  if (!swaggerProxyPattern.test(path)) throw new Error(`Vite Swagger proxy must match ${path}`)
}
if (swaggerProxyPattern.test('/swagger-docs')) {
  throw new Error('Vite Swagger proxy must not intercept the /swagger-docs admin route')
}

const swaggerRoute = readWorkspace('backend/internal/routes/common/swagger.go')
for (const snippet of ['swagger.WrapHandler', 'serveSwaggerIndex(c)', '/swagger/doc.json']) {
  if (!swaggerRoute.includes(snippet)) throw new Error(`backend Swagger route missing ${snippet}`)
}

const swaggerIndex = readWorkspace('backend/internal/routes/common/swagger_index.go')
for (const snippet of ['SwaggerUIBundle({', 'url: "/swagger/doc.json"', 'filter: true', '<title>WeCheckin API</title>']) {
  if (!swaggerIndex.includes(snippet)) throw new Error(`backend Swagger index missing ${snippet}`)
}

const backendMain = readWorkspace('backend/cmd/main.go')
if (!backendMain.includes("--templateDelims '{%,%}'")) {
  throw new Error('Swagger generation must use template delimiters that preserve workflow {{variables}}')
}

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
