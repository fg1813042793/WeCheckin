import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const { createLatestRequestTracker } = await import('../utils/latestRequest.js')

const tracker = createLatestRequestTracker()
const first = tracker.begin()
assert.equal(tracker.isLatest(first), true)

const second = tracker.begin()
assert.equal(tracker.isLatest(first), false)
assert.equal(tracker.isLatest(second), true)

tracker.invalidate()
assert.equal(tracker.isLatest(second), false)

const currentDir = dirname(fileURLToPath(import.meta.url))
const searchPage = readFileSync(resolve(currentDir, '../pages/search/search.vue'), 'utf8')
for (const required of [
  'const keyword = this.keyword.trim()',
  'enrollApi.getList({ keyword, page: 1, pageSize: 20 })',
  'newsApi.getList({ keyword, page: 1, pageSize: 20 })',
  'if (!this.requestTracker.isLatest(requestID)) return',
]) {
  assert.ok(searchPage.includes(required), `search page must query by keyword and ignore stale responses: ${required}`)
}
