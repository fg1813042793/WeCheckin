<template>
  <div class="admin-page user-page">
    <el-card class="admin-card" shadow="never">
      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-input v-model="keyword" placeholder="搜索用户名/姓名" clearable style="width:300px" @keyup.enter="search" />
          <el-button type="primary" @click="search">搜索</el-button>
        </div>
      </div>
      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-button v-if="hasPerm('admin:menu:user:add')" type="success" @click="showAdd">+ 新增用户</el-button>
          <el-button v-if="hasPerm('admin:menu:user:del')" type="danger" :disabled="selected.length === 0" @click="delSelected">批量删除</el-button>
        </div>
        <div class="admin-toolbar__right">
          <el-button circle icon="Refresh" title="刷新" @click="load" />
          <el-button circle icon="Download" title="导出" @click="exportData" />
          <SortPopover :columns="sortColumns" v-model="sortRules" @change="onSortChange" />
        </div>
      </div>
          <el-table :data="list" v-loading="loading" stripe style="width:100%" @selection-change="selected = $event">
            <el-table-column type="selection" width="45" />
            <el-table-column label="头像" width="70">
              <template #default="{ row }">
                <span class="user-avatar" :class="{ 'user-avatar--image': avatarImageReady(row) }">
                  <span v-if="!avatarImageReady(row)" class="user-avatar__initial">{{ avatarInitial(row) }}</span>
                  <img
                    v-if="preferredAvatarUrl(row)"
                    :class="{ 'is-loaded': avatarImageReady(row) }"
                    :src="preferredAvatarUrl(row)"
                    :alt="`${row.name || '用户'}头像`"
                    @load="onAvatarLoad(row)"
                    @error="onAvatarError(row)"
                  />
                </span>
              </template>
            </el-table-column>
            <el-table-column prop="name" label="用户名" width="120" show-overflow-tooltip class-name="user-name-column" />
            <el-table-column prop="mobile" label="手机号" width="130" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="loginCnt" label="登录次数" width="80" />
            <el-table-column label="所属部门" min-width="260">
              <template #default="{ row }">
                <el-tooltip
                  v-if="displayDeptPathItems(row).length"
                  :content="deptPathTooltip(row)"
                  placement="top"
                  :show-after="300"
                >
                  <div class="dept-path-list dept-path-list--compact">
                    <span v-for="item in displayDeptPathItems(row)" :key="item.key" class="dept-path-pill">
                      <template v-for="(segment, index) in item.segments" :key="`${item.key}-${index}`">
                        <span :class="['dept-path-segment', index === item.segments.length - 1 ? 'leaf' : 'parent']">
                          {{ segment }}
                        </span>
                        <span v-if="index < item.segments.length - 1" class="dept-path-separator">/</span>
                      </template>
                    </span>
                    <span v-if="hiddenDeptPathCount(row) > 0" class="compact-more-badge">+{{ hiddenDeptPathCount(row) }}</span>
                  </div>
                </el-tooltip>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="岗位" width="140">
              <template #default="{ row }">{{ row.positionName || '-' }}</template>
            </el-table-column>
            <el-table-column label="绑定角色" width="280">
              <template #default="{ row }">
                <el-tooltip
                  v-if="displayRoleNames(row).length"
                  :disabled="displayRoleNames(row).length <= 1"
                  :content="roleNamesTooltip(row)"
                  placement="top"
                  :show-after="300"
                >
                  <div class="role-tag-list role-tag-list--compact">
                    <el-tag type="primary" size="small">{{ displayRoleNames(row)[0] }}</el-tag>
                    <span v-if="hiddenRoleCount(row) > 0" class="compact-more-badge">+{{ hiddenRoleCount(row) }}</span>
                  </div>
                </el-tooltip>
                <span v-else style="color:#999">未绑定</span>
              </template>
            </el-table-column>
            <el-table-column label="注册时间" width="160">
              <template #default="{ row }">{{ fmtTime(row.addTime) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="220" fixed="right">
              <template #default="{ row }">
                <div class="admin-table-actions">
                  <el-button v-if="hasPerm('admin:menu:user:list')" size="small" @click="showDetail(row)">详情</el-button>
                  <el-button v-if="hasPerm('admin:menu:user:edit')" size="small" type="primary" @click="showEdit(row)">编辑</el-button>
                  <el-dropdown v-if="hasRowMoreActions(row)" trigger="click" @command="(cmd:string) => handleRowCommand(cmd, row)">
                    <el-button size="small">
                      更多<el-icon><ArrowDown /></el-icon>
                    </el-button>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item v-if="hasPerm('admin:menu:user:edit') && row.status === 0" command="approve">审核通过</el-dropdown-item>
                        <el-dropdown-item v-if="hasPerm('admin:menu:user:edit') && row.status === 1" command="disable">禁用</el-dropdown-item>
                        <el-dropdown-item v-if="hasPerm('admin:menu:user:edit') && row.status === 2" command="enable">恢复正常</el-dropdown-item>
                        <el-dropdown-item v-if="hasPerm('admin:menu:user:edit')" command="resetPwd" divided>重置密码</el-dropdown-item>
                        <el-dropdown-item v-if="hasPerm('admin:menu:user:del')" command="delete" divided>删除</el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                </div>
              </template>
            </el-table-column>
          </el-table>
          <div class="admin-pagination">
            <el-pagination
              v-model:current-page="page"
              :page-size="pageSize"
              :page-sizes="[10,20,50,100]"
              :total="total"
              layout="total,sizes,prev,pager,next"
              @current-change="load"
              @size-change="(val:number) => { pageSize = val; page = 1; load() }"
            />
          </div>
    </el-card>

    <el-dialog v-model="dialog.visible" :title="dialog.title" width="min(920px, 92vw)" :close-on-click-modal="false" class="permission-dialog">
      <el-form ref="formRef" :model="form" label-width="100px">
        <el-form-item label="头像" class="avatar-form-item">
          <el-upload action="/upload" :show-file-list="false" :on-success="handleAvatarSuccess" :on-error="()=>ElMessage.error('上传失败')" :headers="{ Authorization: token }" accept="image/*">
            <div class="avatar-upload">
              <el-avatar v-if="form.pic" :src="form.pic" size="large" />
              <div v-else class="avatar-placeholder">+</div>
            </div>
          </el-upload>
        </el-form-item>
        <el-form-item label="用户名" prop="name">
          <el-input v-model="form.name" placeholder="请输入唯一用户名" maxlength="100" clearable />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="form.mobile" placeholder="手机号" />
        </el-form-item>
        <el-form-item label="岗位">
          <el-select v-model="form.positionId" placeholder="请选择岗位" clearable style="width:100%">
            <el-option v-for="item in positionOptions" :key="item.id" :label="item.name" :value="item.id">
              <div class="role-option">
                <span>{{ item.name }}</span>
                <el-tag v-if="item.status === 0" size="small" type="info">已停用</el-tag>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="所属部门">
          <div style="width:100%;position:relative">
            <el-popover trigger="click" placement="bottom-start" :width="420" popper-style="margin-top:4px">
              <template #reference>
                <el-input readonly :model-value="deptDisplayText" placeholder="请选择部门" suffix-icon="ArrowDown" style="width:100%" />
              </template>
              <el-tree
                ref="deptTreeRef"
                :data="deptTreeData"
                :props="{ label: 'name' }"
                show-checkbox
                check-strictly
                node-key="id"
                :default-checked-keys="form.deptIds"
                @check="onDeptCheck"
                style="max-height:300px;overflow-y:auto"
              />
            </el-popover>
          </div>
        </el-form-item>
        <el-form-item :label="dialog.isCreate ? '登录密码' : '新密码'">
          <el-input v-model="form.password" type="password" show-password :placeholder="dialog.isCreate ? '绑定角色时必填' : '留空则不修改密码'" />
        </el-form-item>
        <el-divider content-position="left" class="user-form-divider">额外信息</el-divider>
        <el-form-item v-for="f in formFields" :key="f._key" :label="f.label">
          <el-input v-if="f.type === '文本'" v-model="form.formsData[f._key]" />
          <el-input-number v-else-if="f.type === '数字'" v-model="form.formsData[f._key]" :min="0" />
          <el-input v-else-if="f.type === '多行文本'" v-model="form.formsData[f._key]" type="textarea" />
          <el-select v-else-if="f.type === '选择'" v-model="form.formsData[f._key]" style="width:100%">
            <el-option v-for="opt in (f.options||'').split(',').filter(Boolean)" :key="opt" :label="opt" :value="opt" />
          </el-select>
          <el-date-picker
            v-else-if="f.type === '日期'"
            v-model="form.formsData[f._key]"
            type="date"
            value-format="YYYY-MM-DD"
            placeholder="选择日期"
            style="width:100%"
          />
          <el-time-picker
            v-else-if="f.type === '时间'"
            v-model="form.formsData[f._key]"
            value-format="HH:mm"
            placeholder="选择时间"
            style="width:100%"
          />
          <el-date-picker
            v-else-if="f.type === '日期时间'"
            v-model="form.formsData[f._key]"
            type="datetime"
            value-format="YYYY-MM-DD HH:mm"
            placeholder="选择日期时间"
            style="width:100%"
          />
        </el-form-item>
        <el-divider content-position="left" class="user-form-divider">角色权限</el-divider>
        <el-form-item label="绑定角色">
          <div class="access-role-field">
            <el-select v-model="form.roleIds" placeholder="请选择角色" clearable filterable multiple collapse-tags collapse-tags-tooltip style="width:100%">
              <el-option v-for="r in roleList" :key="r.id" :label="r.name" :value="r.id">
                <div class="role-option">
                  <span>{{ r.name }}</span>
                  <el-tag v-if="r.status === 0" size="small" type="info">已停用</el-tag>
                </div>
              </el-option>
            </el-select>
          </div>
        </el-form-item>
        <el-divider content-position="left" class="user-form-divider">用户额外数据权限</el-divider>
        <el-form-item label="可见部门">
          <div style="width:100%;position:relative">
            <el-popover trigger="click" placement="bottom-start" :width="420" popper-style="margin-top:4px">
              <template #reference>
                <el-input readonly :model-value="extraDataDeptDisplayText" placeholder="请选择额外可见部门" suffix-icon="ArrowDown" style="width:100%" />
              </template>
              <div class="extra-data-picker-search">
                <el-input
                  v-model="extraDataDeptKeyword"
                  placeholder="搜索部门"
                  clearable
                  size="small"
                  @input="filterExtraDataDeptTree"
                />
              </div>
              <el-tree
                ref="extraDataDeptTreeRef"
                :data="deptTreeData"
                :props="{ label: 'name' }"
                :filter-node-method="filterExtraDataDeptNode"
                show-checkbox
                check-strictly
                node-key="id"
                :default-checked-keys="form.extraDataDeptIds"
                @check="onExtraDataDeptCheck"
                style="max-height:300px;overflow-y:auto"
              />
            </el-popover>
          </div>
        </el-form-item>
        <el-form-item label="可见用户">
          <div style="width:100%;position:relative">
            <el-popover trigger="click" placement="bottom-start" :width="520" popper-class="extra-data-user-popover" popper-style="margin-top:4px">
              <template #reference>
                <el-input readonly :model-value="extraDataUserDisplayText" placeholder="请选择额外可见用户" suffix-icon="ArrowDown" style="width:100%" />
              </template>
              <div class="extra-data-picker-search">
                <el-input
                  v-model="extraDataUserKeyword"
                  placeholder="搜索用户/手机号/部门"
                  clearable
                  size="small"
                  @input="filterExtraDataUserTree"
                />
              </div>
              <div class="extra-data-user-tree-wrap">
                <el-tree
                  ref="extraDataUserTreeRef"
                  :data="extraDataUserTreeData"
                  :props="{ label: 'label', children: 'children' }"
                  :filter-node-method="filterExtraDataUserNode"
                  show-checkbox
                  check-on-click-node
                  node-key="key"
                  :default-checked-keys="extraDataUserCheckedNodeKeys"
                  @check="onExtraDataUserCheck"
                  class="extra-data-user-tree"
                >
                  <template #default="{ node, data }">
                    <span :class="['extra-data-user-node', data.type === 'user' ? 'is-user' : 'is-dept']">
                      <span class="extra-data-user-node__label">{{ node.label }}</span>
                      <span v-if="data.type === 'user' && data.mobile" class="extra-data-user-node__meta">{{ data.mobile }}</span>
                    </span>
                  </template>
                </el-tree>
              </div>
            </el-popover>
          </div>
        </el-form-item>

        <el-divider content-position="left" class="user-form-divider collapsible-divider">
          <div class="collapsible-divider__content">
            <span>用户扩展应用权限</span>
            <i class="collapsible-divider__line" aria-hidden="true"></i>
            <el-button
              link
              type="primary"
              size="small"
              :icon="userApplicationPermissionExpanded ? 'ArrowUp' : 'ArrowDown'"
              @click.stop="userApplicationPermissionExpanded = !userApplicationPermissionExpanded"
            >
              {{ userApplicationPermissionExpanded ? '收起' : '展开' }}
            </el-button>
          </div>
        </el-divider>
        <el-form-item v-show="userApplicationPermissionExpanded" label="权限设置" class="permission-form-item">
          <div class="permission-layout">
            <section class="permission-column permission-column--menu">
              <div class="permission-column__header">
                <span>菜单权限</span>
                <small>控制客户端与钉钉 H5 可见入口</small>
              </div>
              <div class="permission-panel">
                <div class="permission-panel__title">额外授权 - 客户端菜单</div>
                <el-tree
                  ref="allowClientMenuTreeRef"
                  :data="clientMenuTreeData"
                  :props="{ label: 'name', children: 'children' }"
                  show-checkbox
                  check-strictly
                  node-key="key"
                  :default-checked-keys="allowClientMenuCheckedKeys"
                  @check="onUserApplicationPermissionCheck('allow')"
                  class="permission-tree"
                />
              </div>
              <div class="permission-panel">
                <div class="permission-panel__title">额外授权 - 钉钉 H5 菜单/按钮</div>
                <el-tree
                  ref="allowDingTalkH5MenuTreeRef"
                  :data="dingtalkH5MenuTreeData"
                  :props="{ label: 'name', children: 'children' }"
                  show-checkbox
                  check-strictly
                  node-key="key"
                  :default-checked-keys="allowDingTalkH5MenuCheckedKeys"
                  @check="onUserApplicationPermissionCheck('allow')"
                  class="permission-tree"
                />
              </div>
            </section>
            <section class="permission-column permission-column--api">
              <div class="permission-column__header">
                <span>接口权限</span>
                <small>控制客户端与钉钉 H5 接口访问</small>
              </div>
              <div class="permission-panel">
                <div class="permission-panel__title">额外授权 - 客户端接口</div>
                <el-tree
                  ref="allowClientApiTreeRef"
                  :data="clientApiTreeData"
                  :props="{ label: 'name', children: 'children' }"
                  show-checkbox
                  node-key="key"
                  :default-checked-keys="allowClientApiCheckedKeys"
                  @check="onUserApplicationPermissionCheck('allow')"
                  class="permission-tree"
                />
              </div>
              <div class="permission-panel">
                <div class="permission-panel__title">额外授权 - 钉钉 H5 接口</div>
                <el-tree
                  ref="allowDingTalkH5ApiTreeRef"
                  :data="dingtalkH5ApiTreeData"
                  :props="{ label: 'name', children: 'children' }"
                  show-checkbox
                  node-key="key"
                  :default-checked-keys="allowDingTalkH5ApiCheckedKeys"
                  @check="onUserApplicationPermissionCheck('allow')"
                  class="permission-tree"
                />
              </div>
            </section>
            <div class="permission-section-divider">
              <span>禁止权限</span>
            </div>
            <section class="permission-column permission-column--menu">
              <div class="permission-panel">
                <div class="permission-panel__title">禁止权限 - 客户端菜单</div>
                <el-tree
                  ref="denyClientMenuTreeRef"
                  :data="clientMenuTreeData"
                  :props="{ label: 'name', children: 'children' }"
                  show-checkbox
                  check-strictly
                  node-key="key"
                  :default-checked-keys="denyClientMenuCheckedKeys"
                  @check="onUserApplicationPermissionCheck('deny')"
                  class="permission-tree"
                />
              </div>
              <div class="permission-panel">
                <div class="permission-panel__title">禁止权限 - 钉钉 H5 菜单/按钮</div>
                <el-tree
                  ref="denyDingTalkH5MenuTreeRef"
                  :data="dingtalkH5MenuTreeData"
                  :props="{ label: 'name', children: 'children' }"
                  show-checkbox
                  check-strictly
                  node-key="key"
                  :default-checked-keys="denyDingTalkH5MenuCheckedKeys"
                  @check="onUserApplicationPermissionCheck('deny')"
                  class="permission-tree"
                />
              </div>
            </section>
            <section class="permission-column permission-column--api">
              <div class="permission-panel">
                <div class="permission-panel__title">禁止权限 - 客户端接口</div>
                <el-tree
                  ref="denyClientApiTreeRef"
                  :data="clientApiTreeData"
                  :props="{ label: 'name', children: 'children' }"
                  show-checkbox
                  node-key="key"
                  :default-checked-keys="denyClientApiCheckedKeys"
                  @check="onUserApplicationPermissionCheck('deny')"
                  class="permission-tree"
                />
              </div>
              <div class="permission-panel">
                <div class="permission-panel__title">禁止权限 - 钉钉 H5 接口</div>
                <el-tree
                  ref="denyDingTalkH5ApiTreeRef"
                  :data="dingtalkH5ApiTreeData"
                  :props="{ label: 'name', children: 'children' }"
                  show-checkbox
                  node-key="key"
                  :default-checked-keys="denyDingTalkH5ApiCheckedKeys"
                  @check="onUserApplicationPermissionCheck('deny')"
                  class="permission-tree"
                />
              </div>
            </section>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveUser">{{ dialog.isCreate ? '创建' : '保存' }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="reasonDialog.visible" title="请输入原因" width="400px">
      <el-input v-model="reasonDialog.reason" type="textarea" placeholder="禁用/审核不通过原因" />
      <template #footer>
        <el-button @click="reasonDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="confirmReason">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailDialog.visible" title="用户详情" width="min(640px, 92vw)" class="user-detail-dialog">
      <div v-if="detailDialog.detail">
        <div style="text-align:center;margin-bottom:16px">
          <span class="user-avatar user-avatar--large" :class="{ 'user-avatar--image': avatarImageReady(detailDialog.detail) }">
            <span v-if="!avatarImageReady(detailDialog.detail)" class="user-avatar__initial">{{ avatarInitial(detailDialog.detail) }}</span>
            <img
              v-if="preferredAvatarUrl(detailDialog.detail)"
              :class="{ 'is-loaded': avatarImageReady(detailDialog.detail) }"
              :src="preferredAvatarUrl(detailDialog.detail)"
              :alt="`${detailDialog.detail.name || '用户'}头像`"
              @load="onAvatarLoad(detailDialog.detail)"
              @error="onAvatarError(detailDialog.detail)"
            />
          </span>
          <div style="margin-top:8px;font-size:18px;font-weight:600">{{ detailDialog.detail.name }}</div>
          <el-tag :type="statusType(detailDialog.detail.status)" size="small">{{ statusLabel(detailDialog.detail.status) }}</el-tag>
        </div>
        <el-descriptions :column="1" border class="user-detail-descriptions">
          <el-descriptions-item label="手机号">{{ detailDialog.detail.mobile || '-' }}</el-descriptions-item>
          <el-descriptions-item label="所属部门">
            <div v-if="detailDeptNames(detailDialog.detail).length" class="user-detail-chip-list">
              <span v-for="name in detailDeptNames(detailDialog.detail)" :key="name" class="user-detail-chip">{{ name }}</span>
            </div>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="岗位">{{ detailDialog.detail.positionName || '-' }}</el-descriptions-item>
          <el-descriptions-item label="绑定角色">
            <div v-if="displayRoleNames(detailDialog.detail).length" class="user-detail-chip-list">
              <el-tag v-for="name in displayRoleNames(detailDialog.detail)" :key="name" type="primary" size="small">{{ name }}</el-tag>
            </div>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="注册时间">{{ fmtTime(detailDialog.detail.addTime) }}</el-descriptions-item>
          <el-descriptions-item label="最后登录">{{ fmtTime(detailDialog.detail.loginTime) }}</el-descriptions-item>
          <el-descriptions-item v-for="f in formFields" :key="f._key" :label="f.label">
            {{ parsedForms(f) }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>
  </div>
</template>

<script lang="ts" setup>
import SortPopover from '../../components/SortPopover.vue'
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { adminApi } from '../../api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { hasPerm } from '../../utils/permission'

const loading = ref(false)
const saving = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const selected = ref<any[]>([])
const token = localStorage.getItem('admin_token') || ''
const sortRules = ref<{field:string;order:string}[]>([])
const userApplicationPermissionExpanded = ref(false)
const sortColumns = [
  { label: '用户名', field: 'name' },
  { label: '手机号', field: 'mobile' },
  { label: '状态', field: 'status' },
  { label: '登录次数', field: 'loginCnt' },
  { label: '注册时间', field: 'addTime' },
]

const formFields = ref<any[]>([])
const deptTreeData = ref<any[]>([])
const deptTreeRef = ref<any>(null)
const extraDataDeptTreeRef = ref<any>(null)
const extraDataUserTreeRef = ref<any>(null)
const extraDataDeptKeyword = ref('')
const extraDataUserKeyword = ref('')
const roleList = ref<any[]>([])
const positionOptions = ref<any[]>([])
const dataScopeUserOptions = ref<any[]>([])
const clientMenuTreeData = ref<any[]>([])
const dingtalkH5MenuTreeData = ref<any[]>([])
const clientApiTreeData = ref<any[]>([])
const dingtalkH5ApiTreeData = ref<any[]>([])
const allowClientMenuTreeRef = ref<any>(null)
const allowDingTalkH5MenuTreeRef = ref<any>(null)
const denyClientMenuTreeRef = ref<any>(null)
const denyDingTalkH5MenuTreeRef = ref<any>(null)
const allowClientApiTreeRef = ref<any>(null)
const allowDingTalkH5ApiTreeRef = ref<any>(null)
const denyClientApiTreeRef = ref<any>(null)
const denyDingTalkH5ApiTreeRef = ref<any>(null)
const deptDisplayText = computed(() => {
  if (!form.deptIds || form.deptIds.length === 0) return ''
  return form.deptIds.map((id: number) => deptTagName(id)).join('、')
})
const extraDataDeptDisplayText = computed(() => {
  if (!form.extraDataDeptIds || form.extraDataDeptIds.length === 0) return ''
  return form.extraDataDeptIds.map((id: number) => deptTagName(id)).join('、')
})
const extraDataUserTreeData = computed(() => buildExtraDataUserTree(deptTreeData.value, dataScopeUserOptions.value))
const extraDataUserDisplayText = computed(() => {
  if (!form.extraDataUserIds || form.extraDataUserIds.length === 0) return ''
  const names = form.extraDataUserIds.map((id: number) => dataScopeUserName(id)).filter(Boolean)
  return names.join('、')
})
const extraDataUserCheckedNodeKeys = computed(() => checkedExtraDataUserNodeKeys(form.extraDataUserIds, extraDataUserTreeData.value))
const dingtalkH5MenuButtonPrefixes = ['dingtalk_h5:menu:', 'dingtalk_h5:button:']
const allowClientMenuKeys = computed(() => applicationKeysByPrefix(form.allowPermissionKeys, 'client:menu:'))
const allowDingTalkH5MenuKeys = computed(() => applicationKeysByPrefixes(form.allowPermissionKeys, dingtalkH5MenuButtonPrefixes))
const allowClientApiKeys = computed(() => applicationKeysByPrefix(form.allowPermissionKeys, 'client:api:'))
const allowDingTalkH5ApiKeys = computed(() => applicationKeysByPrefix(form.allowPermissionKeys, 'dingtalk_h5:api:'))
const denyClientMenuKeys = computed(() => applicationKeysByPrefix(form.denyPermissionKeys, 'client:menu:'))
const denyDingTalkH5MenuKeys = computed(() => applicationKeysByPrefixes(form.denyPermissionKeys, dingtalkH5MenuButtonPrefixes))
const denyClientApiKeys = computed(() => applicationKeysByPrefix(form.denyPermissionKeys, 'client:api:'))
const denyDingTalkH5ApiKeys = computed(() => applicationKeysByPrefix(form.denyPermissionKeys, 'dingtalk_h5:api:'))
const allowClientMenuCheckedKeys = computed(() => checkableKeysForTree(allowClientMenuKeys.value, clientMenuTreeData.value))
const allowDingTalkH5MenuCheckedKeys = computed(() => allowDingTalkH5MenuKeys.value)
const allowClientApiCheckedKeys = computed(() => checkableKeysForTree(allowClientApiKeys.value, clientApiTreeData.value))
const allowDingTalkH5ApiCheckedKeys = computed(() => checkableKeysForTree(allowDingTalkH5ApiKeys.value, dingtalkH5ApiTreeData.value))
const denyClientMenuCheckedKeys = computed(() => checkableKeysForTree(denyClientMenuKeys.value, clientMenuTreeData.value))
const denyDingTalkH5MenuCheckedKeys = computed(() => denyDingTalkH5MenuKeys.value)
const denyClientApiCheckedKeys = computed(() => checkableKeysForTree(denyClientApiKeys.value, clientApiTreeData.value))
const denyDingTalkH5ApiCheckedKeys = computed(() => checkableKeysForTree(denyDingTalkH5ApiKeys.value, dingtalkH5ApiTreeData.value))

function fmtTime(ts: number) {
  if (!ts) return '-'
  return new Date(ts).toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}
function statusType(s: any) {
  const m: Record<string, string> = { '0': 'warning', '1': 'success', '2': 'danger', '9': 'info' }
  return m[String(s)] || 'info'
}
function statusLabel(s: any) {
  const m: Record<string, string> = { '0': '待审核', '1': '正常', '2': '禁用', '9': '管理员' }
  return m[String(s)] || String(s)
}

function deptNameMap() {
  const m: Record<number, string> = {}
  function walk(nodes: any[]) {
    for (const n of nodes) {
      m[n.id] = n.name
      if (n.children) walk(n.children)
    }
  }
  walk(deptTreeData.value)
  return m
}

type DeptPathInfo = {
  path: string[]
}

type DeptPathItem = {
  key: string
  text: string
  segments: string[]
}

function deptPathInfoMap() {
  const m: Record<number, DeptPathInfo> = {}
  function walk(nodes: any[], parentPath: string[] = []) {
    for (const n of nodes || []) {
      const id = Number(n.id)
      if (!id) continue
      const name = String(n.name || id)
      const path = [...parentPath, name]
      m[id] = { path }
      if (Array.isArray(n.children) && n.children.length > 0) {
        walk(n.children, path)
      }
    }
  }
  walk(deptTreeData.value)
  return m
}

function deptPathItems(deptIds: number[]): DeptPathItem[] {
  if (!deptIds || deptIds.length === 0) return []
  const pathMap = deptPathInfoMap()
  const ids = Array.from(new Set((deptIds || []).map((id) => Number(id)).filter(Boolean)))
  return ids.map((id) => {
    const path = pathMap[id]?.path
    const segments = path?.length ? path : [String(id)]
    return {
      key: String(id),
      text: segments.join('|'),
      segments
    }
  }).filter((item) => item.segments.length > 0)
}

function deptPathText(deptIds: number[]) {
  const names = deptPathItems(deptIds).map((item) => item.text)
  return names.filter(Boolean).join('、') || '-'
}

function deptTagName(id: number) {
  return deptNameMap()[id] || id
}

function onDeptCheck() {
  nextTick(() => {
    form.deptIds = deptTreeRef.value?.getCheckedKeys() || []
  })
}

function onExtraDataDeptCheck() {
  nextTick(() => {
    form.extraDataDeptIds = extraDataDeptTreeRef.value?.getCheckedKeys() || []
  })
}

function normalizeFilterText(value: any) {
  return String(value || '').trim().toLowerCase()
}

function filterExtraDataDeptTree() {
  extraDataDeptTreeRef.value?.filter(normalizeFilterText(extraDataDeptKeyword.value))
}

function filterExtraDataUserTree() {
  extraDataUserTreeRef.value?.filter(normalizeFilterText(extraDataUserKeyword.value))
}

function filterExtraDataDeptNode(keyword: string, data: any) {
  const value = normalizeFilterText(keyword)
  if (!value) return true
  const name = normalizeFilterText(data?.name)
  const path = data?.id ? normalizeFilterText(deptPathText([Number(data.id)])) : ''
  return name.includes(value) || path.includes(value)
}

function filterExtraDataUserNode(keyword: string, data: any) {
  const value = normalizeFilterText(keyword)
  if (!value) return true
  return [
    data?.label,
    data?.mobile,
    data?.pathText
  ].some((item) => normalizeFilterText(item).includes(value))
}

function resetExtraDataPickerFilters() {
  extraDataDeptKeyword.value = ''
  extraDataUserKeyword.value = ''
  nextTick(() => {
    filterExtraDataDeptTree()
    filterExtraDataUserTree()
  })
}

function normalizeNumberIDs(ids: any[]) {
  return Array.from(new Set((ids || []).map((id: any) => Number(id)).filter(Boolean)))
}

function dataScopeUserName(id: number) {
  const item = dataScopeUserOptions.value.find((user: any) => Number(user.id) === Number(id))
  return item?.name || String(id)
}

function extraDataUserNodeKey(userID: number, scopeKey: string) {
  return `user:${scopeKey}:${userID}`
}

function extraDataUserIDFromNodeKey(key: string) {
  if (!key.startsWith('user:')) return 0
  const parts = key.split(':')
  return Number(parts[parts.length - 1]) || 0
}

function extraDataUserNode(user: any, scopeKey: string) {
  const userID = Number(user.id)
  return {
    key: extraDataUserNodeKey(userID, scopeKey),
    type: 'user',
    id: userID,
    label: user.name || String(userID),
    mobile: user.mobile || user.phone || '',
    pathText: deptPathText(user.deptIds || [])
  }
}

function buildExtraDataUserTree(depts: any[], users: any[]) {
  const usersByDept: Record<number, any[]> = {}
  const usersWithoutDept: any[] = []
  for (const user of users || []) {
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
    const userNodes = (usersByDept[deptID] || []).map((user: any) => extraDataUserNode(user, `dept-${deptID}`))
    const children = [...childDeptNodes, ...userNodes]
    if (children.length === 0) return null
    return {
      key: `dept-${deptID}`,
      type: 'dept',
      id: deptID,
      label: dept.name || String(deptID),
      children
    }
  }

  const tree = (depts || []).map((dept: any) => buildDeptNode(dept)).filter(Boolean)
  if (usersWithoutDept.length > 0) {
    tree.push({
      key: 'dept-unassigned',
      type: 'dept',
      label: '未分配部门',
      children: usersWithoutDept.map((user: any) => extraDataUserNode(user, 'unassigned'))
    })
  }
  const visibleUserIDs = new Set<number>()
  function collectUserIDs(items: any[]) {
    for (const item of items || []) {
      if (item.type === 'user') visibleUserIDs.add(Number(item.id))
      if (Array.isArray(item.children) && item.children.length > 0) collectUserIDs(item.children)
    }
  }
  collectUserIDs(tree)
  const unmatchedUsers = (users || []).filter((user: any) => !visibleUserIDs.has(Number(user.id)))
  if (unmatchedUsers.length > 0) {
    tree.push({
      key: 'dept-unmatched',
      type: 'dept',
      label: '未匹配部门',
      children: unmatchedUsers.map((user: any) => extraDataUserNode(user, 'unmatched'))
    })
  }
  return tree
}

function checkedExtraDataUserNodeKeys(userIDs: number[], nodes: any[]) {
  const selected = new Set(normalizeNumberIDs(userIDs || []))
  const keys: string[] = []
  function walk(items: any[]) {
    for (const item of items || []) {
      if (item.type === 'user' && selected.has(Number(item.id))) {
        keys.push(item.key)
      }
      if (Array.isArray(item.children) && item.children.length > 0) {
        walk(item.children)
      }
    }
  }
  walk(nodes)
  return keys
}

function extraDataUserIDsFromCheckedKeys(keys: string[]) {
  const ids = (keys || []).map((key) => extraDataUserIDFromNodeKey(String(key))).filter(Boolean)
  return normalizeNumberIDs(ids)
}

function setExtraDataUserTreeKeys() {
  extraDataUserTreeRef.value?.setCheckedKeys(extraDataUserCheckedNodeKeys.value)
}

function onExtraDataUserCheck() {
  nextTick(() => {
    const checkedKeys = extraDataUserTreeRef.value?.getCheckedKeys?.() || []
    form.extraDataUserIds = extraDataUserIDsFromCheckedKeys(checkedKeys)
    nextTick(() => setExtraDataUserTreeKeys())
  })
}

function applicationKeysByPrefix(keys: string[], prefix: string) {
  return (keys || []).filter((key: string) => key.startsWith(prefix))
}

function applicationKeysByPrefixes(keys: string[], prefixes: string[]) {
  return (keys || []).filter((key: string) => prefixes.some((prefix) => key.startsWith(prefix)))
}

function checkedKeys(treeRef: any, options: { includeHalfChecked?: boolean; prefix?: string; prefixes?: string[] } = {}) {
  const checked = treeRef.value?.getCheckedKeys?.() || []
  const halfChecked = options.includeHalfChecked ? (treeRef.value?.getHalfCheckedKeys?.() || []) : []
  const keys = Array.from(new Set([...checked, ...halfChecked])) as string[]
  const prefixes = options.prefixes || (options.prefix ? [options.prefix] : [])
  if (prefixes.length === 0) return keys
  return keys.filter((key) => prefixes.some((prefix) => key.startsWith(prefix)))
}

function onUserApplicationPermissionCheck(source: 'allow' | 'deny') {
  nextTick(() => {
    const allowKeys = Array.from(new Set([
      ...checkedKeys(allowClientMenuTreeRef, { prefix: 'client:menu:' }),
      ...checkedKeys(allowDingTalkH5MenuTreeRef, { prefixes: dingtalkH5MenuButtonPrefixes }),
      ...checkedKeys(allowClientApiTreeRef, { prefix: 'client:api:' }),
      ...checkedKeys(allowDingTalkH5ApiTreeRef, { prefix: 'dingtalk_h5:api:' })
    ])) as string[]
    const denyKeys = Array.from(new Set([
      ...checkedKeys(denyClientMenuTreeRef, { prefix: 'client:menu:' }),
      ...checkedKeys(denyDingTalkH5MenuTreeRef, { prefixes: dingtalkH5MenuButtonPrefixes }),
      ...checkedKeys(denyClientApiTreeRef, { prefix: 'client:api:' }),
      ...checkedKeys(denyDingTalkH5ApiTreeRef, { prefix: 'dingtalk_h5:api:' })
    ])) as string[]
    const allowSet = new Set(allowKeys)
    const denySet = new Set(denyKeys)
    if (source === 'deny') {
      form.allowPermissionKeys = allowKeys.filter((key) => !denySet.has(key))
      form.denyPermissionKeys = denyKeys
    } else {
      form.allowPermissionKeys = allowKeys
      form.denyPermissionKeys = denyKeys.filter((key) => !allowSet.has(key))
    }
    nextTick(() => setUserApplicationPermissionTreeKeys())
  })
}

function setUserApplicationPermissionTreeKeys() {
  allowClientMenuTreeRef.value?.setCheckedKeys(allowClientMenuCheckedKeys.value)
  allowDingTalkH5MenuTreeRef.value?.setCheckedKeys(allowDingTalkH5MenuCheckedKeys.value)
  allowClientApiTreeRef.value?.setCheckedKeys(allowClientApiCheckedKeys.value)
  allowDingTalkH5ApiTreeRef.value?.setCheckedKeys(allowDingTalkH5ApiCheckedKeys.value)
  denyClientMenuTreeRef.value?.setCheckedKeys(denyClientMenuCheckedKeys.value)
  denyDingTalkH5MenuTreeRef.value?.setCheckedKeys(denyDingTalkH5MenuCheckedKeys.value)
  denyClientApiTreeRef.value?.setCheckedKeys(denyClientApiCheckedKeys.value)
  denyDingTalkH5ApiTreeRef.value?.setCheckedKeys(denyDingTalkH5ApiCheckedKeys.value)
}

function checkableKeysForTree(keys: string[], nodes: any[]) {
  const parentKeys = new Set<string>()
  function walk(items: any[]) {
    for (const item of items || []) {
      if (Array.isArray(item.children) && item.children.length > 0) {
        parentKeys.add(String(item.key))
        walk(item.children)
      }
    }
  }
  walk(nodes)
  return keys.filter((key) => !parentKeys.has(key))
}

function normalizeAvatarUrl(value: any) {
  const url = String(value ?? '').trim()
  if (!url || url === '-') return ''
  const lower = url.toLowerCase()
  if (lower === 'null' || lower === 'undefined') return ''
  const path = lower.split('?')[0]
  if (path === 'static/default-avatar.png' || path.endsWith('/static/default-avatar.png')) return ''
  return url
}

function preferredAvatarUrl(row: any) {
  if (!row || row.avatarLoadFailed) return ''
  return normalizeAvatarUrl(row.avatar) || normalizeAvatarUrl(row.pic)
}

function avatarImageReady(row: any) {
  return !!preferredAvatarUrl(row) && row?.avatarLoaded === true
}

function avatarInitial(row: any) {
  const text = String(row?.name || row?.realName || row?.mobile || row?.id || '用').trim()
  return Array.from(text)[0] || '用'
}

function onAvatarLoad(row: any) {
  if (!row) return
  row.avatarLoaded = true
}

function onAvatarError(row: any) {
  if (!row) return
  row.avatarLoadFailed = true
  row.avatarLoaded = false
  row.avatar = ''
  row.pic = ''
}

async function loadDepts() {
  try {
    const res = await adminApi.deptTree()
    deptTreeData.value = Array.isArray(res.data) ? res.data : []
  } catch { deptTreeData.value = [] }
}

async function loadRoles() {
  try {
    const res = await adminApi.roleList({ page: 1, pageSize: 9999 })
    roleList.value = Array.isArray(res.data?.list) ? res.data.list : (Array.isArray(res.data) ? res.data : [])
  } catch { roleList.value = [] }
}

async function loadPositions() {
  try {
    const res = await adminApi.positionList({ page: 1, pageSize: 9999 })
    positionOptions.value = Array.isArray(res.data?.list) ? res.data.list : []
  } catch { positionOptions.value = [] }
}

async function loadDataScopeUserOptions() {
  try {
    const res = await adminApi.userList({ page: 1, pageSize: 9999 })
    dataScopeUserOptions.value = Array.isArray(res.data?.list) ? res.data.list : []
    await nextTick()
    if (dialog.visible) setExtraDataUserTreeKeys()
  } catch { dataScopeUserOptions.value = [] }
}

async function loadApplicationPermissionTree() {
  try {
    const res = await adminApi.appPermissionTree()
    clientMenuTreeData.value = Array.isArray(res.data?.client) ? res.data.client : []
    dingtalkH5MenuTreeData.value = Array.isArray(res.data?.dingtalkH5) ? res.data.dingtalkH5 : []
    clientApiTreeData.value = Array.isArray(res.data?.clientApi) ? res.data.clientApi : []
    dingtalkH5ApiTreeData.value = Array.isArray(res.data?.dingtalkH5Api) ? res.data.dingtalkH5Api : []
    if (clientMenuTreeData.value.length === 0 && dingtalkH5MenuTreeData.value.length === 0 && clientApiTreeData.value.length === 0 && dingtalkH5ApiTreeData.value.length === 0) {
      ElMessage.warning('应用权限配置为空，请确认后端已启动最新版本')
    }
    await nextTick()
    if (dialog.visible) setUserApplicationPermissionTreeKeys()
  } catch {
    clientMenuTreeData.value = []
    dingtalkH5MenuTreeData.value = []
    clientApiTreeData.value = []
    dingtalkH5ApiTreeData.value = []
    ElMessage.error('应用权限加载失败')
  }
}

async function load() {
  loading.value = true
  try {
    const params: any = { page: page.value, pageSize: pageSize.value, keyword: keyword.value }
    if (sortRules.value.length) params.sort = sortRules.value.map(s => s.field + ':' + s.order).join(',')
    const res = await adminApi.userList(params)
    const rawList = res.data?.list || []
    total.value = res.data?.total || 0
    list.value = rawList.map((u: any) => ({
      ...u,
      avatar: normalizeAvatarUrl(u.avatar || u.pic),
      pic: normalizeAvatarUrl(u.pic || u.avatar),
      avatarLoadFailed: false,
      avatarLoaded: false,
      deptNames: deptPathText(u.deptIds || []),
      deptPathItems: deptPathItems(u.deptIds || [])
    }))
  } catch { list.value = []; total.value = 0 }
  loading.value = false
}
function onSortChange() { page.value = 1; load() }
function search() { page.value = 1; load() }

// 状态变更
const reasonDialog = reactive({ visible: false, reason: '', row: null as any, targetStatus: '' })
function showReason(row: any) {
  reasonDialog.row = row
  reasonDialog.targetStatus = '2'
  reasonDialog.reason = ''
  reasonDialog.visible = true
}
function showAuditFail(row: any) {
  reasonDialog.row = row
  reasonDialog.targetStatus = '2'
  reasonDialog.reason = ''
  reasonDialog.visible = true
}
async function confirmReason() {
  if (!reasonDialog.reason) { ElMessage.warning('请输入原因'); return }
  await changeStatus(reasonDialog.row, reasonDialog.targetStatus, reasonDialog.reason)
  reasonDialog.visible = false
}
async function changeStatus(row: any, status: string, reason?: string) {
  await adminApi.userStatus({ id: row.id, status, reason: reason || '' })
  ElMessage.success('操作成功')
  load()
}

function hasRowMoreActions(_row: any) {
  return hasPerm('admin:menu:user:edit') || hasPerm('admin:menu:user:del')
}

async function handleRowCommand(cmd: string, row: any) {
  if (cmd === 'approve' || cmd === 'enable') {
    await changeStatus(row, '1')
  } else if (cmd === 'disable') {
    showReason(row)
  } else if (cmd === 'resetPwd') {
    await resetPwd(row)
  } else if (cmd === 'delete') {
    try {
      await ElMessageBox.confirm(`确定删除用户「${row.name}」？`, '提示', { type: 'warning' })
      await remove(row)
    } catch {}
  }
}

async function remove(row: any) {
  await adminApi.userDel({ id: row.id })
  ElMessage.success('已删除')
  load()
}

async function delSelected() {
  if (selected.value.length === 0) return
  try {
    await ElMessageBox.confirm(`确定删除选中的 ${selected.value.length} 个用户？`, '提示')
    const ids = selected.value.map((r: any) => r.id).join(',')
    await adminApi.userDels({ ids })
    ElMessage.success('已删除')
    selected.value = []
    load()
  } catch {}
}

function exportData() {
  const rows = [['用户名', '姓名', '手机号', '部门', '状态']]
  list.value.forEach((r: any) => {
    rows.push([r.name || '', r.realName || '', r.phone || '', r.deptName || '', r.status === 1 ? '正常' : '禁用'])
  })
  const csv = '\uFEFF' + rows.map(r => r.map(v => '"' + String(v).replace(/"/g, '""') + '"').join(',')).join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = '用户列表.csv'
  a.click()
  URL.revokeObjectURL(a.href)
}

async function resetPwd(row: any) {
  try {
    await ElMessageBox.confirm(`确定将用户「${row.name}」的密码重置为 123456？`, '提示')
    await adminApi.userResetPwd({ id: row.id })
    ElMessage.success('密码已重置为 123456')
  } catch {}
}

// 新增/编辑
const dialog = reactive({ visible: false, title: '', isCreate: false })
const form = reactive({
  id: null as any,
  name: '',
  mobile: '',
  avatar: '',
  pic: '',
  deptIds: [] as number[],
  password: '',
  positionId: null as any,
  roleId: null as any,
  roleIds: [] as number[],
  allowPermissionKeys: [] as string[],
  denyPermissionKeys: [] as string[],
  extraDataDeptIds: [] as number[],
  extraDataUserIds: [] as number[],
  formsData: {} as any
})
function resetForm() {
  form.id = null
  form.name = ''
  form.mobile = ''
  form.avatar = ''
  form.pic = ''
  form.deptIds = []
  form.password = ''
  form.positionId = null
  form.roleId = null
  form.roleIds = []
  form.allowPermissionKeys = []
  form.denyPermissionKeys = []
  form.extraDataDeptIds = []
  form.extraDataUserIds = []
  form.formsData = {}
  resetExtraDataPickerFilters()
}

function showAdd() {
  resetForm()
  dialog.isCreate = true
  dialog.title = '增加用户'
  dialog.visible = true
  nextTick(() => {
    deptTreeRef.value?.setCheckedKeys([])
    extraDataDeptTreeRef.value?.setCheckedKeys([])
    extraDataUserTreeRef.value?.setCheckedKeys([])
    setUserApplicationPermissionTreeKeys()
  })
}
async function showEdit(row: any) {
  resetForm()
  dialog.isCreate = false
  dialog.title = '编辑用户'
  dialog.visible = true
  form.id = row.id
  form.name = row.name
  form.mobile = row.mobile
  form.positionId = row.positionId || null
  form.pic = preferredAvatarUrl(row)
  form.avatar = normalizeAvatarUrl(row.avatar)
  form.deptIds = row.deptIds || []
  form.roleId = row.roleId || null
  form.roleIds = row.roleIds?.length ? row.roleIds : (row.roleId ? [row.roleId] : [])
  try {
    const res = await adminApi.userDetailById(row.id)
    const d = res.data || {}
    form.deptIds = d.deptIds || []
    form.positionId = d.positionId || null
    form.roleId = d.roleId || null
    form.roleIds = d.roleIds?.length ? d.roleIds : (d.roleId ? [d.roleId] : [])
    form.allowPermissionKeys = d.allowPermissionKeys || []
    form.denyPermissionKeys = d.denyPermissionKeys || []
    form.extraDataDeptIds = d.extraDataDeptIds || []
    form.extraDataUserIds = d.extraDataUserIds || []
    if (d.forms) {
      try { form.formsData = JSON.parse(d.forms) } catch { form.formsData = {} }
    }
  } catch {}
  nextTick(() => {
    deptTreeRef.value?.setCheckedKeys(form.deptIds)
    extraDataDeptTreeRef.value?.setCheckedKeys(form.extraDataDeptIds)
    setExtraDataUserTreeKeys()
    setUserApplicationPermissionTreeKeys()
  })
}

function handleAvatarSuccess(res: any) {
  if (res.data?.url) form.pic = res.data.url
}

async function saveUser() {
  const normalizedName = form.name.trim()
  if (!normalizedName) { ElMessage.warning('请输入用户名'); return }
  form.name = normalizedName
  const roleIds = normalizeRoleIds(form.roleIds)
  const primaryRoleId = roleIds[0] || 0
  if (roleIds.length > 0 && dialog.isCreate && !form.password) {
    ElMessage.warning('请输入登录密码')
    return
  }
  saving.value = true
  try {
    const payload: any = {
      name: normalizedName,
      mobile: form.mobile,
	      pic: form.pic,
	      positionId: form.positionId || 0,
	      password: form.password,
	      roleId: primaryRoleId,
	      roleIds: roleIds.join(',')
	    }
    payload.allowPermissionKeys = form.allowPermissionKeys.join(',')
    payload.denyPermissionKeys = form.denyPermissionKeys.join(',')
    payload.extraDataDeptIds = form.extraDataDeptIds.join(',')
    payload.extraDataUserIds = form.extraDataUserIds.join(',')
	    if (form.deptIds.length > 0) payload.deptIds = form.deptIds.join(',')
    const hasAny = Object.values(form.formsData).some(v => v !== undefined && v !== null && v !== '')
    if (hasAny) payload.forms = JSON.stringify(form.formsData)
    if (dialog.isCreate) {
      await adminApi.userAdd(payload)
    } else {
      payload.id = form.id
      await adminApi.userEdit(payload)
    }
    ElMessage.success(dialog.isCreate ? '创建成功' : '保存成功')
    dialog.visible = false
    load()
  } finally { saving.value = false }
}

// 详情
const detailDialog: { visible: boolean; detail: any } = reactive({ visible: false, detail: null })
async function showDetail(row: any) {
  try {
    const res = await adminApi.userDetailById(row.id)
    const detail = res.data || {}
    const avatar = normalizeAvatarUrl(detail.avatar || detail.pic)
    detailDialog.detail = { ...detail, avatar, pic: avatar, avatarLoadFailed: false, avatarLoaded: false }
    detailDialog.visible = true
  } catch {}
}
function deptNameStr(deptIds: number[]) {
  if (!deptIds || deptIds.length === 0) return '-'
  const dmap = deptNameMap()
  return deptIds.map(id => dmap[id] || id).join('、')
}
function normalizedDeptPathItems(row: any) {
  const items = Array.isArray(row?.deptPathItems) ? row.deptPathItems : []
  return items.map((item: any, index: number) => {
    const text = String(item?.text || '').trim()
    const segments = Array.isArray(item?.segments) && item.segments.length > 0
      ? item.segments.map((segment: any) => String(segment).trim()).filter(Boolean)
      : text.split('/').map((segment: string) => segment.trim()).filter(Boolean)
    return {
      key: item?.key || `${text}-${index}`,
      text: text || segments.join('/'),
      segments
    }
  }).filter((item: any) => item.segments.length > 0)
}
function displayDeptPathItems(row: any) {
  return normalizedDeptPathItems(row).slice(0, 1)
}
function hiddenDeptPathCount(row: any) {
  return Math.max(normalizedDeptPathItems(row).length - 1, 0)
}
function deptPathTooltip(row: any) {
  return normalizedDeptPathItems(row).map((item: any) => item.text || item.segments.join('/')).join('、')
}
function detailDeptNames(row: any) {
  const pathItems = normalizedDeptPathItems(row)
  if (pathItems.length > 0) {
    return pathItems.map((item: any) => item.text || item.segments.join('/')).filter(Boolean)
  }
  if (!row?.deptIds || row.deptIds.length === 0) return []
  const dmap = deptNameMap()
  return row.deptIds.map((id: number) => String(dmap[id] || id)).filter(Boolean)
}
function normalizeRoleIds(values: any[]) {
  const result: number[] = []
  const seen = new Set<number>()
  ;(values || []).forEach(value => {
    const id = Number(value)
    if (!id || seen.has(id)) return
    seen.add(id)
    result.push(id)
  })
  return result
}
function displayRoleNames(row: any) {
  const names = Array.isArray(row.roleNames) ? row.roleNames.filter(Boolean) : []
  if (names.length > 0) return names
  if (row.roleName) return [row.roleName]
  if (row.roleId) return ['已绑定']
  return []
}
function hiddenRoleCount(row: any) {
  return Math.max(displayRoleNames(row).length - 1, 0)
}
function roleNamesTooltip(row: any) {
  return displayRoleNames(row).join('、')
}
function parsedForms(f: any) {
  const d: any = detailDialog.detail
  if (!d || !d.forms) return '-'
  try {
    const forms = JSON.parse(d.forms)
    return forms[String(f.id)] || '-'
  } catch { return '-' }
}

async function loadFormFields() {
  try {
    const res = await adminApi.userFormFields()
    formFields.value = (res.data || []).map((f: any, i: number) => ({ ...f, _key: String(f.id || i) }))
  } catch { formFields.value = [] }
}

onMounted(async () => {
  await Promise.all([loadFormFields(), loadDepts(), loadRoles(), loadPositions(), loadDataScopeUserOptions(), loadApplicationPermissionTree()])
  load()
})
</script>

<style scoped>
.user-avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  vertical-align: middle;
  background: #eef2f7;
  color: #667085;
  font-size: 12px;
  font-weight: 600;
  line-height: 1;
}
.user-avatar__initial {
  position: relative;
  z-index: 1;
}
.user-avatar--large {
  width: 56px;
  height: 56px;
  font-size: 20px;
}
.user-avatar--image {
  background: #f8fafc;
}
.user-avatar img {
  width: 100%;
  height: 100%;
  position: absolute;
  inset: 0;
  display: block;
  object-fit: cover;
  opacity: 0;
  transition: opacity 0.16s ease;
}
.user-avatar img.is-loaded {
  opacity: 1;
}
.avatar-upload {
  cursor: pointer;
  display: inline-block;
}
.avatar-placeholder {
  width: 56px;
  height: 56px;
  border: 1px dashed #d9d9d9;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: #999;
  transition: border-color 0.3s;
}
.avatar-placeholder:hover {
  border-color: #409eff;
  color: #409eff;
}
.avatar-form-item :deep(.el-form-item__content) {
  justify-content: flex-start !important;
  display: flex !important;
}
.avatar-form-item {
  display: flex !important;
  align-items: center !important;
  flex-wrap: nowrap !important;
}
.user-detail-dialog :deep(.el-dialog__body) {
  padding-top: 8px;
}
.user-detail-descriptions :deep(.el-descriptions__label) {
  box-sizing: border-box;
  width: 112px;
  min-width: 112px;
  white-space: nowrap;
  color: #475467;
  font-weight: 600;
}
.user-detail-descriptions :deep(.el-descriptions__content) {
  min-width: 0;
  word-break: break-word;
  line-height: 1.7;
}
.user-detail-chip-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
  min-width: 0;
}
.user-detail-chip {
  box-sizing: border-box;
  max-width: 100%;
  min-height: 24px;
  padding: 2px 8px;
  border: 1px solid #dbeafe;
  border-radius: 6px;
  display: inline-flex;
  align-items: center;
  background: #f8fbff;
  color: #344054;
  font-size: 12px;
  line-height: 18px;
  word-break: break-word;
}
.access-role-field {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
}
.role-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.role-tag-list {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 6px;
  max-width: 100%;
}
.role-tag-list--compact {
  flex-wrap: nowrap;
  align-items: center;
  min-width: 0;
}
.role-tag-list--compact :deep(.el-tag) {
  max-width: 210px;
}
.role-tag-list--compact :deep(.el-tag__content) {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
:deep(.user-name-column .cell) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.dept-path-list {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  max-width: 100%;
}
.dept-path-list--compact {
  flex-wrap: nowrap;
  min-width: 0;
}
.dept-path-list--compact .dept-path-pill {
  flex: 0 1 auto;
  min-width: 0;
}
.dept-path-pill {
  box-sizing: border-box;
  max-width: 100%;
  min-height: 24px;
  padding: 2px 8px;
  border: 1px solid #dbeafe;
  border-radius: 6px;
  display: inline-flex;
  align-items: center;
  background: #f8fbff;
  color: #344054;
  font-size: 12px;
  line-height: 18px;
  white-space: nowrap;
}
.dept-path-segment {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}
.dept-path-segment.parent {
  flex: 0 1 auto;
  color: #667085;
}
.dept-path-segment.leaf {
  flex: 1 1 auto;
  color: #1d2129;
  font-weight: 700;
}
.dept-path-separator {
  flex: 0 0 auto;
  padding: 0 5px;
  color: #98a2b3;
}
.compact-more-badge {
  box-sizing: border-box;
  flex: 0 0 auto;
  min-width: 28px;
  height: 24px;
  padding: 0 8px;
  border: 1px solid #dbeafe;
  border-radius: 6px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #f8fbff;
  color: #344054;
  font-size: 12px;
  font-weight: 600;
  line-height: 18px;
}
.extra-data-picker-search {
  box-sizing: border-box;
  padding: 6px 4px 10px;
}
.extra-data-user-tree-wrap {
  box-sizing: border-box;
  width: 100%;
  max-height: 340px;
  overflow-y: auto;
  padding: 6px 4px;
}
.extra-data-user-tree {
  min-width: 420px;
}
.extra-data-user-tree :deep(.el-tree-node__content) {
  min-width: 0;
  height: 32px;
  border-radius: 6px;
}
.extra-data-user-tree :deep(.el-tree-node__content:hover) {
  background: #f5f8ff;
}
.extra-data-user-tree :deep(.el-tree-node__label) {
  min-width: 0;
  flex: 1 1 auto;
}
.extra-data-user-node {
  box-sizing: border-box;
  width: 100%;
  min-width: 0;
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding-right: 8px;
}
.extra-data-user-node.is-dept {
  color: #1d2129;
  font-weight: 600;
}
.extra-data-user-node.is-user {
  color: #344054;
  font-weight: 400;
}
.extra-data-user-node__label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.extra-data-user-node__meta {
  flex: 0 0 auto;
  color: #8a94a6;
  font-size: 12px;
}
.user-form-divider :deep(.el-divider__text) {
  color: var(--el-color-primary);
  font-weight: 700;
}
.collapsible-divider :deep(.el-divider__text) {
  min-width: min(420px, calc(100vw - 96px));
}
.collapsible-divider__content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  width: 100%;
}
.collapsible-divider__content span {
  white-space: nowrap;
}
.collapsible-divider__line {
  flex: 1 1 auto;
  height: 1px;
  background: #dcdfe6;
}
.permission-form-item :deep(.el-form-item__content) {
  min-width: 0;
}
.permission-layout {
  box-sizing: border-box;
  width: 100%;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  align-items: start;
  gap: 16px;
}
.permission-column {
  box-sizing: border-box;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.permission-column__header {
  box-sizing: border-box;
  min-height: 34px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 0 2px;
}
.permission-column__header span {
  color: #1d2129;
  font-size: 14px;
  font-weight: 700;
}
.permission-column__header small {
  min-width: 0;
  overflow: hidden;
  color: #8a94a6;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.permission-column--menu .permission-column__header span {
  color: #2563eb;
}
.permission-column--api .permission-column__header span {
  color: #0f766e;
}
.permission-panel {
  box-sizing: border-box;
  min-width: 0;
  height: 270px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid #e5e8ef;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 6px 16px rgba(29, 41, 57, 0.04);
}
.permission-section-divider {
  box-sizing: border-box;
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  gap: 12px;
  color: #606266;
  font-size: 13px;
  font-weight: 600;
  line-height: 1;
}
.permission-section-divider::before,
.permission-section-divider::after {
  flex: 1 1 auto;
  height: 1px;
  background: #e5e8ef;
  content: '';
}
.permission-section-divider span {
  flex: 0 0 auto;
}
.permission-panel__title {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  min-height: 38px;
  padding: 0 12px;
  border-bottom: 1px solid #eef1f6;
  background: #f7f9fc;
  color: #1d2129;
  font-size: 13px;
  font-weight: 600;
}
.permission-panel__title::before {
  width: 4px;
  height: 14px;
  margin-right: 8px;
  border-radius: 999px;
  background: #409eff;
  content: '';
}
.permission-column--api .permission-panel__title::before {
  background: #14b8a6;
}
.permission-tree {
  box-sizing: border-box;
  width: 100%;
  flex: 1 1 auto;
  min-height: 0;
  padding: 8px 10px 10px;
  overflow-y: auto;
  border: 0;
  border-radius: 0;
  background: #fff;
}
.permission-tree :deep(.el-tree__empty-block) {
  min-height: 150px;
}
.permission-tree :deep(.el-tree__empty-text) {
  color: #8a94a6;
  font-size: 13px;
}
.permission-tree :deep(.el-tree-node__content) {
  min-width: 0;
  height: 30px;
  border-radius: 6px;
}
.permission-tree :deep(.el-tree-node__label) {
  overflow: hidden;
  color: #344054;
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.permission-dialog :deep(.el-dialog__body) {
  max-height: 72vh;
  overflow-y: auto;
}
@media (max-width: 900px) {
  .permission-layout {
    grid-template-columns: 1fr;
  }
  .permission-column__header {
    align-items: flex-start;
    flex-direction: column;
    gap: 2px;
  }
}
</style>
