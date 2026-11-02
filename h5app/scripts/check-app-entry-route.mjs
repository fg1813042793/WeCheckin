import assert from 'node:assert/strict'
import { Buffer } from 'node:buffer'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import ts from 'typescript'

const source = readFileSync(resolve('src/pages/index/app-entry-route.ts'), 'utf8')
const compiled = ts.transpileModule(source, {
  compilerOptions: { module: ts.ModuleKind.ESNext, target: ts.ScriptTarget.ES2022 },
}).outputText
const moduleUrl = `data:text/javascript;base64,${Buffer.from(compiled).toString('base64')}`
const { appEntryRouteQueryFromURL, mergeAppEntryRouteQuery } = await import(moduleUrl)
const indexSource = readFileSync(resolve('src/pages/index/index.vue'), 'utf8')

assert.match(indexSource, /mergeAppEntryRouteQuery/)
assert.match(indexSource, /window\.location\.href/)

assert.deepEqual(
  appEntryRouteQueryFromURL('https://oa.example.com/app?tenant=a&view=workflow%3Astart%3A17#/pages/index/index'),
  { view: 'workflow:start:17' },
)
assert.deepEqual(
  appEntryRouteQueryFromURL('https://oa.example.com/app?tenant=a#/pages/index/index?view=workflow%3Astart%3A18'),
  { view: 'workflow:start:18' },
)
assert.deepEqual(
  mergeAppEntryRouteQuery({ view: 'workflow:start:19', reviewId: '7' }, 'https://oa.example.com/app?view=workflow%3Astart%3A17#/pages/index/index?view=workflow%3Astart%3A18'),
  { view: 'workflow:start:19', reviewId: '7' },
)

console.log('H5App entry route query checks passed')
