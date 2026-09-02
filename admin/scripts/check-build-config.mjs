import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const source = readFileSync(resolve(currentDir, '../vite.config.ts'), 'utf8')

const requiredSnippets = [
  'chunkSizeWarningLimit',
  'manualChunks',
  'vendor-vue',
  'vendor-element-plus',
  'vendor-workflow-graph',
  'vendor-echarts',
  'vendor-editor',
  'vendor-qrcode-map'
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
