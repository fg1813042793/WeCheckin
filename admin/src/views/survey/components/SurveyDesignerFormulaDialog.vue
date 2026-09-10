<template>
  <el-dialog v-model="visible" :title="title" width="720px" :close-on-click-modal="false">
    <div class="formula-layout">
      <div class="formula-main">
        <div class="formula-label">{{ formulaType === 'calculate' ? '计算结果将自动填充到当前题目' : '公式表达式' }}</div>
        <el-input ref="formulaInputRef" v-model="formulaText" type="textarea" :rows="12" :placeholder="placeholder" />
        <div class="formula-help">
          <div class="help-title">使用说明</div>
          <ul>
            <li>运算符、括号、逗号请使用<strong>英文输入法</strong></li>
            <li>判断相等用 <code>==</code>，赋值用 <code>=</code></li>
            <li>手动输入的文本用引号包裹，例如 <code>"优秀"</code></li>
            <li>公式结果类型需与目标题型匹配（数字题不能接收文本结果）</li>
            <li>点击右侧<strong class="primary-text">蓝色高亮标签</strong>可插入题目标签到光标位置</li>
          </ul>
        </div>
        <div class="formula-reference">
          <div class="reference-title">函数参考 &amp; 案例</div>
          <div><code>IF(条件, TRUE, FALSE)</code> — 条件判断</div>
          <div><code>AND(条件1, 条件2)</code> — 与运算</div>
          <div><code>OR(条件1, 条件2)</code> — 或运算</div>
          <div><code>NOT(条件)</code> — 非运算</div>
          <div><code>COUNT(Q1)</code> — 统计多选已选数量</div>
          <div><code>SCORE(Q1)</code> — 读取选项分值</div>
          <div><code>TEXT(Q1)</code> — 读取选项文本</div>
          <div><code>CURRENT_DATE("YYYY/MM/DD")</code> — 当前日期</div>
          <div><code>INCLUDE("学校", Q1)</code> — 判断文本是否包含关键词</div>
          <div v-if="formulaType === 'calculate'" class="formula-examples">
            <div class="reference-title">计算公式案例</div>
            <div><code>SUM(Q1, Q2)</code> — 两题求和（计算总分）</div>
            <div><code>ROUND(Q1 / POWER(Q2 / 100, 2), 2)</code> — BMI 指数</div>
            <div><code>IFS(Q1&lt;60,"不合格",Q1&lt;80,"合格",TRUE(),"优秀")</code> — 多条件等级</div>
            <div><code>CONCATENATE("你的总分是", Q1, "分")</code> — 拼接结果文案</div>
            <div><code>Q1 + Q2 + Q3</code> — 连加运算</div>
            <div><code>(Q1 + Q2) / 2</code> — 求平均值</div>
          </div>
          <div v-if="formulaType === 'endFormula' || formulaType === 'jumpFormula'" class="formula-examples">
            <div class="reference-title">跳转/结束案例</div>
            <div><code>IF Q1A1 THEN BRANCH FROM Q1 TO Q5</code> — 选择选项A跳到Q5</div>
            <div><code>IF Q1A2 THEN BRANCH FROM Q1 TO END</code> — 选择选项B结束问卷</div>
            <div><code>IF AND(Q1A1, Q2&gt;18) THEN BRANCH FROM Q1 TO Q6</code> — 多条件跳转</div>
          </div>
          <div v-if="formulaType === 'requiredFormula'" class="formula-examples">
            <div class="reference-title">必填公式案例</div>
            <div><code>IF Q1A1 THEN REQUIRED Q2</code> — 选A时Q2必填</div>
            <div><code>IF Q1A1 AND Q1A2 THEN REQUIRED Q3</code> — 同时选A、B时Q3必填</div>
            <div><code>IF NOT(ISBLANK(Q1)) THEN REQUIRED Q2</code> — Q1非空时Q2必填</div>
          </div>
          <div v-if="formulaType === 'textReplace'" class="formula-examples">
            <div class="reference-title">文本替换案例</div>
            <div><code>REPLACE Q2 WITH CONCATENATE("你好，", Q1)</code> — 替换Q2标题</div>
            <div><code>REPLACE Q3 WITH CONCATENATE("你的得分：", SUM(Q1, Q2))</code> — 替换为计算结果</div>
          </div>
          <div v-if="formulaType === 'optShow'" class="formula-examples">
            <div class="reference-title">显示隐藏案例</div>
            <div><code>IF Q1A1 THEN SHOW</code> — 选A时显示本选项</div>
            <div><code>IF Q1A2 THEN HIDE</code> — 选A时隐藏本选项</div>
          </div>
          <div v-if="formulaType === 'optAutoCheck'" class="formula-examples">
            <div class="reference-title">自动勾选案例</div>
            <div><code>IF Q1A1 THEN CHECK Q2A1</code> — 选A时自动勾选Q2的选项A</div>
            <div><code>IF Q1A1 THEN CHECK Q2A1 Q2A2</code> — 选A时自动勾选Q2的A、B选项</div>
          </div>
        </div>
      </div>
      <div class="formula-tags">
        <div class="formula-label">题目标签</div>
        <div class="tag-list">
          <div v-for="question in formulaQuestions" :key="question.id" class="tag-question">
            <div class="tag-question-title">{{ question.title?.slice(0, 20) }}</div>
            <div class="tag-items">
              <el-tag size="small" type="primary" @click="insertFormulaTag(questionTag(question))">{{ questionTag(question) }}</el-tag>
              <template v-if="question.props?.options?.length">
                <el-tag
                  v-for="(_, optionIndex) in question.props.options.slice(0, 4)"
                  :key="optionIndex"
                  size="small"
                  type="warning"
                  @click="insertFormulaTag(`${questionTag(question)}A${Number(optionIndex) + 1}`)"
                >
                  {{ questionTag(question) }}A{{ Number(optionIndex) + 1 }}
                </el-tag>
              </template>
            </div>
          </div>
          <el-empty v-if="!questions.length" description="暂无题目" :image-size="30" />
        </div>
      </div>
    </div>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="saveFormula">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import type { SurveyDesignerQuestion, SurveyFormulaPayload } from '../survey-designer-types'

const props = defineProps<{ questions: SurveyDesignerQuestion[] }>()
const emit = defineEmits<{ save: [payload: SurveyFormulaPayload] }>()
const visible = ref(false)
const title = ref('公式编辑')
const formulaType = ref('')
const formulaText = ref('')
const formulaInputRef = ref<any>(null)
const formulaQuestions = computed(() => props.questions.filter(question => !['divider', 'description', 'pagination'].includes(question.type)))
const placeholder = computed(() => ({
  calculate: '输入计算公式，如：\nSUM(Q1, Q2)\nIFS(Q1<60,"不合格","合格")\nCONCATENATE("总分：", Q1)',
  endFormula: '输入结束规则，如：\nIF Q1A2 THEN BRANCH FROM Q1 TO END',
  jumpFormula: '输入跳转规则，如：\nIF Q1A1 THEN BRANCH FROM Q1 TO Q5',
  requiredFormula: '输入必填规则，如：\nIF Q1A1 THEN REQUIRED Q2',
  validate: '输入校验公式，如：\nIF(Q1>100,"不能超过100","")',
  textReplace: '输入文本替换规则，如：\nREPLACE Q2 WITH CONCATENATE("你好，", Q1)',
  optShow: '输入显示/隐藏规则，如：\nIF Q1A1 THEN SHOW\nIF Q1A2 THEN HIDE',
  optAutoCheck: '输入自动勾选规则，如：\nIF Q1A1 THEN CHECK Q2A1',
}[formulaType.value] || '输入公式表达式'))

function open(type: string, initialText = '') {
  const titles: Record<string, string> = {
    endFormula: '结束公式', jumpFormula: '跳转公式', requiredFormula: '必填公式',
    textReplace: '文本替换', optShow: '显示隐藏公式', optAutoCheck: '自动勾选公式', calculate: '计算公式', validate: '校验规则',
  }
  formulaType.value = type
  formulaText.value = initialText
  title.value = titles[type] || '公式编辑'
  visible.value = true
}

function questionTag(question: SurveyDesignerQuestion) {
  return `Q${props.questions.indexOf(question) + 1}`
}

function insertFormulaTag(tag: string) {
  const textarea = formulaInputRef.value?.textarea as HTMLTextAreaElement | undefined
  if (!textarea) {
    formulaText.value += tag
    return
  }
  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  formulaText.value = formulaText.value.slice(0, start) + tag + formulaText.value.slice(end)
  void nextTick(() => {
    textarea.selectionStart = textarea.selectionEnd = start + tag.length
    textarea.focus()
  })
}

function saveFormula() {
  emit('save', { type: formulaType.value, text: formulaText.value })
  visible.value = false
}

defineExpose({ open })
</script>

<style scoped>
.formula-layout { display:flex; min-height:360px; gap:12px; }
.formula-main { flex:1; min-width:0; }
.formula-label { margin-bottom:4px; color:#888; font-size:12px; }
.formula-help { margin-top:8px; padding:8px 10px; border-radius:4px; background:#f5f7fa; color:#606266; font-size:12px; }
.formula-help ul { margin:0; padding-left:16px; line-height:1.8; }
.help-title, .reference-title { margin-bottom:4px; color:#606266; font-weight:600; }
.primary-text { color:#2563eb; }
.formula-reference { margin-top:8px; color:#999; font-size:11px; }
.formula-examples { margin-top:6px; padding-top:6px; border-top:1px solid #eee; }
.formula-tags { width:200px; flex-shrink:0; }
.tag-list { max-height:320px; overflow-y:auto; padding:8px; border:1px solid #eee; border-radius:4px; }
.tag-question { margin-bottom:6px; }
.tag-question-title { margin-bottom:2px; color:#999; font-size:11px; }
.tag-items { display:flex; flex-wrap:wrap; gap:4px; }
.tag-items :deep(.el-tag) { cursor:pointer; }
</style>
