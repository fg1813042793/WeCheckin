import assert from 'node:assert/strict'
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { dirname, extname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const root = resolve(currentDir, '..')
const uploadPath = resolve(root, 'utils/upload.js')

function collectSourceFiles(path) {
  const stat = statSync(path)
  if (stat.isFile()) return ['.js', '.vue'].includes(extname(path)) ? [path] : []
  return readdirSync(path).flatMap((entry) => collectSourceFiles(resolve(path, entry)))
}

for (const file of [resolve(root, 'components'), resolve(root, 'pages')].flatMap(collectSourceFiles)) {
  const source = readFileSync(file, 'utf8')
  assert.ok(!source.includes('uni.uploadFile('), `${file} must use utils/upload.js`)
}

const uploadSource = readFileSync(uploadPath, 'utf8')
for (const required of [
  "path: '/api/v2/uploads'",
  "path: '/api/v2/admin/uploads'",
  "path: '/upload'",
  'getClientToken()',
  'getAdminToken()',
]) {
  assert.ok(uploadSource.includes(required), `upload wrapper missing: ${required}`)
}
