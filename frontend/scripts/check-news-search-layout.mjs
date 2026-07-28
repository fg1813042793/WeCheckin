import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const page = readFileSync(resolve(currentDir, '../pages/news/news_index.vue'), 'utf8')

for (const forbidden of [
  'fixedTop',
  'containerPad',
  'uni.getSystemInfoSync()',
  'class="search-btn"',
  'top: 0',
]) {
  if (page.includes(forbidden)) {
    throw new Error(`news search layout must use fixed search below native nav: ${forbidden}`)
  }
}

for (const required of [
  'class="search-shell"',
  'position: fixed',
  'left: 0',
  'right: 0',
  'top: var(--window-top, 0px)',
  'z-index: 20',
  'padding: 20rpx 30rpx',
  'background-color: #fff',
  'padding-top: 110rpx',
  'height: 70rpx',
  'background-color: #f5f5f5',
  'border-radius: 35rpx',
  'padding: 0 24rpx',
]) {
  if (!page.includes(required)) {
    throw new Error(`news search layout missing fixed search below native nav style: ${required}`)
  }
}
