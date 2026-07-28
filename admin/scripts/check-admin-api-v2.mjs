import { readdirSync, readFileSync, statSync } from 'node:fs'
import { dirname, extname, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const srcDir = resolve(currentDir, '../src')

const files = []

function collect(dir) {
  for (const name of readdirSync(dir)) {
    const path = join(dir, name)
    const stat = statSync(path)
    if (stat.isDirectory()) {
      collect(path)
      continue
    }
    if (['.ts', '.vue'].includes(extname(path))) {
      files.push(path)
    }
  }
}

collect(srcDir)

const patterns = [
  /request\.(?:get|post|put|delete|patch)\(\s*['"`]\/admin\//,
  /request\.(?:get|post|put|delete|patch)\(\s*['"`]\/home\//,
  /request\.(?:get|post|put|delete|patch)\(\s*['"`]\/user_form_fields['"`]/,
  /(?:action|:action)=["{`]+\/admin\//,
  /fetch\(\s*(?:baseUrl\s*\+\s*)?[`'"]\/admin\//
]

const offenders = []
for (const file of files) {
  const source = readFileSync(file, 'utf8')
  const lines = source.split(/\r?\n/)
  for (let i = 0; i < lines.length; i++) {
    if (patterns.some(pattern => pattern.test(lines[i]))) {
      offenders.push(`${relative(resolve(currentDir, '..'), file)}:${i + 1}: ${lines[i].trim()}`)
    }
  }
}

if (offenders.length > 0) {
  const shown = offenders.slice(0, 80)
  const more = offenders.length > shown.length ? `\n...and ${offenders.length - shown.length} more` : ''
  throw new Error(`admin frontend must use /api/v2 admin endpoints (${offenders.length} offenders):\n${shown.join('\n')}${more}`)
}
