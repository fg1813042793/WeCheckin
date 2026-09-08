<template>
  <section class="outline-tree" aria-label="题目大纲">
    <header class="tree-root">
      <div class="tree-root-main">
        <el-icon class="tree-root-icon"><Document /></el-icon>
        <span class="tree-root-title" :title="title || '未命名问卷'">{{ title || '未命名问卷' }}</span>
      </div>
      <span class="tree-root-count">{{ resultCountText }}</span>
    </header>

    <div v-if="outlineChildren.length" class="outline-tools">
      <el-input v-model="keyword" :prefix-icon="Search" clearable placeholder="搜索题目、题型或 ID" />
      <div class="outline-filter-row">
        <el-checkbox v-model="onlyIssues" :disabled="!invalidIds.length">只看待完善</el-checkbox>
        <span v-if="hasActiveFilter" class="outline-filter-result">找到 {{ filteredChildren.length }} 题</span>
      </div>
    </div>

    <div v-if="filteredChildren.length" class="tree-children">
      <button
        v-for="child in filteredChildren"
        :key="`${child.q.id || 'missing'}-${child.index}`"
        type="button"
        class="tree-child"
        :class="{ active: child.q.id === selectedId, invalid: isInvalid(child.q.id) }"
        @click="selectQuestion(child.q.id)"
      >
        <span class="tree-index">{{ child.index }}.</span>
        <span class="tree-child-body">
          <QuestionIcon :type="child.q.type" class="tree-icon" />
          <span class="tree-title" :title="questionTitle(child.q)">{{ questionTitle(child.q) }}</span>
          <span v-if="child.q.required" class="tree-required">必</span>
          <el-tooltip v-if="isInvalid(child.q.id)" content="该题待完善" placement="top">
            <el-icon class="tree-issue"><WarningFilled /></el-icon>
          </el-tooltip>
          <span class="tree-type">{{ typeName(child.q.type) }}</span>
        </span>
      </button>
    </div>
    <el-empty v-else class="outline-empty" :description="emptyDescription" :image-size="40" />
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Document, Search, WarningFilled } from '@element-plus/icons-vue'
import QuestionIcon from '../formkit/QuestionIcon.vue'

interface OutlineQuestion {
  id?: string
  type: string
  title?: string
  required?: boolean
}

interface OutlineChild {
  q: OutlineQuestion
  index: number
}

const props = withDefaults(defineProps<{
  title?: string
  questions: OutlineQuestion[]
  selectedId?: string | null
  invalidIds?: string[]
  typeName: (type: string) => string
}>(), {
  title: '',
  selectedId: null,
  invalidIds: () => [],
})

const emit = defineEmits<{ select: [questionId: string] }>()
const keyword = ref('')
const onlyIssues = ref(false)

const invalidIdSet = computed(() => new Set(props.invalidIds.map(id => String(id))))
const outlineChildren = computed<OutlineChild[]>(() => props.questions
  .filter(question => !['divider', 'pagination', 'description', 'questionSet'].includes(question.type))
  .map((question, index) => ({ q: question, index: index + 1 })))

const normalizedKeyword = computed(() => keyword.value.trim().toLocaleLowerCase())
const hasActiveFilter = computed(() => Boolean(normalizedKeyword.value || onlyIssues.value))
const filteredChildren = computed(() => outlineChildren.value.filter(child => {
  if (onlyIssues.value && !isInvalid(child.q.id)) return false
  if (!normalizedKeyword.value) return true
  const searchable = [questionTitle(child.q), props.typeName(child.q.type), child.q.type, child.q.id]
    .filter(value => value !== undefined && value !== null)
    .join(' ')
    .toLocaleLowerCase()
  return searchable.includes(normalizedKeyword.value)
}))

const resultCountText = computed(() => hasActiveFilter.value
  ? `${filteredChildren.value.length}/${outlineChildren.value.length}题`
  : `${outlineChildren.value.length}题`)

const emptyDescription = computed(() => {
  if (!outlineChildren.value.length) return '暂无题目'
  if (onlyIssues.value && !normalizedKeyword.value) return '暂无待完善题目'
  return '没有匹配题目'
})

watch(() => props.invalidIds.length, count => {
  if (!count) onlyIssues.value = false
})

function questionTitle(question: OutlineQuestion) {
  const title = String(question.title || '')
    .replace(/<[^>]*>/g, '')
    .split(/\r?\n/, 1)[0]
    .replace(/\s+/g, ' ')
    .trim()
  return title || '未命名'
}

function isInvalid(questionId: string | undefined) {
  return questionId !== undefined && invalidIdSet.value.has(String(questionId))
}

function selectQuestion(questionId: string | undefined) {
  if (questionId) emit('select', questionId)
}
</script>

<style scoped>
.outline-tree {
  height:100%; min-height:0; padding:6px 7px 10px; display:flex; flex-direction:column;
  background:#fff; box-sizing:border-box;
}
.tree-root {
  display:flex; align-items:center; justify-content:space-between; gap:8px; flex-shrink:0;
  padding:9px 8px 10px; border-bottom:1px solid var(--designer-border); background:#fff;
}
.tree-root-main { display:flex; align-items:center; gap:8px; min-width:0; }
.tree-root-icon { width:16px; height:16px; color:#667085; flex-shrink:0; }
.tree-root-title {
  flex:1; min-width:0; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;
  font-weight:650; font-size:13px; color:#1f2937;
}
.tree-root-count {
  flex-shrink:0; font-size:11px; color:var(--designer-accent); font-weight:600;
  background:var(--designer-accent-soft); border:1px solid var(--designer-accent-border);
  border-radius:999px; padding:1px 8px; line-height:18px;
}
.outline-tools { flex-shrink:0; padding:8px 5px 6px; border-bottom:1px solid #eef2f6; }
.outline-tools :deep(.el-input__wrapper) { border-radius:7px; }
.outline-filter-row {
  min-height:28px; padding:4px 2px 0; display:flex; align-items:center; justify-content:space-between; gap:8px;
}
.outline-filter-row :deep(.el-checkbox) { height:24px; font-size:12px; }
.outline-filter-result { color:#98a2b3; font-size:11px; white-space:nowrap; }
.tree-children {
  position:relative; flex:1; min-height:0; overflow-y:auto; padding:6px 2px 2px 18px;
  display:flex; flex-direction:column; gap:3px;
}
.tree-children::before {
  content:''; position:absolute; left:9px; top:12px; bottom:10px;
  border-left:1px dashed #cbd5e1; pointer-events:none;
}
.tree-child {
  width:100%; display:flex; align-items:center; gap:7px; cursor:pointer; border-radius:8px; min-height:34px;
  padding:5px 6px; border:1px solid transparent; position:relative; background:transparent; text-align:left;
  transition:background 0.12s, border-color 0.12s, box-shadow 0.12s;
}
.tree-child::after {
  content:''; position:absolute; left:-9px; top:50%; width:9px;
  border-top:1px dashed #cbd5e1; transform:translateY(-50%); pointer-events:none;
}
.tree-child:hover { background:#f8fbff; border-color:#dbeafe; }
.tree-child.active { background:var(--designer-accent-soft); border-color:var(--designer-accent-border); }
.tree-child.active::before {
  content:''; position:absolute; left:0; top:7px; bottom:7px; width:3px;
  border-radius:999px; background:var(--designer-accent);
}
.tree-child.invalid:not(.active) { border-color:#fecaca; background:#fffafa; }
.tree-child-body { display:flex; align-items:center; gap:6px; flex:1; min-width:0; }
.tree-index {
  display:flex; align-items:center; justify-content:flex-end; flex:0 0 26px; height:22px;
  color:#98a2b3; font-size:12px; font-weight:600; line-height:1;
}
.tree-child.active .tree-index { color:var(--designer-accent); }
.tree-icon { flex-shrink:0; font-size:13px; width:16px; text-align:center; color:#98a2b3; }
.tree-child:hover .tree-icon,
.tree-child.active .tree-icon { color:var(--designer-accent); }
.tree-title {
  flex:1; min-width:0; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;
  color:#344054; font-size:13px; line-height:20px;
}
.tree-child.active .tree-title { color:#1f2937; font-weight:650; }
.tree-required {
  display:flex; align-items:center; justify-content:center; flex:0 0 16px; height:16px;
  font-size:10px; color:#dc6803; border-radius:999px; background:#fff7ed;
}
.tree-issue { flex:0 0 14px; width:14px; height:14px; color:#dc2626; }
.tree-type {
  max-width:64px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;
  font-size:10px; color:#667085; padding:1px 6px; border-radius:999px;
  background:#f2f4f7; flex-shrink:0; line-height:16px;
}
.tree-child:hover .tree-type,
.tree-child.active .tree-type { background:var(--designer-accent-soft); color:var(--designer-accent); }
.outline-empty { flex:1; display:flex; align-items:center; justify-content:center; }
.tree-children::-webkit-scrollbar { width:8px; height:8px; }
.tree-children::-webkit-scrollbar-thumb {
  background:#cfd6e2; border-radius:999px; border:2px solid transparent; background-clip:content-box;
}
.tree-children::-webkit-scrollbar-thumb:hover {
  background:#aeb8c8; border:2px solid transparent; background-clip:content-box;
}
</style>
