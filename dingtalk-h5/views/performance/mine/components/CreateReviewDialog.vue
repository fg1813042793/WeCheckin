<template>
  <view v-if="dialog.visible" class="review-create-modal" @click="$emit('close')">
    <view class="review-create-card" @click.stop="$emit('close-month-picker')">
      <view class="review-create-head">
        <view>
          <text class="review-create-title">新建考评单</text>
          <text class="review-create-desc">选择被考评人和考评月份，支持按部门多选创建</text>
        </view>
        <button class="review-create-close" @click="$emit('close')">×</button>
      </view>
      <view class="review-create-form">
        <view class="review-create-field review-create-field-users">
          <text class="review-create-label">被考评人</text>
          <view class="create-target-search">
            <input
              class="field-input create-target-search-input"
              :value="userKeyword"
              placeholder="搜索姓名/账号/部门/岗位"
              @input="$emit('update:userKeyword', readInputValue($event))"
            />
            <button
              v-if="userKeyword"
              class="create-target-search-clear"
              @click="$emit('update:userKeyword', '')"
            >清空</button>
          </view>
          <view class="department-user-tree">
            <view
              v-for="row in treeRows"
              :key="row.key"
              :class="['create-target-row', row.type, `depth-${row.depth}`]"
            >
              <view
                v-if="row.type === 'department'"
                class="create-target-dept-head"
                :class="{ expanded: row.expanded }"
                @click="$emit('toggle-dept', row.key)"
              >
                <view
                  class="create-target-dept-check"
                  :class="[
                    departmentCheckState(row) === 'checked' ? 'checked' : '',
                    departmentCheckState(row) === 'indeterminate' ? 'create-target-dept-check-indeterminate' : ''
                  ]"
                  @click.stop="$emit('toggle-department', row)"
                >
                  <text v-if="departmentCheckState(row) === 'checked'">✓</text>
                  <text v-else-if="departmentCheckState(row) === 'indeterminate'">-</text>
                </view>
                <view class="create-target-dept-title">
                  <text class="create-target-dept-chevron" :class="{ expanded: row.expanded }"></text>
                  <text class="create-target-dept-name">{{ row.name }}</text>
                </view>
                <text class="create-target-dept-count">{{ departmentUserIds(row).length }} 人</text>
              </view>
              <view
                v-else-if="row.type === 'employee'"
                class="create-target-user-tree"
                :class="{ selected: form.employeeIds.includes(row.user.id) }"
                @click="$emit('toggle-employee', row.user.id)"
              >
                <view class="create-target-check">
                  <text v-if="form.employeeIds.includes(row.user.id)">✓</text>
                </view>
                <view class="create-target-user-main">
                  <text class="create-target-user-name">{{ row.user.name || row.user.id }}</text>
                  <text class="create-target-user-meta">{{ userMeta(row.user) }}</text>
                </view>
              </view>
            </view>
            <view v-if="treeEmpty" class="create-target-empty">{{ emptyText }}</view>
          </view>
        </view>
        <view class="review-create-inline-fields single">
          <view class="review-create-field">
            <text class="review-create-label">考评月份</text>
            <view class="review-create-month-picker">
              <button
                class="field-input review-create-month-trigger"
                :class="{ selected: form.period }"
                @click.stop="$emit('toggle-month-picker')"
              >
                <text class="review-create-month-text">{{ monthText(form.period) }}</text>
                <text class="review-create-month-arrow" :class="{ open: monthPickerOpen }"></text>
              </button>
              <view v-if="monthPickerOpen" class="review-create-month-dropdown" @click.stop>
                <view class="review-create-month-head">
                  <button class="review-create-month-nav" @click="$emit('change-month-year', -1)">‹</button>
                  <text class="review-create-month-year">{{ monthPickerYear }}年</text>
                  <button class="review-create-month-nav" @click="$emit('change-month-year', 1)">›</button>
                </view>
                <view class="review-create-month-grid">
                  <button
                    v-for="month in monthOptions"
                    :key="month.value"
                    class="review-create-month-option"
                    :class="{ active: isMonthSelected(month.value) }"
                    @click="$emit('select-month', month.value)"
                  >{{ month.label }}</button>
                </view>
              </view>
            </view>
          </view>
        </view>
      </view>
      <view class="review-create-actions">
        <button class="dt-btn dt-btn-light" @click="$emit('close')">取消</button>
        <button class="dt-btn dt-btn-primary" :loading="dialog.loading" @click="$emit('create')">创建</button>
      </view>
    </view>
  </view>
</template>

<script setup>
const props = defineProps({
  departmentCheckState: { type: Function, required: true },
  departmentUserIds: { type: Function, required: true },
  dialog: { type: Object, required: true },
  emptyText: { type: String, default: '暂无可创建人员' },
  form: { type: Object, required: true },
  isMonthSelected: { type: Function, required: true },
  monthOptions: { type: Array, default: () => [] },
  monthPickerOpen: { type: Boolean, default: false },
  monthPickerYear: { type: Number, default: new Date().getFullYear() },
  monthText: { type: Function, required: true },
  treeEmpty: { type: Boolean, default: true },
  treeRows: { type: Array, default: () => [] },
  userKeyword: { type: String, default: '' },
  userMeta: { type: Function, required: true }
})

defineEmits([
  'change-month-year',
  'close',
  'close-month-picker',
  'create',
  'select-month',
  'toggle-department',
  'toggle-dept',
  'toggle-employee',
  'toggle-month-picker',
  'update:userKeyword'
])

function readInputValue(event) {
  return event.detail?.value ?? event.target?.value ?? ''
}
</script>

<style>
/* 页面专属样式：从 styles/performance.css 拆分 */
.review-create-modal {
  position: fixed;
  inset: 0;
  z-index: 120;
  padding: 40px 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(17, 24, 39, 0.36);
  overscroll-behavior: contain;
  pointer-events: auto;
  touch-action: none;
}

.review-create-card {
  width: min(860px, 100%);
  max-height: calc(100vh - 80px);
  border: 1px solid #e5e6eb;
  border-radius: 8px;
  display: grid;
  grid-template-rows: auto minmax(0, 1fr) auto;
  background: #fff;
  box-shadow: 0 24px 70px rgba(31, 35, 41, 0.18);
  overflow: hidden;
  overscroll-behavior: contain;
  pointer-events: auto;
  touch-action: auto;
}

.review-create-head,
.review-create-actions {
  padding: 16px 18px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.review-create-head {
  position: relative;
  padding-right: 58px;
  border-bottom: 1px solid #edf0f5;
}

.review-create-actions {
  border-top: 1px solid #edf0f5;
  justify-content: flex-end;
}

.review-create-title,
.review-create-label {
  display: block;
  color: #1f2329;
  font-weight: 700;
}

.review-create-title {
  font-size: 17px;
}

.review-create-desc {
  display: block;
  margin-top: 4px;
  color: #86909c;
  font-size: 13px;
}

.review-create-close {
  position: absolute;
  top: 50%;
  right: 16px;
  transform: translateY(-50%);
  width: 32px;
  height: 32px;
  min-height: 32px;
  padding: 0;
  border: 0;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  color: #86909c;
  font-size: 22px;
  line-height: 30px;
}

.review-create-close:hover {
  background: #f2f3f5;
  color: #4e5969;
}

.review-create-form {
  min-height: 0;
  padding: 18px;
  display: grid;
  gap: 16px;
  overflow: auto;
}

.review-create-field {
  min-width: 0;
  display: grid;
  gap: 8px;
}

.review-create-field-users {
  min-height: 260px;
}

.create-target-search {
  display: flex;
  align-items: center;
  gap: 8px;
}

.create-target-search-input {
  flex: 1 1 auto;
  min-width: 0;
}

.create-target-search-clear {
  width: 58px;
  height: 32px;
  min-height: 32px;
  padding: 0;
  border: 1px solid #d9e6ff;
  border-radius: 4px;
  background: #f2f7ff;
  color: #1677ff;
  font-size: 13px;
  font-weight: 600;
  line-height: 30px;
  cursor: pointer;
}

.create-target-search-clear:hover {
  border-color: #91caff;
  background: #eaf3ff;
}

.review-create-inline-fields {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 180px));
  gap: 14px;
}

.review-create-inline-fields.single {
  grid-template-columns: minmax(0, 180px);
}

.review-create-month-picker {
  position: relative;
  width: 180px;
  max-width: 100%;
  z-index: 8;
}

.review-create-month-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  color: #86909c;
  text-align: left;
  cursor: pointer;
}

.review-create-month-trigger.selected {
  color: #1f2329;
}

.review-create-month-text {
  min-width: 0;
  flex: 1 1 auto;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.review-create-month-arrow {
  position: relative;
  width: 14px;
  height: 14px;
  flex: 0 0 14px;
}

.review-create-month-arrow::before {
  content: '';
  position: absolute;
  left: 4px;
  top: 4px;
  width: 6px;
  height: 6px;
  border-right: 1.5px solid #86909c;
  border-bottom: 1.5px solid #86909c;
  transform: rotate(45deg);
  transition: transform 0.16s ease;
}

.review-create-month-arrow.open::before {
  top: 6px;
  transform: rotate(225deg);
}

.review-create-month-dropdown {
  position: absolute;
  left: 0;
  bottom: calc(100% + 6px);
  width: 300px;
  max-width: calc(100vw - 48px);
  padding: 14px;
  border: 1px solid #e5eaf3;
  border-radius: 10px;
  background: #fff;
  box-shadow: 0 16px 40px rgba(31, 35, 41, 0.14);
  z-index: 60;
  transform-origin: left bottom;
}

.review-create-month-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  padding: 0 2px;
}

.review-create-month-year {
  color: #1f2329;
  font-size: 15px;
  font-weight: 700;
}

.review-create-month-nav {
  width: 32px;
  height: 32px;
  min-height: 32px;
  padding: 0;
  border: 1px solid #e5e6eb;
  border-radius: 8px;
  background: #f7f9fc;
  color: #4e5969;
  font-size: 18px;
  font-weight: 700;
  line-height: 30px;
  cursor: pointer;
}

.review-create-month-nav:hover {
  border-color: #91caff;
  background: #f2f7ff;
  color: #1677ff;
}

.review-create-month-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.review-create-month-option {
  width: 100%;
  height: 38px;
  min-height: 38px;
  margin: 0;
  padding: 0 8px;
  border: 1px solid #e5eaf3;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f8fafc;
  color: #4e5969;
  font-size: 13px;
  font-weight: 600;
  line-height: 1;
  cursor: pointer;
  transition: border-color 0.16s ease, background-color 0.16s ease, color 0.16s ease, box-shadow 0.16s ease;
}

.review-create-month-option:hover {
  border-color: #1677ff;
  background: #eaf3ff;
  color: #1677ff;
}

.review-create-month-option.active {
  border-color: #1677ff;
  background: #1677ff;
  color: #fff;
  font-weight: 800;
  box-shadow: 0 6px 14px rgba(22, 119, 255, 0.22);
}

.create-target-row + .create-target-row,
.create-target-dept + .create-target-dept {
  border-top: 1px solid #f2f3f5;
}

.create-target-dept-head {
  height: 40px;
  padding: 0 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  background: #fafafa;
  cursor: pointer;
  transition: background 0.16s ease;
}

.create-target-dept-head.expanded {
  background: #f5f8ff;
}

.create-target-dept-head:hover {
  background: #f5f8ff;
}

.create-target-dept-title {
  min-width: 0;
  flex: 1 1 auto;
  display: flex;
  align-items: center;
  gap: 8px;
}

.create-target-dept-check {
  width: 18px;
  height: 18px;
  flex: 0 0 18px;
  border: 1px solid #c9d3e2;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  color: #1677ff;
  font-size: 12px;
  font-weight: 800;
  transition: border-color 0.16s ease, background 0.16s ease, color 0.16s ease;
}

.create-target-dept-check.checked,
.create-target-dept-check-indeterminate {
  border-color: #1677ff;
  background: #1677ff;
  color: #fff;
}

.create-target-dept-chevron {
  width: 16px;
  height: 16px;
  flex: 0 0 16px;
  border-radius: 4px;
  position: relative;
  color: #86909c;
  transition: background 0.16s ease;
}

.create-target-dept-chevron::before {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 6px;
  height: 6px;
  border-top: 1.5px solid currentColor;
  border-right: 1.5px solid currentColor;
  transform: translate(-60%, -50%) rotate(45deg);
  transform-origin: center;
  transition: transform 0.16s ease;
  content: '';
}

.create-target-dept-chevron.expanded {
  color: #1677ff;
}

.create-target-dept-chevron.expanded::before {
  transform: translate(-50%, -65%) rotate(135deg);
}

.create-target-dept-head:hover .create-target-dept-chevron {
  background: #e8f3ff;
}

.create-target-dept-name {
  min-width: 0;
  color: #1f2329;
  font-size: 13px;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.create-target-dept-count {
  flex: 0 0 auto;
  padding-left: 12px;
  color: #86909c;
  font-size: 12px;
}

.create-target-row.depth-2 .create-target-dept-head,
.create-target-row.depth-2 .create-target-user-tree {
  padding-left: 30px;
}

.create-target-row.depth-3 .create-target-dept-head,
.create-target-row.depth-3 .create-target-user-tree {
  padding-left: 46px;
}

.create-target-row.depth-4 .create-target-dept-head,
.create-target-row.depth-4 .create-target-user-tree,
.create-target-row.depth-5 .create-target-dept-head,
.create-target-row.depth-5 .create-target-user-tree {
  padding-left: 62px;
}

.create-target-user-list {
  padding: 6px;
  display: grid;
  gap: 4px;
}

.create-target-user-tree {
  min-height: 46px;
  padding: 7px 14px;
  border-radius: 0;
  display: flex;
  align-items: center;
  gap: 10px;
  background: #fff;
  cursor: pointer;
  transition: background 0.16s ease, box-shadow 0.16s ease;
}

.create-target-user-tree:hover,
.create-target-user-tree.selected {
  background: #f0f7ff;
}

.create-target-user-tree.selected {
  box-shadow: inset 0 0 0 1px #91caff;
}

.create-target-check {
  width: 18px;
  height: 18px;
  flex: 0 0 18px;
  border: 1px solid #c9d3e2;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  color: #1677ff;
  font-size: 12px;
  font-weight: 800;
}

.create-target-user-tree.selected .create-target-check {
  border-color: #1677ff;
  background: #1677ff;
  color: #fff;
}

.create-target-user-main {
  min-width: 0;
  display: grid;
  gap: 3px;
}

.create-target-user-name {
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
}

.create-target-user-meta {
  color: #86909c;
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.create-target-empty {
  padding: 48px 16px;
  color: #86909c;
  font-size: 13px;
  text-align: center;
}

@media (max-width: 960px) {
  .review-create-modal {
    padding: 12px;
    align-items: flex-end;
  }
  .review-create-card {
    width: 100%;
    max-height: calc(100vh - 32px);
    border-radius: 8px;
  }
  .review-create-head,
  .review-create-actions,
  .review-create-form {
    padding: 16px;
  }
  .review-create-actions .dt-btn {
    flex: 1 1 0;
  }
  .review-create-inline-fields {
    grid-template-columns: 1fr;
  }
  .review-create-month-picker {
    width: 100%;
  }
  .review-create-month-dropdown {
    width: 100%;
    max-width: none;
  }
  .review-create-month-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
  .create-target-search {
    gap: 6px;
  }
  .create-target-search-clear {
    width: 52px;
    font-size: 12px;
  }
}
</style>
