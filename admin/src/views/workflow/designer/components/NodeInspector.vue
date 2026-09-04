<template>
  <aside class="node-inspector" :class="{ readonly }">
    <el-empty v-if="!selectedNode" :image-size="72" description="选择节点后在此配置" />

    <div v-else class="inspector-content">
      <section class="inspector-section">
        <label class="field-label">节点名称</label>
        <el-input v-model="selectedNode.name" maxlength="40" :disabled="readonly" @input="$emit('change')" />
      </section>

      <template v-if="['approval', 'handle', 'cc', 'notify'].includes(selectedNode.type)">
        <section class="inspector-section">
          <h4>{{ selectedNode.type === 'approval' ? '审批规则' : selectedNode.type === 'handle' ? '办理规则' : selectedNode.type === 'cc' ? '抄送规则' : '接收人规则' }}</h4>
          <template v-if="selectedNode.type === 'approval'">
            <label class="field-label">审批方式</label>
            <el-select :model-value="selectedNode.approvalMode" :disabled="readonly" style="width: 100%" @change="updateApprovalMode">
              <el-option label="单人审批" value="single" />
              <el-option label="依次审批" value="sequential" />
              <el-option label="并行审批" value="parallel" />
              <el-option label="会签审批" value="countersign" />
            </el-select>
            <div class="setting-switch-row spacing">
              <label class="field-label">逐级向上审批</label>
              <el-switch :model-value="departmentApprovalChainEnabled" :disabled="readonly" @change="updateDepartmentApprovalChainEnabled" />
            </div>
            <template v-if="departmentApprovalChainEnabled">
              <label class="field-label spacing">终止范围</label>
              <el-radio-group :model-value="departmentApprovalChainStopMode" :disabled="readonly" class="scope-group" @change="updateDepartmentApprovalChainStopMode">
                <el-radio-button label="root">组织根部门</el-radio-button>
                <el-radio-button label="department">指定部门</el-radio-button>
              </el-radio-group>
              <template v-if="departmentApprovalChainStopMode === 'department'">
                <label class="field-label spacing">终止部门</label>
                <el-tree-select
                  :model-value="selectedNode.departmentApprovalChain?.stopDepartmentId || undefined"
                  :data="departments"
                  :props="{ label: 'name', children: 'children' }"
                  node-key="id"
                  check-strictly
                  clearable
                  :render-after-expand="false"
                  :disabled="readonly"
                  placeholder="请选择终止部门"
                  style="width: 100%"
                  @change="updateDepartmentApprovalChainStopDepartment"
                />
              </template>
              <label class="field-label spacing">部门未配置该身份</label>
              <el-radio-group
                :model-value="selectedNode.departmentApprovalChain?.missingAssigneePolicy || 'skip'"
                :disabled="readonly"
                class="scope-group"
                @change="updateDepartmentApprovalChainMissingPolicy"
              >
                <el-radio-button label="skip">跳过该部门</el-radio-button>
                <el-radio-button label="error">阻止发起</el-radio-button>
              </el-radio-group>
              <div class="setting-switch-row spacing">
                <label class="field-label">跳过发起人本人</label>
                <el-switch
                  :model-value="selectedNode.departmentApprovalChain?.skipStarter === true"
                  :disabled="readonly"
                  @change="updateDepartmentApprovalChainSkipStarter"
                />
              </div>
            </template>
          </template>

          <label class="field-label spacing">{{ ['cc', 'notify'].includes(selectedNode.type) ? '接收人来源' : '处理人来源' }}</label>
          <el-select :model-value="selectedNode.assignee?.type" :disabled="readonly" style="width: 100%" @change="updateAssigneeType">
            <el-option label="发起人" value="initiator" />
            <el-option label="指定用户" value="user" />
            <el-option label="系统权限角色" value="role" />
            <el-option label="组织审批身份" value="org_identity" />
            <el-option label="部门负责人" value="department_leader" />
            <el-option label="直属上级" value="manager" />
            <el-option label="流程变量" value="variable" />
          </el-select>

          <template v-if="selectedNode.assignee?.type === 'user'">
            <label class="field-label spacing">处理人</label>
            <el-popover trigger="click" placement="bottom-start" :width="520" :disabled="readonly" popper-class="assignee-user-popover" popper-style="margin-top:4px">
              <template #reference>
                <el-input
                  readonly
                  class="assignee-user-tree"
                  :model-value="assigneeUserDisplayText"
                  :placeholder="canSelectMultipleAssignees ? '请选择一个或多个用户' : '请选择一个用户'"
                  suffix-icon="ArrowDown"
                  style="width: 100%"
                />
              </template>
              <div class="assignee-picker-search">
                <el-input
                  v-model="assigneeUserKeyword"
                  placeholder="搜索用户/手机号/部门"
                  clearable
                  size="small"
                  @input="filterAssigneeUserTree"
                />
              </div>
              <div class="assignee-user-tree-wrap">
                <el-tree
                  ref="assigneeUserTreeRef"
                  :data="assigneeUserTreeData"
                  :props="{ label: 'label', children: 'children', disabled: 'disabled' }"
                  :filter-node-method="filterAssigneeUserNode"
                  show-checkbox
                  check-on-click-node
                  node-key="key"
                  :default-checked-keys="assigneeUserCheckedNodeKeys"
                  @check="onAssigneeUserCheck"
                  class="assignee-user-tree"
                >
                  <template #default="{ node, data }">
                    <span :class="['assignee-user-node', data.type === 'user' ? 'is-user' : 'is-dept']">
                      <span class="assignee-user-node__label">{{ node.label }}</span>
                      <span v-if="data.type === 'user' && data.mobile" class="assignee-user-node__meta">{{ data.mobile }}</span>
                    </span>
                  </template>
                </el-tree>
              </div>
            </el-popover>
            <div v-if="selectedAssigneeUsers.length" class="selected-assignee-list">
              <div v-for="(user, index) in selectedAssigneeUsers" :key="user.id" class="selected-assignee-row">
                <el-tag :closable="!readonly" @close="removeAssigneeUser(user.id)">
                  {{ selectedNode.approvalMode === 'sequential' ? `${index + 1}. ` : '' }}{{ user.name }}
                </el-tag>
                <span v-if="user.mobile" class="selected-assignee-row__meta">{{ user.mobile }}</span>
                <template v-if="selectedNode.type === 'approval' && selectedNode.approvalMode === 'sequential' && selectedAssigneeUsers.length > 1">
                  <el-button circle size="small" icon="ArrowUp" :disabled="readonly || index === 0" @click="moveSelectedAssigneeUser(index, -1)" />
                  <el-button circle size="small" icon="ArrowDown" :disabled="readonly || index === selectedAssigneeUsers.length - 1" @click="moveSelectedAssigneeUser(index, 1)" />
                </template>
              </div>
            </div>
          </template>

          <template v-else-if="selectedNode.assignee?.type === 'org_identity'">
            <label class="field-label spacing">组织范围</label>
            <el-radio-group :model-value="orgIdentityScope" :disabled="readonly" class="scope-group" @change="updateOrgIdentityScope">
              <el-radio-button label="starter_department">发起人部门</el-radio-button>
              <el-radio-button label="department">指定部门</el-radio-button>
            </el-radio-group>

            <template v-if="orgIdentityScope === 'department'">
              <label class="field-label spacing">指定部门</label>
              <el-tree-select
                :model-value="orgIdentityDepartmentID || undefined"
                :data="departments"
                :props="{ label: 'name', children: 'children' }"
                node-key="id"
                check-strictly
                clearable
                :render-after-expand="false"
                :disabled="readonly"
                placeholder="请选择部门"
                style="width: 100%"
                @change="updateOrgIdentityDepartment"
              />
            </template>

            <label class="field-label spacing">组织审批身份</label>
            <el-select :model-value="orgIdentityCode" :disabled="readonly" style="width: 100%" @change="updateOrgIdentityCode">
              <el-option v-for="identity in orgIdentityOptions" :key="identity.code" :label="identity.name" :value="identity.code" />
            </el-select>
            <p v-if="selectedNode.type === 'approval'" class="assignee-helper">审批方式应用于同一部门内解析出的处理人；逐级审批固定按部门层级顺序流转。</p>
          </template>

          <template v-else-if="selectedNode.assignee?.type !== 'initiator'">
            <label class="field-label spacing">处理人标识</label>
            <el-input
              :model-value="selectedNode.assignee?.value"
              :disabled="readonly"
              :placeholder="assigneePlaceholder(selectedNode.assignee?.type)"
              @input="updateAssigneeValue"
            />
          </template>

          <template v-if="selectedNode.approvalMode === 'countersign'">
            <label class="field-label spacing">通过比例</label>
            <el-input-number
              :model-value="selectedNode.completionRate ?? 100"
              :min="1"
              :max="100"
              :step="5"
              :disabled="readonly"
              controls-position="right"
              style="width: 100%"
              @change="updateCompletionRate"
            />
          </template>
        </section>
      </template>

      <template v-if="['approval', 'handle', 'cc', 'notify'].includes(selectedNode.type)">
        <section class="inspector-section notification-section">
          <div class="section-title-row">
            <div class="section-title-with-help">
              <h4>{{ ['approval', 'handle'].includes(selectedNode.type) ? '任务到达通知' : selectedNode.type === 'cc' ? '抄送通知' : '通知配置' }}</h4>
              <el-tooltip content="通知消息配置说明" placement="top">
                <el-button
                  class="notification-help-button"
                  link
                  type="primary"
                  :icon="QuestionFilled"
                  aria-label="打开通知消息配置说明"
                  @click="notificationHelpVisible = true"
                />
              </el-tooltip>
            </div>
            <el-switch
              v-if="selectedNode.type !== 'notify'"
              :model-value="notificationEnabled"
              :disabled="readonly"
              @change="updateNotificationEnabled"
            />
            <el-tag v-else size="small" type="success" effect="plain">必须发送</el-tag>
          </div>
          <template v-if="notificationEnabled">
            <label class="field-label">通知渠道</label>
            <el-checkbox-group :model-value="notificationChannels" :disabled="readonly" class="notification-channels" @change="updateNotificationChannels">
              <el-checkbox value="in_app">站内通知</el-checkbox>
              <el-checkbox value="dingtalk_oa">钉钉 OA</el-checkbox>
            </el-checkbox-group>
            <label class="field-label spacing">通知标题</label>
            <el-input
              :model-value="selectedNode.notification?.title || ''"
              maxlength="256"
              :disabled="readonly"
              placeholder="{{workflowName}}"
              @input="updateNotificationTitle"
            />
            <label class="field-label spacing">通知正文</label>
            <el-input
              :model-value="selectedNode.notification?.content || ''"
              type="textarea"
              :rows="5"
              maxlength="2000"
              show-word-limit
              :disabled="readonly"
              @input="updateNotificationContent"
            />
          </template>
        </section>
      </template>

      <WorkflowResultNotificationEditor
        v-if="selectedNode.type === 'approval'"
        :model-value="selectedNode.resultNotification"
        :readonly="readonly"
        @update:model-value="updateResultNotification"
      />

      <template v-if="selectedNode.type === 'automation'">
        <section class="inspector-section">
          <h4>自动动作</h4>
          <label class="field-label">写入流程变量</label>
          <el-input v-model="automationVariablesText" type="textarea" :rows="8" :disabled="readonly" placeholder="JSON 对象" />
          <el-button class="section-action" type="primary" plain :disabled="readonly" @click="applyAutomationVariables">应用变量</el-button>
        </section>
      </template>

      <template v-if="selectedNode.type === 'timer'">
        <section class="inspector-section">
          <h4>定时等待</h4>
          <label class="field-label">等待秒数</label>
          <el-input-number
            :model-value="selectedNode.timer?.delaySeconds || 86400"
            :min="1"
            :max="31536000"
            :step="60"
            :disabled="readonly"
            controls-position="right"
            style="width: 100%"
            @change="updateTimerDelay"
          />
        </section>
      </template>

      <template v-if="selectedNode.gatewayMode === 'split'">
        <section class="inspector-section">
          <div class="section-title-row">
            <h4>分支配置</h4>
            <el-button size="small" type="primary" plain icon="Plus" :disabled="readonly" @click="$emit('add-branch', selectedNode.id)">新增分支</el-button>
          </div>
          <div v-for="(branch, index) in branchEdges" :key="branch.id" class="branch-editor">
            <div class="branch-editor__heading">
              <b>分支 {{ index + 1 }}</b>
              <el-button
                link
                type="danger"
                :disabled="readonly || branchEdges.length <= 2"
                @click="$emit('delete-branch', selectedNode.id, branch.id)"
              >删除</el-button>
            </div>
            <label class="field-label">分支名称</label>
            <el-input :model-value="branch.name" size="small" :disabled="readonly" @input="updateBranchName(branch, $event)" />
            <template v-if="selectedNode.type === 'exclusive'">
              <el-checkbox :model-value="branch.default === true" :disabled="readonly" class="default-check" @change="updateDefaultBranch(branch, $event)">默认分支</el-checkbox>
              <template v-if="!branch.default">
                <label class="field-label">条件字段</label>
                <el-select
                  :model-value="branch.condition?.field"
                  size="small"
                  class="condition-field-select"
                  :disabled="readonly"
                  filterable
                  allow-create
                  default-first-option
                  placeholder="选择表单字段"
                  @change="updateConditionField(branch, $event)"
                >
                  <el-option
                    v-for="field in conditionFieldOptions"
                    :key="field.key"
                    :label="`${field.label}（${field.key}）`"
                    :value="field.key"
                  />
                </el-select>
                <div class="condition-row">
                  <el-select :model-value="branch.condition?.operator || 'eq'" size="small" :disabled="readonly" @change="updateConditionOperator(branch, $event)">
                    <el-option label="等于" value="eq" />
                    <el-option label="不等于" value="ne" />
                    <el-option label="大于" value="gt" />
                    <el-option label="大于等于" value="gte" />
                    <el-option label="小于" value="lt" />
                    <el-option label="小于等于" value="lte" />
                  </el-select>
                  <el-input :model-value="String(branch.condition?.value ?? '')" size="small" :disabled="readonly" placeholder="条件值" @input="updateConditionValue(branch, $event)" />
                </div>
              </template>
            </template>
          </div>
        </section>
      </template>

      <section v-if="selectedNode.gatewayMode === 'join'" class="inspector-section join-note">
        <el-icon><Connection /></el-icon>
        <span>该节点由对应分支自动汇聚。删除此节点会同时删除整组分支。</span>
      </section>

      <div v-if="canDelete" class="inspector-footer">
        <el-button type="danger" plain icon="Delete" @click="$emit('delete', selectedNode.id)">删除节点</el-button>
      </div>
    </div>
  </aside>

  <el-dialog
    v-model="notificationHelpVisible"
    title="通知消息配置说明"
    width="min(680px, calc(100vw - 32px))"
    append-to-body
    destroy-on-close
  >
    <div class="notification-help-content">
      <section class="notification-help-section">
        <h3>如何配置</h3>
        <ol class="notification-help-steps">
          <li>选择站内通知、钉钉 OA，或同时选择两个渠道。</li>
          <li>在通知标题和通知正文中输入固定文字，并按需插入下方占位符。</li>
          <li>保存并发布流程后，系统会在任务到达或通知节点触发时替换占位符并发送消息。</li>
        </ol>
        <el-alert
          title="当前所选渠道共用同一套标题和正文；钉钉通知点击后进入企业配置的应用首页。"
          type="info"
          :closable="false"
          show-icon
        />
      </section>

      <section class="notification-help-section">
        <h3>可用占位符</h3>
        <el-table :data="notificationPlaceholderRows" border size="small" table-layout="fixed">
          <el-table-column label="占位符" width="180">
            <template #default="{ row }"><code class="placeholder-code" v-text="row.placeholder" /></template>
          </el-table-column>
          <el-table-column prop="description" label="替换内容" />
          <el-table-column prop="example" label="示例" width="150" />
        </el-table>
      </section>

      <section class="notification-help-section">
        <h3>配置示例</h3>
        <dl class="notification-example-list">
          <div><dt>标题模板</dt><dd><code v-text="notificationExample.titleTemplate" /></dd></div>
          <div><dt>正文模板</dt><dd><code v-text="notificationExample.contentTemplate" /></dd></div>
          <div><dt>发送标题</dt><dd>{{ notificationExample.renderedTitle }}</dd></div>
          <div><dt>发送正文</dt><dd>{{ notificationExample.renderedContent }}</dd></div>
        </dl>
      </section>

      <section class="notification-help-section notification-help-limits">
        <h3>格式限制</h3>
        <p>配置时标题最多 256 个字符，正文最多 2000 个字符。</p>
        <p>发送时标题最多 64 个字符，正文最多 1000 个字符，超出部分会被截断。</p>
        <p>为避免表单数据泄露，通知模板不支持直接读取表单字段或流程变量。</p>
      </section>
    </div>
    <template #footer>
      <el-button type="primary" @click="notificationHelpVisible = false">知道了</el-button>
    </template>
  </el-dialog>
</template>

<script lang="ts" setup>
import { computed, nextTick, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { QuestionFilled } from '@element-plus/icons-vue'
import { defaultNotificationConfig } from '../graph'
import WorkflowResultNotificationEditor from './WorkflowResultNotificationEditor.vue'
import type {
  ApprovalMode,
  AssigneeType,
  ConditionOperator,
  WorkflowAssigneeUser,
  WorkflowDraft,
  WorkflowEdge,
  WorkflowOrgApproverIdentity,
  WorkflowNotificationChannel,
  WorkflowNotificationConfig,
} from '../../types'

const props = defineProps<{
  draft: WorkflowDraft
  selectedNodeId: string
  readonly?: boolean
  departments?: any[]
  users?: WorkflowAssigneeUser[]
  orgIdentities?: WorkflowOrgApproverIdentity[]
}>()
const emit = defineEmits<{
  change: []
  delete: [nodeId: string]
  'delete-branch': [splitId: string, edgeId: string]
  'add-branch': [splitId: string]
}>()

const assigneeUserTreeRef = ref<any>(null)
const assigneeUserKeyword = ref('')
const automationVariablesText = ref('{}')
const notificationHelpVisible = ref(false)
const notificationPlaceholderRows = [
  { placeholder: '{{workflowName}}', description: '当前流程定义名称', example: '请假审批' },
  { placeholder: '{{nodeName}}', description: '当前审批、办理、抄送或通知节点名称', example: '部门负责人审批' },
  { placeholder: '{{starterName}}', description: '流程业务发起人的显示名称', example: '张三' },
  { placeholder: '{{instanceId}}', description: '当前流程实例 ID', example: 'instance_123' },
  { placeholder: '{{taskId}}', description: '当前审批或办理任务 ID；非任务通知为空', example: 'task_456' },
]
const notificationExample = {
  titleTemplate: '{{workflowName}}',
  contentTemplate: '{{starterName}} 发起的流程已到达 {{nodeName}}，请及时处理。',
  renderedTitle: '请假审批',
  renderedContent: '张三发起的流程已到达部门负责人审批，请及时处理。',
}

const selectedNode = computed(() => props.draft.nodes.find(item => item.id === props.selectedNodeId))
const branchEdges = computed(() => selectedNode.value?.gatewayMode === 'split'
  ? props.draft.edges.filter(item => item.source === selectedNode.value?.id)
  : [])
const conditionFieldOptions = computed(() => props.draft.form || [])
const canDelete = computed(() => !props.readonly && selectedNode.value && !['start', 'end'].includes(selectedNode.value.type))
const departments = computed(() => props.departments || [])
const users = computed(() => props.users || [])
const orgIdentityOptions = computed(() => {
  if (props.orgIdentities?.length) return props.orgIdentities
  return [
    { code: 'department_leader', name: '部门负责人' },
    { code: 'supervisor', name: '主管' },
  ]
})
const canSelectMultipleAssignees = computed(() => selectedNode.value?.type === 'cc' || selectedNode.value?.type === 'notify'
  || (selectedNode.value?.type === 'approval' && selectedNode.value.approvalMode !== 'single'))
const assigneeUserTreeData = computed(() => buildAssigneeUserTree(departments.value, users.value))
const selectedAssigneeUserIDs = computed(() => parseNumberIDs(selectedNode.value?.assignee?.type === 'user' ? selectedNode.value.assignee.value : ''))
const selectedAssigneeUsers = computed(() => {
  const byID = new Map(users.value.map(user => [Number(user.id), user]))
  return selectedAssigneeUserIDs.value.map((id) => {
    const user = byID.get(id)
    return user || { id, name: String(id) }
  })
})
const assigneeUserDisplayText = computed(() => selectedAssigneeUsers.value.map(user => user.name || String(user.id)).join('、'))
const assigneeUserCheckedNodeKeys = computed(() => checkedAssigneeUserNodeKeys(selectedAssigneeUserIDs.value, assigneeUserTreeData.value))
const orgIdentitySelection = computed(() => parseOrgIdentityValue(selectedNode.value?.assignee?.value || ''))
const orgIdentityScope = computed(() => orgIdentitySelection.value.scope)
const orgIdentityDepartmentID = computed(() => orgIdentitySelection.value.departmentID)
const orgIdentityCode = computed(() => orgIdentitySelection.value.identityCode || defaultOrgIdentityCode())
const departmentApprovalChainEnabled = computed(() => selectedNode.value?.departmentApprovalChain?.enabled === true)
const departmentApprovalChainStopMode = computed(() => selectedNode.value?.departmentApprovalChain?.stopMode || 'root')
const notificationEnabled = computed(() => selectedNode.value?.type === 'notify' || selectedNode.value?.notification?.enabled === true)
const notificationChannels = computed(() => selectedNode.value?.notification?.channels || [])

watch(() => props.selectedNodeId, () => {
  syncAutomationVariablesText()
  nextTick(syncAssigneeUserTreeKeys)
}, { immediate: true })
watch(() => selectedNode.value?.assignee?.value, () => nextTick(syncAssigneeUserTreeKeys))
watch(() => users.value.length, () => nextTick(syncAssigneeUserTreeKeys))

function ensureAssignee() {
  if (!selectedNode.value) return undefined
  selectedNode.value.assignee ||= { type: 'manager', value: 'direct_manager' }
  return selectedNode.value.assignee
}

function ensureNotification() {
  if (!selectedNode.value || !['approval', 'handle', 'cc', 'notify'].includes(selectedNode.value.type)) return undefined
  if (!selectedNode.value.notification) {
    selectedNode.value.notification = defaultNotificationConfig(selectedNode.value.type as 'approval' | 'handle' | 'cc' | 'notify')
  }
  if (selectedNode.value.type === 'notify') selectedNode.value.notification.enabled = true
  return selectedNode.value.notification
}

function updateNotificationEnabled(value: string | number | boolean) {
  if (!selectedNode.value || selectedNode.value.type === 'notify') return
  const notification = ensureNotification()
  if (!notification) return
  notification.enabled = Boolean(value)
  emit('change')
}

function updateNotificationChannels(value: unknown) {
  const notification = ensureNotification()
  if (!notification) return
  const allowed = new Set<WorkflowNotificationChannel>(['in_app', 'dingtalk_oa'])
  notification.channels = Array.from(new Set(Array.isArray(value) ? value : []))
    .filter((item): item is WorkflowNotificationChannel => allowed.has(item as WorkflowNotificationChannel))
  emit('change')
}

function updateNotificationTitle(value: string) {
  const notification = ensureNotification()
  if (!notification) return
  notification.title = String(value || '')
  emit('change')
}

function updateNotificationContent(value: string) {
  const notification = ensureNotification()
  if (!notification) return
  notification.content = String(value || '')
  emit('change')
}

function updateResultNotification(value: WorkflowNotificationConfig) {
  if (!selectedNode.value || selectedNode.value.type !== 'approval') return
  selectedNode.value.resultNotification = value
  emit('change')
}

function updateApprovalMode(value: ApprovalMode) {
  if (!selectedNode.value) return
  selectedNode.value.approvalMode = value
  if (value !== 'countersign') delete selectedNode.value.completionRate
  else selectedNode.value.completionRate ||= 100
  if (value === 'single' && selectedNode.value.assignee?.type === 'user') {
    const firstID = selectedAssigneeUserIDs.value[0]
    selectedNode.value.assignee.value = firstID ? String(firstID) : ''
    nextTick(syncAssigneeUserTreeKeys)
  }
  emit('change')
}

function updateAssigneeType(value: AssigneeType) {
  const assignee = ensureAssignee()
  if (!assignee) return
  assignee.type = value
  assignee.value = defaultAssigneeValue(value)
  if (value !== 'org_identity' && selectedNode.value) delete selectedNode.value.departmentApprovalChain
  assigneeUserKeyword.value = ''
  nextTick(syncAssigneeUserTreeKeys)
  emit('change')
}

function defaultAssigneeValue(type: AssigneeType) {
  if (type === 'manager') return 'direct_manager'
  if (type === 'department_leader') return 'current_department'
  if (type === 'org_identity') return `starter_department:${defaultOrgIdentityCode()}`
  return ''
}

function updateAssigneeValue(value: string) {
  const assignee = ensureAssignee()
  if (!assignee) return
  assignee.value = value
  emit('change')
}

function updateCompletionRate(value: number | undefined) {
  if (!selectedNode.value) return
  selectedNode.value.completionRate = Number(value || 100)
  emit('change')
}

function syncAutomationVariablesText() {
  automationVariablesText.value = JSON.stringify(selectedNode.value?.automation?.variables || {}, null, 2)
}

function applyAutomationVariables() {
  if (!selectedNode.value || selectedNode.value.type !== 'automation') return
  try {
    const variables = JSON.parse(automationVariablesText.value || '{}')
    if (!variables || Array.isArray(variables) || typeof variables !== 'object' || Object.keys(variables).length === 0) {
      throw new Error('请填写非空 JSON 对象')
    }
    selectedNode.value.automation = { type: 'set_variables', variables }
    automationVariablesText.value = JSON.stringify(variables, null, 2)
    emit('change')
  } catch (error) {
    ElMessage.warning(error instanceof Error ? error.message : '变量 JSON 格式无效')
  }
}

function updateTimerDelay(value: number | undefined) {
  if (!selectedNode.value || selectedNode.value.type !== 'timer') return
  selectedNode.value.timer = { delaySeconds: Math.min(31536000, Math.max(1, Number(value || 1))) }
  emit('change')
}

function assigneePlaceholder(type?: AssigneeType) {
  return ({
    initiator: '流程业务发起人',
    user: '请选择用户',
    role: '输入系统权限角色 ID，多个用逗号分隔',
    department_leader: '例如：current_department',
    manager: '例如：direct_manager',
    variable: '输入流程变量名',
    org_identity: 'starter_department:department_leader',
  } as Record<AssigneeType, string>)[type || 'manager']
}

function filterAssigneeUserTree() {
  assigneeUserTreeRef.value?.filter(normalizeFilterText(assigneeUserKeyword.value))
}

function filterAssigneeUserNode(keyword: string, data: any) {
  const value = normalizeFilterText(keyword)
  if (!value) return true
  return [
    data?.label,
    data?.mobile,
    data?.pathText
  ].some((item) => normalizeFilterText(item).includes(value))
}

function onAssigneeUserCheck() {
  nextTick(() => {
    const checkedKeys = assigneeUserTreeRef.value?.getCheckedKeys?.() || []
    const checkedIDs = assigneeUserIDsFromCheckedKeys(checkedKeys)
    const currentIDs = selectedAssigneeUserIDs.value
    let nextIDs = mergeSelectedUserOrder(currentIDs, checkedIDs)
    if (!canSelectMultipleAssignees.value) {
      const currentSet = new Set(currentIDs)
      const newID = checkedIDs.find(id => !currentSet.has(id)) || checkedIDs[checkedIDs.length - 1] || 0
      nextIDs = newID ? [newID] : []
    }
    updateAssigneeValue(nextIDs.join(','))
    nextTick(syncAssigneeUserTreeKeys)
  })
}

function removeAssigneeUser(userID: number) {
  if (props.readonly) return
  updateAssigneeValue(selectedAssigneeUserIDs.value.filter(id => id !== Number(userID)).join(','))
  nextTick(syncAssigneeUserTreeKeys)
}

function moveSelectedAssigneeUser(index: number, offset: number) {
  if (props.readonly) return
  const target = index + offset
  const ids = [...selectedAssigneeUserIDs.value]
  if (target < 0 || target >= ids.length) return
  const current = ids[index]
  ids[index] = ids[target]
  ids[target] = current
  updateAssigneeValue(ids.join(','))
}

function syncAssigneeUserTreeKeys() {
  assigneeUserTreeRef.value?.setCheckedKeys(assigneeUserCheckedNodeKeys.value)
}

function mergeSelectedUserOrder(currentIDs: number[], checkedIDs: number[]) {
  const checkedSet = new Set(checkedIDs)
  const result = currentIDs.filter(id => checkedSet.has(id))
  checkedIDs.forEach((id) => {
    if (!result.includes(id)) result.push(id)
  })
  return result
}

function updateOrgIdentityScope(value: string | number | boolean) {
  const scope = String(value) === 'department' ? 'department' : 'starter_department'
  if (scope === 'department' && selectedNode.value) delete selectedNode.value.departmentApprovalChain
  updateOrgIdentityValue(scope, orgIdentityDepartmentID.value, orgIdentityCode.value)
}

function updateOrgIdentityDepartment(value: string | number | undefined) {
  updateOrgIdentityValue(orgIdentityScope.value, Number(value) || 0, orgIdentityCode.value)
}

function updateOrgIdentityCode(value: string) {
  updateOrgIdentityValue(orgIdentityScope.value, orgIdentityDepartmentID.value, value)
}

function updateOrgIdentityValue(scope: string, departmentID: number, identityCode: string) {
  const assignee = ensureAssignee()
  if (!assignee) return
  const code = identityCode || defaultOrgIdentityCode()
  assignee.type = 'org_identity'
  assignee.value = scope === 'department'
    ? `department:${Number(departmentID) || 0}:${code}`
    : `starter_department:${code}`
  emit('change')
}

function ensureDepartmentApprovalChain() {
  if (!selectedNode.value || selectedNode.value.type !== 'approval') return undefined
  selectedNode.value.departmentApprovalChain ||= {
    enabled: true,
    stopMode: 'root',
    missingAssigneePolicy: 'skip',
  }
  return selectedNode.value.departmentApprovalChain
}

function updateDepartmentApprovalChainEnabled(value: string | number | boolean) {
  if (!selectedNode.value) return
  if (!Boolean(value)) {
    delete selectedNode.value.departmentApprovalChain
    emit('change')
    return
  }
  const config = ensureDepartmentApprovalChain()
  if (!config) return
  config.enabled = true
  updateOrgIdentityValue('starter_department', 0, orgIdentityCode.value)
}

function updateDepartmentApprovalChainStopMode(value: string | number | boolean) {
  const config = ensureDepartmentApprovalChain()
  if (!config) return
  config.stopMode = String(value) === 'department' ? 'department' : 'root'
  if (config.stopMode === 'root') delete config.stopDepartmentId
  emit('change')
}

function updateDepartmentApprovalChainStopDepartment(value: string | number | undefined) {
  const config = ensureDepartmentApprovalChain()
  if (!config) return
  config.stopDepartmentId = Number(value) || undefined
  emit('change')
}

function updateDepartmentApprovalChainMissingPolicy(value: string | number | boolean) {
  const config = ensureDepartmentApprovalChain()
  if (!config) return
  config.missingAssigneePolicy = String(value) === 'error' ? 'error' : 'skip'
  emit('change')
}

function updateDepartmentApprovalChainSkipStarter(value: string | number | boolean) {
  const config = ensureDepartmentApprovalChain()
  if (!config) return
  config.skipStarter = Boolean(value)
  emit('change')
}

function parseOrgIdentityValue(value: string) {
  const parts = String(value || '').split(':').map(item => item.trim()).filter(Boolean)
  if (parts[0] === 'department') {
    return { scope: 'department', departmentID: Number(parts[1]) || 0, identityCode: parts[2] || defaultOrgIdentityCode() }
  }
  if (parts[0] === 'starter_department') {
    return { scope: 'starter_department', departmentID: 0, identityCode: parts[1] || defaultOrgIdentityCode() }
  }
  return { scope: 'starter_department', departmentID: 0, identityCode: parts[0] || defaultOrgIdentityCode() }
}

function defaultOrgIdentityCode() {
  return orgIdentityOptions.value[0]?.code || 'department_leader'
}

function normalizeFilterText(value: any) {
  return String(value || '').trim().toLowerCase()
}

function parseNumberIDs(value: string) {
  const seen = new Set<number>()
  const result: number[] = []
  String(value || '').split(',').forEach((item) => {
    const id = Number(item)
    if (!id || seen.has(id)) return
    seen.add(id)
    result.push(id)
  })
  return result
}

function normalizeNumberIDs(ids: any[]) {
  return Array.from(new Set((ids || []).map((id: any) => Number(id)).filter(Boolean)))
}

function workflowDeptPathMap(depts: any[]) {
  const result: Record<number, string[]> = {}
  function walk(nodes: any[], parentPath: string[] = []) {
    for (const node of nodes || []) {
      const id = Number(node.id)
      if (!id) continue
      const name = String(node.name || id)
      const path = [...parentPath, name]
      result[id] = path
      walk(node.children || [], path)
    }
  }
  walk(depts)
  return result
}

function workflowDeptPathText(deptIDs: number[]) {
  const pathMap = workflowDeptPathMap(departments.value)
  const texts = normalizeNumberIDs(deptIDs || []).map((deptID) => (pathMap[deptID] || [String(deptID)]).join('/'))
  return texts.join('、')
}

function assigneeUserNodeKey(userID: number, scopeKey: string) {
  return `user:${scopeKey}:${userID}`
}

function assigneeUserIDFromNodeKey(key: string) {
  if (!key.startsWith('user:')) return 0
  const parts = key.split(':')
  return Number(parts[parts.length - 1]) || 0
}

function assigneeUserNode(user: WorkflowAssigneeUser, scopeKey: string) {
  const userID = Number(user.id)
  return {
    key: assigneeUserNodeKey(userID, scopeKey),
    type: 'user',
    id: userID,
    label: user.name || String(userID),
    mobile: user.mobile || '',
    pathText: workflowDeptPathText(user.deptIds || []),
    disabled: false,
  }
}

function buildAssigneeUserTree(depts: any[], userList: WorkflowAssigneeUser[]) {
  const usersByDept: Record<number, WorkflowAssigneeUser[]> = {}
  const usersWithoutDept: WorkflowAssigneeUser[] = []
  for (const user of userList || []) {
    const deptIDs = normalizeNumberIDs(user.deptIds || [])
    if (deptIDs.length === 0) {
      usersWithoutDept.push(user)
      continue
    }
    for (const deptID of deptIDs) {
      if (!usersByDept[deptID]) usersByDept[deptID] = []
      usersByDept[deptID].push(user)
    }
  }

  function buildDeptNode(dept: any): any | null {
    const deptID = Number(dept.id)
    if (!deptID) return null
    const childDeptNodes = (dept.children || []).map((item: any) => buildDeptNode(item)).filter(Boolean)
    const userNodes = (usersByDept[deptID] || []).map((user: WorkflowAssigneeUser) => assigneeUserNode(user, `dept-${deptID}`))
    const children = [...childDeptNodes, ...userNodes]
    if (children.length === 0) return null
    return {
      key: `dept-${deptID}`,
      type: 'dept',
      id: deptID,
      label: dept.name || String(deptID),
      disabled: true,
      children
    }
  }

  const tree = (depts || []).map((dept: any) => buildDeptNode(dept)).filter(Boolean)
  if (usersWithoutDept.length > 0) {
    tree.push({
      key: 'dept-unassigned',
      type: 'dept',
      label: '未分配部门',
      disabled: true,
      children: usersWithoutDept.map((user: WorkflowAssigneeUser) => assigneeUserNode(user, 'unassigned'))
    })
  }
  const visibleUserIDs = new Set<number>()
  function collectUserIDs(items: any[]) {
    for (const item of items || []) {
      if (item.type === 'user') visibleUserIDs.add(Number(item.id))
      collectUserIDs(item.children || [])
    }
  }
  collectUserIDs(tree)
  const unmatchedUsers = (userList || []).filter((user: WorkflowAssigneeUser) => !visibleUserIDs.has(Number(user.id)))
  if (unmatchedUsers.length > 0) {
    tree.push({
      key: 'dept-unmatched',
      type: 'dept',
      label: '未匹配部门',
      disabled: true,
      children: unmatchedUsers.map((user: WorkflowAssigneeUser) => assigneeUserNode(user, 'unmatched'))
    })
  }
  return tree
}

function checkedAssigneeUserNodeKeys(userIDs: number[], nodes: any[]) {
  const selected = new Set(normalizeNumberIDs(userIDs || []))
  const keys: string[] = []
  function walk(items: any[]) {
    for (const item of items || []) {
      if (item.type === 'user' && selected.has(Number(item.id))) keys.push(item.key)
      walk(item.children || [])
    }
  }
  walk(nodes)
  return keys
}

function assigneeUserIDsFromCheckedKeys(keys: string[]) {
  const ids = (keys || []).map(key => assigneeUserIDFromNodeKey(String(key))).filter(Boolean)
  return normalizeNumberIDs(ids)
}

function updateEdge(edge: WorkflowEdge, key: 'name', value: unknown) {
  edge[key] = String(value || '')
  emit('change')
}

function updateBranchName(edge: WorkflowEdge, value: string) {
  updateEdge(edge, 'name', value)
}

function ensureCondition(edge: WorkflowEdge) {
  edge.condition ||= { field: '', operator: 'eq', value: '' }
  return edge.condition
}

function updateCondition(edge: WorkflowEdge, key: 'field' | 'operator' | 'value', value: unknown) {
  const condition = ensureCondition(edge)
  if (key === 'operator') condition.operator = value as ConditionOperator
  else if (key === 'field') condition.field = String(value || '')
  else condition.value = String(value ?? '')
  emit('change')
}

function updateConditionField(edge: WorkflowEdge, value: string) {
  updateCondition(edge, 'field', value)
}

function updateConditionOperator(edge: WorkflowEdge, value: ConditionOperator) {
  updateCondition(edge, 'operator', value)
}

function updateConditionValue(edge: WorkflowEdge, value: string) {
  updateCondition(edge, 'value', value)
}

function updateDefaultBranch(edge: WorkflowEdge, value: string | number | boolean) {
  setDefaultBranch(edge, Boolean(value))
}

function setDefaultBranch(edge: WorkflowEdge, checked: boolean) {
  if (!selectedNode.value) return
  if (checked) {
    branchEdges.value.forEach(item => { item.default = item.id === edge.id })
    delete edge.condition
  } else {
    edge.default = false
    ensureCondition(edge)
  }
  emit('change')
}
</script>

<style scoped>
.node-inspector { width: 100%; min-height: 100%; background: #fff; }
.inspector-content { padding: 0 2px 24px; }
.inspector-section { padding: 18px 0; border-bottom: 1px solid #edf0f5; }
.inspector-section h4 { margin: 0 0 14px; color: #334155; font-size: 13px; }
.field-label { display: block; margin-bottom: 7px; color: #64748b; font-size: 12px; font-weight: 500; }
.field-label.spacing { margin-top: 14px; }
.section-action { width: 100%; margin-top: 10px; }
.scope-group { width: 100%; }
.scope-group :deep(.el-radio-button) { width: 50%; }
.scope-group :deep(.el-radio-button__inner) { width: 100%; }
.setting-switch-row { display: flex; min-height: 32px; align-items: center; justify-content: space-between; gap: 12px; }
.setting-switch-row.spacing { margin-top: 14px; }
.setting-switch-row .field-label { margin-bottom: 0; }
.assignee-helper { margin: 8px 0 0; color: #64748b; font-size: 12px; line-height: 1.6; }
.assignee-picker-search { margin-bottom: 10px; }
.assignee-user-tree-wrap { max-height: 320px; overflow-y: auto; }
.assignee-user-tree { width: 100%; }
.assignee-user-node { display: inline-flex; min-width: 0; align-items: center; gap: 8px; }
.assignee-user-node__label { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.assignee-user-node__meta { color: #94a3b8; font-size: 11px; }
.selected-assignee-list { display: flex; flex-direction: column; gap: 7px; margin-top: 10px; }
.selected-assignee-row { display: flex; min-height: 26px; align-items: center; gap: 6px; }
.selected-assignee-row__meta { flex: 1; min-width: 0; overflow: hidden; color: #94a3b8; font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }
.section-title-row { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
.section-title-row h4 { margin: 0; }
.section-title-with-help { display: inline-flex; min-width: 0; align-items: center; gap: 3px; }
.notification-help-button { width: 28px; height: 28px; padding: 0; font-size: 16px; }
.notification-channels { display: flex; flex-wrap: wrap; gap: 4px 18px; }
.notification-channels :deep(.el-checkbox) { margin-right: 0; }
.branch-editor { margin-bottom: 12px; padding: 12px; border: 1px solid #e5eaf0; border-radius: 7px; background: #fafbfd; }
.branch-editor:last-child { margin-bottom: 0; }
.branch-editor__heading { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; color: #334155; font-size: 12px; }
.default-check { margin: 9px 0; }
.condition-field-select { width: 100%; }
.condition-row { display: grid; grid-template-columns: 105px 1fr; gap: 7px; margin-top: 7px; }
.join-note { display: flex; align-items: flex-start; gap: 8px; color: #64748b; font-size: 12px; line-height: 1.6; }
.join-note .el-icon { margin-top: 3px; color: #0f766e; }
.inspector-footer { padding-top: 18px; }
.notification-help-content { color: #475569; }
.notification-help-section + .notification-help-section { margin-top: 22px; }
.notification-help-section h3 { margin: 0 0 10px; color: #1f2937; font-size: 15px; }
.notification-help-steps { margin: 0 0 14px; padding-left: 22px; line-height: 1.8; }
.notification-example-list { margin: 0; border-top: 1px solid #e5e7eb; }
.notification-example-list > div { display: grid; grid-template-columns: 92px minmax(0, 1fr); border-bottom: 1px solid #e5e7eb; }
.notification-example-list dt,
.notification-example-list dd { margin: 0; padding: 10px 12px; line-height: 1.6; }
.notification-example-list dt { background: #f8fafc; color: #64748b; font-weight: 600; }
.notification-example-list dd { min-width: 0; overflow-wrap: anywhere; color: #334155; }
.placeholder-code,
.notification-example-list code { color: #2563eb; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; }
.notification-help-limits { padding: 12px 14px; border-left: 3px solid #409eff; background: #f6faff; }
.notification-help-limits p { margin: 0; line-height: 1.8; }
@media (max-width: 560px) {
  .notification-example-list > div { grid-template-columns: 1fr; }
  .notification-example-list dt { padding-bottom: 4px; }
  .notification-example-list dd { padding-top: 4px; }
}
</style>
