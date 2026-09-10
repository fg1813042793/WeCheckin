import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

function read(path) {
  return readFileSync(new URL(path, import.meta.url), 'utf8')
}

const feedbackCenter = read('../src/pages/feedback/components/FeedbackCenter.vue')
const zhCN = read('../src/locale/lang/zh-CN.json')
const enUS = read('../src/locale/lang/en-US.json')

for (const snippet of [
  'const mobilePresentation = ref(resolveMobilePresentation())',
  'const filtersExpanded = ref(!mobilePresentation.value)',
  'const filtersVisible = computed(() => !mobilePresentation.value || filtersExpanded.value)',
  'const activeFilterCount = computed',
  'v-if="mobilePresentation"',
  ':aria-expanded="filtersExpanded"',
  'v-show="filtersVisible"',
  'feedback-center__filter-toggle',
  't(\'filterPanel.title\')',
  't(\'filterPanel.activeCount\', { count: activeFilterCount })',
]) {
  assert.ok(feedbackCenter.includes(snippet), `FeedbackCenter.vue missing ${snippet}`)
}

assert.match(
  feedbackCenter,
  /@media screen and \(max-width: 768px\)[\s\S]*?\.feedback-center \.app-page-header__copy\s*\{[^}]*display:\s*none;/,
  'feedback center must hide only its mobile page header copy',
)

for (const locale of [zhCN, enUS]) {
  assert.match(locale, /"filterPanel"\s*:\s*\{[\s\S]*?"title"\s*:/)
  assert.match(locale, /"filterPanel"\s*:\s*\{[\s\S]*?"activeCount"\s*:/)
}

console.log('Feedback mobile layout checks passed')
