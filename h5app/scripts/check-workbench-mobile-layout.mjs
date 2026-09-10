import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(
  new URL('../src/pages/performance/components/PerformanceWorkbench.vue', import.meta.url),
  'utf8',
)

assert.match(
  source,
  /\.todo-item\s*\{[^}]*grid-template-columns:\s*48px minmax\(0, 1fr\) auto;/,
  'desktop todo cards must keep their three-column layout',
)

const mobileStyles = source.match(/@media \(max-width: 768px\) \{([\s\S]*)\}\s*<\/style>/)?.[1] || ''

assert.match(
  mobileStyles,
  /\.todo-item\s*\{[^}]*grid-template-columns:\s*72rpx minmax\(0, 1fr\) auto;/,
  'mobile todo cards must keep the action in a dedicated third column',
)
assert.match(
  mobileStyles,
  /\.todo-item__side\s*\{[^}]*grid-column:\s*3;[^}]*grid-row:\s*1;/,
  'mobile todo actions must stay on the first row',
)
assert.match(
  mobileStyles,
  /\.todo-item__action-button,[\s\S]*?:deep\(\.todo-item__action-button\)\s*\{[^}]*white-space:\s*nowrap;/,
  'mobile todo action labels must not wrap',
)

console.log('Workbench mobile todo layout checks passed')
