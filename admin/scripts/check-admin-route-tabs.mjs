import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const layoutPath = resolve(currentDir, '../src/views/layout/index.vue')
const loginPath = resolve(currentDir, '../src/views/login/index.vue')
const requestPath = resolve(currentDir, '../src/utils/request.ts')
const sessionPath = resolve(currentDir, '../src/utils/adminSession.ts')

if (!existsSync(layoutPath)) {
  throw new Error('admin layout missing: src/views/layout/index.vue')
}
if (!existsSync(loginPath)) {
  throw new Error('admin login missing: src/views/login/index.vue')
}
if (!existsSync(requestPath)) {
  throw new Error('admin request util missing: src/utils/request.ts')
}
if (!existsSync(sessionPath)) {
  throw new Error('admin session util missing: src/utils/adminSession.ts')
}

const layoutSource = readFileSync(layoutPath, 'utf8')
const loginSource = readFileSync(loginPath, 'utf8')
const requestSource = readFileSync(requestPath, 'utf8')
const sessionSource = readFileSync(sessionPath, 'utf8')

for (const snippet of [
  'class="admin-route-tabs"',
  'class="admin-route-tabs__nav admin-route-tabs__nav--left"',
  'class="admin-route-tabs__nav admin-route-tabs__nav--right"',
  'aria-label="向左翻动页签"',
  'aria-label="向右翻动页签"',
  "scrollRouteTabs('left')",
  "scrollRouteTabs('right')",
  'visitedTabs',
  'routeTabsScrollbarRef',
  'routeTabsOverflowing',
  'routeTabsCanScrollLeft',
  'routeTabsCanScrollRight',
  'addVisitedTab',
  'closeVisitedTab',
  'isAffixTab',
  'getRouteTabsScrollWrap',
  'updateRouteTabsScrollState',
  'scrollActiveRouteTabIntoView',
  'router.push(tab.fullPath)',
]) {
  if (!layoutSource.includes(snippet)) {
    throw new Error(`admin route tabs missing: ${snippet}`)
  }
}

for (const snippet of [
  '.admin-route-tabs',
  '.admin-route-tabs__scroll',
  '.admin-route-tabs__nav',
  '.admin-route-tabs__nav:disabled',
  '.admin-route-tab',
  '.admin-route-tab.is-active',
  '.admin-route-tab__close',
  '.admin-route-tab__dot {',
  'display: none;',
  'border-radius: 8px 8px 0 0',
]) {
  if (!layoutSource.includes(snippet)) {
    throw new Error(`admin route tab style missing: ${snippet}`)
  }
}

for (const forbidden of [
  '.admin-route-tab::before',
  '.admin-route-tab.is-active::before',
]) {
  if (layoutSource.includes(forbidden)) {
    throw new Error(`admin route active tab should not render top line: ${forbidden}`)
  }
}

for (const snippet of [
  "export const ADMIN_ROUTE_TABS_STORAGE_KEY = 'admin_route_tabs'",
  'export function resetAdminRouteTabs()',
  'localStorage.removeItem(ADMIN_ROUTE_TABS_STORAGE_KEY)',
  'export function clearAdminSession()',
]) {
  if (!sessionSource.includes(snippet)) {
    throw new Error(`admin session util should reset route tabs: ${snippet}`)
  }
}

for (const [name, source] of [
  ['layout', layoutSource],
  ['login', loginSource],
  ['request', requestSource],
]) {
  if (!source.includes('clearAdminSession')) {
    throw new Error(`${name} should use shared clearAdminSession when the admin session changes`)
  }
}

if (!layoutSource.includes('ADMIN_ROUTE_TABS_STORAGE_KEY')) {
  throw new Error('layout route tabs should use the shared storage key')
}
if (!loginSource.includes('clearAdminSession()')) {
  throw new Error('login should clear stale route tabs before writing a new admin session')
}
