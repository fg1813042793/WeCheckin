import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import path from 'node:path'
import ts from 'typescript'

const sourcePath = path.resolve('src/config/app-icons.ts')
const source = await readFile(sourcePath, 'utf8')
const output = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.ESNext,
    target: ts.ScriptTarget.ES2022,
  },
})

const moduleUrl = `data:text/javascript;base64,${Buffer.from(output.outputText).toString('base64')}`
const { resolveAppMenuIcon } = await import(moduleUrl)

assert.equal(resolveAppMenuIcon('dashboard'), 'grid')
assert.equal(resolveAppMenuIcon('performance'), 'list')
assert.equal(resolveAppMenuIcon('mine'), 'file-text')
assert.equal(resolveAppMenuIcon('history'), 'order')
assert.equal(resolveAppMenuIcon('manager'), 'account-fill')
assert.equal(resolveAppMenuIcon('hrbp'), 'man-add')
assert.equal(resolveAppMenuIcon('summary'), 'order')
assert.equal(resolveAppMenuIcon('org'), 'setting')
assert.equal(resolveAppMenuIcon('template'), 'file-text')
assert.equal(resolveAppMenuIcon('account'), 'account-fill')
assert.equal(resolveAppMenuIcon('custom-icon'), 'custom-icon')
assert.equal(resolveAppMenuIcon('', 'home'), 'home')

console.log('menu icon aliases ok')
