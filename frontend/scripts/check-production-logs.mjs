import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs'
import { dirname, extname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const root = resolve(currentDir, '..')
const scanTargets = [
  resolve(root, 'App.vue'),
  resolve(root, 'pages'),
  resolve(root, 'components'),
  resolve(root, 'utils'),
  resolve(root, 'config'),
]
const sourceExts = new Set(['.vue', '.js', '.ts'])
const debugConsolePattern = /\bconsole\.(log|debug|info)\s*\(/g

function collectFiles(path) {
  if (!existsSync(path)) return []
  const stat = statSync(path)
  if (stat.isFile()) {
    return sourceExts.has(extname(path)) ? [path] : []
  }
  return readdirSync(path)
    .flatMap((entry) => collectFiles(resolve(path, entry)))
}

const violations = []
for (const file of scanTargets.flatMap(collectFiles)) {
  const source = readFileSync(file, 'utf8')
  const lines = source.split(/\r?\n/)
  lines.forEach((line, index) => {
    if (debugConsolePattern.test(line)) {
      violations.push(`${file.replace(root + '/', '')}:${index + 1}: ${line.trim()}`)
    }
    debugConsolePattern.lastIndex = 0
  })
}

if (violations.length > 0) {
  throw new Error(`frontend production debug logs found:\n${violations.join('\n')}`)
}
