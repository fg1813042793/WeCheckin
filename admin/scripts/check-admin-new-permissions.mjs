import { readdirSync, readFileSync, statSync } from 'node:fs'
import { dirname, resolve, relative } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const srcDir = resolve(currentDir, '../src')
const hasPermPattern = /hasPerm\(\s*['"`]([^'"`]+)['"`]\s*\)/g

function walk(dir) {
  const files = []
  for (const name of readdirSync(dir)) {
    const path = resolve(dir, name)
    const stat = statSync(path)
    if (stat.isDirectory()) {
      files.push(...walk(path))
      continue
    }
    if (/\.(vue|ts)$/.test(name)) files.push(path)
  }
  return files
}

const legacyUsages = []
for (const file of walk(srcDir)) {
  const source = readFileSync(file, 'utf8')
  for (const match of source.matchAll(hasPermPattern)) {
    const key = match[1]
    if (!key.startsWith('admin:menu:') && !key.startsWith('admin:api:')) {
      legacyUsages.push(`${relative(srcDir, file)} -> hasPerm('${key}')`)
    }
  }
}

if (legacyUsages.length > 0) {
  throw new Error(`admin hasPerm must use unified permission keys:\n${legacyUsages.join('\n')}`)
}

const menuService = readFileSync(resolve(currentDir, '../../backend/internal/service/admin/menu/service.go'), 'utf8')
for (const forbidden of [
  'permissionsupport.AdminPermCodesContext',
  'allAdminPermissionsWithPermCodesContext',
  'permissionPermCodes',
  '`permission_perms` <> \'\'',
]) {
  if (menuService.includes(forbidden)) {
    throw new Error(`admin me/perms still depends on legacy permission code path: ${forbidden}`)
  }
}
