<template>
  <div class="user-list-filters" aria-label="用户筛选条件">
    <el-input
      :model-value="modelValue.keyword"
      aria-label="用户名或姓名"
      clearable
      placeholder="搜索用户名/姓名"
      @update:model-value="updateText('keyword', $event)"
      @keyup.enter="emit('search')"
    />
    <el-input
      :model-value="modelValue.mobile"
      aria-label="手机号"
      clearable
      placeholder="搜索手机号"
      @update:model-value="updateText('mobile', $event)"
      @keyup.enter="emit('search')"
    />
    <el-select
      :model-value="modelValue.status"
      aria-label="状态"
      clearable
      placeholder="全部状态"
      @update:model-value="updateNumber('status', $event)"
    >
      <el-option label="待审核" :value="0" />
      <el-option label="正常" :value="1" />
      <el-option label="禁用" :value="2" />
      <el-option label="管理员" :value="9" />
    </el-select>
    <el-select
      :model-value="modelValue.positionId"
      aria-label="岗位"
      clearable
      filterable
      placeholder="全部岗位"
      @update:model-value="updateNumber('positionId', $event)"
    >
      <el-option v-for="item in positions" :key="String(item.id)" :label="String(item.name || '')" :value="Number(item.id)" />
    </el-select>
    <el-select
      :model-value="modelValue.managerUserId"
      aria-label="直属上级"
      clearable
      filterable
      placeholder="全部直属上级"
      @update:model-value="updateNumber('managerUserId', $event)"
    >
      <el-option v-for="item in managers" :key="item.id" :label="managerLabel(item)" :value="item.id" />
    </el-select>
    <el-select
      :model-value="modelValue.roleId"
      aria-label="角色"
      clearable
      filterable
      placeholder="全部角色"
      @update:model-value="updateNumber('roleId', $event)"
    >
      <el-option v-for="item in roles" :key="String(item.id)" :label="String(item.name || '')" :value="Number(item.id)" />
    </el-select>
    <AdminFilterActions class="user-list-filters__actions" @search="emit('search')" @reset="emit('reset')" />
  </div>
</template>

<script setup lang="ts">
import type { AdminListRecord, AdminUser, AdminUserListQuery } from '../../../api/types'
import { AdminFilterActions } from '@/components/admin-ui'

const props = defineProps<{
  modelValue: AdminUserListQuery
  positions: AdminListRecord[]
  managers: AdminUser[]
  roles: AdminListRecord[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: AdminUserListQuery]
  search: []
  reset: []
}>()

function updateText(field: 'keyword' | 'mobile', value: unknown) {
  emit('update:modelValue', { ...props.modelValue, [field]: String(value || '') })
}

function updateNumber(field: 'status' | 'positionId' | 'managerUserId' | 'roleId', value: unknown) {
  const numberValue = value === '' || value === null || value === undefined ? undefined : Number(value)
  emit('update:modelValue', { ...props.modelValue, [field]: Number.isFinite(numberValue) ? numberValue : undefined })
}

function managerLabel(item: AdminUser) {
  return item.mobile ? `${item.name} · ${item.mobile}` : item.name
}
</script>

<style scoped>
.user-list-filters {
  box-sizing: border-box;
  width: 100%;
  padding: 12px 14px;
  display: flex;
  flex-wrap: wrap;
  gap: var(--admin-ui-space-2);
  align-items: center;
}
.user-list-filters > .el-input,
.user-list-filters > .el-select { flex: 0 1 180px; width: 180px; max-width: 100%; }
.user-list-filters > .el-input:first-child { flex-basis: 240px; width: 240px; }
@media (max-width: 767px) {
  .user-list-filters > .el-input,
  .user-list-filters > .el-input:first-child,
  .user-list-filters > .el-select { flex-basis: 100%; width: 100%; }
}
</style>
