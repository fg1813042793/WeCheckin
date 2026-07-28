import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const srcDir = resolve(currentDir, '../src')

function read(path) {
  return readFileSync(resolve(srcDir, path), 'utf8')
}

function requireSnippet(source, snippet, label) {
  if (!source.includes(snippet)) {
    throw new Error(`${label} missing required snippet: ${snippet}`)
  }
}

function forbidSnippet(source, snippet, label) {
  if (source.includes(snippet)) {
    throw new Error(`${label} still contains loose type snippet: ${snippet}`)
  }
}

const sharedTypesPath = resolve(srcDir, 'views/formkit/shared/types.ts')
if (!existsSync(sharedTypesPath)) {
  throw new Error('missing shared formkit type file: views/formkit/shared/types.ts')
}

for (const [path, label] of [
  ['views/survey/formkit/QuestionPreview.vue', 'survey question preview'],
  ['views/exam/formkit/QuestionPreview.vue', 'exam question preview'],
]) {
  const source = read(path)
  requireSnippet(source, "import type { FormKitQuestion } from '../../formkit/shared/types'", label)
  requireSnippet(source, 'defineProps<{ q: FormKitQuestion; editing?: boolean }>()', label)
  forbidSnippet(source, 'defineProps<{ q: any; editing?: boolean }>()', label)
}

