<template>
  <view class="org-config-modal-mask" @click.self="emit('close')">
    <view class="org-config-modal">
      <view class="org-config-modal-head">
        <view>
          <text class="org-config-title">配置流程审批人</text>
          <text class="org-config-subtitle">
            {{ props.selectedUser.name || props.selectedUser.id }} · {{ departmentText(props.selectedUser) }}
          </text>
        </view>
        <button class="process-modal-close" @click="emit('close')">×</button>
      </view>

      <view class="org-config-user">
        <view class="org-avatar">{{ initials(props.selectedUser.name || props.selectedUser.id) }}</view>
        <view>
          <text class="user-name">{{ props.selectedUser.name || props.selectedUser.id }}</text>
          <text class="org-account">
            {{ [props.selectedUser.id, props.selectedUser.position || '未设置岗位'].filter(Boolean).join(' · ') }}
          </text>
        </view>
      </view>

      <view class="org-config-form">
        <view class="org-form-field relation-picker-field" :class="{ active: props.activeRelationPicker === 'manager' }">
          <text>直属上级</text>
          <button
            class="relation-picker-trigger"
            :class="{ active: props.activeRelationPicker === 'manager' }"
            @click="emit('toggle-relation-picker', 'manager')"
          >
            <text class="relation-picker-value">{{ relationPickerText(props.configForm.managerId, '无直属上级') }}</text>
            <text :class="['relation-picker-arrow', props.activeRelationPicker === 'manager' ? 'expanded' : '']"></text>
          </button>
          <view v-if="props.activeRelationPicker === 'manager'" class="relation-picker-panel">
            <button class="relation-picker-clear" @click="emit('select-relation-user', 'manager', '')">无直属上级</button>
            <view class="relation-picker-tree">
              <view
                v-for="row in props.managerPickerRows"
                :key="`manager-${row.key}`"
                :class="['relation-picker-row', row.type, `depth-${row.depth}`]"
              >
                <button
                  v-if="row.type === 'department'"
                  class="relation-picker-dept"
                  @click="emit('toggle-relation-department', 'manager', row)"
                >
                  <text :class="['relation-picker-chevron', row.expandable ? (row.expanded ? 'expanded' : 'collapsed') : 'placeholder']"></text>
                  <text class="relation-picker-dept-name">{{ row.name }}</text>
                  <text class="relation-picker-count">{{ row.count }} 人</text>
                </button>
                <button
                  v-else
                  class="relation-picker-user"
                  :class="{ selected: props.configForm.managerId === row.user.id }"
                  @click="emit('select-relation-user', 'manager', row.user.id)"
                >
                  <text class="relation-picker-radio" :class="{ checked: props.configForm.managerId === row.user.id }"></text>
                  <view class="relation-picker-user-main">
                    <text class="relation-picker-user-name">{{ row.user.name || row.user.id }}</text>
                    <text class="relation-picker-user-meta">
                      {{ [row.user.position, departmentText(row.user)].filter(Boolean).join(' · ') }}
                    </text>
                  </view>
                </button>
              </view>
            </view>
          </view>
        </view>

        <view class="org-form-field relation-picker-field" :class="{ active: props.activeRelationPicker === 'hrbp' }">
          <text>HRBP</text>
          <button
            class="relation-picker-trigger"
            :class="{ active: props.activeRelationPicker === 'hrbp' }"
            @click="emit('toggle-relation-picker', 'hrbp')"
          >
            <text class="relation-picker-value">{{ relationPickerText(props.configForm.hrbpId, '无HRBP') }}</text>
            <text :class="['relation-picker-arrow', props.activeRelationPicker === 'hrbp' ? 'expanded' : '']"></text>
          </button>
          <view v-if="props.activeRelationPicker === 'hrbp'" class="relation-picker-panel">
            <button class="relation-picker-clear" @click="emit('select-relation-user', 'hrbp', '')">无HRBP</button>
            <view class="relation-picker-tree">
              <view
                v-for="row in props.hrbpPickerRows"
                :key="`hrbp-${row.key}`"
                :class="['relation-picker-row', row.type, `depth-${row.depth}`]"
              >
                <button
                  v-if="row.type === 'department'"
                  class="relation-picker-dept"
                  @click="emit('toggle-relation-department', 'hrbp', row)"
                >
                  <text :class="['relation-picker-chevron', row.expandable ? (row.expanded ? 'expanded' : 'collapsed') : 'placeholder']"></text>
                  <text class="relation-picker-dept-name">{{ row.name }}</text>
                  <text class="relation-picker-count">{{ row.count }} 人</text>
                </button>
                <button
                  v-else
                  class="relation-picker-user"
                  :class="{ selected: props.configForm.hrbpId === row.user.id }"
                  @click="emit('select-relation-user', 'hrbp', row.user.id)"
                >
                  <text class="relation-picker-radio" :class="{ checked: props.configForm.hrbpId === row.user.id }"></text>
                  <view class="relation-picker-user-main">
                    <text class="relation-picker-user-name">{{ row.user.name || row.user.id }}</text>
                    <text class="relation-picker-user-meta">
                      {{ [row.user.position, departmentText(row.user)].filter(Boolean).join(' · ') }}
                    </text>
                  </view>
                </button>
              </view>
            </view>
          </view>
        </view>
      </view>

      <view class="process-modal-actions">
        <button class="dt-btn dt-btn-light" @click="emit('close')">取消</button>
        <button class="dt-btn dt-btn-primary" @click="emit('save')">保存配置</button>
      </view>
    </view>
  </view>
</template>

<script setup>
import { usePerformanceContext } from '../../common/context'
import {
  departmentText,
  initials,
  userOptionText
} from '../composables/useOrgDirectory'

const props = defineProps({
  selectedUser: {
    type: Object,
    required: true
  },
  configForm: {
    type: Object,
    required: true
  },
  activeRelationPicker: {
    type: String,
    default: ''
  },
  managerPickerRows: {
    type: Array,
    default: () => []
  },
  hrbpPickerRows: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits([
  'close',
  'save',
  'toggle-relation-picker',
  'toggle-relation-department',
  'select-relation-user'
])

const ctx = usePerformanceContext()

function relationPickerText(id, emptyText) {
  if (!id) return emptyText
  const user = ctx.state.users.find((item) => item.id === id)
  return user ? userOptionText(user) : id
}
</script>

<style>
/* 页面专属样式：从 styles/performance.css 拆分 */
.org-config-btn {
  justify-self: end;
  white-space: nowrap;
}

.org-config-modal-mask {
  position: fixed;
  inset: 0;
  z-index: 120;
  padding: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(15, 23, 42, 0.34);
  backdrop-filter: blur(2px);
  overscroll-behavior: contain;
  pointer-events: auto;
  touch-action: none;
}

.org-config-modal {
  width: min(560px, 100%);
  height: min(640px, calc(100vh - 48px));
  max-height: calc(100vh - 48px);
  border-radius: 12px;
  display: grid;
  grid-template-rows: auto auto minmax(0, 1fr) auto;
  background: #fff;
  box-shadow: 0 24px 64px rgba(15, 23, 42, 0.2);
  overflow: hidden;
  overscroll-behavior: contain;
  pointer-events: auto;
  touch-action: auto;
}

.org-config-modal-head {
  padding: 20px 22px 14px;
  border-bottom: 1px solid #eef1f6;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.org-config-title {
  display: block;
  color: #1f2329;
  font-size: 18px;
  font-weight: 800;
}

.org-config-subtitle {
  display: block;
  margin-top: 6px;
  color: #86909c;
  font-size: 12px;
}

.org-config-user {
  margin: 18px 22px 0;
  padding: 14px;
  border: 1px solid #e8edf5;
  border-radius: 10px;
  display: flex;
  align-items: center;
  gap: 12px;
  background: #fbfdff;
}

.org-config-form {
  min-height: 0;
  padding: 18px 22px 4px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  align-content: start;
  gap: 14px;
  overflow: visible;
}

.org-form-field {
  min-width: 0;
  display: grid;
  gap: 6px;
}

.org-form-field > text {
  color: #86909c;
  font-size: 12px;
  font-weight: 700;
}

.org-form-field.span-2 {
  grid-column: span 2;
}

@media (max-width: 960px) {
  .org-config-btn {
    grid-area: action;
    align-self: center;
    justify-self: start;
    min-height: 28px;
    padding: 0 9px;
    border-radius: 7px;
    line-height: 28px;
  }

  .org-config-modal-mask {
    align-items: center;
    padding: 24px 12px;
  }

  .org-config-modal {
    width: 100%;
    height: min(620px, calc(100vh - 96px));
    max-height: calc(100vh - 96px);
    border-radius: 12px;
  }

  .org-config-modal-head,
  .org-config-form {
    padding-left: 16px;
    padding-right: 16px;
  }

  .org-config-user {
    margin-left: 16px;
    margin-right: 16px;
  }

  .org-config-form {
    grid-template-columns: minmax(0, 1fr);
  }

  .org-form-field.span-2 {
    grid-column: auto;
  }
}
</style>
