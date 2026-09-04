import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import process from 'node:process'

const source = readFileSync(
  resolve(process.cwd(), 'src/pages/workflow/components/WorkflowFieldControl.vue'),
  'utf8',
)

const selectCondition = source.indexOf(`v-else-if="field.type === 'select'"`)
const selectOpeningStart = source.lastIndexOf('<view', selectCondition)
const selectOpeningEnd = source.indexOf('>', selectCondition)
const selectOpeningTag = source.slice(selectOpeningStart, selectOpeningEnd + 1)
const selectType = source.indexOf('type="select"', selectOpeningEnd)
const selectInputStart = source.lastIndexOf('<u-input', selectType)
const selectInputEnd = source.indexOf('/>', selectType)
const selectInput = source.slice(selectInputStart, selectInputEnd + 2)

assert.ok(selectInput.includes('@click="openSelect"'), '选择输入框必须直接触发 openSelect')
assert.ok(!selectOpeningTag.includes('@click="openSelect"'), '不能依赖会被 u-input 阻止的外层点击冒泡')
assert.ok(source.includes('function resolveMobileSelect()'), 'select 必须区分 PC 和移动端交互')
assert.ok(source.includes('v-if="desktopSelectOpen" class="workflow-control__select-panel"'), 'PC 端必须使用输入框下方的选项面板')
assert.ok(source.includes('v-model="mobileSelectVisible"'), '移动端必须保留底部选择器')
assert.ok(source.includes('@click="selectDesktopOption(option)"'), 'PC 端选项必须可直接点击选中')
assert.ok(!source.includes(':custom-class="`workflow-control__select-option'), 'PC 端选项不应使用自带框线的 u-button')

console.log('workflow option source checks passed')
