<template>
  <div class="logic-panel logic-full-panel">
    <div class="logic-toolbar">
      <h4>自定义逻辑</h4>
      <div class="logic-actions">
        <el-button size="small" type="primary" @click="addRule">+ 增加规则</el-button>
        <el-button size="small" @click="emit('save')">保存全部规则</el-button>
      </div>
    </div>
    <div class="logic-guide">
      <strong>使用说明：</strong>每一条规则由「条件 + 动作 + 目标」组成。例如：<code>如果回答了Q1第1选项，则显示Q2</code>。
      条件可组合多个（且/或/非），动作可选显示隐藏、必填、跳转等。
    </div>
    <div class="logic-body">
      <div class="logic-editor-area">
        <div class="logic-list-title">规则列表（{{ rules.length }}条）</div>
        <div v-if="!rules.length" class="logic-empty">暂无规则，点击上方「+ 增加规则」添加</div>
        <div v-for="(rule, index) in rules" :key="rule.id" class="logic-rule-card">
          <div class="rule-header">
            <span class="rule-index">#{{ Number(index) + 1 }}</span>
            <span class="rule-dsl">{{ renderRuleDSL(rule) }}</span>
            <div class="rule-actions">
              <el-button text size="small" @click="editRule(Number(index))">编辑</el-button>
              <el-button text size="small" type="danger" @click="removeRule(Number(index))">删除</el-button>
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
          <div><code>IF Q1A1 THEN BRANCH FROM Q1 TO END</code> - 选A结束考试</div>
          <div><code>IF Q1A1 THEN CHECK Q2A1</code> - 选A时勾选Q2的选项A</div>
          <div><code>ASSIGNMENT Q3 WITH SUM(Q1,Q2)</code> - Q3赋值为Q1+Q2</div>
        </div>
      </div>
    </div>

    <el-dialog v-model="dialogVisible" :title="editingRuleIndex >= 0 ? '编辑规则' : '增加规则'" width="560px" :close-on-click-modal="false">
      <el-form label-position="top" size="small">
        <el-form-item label="条件类型">
          <el-select v-model="ruleForm.conditionType" class="full-width">
            <el-option label="简单条件 (单个)" value="simple" />
            <el-option label="且 (AND)" value="and" />
            <el-option label="或 (OR)" value="or" />
            <el-option label="非 (NOT)" value="not" />
            <el-option label="无条件 (赋值/校验/替换)" value="none" />
          </el-select>
        </el-form-item>
        <template v-if="ruleForm.conditionType !== 'none'">
          <el-form-item v-for="(condition, index) in ruleForm.conditions" :key="index" :label="Number(index) === 0 ? '条件' : `条件 ${Number(index) + 1}`">
            <div class="condition-row">
              <el-select v-model="condition.questionIdx" placeholder="选择题目" class="condition-question" @change="condition.optionIdx=undefined;condition.operator=undefined">
                <el-option v-for="(question, questionIndex) in questions" :key="questionIndex" :label="`Q${Number(questionIndex) + 1}: ${question.title?.slice(0,20)}`" :value="Number(questionIndex)" />
              </el-select>
              <el-select v-if="hasOptionsByIndex(condition.questionIdx)" v-model="condition.optionIdx" placeholder="选择选项" class="condition-option" clearable @change="condition.operator='optSelected'">
                <el-option v-for="(item, optionIndex) in getOptionsByIndex(condition.questionIdx)" :key="optionIndex" :label="item.label" :value="Number(optionIndex)" />
              </el-select>
              <el-select v-if="condition.optionIdx === undefined" v-model="condition.operator" placeholder="判断条件" class="condition-option">
                <el-option label="已填写" value="filled" /><el-option label="未填写" value="empty" />
                <el-option label="等于" value="eq" /><el-option label="大于" value="gt" /><el-option label="小于" value="lt" />
              </el-select>
              <el-input v-if="['eq','gt','lt'].includes(condition.operator || '')" v-model="condition.compareValue" placeholder="比较值" class="condition-value" />
              <el-button v-if="ruleForm.conditions.length > 1 && Number(index) === ruleForm.conditions.length - 1" text size="small" type="danger" @click="ruleForm.conditions.splice(Number(index), 1)">✕</el-button>
            </div>
          </el-form-item>
          <el-button v-if="['and','or'].includes(ruleForm.conditionType)" text size="small" @click="ruleForm.conditions.push({})">+ 添加条件</el-button>
        </template>
        <el-form-item label="动作">
          <el-select v-model="ruleForm.action" class="full-width">
            <el-option label="显示 (SHOW)" value="show" /><el-option label="隐藏 (HIDE)" value="hide" />
            <el-option label="必填 (REQUIRED)" value="required" /><el-option label="跳转 (BRANCH)" value="branch" />
            <el-option label="勾选 (CHECK)" value="check" /><el-option label="赋值 (ASSIGNMENT)" value="assignment" />
            <el-option label="校验 (VALIDATE)" value="validate" /><el-option label="文本替换 (REPLACE)" value="replace" />
            <el-option label="结束考试 (END)" value="end" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="['show','hide','required','assignment','validate','replace'].includes(ruleForm.action)" label="目标题目">
          <el-select v-model="ruleForm.targetQuestionIdx" placeholder="选择题目" class="full-width">
            <el-option v-for="(question, index) in questions" :key="index" :label="`Q${Number(index) + 1}: ${question.title?.slice(0,30)}`" :value="Number(index)" />
          </el-select>
        </el-form-item>
        <template v-if="ruleForm.action === 'check'">
          <el-form-item label="目标题目">
            <el-select v-model="ruleForm.targetQuestionIdx" placeholder="选择题目" class="full-width" @change="ruleForm.targetOptionIdxs=[]">
              <el-option v-for="(question, index) in questions" :key="index" :label="`Q${Number(index) + 1}: ${question.title?.slice(0,30)}`" :value="Number(index)" />
            </el-select>
          </el-form-item>
          <el-form-item v-if="ruleForm.targetQuestionIdx !== undefined" label="目标选项（可多选）">
            <el-select v-model="ruleForm.targetOptionIdxs" multiple placeholder="选择选项" class="full-width">
              <el-option v-for="(item, index) in getOptionsByIndex(ruleForm.targetQuestionIdx)" :key="index" :label="item.label" :value="Number(index)" />
            </el-select>
          </el-form-item>
        </template>
        <template v-if="ruleForm.action === 'branch'">
          <el-form-item label="从题目"><el-select v-model="ruleForm.branchFromIdx" placeholder="选择题目" class="full-width"><el-option v-for="(question, index) in questions" :key="index" :label="`Q${Number(index) + 1}: ${question.title?.slice(0,30)}`" :value="Number(index)" /></el-select></el-form-item>
          <el-form-item label="跳到"><div class="branch-row"><el-select v-model="ruleForm.branchToIdx" placeholder="选择目标题目" :disabled="ruleForm.branchToEnd"><el-option v-for="(question, index) in questions" :key="index" :label="`Q${Number(index) + 1}: ${question.title?.slice(0,30)}`" :value="Number(index)" /></el-select><el-checkbox v-model="ruleForm.branchToEnd">结束考试</el-checkbox></div></el-form-item>
        </template>
        <el-form-item v-if="ruleForm.action === 'assignment'" label="赋值公式"><el-input v-model="ruleForm.formula" placeholder="如 SUM(Q1, Q2)" /></el-form-item>
        <el-form-item v-if="ruleForm.action === 'validate'" label="校验公式（返回空字符串表示通过）"><el-input v-model="ruleForm.formula" placeholder='如 IF(Q2>100,"不能超过100","")' /></el-form-item>
        <el-form-item v-if="ruleForm.action === 'replace'" label="替换文本/公式"><el-input v-model="ruleForm.formula" placeholder='如 CONCATENATE("你好，",Q1)' /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button type="primary" @click="confirmRule">确定</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { ExamDesignerQuestion, ExamLogicCondition, ExamLogicRule } from '../exam-designer-types'

const props = defineProps<{ questions: ExamDesignerQuestion[] }>()
const rules = defineModel<ExamLogicRule[]>({ required: true })
const emit = defineEmits<{ save: [] }>()
const dialogVisible = ref(false)
const editingRuleIndex = ref(-1)
const ruleForm = ref<ExamLogicRule>(createRule())
const choiceTypes = ['select', 'radio', 'checkbox', 'picker', 'cascade', 'judge', 'multiInput', 'hInput']

function createRule(): ExamLogicRule {
  return { id: '', conditionType: 'simple', conditions: [{}], action: 'show', scope: 'frontend', targetOptionIdxs: [], branchToEnd: false, formula: '' }
}
function addRule() { editingRuleIndex.value = -1; ruleForm.value = createRule(); dialogVisible.value = true }
function hasOptionsByIndex(index?: number) { const question = index === undefined ? null : props.questions[index]; return Boolean(question && choiceTypes.includes(question.type) && question.props?.options?.length) }
function getOptionsByIndex(index?: number) { const question = index === undefined ? null : props.questions[index]; return question && choiceTypes.includes(question.type) ? question.props?.options || [] : [] }
function conditionToDSL(condition: ExamLogicCondition) {
  const tag = `Q${(condition.questionIdx ?? -1) + 1}`
  if (condition.optionIdx !== undefined) return `${tag}A${condition.optionIdx + 1}`
  if (condition.operator === 'filled') return `NOT(ISBLANK(${tag}))`
  if (condition.operator === 'empty') return `ISBLANK(${tag})`
  if (condition.operator) return `${tag}${condition.operator === 'eq' ? '==' : condition.operator}${condition.compareValue || ''}`
  return tag
}
function renderRuleDSL(rule: ExamLogicRule) {
  const condition = rule.conditionType === 'none' ? '' : rule.conditions.length === 1
    ? conditionToDSL(rule.conditions[0])
    : `${rule.conditionType.toUpperCase()}(${rule.conditions.map(conditionToDSL).join(', ')})`
  const target = rule.targetQuestionIdx !== undefined ? `Q${rule.targetQuestionIdx + 1}` : '?'
  let action = `${({ show: 'SHOW', hide: 'HIDE', required: 'REQUIRED' } as Record<string, string>)[rule.action] || rule.action.toUpperCase()} ${target}`
  if (rule.action === 'branch') action = `BRANCH FROM Q${(rule.branchFromIdx ?? -1) + 1} TO ${rule.branchToEnd ? 'END' : `Q${(rule.branchToIdx ?? -1) + 1}`}`
  else if (rule.action === 'check') action = `CHECK ${(rule.targetOptionIdxs || []).map(index => `${target}A${index + 1}`).join(' ')}`
  else if (['assignment', 'validate', 'replace'].includes(rule.action)) action = `${rule.action.toUpperCase()} ${target} WITH ${rule.formula || '?'}`
  else if (rule.action === 'end') action = `BRANCH FROM Q${(rule.branchFromIdx ?? -1) + 1} TO END`
  return rule.conditionType === 'none' ? action : `IF ${condition} THEN ${action}`
}
function confirmRule() {
  const value = ruleForm.value
  if (value.conditionType !== 'none' && value.conditions.some(condition => condition.questionIdx === undefined)) return void ElMessage.warning('请选择条件题目')
  if (['show','hide','required','check','assignment','validate','replace'].includes(value.action) && value.targetQuestionIdx === undefined) return void ElMessage.warning('请选择目标题目')
  if (value.action === 'branch' && value.branchFromIdx === undefined) return void ElMessage.warning('请选择跳转来源题目')
  if (value.action === 'branch' && !value.branchToEnd && value.branchToIdx === undefined) return void ElMessage.warning('请选择跳转目标题目或勾选结束考试')
  value.id = `rule_${Date.now()}_${Math.random().toString(36).slice(2, 6)}`
  if (editingRuleIndex.value >= 0) rules.value[editingRuleIndex.value] = { ...value }
  else rules.value.push({ ...value })
  dialogVisible.value = false
}
function editRule(index: number) { editingRuleIndex.value = index; ruleForm.value = JSON.parse(JSON.stringify(rules.value[index])); dialogVisible.value = true }
function removeRule(index: number) { rules.value.splice(index, 1) }
</script>

<style scoped>
.logic-panel { flex:1; min-width:0; overflow:auto; padding:16px; background:var(--designer-bg); }
.logic-toolbar, .rule-header, .logic-actions, .condition-row, .branch-row { display:flex; align-items:center; gap:8px; }
.logic-toolbar { justify-content:space-between; margin-bottom:12px; }
.logic-toolbar h4 { margin:0; font-size:14px; }
.logic-guide { margin-bottom:12px; padding:8px 12px; border:1px solid #faf0c7; border-radius:4px; background:#fefced; color:#606266; font-size:12px; line-height:1.8; }
.logic-body { display:grid; grid-template-columns:minmax(0,1fr) 280px; gap:14px; }
.logic-editor-area, .logic-sidebar { min-width:0; padding:12px; border:1px solid var(--designer-border); border-radius:8px; background:#fff; }
.logic-list-title, .sidebar-title { margin-bottom:8px; color:#606266; font-size:12px; font-weight:600; }
.logic-empty { padding:40px 0; color:#999; font-size:13px; text-align:center; }
.logic-rule-card { margin-bottom:8px; padding:10px; border:1px solid #e5e7eb; border-radius:6px; }
.rule-dsl { flex:1; min-width:0; font-family:monospace; font-size:12px; }
.rule-index { flex-shrink:0; color:var(--designer-accent); font-weight:600; }
.sidebar-examples { font-size:12px; line-height:2.2; }
.full-width { width:100%; }
.condition-row { flex-wrap:wrap; }
.condition-question { width:160px; }.condition-option { width:120px; }.condition-value { width:100px; }
.branch-row { width:100%; }.branch-row :deep(.el-select) { flex:1; min-width:280px; }.branch-row :deep(.el-checkbox) { white-space:nowrap; }
</style>
