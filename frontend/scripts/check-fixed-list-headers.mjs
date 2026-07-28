import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))

const searchAndTabPages = [
  ['打卡任务', '../pages/enroll/enroll_index.vue'],
  ['赛事活动', '../pages/event/event_index.vue'],
  ['赛事管理', '../pages/event/my_event_manage.vue'],
]

const tabOnlyPages = [
  ['我的打卡', '../pages/enroll/my_user_list.vue'],
  ['我的活动入口', '../pages/my/my_activity.vue'],
  ['我的赛事入口', '../pages/my/my_competition.vue'],
  ['我的活动', '../pages/event/my_event_activity.vue'],
  ['我的赛事', '../pages/event/my_event_competition.vue'],
]

function assertNoManualOffset(label, page) {
  for (const forbidden of [
    'fixedTop',
    'containerPad',
    'uni.getSystemInfoSync()',
    ':style="{ paddingTop:',
    ':style="{ top:',
  ]) {
    if (page.includes(forbidden)) {
      throw new Error(`${label} fixed list header must not use manual nav offsets: ${forbidden}`)
    }
  }
}

function assertRequired(label, page, snippets) {
  for (const required of snippets) {
    if (!page.includes(required)) {
      throw new Error(`${label} fixed list header missing style: ${required}`)
    }
  }
}

for (const [label, pagePath] of searchAndTabPages) {
  const page = readFileSync(resolve(currentDir, pagePath), 'utf8')
  assertNoManualOffset(label, page)
  assertRequired(label, page, [
    'class="header-sticky"',
    'class="search-bar"',
    'class="tabs"',
    'position: fixed',
    'left: 0',
    'right: 0',
    'top: var(--window-top, 0px)',
    'z-index: 20',
    'padding-top: 192rpx',
  ])
}

for (const [label, pagePath] of tabOnlyPages) {
  const page = readFileSync(resolve(currentDir, pagePath), 'utf8')
  assertNoManualOffset(label, page)
  assertRequired(label, page, [
    'class="tabs"',
    'position: fixed',
    'left: 0',
    'right: 0',
    'top: var(--window-top, 0px)',
    'z-index: 20',
    'padding-top: 92rpx',
  ])
}
