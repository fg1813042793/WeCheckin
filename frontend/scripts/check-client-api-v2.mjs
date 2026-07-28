import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const root = resolve(currentDir, '..')

const checkedFiles = [
  'api/index.js',
  'api/admin.js',
  'pages/about/agreement.vue',
  'pages/about/privacy.vue',
]

const requestCallPattern = /\b(?:get|post|postJSON|put|patch|del|apiGet|apiPost)\(\s*['"]([^'"]+)['"]/g
const legacyCalls = []

for (const file of checkedFiles) {
  const source = readFileSync(resolve(root, file), 'utf8')
  for (const match of source.matchAll(requestCallPattern)) {
    const endpoint = match[1]
    if (endpoint.startsWith('/') && !endpoint.startsWith('/api/v2/')) {
      legacyCalls.push(`${file}: ${endpoint}`)
    }
  }
}

if (legacyCalls.length) {
  throw new Error(`client API still uses legacy endpoints:\n${legacyCalls.join('\n')}`)
}

const apiIndex = readFileSync(resolve(root, 'api/index.js'), 'utf8')
const adminApi = readFileSync(resolve(root, 'api/admin.js'), 'utf8')
const requestLayer = readFileSync(resolve(root, 'utils/request.js'), 'utf8')

function sourceIncludesEndpoint(source, endpoint) {
  return source.includes(endpoint) ||
    source.includes(endpoint.replace('/api/v2/admin', '${ADMIN_V2}')) ||
    source.includes(endpoint.replace('/api/v2', '${API_V2}'))
}

const requiredClientEndpoints = [
  '/api/v2/home',
  '/api/v2/auth/login',
  '/api/v2/me',
  '/api/v2/user-form-fields',
  '/api/v2/news/categories',
  '/api/v2/enrollments',
  '/api/v2/events',
  '/api/v2/surveys',
  '/api/v2/exams',
]

for (const endpoint of requiredClientEndpoints) {
  if (!sourceIncludesEndpoint(apiIndex, endpoint)) {
    throw new Error(`client API missing v2 endpoint: ${endpoint}`)
  }
}

const requiredAdminEndpoints = [
  '/api/v2/admin/auth/login',
  '/api/v2/admin/home',
  '/api/v2/admin/users',
  '/api/v2/admin/enrollments',
  '/api/v2/admin/news',
  '/api/v2/admin/events',
  '/api/v2/admin/managers',
  '/api/v2/admin/me/perms',
]

for (const endpoint of requiredAdminEndpoints) {
  if (!sourceIncludesEndpoint(adminApi, endpoint)) {
    throw new Error(`mobile admin API missing v2 endpoint: ${endpoint}`)
  }
}

if (!requestLayer.includes("startsWith('/api/v2/admin/')")) {
  throw new Error('request layer must treat /api/v2/admin/* as admin requests')
}
