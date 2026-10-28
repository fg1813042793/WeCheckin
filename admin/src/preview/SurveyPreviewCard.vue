<script setup lang="ts">
import { Document, MoreFilled, EditPen, DataAnalysis, CopyDocument, Delete, SwitchButton } from '@element-plus/icons-vue'
import { statusLabels, type DemoSurvey } from './survey-demo'

defineProps<{ survey: DemoSurvey }>()
const emit = defineEmits<{
  edit: []; detail: []; statistics: []; copy: []; toggle: []; remove: []
}>()
function command(value: string) {
  if (value === 'copy') emit('copy')
  if (value === 'toggle') emit('toggle')
  if (value === 'remove') emit('remove')
}
</script>

<template>
  <article class="survey-preview-card">
    <div class="paper-topline">
      <span class="paper-category"><el-icon><Document /></el-icon>{{ survey.category }}</span>
      <span class="status-label" :class="survey.status"><i />{{ statusLabels[survey.status] }}</span>
    </div>
    <button class="paper-content" type="button" :aria-label="`查看 ${survey.title}`" @click="emit('detail')">
      <h2>{{ survey.title }}</h2>
      <p class="paper-description">{{ survey.description }}</p>
      <div class="paper-question">
        <p><span>01</span>{{ survey.sampleQuestion }}</p>
        <div v-for="option in survey.sampleOptions" :key="option" class="paper-option"><i />{{ option }}</div>
      </div>
    </button>
    <div class="paper-metadata"><span><strong>{{ survey.responses }}</strong> 份答卷<span class="meta-divider">/</span>{{ survey.questions }} 题</span><time :datetime="survey.updated">{{ survey.updated.slice(5).replace('-', '/') }} 更新</time></div>
    <footer class="paper-actions">
      <el-button text type="primary" :icon="EditPen" @click="emit('edit')">编辑</el-button>
      <el-button text :icon="Document" @click="emit('detail')">数据</el-button>
      <el-button text :icon="DataAnalysis" @click="emit('statistics')">统计</el-button>
      <el-dropdown trigger="click" @command="command">
        <el-button text :icon="MoreFilled" :aria-label="`${survey.title} 更多操作`" title="更多操作" />
        <template #dropdown><el-dropdown-menu>
          <el-dropdown-item command="copy" :icon="CopyDocument">复制问卷</el-dropdown-item>
          <el-dropdown-item command="toggle" :icon="SwitchButton">{{ survey.status === 'published' ? '停用问卷' : '发布问卷' }}</el-dropdown-item>
          <el-dropdown-item command="remove" :icon="Delete" divided>删除问卷</el-dropdown-item>
        </el-dropdown-menu></template>
      </el-dropdown>
    </footer>
  </article>
</template>
