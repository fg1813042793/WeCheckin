<template>
  <div class="setting-wrapper">
    <div class="setting-header">
      <div>
        <div class="setting-page-title">配置中心</div>
        <div class="setting-page-desc">管理问卷展示、回收规则、投放链接和协作人员</div>
      </div>
      <div class="setting-header-actions">
        <el-tag size="small" :type="saveStatusTagType">{{ saveStatusLabel }}</el-tag>
        <el-button type="primary" size="small" :loading="saving" @click="emit('save')">保存配置</el-button>
      </div>
    </div>

    <div class="setting-scroll">
      <section class="setting-group">
        <div class="group-title">问卷显示设置</div>
        <el-form label-position="top">
          <div class="settings-grid">
            <el-form-item label="填写端自动暂存"><el-switch v-model="form.autoSave" /></el-form-item>
            <el-form-item label="显示题目序号"><el-switch v-model="form.questionNumber" /></el-form-item>
            <el-form-item label="显示进度条"><el-switch v-model="form.progressBar" /></el-form-item>
            <el-form-item label="一页一题"><el-switch v-model="form.onePageOneQuestion" /></el-form-item>
            <el-form-item label="显示答题卡"><el-switch v-model="form.answerSheetVisible" /></el-form-item>
          </div>
        </el-form>
      </section>

      <section class="setting-group">
        <div class="group-title">回收设置</div>
        <el-form label-position="top">
          <div class="settings-grid">
            <el-form-item class="settings-full" label="可见性">
              <el-radio-group v-model="form.visibility" class="visibility-options">
                <el-radio :value="0" border>公开链接</el-radio>
                <el-radio :value="1" border>登录可见</el-radio>
                <el-radio :value="2" border>部门限定</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item v-if="form.visibility === 2" class="settings-full" label="限定部门" :class="{ 'is-error': selectedDeptCount === 0 }">
              <div class="setting-tree-box" :class="{ invalid: selectedDeptCount === 0 }">
                <el-tree
                  ref="deptTreeRef"
                  :data="deptTreeData"
                  :props="{ label: 'name', children: 'children' }"
                  node-key="id"
                  show-checkbox
                  check-strictly
                  :default-checked-keys="deptCheckedKeys"
                  @check="onDeptCheck"
                />
              </div>
              <div class="setting-help" :class="{ danger: selectedDeptCount === 0 }">
                {{ selectedDeptCount ? `已选择 ${selectedDeptCount} 个部门` : '部门限定需要至少选择一个部门' }}
              </div>
            </el-form-item>
            <el-form-item label="每台设备答题次数">
              <div class="setting-control-stack">
                <el-input-number v-model="form.deviceLimit" :min="0" />
                <span class="setting-help">0 表示不限</span>
              </div>
            </el-form-item>
            <el-form-item label="每个 IP 答题次数">
              <div class="setting-control-stack">
                <el-input-number v-model="form.ipLimit" :min="0" />
                <span class="setting-help">0 表示不限</span>
              </div>
            </el-form-item>
            <el-form-item label="允许多次填写"><el-switch v-model="form.allowMultiBool" :active-value="1" :inactive-value="0" active-text="允许" inactive-text="仅一次" /></el-form-item>
            <el-form-item label="开始时间"><el-date-picker v-model="form.startDate" type="datetime" placeholder="不限" value-format="x" /></el-form-item>
            <el-form-item label="结束时间">
              <el-date-picker v-model="form.endDate" type="datetime" placeholder="不限" value-format="x" />
              <div v-if="hasInvalidTimeRange" class="setting-help danger">结束时间不能早于开始时间</div>
            </el-form-item>
            <el-form-item label="回收上限">
              <div class="setting-control-stack">
                <el-input-number v-model="form.maxResponse" :min="0" />
                <span class="setting-help">0 表示不限</span>
              </div>
            </el-form-item>
            <el-form-item label="作答时间（分钟）">
              <div class="setting-control-stack">
                <el-input-number v-model="form.timeLimit" :min="0" />
                <span class="setting-help">0 表示不限</span>
              </div>
            </el-form-item>
          </div>
        </el-form>
      </section>

      <section v-if="form.mode === 'exam'" class="setting-group">
        <div class="group-title">考试设置</div>
        <el-form label-position="top">
          <div class="settings-grid">
            <el-form-item label="显示排名"><el-switch v-model="form.examRankingEnabled" /></el-form-item>
            <el-form-item label="练习模式"><el-switch v-model="form.exerciseMode" /></el-form-item>
            <el-form-item label="随机顺序"><el-switch v-model="form.randomOrder" /></el-form-item>
            <el-form-item label="最短交卷(分钟)"><el-input-number v-model="form.minSubmitMinutes" :min="0" /></el-form-item>
            <el-form-item label="最长交卷(分钟)"><el-input-number v-model="form.maxSubmitMinutes" :min="0" /></el-form-item>
          </div>
        </el-form>
      </section>

      <section class="setting-group">
        <div class="group-title">投放与分享</div>
        <el-form label-position="top">
          <el-form-item class="settings-full" label="提交后处理">
            <el-radio-group v-model="completionAction" class="completion-radio-group">
              <el-radio-button value="default">默认完成页</el-radio-button>
              <el-radio-button value="content">自定义页面</el-radio-button>
              <el-radio-button value="redirect">跳转链接</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <div class="settings-grid">
            <el-form-item v-if="completionAction === 'content'" label="自定义完成页"><el-input v-model="form.endContent" type="textarea" :rows="3" placeholder="答题完成后显示的 HTML 内容" /></el-form-item>
            <el-form-item v-if="completionAction === 'redirect'" label="跳转链接">
              <el-input v-model="form.redirectUrl" placeholder="https://example.com 或站内路径" />
              <div v-if="form.redirectUrl && !isValidRedirectUrl(form.redirectUrl)" class="setting-help danger">请输入 http(s) 链接或以 / 开头的站内路径</div>
            </el-form-item>
          </div>
          <el-form-item class="settings-full" label="填写模版">
            <el-select v-model="form.fillTemplate">
              <el-option v-for="(label, value) in fillTemplates" :key="value" :value="value" :label="label" />
            </el-select>
          </el-form-item>
          <el-form-item class="settings-full" label="问卷链接">
            <div class="share-settings">
              <el-input v-if="form.id" :model-value="publicUrl" readonly><template #append><el-button @click="emit('copyLink')">复制</el-button></template></el-input>
              <span v-else class="setting-help">请先保存问卷</span>
              <div v-if="form.id" class="qr-settings">
                <el-image v-if="qrUrl" :src="qrUrl" class="qr-image" />
                <div class="qr-actions">
                  <el-button size="small" @click="emit('generateQr')">生成二维码</el-button>
                  <el-button v-if="qrUrl" size="small" @click="emit('downloadQr')">下载二维码</el-button>
                </div>
              </div>
            </div>
          </el-form-item>
          <div class="settings-grid">
            <el-form-item label="问卷状态"><el-switch v-model="form.statusBool" :active-value="1" :inactive-value="0" active-text="已开启" inactive-text="已停用" /></el-form-item>
          </div>
        </el-form>
      </section>

      <section class="setting-group">
        <div class="group-title">协作管理员</div>
        <el-form label-position="top">
          <div class="settings-grid">
            <el-form-item label="问卷创建者"><el-input :model-value="creatorName" disabled placeholder="加载中..." /></el-form-item>
          </div>
          <el-form-item class="settings-full" label="协作管理员">
            <div class="setting-tree-box admin-tree-box">
              <el-tree
                ref="adminTreeRef"
                :data="adminTreeData"
                :props="{ label: 'label', children: 'children' }"
                node-key="id"
                show-checkbox
                :default-checked-keys="collaboratorCheckedKeys"
                @check="onAdminTreeCheck"
              />
            </div>
          </el-form-item>
        </el-form>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { adminApi } from '../../../api'
import { isValidRedirectUrl } from '../survey-designer-validation'
import type { SurveyCompletionAction } from '../survey-designer-types'

const props = defineProps<{
  active: boolean
  saving: boolean
  saveStatusLabel: string
  saveStatusTagType: 'success' | 'warning' | 'danger' | 'info'
  publicUrl: string
  qrUrl: string
  fillTemplates: Record<string, string>
  hasInvalidTimeRange: boolean
}>()

const emit = defineEmits<{
  save: []
  copyLink: []
  generateQr: []
  downloadQr: []
}>()
const form = defineModel<Record<string, any>>('form', { required: true })
const completionAction = defineModel<SurveyCompletionAction>({ required: true })

const deptTreeRef = ref<any>(null)
const deptTreeData = ref<any[]>([])
const deptCheckedKeys = ref<number[]>([])
const creatorName = ref('')
const adminTreeRef = ref<any>(null)
const adminTreeData = ref<any[]>([])
const collaboratorCheckedKeys = ref<string[]>([])
const selectedDeptCount = computed(() => parseDeptIds(form.value.deptIds).length)

function parseDeptIds(value: string) {
  return String(value || '').split(',').map(id => Number(id.trim())).filter(Boolean)
}
function onDeptCheck() {
  form.value.deptIds = deptTreeRef.value?.getCheckedKeys()?.join(',') || ''
}
function onAdminTreeCheck() {
  form.value.collaborators = adminTreeRef.value?.getCheckedKeys(true)
    ?.filter((key: string) => key.startsWith('admin-'))
    ?.map((key: string) => key.replace('admin-', ''))
    ?.join(',') || ''
}
async function loadDeptTree() {
  if (form.value.visibility !== 2) return
  try {
    const response: any = await adminApi.deptTree()
    deptTreeData.value = response.data || []
    deptCheckedKeys.value = parseDeptIds(form.value.deptIds)
    await nextTick()
    deptTreeRef.value?.setCheckedKeys(deptCheckedKeys.value)
  } catch {
    deptTreeData.value = []
  }
}
async function loadAdminTree() {
  try {
    const [departmentResponse, managerResponse] = await Promise.all([
      adminApi.deptTree(),
      adminApi.mgrList({ page: 1, pageSize: 9999 }),
    ])
    const departments: any[] = departmentResponse.data || []
    const managers: any[] = managerResponse.data?.list || []
    const creator = managers.find(manager => manager.id === form.value.createBy)
    creatorName.value = creator ? creator.name : (form.value.createBy ? '未知用户' : '未指定')
    adminTreeData.value = buildAdminTree(departments, managers)
    collaboratorCheckedKeys.value = String(form.value.collaborators || '').split(',').filter(Boolean).map(id => `admin-${id}`)
    await nextTick()
    adminTreeRef.value?.setCheckedKeys(collaboratorCheckedKeys.value)
  } catch {
    creatorName.value = '加载失败'
    adminTreeData.value = []
  }
}
function buildAdminTree(departments: any[], managers: any[]): any[] {
  return departments.map(department => {
    const departmentManagers = managers.filter(manager => (manager.deptIds || []).includes(department.id))
    const managerNodes = departmentManagers.map(manager => ({ id: `admin-${manager.id}`, label: manager.name, type: 'admin' }))
    const children = department.children?.length ? buildAdminTree(department.children, managers) : []
    if (managerNodes.length) children.push(...managerNodes)
    return { id: `dept-${department.id}`, label: department.name, type: 'dept', children }
  }).filter(department => department.children?.length)
}

watch(() => form.value.visibility, value => {
  if (value === 2) void loadDeptTree()
})
watch(() => form.value.deptIds, value => {
  deptCheckedKeys.value = parseDeptIds(value)
  void nextTick(() => deptTreeRef.value?.setCheckedKeys(deptCheckedKeys.value))
})
watch(() => form.value.collaborators, value => {
  collaboratorCheckedKeys.value = String(value || '').split(',').filter(Boolean).map(id => `admin-${id}`)
  void nextTick(() => adminTreeRef.value?.setCheckedKeys(collaboratorCheckedKeys.value))
})
watch(() => form.value.createBy, () => {
  if (props.active && adminTreeData.value.length) void loadAdminTree()
})
watch(() => form.value.id, () => {
  if (!props.active) return
  void loadAdminTree()
  if (form.value.visibility === 2) void loadDeptTree()
})
watch(() => props.active, active => {
  if (!active) return
  void loadAdminTree()
  if (form.value.visibility === 2) void loadDeptTree()
}, { immediate: true })
</script>

<style scoped>
.setting-wrapper { height:100%; overflow-y:auto; background:var(--designer-bg); }
.setting-header { position:sticky; top:0; z-index:5; display:flex; align-items:center; justify-content:space-between; gap:16px; padding:18px 24px; border-bottom:1px solid var(--designer-border); background:rgba(248,250,252,.94); backdrop-filter:saturate(120%) blur(10px); }
.setting-page-title { color:#1f2937; font-size:18px; font-weight:700; line-height:1.3; }
.setting-page-desc { margin-top:3px; color:#667085; font-size:12px; line-height:1.5; }
.setting-header-actions { display:flex; flex-shrink:0; align-items:center; justify-content:flex-end; gap:10px; }
.setting-scroll { display:grid; grid-template-columns:repeat(auto-fill,minmax(360px,1fr)); align-items:start; gap:18px; padding:24px; }
.setting-group { overflow:hidden; padding:18px 20px; border:1px solid var(--designer-border); border-radius:10px; background:#fff; box-shadow:0 8px 18px rgba(15,23,42,.04); }
.group-title { margin-bottom:14px; color:#344054; font-size:14px; font-weight:650; letter-spacing:0; }
.settings-grid { display:flex; flex-direction:column; }
.settings-grid :deep(.el-form-item), .settings-full { width:100%; min-width:0; box-sizing:border-box; }
.settings-grid :deep(.el-input), .settings-grid :deep(.el-select), .settings-grid :deep(.el-input-number), .settings-grid :deep(.el-date-editor), .settings-full :deep(.el-select) { width:100%; }
.settings-grid :deep(.el-form-item__content) { min-width:0; overflow:hidden; }
.visibility-options { display:flex; flex-direction:column; align-items:flex-start; gap:8px; }
.setting-control-stack, .share-settings { display:flex; width:100%; min-width:0; flex-direction:column; gap:4px; }
.setting-control-stack :deep(.el-input-number) { width:100%; }
.setting-help { width:100%; margin-top:4px; color:#98a2b3; font-size:12px; line-height:1.5; }
.setting-help.danger { color:#f04438; }
.setting-tree-box { width:100%; max-height:300px; overflow:auto; box-sizing:border-box; padding:8px; border:1px solid #dcdfe6; border-radius:7px; transition:border-color .15s,box-shadow .15s; }
.setting-tree-box.invalid { border-color:#fda29b; box-shadow:0 0 0 2px rgba(240,68,56,.08); }
.admin-tree-box { max-height:400px; }
.completion-radio-group { display:flex; width:100%; min-width:0; }
.completion-radio-group :deep(.el-radio-button) { min-width:0; flex:1; }
.completion-radio-group :deep(.el-radio-button__inner) { width:100%; padding-right:8px; padding-left:8px; }
.share-settings { gap:8px; }
.qr-settings, .qr-actions { display:flex; align-items:center; gap:12px; }
.qr-actions { gap:6px; }
.qr-image { width:100px; height:100px; border:1px solid #eee; border-radius:4px; }
@media (max-width:1240px) {
  .setting-header { align-items:flex-start; flex-direction:column; }
  .setting-header-actions { width:100%; flex-wrap:wrap; justify-content:flex-start; }
  .setting-scroll { grid-template-columns:1fr; }
}
</style>
