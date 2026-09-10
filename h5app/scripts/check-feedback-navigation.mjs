import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

function read(path) {
  return readFileSync(new URL(path, import.meta.url), 'utf8')
}

const store = read('../src/stores/appContent.ts')
const shell = read('../src/components/app-shell/app-shell.vue')
const createPage = read('../src/pages/feedback/components/FeedbackCreatePage.vue')

for (const snippet of [
  'const closeRequestTargetKey = ref(\'\')',
  'function requestCloseTab(key: string, targetKey = \'\')',
  'closeRequestTargetKey.value = String(targetKey || \'\').trim()',
  'closeRequestTargetKey,',
]) {
  assert.ok(store.includes(snippet), `appContent store missing ${snippet}`)
}

for (const snippet of [
  'const pendingCloseTargetKey = ref(\'\')',
  'function completeTabClose(item: AppNavItem, targetKey = \'\')',
  'allTabItems.value.find(tab => tab.key === targetKey',
  'appContent.closeRequestTargetKey',
  'completeTabClose(item, targetKey)',
  'pendingCloseTargetKey.value',
]) {
  assert.ok(shell.includes(snippet), `app shell missing ${snippet}`)
}

for (const snippet of [
  'FEEDBACK_CONTENT_KEY',
  'appContent.switchContent(FEEDBACK_CONTENT_KEY)',
  'appContent.removeDynamicTab(props.contentKey)',
  'appContent.requestCloseTab(props.contentKey, FEEDBACK_CONTENT_KEY)',
]) {
  assert.ok(createPage.includes(snippet), `feedback create page missing ${snippet}`)
}

assert.equal(createPage.includes('feedbackDetailContentKey'), false, 'create success must not open a feedback detail tab')
assert.equal(createPage.includes('createFeedbackDynamicTab'), false, 'create success must not create a feedback detail tab')

console.log('Feedback navigation checks passed')
