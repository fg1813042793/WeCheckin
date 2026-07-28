<template>
  <view class="org-page">
    <view class="page-head">
      <view>
        <text class="page-title">流程执行</text>
        <text class="page-desc">维护流程人员、角色、直属上级和 HRBP 关系</text>
      </view>
      <button class="dt-btn dt-btn-light" @click="ctx.refreshData">刷新</button>
    </view>

    <view class="org-toolbar">
      <view class="org-stat">
        <text class="org-stat-value">{{ ctx.state.users.length }}</text>
        <text class="org-stat-label">人员总数</text>
      </view>
      <view class="org-stat">
        <text class="org-stat-value">{{ employeeCount }}</text>
        <text class="org-stat-label">业务人员</text>
      </view>
      <view class="org-stat">
        <text class="org-stat-value">{{ hrbpCount }}</text>
        <text class="org-stat-label">HRBP</text>
      </view>
      <view class="org-stat">
        <text class="org-stat-value">{{ managerCount }}</text>
        <text class="org-stat-label">管理角色</text>
      </view>
    </view>

    <section v-if="isAdmin" class="panel create-panel">
      <view class="panel-head">
        <text class="panel-title">新增人员</text>
        <text class="count-pill">默认密码 123456</text>
      </view>
      <view class="org-create-grid">
        <input v-model="ctx.newUser.id" class="field-input" placeholder="账号" />
        <input v-model="ctx.newUser.name" class="field-input" placeholder="姓名" />
        <input v-model="ctx.newUser.password" class="field-input" placeholder="初始密码" />
        <select v-model="ctx.newUser.role" class="field-select">
          <option v-for="role in roleOptions" :key="role" :value="role">{{ ctx.roleText(role) }}</option>
        </select>
        <input v-model="ctx.newUser.position" class="field-input" placeholder="岗位" />
        <input v-model="ctx.newUser.departmentLevel1" class="field-input" placeholder="一级部门" />
        <input v-model="ctx.newUser.departmentLevel2" class="field-input" placeholder="二级部门" />
        <input v-model="ctx.newUser.departmentLevel3" class="field-input" placeholder="三级部门" />
        <select v-model="ctx.newUser.managerId" class="field-select">
          <option value="">无直属上级</option>
          <option v-for="user in managerOptions(ctx.newUser.id)" :key="user.id" :value="user.id">
            {{ user.name }} · {{ ctx.roleText(user.role) }}
          </option>
        </select>
        <select v-model="ctx.newUser.hrbpId" class="field-select">
          <option value="">无HRBP</option>
          <option v-for="user in hrbpOptions" :key="user.id" :value="user.id">
            {{ user.name }} · {{ ctx.roleText(user.role) }}
          </option>
        </select>
        <input v-model="ctx.newUser.responsibleDepartments" class="field-input org-wide-field" placeholder="负责部门，多个用逗号分隔" />
        <button class="dt-btn dt-btn-primary" @click="ctx.createUser">添加</button>
      </view>
    </section>

    <section class="panel org-list-panel">
      <view class="panel-head">
        <text class="panel-title">人员列表</text>
        <text class="count-pill">{{ ctx.state.users.length }} 人</text>
      </view>

      <view class="org-table-shell">
        <table class="org-table">
          <thead>
            <tr>
              <th>人员</th>
              <th>角色</th>
              <th>岗位</th>
              <th>部门</th>
              <th>直属上级</th>
              <th>HRBP</th>
              <th>负责部门</th>
              <th>密码</th>
              <th v-if="isAdmin">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in ctx.state.users" :key="user.id">
              <td>
                <view class="org-person-cell">
                  <view class="org-avatar">{{ initials(user.name || user.id) }}</view>
                  <view class="org-person-main">
                    <input v-model="user.name" class="field-input compact-input" :disabled="!isAdmin" placeholder="姓名" />
                    <text class="org-account">{{ user.id }}</text>
                  </view>
                </view>
              </td>
              <td>
                <select v-model="user.role" class="field-select compact-select" :disabled="!isAdmin">
                  <option v-for="role in roleOptions" :key="role" :value="role">{{ ctx.roleText(role) }}</option>
                </select>
              </td>
              <td>
                <input v-model="user.position" class="field-input compact-input" :disabled="!isAdmin" placeholder="岗位" />
              </td>
              <td>
                <view class="org-dept-fields">
                  <input v-model="user.departmentLevel1" class="field-input compact-input" :disabled="!isAdmin" placeholder="一级部门" />
                  <input v-model="user.departmentLevel2" class="field-input compact-input" :disabled="!isAdmin" placeholder="二级部门" />
                  <input v-model="user.departmentLevel3" class="field-input compact-input" :disabled="!isAdmin" placeholder="三级部门" />
                </view>
              </td>
              <td>
                <select v-model="user.managerId" class="field-select compact-select" :disabled="!isAdmin">
                  <option value="">无直属上级</option>
                  <option v-for="item in managerOptions(user.id)" :key="item.id" :value="item.id">
                    {{ item.name }} · {{ ctx.roleText(item.role) }}
                  </option>
                </select>
              </td>
              <td>
                <select v-model="user.hrbpId" class="field-select compact-select" :disabled="!isAdmin">
                  <option value="">无HRBP</option>
                  <option v-for="item in hrbpOptions" :key="item.id" :value="item.id">
                    {{ item.name }} · {{ ctx.roleText(item.role) }}
                  </option>
                </select>
              </td>
              <td>
                <input
                  class="field-input compact-input"
                  :value="responsibleValue(user)"
                  :disabled="!isAdmin"
                  placeholder="负责部门"
                  @input="updateResponsible(user, $event)"
                />
              </td>
              <td>
                <input v-model="user.password" class="field-input compact-input" :disabled="!isAdmin" placeholder="留空不改密码" />
              </td>
              <td v-if="isAdmin">
                <view class="table-actions">
                  <button class="dt-btn dt-btn-primary small" @click="ctx.saveUser(user)">保存</button>
                  <button class="dt-btn dt-btn-danger-light small" @click="ctx.deleteUser(user.id)">删除</button>
                </view>
              </td>
            </tr>
          </tbody>
        </table>
      </view>

      <view class="user-card-list">
        <view v-for="user in ctx.state.users" :key="user.id" class="user-card org-mobile-card">
          <view class="user-card-head">
            <view class="org-person-cell">
              <view class="org-avatar">{{ initials(user.name || user.id) }}</view>
              <view>
                <text class="user-name">{{ user.name }} · {{ user.id }}</text>
                <text class="review-meta">{{ user.position || '未设置岗位' }}</text>
                <text class="review-meta">{{ departmentText(user) }}</text>
              </view>
            </view>
            <text class="count-pill">{{ ctx.roleText(user.role) }}</text>
          </view>
          <view class="org-mobile-meta">
            <text>岗位：{{ user.position || '未设置' }}</text>
            <text>上级：{{ ctx.userName(user.managerId) }}</text>
            <text>HRBP：{{ ctx.userName(user.hrbpId) }}</text>
            <text>负责：{{ responsibleValue(user) || '未设置' }}</text>
          </view>
          <view v-if="isAdmin" class="row-actions">
            <button class="dt-btn dt-btn-primary small" @click="ctx.saveUser(user)">保存</button>
            <button class="dt-btn dt-btn-danger-light small" @click="ctx.deleteUser(user.id)">删除</button>
          </view>
        </view>
      </view>
    </section>
  </view>
</template>

<script setup>
import { computed } from 'vue'
import { usePerformanceContext } from './context'

const ctx = usePerformanceContext()

const roleOptions = computed(() => Object.keys(ctx.roleName))
const isAdmin = computed(() => ctx.state.user?.role === 'admin')
const hrbpOptions = computed(() => ctx.state.users.filter((user) => ['hrbp', 'hrbp_manager', 'admin'].includes(user.role)))
const employeeCount = computed(() => ctx.state.users.filter((user) => user.role === 'employee').length)
const hrbpCount = computed(() => ctx.state.users.filter((user) => ['hrbp', 'hrbp_manager', 'admin'].includes(user.role)).length)
const managerCount = computed(() => ctx.state.users.filter((user) => ['director', 'manager', 'supervisor'].includes(user.role)).length)

function managerOptions(currentId) {
  return ctx.state.users.filter((user) => user.id !== currentId)
}

function responsibleValue(user) {
  return Array.isArray(user.responsibleDepartments) ? user.responsibleDepartments.join(',') : (user.responsibleDepartments || '')
}

function updateResponsible(user, event) {
  user.responsibleDepartments = event.detail?.value ?? event.target.value
}

function departmentText(user) {
  return [user.departmentLevel1, user.departmentLevel2, user.departmentLevel3].filter(Boolean).join(' / ') || user.department || '未设置部门'
}

function initials(value) {
  return String(value || '人').slice(0, 1).toUpperCase()
}
</script>
