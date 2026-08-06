import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const userSource = readFileSync(resolve(currentDir, '../src/views/user/index.vue'), 'utf8')
const roleSource = readFileSync(resolve(currentDir, '../src/views/role/index.vue'), 'utf8')

function treeBlock(source, refName) {
  const refIndex = source.indexOf(`ref="${refName}"`)
  if (refIndex < 0) throw new Error(`missing tree ref: ${refName}`)
  const start = source.lastIndexOf('<el-tree', refIndex)
  const end = source.indexOf('/>', refIndex)
  if (start < 0 || end < 0) throw new Error(`invalid tree block: ${refName}`)
  return source.slice(start, end)
}

for (const [source, refs] of [
  [userSource, ['allowDingTalkH5MenuTreeRef', 'denyDingTalkH5MenuTreeRef']],
  [roleSource, ['dingtalkH5MenuTreeRef']],
]) {
  for (const refName of refs) {
    const block = treeBlock(source, refName)
    if (!block.includes('check-strictly')) {
      throw new Error(`DingTalk H5 menu/button tree must be strict: ${refName}`)
    }
  }
}

for (const snippet of [
  'const allowDingTalkH5MenuCheckedKeys = computed(() => allowDingTalkH5MenuKeys.value)',
  'const denyDingTalkH5MenuCheckedKeys = computed(() => denyDingTalkH5MenuKeys.value)',
  '...checkedKeys(allowDingTalkH5MenuTreeRef, { prefixes: dingtalkH5MenuButtonPrefixes })',
  '...checkedKeys(denyDingTalkH5MenuTreeRef, { prefixes: dingtalkH5MenuButtonPrefixes })',
]) {
  if (!userSource.includes(snippet)) {
    throw new Error(`user extra DingTalk H5 menu/button permission tree is not independent: ${snippet}`)
  }
}

for (const snippet of [
  'const dingtalkH5MenuCheckedKeys = computed(() => form.dingtalkH5MenuKeys)',
  'form.dingtalkH5MenuKeys = checked.filter((key: string) => dingtalkH5MenuButtonPrefixes.some((prefix) => key.startsWith(prefix)))',
]) {
  if (!roleSource.includes(snippet)) {
    throw new Error(`role DingTalk H5 menu/button permission tree is not independent: ${snippet}`)
  }
}
