import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import process from 'node:process'

const root = process.cwd()
const read = relativePath => readFileSync(resolve(root, relativePath), 'utf8')

const htmlSource = read('index.html')
const dingTalkSource = read('src/utils/dingtalk.ts')
const authStoreSource = read('src/stores/dingtalkAuth.ts')

assert.ok(htmlSource.includes('id="app-favicon"'), 'H5 入口缺少可动态更新的 favicon 标识')
assert.ok(htmlSource.includes('data-default-href="static/logo.png"'), 'H5 入口缺少 favicon 默认回退地址')
assert.ok(dingTalkSource.includes('setBrowserFavicon'), 'H5 工具缺少动态 favicon 更新方法')
assert.ok(dingTalkSource.includes('link[rel~="icon"]'), '动态 favicon 方法未查找浏览器标签图标')
assert.ok(dingTalkSource.includes('data-default-href'), '动态 favicon 方法未保留默认回退地址')
assert.ok(dingTalkSource.includes('favicon.onerror'), '动态 favicon 方法缺少图片加载失败回退')
assert.ok(authStoreSource.includes('setBrowserFavicon(appConfig.value.logoUrl)'), '公开应用配置未同步到浏览器 favicon')

console.log('H5 app branding checks passed')
