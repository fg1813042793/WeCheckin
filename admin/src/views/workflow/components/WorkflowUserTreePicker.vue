<template>
  <div class="workflow-user-picker">
    <el-popover trigger="click" placement="bottom-start" :width="520" :disabled="disabled" popper-style="margin-top:4px">
      <template #reference>
        <div
          ref="selectionRef"
          :class="['workflow-user-picker__selection', { 'is-disabled': disabled }]"
          role="combobox"
          :aria-disabled="disabled"
          tabindex="0"
        >
          <span v-if="!selectedItems.length" class="workflow-user-picker__placeholder">
            {{ placeholder || (multiple ? '请选择一个或多个用户' : '请选择一个用户') }}
          </span>
          <span v-else class="workflow-user-picker__value">
            <span v-if="visibleSelectedItems.length" class="workflow-user-picker__names">
              {{ visibleSelectedItems.map(item => item.label).join('、') }}
            </span>
            <el-tooltip
              v-if="hiddenSelectedItems.length"
              placement="top"
              popper-class="workflow-user-picker__overflow-tooltip"
            >
              <template #content>
                <div class="workflow-user-picker__overflow-list">
                  <span v-for="item in hiddenSelectedItems" :key="item.key">
                    {{ item.label }}
                  </span>
                </div>
              </template>
              <span class="workflow-user-picker__overflow">+{{ hiddenSelectedItems.length }}</span>
            </el-tooltip>
          </span>
          <el-icon class="workflow-user-picker__arrow"><ArrowDown /></el-icon>
        </div>
      </template>
      <el-input
        v-model="keyword"
        placeholder="搜索用户/手机号/部门"
        clearable
        size="small"
        @input="filterTree"
      />
      <div class="workflow-user-picker__tree">
        <el-tree
          ref="treeRef"
          :data="treeData"
          :props="{ label: 'label', children: 'children', disabled: 'disabled' }"
          :filter-node-method="filterNode"
          show-checkbox
          check-on-click-node
          :check-strictly="selectDepartmentRules"
          node-key="key"
          :default-checked-keys="checkedNodeKeys"
          @check="onCheck"
        >
          <template #default="{ node, data }">
            <span :class="['workflow-user-picker__node', data.type === 'user' ? 'is-user' : 'is-dept']">
              <span>{{ node.label }}</span>
              <small v-if="data.type === 'user' && data.mobile">{{ data.mobile }}</small>
            </span>
          </template>
        </el-tree>
      </div>
    </el-popover>
  </div>
</template>

<script lang="ts" setup>
import { ArrowDown } from '@element-plus/icons-vue'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { WorkflowAssigneeUser } from '../types'

type DepartmentNode = {
  id: number
  name: string
  children?: DepartmentNode[]
}

type UserTreeNode = {
  key: string
  type: 'dept' | 'user'
  id?: number
  label: string
  mobile?: string
  pathText?: string
  disabled?: boolean
  children?: UserTreeNode[]
}

type SelectionItem = {
  key: string
  type: 'dept' | 'user'
  id: number
  label: string
}

const props = defineProps<{
  modelValue: number[]
  departmentModelValue?: number[]
  departments?: DepartmentNode[]
  users?: WorkflowAssigneeUser[]
  multiple?: boolean
  selectDepartmentRules?: boolean
  disabled?: boolean
  placeholder?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: number[]]
  'update:departmentModelValue': [value: number[]]
}>()
const treeRef = ref<any>()
const selectionRef = ref<HTMLElement>()
const keyword = ref('')
const visibleSelectionCount = ref(Number.MAX_SAFE_INTEGER)
let selectionResizeObserver: ResizeObserver | undefined
let textMeasureContext: CanvasRenderingContext2D | null = null

const normalizedValue = computed(() => normalizeIDs(props.modelValue || []))
const normalizedDepartmentValue = computed(() => normalizeIDs(props.departmentModelValue || []))
const availableUsers = computed(() => (props.users || []).filter(user => user.status === undefined || Number(user.status) === 1))
const treeData = computed(() => buildUserTree(
  props.departments || [],
  availableUsers.value,
  Boolean(props.multiple),
  Boolean(props.selectDepartmentRules),
))
const selectedUsers = computed(() => {
  const usersByID = new Map((props.users || []).map(user => [Number(user.id), user]))
  return normalizedValue.value.map(id => usersByID.get(id) || { id, name: String(id) })
})
const selectedDepartments = computed(() => {
  const departmentsByID = new Map<number, DepartmentNode>()
  walkDepartments(props.departments || [], department => departmentsByID.set(Number(department.id), department))
  return normalizedDepartmentValue.value.map(id => departmentsByID.get(id) || { id, name: String(id) })
})
const selectedItems = computed<SelectionItem[]>(() => [
  ...selectedDepartments.value.map(department => ({
    key: `dept:${department.id}`, type: 'dept' as const, id: Number(department.id), label: department.name || String(department.id),
  })),
  ...selectedUsers.value.map(user => ({
    key: `user:${user.id}`, type: 'user' as const, id: Number(user.id), label: user.name || String(user.id),
  })),
])
const visibleSelectedItems = computed(() => selectedItems.value.slice(0, visibleSelectionCount.value))
const hiddenSelectedItems = computed(() => selectedItems.value.slice(visibleSelectionCount.value))
const checkedNodeKeys = computed(() => checkedKeys(normalizedValue.value, normalizedDepartmentValue.value, treeData.value))

watch([normalizedValue, normalizedDepartmentValue, treeData], () => nextTick(syncTreeKeys), { immediate: true })
watch(selectedItems, () => nextTick(updateVisibleSelection), { immediate: true })

onMounted(() => {
  updateVisibleSelection()
  if (typeof ResizeObserver !== 'undefined' && selectionRef.value) {
    selectionResizeObserver = new ResizeObserver(updateVisibleSelection)
    selectionResizeObserver.observe(selectionRef.value)
  } else {
    window.addEventListener('resize', updateVisibleSelection)
  }
})

onBeforeUnmount(() => {
  selectionResizeObserver?.disconnect()
  window.removeEventListener('resize', updateVisibleSelection)
})

function normalizeIDs(values: Array<number | string>) {
  return Array.from(new Set(values.map(value => Number(value)).filter(value => Number.isInteger(value) && value > 0)))
}

function buildUserTree(
  departments: DepartmentNode[],
  users: WorkflowAssigneeUser[],
  selectDepartments: boolean,
  selectDepartmentRules: boolean,
) {
  const usersByDepartment = new Map<number, WorkflowAssigneeUser[]>()
  const usersWithoutDepartment: WorkflowAssigneeUser[] = []
  users.forEach((user) => {
    const departmentIDs = normalizeIDs(user.deptIds || [])
    if (!departmentIDs.length) usersWithoutDepartment.push(user)
    departmentIDs.forEach((departmentID) => {
      usersByDepartment.set(departmentID, [...(usersByDepartment.get(departmentID) || []), user])
    })
  })

  function userNode(user: WorkflowAssigneeUser, scope: string, pathText: string): UserTreeNode {
    return {
      key: `user:${scope}:${user.id}`,
      type: 'user',
      id: Number(user.id),
      label: user.name || String(user.id),
      mobile: user.mobile || '',
      pathText,
    }
  }

  function departmentNode(department: DepartmentNode, parentPath: string[] = []): UserTreeNode | null {
    const id = Number(department.id)
    if (!id) return null
    const path = [...parentPath, department.name || String(id)]
    const childDepartments = (department.children || []).map(child => departmentNode(child, path)).filter(Boolean) as UserTreeNode[]
    const departmentUsers = (usersByDepartment.get(id) || []).map(user => userNode(user, `dept-${id}`, path.join('/')))
    const children = [...childDepartments, ...departmentUsers]
    if (!children.length && !selectDepartmentRules) return null
    return {
      key: `dept:${id}`,
      type: 'dept',
      id,
      label: department.name || String(id),
      pathText: path.join('/'),
      disabled: !selectDepartments,
      children,
    }
  }

  const result = departments.map(department => departmentNode(department)).filter(Boolean) as UserTreeNode[]
  if (usersWithoutDepartment.length) {
    result.push({
      key: 'dept:unassigned',
      type: 'dept',
      label: '未分配部门',
      disabled: selectDepartmentRules || !selectDepartments,
      children: usersWithoutDepartment.map(user => userNode(user, 'unassigned', '未分配部门')),
    })
  }

  const visible = new Set<number>()
  walkTree(result, (node) => {
    if (node.type === 'user' && node.id) visible.add(node.id)
  })
  const unmatched = users.filter(user => !visible.has(Number(user.id)))
  if (unmatched.length) {
    result.push({
      key: 'dept:unmatched',
      type: 'dept',
      label: '未匹配部门',
      disabled: selectDepartmentRules || !selectDepartments,
      children: unmatched.map(user => userNode(user, 'unmatched', '未匹配部门')),
    })
  }
  return result
}

function walkDepartments(departments: DepartmentNode[], visit: (department: DepartmentNode) => void) {
  departments.forEach((department) => {
    visit(department)
    walkDepartments(department.children || [], visit)
  })
}

function walkTree(nodes: UserTreeNode[], visit: (node: UserTreeNode) => void) {
  nodes.forEach((node) => {
    visit(node)
    walkTree(node.children || [], visit)
  })
}

function checkedKeys(userIDs: number[], departmentIDs: number[], nodes: UserTreeNode[]) {
  const selectedUsers = new Set(userIDs)
  const selectedDepartments = new Set(departmentIDs)
  const result: string[] = []
  walkTree(nodes, (node) => {
    if (node.type === 'user' && node.id && selectedUsers.has(node.id)) result.push(node.key)
    if (node.type === 'dept' && node.id && selectedDepartments.has(node.id)) result.push(node.key)
  })
  return result
}

function collectDescendantUserIDs(node: UserTreeNode) {
  const result: number[] = []
  walkTree(node.children || [], (child) => {
    if (child.type === 'user' && child.id) result.push(child.id)
  })
  return normalizeIDs(result)
}

function onCheck(data: UserTreeNode) {
  nextTick(() => {
    const eventCheckedKeys: string[] = treeRef.value?.getCheckedKeys?.() || []
    const isChecked = eventCheckedKeys.includes(data.key)
    const selected = new Set(normalizedValue.value)

    if (data.type === 'dept') {
      if (!props.multiple) return
      if (props.selectDepartmentRules && data.id) {
        const selectedDepartments = new Set(normalizedDepartmentValue.value)
        if (isChecked) selectedDepartments.add(Number(data.id))
        else selectedDepartments.delete(Number(data.id))
        emit('update:departmentModelValue', Array.from(selectedDepartments))
        nextTick(syncTreeKeys)
        return
      }
      collectDescendantUserIDs(data).forEach((userID) => {
        if (isChecked) selected.add(userID)
        else selected.delete(userID)
      })
      emit('update:modelValue', Array.from(selected))
      nextTick(syncTreeKeys)
      return
    }

    if (!data.id) return
    const userID = Number(data.id)
    if (isChecked) selected.add(userID)
    else selected.delete(userID)
    const nextValue = props.multiple ? Array.from(selected) : (isChecked ? [userID] : [])
    emit('update:modelValue', nextValue)
    nextTick(syncTreeKeys)
  })
}

function syncTreeKeys() {
  treeRef.value?.setCheckedKeys(checkedNodeKeys.value)
}

function updateVisibleSelection() {
  const element = selectionRef.value
  const labels = selectedItems.value.map(item => item.label)
  if (!element || labels.length === 0) {
    visibleSelectionCount.value = labels.length
    return
  }

  const availableWidth = Math.max(0, element.clientWidth - 50)
  if (measureText(labels.join('、'), element) <= availableWidth) {
    visibleSelectionCount.value = labels.length
    return
  }

  for (let count = labels.length - 1; count >= 0; count -= 1) {
    const visibleText = labels.slice(0, count).join('、')
    const overflowText = `+${labels.length - count}`
    const requiredWidth = measureText(visibleText, element)
      + (count > 0 ? 8 : 0)
      + measureText(overflowText, element)
      + 10
    if (requiredWidth <= availableWidth) {
      visibleSelectionCount.value = count
      return
    }
  }
  visibleSelectionCount.value = 0
}

function measureText(value: string, element: HTMLElement) {
  if (!textMeasureContext) textMeasureContext = document.createElement('canvas').getContext('2d')
  if (!textMeasureContext) return value.length * 14
  const style = window.getComputedStyle(element)
  textMeasureContext.font = `${style.fontWeight} ${style.fontSize} ${style.fontFamily}`
  return textMeasureContext.measureText(value).width
}

function normalizeText(value: unknown) {
  return String(value || '').trim().toLowerCase()
}

function filterTree() {
  treeRef.value?.filter(normalizeText(keyword.value))
}

function filterNode(value: string, data: UserTreeNode) {
  const keywordValue = normalizeText(value)
  if (!keywordValue) return true
  return [data.label, data.mobile, data.pathText].some(item => normalizeText(item).includes(keywordValue))
}
</script>

<style scoped>
.workflow-user-picker { width: 100%; }
.workflow-user-picker__selection { display: flex; box-sizing: border-box; width: 100%; min-width: 0; height: 32px; align-items: center; gap: 8px; padding: 0 11px; border: 1px solid var(--el-border-color); border-radius: var(--el-border-radius-base); background: var(--el-fill-color-blank); color: var(--el-text-color-regular); cursor: pointer; transition: var(--el-transition-border); }
.workflow-user-picker__selection:hover { border-color: var(--el-border-color-hover); }
.workflow-user-picker__selection:focus-visible { border-color: var(--el-color-primary); outline: none; box-shadow: 0 0 0 1px var(--el-color-primary) inset; }
.workflow-user-picker__selection.is-disabled { border-color: var(--el-disabled-border-color); background: var(--el-disabled-bg-color); color: var(--el-disabled-text-color); cursor: not-allowed; }
.workflow-user-picker__placeholder { min-width: 0; flex: 1; overflow: hidden; color: var(--el-text-color-placeholder); text-overflow: ellipsis; white-space: nowrap; }
.workflow-user-picker__value { display: flex; min-width: 0; flex: 1; align-items: center; gap: 8px; overflow: hidden; }
.workflow-user-picker__names { min-width: 0; overflow: hidden; text-overflow: clip; white-space: nowrap; }
.workflow-user-picker__overflow { flex: 0 0 auto; padding: 1px 5px; border-radius: 3px; color: var(--el-color-primary); background: var(--el-color-primary-light-9); font-size: 12px; line-height: 20px; cursor: help; }
.workflow-user-picker__arrow { flex: 0 0 auto; color: var(--el-text-color-placeholder); font-size: 14px; }
.workflow-user-picker__tree { max-height: 330px; margin-top: 10px; overflow: auto; border-top: 1px solid #edf0f4; padding-top: 8px; }
.workflow-user-picker__node { display: inline-flex; min-width: 0; align-items: center; gap: 10px; }
.workflow-user-picker__node span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.workflow-user-picker__node small { flex: 0 0 auto; color: #98a2b3; font-size: 11px; }
:global(.workflow-user-picker__overflow-tooltip) { max-width: 360px; }
:global(.workflow-user-picker__overflow-list) { display: flex; max-height: 240px; flex-direction: column; gap: 4px; overflow: auto; }
</style>
