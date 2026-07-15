import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const source = readFileSync(resolve(currentDir, '../components/formkit/FormRender.vue'), 'utf8')
const formkitSource = readFileSync(resolve(currentDir, '../utils/formkit.js'), 'utf8')
const calcPath = resolve(currentDir, '../utils/formkitCalc.js')

const requiredSnippets = [
  'evaluateFrontendRules',
  'applyFormkitCalcValues',
  'logicRules',
  'hiddenIds',
  'requiredIds',
  'calculatedIds',
  'reevaluateRules',
  'applyCalculatedAnswers',
  'extractCalculatedIds',
  'isCalculated(q)',
  'isRequired(q)',
  'this.reevaluateRules()'
]

for (const snippet of requiredSnippets) {
  if (!source.includes(snippet)) {
    throw new Error(`frontend FormRender logic missing: ${snippet}`)
  }
}

if (source.includes('不做 logic 求值')) {
  throw new Error('frontend FormRender still contains the old no-logic placeholder')
}

if (!existsSync(calcPath)) {
  throw new Error('frontend formkit calc runtime missing: frontend/utils/formkitCalc.js')
}

const calcSource = readFileSync(calcPath, 'utf8')
const calcSnippets = [
  'export function evalFormkitExpression',
  'export function applyFormkitCalcValues',
  'buildCalcEnv',
  'normalizeExpression',
  'CONCATENATE',
  'IFS',
  'SUM',
  'AVG'
]

for (const snippet of calcSnippets) {
  if (!calcSource.includes(snippet)) {
    throw new Error(`frontend formkit calc runtime missing: ${snippet}`)
  }
}

if (calcSource.includes('new Function') || calcSource.includes('eval(')) {
  throw new Error('frontend formkit calc runtime must not use eval or Function')
}

if (!formkitSource.includes('props: q.props || {}')) {
  throw new Error('frontend formkit normalizeSchema should preserve question props for calc compatibility')
}
