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
  'vendor-echarts',
  'vendor-editor',
  'vendor-qrcode-map'
]

for (const snippet of requiredSnippets) {
  if (!source.includes(snippet)) {
    throw new Error(`admin Vite build config missing: ${snippet}`)
  }
}
