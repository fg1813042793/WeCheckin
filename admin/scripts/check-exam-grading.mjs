import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import ts from 'typescript'

const require = createRequire(import.meta.url)
const modelUrl = new URL('../src/views/exam/components/examGradingModel.ts', import.meta.url)
assert.ok(existsSync(modelUrl), '人工批改必须提供独立的评分校验与草稿计算')
const compiled = ts.transpileModule(readFileSync(modelUrl, 'utf8'), {
  compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS },
}).outputText
const module = { exports: {} }
new Function('require', 'module', 'exports', compiled)(require, module, module.exports)
const { scoreError, buildSubmission, gradingPreview, retainGradingDraft, gradingImageUrl } = module.exports

const context = {
  record: { id: 8, examId: 5, status: 1, score: 6, totalScore: 20 },
  version: 'revision-1', canGrade: true, passScore: 12,
  questions: [
    { questionId: 1, fullScore: 6, gotScore: 6, needManual: false },
    { questionId: 2, fullScore: 10, gotScore: 0, needManual: true },
    { questionId: 3, fullScore: 4, gotScore: 0, needManual: true },
  ],
}
for (const value of [undefined, null, NaN, Infinity, -1, 2.5, 11, '0', '']) {
  assert.ok(scoreError(value, 10), `必须拒绝空值或非法分数 ${String(value)}`)
}
assert.equal(scoreError(0, 10), '')
assert.equal(scoreError(10, 10), '')
assert.throws(() => buildSubmission(context, {}, ''), /全部|评分/)
assert.throws(() => buildSubmission(context, { 2: 0 }, ''), /全部|评分/)
assert.throws(() => buildSubmission(context, { 2: 11, 3: 0 }, ''), /评分|分数/)
assert.deepEqual(buildSubmission(context, { 1: 999, 2: 0, 3: 3 }, '继续努力'), {
  expectedVersion: 'revision-1', scores: [{ questionId: 2, score: 0 }, { questionId: 3, score: 3 }], comment: '继续努力',
})
assert.deepEqual(gradingPreview(context, {}), { score: 6, remaining: 2 })
assert.deepEqual(gradingPreview(context, { 2: 5, 3: 0 }), { score: 11, remaining: 0 })
for (const patch of [{ record: { ...context.record, status: 0 } }, { record: { ...context.record, status: 2 } }, { canGrade: false }, { version: '' }, { questions: [] }]) {
  assert.throws(() => buildSubmission({ ...context, ...patch }, { 2: 5, 3: 0 }, ''), /批改|版本|题目/)
}
assert.equal(typeof retainGradingDraft, 'function', '重新加载必须按最新题目筛选可保留的评分')
assert.deepEqual(retainGradingDraft(context, { 1: 6, 2: 0, 3: 4, 99: 1 }), { 2: 0, 3: 4 })
assert.deepEqual(retainGradingDraft(context, { 2: 11, 3: -1 }), {})
assert.deepEqual(retainGradingDraft(context, { 2: null, 3: 2.5 }), {})
assert.deepEqual(retainGradingDraft({ ...context, questions: [{ ...context.questions[1], fullScore: 2 }] }, { 2: 3, 3: 1 }), {})

assert.equal(typeof gradingImageUrl, 'function', '题目图片必须校验 URL')
for (const url of ['https://example.test/a.png', 'http://example.test/a.png?q=1#preview', 'HTTPS://example.test/a.png', '/uploads/a.png', '/']) {
  assert.equal(gradingImageUrl(url), url)
}
for (const url of [undefined, null, 42, '', ' ', 'javascript:alert(1)', 'data:image/png;base64,aaa', 'blob:https://example.test/id', 'file:///tmp/a', '//example.test/a', '/\\example.test/a', '\\example.test/a', 'uploads/a.png', 'https:example.test/a', 'http://', 'https:///a.png', 'https://example.test/\na.png', '/\texample.test/a', 'https://example.test/\\a.png']) {
  assert.equal(gradingImageUrl(url), '', `不得加载不安全图片地址 ${JSON.stringify(url)}`)
}
console.log('exam grading score validation, draft recovery, image URLs and payload checks passed')
