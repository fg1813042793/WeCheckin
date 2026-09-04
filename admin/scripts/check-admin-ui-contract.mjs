import { existsSync, readFileSync, readdirSync, statSync } from 'node:fs'
import { dirname, extname, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import ts from 'typescript'

const currentDir = dirname(fileURLToPath(import.meta.url))
const adminDir = resolve(currentDir, '..')
const srcDir = resolve(adminDir, 'src')
const routerDir = resolve(srcDir, 'router')
const baselinePath = resolve(currentDir, 'admin-ui-legacy-routes.json')
const allowedPatterns = new Set(['list', 'filter-list', 'form', 'detail', 'workspace'])

const requiredFoundationFiles = [
  'docs/ui-guidelines.md',
  'src/styles/admin-ui-tokens.css',
  'src/components/admin-ui/index.ts',
  'src/components/admin-ui/AdminPageShell.vue',
  'src/components/admin-ui/AdminPageHeader.vue',
  'src/components/admin-ui/AdminSearchBar.vue',
  'src/components/admin-ui/AdminTablePanel.vue',
  'src/components/admin-ui/AdminDialog.vue',
  'src/components/admin-ui/AdminDrawer.vue',
  'src/examples/admin-ui/SimpleListExample.vue',
  'src/examples/admin-ui/AdvancedFilterListExample.vue',
]

for (const file of requiredFoundationFiles) {
  if (!existsSync(resolve(adminDir, file))) {
    throw new Error(`admin UI foundation missing: ${file}`)
  }
}

if (!existsSync(baselinePath)) {
  throw new Error('admin UI legacy route baseline missing: scripts/admin-ui-legacy-routes.json')
}

const baseline = JSON.parse(readFileSync(baselinePath, 'utf8'))
if (baseline.version !== 1 || !Array.isArray(baseline.routes)) {
  throw new Error('admin UI legacy route baseline must use version 1 with a routes array')
}

function routeKey(route) {
  return JSON.stringify([route.name || '', route.path])
}

const legacyRouteKeys = new Set()
for (const route of baseline.routes) {
  if (typeof route?.name !== 'string' || typeof route?.path !== 'string') {
    throw new Error('admin UI legacy route baseline contains an invalid route entry')
  }
  const key = routeKey(route)
  if (legacyRouteKeys.has(key)) {
    throw new Error(`admin UI legacy route baseline contains duplicate route: ${key}`)
  }
  legacyRouteKeys.add(key)
}

function walkFiles(directory) {
  const files = []
  for (const entry of readdirSync(directory)) {
    const absolutePath = resolve(directory, entry)
    if (statSync(absolutePath).isDirectory()) {
      files.push(...walkFiles(absolutePath))
    } else if (extname(entry) === '.ts') {
      files.push(absolutePath)
    }
  }
  return files
}

function propertyName(node, sourceFile) {
  if (!node) return ''
  if (ts.isIdentifier(node) || ts.isStringLiteralLike(node)) return node.text
  return node.getText(sourceFile)
}

function property(object, name, sourceFile) {
  return object.properties.find((item) => {
    return ts.isPropertyAssignment(item) && propertyName(item.name, sourceFile) === name
  })
}

function stringValue(node) {
  return node && ts.isStringLiteralLike(node) ? node.text : undefined
}

function numberValue(node) {
  return node && ts.isNumericLiteral(node) ? Number(node.text) : undefined
}

function componentImport(node) {
  let value
  const visit = (child) => {
    if (
      ts.isCallExpression(child)
      && child.expression.kind === ts.SyntaxKind.ImportKeyword
      && child.arguments.length === 1
    ) {
      value = stringValue(child.arguments[0])
      return
    }
    ts.forEachChild(child, visit)
  }
  visit(node)
  return value
}

function adminUiContract(metaNode, sourceFile) {
  if (!metaNode || !ts.isObjectLiteralExpression(metaNode)) return undefined
  const adminUi = property(metaNode, 'adminUi', sourceFile)
  if (!adminUi || !ts.isObjectLiteralExpression(adminUi.initializer)) return undefined
  return {
    version: numberValue(property(adminUi.initializer, 'version', sourceFile)?.initializer),
    pattern: stringValue(property(adminUi.initializer, 'pattern', sourceFile)?.initializer),
  }
}

function routeComponentPath(routerFile, importPath) {
  if (!importPath?.endsWith('.vue')) return undefined
  if (importPath.startsWith('@/')) return resolve(srcDir, importPath.slice(2))
  return resolve(dirname(routerFile), importPath)
}

function extractRoutes(routerFile) {
  const source = readFileSync(routerFile, 'utf8')
  const sourceFile = ts.createSourceFile(routerFile, source, ts.ScriptTarget.Latest, true)
  const routes = []
  const visit = (node) => {
    if (ts.isObjectLiteralExpression(node)) {
      const componentProp = property(node, 'component', sourceFile)
      if (componentProp) {
        const pathProp = property(node, 'path', sourceFile)
        const path = stringValue(pathProp?.initializer)
        if (path === undefined) {
          throw new Error(`${relative(adminDir, routerFile)} contains a component route without a literal path`)
        }
        const name = stringValue(property(node, 'name', sourceFile)?.initializer) || ''
        const importPath = componentImport(componentProp.initializer)
        routes.push({
          name,
          path,
          source: relative(adminDir, routerFile),
          component: routeComponentPath(routerFile, importPath),
          contract: adminUiContract(property(node, 'meta', sourceFile)?.initializer, sourceFile),
        })
      }
    }
    ts.forEachChild(node, visit)
  }
  visit(sourceFile)
  return routes
}

const routes = walkFiles(routerDir).flatMap(extractRoutes)
const currentRouteKeys = new Set()
for (const route of routes) {
  const key = routeKey(route)
  if (currentRouteKeys.has(key)) {
    throw new Error(`duplicate admin route name/path pair: ${key}`)
  }
  currentRouteKeys.add(key)
}

const newRoutes = routes.filter((route) => !legacyRouteKeys.has(routeKey(route)))
const errors = []
for (const route of newRoutes) {
  const label = route.name || route.path
  if (route.contract?.version !== 1 || !allowedPatterns.has(route.contract?.pattern)) {
    errors.push(`${label}: meta.adminUi must declare version 1 and one of ${[...allowedPatterns].join(', ')}`)
  }
  if (!route.component || !existsSync(route.component)) {
    errors.push(`${label}: route component must be a lazy-loaded local Vue SFC`)
    continue
  }
  const componentSource = readFileSync(route.component, 'utf8')
  const importsFoundation = /from\s+['"]@\/components\/admin-ui(?:\/index)?['"]/.test(componentSource)
  const usesPageShell = /<AdminPageShell(?:\s|>)/.test(componentSource)
  if (!importsFoundation || !usesPageShell) {
    errors.push(`${label}: ${relative(adminDir, route.component)} must import and render AdminPageShell from @/components/admin-ui`)
  }
}

const routedComponents = new Set(routes.map((route) => route.component).filter(Boolean))
for (const example of [
  'src/examples/admin-ui/SimpleListExample.vue',
  'src/examples/admin-ui/AdvancedFilterListExample.vue',
]) {
  const absolutePath = resolve(adminDir, example)
  if (routedComponents.has(absolutePath)) {
    errors.push(`${example} must not be registered as a route`)
  }
  const source = readFileSync(absolutePath, 'utf8')
  if (/@\/api|utils\/request|\baxios\b|\bfetch\s*\(/.test(source)) {
    errors.push(`${example} must remain static and must not request APIs`)
  }
}

if (errors.length > 0) {
  throw new Error(`admin UI contract check failed:\n- ${errors.join('\n- ')}`)
}

console.log(`Admin UI contract passed: ${legacyRouteKeys.size} legacy routes exempted, ${newRoutes.length} new routes checked.`)
