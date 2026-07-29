import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const requestSource = readFileSync(resolve(currentDir, '../utils/request.js'), 'utf8')
const packageSource = readFileSync(resolve(currentDir, '../package.json'), 'utf8')

for (const snippet of [
  'const inflightRequests = new Map()',
  'function buildRequestKey',
  "method === 'GET'",
  'inflightRequests.has(requestKey)',
  'inflightRequests.delete(requestKey)',
]) {
  if (!requestSource.includes(snippet)) {
    throw new Error(`frontend request layer must dedupe in-flight GET requests with: ${snippet}`)
  }
}

for (const snippet of [
  '"check:request-dedupe"',
  'check-request-dedupe.mjs',
]) {
  if (!packageSource.includes(snippet)) {
    throw new Error(`frontend package scripts must include request dedupe check: ${snippet}`)
  }
}
