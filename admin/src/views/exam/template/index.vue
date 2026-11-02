<template>
  <AdminPageShell width="fluid" density="compact" class="exam-template-page">
    <AdminPageHeader title="考试通知模版" description="管理考试交卷后的统计通知内容，内置模版全局可用，自定义模版仅属于当前管理员。">
      <template #actions><el-button circle :icon="Refresh" title="刷新" :loading="loading" @click="loadTemplates" /></template>
    </AdminPageHeader>
    <el-alert type="info" :closable="false" show-icon>
      <template #title>支持变量：<code>{title}</code> 考试名称、<code>{total}</code> 总提交数、<code>{submitter}</code> 提交人、<code>{date}</code> 提交时间、<code>{score}</code> 得分、<code>{fullScore}</code> 满分、<code>{result}</code> 统计结果</template>
    </el-alert>
    <el-alert v-if="error" :title="error" type="error" :closable="false" show-icon><template #default><el-button size="small" @click="loadTemplates">重新加载</el-button></template></el-alert>
    <section v-loading="loading" class="template-panel">
      <el-tabs v-model="activeScope"><el-tab-pane label="内置模版" name="builtin" /><el-tab-pane label="自定义模版" name="custom" /></el-tabs>
      <div class="template-toolbar">
        <div><strong>{{ activeScope === 'builtin' ? '内置模版' : '自定义模版' }}</strong><small>{{ activeScope === 'builtin' ? '所有考试设计器均可使用' : '仅当前管理员可见' }}</small></div>
        <el-button v-if="canEdit" type="primary" :icon="Plus" @click="openAdd">新增模版</el-button>
      </div>
      <el-table v-if="activePresets.length" :data="activePresets" border stripe>
        <el-table-column label="名称" prop="label" width="200" />
        <el-table-column label="模版内容" min-width="480"><template #default="{ row }"><div class="template-content">{{ row.value }}</div></template></el-table-column>
        <el-table-column v-if="canEdit" label="操作" width="170" fixed="right"><template #default="{ $index }"><div class="admin-table-actions"><el-button text type="primary" @click="openEdit($index)">编辑</el-button><el-button text type="danger" @click="deletePreset($index)">删除</el-button></div></template></el-table-column>
      </el-table>
      <el-empty v-else-if="!loading" :description="activeScope === 'builtin' ? '暂无内置模版' : '暂无自定义模版'" />
    </section>
    <el-dialog v-model="dialog.visible" :title="dialog.isEdit ? '编辑模版' : '新增模版'" width="min(620px, 92vw)" :close-on-click-modal="false">
      <el-form label-position="top"><el-form-item label="模版名称"><el-input v-model="dialog.label" maxlength="50" placeholder="请输入模版名称" /></el-form-item><el-form-item label="模版内容"><el-input v-model="dialog.value" type="textarea" :rows="10" placeholder="请输入模版内容" /></el-form-item></el-form>
      <template #footer><el-button @click="dialog.visible=false">取消</el-button><el-button type="primary" :loading="Boolean(savingScope)" @click="confirmDialog">保存</el-button></template>
    </el-dialog>
  </AdminPageShell>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh } from '@element-plus/icons-vue'
import { AdminPageHeader, AdminPageShell } from '@/components/admin-ui'
import { hasPerm } from '@/utils/permission'
import type { ExamNotificationTemplateScope, TemplatePreset } from '@/api'
import { useExamNotificationTemplates } from './useExamNotificationTemplates'

const activeScope = ref<ExamNotificationTemplateScope>('builtin')
const canEdit = computed(() => hasPerm('admin:menu:exam:templates:edit') || hasPerm('admin:menu:exam:edit'))
const { builtinPresets, customPresets, loading, error, savingScope, loadTemplates, saveTemplates } = useExamNotificationTemplates()
const activePresets = computed(() => activeScope.value === 'builtin' ? builtinPresets.value : customPresets.value)
const dialog = reactive({ visible:false, isEdit:false, index:-1, label:'', value:'' })
function openAdd() { Object.assign(dialog, { visible:true, isEdit:false, index:-1, label:'', value:'' }) }
function openEdit(index:number) { const item=activePresets.value[index]; if(item) Object.assign(dialog,{visible:true,isEdit:true,index,label:item.label,value:item.value}) }
async function confirmDialog() { const label=dialog.label.trim(), value=dialog.value.trim(); if(!label)return void ElMessage.warning('请输入模版名称'); if(!value)return void ElMessage.warning('请输入模版内容'); const next=activePresets.value.map(item=>({...item})); if(dialog.isEdit&&dialog.index>=0)next[dialog.index]={label,value}; else next.push({label,value}); await saveTemplates(activeScope.value,next); dialog.visible=false; ElMessage.success('保存成功') }
async function deletePreset(index:number) { const item=activePresets.value[index]; if(!item)return; try{await ElMessageBox.confirm(`确认删除模版「${item.label}」？`,'删除模版',{type:'warning'})}catch{return}; await saveTemplates(activeScope.value,activePresets.value.filter((_,i)=>i!==index).map(value=>({...value}))); ElMessage.success('删除成功') }
onMounted(loadTemplates)
</script>

<style scoped>
.template-panel{min-height:360px;padding:0 18px 18px;border:1px solid var(--admin-ui-color-border);border-radius:8px;background:#fff}.template-toolbar{min-height:56px;display:flex;align-items:center;justify-content:space-between;gap:16px}.template-toolbar>div{min-width:0;display:flex;flex-direction:column;gap:4px}.template-toolbar strong{color:#303133;font-size:14px}.template-toolbar small{color:#909399;font-size:12px}.template-content{max-height:220px;overflow:auto;white-space:pre-wrap;color:#475467;font-size:13px;line-height:1.65}code{padding:1px 4px;border-radius:4px;background:#eef4ff;color:#2563eb}@media(max-width:767px){.template-toolbar{align-items:stretch;flex-direction:column;padding:10px 0}}
</style>
