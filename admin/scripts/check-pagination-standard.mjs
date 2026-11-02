import { readdirSync, readFileSync } from 'node:fs'
import { join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = resolve(fileURLToPath(new URL('..', import.meta.url)))
const viewsRoot = resolve(root, 'src/views')
const compactExceptions = new Set([
  'src/views/notification/components/TemplateWorkflowSelect.vue',
  'src/views/question-bank/components/QuestionBankNavigator.vue',
])

function vueFiles(directory) {
  const result = []
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) result.push(...vueFiles(path))
    else if (entry.isFile() && entry.name.endsWith('.vue')) result.push(path)
  }
  return result
}

const errors = []
let count = 0
for (const path of vueFiles(viewsRoot)) {
  const source = readFileSync(path, 'utf8')
  const file = relative(root, path)
  const tags = [...source.matchAll(/<el-pagination\b[\s\S]*?\/>/g)]
  for (const [index, match] of tags.entries()) {
    count++
    const tag = match[0]
    const label = `${file} pagination #${index + 1}`
    if (/\sbackground(?:\s|=|\/>)/.test(tag)) errors.push(`${label} must not use background pagination`)
    if (compactExceptions.has(file)) continue
    const layout = tag.match(/\blayout="([^"]+)"/)?.[1].replaceAll(' ', '')
    if (layout !== 'total,sizes,prev,pager,next') errors.push(`${label} must use total,sizes,prev,pager,next`)
    if (!/:page-sizes=/.test(tag)) errors.push(`${label} must expose page-size options`)
    if (/\bsmall\b|\bsize="small"/.test(tag)) errors.push(`${label} must use the standard pagination size`)
    if (/:pager-count="5"/.test(tag)) errors.push(`${label} must use the standard seven-item pager`)
  }
}

const adminCSS = readFileSync(resolve(root, 'src/styles/admin.css'), 'utf8')
const paginationBlock = adminCSS.match(/\.admin-pagination\s*\{([^}]*)\}/)?.[1] || ''
if (!paginationBlock.includes('justify-content: flex-end')) errors.push('global .admin-pagination must align pagination to the right')
const elementPaginationBlock = adminCSS.match(/\.admin-main \.el-pagination\s*\{([^}]*)\}/)?.[1] || ''
if (!elementPaginationBlock.includes('margin-left: auto')) errors.push('all admin pagination components must use right auto alignment')
if (!elementPaginationBlock.includes('--el-pagination-bg-color: transparent')) errors.push('non-active pagination buttons must use a transparent background')
if (!elementPaginationBlock.includes('--el-pagination-button-disabled-bg-color: transparent')) errors.push('disabled pagination buttons must use a transparent background')

if (errors.length) throw new Error(`pagination standard failed (${errors.length}/${count}):\n${errors.join('\n')}`)
console.log(`Pagination standard passed for ${count} pagination components`)
