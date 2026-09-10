<template>
  <div class="setting-wrapper">
    <div class="setting-header">
      <div><div class="setting-page-title">考试配置</div><div class="setting-page-desc">集中管理考试展示、交卷规则、访问限制、投放链接和协作人员</div></div>
      <div class="setting-header-actions"><el-tag size="small" :type="form.id ? 'success' : 'info'">{{ form.id ? '配置可保存' : '保存后生成链接' }}</el-tag><el-button type="primary" size="small" :loading="saving" @click="emit('save')">保存配置</el-button></div>
    </div>
    <div class="setting-scroll">
      <section class="setting-group">
        <div class="group-title">展示与答题体验</div><div class="group-desc">控制考生端页面结构、校验时机和答题辅助能力</div>
        <el-form label-position="top"><div class="settings-grid">
          <el-form-item label="显示题号"><el-switch v-model="form.questionNumber" /></el-form-item><el-form-item label="显示进度条"><el-switch v-model="form.progressBar" /></el-form-item>
          <el-form-item label="自动暂存"><el-switch v-model="form.autoSave" /></el-form-item><el-form-item label="显示分数"><el-switch v-model="form.showScore" :active-value="1" :inactive-value="0" /></el-form-item>
          <el-form-item label="一页一题"><el-switch v-model="form.onePageOneQuestion" /></el-form-item><el-form-item label="显示答题卡"><el-switch v-model="form.answerSheetVisible" /></el-form-item>
          <el-form-item label="显示答案"><el-switch v-model="form.answerVisible" /></el-form-item><el-form-item label="默认语言"><el-select v-model="form.language"><el-option value="zh" label="中文" /><el-option value="en" label="英文" /></el-select></el-form-item>
          <el-form-item label="校验触发"><el-select v-model="form.triggerType"><el-option value="onInput" label="输入时" /><el-option value="onBlur" label="失焦时" /><el-option value="onSubmit" label="提交时" /></el-select></el-form-item>
        </div></el-form>
      </section>
      <section class="setting-group">
        <div class="group-title">考试规则</div><div class="group-desc">设置交卷时间、练习模式、题目顺序和考试次数</div>
        <el-form label-position="top"><div class="settings-grid">
          <el-form-item label="最短交卷时间"><number-setting v-model="form.minSubmitMinutes" help="0 表示不限制最短作答时间" /></el-form-item>
          <el-form-item label="最长交卷时间"><number-setting v-model="form.maxSubmitMinutes" help="单位：分钟，0 表示不限时" /></el-form-item>
          <el-form-item label="每人考试次数"><number-setting v-model="form.maxAttempts" help="0 表示不限次数" /></el-form-item>
          <el-form-item label="练习模式"><el-switch v-model="form.exerciseMode" /></el-form-item><el-form-item label="随机顺序"><el-switch v-model="form.randomOrder" /></el-form-item>
          <el-form-item label="开始时间"><el-date-picker v-model="form.startDate" type="datetime" placeholder="不限" value-format="x" /></el-form-item><el-form-item label="结束时间"><el-date-picker v-model="form.endDate" type="datetime" placeholder="不限" value-format="x" /></el-form-item>
        </div></el-form>
      </section>
      <section class="setting-group">
        <div class="group-title">访问与回收</div><div class="group-desc">控制谁可以参加考试，以及答卷、设备和 IP 限制</div>
        <el-form label-position="top"><div class="settings-grid">
          <el-form-item class="settings-full" label="可见性"><el-radio-group v-model="form.visibility" class="setting-radio-group"><el-radio :value="0" border>公开链接</el-radio><el-radio :value="1" border>登录可见</el-radio><el-radio :value="2" border>部门限定</el-radio></el-radio-group></el-form-item>
          <el-form-item v-if="form.visibility===2" class="settings-full" label="限定部门"><div class="setting-tree-box"><el-tree ref="deptTreeRef" :data="deptTreeOptions" :props="{ label: 'name', children: 'children' }" node-key="id" show-checkbox default-expand-all @check-change="onDeptCheckChange" /></div></el-form-item>
          <el-form-item label="最大答卷数"><number-setting v-model="form.maxResponse" help="0 表示不限总答卷数" /></el-form-item><el-form-item label="填写密码"><el-input v-model="form.password" placeholder="留空不设密码" /></el-form-item>
          <el-form-item label="每台设备答题次数"><number-setting v-model="form.deviceLimit" help="0 表示不限" /></el-form-item><el-form-item label="每个 IP 答题次数"><number-setting v-model="form.ipLimit" help="0 表示不限" /></el-form-item>
        </div></el-form>
      </section>
      <section class="setting-group">
        <div class="group-title">投放与结果</div><div class="group-desc">管理考后展示、填写模板、公开链接和二维码</div>
        <el-form label-position="top"><div class="settings-grid">
          <el-form-item label="显示排名"><el-switch v-model="form.examRankingEnabled" /></el-form-item><el-form-item label="显示成绩单"><el-switch v-model="form.transcriptVisible" /></el-form-item><el-form-item label="可查看正确答案和解析"><el-switch v-model="form.showAnalysis" /></el-form-item><el-form-item label="考试状态"><el-switch v-model="form.statusBool" :active-value="1" :inactive-value="0" active-text="已开启" inactive-text="已停用" /></el-form-item>
          <el-form-item class="settings-full" label="填写模版"><el-select v-model="form.fillTemplate"><el-option v-for="(label, value) in fillTemplates" :key="value" :value="value" :label="label" /></el-select></el-form-item>
          <el-form-item class="settings-full" label="考试链接"><el-input v-if="form.id" :model-value="publicUrl" readonly><template #append><el-button @click="emit('copyLink')">复制</el-button></template></el-input><span v-else class="setting-empty-tip">请先保存考试</span></el-form-item>
          <el-form-item class="settings-full" label="二维码"><div v-if="form.id" class="qr-preview"><img :src="qrUrl" alt="二维码" /><div class="qr-preview-meta"><strong>扫码进入考试</strong><span>二维码内容随考试链接自动生成</span></div></div><span v-else class="setting-empty-tip">请先保存考试</span></el-form-item>
          <el-form-item class="settings-full" label="提交后处理"><el-radio-group v-model="completionAction" class="completion-radio-group"><el-radio-button value="default">默认完成页</el-radio-button><el-radio-button value="content">自定义页面</el-radio-button><el-radio-button value="redirect">跳转链接</el-radio-button></el-radio-group></el-form-item>
          <el-form-item v-if="completionAction==='content'" label="提交完成页内容"><el-input v-model="form.endContent" type="textarea" :rows="3" placeholder="提交后显示的 HTML 内容" /></el-form-item><el-form-item v-if="completionAction==='redirect'" label="提交后跳转链接"><el-input v-model="form.redirectUrl" placeholder="https://example.com 或站内路径" /><div v-if="form.redirectUrl && !isValidRedirectUrl(form.redirectUrl)" class="setting-help danger">请输入 http(s) 链接或以 / 开头的站内路径</div></el-form-item>
        </div></el-form>
      </section>
      <section class="setting-group">
        <div class="group-title">协作管理员</div><div class="group-desc">查看创建者，并指定可以共同维护考试的管理员</div>
        <el-form label-position="top"><el-form-item label="考试创建者"><el-input :model-value="creatorName" disabled placeholder="加载中..." /></el-form-item><el-form-item class="settings-full" label="协作管理员"><div class="setting-tree-box large"><el-tree ref="adminTreeRef" :data="adminTreeData" :props="{ label: 'label', children: 'children' }" node-key="id" show-checkbox :default-checked-keys="collaboratorCheckedKeys" @check="onAdminTreeCheck" /></div></el-form-item></el-form>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, nextTick, ref, watch } from 'vue'
import { ElInputNumber } from 'element-plus'
import { adminApi } from '../../../api'
import { isValidRedirectUrl } from '../../survey/survey-designer-validation'
import type { ExamCompletionAction } from '../exam-designer-types'

const props = defineProps<{ saving: boolean; publicUrl: string; fillTemplates: Record<string, string> }>()
const form = defineModel<Record<string, any>>('form', { required: true })
const completionAction = defineModel<ExamCompletionAction>({ required: true })
const emit = defineEmits<{ save: []; copyLink: [] }>()
const deptTreeRef = ref<any>(null), adminTreeRef = ref<any>(null)
const deptTreeOptions = ref<any[]>([]), adminTreeData = ref<any[]>([])
const collaboratorCheckedKeys = ref<string[]>([]), creatorName = ref('')
const qrUrl = computed(() => props.publicUrl ? `https://api.qrserver.com/v1/create-qr-code/?size=120x120&data=${encodeURIComponent(props.publicUrl)}` : '')
const NumberSetting = defineComponent({ props: { modelValue: Number, help: String }, emits: ['update:modelValue'], setup(componentProps, { emit: componentEmit }) { return () => h('div', { class: 'setting-control-stack' }, [h(ElInputNumber, { modelValue: componentProps.modelValue, min: 0, 'onUpdate:modelValue': (value: number | undefined) => componentEmit('update:modelValue', value) }), h('span', { class: 'setting-help' }, componentProps.help)]) } })
function onAdminTreeCheck() { form.value.collaborators = adminTreeRef.value?.getCheckedKeys(true)?.filter((key: string) => key.startsWith('admin-'))?.map((key: string) => key.replace('admin-', ''))?.join(',') || '' }
function onDeptCheckChange() { form.value.deptIds = deptTreeRef.value?.getCheckedKeys(false).join(',') || '' }
function buildAdminTree(departments: any[], managers: any[]): any[] { return departments.map(department => { const managerNodes = managers.filter(manager => (manager.deptIds || []).includes(department.id)).map(manager => ({ id: `admin-${manager.id}`, label: manager.name, type: 'admin' })); const children = department.children?.length ? buildAdminTree(department.children, managers) : []; children.push(...managerNodes); return { id: `dept-${department.id}`, label: department.name, type: 'dept', children } }).filter(department => department.children?.length) }
async function loadData() {
  await loadDepartments()
  await loadManagers()
}
async function loadDepartments() {
  try {
    const response: any = await adminApi.deptTree()
    deptTreeOptions.value = response.data || []
    await nextTick()
    if (form.value.visibility === 2) deptTreeRef.value?.setCheckedKeys(String(form.value.deptIds || '').split(',').map(Number).filter(Boolean))
  } catch { deptTreeOptions.value = [] }
}
async function loadManagers() {
  try {
    const response: any = await adminApi.mgrList({ page: 1, pageSize: 9999 })
    const managers = response.data?.list || []
    adminTreeData.value = buildAdminTree(deptTreeOptions.value, managers)
    creatorName.value = managers.find((manager: any) => manager.id === form.value.createBy)?.name || (form.value.createBy ? '未知用户' : '未指定')
    collaboratorCheckedKeys.value = String(form.value.collaborators || '').split(',').filter(Boolean).map(id => `admin-${id}`)
    await nextTick()
    adminTreeRef.value?.setCheckedKeys(collaboratorCheckedKeys.value)
  } catch { creatorName.value = '加载失败'; adminTreeData.value = [] }
}
watch(() => form.value.visibility, value => { if (value === 2) void loadData() })
defineExpose({ loadData })
</script>

<style scoped>
.setting-wrapper { height:100%; overflow-y:auto; background:var(--designer-bg); }.setting-header { position:sticky; top:0; z-index:5; display:flex; align-items:center; justify-content:space-between; gap:16px; padding:18px 24px; border-bottom:1px solid var(--designer-border); background:#fff; }.setting-page-title { font-size:18px; font-weight:700; }.setting-page-desc,.group-desc,.setting-help,.setting-empty-tip { color:#98a2b3; font-size:12px; }.setting-header-actions { display:flex; gap:10px; }.setting-scroll { display:grid; grid-template-columns:repeat(auto-fill,minmax(360px,1fr)); align-items:start; gap:18px; padding:24px; }.setting-group { padding:18px 20px; border:1px solid var(--designer-border); border-radius:8px; background:#fff; }.group-title { margin-bottom:4px; font-size:14px; font-weight:650; }.group-desc { margin-bottom:14px; }.settings-grid { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:0 16px; }.settings-full { grid-column:1/-1; }.settings-grid :deep(.el-select),.settings-grid :deep(.el-date-editor),.settings-grid :deep(.el-input-number) { width:100%; }.setting-radio-group { display:flex; flex-wrap:wrap; }.setting-control-stack { display:flex; flex-direction:column; width:100%; gap:4px; }.setting-tree-box { width:100%; max-height:300px; overflow:auto; padding:8px; border:1px solid #dcdfe6; border-radius:7px; }.setting-tree-box.large { max-height:400px; }.completion-radio-group { display:flex; width:100%; }.completion-radio-group :deep(.el-radio-button) { flex:1; }.qr-preview { display:flex; align-items:center; gap:12px; }.qr-preview img { width:120px; height:120px; }.qr-preview-meta { display:flex; flex-direction:column; gap:4px; }.setting-help.danger { color:#f04438; }
</style>
