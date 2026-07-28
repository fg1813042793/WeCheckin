import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const rootDir = resolve(currentDir, '..')

function readJson(path) {
  return JSON.parse(readFileSync(resolve(rootDir, path), 'utf8'))
}

function requireScript(pkg, scriptName, requiredSnippet, label) {
  const script = pkg.scripts?.[scriptName]
  if (!script) {
    throw new Error(`${label} missing npm script: ${scriptName}`)
  }
  if (requiredSnippet && !script.includes(requiredSnippet)) {
    throw new Error(`${label} script ${scriptName} must include: ${requiredSnippet}`)
  }
}

const adminPkg = readJson('admin/package.json')
const frontendPkg = readJson('frontend/package.json')
const verifyLocal = readFileSync(resolve(rootDir, 'scripts/verify-local.sh'), 'utf8')
const maintenanceDoc = readFileSync(resolve(rootDir, 'docs/project-maintenance.md'), 'utf8')

for (const [scriptName, snippet] of [
  ['check:all', 'check:formkit-shared-types'],
  ['check:all', 'check:bundle'],
  ['check:formkit-shared-types', 'check-formkit-shared-types.mjs'],
  ['check:bundle', 'check-bundle-budget.mjs'],
]) {
  requireScript(adminPkg, scriptName, snippet, 'admin')
}

for (const [scriptName, snippet] of [
  ['check:all', 'check:form-fill'],
  ['check:all', 'check:news-search-layout'],
  ['check:all', 'check:fixed-list-headers'],
  ['check:all', 'check:search-input-style'],
  ['check:all', 'check:logs'],
  ['check:form-fill', 'check-form-fill.mjs'],
  ['check:news-search-layout', 'check-news-search-layout.mjs'],
  ['check:fixed-list-headers', 'check-fixed-list-headers.mjs'],
  ['check:search-input-style', 'check-search-input-style.mjs'],
]) {
  requireScript(frontendPkg, scriptName, snippet, 'frontend')
}

for (const snippet of [
  'node scripts/check-quality-gates.mjs',
  'go test ./backend/...',
  'npm run check:all',
]) {
  if (!verifyLocal.includes(snippet)) {
    throw new Error(`scripts/verify-local.sh must include: ${snippet}`)
  }
}

for (const snippet of [
  'bash scripts/verify-local.sh',
  '管理后台',
  '客户端',
  '后端',
]) {
  if (!maintenanceDoc.includes(snippet)) {
    throw new Error(`docs/project-maintenance.md must mention: ${snippet}`)
  }
}
