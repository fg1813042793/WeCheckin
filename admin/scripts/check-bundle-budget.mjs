import { existsSync, readdirSync, statSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const assetsDir = resolve(currentDir, '../dist/assets')

if (!existsSync(assetsDir)) {
  throw new Error('admin dist/assets not found. Run npm run build before npm run check:bundle.')
}

const kib = 1024
const budgets = [
  { pattern: /^vendor-element-plus-.*\.js$/, max: 1024 * kib },
  { pattern: /^vendor-element-plus-.*\.css$/, max: 380 * kib },
  { pattern: /^vendor-.*\.js$/, max: 560 * kib },
  { pattern: /^vendor-echarts-.*\.js$/, max: 430 * kib },
  { pattern: /^vendor-qrcode-map-.*\.js$/, max: 520 * kib },
  { pattern: /^vendor-editor-.*\.js$/, max: 160 * kib },
  { pattern: /^SurveyDesigner-.*\.js$/, max: 190 * kib },
  { pattern: /^ExamDesigner-.*\.js$/, max: 170 * kib },
  { pattern: /^QuestionBank-.*\.js$/, max: 60 * kib },
]

const files = readdirSync(assetsDir)
const violations = []

for (const file of files) {
  const budget = budgets.find((item) => item.pattern.test(file))
  if (!budget) continue

  const size = statSync(resolve(assetsDir, file)).size
  if (size > budget.max) {
    violations.push(`${file}: ${(size / kib).toFixed(1)} KiB > ${(budget.max / kib).toFixed(1)} KiB`)
  }
}

for (const budget of budgets) {
  if (!files.some((file) => budget.pattern.test(file))) {
    violations.push(`missing expected bundle matching ${budget.pattern}`)
  }
}

if (violations.length > 0) {
  throw new Error(`admin bundle budget exceeded:\n${violations.join('\n')}`)
}

