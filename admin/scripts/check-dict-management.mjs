import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const root = resolve(currentDir, '..')
const read = path => readFileSync(resolve(root, path), 'utf8')

const page = read('src/views/dict/index.vue')
for (const snippet of [
  'class="dict-workspace"',
  'selectedType',
  'itemLoading',
  'dictTypeAdd',
  'dictTypeDelete',
  'dictTypeClearItems',
  'itemForm.status',
  ':disabled="!typeDialog.isCreate"',
]) {
  if (!page.includes(snippet)) throw new Error(`dictionary management page missing ${snippet}`)
}
if (/dictAdd\(\{[^}]*value:\s*['"]{2}/s.test(page)) {
  throw new Error('dictionary types must not be represented by empty-value dictionary items')
}

const api = read('src/api/index.ts')
for (const snippet of ['dictTypeAdd', 'dictTypeDelete', 'dictTypeClearItems', 'dictActiveItems']) {
  if (!api.includes(snippet)) throw new Error(`dictionary API missing ${snippet}`)
}

const types = read('src/api/types.ts')
for (const snippet of ['interface DictTypeSummary', 'interface DictItem', 'interface DictTypePayload', 'interface DictItemPayload']) {
  if (!types.includes(snippet)) throw new Error(`dictionary API types missing ${snippet}`)
}

for (const path of ['src/views/news/index.vue', 'src/views/enroll/index.vue', 'src/views/event/index.vue']) {
  const source = read(path)
  if (!source.includes('dictActiveItems(')) throw new Error(`${path} must load enabled public dictionary items`)
}
