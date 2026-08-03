import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

const currentDir = dirname(fileURLToPath(import.meta.url))
const srcDir = resolve(currentDir, '../src')
const mainPath = resolve(srcDir, 'main.ts')
const iconsPath = resolve(srcDir, 'icons.ts')
const layoutPath = resolve(srcDir, 'views/layout/index.vue')
const menuViewPath = resolve(srcDir, 'views/menu/index.vue')
const iconPickerPath = resolve(srcDir, 'components/IconPicker.vue')
const adminRoutesPath = resolve(srcDir, 'router/adminRoutes.ts')

if (!existsSync(iconsPath)) {
  throw new Error('admin icon registry missing: src/icons.ts')
}

const mainSource = readFileSync(mainPath, 'utf8')
const iconsSource = readFileSync(iconsPath, 'utf8')
const layoutSource = readFileSync(layoutPath, 'utf8')
const menuViewSource = readFileSync(menuViewPath, 'utf8')
const iconPickerSource = readFileSync(iconPickerPath, 'utf8')
const adminRoutesSource = readFileSync(adminRoutesPath, 'utf8')

for (const forbidden of [
  'import * as ElementPlusIconsVue',
  'Object.entries(ElementPlusIconsVue)'
]) {
  if (mainSource.includes(forbidden)) {
    throw new Error(`admin main.ts still imports all icons: ${forbidden}`)
  }
}

for (const snippet of [
  "import { registerAdminIcons } from './icons'",
  'registerAdminIcons(app)'
]) {
  if (!mainSource.includes(snippet)) {
    throw new Error(`admin main.ts missing icon registry usage: ${snippet}`)
  }
}

for (const snippet of [
  'export const adminIconMap',
  'export const ADMIN_ICON_NAMES',
  'export const DEFAULT_ADMIN_ICON_NAME',
  'export function resolveAdminIcon',
  'export function registerAdminIcons',
  "import * as ElementPlusIconsVue from '@element-plus/icons-vue'",
  "Object.entries(ElementPlusIconsVue).filter(([name]) => name !== 'default')"
]) {
  if (!iconsSource.includes(snippet)) {
    throw new Error(`admin icon registry missing: ${snippet}`)
  }
}

for (const snippet of [
  "import { resolveAdminIcon } from '../../icons'",
  'resolveAdminIcon(row.icon)'
]) {
  if (!menuViewSource.includes(snippet)) {
    throw new Error(`admin menu view missing resolved icon usage: ${snippet}`)
  }
}

for (const snippet of [
  "import { ADMIN_ICON_NAMES, resolveAdminIcon } from '../icons'",
  'icons.value = ADMIN_ICON_NAMES',
  'resolveAdminIcon(props.modelValue)'
]) {
  if (!iconPickerSource.includes(snippet)) {
    throw new Error(`admin icon picker must use controlled registry: ${snippet}`)
  }
}

if (iconPickerSource.includes("import * as ElIcons from '@element-plus/icons-vue'")) {
  throw new Error('admin icon picker should reuse src/icons.ts instead of importing Element Plus icons again')
}

const routeIconMatches = adminRoutesSource.matchAll(/icon:\s*['"]([^'"]+)['"]/g)
for (const match of routeIconMatches) {
  const iconName = match[1]
  if (!Object.prototype.hasOwnProperty.call(ElementPlusIconsVue, iconName)) {
    throw new Error(`admin route icon does not exist in Element Plus: ${iconName}`)
  }
}

for (const snippet of [
  "import { resolveAdminIcon } from '../../icons'",
  'resolveAdminIcon(item.icon)',
  "resolveAdminIcon('Expand')",
  "resolveAdminIcon('Fold')"
]) {
  if (!layoutSource.includes(snippet)) {
    throw new Error(`admin layout missing resolved icon usage: ${snippet}`)
  }
}
