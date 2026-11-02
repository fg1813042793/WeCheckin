<script setup lang="ts">
import { onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { AdminTablePanel } from '@/components/admin-ui'
import type { DingTalkSendTemplate, DingTalkSendTemplateInput } from '@/types/notificationTemplate'
import { targetSummary, templateMessageTypes } from '../templateHelpers'
import { useTemplateManager } from '../useTemplateManager'
import NotificationTemplateEditor from './NotificationTemplateEditor.vue'

const props = defineProps<{ canEdit: boolean }>()
const { items, loading, error, keyword, page, pageSize, total, busy, mutationError, editorVisible, editing, draft, load, search, openEditor, save, toggle, remove } = useTemplateManager(
  () => props.canEdit,
  async name => {
    try {
      await ElMessageBox.confirm(`删除模板“${name}”？删除只影响未来选择，已发送的消息不会撤回。`, '删除发送模板', { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' })
      return true
    } catch { return false }
  },
)
async function saveTemplate(value: DingTalkSendTemplateInput) { if (await save(value)) ElMessage.success('发送模板已保存') }
async function toggleTemplate(row: DingTalkSendTemplate) { if (await toggle(row)) ElMessage.success(row.enabled ? '模板已停用' : '模板已启用') }
async function deleteTemplate(row: DingTalkSendTemplate) { if (await remove(row)) ElMessage.success('模板已删除') }
onMounted(load)
</script>

<template>
  <div class="template-manager">
    <p class="template-intro">配置可重复使用的钉钉手动通知。在“发送钉钉通知”中选择后，可单独修改本次标题和正文。</p>
    <el-alert v-if="!canEdit" title="当前为只读模式，编辑模板需要消息样式编辑权限。" type="info" :closable="false" />
    <el-alert v-if="mutationError && !editorVisible" :title="mutationError" type="error" :closable="false" />
    <div class="template-search">
      <el-input v-model="keyword" aria-label="搜索模板名称" placeholder="搜索模板名称" maxlength="80" clearable :disabled="loading || busy" @keyup.enter="search" @clear="search" />
      <el-button type="primary" :disabled="loading || busy" @click="search">查询</el-button>
    </div>
    <AdminTablePanel title="发送模板" :count="total" :loading="loading" :error="error" :empty="!items.length" empty-text="暂无匹配模板，可更换关键词或新建模板。">
      <template #actions>
        <el-button :loading="loading" :disabled="busy" @click="load">{{ error ? '重试' : '刷新' }}</el-button>
        <el-button v-if="canEdit" type="primary" :disabled="busy || loading" @click="openEditor()">新建模板</el-button>
      </template>
      <el-table :data="items" row-key="id">
        <el-table-column prop="name" label="名称" min-width="150" show-overflow-tooltip />
        <el-table-column label="消息类型" width="110"><template #default="{ row }">{{ templateMessageTypes.find(type => type.value === row.messageType)?.label }}</template></el-table-column>
        <el-table-column label="跳转目标" min-width="200" show-overflow-tooltip><template #default="{ row }">{{ targetSummary(row.target) }}</template></el-table-column>
        <el-table-column label="状态" width="85"><template #default="{ row }"><el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '停用' }}</el-tag></template></el-table-column>
        <el-table-column prop="version" label="版本" width="65" />
        <el-table-column v-if="canEdit" label="操作" fixed="right" width="230">
          <template #default="{ row }">
            <el-button link type="primary" :disabled="busy || loading" @click="openEditor(row)">编辑</el-button>
            <el-button link type="primary" :disabled="busy || loading" @click="openEditor(row, true)">复制</el-button>
            <el-button link type="primary" :disabled="busy || loading" @click="toggleTemplate(row)">{{ row.enabled ? '停用' : '启用' }}</el-button>
            <el-button link type="danger" :disabled="busy || loading" @click="deleteTemplate(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next" :pager-count="7" :disabled="loading || busy" @current-change="load" />
      </template>
    </AdminTablePanel>
    <NotificationTemplateEditor v-model="editorVisible" :draft="draft" :editing="Boolean(editing)" :can-edit="canEdit" :saving="busy" :error="mutationError" @save="saveTemplate" />
  </div>
</template>

<style scoped>
.template-manager { display: grid; gap: var(--admin-ui-space-4); min-width: 0; }
.template-intro { margin: 0; color: var(--admin-ui-color-text-secondary); line-height: var(--admin-ui-line-height-body); }
.template-search { display: flex; flex-wrap: wrap; align-items: center; gap: var(--admin-ui-space-2); }
.template-search .el-input { flex: 0 1 280px; width: 280px; max-width: 100%; }
</style>
