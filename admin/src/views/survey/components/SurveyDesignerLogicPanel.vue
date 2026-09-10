<template>
  <div class="logic-panel logic-full-panel">
    <div class="logic-toolbar">
      <h4 class="logic-title">自定义逻辑</h4>
      <div class="logic-actions">
        <el-button size="small" type="primary" @click="addNewRule">+ 增加规则</el-button>
        <el-button size="small" @click="emit('save', true)">保存全部规则</el-button>
      </div>
    </div>
    <div class="logic-guide">
      <strong>使用说明：</strong>每一条规则由「条件 + 动作 + 目标」组成。例如：<code>如果回答了Q1第1选项，则显示Q2</code>。
      条件可组合多个（且/或/非），动作可选显示隐藏、必填、跳转等。点击上方「+ 增加规则」添加。
    </div>
    <div class="logic-body">
      <div class="logic-editor-area">
        <div class="logic-list-title">规则列表（{{ rules.length }}条）</div>
        <div v-if="rules.length === 0" class="logic-empty">暂无规则，点击上方「+ 增加规则」添加</div>
        <div v-for="(rule, ruleIndex) in rules" :key="rule.id" class="logic-rule-card">
          <div class="rule-header">
            <span class="rule-index">#{{ Number(ruleIndex) + 1 }}</span>
            <span class="rule-dsl">{{ renderRuleDSL(rule) }}</span>
            <div class="rule-actions">
              <el-button text size="small" @click="editRule(Number(ruleIndex))">编辑</el-button>
              <el-button text size="small" type="danger" @click="removeRule(Number(ruleIndex))">删除</el-button>
            </div>
          </div>
        </div>
      </div>
      <div class="logic-sidebar">
        <div class="sidebar-title">案例参考</div>
        <div class="sidebar-examples">
          <div><code>IF Q1A1 THEN SHOW Q2</code> - 选A时显示Q2</div>
          <div><code>IF Q1A2 THEN HIDE Q3</code> - 选B时隐藏Q3</div>
          <div><code>IF Q1A1 THEN REQUIRED Q2</code> - 选A时Q2必填</div>
          <div><code>IF Q1A1 THEN BRANCH FROM Q1 TO END</code> - 选A结束问卷</div>
          <div><code>IF Q1A1 THEN BRANCH FROM Q1 TO Q5</code> - 选A跳到Q5</div>
          <div><code>IF AND(Q1A1,Q2&gt;18) THEN BRANCH FROM Q2 TO Q6</code> - 多条件跳转</div>
          <div><code>IF Q1A1 THEN CHECK Q2A1</code> - 选A时勾选Q2的选项A</div>
          <div><code>ASSIGNMENT Q3 WITH SUM(Q1,Q2)</code> - Q3赋值为Q1+Q2</div>
          <div><code>VALIDATE Q2 WITH IF(Q2&gt;100,"超出","")</code> - 校验Q2不超100</div>
          <div><code>REPLACE Q2 WITH "你好"</code> - 替换Q2标题</div>
          <div><code>POST_STAT VALUE CHANNEL(webhook)</code> - 交卷后统计并通知</div>
        </div>
      </div>
    </div>

    <el-dialog v-model="dialogVisible" :title="editingRuleIndex >= 0 ? '编辑规则' : '增加规则'" width="680px" :close-on-click-modal="false">
      <el-form label-position="top" size="small">
        <el-form-item label="条件类型">
          <el-select v-model="ruleForm.conditionType" class="full-width" :disabled="ruleForm.action === 'postStat'">
            <el-option label="简单条件 (单个)" value="simple" />
            <el-option label="且 (AND)" value="and" />
            <el-option label="或 (OR)" value="or" />
            <el-option label="非 (NOT)" value="not" />
            <el-option label="无条件 (赋值/校验/替换/交卷后统计)" value="none" />
          </el-select>
        </el-form-item>
        <template v-if="ruleForm.conditionType !== 'none'">
          <el-form-item v-for="(condition, conditionIndex) in ruleForm.conditions" :key="conditionIndex" :label="Number(conditionIndex) === 0 ? '条件' : `条件 ${Number(conditionIndex) + 1}`">
            <div class="condition-row">
              <el-select v-model="condition.questionIdx" placeholder="选择题目" class="condition-question" @change="condition.optionIdx = undefined; condition.operator = undefined">
                <el-option v-for="(question, questionIndex) in questions" :key="question.id" :label="`Q${Number(questionIndex) + 1}: ${question.title?.slice(0, 20)}`" :value="Number(questionIndex)" />
              </el-select>
              <el-select v-if="hasOptionsByIndex(condition.questionIdx)" v-model="condition.optionIdx" placeholder="选择选项" class="condition-option" clearable @change="condition.operator = 'optSelected'">
                <el-option v-for="(option, optionIndex) in getOptionsByIndex(condition.questionIdx)" :key="optionIndex" :label="option.label" :value="Number(optionIndex)" />
              </el-select>
              <el-select v-if="condition.optionIdx === undefined" v-model="condition.operator" placeholder="判断条件" class="condition-option">
                <el-option label="已填写" value="filled" />
                <el-option label="未填写" value="empty" />
                <el-option label="等于" value="eq" />
                <el-option label="大于" value="gt" />
                <el-option label="小于" value="lt" />
              </el-select>
              <el-input v-if="['eq', 'gt', 'lt'].includes(condition.operator || '')" v-model="condition.compareValue" placeholder="比较值" class="condition-value" />
              <el-button v-if="ruleForm.conditions.length > 1 && Number(conditionIndex) === ruleForm.conditions.length - 1" text size="small" type="danger" @click="ruleForm.conditions.splice(Number(conditionIndex), 1)">×</el-button>
            </div>
          </el-form-item>
          <el-button v-if="['and', 'or'].includes(ruleForm.conditionType)" text size="small" @click="ruleForm.conditions.push({})">+ 添加条件</el-button>
        </template>

        <el-form-item label="动作">
          <el-select v-model="ruleForm.action" class="full-width">
            <el-option label="显示 (SHOW)" value="show" />
            <el-option label="隐藏 (HIDE)" value="hide" />
            <el-option label="必填 (REQUIRED)" value="required" />
            <el-option label="跳转 (BRANCH)" value="branch" />
            <el-option label="勾选 (CHECK)" value="check" />
            <el-option label="赋值 (ASSIGNMENT)" value="assignment" />
            <el-option label="校验 (VALIDATE)" value="validate" />
            <el-option label="文本替换 (REPLACE)" value="replace" />
            <el-option label="结束问卷 (END)" value="end" />
            <el-option label="交卷后统计 (POST_STAT)" value="postStat" />
          </el-select>
        </el-form-item>

        <el-form-item v-if="['show', 'hide', 'required'].includes(ruleForm.action)" label="目标题目">
          <QuestionSelect v-model="ruleForm.targetQuestionIdx" :questions="questions" />
        </el-form-item>
        <template v-if="ruleForm.action === 'check'">
          <el-form-item label="目标题目">
            <QuestionSelect v-model="ruleForm.targetQuestionIdx" :questions="questions" @change="ruleForm.targetOptionIdxs = []" />
          </el-form-item>
          <el-form-item v-if="ruleForm.targetQuestionIdx !== undefined" label="目标选项（可多选）">
            <el-select v-model="ruleForm.targetOptionIdxs" multiple placeholder="选择选项" class="full-width">
              <el-option v-for="(option, optionIndex) in getOptionsByIndex(ruleForm.targetQuestionIdx)" :key="optionIndex" :label="option.label" :value="Number(optionIndex)" />
            </el-select>
          </el-form-item>
        </template>
        <template v-if="ruleForm.action === 'branch'">
          <el-form-item label="从题目"><QuestionSelect v-model="ruleForm.branchFromIdx" :questions="questions" /></el-form-item>
          <el-form-item label="跳到">
            <div class="branch-target">
              <QuestionSelect v-model="ruleForm.branchToIdx" :questions="questions" :disabled="ruleForm.branchToEnd" />
              <el-checkbox v-model="ruleForm.branchToEnd">结束问卷</el-checkbox>
            </div>
          </el-form-item>
        </template>
        <el-form-item v-if="ruleForm.action === 'end'" label="从题目">
          <QuestionSelect v-model="ruleForm.branchFromIdx" :questions="questions" placeholder="选择触发结束的题目" />
        </el-form-item>
        <el-form-item v-if="['assignment', 'validate', 'replace'].includes(ruleForm.action)" label="目标题目">
          <QuestionSelect v-model="ruleForm.targetQuestionIdx" :questions="questions" />
        </el-form-item>
        <el-form-item v-if="ruleForm.action === 'assignment'" label="赋值公式"><el-input v-model="ruleForm.formula" placeholder="如 SUM(Q1, Q2)" /></el-form-item>
        <el-form-item v-if="ruleForm.action === 'validate'" label="校验公式（返回空字符串表示通过）"><el-input v-model="ruleForm.formula" placeholder='如 IF(Q2>100,"不能超过100","")' /></el-form-item>
        <el-form-item v-if="ruleForm.action === 'replace'" label="替换文本/公式"><el-input v-model="ruleForm.formula" placeholder='如 CONCATENATE("你好，",Q1)' /></el-form-item>

        <template v-if="ruleForm.action === 'postStat'">
          <el-form-item label="统计字段">
            <el-radio-group v-model="ruleForm.statField">
              <el-radio value="label">选项标签 (label)</el-radio>
              <el-radio value="value">选项值 (value)</el-radio>
            </el-radio-group>
            <div class="field-help">选择统计时使用的字段，与「统计配置」中的统计方式联动</div>
          </el-form-item>
          <el-form-item label="统计范围">
            <el-radio-group v-model="ruleForm.statScope">
              <el-radio value="all">全部统计（统计所有答卷）</el-radio>
              <el-radio value="single">单卷统计（仅统计当前答卷）</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-divider />
          <el-form-item label="通知方式">
            <el-radio-group v-model="ruleForm.notifyChannel">
              <el-radio value="internal">站内通知</el-radio>
              <el-radio value="webhook">Webhook</el-radio>
              <el-radio value="both">全部</el-radio>
            </el-radio-group>
          </el-form-item>
          <template v-if="['webhook', 'both'].includes(ruleForm.notifyChannel || '')">
            <el-form-item label="Webhook 类型">
              <el-select v-model="ruleForm.webhookType" class="full-width">
                <el-option label="钉钉" value="dingtalk" />
                <el-option label="企业微信" value="wecom" />
                <el-option label="飞书" value="lark" />
                <el-option label="自定义" value="custom" />
              </el-select>
            </el-form-item>
            <el-form-item label="Webhook URL">
              <el-input v-model="ruleForm.webhookUrl" :placeholder="webhookPlaceholder" clearable />
              <div class="field-help">{{ webhookHint }}</div>
            </el-form-item>
          </template>
          <el-divider />
          <el-form-item label="通知消息模板">
            <div class="template-actions">
              <el-select v-model="templatePreset" size="small" class="template-select" placeholder="模版预设" clearable @change="applyTemplatePreset">
                <el-option-group label="内置模版">
                  <el-option v-for="preset in allPresets.filter(item => item.group === 'builtin')" :key="preset.value" :label="preset.label" :value="preset.value" />
                </el-option-group>
                <el-option-group v-if="customPresets.length" label="自定义模版">
                  <el-option v-for="preset in allPresets.filter(item => item.group === 'custom')" :key="preset.value" :label="preset.label" :value="preset.value">
                    <span>{{ preset.label }}</span>
                    <span class="preset-delete" @click.stop="deleteCustomPreset(preset.label)">删除</span>
                  </el-option>
                </el-option-group>
              </el-select>
              <el-button text size="small" @click="presetDialogRef?.open()">另存为</el-button>
              <el-button text size="small" :disabled="!ruleForm.messageTemplate" @click="ruleForm.messageTemplate = ''">恢复默认</el-button>
            </div>
            <el-input v-model="ruleForm.messageTemplate" type="textarea" :rows="5" placeholder="📋 问卷「{title}」收到新答卷&#10;提交人：{submitter}　时间：{date}&#10;共 {total} 份提交&#10;&#10;{result}" />
            <div class="field-help template-help">
              可用变量：<code>{title}</code> 问卷标题 · <code>{questionCount}</code> 题目数 · <code>{total}</code> 总答卷数 ·
              <code>{submitter}</code> 提交人 · <code>{date}</code> 提交时间 · <code>{result}</code> 统计结果<br>
              支持插入 <code>\n</code> 换行，留空使用默认模版
            </div>
          </el-form-item>
          <el-divider />
          <el-form-item label="通知对象">
            <el-checkbox v-model="ruleForm.notifyAdmin" class="notify-admin">通知管理员</el-checkbox>
            <el-input v-model="ruleForm.notifyUserIds" placeholder="指定成员 ID（多个用逗号隔开）" clearable />
            <div class="field-help">通知管理员和指定成员为选填，仅用于站内通知</div>
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmRule">确定</el-button>
      </template>
    </el-dialog>

    <SurveyDesignerPresetDialog ref="presetDialogRef" @save="savePreset" />
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, ref } from 'vue'
import { ElMessage, ElMessageBox, ElOption, ElSelect } from 'element-plus'
import { adminApi } from '../../../api'
import request from '../../../utils/request'
import type { SurveyDesignerQuestion, SurveyLogicCondition, SurveyLogicRule } from '../survey-designer-types'
import SurveyDesignerPresetDialog from './SurveyDesignerPresetDialog.vue'

const props = defineProps<{ questions: SurveyDesignerQuestion[]; surveyExists: boolean }>()
const emit = defineEmits<{ save: [showMessage: boolean] }>()
const rules = defineModel<SurveyLogicRule[]>({ required: true })

const dialogVisible = ref(false)
const editingRuleIndex = ref(-1)
const ruleForm = ref<SurveyLogicRule>(createRule())
const templatePreset = ref('')
const customPresets = ref<Array<{ label: string; value: string }>>([])
const builtinPresets = ref<Record<string, string>>({
  '简洁': '📋 问卷「{title}」收到新答卷\n新提交人：{submitter}\n时间：{date}\n共 {total} 份提交\n\n{result}',
  '详细': '📊 问卷统计报告\n━━━━━━━━━━━━━━━━━━\n📌 问卷：{title}\n👤 新提交人：{submitter}\n🕐 时间：{date}\n📈 总提交数：{total}\n\n{result}\n━━━━━━━━━━━━━━━━━━',
  '仅统计结果': '{result}',
})
const presetDialogRef = ref<{ open: () => void; close: () => void } | null>(null)

const QuestionSelect = defineComponent({
  props: {
    modelValue: Number,
    questions: { type: Array as () => SurveyDesignerQuestion[], required: true },
    disabled: Boolean,
    placeholder: { type: String, default: '选择题目' },
  },
  emits: ['update:modelValue', 'change'],
  setup(selectProps, { emit: selectEmit }) {
    return () => h(ElSelect, {
      modelValue: selectProps.modelValue,
      disabled: selectProps.disabled,
      placeholder: selectProps.placeholder,
      style: 'width:100%',
      'onUpdate:modelValue': (value: number) => selectEmit('update:modelValue', value),
      onChange: (value: number) => selectEmit('change', value),
    }, () => selectProps.questions.map((question, index) => h(ElOption, {
      key: question.id,
      label: `Q${index + 1}: ${question.title?.slice(0, 30)}`,
      value: index,
    })))
  },
})

const webhookPlaceholder = computed(() => ({
  dingtalk: 'https://oapi.dingtalk.com/robot/send?access_token=xxx',
  wecom: 'https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx',
  lark: 'https://open.feishu.cn/open-apis/bot/v2/hook/xxx',
  custom: 'https://your-domain.com/webhook/xxx',
}[ruleForm.value.webhookType || 'dingtalk'] || ''))
const webhookHint = computed(() => ({
  dingtalk: '在钉钉群中添加机器人获取 Webhook 地址',
  wecom: '在企业微信群中添加机器人获取 Webhook 地址',
  lark: '在飞书群中添加机器人获取 Webhook 地址',
  custom: '通用 Webhook，POST JSON 格式消息体',
}[ruleForm.value.webhookType || 'dingtalk'] || ''))
const allPresets = computed(() => [
  ...Object.keys(builtinPresets.value).map(label => ({ label, value: label, group: 'builtin' })),
  ...customPresets.value.map(item => ({ label: item.label, value: `custom:${item.label}`, group: 'custom' })),
])

void loadPresets()

function createRule(): SurveyLogicRule {
  return {
    id: '', conditionType: 'simple', conditions: [{}], action: 'show', scope: 'frontend',
    targetOptionIdxs: [], branchToEnd: false, formula: '', statField: 'label', statScope: 'all',
    notifyAdmin: false, notifyUserIds: '', notifyChannel: 'internal', webhookType: 'dingtalk', webhookUrl: '', messageTemplate: '',
  }
}
function addNewRule() {
  editingRuleIndex.value = -1
  ruleForm.value = createRule()
  dialogVisible.value = true
}
function editRule(index: number) {
  editingRuleIndex.value = index
  ruleForm.value = JSON.parse(JSON.stringify(rules.value[index]))
  dialogVisible.value = true
}
function confirmRule() {
  const rule = ruleForm.value
  if (rule.conditionType !== 'none' && rule.conditions.some(condition => condition.questionIdx === undefined)) return ElMessage.warning('请选择条件题目')
  if (!rule.action) return ElMessage.warning('请选择动作')
  if (rule.action === 'postStat') {
    if (!rule.statField) return ElMessage.warning('请选择统计字段')
    if (!rule.notifyChannel) return ElMessage.warning('请选择通知方式')
    if (['webhook', 'both'].includes(rule.notifyChannel) && !rule.webhookUrl) return ElMessage.warning('请填写 Webhook URL')
    if (['webhook', 'both'].includes(rule.notifyChannel) && !rule.webhookType) return ElMessage.warning('请选择 Webhook 类型')
  }
  if (['show', 'hide', 'required', 'check', 'assignment', 'validate', 'replace'].includes(rule.action) && rule.targetQuestionIdx === undefined) return ElMessage.warning('请选择目标题目')
  if (['branch', 'end'].includes(rule.action) && rule.branchFromIdx === undefined) return ElMessage.warning(rule.action === 'end' ? '请选择结束来源题目' : '请选择跳转来源题目')
  if (rule.action === 'branch' && !rule.branchToEnd && rule.branchToIdx === undefined) return ElMessage.warning('请选择跳转目标题目或勾选结束问卷')

  const savedRule = { ...rule, id: `rule_${Date.now()}_${Math.random().toString(36).slice(2, 6)}`, scope: rule.action === 'postStat' ? 'backend' : 'frontend' } as SurveyLogicRule
  const nextRules = [...rules.value]
  if (editingRuleIndex.value >= 0) nextRules[editingRuleIndex.value] = savedRule
  else nextRules.push(savedRule)
  rules.value = nextRules
  dialogVisible.value = false
  ruleForm.value = createRule()
  editingRuleIndex.value = -1
  if (props.surveyExists) emit('save', false)
}
function removeRule(index: number) {
  ElMessageBox.confirm('确定删除该条规则？', '提示', { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }).then(() => {
    const nextRules = [...rules.value]
    nextRules.splice(index, 1)
    rules.value = nextRules
    emit('save', true)
  }).catch(() => {})
}
function hasOptionsByIndex(index?: number) {
  const question = index === undefined ? undefined : props.questions[index]
  return !!question && ['select', 'radio', 'checkbox', 'picker', 'cascade', 'judge', 'multiInput', 'hInput'].includes(question.type) && !!question.props?.options?.length
}
function getOptionsByIndex(index?: number) {
  return index === undefined ? [] : props.questions[index]?.props?.options || []
}

function conditionToDSL(condition: SurveyLogicCondition) {
  const questionTag = condition.questionIdx !== undefined ? `Q${condition.questionIdx + 1}` : '?'
  if (condition.optionIdx !== undefined) return `${questionTag}A${condition.optionIdx + 1}`
  if (condition.operator === 'filled') return `NOT(ISBLANK(${questionTag}))`
  if (condition.operator === 'empty') return `ISBLANK(${questionTag})`
  if (condition.operator) return `${questionTag}${condition.operator === 'eq' ? '==' : condition.operator}${condition.compareValue || ''}`
  return questionTag
}
function renderRuleDSL(rule: SurveyLogicRule) {
  const condition = rule.conditionType === 'none'
    ? ''
    : rule.conditions.length === 1
      ? conditionToDSL(rule.conditions[0])
      : `${rule.conditionType.toUpperCase()}(${rule.conditions.map(conditionToDSL).join(', ')})`
  const target = rule.targetQuestionIdx !== undefined ? `Q${rule.targetQuestionIdx + 1}` : '?'
  const branchFrom = rule.branchFromIdx !== undefined ? `Q${rule.branchFromIdx + 1}` : '?'
  const branchTo = rule.branchToIdx !== undefined ? `Q${rule.branchToIdx + 1}` : '?'
  let action = ''
  if (rule.action === 'branch') action = `BRANCH FROM ${branchFrom} TO ${rule.branchToEnd ? 'END' : branchTo}`
  else if (rule.action === 'check') action = `CHECK ${(rule.targetOptionIdxs || []).map(index => `${target}A${index + 1}`).join(' ')}`
  else if (['assignment', 'validate', 'replace'].includes(rule.action)) action = `${rule.action.toUpperCase()} ${target} WITH ${rule.formula || '?'}`
  else if (rule.action === 'end') action = `BRANCH FROM ${branchFrom} TO END`
  else if (rule.action === 'postStat') {
    const parts = [`POST_STAT ${rule.statField === 'value' ? 'VALUE' : 'LABEL'} CHANNEL(${rule.notifyChannel || 'internal'})`]
    if (rule.webhookUrl) parts.push(`WEBHOOK(${rule.webhookType || 'dingtalk'}:${rule.webhookUrl})`)
    if (rule.notifyAdmin) parts.push('NOTIFY_ADMIN')
    if (rule.notifyUserIds) parts.push(`USERS(${rule.notifyUserIds})`)
    if (rule.messageTemplate) parts.push(`TEMPLATE("${rule.messageTemplate.replace(/"/g, '\\"')}")`)
    action = parts.join(' ')
  } else action = `${({ show: 'SHOW', hide: 'HIDE', required: 'REQUIRED' } as Record<string, string>)[rule.action] || rule.action.toUpperCase()} ${target}`
  return condition ? `IF ${condition} THEN ${action}` : action
}

async function loadPresets() {
  try {
    const [customResponse, builtinResponse] = await Promise.all([
      adminApi.surveyTemplatePresetsGet(),
      request.get<string | Array<{ label: string; value: string }>>('/api/v2/home/setup', { params: { key: 'BUILTIN_TEMPLATE_PRESETS' } }),
    ])
    if (customResponse.code === 0 && Array.isArray(customResponse.data)) customPresets.value = customResponse.data
    if (builtinResponse.code === 0 && builtinResponse.data) {
      const parsed = typeof builtinResponse.data === 'string' ? JSON.parse(builtinResponse.data) : builtinResponse.data
      if (Array.isArray(parsed) && parsed.length) builtinPresets.value = Object.fromEntries(parsed.map(item => [item.label, item.value]))
    }
  } catch {}
}
async function persistCustomPresets() {
  try { await adminApi.surveyTemplatePresetsSave({ presets: customPresets.value }) } catch {}
}
function applyTemplatePreset(value: string) {
  if (!value) return
  if (value.startsWith('custom:')) ruleForm.value.messageTemplate = customPresets.value.find(item => item.label === value.slice(7))?.value || ''
  else ruleForm.value.messageTemplate = builtinPresets.value[value] || ''
}
function savePreset(rawName: string) {
  const name = rawName.trim()
  if (!name) return ElMessage.warning('请输入模版名称')
  if (!ruleForm.value.messageTemplate) return ElMessage.warning('请先填写模版内容')
  const entry = { label: name, value: ruleForm.value.messageTemplate }
  const index = customPresets.value.findIndex(item => item.label === name)
  if (index >= 0) customPresets.value[index] = entry
  else customPresets.value.push(entry)
  void persistCustomPresets()
  presetDialogRef.value?.close()
  ElMessage.success('已保存自定义模版')
}
function deleteCustomPreset(label: string) {
  customPresets.value = customPresets.value.filter(item => item.label !== label)
  void persistCustomPresets()
}
</script>

<style scoped>
.logic-full-panel { display:flex; flex:1; flex-direction:column; overflow:hidden; padding:24px 32px; background:var(--designer-bg); }
.logic-toolbar { display:flex; flex-shrink:0; align-items:center; justify-content:space-between; margin-bottom:14px; padding:12px 14px; border:1px solid var(--designer-border); border-radius:10px; background:#fff; box-shadow:0 8px 18px rgba(15,23,42,.04); }
.logic-title { margin:0; font-size:14px; }
.logic-actions, .template-actions { display:flex; align-items:center; gap:8px; }
.logic-guide { margin-bottom:12px; padding:8px 12px; border:1px solid #faf0c7; border-radius:4px; background:#fefced; color:#606266; font-size:12px; line-height:1.8; }
.logic-body { display:flex; flex:1; gap:16px; overflow:hidden; }
.logic-editor-area { display:flex; min-width:0; flex:1; flex-direction:column; overflow-y:auto; }
.logic-list-title { margin-bottom:8px; color:#888; font-size:12px; }
.logic-empty { padding:40px 0; color:#999; font-size:13px; text-align:center; }
.logic-sidebar { width:260px; flex-shrink:0; overflow-y:auto; padding:14px; border:1px solid var(--designer-border); border-radius:10px; background:#fff; box-shadow:0 8px 18px rgba(15,23,42,.04); }
.sidebar-title { margin-bottom:6px; color:#606266; font-size:12px; font-weight:600; }
.sidebar-examples { font-size:12px; line-height:2.2; }
.logic-rule-card { margin-bottom:8px; border:1px solid var(--designer-border); border-radius:9px; background:#fff; transition:box-shadow .15s,border-color .15s; }
.logic-rule-card:hover { border-color:var(--designer-accent-border); box-shadow:0 8px 18px rgba(37,99,235,.08); }
.rule-header { display:flex; align-items:center; gap:8px; padding:10px 12px; }
.rule-index { width:24px; flex-shrink:0; color:#999; font-size:11px; font-weight:600; }
.rule-dsl { flex:1; color:#303133; font-family:Consolas,Monaco,"Courier New",monospace; font-size:12px; }
.rule-actions { flex-shrink:0; }
.full-width { width:100%; }
.condition-row { display:flex; flex-wrap:wrap; align-items:center; gap:6px; }
.condition-question { width:160px; }
.condition-option { width:120px; }
.condition-value { width:100px; }
.branch-target { display:flex; width:100%; align-items:center; gap:8px; }
.branch-target :deep(.el-select) { min-width:280px; flex:1; }
.branch-target :deep(.el-checkbox) { white-space:nowrap; }
.field-help { margin-top:4px; color:#999; font-size:11px; }
.template-select { width:140px; }
.preset-delete { float:right; color:#999; font-size:11px; cursor:pointer; }
.template-help { line-height:1.8; }
.notify-admin { display:block; margin-bottom:8px; }
@media (max-width:1080px) {
  .logic-full-panel { padding:18px; }
  .logic-body { flex-direction:column; overflow-y:auto; }
  .logic-sidebar { width:auto; }
}
</style>
