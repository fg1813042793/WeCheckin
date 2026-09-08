<template>
  <section v-if="issues.length" class="designer-issues" :class="{ 'has-error': errorCount > 0 }">
    <button class="designer-issues-summary" type="button" :aria-expanded="expanded" @click="expanded = !expanded">
      <span class="issues-indicator" />
      <strong>{{ errorCount ? `发现 ${errorCount} 个问题` : `${warningCount} 项待完善` }}</strong>
      <span class="issues-description">问题已在画布标记，发布前建议完成处理</span>
      <span class="issues-toggle">{{ expanded ? '收起' : '查看' }}</span>
    </button>
    <div v-show="expanded" class="designer-issues-list">
      <button v-for="(issue, index) in issues" :key="`${issue.code}-${issue.questionId || 'survey'}-${index}`" type="button" class="designer-issue" @click="locate(issue)">
        <el-tag size="small" :type="issue.severity === 'error' ? 'danger' : 'warning'">
          {{ issue.questionIndex === undefined ? '问卷设置' : `第 ${issue.questionIndex + 1} 项` }}
        </el-tag>
        <span class="issue-message">{{ issue.message }}</span>
        <span class="issue-locate">定位</span>
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { SurveyDesignerIssue } from '../survey-designer-validation'

const props = defineProps<{ issues: SurveyDesignerIssue[] }>()
const emit = defineEmits<{ locate: [issue: SurveyDesignerIssue] }>()
const expanded = ref(false)
const errorCount = computed(() => props.issues.filter(issue => issue.severity === 'error').length)
const warningCount = computed(() => props.issues.length - errorCount.value)

function locate(issue: SurveyDesignerIssue) {
  emit('locate', issue)
}
</script>

<style scoped>
.designer-issues { flex-shrink:0; border-bottom:1px solid #fde7b2; background:#fffbeb; }
.designer-issues.has-error { border-bottom-color:#fecaca; background:#fff7f7; }
.designer-issues-summary {
  width:100%; min-height:38px; padding:7px 18px; border:0; background:transparent; color:#92400e;
  display:flex; align-items:center; gap:8px; text-align:left; cursor:pointer;
}
.has-error .designer-issues-summary { color:#b42318; }
.issues-indicator { width:7px; height:7px; border-radius:50%; background:#f59e0b; flex-shrink:0; }
.has-error .issues-indicator { background:#ef4444; }
.designer-issues-summary strong { font-size:12px; white-space:nowrap; }
.issues-description { min-width:0; flex:1; color:#a16207; font-size:12px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.has-error .issues-description { color:#b5473c; }
.issues-toggle { flex-shrink:0; font-size:12px; font-weight:600; }
.designer-issues-list { max-height:180px; overflow-y:auto; border-top:1px solid rgba(245,158,11,.18); padding:5px 10px 7px; }
.designer-issue {
  width:100%; min-height:34px; padding:4px 8px; border:0; border-radius:5px; background:transparent;
  display:flex; align-items:center; gap:8px; color:#475467; text-align:left; cursor:pointer;
}
.designer-issue:hover { background:rgba(255,255,255,.75); }
.issue-message { flex:1; min-width:0; font-size:12px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.issue-locate { flex-shrink:0; color:#2563eb; font-size:12px; font-weight:600; }

@media (max-width: 767px) {
  .designer-issues-summary { padding:7px 12px; }
  .issues-description { display:none; }
  .designer-issue { align-items:flex-start; }
  .issue-message { white-space:normal; }
}
</style>
