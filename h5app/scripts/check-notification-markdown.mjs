import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { resolve } from 'node:path'
import process from 'node:process'
import ts from 'typescript'

const root = process.cwd()
const require = createRequire(import.meta.url)
const markdownModulePath = 'src/pages/notifications/notification-markdown.ts'
const markdownComponentPath = 'src/pages/notifications/components/NotificationMarkdown.vue'

for (const file of [markdownModulePath, markdownComponentPath]) {
  assert.equal(existsSync(resolve(root, file)), true, `missing notification Markdown file: ${file}`)
}

function source(file) {
  return readFileSync(resolve(root, file), 'utf8')
}

const markdownSource = source(markdownModulePath)
assert.equal(markdownSource.includes('new MarkdownIt'), true, 'notification Markdown must use markdown-it')
assert.equal(markdownSource.includes('linkify: true'), true, 'notification Markdown must autolink URLs')

const componentSource = source(markdownComponentPath)
assert.equal(componentSource.includes('import DOMPurify from \'dompurify\''), true, 'notification HTML must use DOMPurify')
assert.equal(componentSource.includes('purifier.sanitize'), true, 'notification HTML must be sanitized before rendering')
assert.equal(componentSource.includes('USE_PROFILES: { html: true }'), true, 'notification sanitizer must only use the HTML profile')
assert.equal(componentSource.includes('FORBID_ATTR: [\'style\']'), true, 'notification sanitizer must reject inline styles')
assert.equal(componentSource.includes('<rich-text'), true, 'notification Markdown must render through uni-app rich-text')

for (const file of [
  'src/components/app-notification-panel/app-notification-panel.vue',
  'src/pages/notifications/components/NotificationHistoryPage.vue',
]) {
  const content = source(file)
  assert.equal(content.includes('import NotificationMarkdown from \'@/pages/notifications/components/NotificationMarkdown.vue\''), true, `${file} must import NotificationMarkdown`)
  assert.equal(content.includes('<NotificationMarkdown'), true, `${file} must render notification detail with NotificationMarkdown`)
}

const output = ts.transpileModule(markdownSource, {
  compilerOptions: {
    module: ts.ModuleKind.CommonJS,
    target: ts.ScriptTarget.ES2020,
    esModuleInterop: true,
  },
  fileName: markdownModulePath,
}).outputText
const module = { exports: {} }
// eslint-disable-next-line no-new-func
new Function('module', 'exports', 'require', output)(module, module.exports, require)

const { renderNotificationMarkdownSource, renderNotificationMarkdownWithoutHtml } = module.exports
const rendered = renderNotificationMarkdownSource('# 标题\n\n- 第一项\n- 第二项\n\n**重点**\n\n<strong>HTML</strong>')
assert.match(rendered, /<h1>标题<\/h1>/, 'notification Markdown must render headings')
assert.match(rendered, /<ul>[\s\S]*<li>第一项<\/li>/, 'notification Markdown must render lists')
assert.match(rendered, /<strong>重点<\/strong>/, 'notification Markdown must render emphasis')
assert.match(rendered, /<strong>HTML<\/strong>/, 'notification Markdown must preserve basic HTML for sanitization')
assert.doesNotMatch(renderNotificationMarkdownSource('[危险链接](javascript:alert(1))'), /href=/, 'notification Markdown must reject dangerous Markdown links')
assert.doesNotMatch(renderNotificationMarkdownWithoutHtml('<script>alert(1)</script>'), /<script>/, 'non-H5 fallback must escape raw HTML')

console.log('notification Markdown checks passed')
