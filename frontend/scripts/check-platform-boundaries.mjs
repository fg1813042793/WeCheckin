import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const root = resolve(currentDir, '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

const app = read('App.vue')
assert.ok(app.includes("from './utils/auth'"), 'app startup must import client auth helpers')
assert.ok(app.includes('hasClientAuth'), 'app startup must check client auth')
assert.ok(!app.includes('hasAnyAuth'), 'admin auth must not satisfy client startup auth')

for (const page of ['pages/survey/fill.vue', 'pages/exam/fill.vue']) {
  assert.ok(!read(page).includes('window.location'), `${page} must use the shared cross-platform navigation helper`)
}

const scoreEntry = read('pages/event/my_event_score_entry.vue')
for (const required of ['// #ifdef H5', '// #ifndef H5', '当前环境不支持导入']) {
  assert.ok(scoreEntry.includes(required), `score CSV behavior missing platform boundary: ${required}`)
}

const manifest = read('manifest.json')
for (const forbidden of [
  'CHANGE_NETWORK_STATE',
  'MOUNT_UNMOUNT_FILESYSTEMS',
  'READ_LOGS',
  'GET_ACCOUNTS',
  'READ_PHONE_STATE',
  'CHANGE_WIFI_STATE',
  'WAKE_LOCK',
  'WRITE_SETTINGS',
]) {
  assert.ok(!manifest.includes(forbidden), `Android manifest contains unnecessary permission: ${forbidden}`)
}
