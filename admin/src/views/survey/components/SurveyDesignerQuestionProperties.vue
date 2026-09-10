<template>
  <div class="survey-setting-panel" @click.self="deselectQuestion">
    <div v-if="selected" class="props-panel">
      <template v-if="selectedOptIdx >= 0 && selected.type === 'file'">
        <SurveyDesignerPanelTitle title="文件上传设置" :context="context" />
        <el-form label-position="top" size="small">
          <el-form-item label="允许类型">
            <el-checkbox-group v-model="selected.fileTypes">
              <el-checkbox value="image">图片</el-checkbox>
              <el-checkbox value="video">视频</el-checkbox>
              <el-checkbox value="audio">音频</el-checkbox>
              <el-checkbox value="doc">文档</el-checkbox>
              <el-checkbox value="other">其他</el-checkbox>
            </el-checkbox-group>
          </el-form-item>
          <el-form-item label="自定义扩展名"><el-input v-model="selected.fileExtensions" placeholder="jpg,png,pdf 逗号分隔" /></el-form-item>
          <el-row :gutter="8">
            <el-col :span="12"><el-form-item label="单文件限制"><el-input-number v-model="selected.maxFileSize" :min="0" style="width:100%" /><span class="field-unit">MB</span></el-form-item></el-col>
            <el-col :span="12"><el-form-item label="文件数量"><el-input-number v-model="selected.maxFileCount" :min="1" :max="99" style="width:100%" /></el-form-item></el-col>
          </el-row>
          <BackButton @click="selectedOptIdx = -1" />
        </el-form>
      </template>

      <template v-else-if="selectedOptIdx >= 0 && isMatrixAll(selected.type) && selected.props?.rows?.[selectedOptIdx]">
        <SurveyDesignerPanelTitle title="矩阵行设置" :context="context" />
        <el-form label-position="top" size="small">
          <el-form-item label="行名"><el-input v-model="selected.props.rows[selectedOptIdx].title" /></el-form-item>
          <el-form-item label="行ID"><el-input v-model="selected.props.rows[selectedOptIdx].id" disabled /></el-form-item>
          <BackButton @click="selectedOptIdx = -1" />
        </el-form>
      </template>

      <template v-else-if="selectedOptIdx >= 0 && (selected.type === 'user' || selected.type === 'dept')">
        <SurveyDesignerPanelTitle :title="`${selected.type === 'user' ? '成员' : '部门'}设置`" :context="context" />
        <el-form label-position="top" size="small">
          <template v-if="selected.type === 'user'">
            <el-form-item label="成员名称"><el-input v-model="selected.props.options[selectedOptIdx].label" placeholder="输入成员名称" /></el-form-item>
            <el-form-item label="成员ID"><el-input v-model="selected.props.options[selectedOptIdx].value" placeholder="输入成员ID" /></el-form-item>
            <el-form-item label="部门名称"><el-input v-model="selected.props.options[selectedOptIdx].deptName" placeholder="输入部门名称" /></el-form-item>
            <el-form-item label="部门ID"><el-input v-model="selected.props.options[selectedOptIdx].deptId" placeholder="输入部门ID" /></el-form-item>
            <el-form-item label="父级部门ID"><el-input v-model="selected.props.options[selectedOptIdx].parentDeptId" placeholder="输入父级部门ID" /></el-form-item>
          </template>
          <template v-else>
            <el-form-item label="部门名称"><el-input v-model="selected.props.options[selectedOptIdx].label" placeholder="输入部门名称" /></el-form-item>
            <el-form-item label="部门ID"><el-input v-model="selected.props.options[selectedOptIdx].value" placeholder="输入部门ID" /></el-form-item>
            <el-form-item label="父级部门ID"><el-input v-model="selected.props.options[selectedOptIdx].parentId" placeholder="输入父级部门ID" /></el-form-item>
          </template>
          <BackButton @click="selectedOptIdx = -1" />
        </el-form>
      </template>

      <template v-else-if="selectedOptIdx >= 0 && selected.props?.options?.[selectedOptIdx]">
        <SurveyDesignerPanelTitle :title="`选项设置 - ${stripHtml(selected.props.options[selectedOptIdx].label || '')}`" :context="context" />
        <el-form label-position="top" size="small">
          <el-form-item label="选项值"><el-input v-model="selected.props.options[selectedOptIdx].value" placeholder="默认同选项名称" /></el-form-item>
          <template v-if="hasOptions(selected)">
            <el-form-item label="选项配额"><el-input-number v-model="selected.props.options[selectedOptIdx].quota" :min="0" style="width:100%" placeholder="0=不限制" /></el-form-item>
            <FormulaSetting label="显示隐藏公式" @click="emit('openFormula', 'optShow')" />
            <FormulaSetting label="自动勾选公式" @click="emit('openFormula', 'optAutoCheck')" />
          </template>
          <template v-if="isInput(selected.type)">
            <el-form-item label="必填"><el-switch v-model="selected.props.options[selectedOptIdx].required" /></el-form-item>
            <el-form-item label="只读"><el-switch v-model="selected.props.options[selectedOptIdx].readOnly" /></el-form-item>
            <el-form-item label="计算公式"><el-input v-model="selected.props.options[selectedOptIdx].calculate" placeholder="#{qId} 引用其他题目" /></el-form-item>
            <el-form-item label="内容限制"><DataTypeSelect v-model="selected.props.options[selectedOptIdx].dataType" /></el-form-item>
            <el-form-item v-if="selected.props.options[selectedOptIdx].dataType === 'number'" label="小数位数"><el-input-number v-model="selected.props.options[selectedOptIdx].decimalPlaces" :min="0" :max="6" style="width:100%" /></el-form-item>
            <el-form-item label="最少填写"><el-input-number v-model="selected.props.options[selectedOptIdx].minLength" :min="0" style="width:100%" /></el-form-item>
            <el-form-item label="最多填写"><el-input-number v-model="selected.props.options[selectedOptIdx].maxLength" :min="0" style="width:100%" /></el-form-item>
            <el-form-item label="后缀文字"><el-input v-model="selected.props.options[selectedOptIdx].suffix" placeholder="如 cm / kg" /></el-form-item>
            <el-form-item label="答案唯一"><el-switch v-model="selected.props.options[selectedOptIdx].unique" /></el-form-item>
            <el-form-item label="提示语"><el-input v-model="selected.props.options[selectedOptIdx].placeholder" /></el-form-item>
          </template>
          <BackButton @click="selectedOptIdx = -1" />
        </el-form>
      </template>

      <template v-else-if="selectedOptIdx >= 0 && (selected.type === 'multiInput' || selected.type === 'hInput') && selected.props?.fields?.[selectedOptIdx]">
        <SurveyDesignerPanelTitle :title="`字段设置 - ${selected.props.fields[selectedOptIdx].label}`" :context="context" />
        <el-form label-position="top" size="small">
          <el-form-item label="字段名"><el-input v-model="selected.props.fields[selectedOptIdx].label" placeholder="输入字段名" /></el-form-item>
          <el-form-item label="占位提示"><el-input v-model="selected.props.fields[selectedOptIdx].placeholder" placeholder="输入占位提示" /></el-form-item>
          <el-form-item label="数据类型"><DataTypeSelect v-model="selected.props.fields[selectedOptIdx].dataType" /></el-form-item>
          <el-form-item v-if="selected.props.fields[selectedOptIdx].dataType === 'number'" label="小数位数"><el-input-number v-model="selected.props.fields[selectedOptIdx].decimalPlaces" :min="0" :max="6" style="width:100%" /></el-form-item>
          <BackButton @click="selectedOptIdx = -1" />
        </el-form>
      </template>

      <template v-else-if="selectedOptIdx >= 0 && (selected.type === 'rating' || selected.type === 'nps')">
        <SurveyDesignerPanelTitle :title="`${selected.type === 'rating' ? '评分' : 'NPS'}设置`" :context="context" />
        <el-form label-position="top" size="small">
          <el-form-item label="最大分值"><el-input-number v-model="selected.props.maxRating" :min="2" :max="10" :step="1" style="width:100%" /></el-form-item>
          <el-form-item v-if="selected.type === 'rating'" label="图标类型">
            <el-radio-group v-model="selected.props.icon"><el-radio value="star">星标</el-radio><el-radio value="heart">爱心</el-radio><el-radio value="smiley">笑脸</el-radio></el-radio-group>
          </el-form-item>
        </el-form>
        <BackButton @click="selectedOptIdx = -1" />
        <el-button type="danger" text class="delete-button" @click="emit('remove')">删除此题</el-button>
      </template>

      <template v-else>
        <SurveyDesignerPanelTitle :title="`${typeName(selected.type)}设置`" :context="context" />
        <el-form label-position="top" size="small">
          <el-form-item label="ID"><el-input v-model="selected.id" :disabled="!!selected._existing" /></el-form-item>
          <el-row v-if="!isPureLayout(selected.type) && selected.type !== 'questionSet'" :gutter="8">
            <el-col :span="12"><el-form-item label="必填"><el-switch v-model="selected.required" /></el-form-item></el-col>
            <el-col :span="12"><el-form-item label="只读"><el-switch v-model="selected.readOnly" /></el-form-item></el-col>
          </el-row>
          <el-form-item v-if="!isPureLayout(selected.type) && selected.type !== 'questionSet'" label="说明"><el-switch v-model="selected.showDescription" /></el-form-item>

          <template v-if="!isPureLayout(selected.type)">
            <FormulaSetting label="结束公式" @click="emit('openFormula', 'endFormula')" />
            <FormulaSetting label="跳转公式" @click="emit('openFormula', 'jumpFormula')" />
            <el-row :gutter="8">
              <el-col :span="12"><el-form-item label="默认隐藏"><el-switch v-model="selected.defaultHidden" /></el-form-item></el-col>
              <el-col v-if="selected.type === 'user' || selected.type === 'dept'" :span="12"><el-form-item label="多选"><el-switch v-model="selected.multiple" /></el-form-item></el-col>
              <el-col v-if="hasOptions(selected) && !isMatrixAll(selected.type)" :span="12"><el-form-item label="选项布局"><el-input-number v-model="selected.optionLayout" :min="1" :max="6" :step="1" style="width:100%" /></el-form-item></el-col>
            </el-row>
            <FormulaSetting label="必填公式" @click="emit('openFormula', 'requiredFormula')" />
            <FormulaSetting label="文本替换" @click="emit('openFormula', 'textReplace')" />
            <FormulaSetting v-if="['input','textarea','number','date','time','dateRange','calculate'].includes(selected.type) || isPersonal(selected.type)" label="计算公式" @click="emit('openFormula', 'calculate')" />
            <el-form-item v-if="isInput(selected.type) || isPersonal(selected.type) || selected.type === 'signature' || selected.type === 'scanCode'" label="占位提示"><el-input v-model="selected.placeholder" /></el-form-item>
            <el-form-item v-if="selected.type === 'richText'" label="占位提示"><el-input v-model="selected.placeholder" placeholder="输入富文本编辑器占位提示" /></el-form-item>
          </template>
          <template v-else-if="selected.type === 'description' || selected.type === 'questionSet'">
            <FormulaSetting label="文本替换" @click="emit('openFormula', 'textReplace')" />
            <el-form-item v-if="selected.type === 'description'" label="文本替换规则"><el-input v-model="selected.props.replaceTextRule" type="textarea" :rows="2" placeholder="@公式" /></el-form-item>
          </template>

          <div v-if="selected.type === 'matrixAuto' && selected.props" class="props-options-section">
            <h4>列定义</h4>
            <div v-for="(col, ci) in (selected.props.columns || [])" :key="ci" class="setting-opt-row">
              <el-input v-model="col.label" size="small" placeholder="列名" style="flex:1" />
              <el-select v-model="col.type" size="small" style="width:90px"><el-option label="文本" value="input" /><el-option label="数字" value="number" /><el-option label="多行" value="textarea" /><el-option label="日期" value="date" /></el-select>
              <el-button text size="small" type="danger" @click="removeAutoCol(Number(ci))">×</el-button>
            </div>
            <el-button size="small" plain @click="addAutoCol">+ 添加列</el-button>
            <el-row :gutter="8" style="margin-top:8px">
              <el-col :span="12"><el-form-item label="最少行数"><el-input-number v-model="selected.props.minRows" :min="0" style="width:100%" /></el-form-item></el-col>
              <el-col :span="12"><el-form-item label="最多行数"><el-input-number v-model="selected.props.maxRows" :min="0" style="width:100%" /></el-form-item></el-col>
            </el-row>
          </div>

          <div v-if="selected.type === 'number' && selected.props" class="props-options-section">
            <h4>数值范围</h4>
            <el-row :gutter="8">
              <el-col :span="12"><el-form-item label="最小值"><el-input-number v-model="selected.props.min" :step="1" style="width:100%" /></el-form-item></el-col>
              <el-col :span="12"><el-form-item label="最大值"><el-input-number v-model="selected.props.max" :step="1" style="width:100%" /></el-form-item></el-col>
            </el-row>
            <el-form-item label="小数位数"><el-input-number v-model="selected.props.decimalPlaces" :min="0" :max="6" style="width:100%" /></el-form-item>
          </div>

          <div v-if="selected.type === 'pagination'" class="props-options-section">
            <h4>分页</h4>
            <el-row :gutter="8">
              <el-col :span="12"><el-form-item label="当前页"><el-input-number v-model="selected.props.currentPage" :min="1" style="width:100%" /></el-form-item></el-col>
              <el-col :span="12"><el-form-item label="总页数"><el-input-number v-model="selected.props.totalPage" :min="1" style="width:100%" /></el-form-item></el-col>
            </el-row>
          </div>

          <div v-if="(selected.type === 'user' || selected.type === 'dept') && selected.props" class="props-options-section">
            <div class="option-list-title">
              <h4>{{ selected.type === 'user' ? '成员' : '部门' }}列表</h4>
              <el-tooltip :content="importFormatHint" placement="right"><el-button text size="small" class="format-button">格式</el-button></el-tooltip>
            </div>
            <div v-for="(option, i) in (selected.props.options || [])" :key="i" class="setting-opt-row">
              <span class="option-summary">
                {{ option.label || (selected.type === 'user' ? '成员' : '部门') }}
                <span v-if="selected.type === 'user' && option.deptName">({{ option.deptName }})</span>
                <span v-if="selected.type === 'user' && option.parentDeptId">父级:{{ option.parentDeptId }}</span>
                <span v-if="selected.type === 'dept' && option.parentId">父级:{{ option.parentId }}</span>
              </span>
              <el-button text size="small" type="primary" @click="selectedOptIdx = Number(i)">编辑</el-button>
              <el-button text size="small" type="danger" @click="removeUserOption(Number(i))">×</el-button>
            </div>
            <div class="option-actions"><el-button size="small" plain @click="addUserOption">{{ selected.type === 'user' ? '+ 添加成员' : '+ 添加部门' }}</el-button><el-button size="small" plain @click="importUserOptions">导入</el-button></div>
          </div>

          <el-form-item v-if="selected.type === 'signature'" label="考试答案模式">
            <el-select v-model="selected.examAnswerMode" clearable placeholder="不限制" style="width:100%"><el-option label="不限制" value="" /><el-option label="无" value="none" /><el-option label="文字" value="text" /><el-option label="图片" value="image" /></el-select>
          </el-form-item>
          <el-form-item v-if="hasDataType(selected)" label="数据类型"><FullDataTypeSelect v-model="selected.dataType" /></el-form-item>
          <el-collapse v-if="!isPureLayout(selected.type) && form.mode === 'exam'" v-model="collapseActive" accordion>
            <el-collapse-item title="考试属性" name="exam">
              <el-form-item label="分值"><el-input-number v-model="selected.examScore" :min="0" :step="0.5" style="width:100%" /></el-form-item>
              <el-form-item v-if="selected.examScore" label="正确答案"><el-input v-model="selected.examCorrectAnswer" placeholder="填写正确答案" /></el-form-item>
              <el-form-item v-if="selected.examScore" label="答案解析"><el-input v-model="selected.examAnalysis" type="textarea" :rows="2" placeholder="选填" /></el-form-item>
            </el-collapse-item>
          </el-collapse>
        </el-form>
        <div class="props-footer"><el-button type="danger" text class="delete-button" @click="emit('remove')">删除此题</el-button></div>
      </template>
    </div>

    <div v-else-if="showSurveySettings" class="props-panel">
      <SurveyDesignerPanelTitle title="问卷设置" :context="form.title || '未命名问卷'" />
      <div class="setting-row"><span class="setting-label">标题</span><el-input v-model="form.title" placeholder="问卷标题" style="flex:1" /></div>
      <div class="setting-row description-row"><span class="setting-label">描述</span><el-input v-model="form.description" type="textarea" :autosize="{ minRows: 3, maxRows: 8 }" placeholder="问卷描述" style="flex:1" /></div>
      <div class="setting-row"><span class="setting-label">模式</span><span>{{ form.mode === 'exam' ? '考试' : '问卷' }}</span></div>
      <el-divider />
      <div class="group-title">已应用图片</div>
      <div v-for="(item, i) in form.backgroundImages" :key="`bg-${i}`" class="applied-image-row"><img :src="typeof item === 'string' ? item : item.url" class="applied-image-thumb" /><span class="applied-image-label">背景图</span><button class="applied-image-remove" @click="removeAppliedImage('backgroundImages')">✕</button></div>
      <div v-for="(item, i) in form.headerImages" :key="`hd-${i}`" class="applied-image-row"><img :src="typeof item === 'string' ? item : item.url" class="applied-image-thumb" /><span class="applied-image-label">页眉图</span><button class="applied-image-remove" @click="removeAppliedImage('headerImages')">✕</button></div>
      <el-empty v-if="!form.backgroundImages.length && !form.headerImages.length" description="暂无已应用图片" :image-size="30" />
    </div>
    <div v-else class="props-panel props-empty-panel"><el-empty description="未选择题目" :image-size="60"><el-button size="small" type="primary" @click="openSurveySettings">问卷设置</el-button></el-empty></div>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, ref } from 'vue'
import { ElButton, ElMessage, ElOption, ElSelect } from 'element-plus'
import SurveyDesignerPanelTitle from './SurveyDesignerPanelTitle.vue'
import type { SurveyDesignerQuestion } from '../survey-designer-types'

const props = defineProps<{ form: Record<string, any>; context: string; typeName: (type: string) => string; generateId: () => string }>()
const selected = defineModel<SurveyDesignerQuestion | null>({ required: true })
const selectedOptIdx = defineModel<number>('selectedOptIdx', { required: true })
const showSurveySettings = defineModel<boolean>('showSurveySettings', { required: true })
const emit = defineEmits<{ remove: []; openFormula: [type: string]; save: [] }>()
const form = computed(() => props.form)
const collapseActive = ref('')

const BackButton = defineComponent({ emits: ['click'], setup(_, { emit: click }) { return () => h(ElButton, { text: true, size: 'small', class: 'back-button', onClick: () => click('click') }, () => '← 返回题目设置') } })
const FormulaSetting = defineComponent({ props: { label: { type: String, required: true } }, emits: ['click'], setup(componentProps, { emit: click }) { return () => h('div', { class: 'setting-row' }, [h('span', { class: 'setting-label' }, componentProps.label), h(ElButton, { text: true, size: 'small', onClick: () => click('click') }, () => '点击设置')]) } })
const DataTypeSelect = createDataTypeSelect([['不限制', ''], ['数字', 'number'], ['手机号', 'mobile'], ['邮箱', 'email'], ['身份证', 'idCard']])
const FullDataTypeSelect = createDataTypeSelect([['不限制', ''], ['数字', 'number'], ['邮箱', 'email'], ['手机号', 'mobile'], ['身份证', 'idCard'], ['日期', 'date'], ['日期时间', 'dateTime'], ['时间', 'time'], ['中文', 'chinese'], ['字母', 'alphabet']])

function createDataTypeSelect(options: string[][]) {
  return defineComponent({ props: { modelValue: String }, emits: ['update:modelValue'], setup(selectProps, { emit: update }) { return () => h(ElSelect, { modelValue: selectProps.modelValue, placeholder: '不限制', clearable: true, style: 'width:100%', 'onUpdate:modelValue': value => update('update:modelValue', value) }, () => options.map(([label, value]) => h(ElOption, { label, value }))) } })
}
function stripHtml(value: string) { return value.replace(/<[^>]*>/g, '') }
function hasOptions(question: SurveyDesignerQuestion) { return ['select', 'radio', 'checkbox', 'picker', 'matrixRadio', 'matrixCheckbox', 'judge', 'cascade'].includes(question.type) }
function hasDataType(question: SurveyDesignerQuestion) { return ['input', 'textarea', 'number', 'name', 'studentId', 'employeeId', 'class', 'phone', 'email', 'idCard', 'password', 'date', 'time', 'dateRange', 'switch', 'location'].includes(question.type) }
function isPureLayout(type: string) { return ['divider', 'description', 'pagination'].includes(type) }
function isInput(type: string) { return ['input', 'textarea', 'number', 'multiInput', 'hInput'].includes(type) }
function isMatrixAll(type: string) { return ['matrixRadio', 'matrixCheckbox', 'matrixFillBlank', 'matrixAuto'].includes(type) }
function isPersonal(type: string) { return ['name', 'studentId', 'employeeId', 'class', 'phone', 'email', 'idCard', 'password', 'date', 'time', 'dateRange', 'switch', 'location'].includes(type) }
const importFormatHint = computed(() => selected.value?.type === 'user'
  ? 'JSON 格式: [{"label":"姓名","value":"ID","deptName":"部门","deptId":"部门ID","parentDeptId":"父级部门ID"}]'
  : 'JSON 格式: [{"label":"部门名称","value":"部门ID","parentId":"父级部门ID"}]')

function addAutoCol() {
  if (!selected.value?.props) return
  if (!Array.isArray(selected.value.props.columns)) selected.value.props.columns = []
  selected.value.props.columns.push({ label: `字段${selected.value.props.columns.length + 1}`, id: props.generateId(), width: 150, type: 'input' })
}
function removeAutoCol(index: number) { selected.value?.props?.columns?.splice(index, 1) }
function addUserOption() {
  if (!selected.value?.props) return
  if (!Array.isArray(selected.value.props.options)) selected.value.props.options = []
  const number = selected.value.props.options.length + 1
  selected.value.props.options.push(selected.value.type === 'user'
    ? { label: `成员${number}`, value: `member${number}`, deptName: '', deptId: '', parentDeptId: '' }
    : { label: `部门${number}`, value: `dept${number}`, parentId: '' })
}
function removeUserOption(index: number) { selected.value?.props?.options?.splice(index, 1) }
function importUserOptions() {
  if (!selected.value) return
  const question = selected.value
  const input = document.createElement('input')
  input.type = 'file'; input.accept = '.json'
  input.onchange = () => {
    const file = input.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = () => {
      try {
        const data = JSON.parse(String(reader.result || ''))
        if (!Array.isArray(data)) return ElMessage.error('JSON 格式错误，应为数组')
        if (!Array.isArray(question.props.options)) question.props.options = []
        data.filter(item => item.label || item.value).forEach(item => question.props.options.push(question.type === 'user'
          ? { label: item.label || item.name || '', value: item.value || item.id || '', deptName: item.deptName || item.department || '', deptId: item.deptId || item.departmentId || '', parentDeptId: item.parentDeptId || '' }
          : { label: item.label || item.name || '', value: item.value || item.id || '', parentId: item.parentId || '' }))
        ElMessage.success(`成功导入 ${data.filter(item => item.label || item.value).length} 条`)
      } catch { ElMessage.error('JSON 解析失败') }
    }
    reader.readAsText(file)
  }
  input.click()
}
function removeAppliedImage(field: 'backgroundImages' | 'headerImages') { form.value[field] = []; emit('save') }
function deselectQuestion() { selected.value = null; showSurveySettings.value = false }
function openSurveySettings() { selected.value = null; showSurveySettings.value = true }
</script>

<style scoped>
.survey-setting-panel { width:380px; flex-shrink:0; border-left:1px solid var(--designer-border); background:#fff; overflow-y:auto; box-shadow:-2px 0 14px rgba(15,23,42,.04); }
.props-panel { padding:14px; min-height:100%; box-sizing:border-box; }
.props-panel :deep(.el-form-item) { margin-bottom:10px; }
.props-panel :deep(.el-form-item__label) { font-size:12px; color:#667085; padding-bottom:4px; font-weight:600; line-height:1.2; }
.props-panel :deep(.el-divider) { margin:12px 0; }
.props-panel :deep(.el-collapse-item__header) { font-size:12px; font-weight:600; }
.props-panel :deep(.el-input__wrapper), .props-panel :deep(.el-textarea__inner), .props-panel :deep(.el-select__wrapper), .props-panel :deep(.el-input-number .el-input__wrapper) { border-radius:7px; }
.props-footer { padding:10px 0 0; border-top:1px solid var(--designer-border); background:#fff; position:sticky; bottom:0; z-index:1; margin-top:12px; }
.setting-row { display:flex; align-items:center; justify-content:space-between; gap:10px; margin-bottom:10px; font-size:12px; color:#667085; }
.setting-label { font-weight:500; }
.description-row { align-items:flex-start; }
.description-row .setting-label { margin-top:4px; }
.props-options-section { margin-bottom:10px; padding:10px; background:#f8fafc; border:1px solid #edf0f5; border-radius:8px; }
.props-options-section h4, .group-title { font-size:12px; color:#667085; margin:0 0 6px; font-weight:600; }
.setting-opt-row, .option-list-title, .option-actions { display:flex; align-items:center; gap:6px; margin-bottom:6px; min-width:0; }
.option-list-title h4 { margin:0; }
.format-button { font-size:11px; color:#999; padding:0 4px; }
.option-summary { font-size:12px; flex:1; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.option-summary span { color:#999; margin-left:4px; font-size:11px; }
.props-empty-panel { display:flex; align-items:center; justify-content:center; }
.props-empty-panel :deep(.el-empty) { width:100%; }
.delete-button { width:100%; }
.back-button { margin-top:8px; }
.field-unit { font-size:11px; color:#999; margin-left:4px; }
.applied-image-row { display:flex; align-items:center; gap:8px; margin-bottom:8px; padding:6px; background:#fff; border-radius:6px; border:1px solid #eee; }
.applied-image-thumb { width:40px; height:40px; border-radius:4px; object-fit:cover; flex-shrink:0; }
.applied-image-label { font-size:12px; color:#666; flex:1; }
.applied-image-remove { width:22px; height:22px; border-radius:50%; border:none; background:#fee; color:#fb454c; cursor:pointer; display:flex; align-items:center; justify-content:center; font-size:12px; flex-shrink:0; }
.applied-image-remove:hover { background:#fb454c; color:#fff; }
</style>
