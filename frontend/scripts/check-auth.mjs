import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs'
import { dirname, extname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const root = resolve(currentDir, '..')
const authPath = resolve(root, 'utils/auth.js')
const scanTargets = [
  resolve(root, 'App.vue'),
  resolve(root, 'utils'),
  resolve(root, 'pages'),
]
const sourceExts = new Set(['.vue', '.js', '.ts'])
const directAuthStoragePattern = /uni\.(getStorageSync|setStorageSync|removeStorageSync)\(['"](token|userInfo|admin_token|admin_info)['"]\)/g

if (!existsSync(authPath)) {
  throw new Error('frontend auth wrapper missing: utils/auth.js')
}

function collectFiles(path) {
  if (!existsSync(path)) return []
  const stat = statSync(path)
  if (stat.isFile()) {
    return sourceExts.has(extname(path)) ? [path] : []
  }
  return readdirSync(path)
    .flatMap((entry) => collectFiles(resolve(path, entry)))
}

const violations = []
for (const file of scanTargets.flatMap(collectFiles)) {
  if (file === authPath) continue
  const source = readFileSync(file, 'utf8')
  const lines = source.split(/\r?\n/)
  lines.forEach((line, index) => {
    if (directAuthStoragePattern.test(line)) {
      violations.push(`${file.replace(root + '/', '')}:${index + 1}: ${line.trim()}`)
    }
    directAuthStoragePattern.lastIndex = 0
  })
}

if (violations.length > 0) {
  throw new Error(`frontend auth storage must go through utils/auth.js:\n${violations.join('\n')}`)
}

const authSource = readFileSync(authPath, 'utf8')
const requiredAuthSnippets = [
  'export const CLIENT_TOKEN_KEY',
  'export const CLIENT_INFO_KEY',
  'export const ADMIN_TOKEN_KEY',
  'export const ADMIN_INFO_KEY',
  'export function getClientAuth',
  'export function setClientAuth',
  'export function clearClientAuth',
  'export function getAdminAuth',
  'export function setAdminAuth',
  'export function clearAdminAuth',
  'export function getClientUserId',
  'export function getRequestAuthState',
  'export function clearRequestAuthState',
]

for (const snippet of requiredAuthSnippets) {
  if (!authSource.includes(snippet)) {
    throw new Error(`frontend auth wrapper missing: ${snippet}`)
  }
}

const requestSource = readFileSync(resolve(root, 'utils/request.js'), 'utf8')
if (!requestSource.includes("from './auth'")) {
  throw new Error('frontend request layer must import auth helpers')
}
