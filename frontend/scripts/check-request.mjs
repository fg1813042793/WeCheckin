import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const source = readFileSync(resolve(currentDir, '../utils/request.js'), 'utf8')

const requiredSnippets = [
  'const LOGIN_EXPIRED_MESSAGES = new Set',
  'function getAuthState',
  'function clearAuthState',
  'function redirectToLogin',
  'LOGIN_EXPIRED_MESSAGES.has',
  'reject(res.data)',
  "method: (options.method || 'GET').toUpperCase()",
  'timeout: options.timeout || 15000',
  'data,',
]

for (const snippet of requiredSnippets) {
  if (!source.includes(snippet)) {
    throw new Error(`frontend request layer missing: ${snippet}`)
  }
}

if (source.includes('data: JSON.stringify(data)')) {
  throw new Error('postJSON should pass object data to uni.request instead of pre-stringifying')
}
