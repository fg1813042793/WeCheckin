import { readdirSync, readFileSync, statSync } from 'node:fs'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const pagesDir = resolve(currentDir, '../pages')

function collectVueFiles(dir) {
  const files = []
  for (const item of readdirSync(dir)) {
    const path = join(dir, item)
    const stat = statSync(path)
    if (stat.isDirectory()) {
      files.push(...collectVueFiles(path))
    } else if (path.endsWith('.vue')) {
      files.push(path)
    }
  }
  return files
}

for (const file of collectVueFiles(pagesDir)) {
  const source = readFileSync(file, 'utf8')
  if (!source.includes('placeholder="搜索')) continue

  for (const forbidden of [
    'class="search-btn"',
    "class='search-btn'",
    '>搜索</text>',
    '>搜索</view>',
    '>搜索</button>',
  ]) {
    if (source.includes(forbidden)) {
      throw new Error(`${file} must not render a text search button: ${forbidden}`)
    }
  }

  for (const chunk of source.split('<input').slice(1)) {
    const tag = chunk.split('>')[0] || ''
    if (tag.includes('placeholder="搜索') && !tag.includes('confirm-type="search"')) {
      throw new Error(`${file} search input must use confirm-type="search"`)
    }
  }

  for (const required of [
    'height: 70rpx',
    'border-radius: 35rpx',
  ]) {
    if (!source.includes(required)) {
      throw new Error(`${file} search input missing capsule style: ${required}`)
    }
  }

  if (!source.includes('background-color: #f5f5f5') && !source.includes('background: #f5f5f5')) {
    throw new Error(`${file} search input missing light gray background`)
  }
}

for (const page of [
  '../pages/admin/user/admin_user_list.vue',
  '../pages/admin/news/admin_news_list.vue',
  '../pages/admin/enroll/admin_enroll_list.vue',
  '../pages/admin/event/admin_event_list.vue',
]) {
  const file = resolve(currentDir, page)
  const source = readFileSync(file, 'utf8')
  for (const required of [
    'action-row',
    'flex-direction: column',
    'align-items: flex-start',
    'gap: 16rpx',
    'width: 100%',
    'height: 56rpx',
    'line-height: 56rpx',
    'padding: 0 28rpx',
    'border-radius: 16rpx',
    'background-color: #eaf3ff',
    'color: #1677ff',
    '<text class="add-text">创建</text>',
    'v-if="hasPerm',
  ]) {
    if (!source.includes(required)) {
      throw new Error(`${file} admin toolbar action must use always-visible create text button: ${required}`)
    }
  }

  for (const forbidden of [
    'actionsOpen',
    'container--actions-open',
    'toolbar--expanded',
    'fold-btn',
    'fold-icon',
    '<text class="add-icon">+</text>',
    '+ 新增',
    '+ 增加用户',
    '+ 添加通知',
  ]) {
    if (source.includes(forbidden)) {
      throw new Error(`${file} admin toolbar action must not use folded or icon-only action: ${forbidden}`)
    }
  }
}

{
  const file = resolve(currentDir, '../pages/admin/user/admin_user_list.vue')
  const source = readFileSync(file, 'utf8')
  for (const forbidden of ['#fb454c', '#c62828']) {
    if (source.includes(forbidden)) {
      throw new Error(`${file} user list should avoid saturated red UI color: ${forbidden}`)
    }
  }

  for (const required of [
    'background-color: #eaf3ff',
    'background-color: #f1f5f9',
    'color: #64748b',
    'color: #d97706',
    'background-color: #fff1f2',
    'color: #be123c',
    'background-color: #1677ff',
  ]) {
    if (!source.includes(required)) {
      throw new Error(`${file} user list missing softer visual color: ${required}`)
    }
  }
}

{
  const file = resolve(currentDir, '../pages/admin/event/admin_event_list.vue')
  const source = readFileSync(file, 'utf8')
  for (const required of [
    'class="event-header"',
    '.event-header { position: fixed',
    '.tabs { display: flex',
    '.tab-item.active::after',
  ]) {
    if (!source.includes(required)) {
      throw new Error(`${file} event admin top header should be a unified fixed header: ${required}`)
    }
  }

  for (const forbidden of [
    '.tabs { position: fixed',
    'top: calc(var(--window-top',
  ]) {
    if (source.includes(forbidden)) {
      throw new Error(`${file} event admin tabs should not be a separate fixed layer: ${forbidden}`)
    }
  }
}

{
  const file = resolve(currentDir, '../pages/admin/mgr/admin_mgr_list.vue')
  const source = readFileSync(file, 'utf8')
  for (const required of [
    'placeholder="搜索管理员姓名/手机号"',
    'confirm-type="search"',
    '@confirm="handleSearch"',
    'action-row',
    'flex-direction: column',
    'align-items: flex-start',
    'padding-top: 206rpx',
    'height: 56rpx',
    'line-height: 56rpx',
    'padding: 0 28rpx',
    'border-radius: 16rpx',
    'background-color: #eaf3ff',
    'color: #1677ff',
    '<text class="add-text">创建</text>',
  ]) {
    if (!source.includes(required)) {
      throw new Error(`${file} admin manager list top style is not unified: ${required}`)
    }
  }

  for (const forbidden of ['+ 添加', '#fb454c']) {
    if (source.includes(forbidden)) {
      throw new Error(`${file} admin manager list must not use old red add button: ${forbidden}`)
    }
  }
}
