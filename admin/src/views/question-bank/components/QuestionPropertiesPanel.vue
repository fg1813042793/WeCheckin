<template>
  <aside class="question-properties" aria-label="题目属性">
    <header class="question-properties__header">
      <strong>题目属性</strong>
      <el-tag v-if="form.type" size="small" type="info">{{ typeName(form.type) }}</el-tag>
    </header>
    <div v-if="!active" class="question-properties__empty">选择或新增题目后编辑属性</div>
    <el-form v-else label-position="top" class="question-properties__form" :disabled="readonly">
      <section class="question-properties__section">
        <h3>归属信息</h3>
        <el-form-item label="题型">
          <el-select v-model="form.type" filterable @change="emit('type-change', String($event || ''))">
            <el-option v-for="item in typeOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="所属题库">
          <el-select v-model="form.category" filterable allow-create clearable placeholder="未分类">
            <el-option v-for="item in categories" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="form.tags" placeholder="多个标签用逗号分隔" />
        </el-form-item>
      </section>

      <section class="question-properties__section">
        <h3>基础设置</h3>
        <div class="question-properties__switches">
          <label><span>必填</span><el-switch v-model="visualQuestion.required" /></label>
          <label><span>只读</span><el-switch v-model="visualQuestion.readOnly" /></label>
        </div>
        <el-form-item label="提示语"><el-input v-model="visualQuestion.placeholder" placeholder="请输入提示语" /></el-form-item>
        <el-form-item label="题目说明"><el-input v-model="visualQuestion.description" type="textarea" :rows="3" placeholder="请输入题目说明" /></el-form-item>
      </section>

      <section v-if="visualQuestion.type === 'file'" class="question-properties__section">
        <h3>上传设置</h3>
        <el-form-item label="文件类型">
          <el-select v-model="visualQuestion.fileTypes" multiple clearable>
            <el-option label="图片" value="image" /><el-option label="文档" value="document" /><el-option label="视频" value="video" /><el-option label="音频" value="audio" />
          </el-select>
        </el-form-item>
        <div class="question-properties__number-grid">
          <el-form-item label="大小 MB"><el-input-number v-model="visualQuestion.maxFileSize" :min="1" :max="200" /></el-form-item>
          <el-form-item label="数量"><el-input-number v-model="visualQuestion.maxFileCount" :min="1" :max="20" /></el-form-item>
        </div>
      </section>

      <section v-if="visualQuestion.type === 'rating' || visualQuestion.type === 'nps'" class="question-properties__section">
        <h3>评分设置</h3>
        <el-form-item label="最大分值"><el-input-number v-model="visualQuestion.props.maxRating" :min="1" :max="20" /></el-form-item>
      </section>

      <section v-if="activeScope === 'exam'" class="question-properties__section">
        <h3>考试设置</h3>
        <el-form-item label="分值"><el-input-number v-model="visualQuestion.examScore" :min="0" :step="1" /></el-form-item>
        <el-form-item label="正确答案"><el-input v-model="visualQuestion.examCorrectAnswer" placeholder="请输入正确答案" /></el-form-item>
        <el-form-item label="答案解析"><el-input v-model="visualQuestion.examAnalysis" type="textarea" :rows="4" placeholder="请输入答案解析" /></el-form-item>
      </section>
    </el-form>
    <footer v-if="active" class="question-properties__footer">
      <el-tooltip :disabled="canDelete" content="缺少题库删除权限" placement="top">
        <span><el-button type="danger" plain :disabled="!canDelete" @click="emit('delete-question')">删除当前题目</el-button></span>
      </el-tooltip>
    </footer>
  </aside>
</template>

<script setup lang="ts">
import type { BankScope } from '../utils/importExport'

defineProps<{
  active: boolean
  activeScope: BankScope
  form: Record<string, any>
  visualQuestion: Record<string, any>
  categories: string[]
  typeOptions: Array<{ value: string; label: string }>
  readonly: boolean
  canDelete: boolean
  typeName: (type: string) => string
}>()

const emit = defineEmits<{
  (event: 'type-change', value: string): void
  (event: 'delete-question'): void
}>()
</script>

<style scoped>
.question-properties { min-width:0; height:100%; display:flex; flex-direction:column; border-left:1px solid var(--el-border-color-light); background:#fff; }
.question-properties__header { min-height:48px; display:flex; align-items:center; justify-content:space-between; gap:10px; padding:0 14px; border-bottom:1px solid #eef0f3; }
.question-properties__header strong { color:#303133; font-size:14px; }
.question-properties__form { flex:1; min-height:0; overflow:auto; padding:14px 14px 24px; }
.question-properties__section { padding-bottom:4px; }
.question-properties__section + .question-properties__section { margin-top:12px; padding-top:14px; border-top:1px solid #eef0f3; }
.question-properties__section h3 { margin:0 0 12px; color:#344054; font-size:13px; }
.question-properties__form :deep(.el-form-item) { margin-bottom:14px; }
.question-properties__form :deep(.el-form-item__label) { margin-bottom:5px; color:#667085; font-size:12px; font-weight:600; }
.question-properties__form :deep(.el-select),.question-properties__form :deep(.el-input-number) { width:100%; }
.question-properties__switches { display:grid; grid-template-columns:1fr 1fr; gap:10px; margin-bottom:14px; }
.question-properties__switches label { min-height:34px; display:flex; align-items:center; justify-content:space-between; gap:8px; padding:0 10px; border:1px solid #e4e7ed; border-radius:6px; color:#667085; font-size:12px; }
.question-properties__number-grid { display:grid; grid-template-columns:1fr 1fr; gap:10px; }
.question-properties__empty { flex:1; display:flex; align-items:center; justify-content:center; padding:20px; color:#909399; font-size:13px; text-align:center; }
.question-properties__footer { padding:12px 14px; border-top:1px solid #eef0f3; }
.question-properties__footer .el-button,.question-properties__footer > span { width:100%; }
</style>
