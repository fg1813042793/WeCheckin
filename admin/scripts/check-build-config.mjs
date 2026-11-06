import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { executeTypeScriptModule } from './lib/typescript-runtime.mjs'

const currentDir = dirname(fileURLToPath(import.meta.url))
const source = readFileSync(resolve(currentDir, '../vite.config.ts'), 'utf8')
const read = path => readFileSync(resolve(currentDir, `../${path}`), 'utf8')

const requiredSnippets = [
  'chunkSizeWarningLimit',
  'manualChunks',
  'vendor-vue',
  'vendor-element-plus',
  'vendor-workflow-graph',
  'vendor-echarts',
  'vendor-editor',
  'vendor-qrcode-map',
  'exam\\/(list|designer|responses|statistic|formkit|templates)'
]

for (const snippet of requiredSnippets) {
  if (!source.includes(snippet)) {
    throw new Error(`admin Vite build config missing: ${snippet}`)
  }
}

const requiredProxyPatterns = [
  {
    name: 'api v2 dev proxy',
    pattern: /['"]\/api['"]\s*:\s*\{[\s\S]*?target:\s*['"]http:\/\/localhost:8083['"]/
  }
]

for (const { name, pattern } of requiredProxyPatterns) {
  if (!pattern.test(source)) {
    throw new Error(`admin Vite dev proxy missing: ${name}`)
  }
}

const mainSource = read('src/main.ts')
assert.match(mainSource, /installVitePreloadErrorRecovery\(\)/, 'admin startup must install stale chunk recovery')

const {
  CHUNK_RELOAD_COOLDOWN_MS,
  installVitePreloadErrorRecovery,
} = executeTypeScriptModule(read('src/utils/chunkLoadRecovery.ts'), 'chunkLoadRecovery.ts')

class FakeTarget {
  listeners = new Map()

  addEventListener(type, listener) {
    this.listeners.set(type, listener)
  }

  removeEventListener(type, listener) {
    if (this.listeners.get(type) === listener)
      this.listeners.delete(type)
  }

  dispatch(type, event) {
    this.listeners.get(type)?.(event)
  }
}

const target = new FakeTarget()
const values = new Map()
const storage = {
  getItem: key => values.get(key) ?? null,
  setItem: (key, value) => values.set(key, value),
  removeItem: key => values.delete(key),
}
let now = 100_000
let reloads = 0
const dispose = installVitePreloadErrorRecovery({
  target,
  storage,
  now: () => now,
  reload: () => { reloads++ },
})

const first = { prevented: false, preventDefault() { this.prevented = true } }
target.dispatch('vite:preloadError', first)
assert.equal(first.prevented, true, 'first stale chunk failure must suppress the broken async render')
assert.equal(reloads, 1, 'first stale chunk failure must reload the current page')

const repeated = { prevented: false, preventDefault() { this.prevented = true } }
target.dispatch('vite:preloadError', repeated)
assert.equal(repeated.prevented, false, 'repeated failure inside the cooldown must remain visible')
assert.equal(reloads, 1, 'repeated failure inside the cooldown must not create a reload loop')

now += CHUNK_RELOAD_COOLDOWN_MS + 1
const later = { prevented: false, preventDefault() { this.prevented = true } }
target.dispatch('vite:preloadError', later)
assert.equal(later.prevented, true)
assert.equal(reloads, 2, 'a later deployment may recover with a fresh single reload')

dispose()
assert.equal(target.listeners.has('vite:preloadError'), false, 'chunk recovery listener must be removable')

for (const configPath of ['nginx/default.conf.template', '../nginx/default.conf.template']) {
  const nginxSource = read(configPath)
  assert.match(nginxSource, /location = \/index\.html\s*\{[\s\S]*?Cache-Control "no-cache, no-store, must-revalidate"/, `${configPath} must not cache the SPA entry document`)
}
