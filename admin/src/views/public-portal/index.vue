<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { AdminPageShell, AdminPageHeader, AdminTablePanel } from '@/components/admin-ui'
import type { PublicPortalAdminItem } from '@/types/publicPortal'
import { activityCategory, copyPortalLink, errorText } from './portalModel'
import { usePublicPortalAdmin } from './usePublicPortalAdmin'
import SettingsPanel from './SettingsPanel.vue'
import ItemDialog from './ItemDialog.vue'

const { settings, sortedItems, loading, busy, loadError, actionError, dialogVisible, editing, publicUrl, canEdit, mayEditItem, mayRemoveItem, mayToggleItem, load, saveSettings, openCreate, openEdit, saveItem, changeItem, removeItem } = usePublicPortalAdmin()
const tab = ref('settings')
const copying = ref(false)
async function copyLink() {
  if (!publicUrl.value || copying.value) return
  copying.value = true
  try { await copyPortalLink(publicUrl.value); ElMessage.success('公共链接已复制') }
  catch (error) { ElMessage.error(errorText(error)) }
  finally { copying.value = false }
}
function preview() { window.open('/public/activities?preview=1', '_blank', 'noopener,noreferrer') }
function changeSort(item: PublicPortalAdminItem, value: number | undefined) {
  if (typeof value === 'number' && value !== item.sort) void changeItem(item, { sort: value })
}
</script>

<template>
  <AdminPageShell width="fluid">
    <AdminPageHeader title="公共问卷 / 考试入口" description="设置一个统一入口，供参与者浏览问卷、考试并扫码参加。">
      <template #status><el-tag v-if="settings" :type="settings.enabled ? 'success' : 'info'">{{ settings.enabled ? '入口已开启' : '入口已关闭' }}</el-tag></template>
      <template #actions>
        <el-button :disabled="loading || busy" @click="load()">刷新已保存配置</el-button>
        <el-button :disabled="!settings || loading" @click="preview">预览已保存内容</el-button>
        <el-button type="primary" plain :disabled="!publicUrl || loading" :loading="copying" @click="copyLink">复制公共链接</el-button>
      </template>
    </AdminPageHeader>
    <el-alert v-if="!canEdit()" title="当前为只读模式，可查看配置和预览；修改需要公共入口编辑权限。" type="info" show-icon :closable="false" />
    <el-alert v-if="loadError || actionError" :title="loadError || actionError" type="error" show-icon :closable="false" />
    <div v-if="settings" class="saved-link">
      <span>当前站点公共链接（自动生成）</span>
      <el-link v-if="publicUrl" :href="publicUrl" target="_blank" rel="noopener noreferrer" type="primary">{{ publicUrl }}</el-link>
      <span v-else>当前站点地址无效，请通过 HTTP(S) 访问</span>
      <span v-if="!settings.enabled" class="muted">入口关闭时公众无法浏览</span>
    </div>
    <AdminTablePanel v-if="!settings" :loading="loading" :error="loadError" :empty="!loading" empty-text="尚未读取公共入口配置" />
    <el-tabs v-else v-model="tab">
      <el-tab-pane label="页面设置" name="settings">
        <AdminTablePanel title="页面设置" :loading="loading" :min-height="0">
          <SettingsPanel :settings="settings" :editable="canEdit() && !loadError" :busy="busy || loading" @save="saveSettings" />
        </AdminTablePanel>
      </el-tab-pane>
      <el-tab-pane label="展示内容" name="items">
        <AdminTablePanel title="展示内容" :count="sortedItems.length" :loading="loading" :empty="!sortedItems.length" empty-text="还没有展示内容，添加已发布且公开的问卷或考试。">
          <template #actions><el-button type="primary" :disabled="!canEdit() || busy || loading || !!loadError || sortedItems.length >= 200" @click="openCreate">添加内容</el-button></template>
          <el-table :data="sortedItems" row-key="id">
            <el-table-column label="分类 / 来源 ID" width="150" show-overflow-tooltip><template #default="{ row }">{{ activityCategory(row.kind, row.category) }} #{{ row.sourceId }}</template></el-table-column>
            <el-table-column label="管理名称" min-width="160" show-overflow-tooltip><template #default="{ row }">{{ row.managementTitle || (row.canManage ? '来源不可用' : '无来源管理权限') }}</template></el-table-column>
            <el-table-column label="展示名称" min-width="180" show-overflow-tooltip><template #default="{ row }">{{ row.title || row.displayTitle || row.managementTitle || '未命名' }}</template></el-table-column>
            <el-table-column label="展示状态" min-width="230">
              <template #default="{ row }">
                <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '已上架' : '已下架' }}</el-tag>
                <p v-if="!row.available" class="unavailable">{{ row.unavailableReason || '来源不可用，已自动隐藏' }}</p>
                <p v-if="!row.canManage" class="muted">不可管理此来源，修改操作已禁用</p>
                <p v-else-if="row.available && row.enabled" class="muted">公开显示仍受页面开关、活动时间限制</p>
              </template>
            </el-table-column>
            <el-table-column label="排序（小值在前）" width="165"><template #default="{ row }"><el-input-number :model-value="row.sort" :min="0" :max="99999" :precision="0" controls-position="right" :disabled="!mayEditItem(row) || !row.available || busy || loading || !!loadError" :aria-label="`调整 ${row.title || row.displayTitle || row.sourceId} 的排序`" @change="(value: number | undefined) => changeSort(row, value)" /></template></el-table-column>
            <el-table-column label="操作" width="220" fixed="right"><template #default="{ row }">
              <el-button link type="primary" :disabled="!mayEditItem(row) || busy || loading || !!loadError" @click="openEdit(row)">编辑</el-button>
              <el-button link type="primary" :disabled="!mayToggleItem(row) || busy || loading || !!loadError" @click="changeItem(row, { enabled: !row.enabled })">{{ row.enabled ? '下架' : '上架' }}</el-button>
              <el-button link type="danger" :disabled="!mayRemoveItem(row) || busy || loading || !!loadError" @click="removeItem(row)">移除</el-button>
            </template></el-table-column>
          </el-table>
        </AdminTablePanel>
      </el-tab-pane>
    </el-tabs>
    <ItemDialog v-model="dialogVisible" :item="editing" :editable="canEdit() && !loading && !loadError" :busy="busy" :error="actionError" @save="saveItem" />
  </AdminPageShell>
</template>

<style scoped>
.saved-link { display: flex; align-items: baseline; flex-wrap: wrap; gap: var(--admin-ui-space-2) var(--admin-ui-space-4); font-size: var(--admin-ui-font-size-body); overflow-wrap: anywhere; }
.muted { color: var(--admin-ui-color-text-secondary); font-size: var(--admin-ui-font-size-caption); margin: var(--admin-ui-space-1) 0 0; }
.unavailable { color: var(--admin-ui-color-danger); font-size: var(--admin-ui-font-size-caption); margin: var(--admin-ui-space-1) 0 0; }
:deep(.el-input-number) { max-width: 100%; }
</style>
