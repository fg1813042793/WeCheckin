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
const clientIdentityParamPattern = /\buser_id\s*:/
const clientIdentityParamAllowlist = new Set([
  'pages/index/index.vue',
  'pages/login/login.vue',
  'pages/login/login_pwd.vue',
  'pages/my/my_reg.vue'
])

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
  const relativeFile = file.replace(root + '/', '')
  const source = readFileSync(file, 'utf8')
  const lines = source.split(/\r?\n/)
  lines.forEach((line, index) => {
    if (directAuthStoragePattern.test(line)) {
      violations.push(`${relativeFile}:${index + 1}: ${line.trim()}`)
    }
    directAuthStoragePattern.lastIndex = 0
    if (clientIdentityParamPattern.test(line) && !clientIdentityParamAllowlist.has(relativeFile)) {
      violations.push(`${relativeFile}:${index + 1}: authenticated client requests must not submit user_id`)
    }
  })
}

if (violations.length > 0) {
  throw new Error(`frontend auth storage must go through utils/auth.js:\n${violations.join('\n')}`)
}

const authSource = readFileSync(authPath, 'utf8')
if (/getClientToken\(\)[\s\S]{0,160}\|\|\s*token\s*\|\|/.test(authSource)) {
  throw new Error('getClientUserId must never fall back to the client token')
}

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
