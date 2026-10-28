<template>
  <aside class="question-bank-navigator" aria-label="题库和题目导航">
    <section class="question-bank-navigator__banks">
      <header class="question-bank-navigator__header">
        <strong>题库</strong>
        <span>{{ categories.length }} 个分类</span>
      </header>
      <div class="question-bank-navigator__list question-bank-navigator__list--banks">
        <button
          v-for="item in bankItems"
          :key="item.value || '__all__'"
          type="button"
          :class="['question-bank-navigator__bank', { active: activeCategory === item.value }]"
          @click="emit('select-category', item.value)"
        >
          <span>{{ item.label }}</span>
          <small v-if="item.value === activeCategory">{{ total }}</small>
        </button>
      </div>
    </section>

    <section class="question-bank-navigator__questions">
      <header class="question-bank-navigator__header">
        <strong>{{ activeCategoryLabel }}</strong>
        <el-tooltip :disabled="canAdd" content="缺少题库新增权限" placement="top">
          <span>
            <el-button type="primary" size="small" :icon="Plus" :disabled="!canAdd" @click="emit('add-question')">
              新增
            </el-button>
          </span>
        </el-tooltip>
      </header>
      <div class="question-bank-navigator__filters">
        <el-input
          :model-value="keyword"
          clearable
          placeholder="搜索题目标题或题型"
          @update:model-value="emit('update:keyword', String($event || ''))"
          @keyup.enter="emit('search')"
          @clear="emit('search')"
        />
        <el-select
          :model-value="questionType"
          clearable
          placeholder="全部题型"
          @update:model-value="emit('update:questionType', String($event || ''))"
          @change="emit('search')"
        >
          <el-option v-for="item in typeOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
      </div>

      <div v-if="loading" class="question-bank-navigator__state" aria-live="polite">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>正在加载题目...</span>
      </div>
      <div v-else-if="error" class="question-bank-navigator__state question-bank-navigator__state--error">
        <span>{{ error }}</span>
        <el-button size="small" @click="emit('retry')">重新加载</el-button>
      </div>
      <div v-else-if="list.length === 0" class="question-bank-navigator__state">
        <el-empty description="当前题库暂无题目" :image-size="54" />
        <el-button v-if="canAdd" type="primary" plain size="small" @click="emit('add-question')">新增第一道题</el-button>
      </div>
      <div v-else class="question-bank-navigator__list question-bank-navigator__list--questions">
        <button
          v-for="row in list"
          :key="row.id"
          type="button"
          :class="['question-bank-navigator__question', { active: selectedId === Number(row.id) }]"
          @click="emit('select-question', row)"
        >
          <span class="question-bank-navigator__question-icon"><QuestionIcon :type="row.type" /></span>
          <span class="question-bank-navigator__question-copy">
            <strong>{{ questionTitle(row) }}</strong>
            <small>{{ typeName(row.type) }}<template v-if="row.tags"> · {{ row.tags }}</template></small>
          </span>
          <el-icon><ArrowRight /></el-icon>
        </button>
      </div>

      <footer v-if="total > pageSize" class="question-bank-navigator__pagination">
        <el-pagination
          small
          background
          layout="prev, pager, next"
          :current-page="page"
          :page-size="pageSize"
          :total="total"
          :pager-count="5"
          @current-change="emit('update:page', $event)"
        />
      </footer>
    </section>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ArrowRight, Loading, Plus } from '@element-plus/icons-vue'
import QuestionIcon from '../../survey/formkit/QuestionIcon.vue'

type QuestionRow = Record<string, any> & { id: number; type: string }

const props = defineProps<{
  categories: string[]
  activeCategory: string
  keyword: string
  questionType: string
  typeOptions: Array<{ value: string; label: string }>
  list: QuestionRow[]
  selectedId: number
  loading: boolean
  error: string
  total: number
  page: number
  pageSize: number
  canAdd: boolean
  questionTitle: (row: QuestionRow) => string
  typeName: (type: string) => string
}>()

const emit = defineEmits<{
  (event: 'select-category', value: string): void
  (event: 'select-question', row: QuestionRow): void
  (event: 'add-question'): void
  (event: 'update:keyword', value: string): void
  (event: 'update:questionType', value: string): void
  (event: 'update:page', value: number): void
  (event: 'search'): void
  (event: 'retry'): void
}>()

const bankItems = computed(() => [
  { value: '', label: '全部题目' },
  ...props.categories.map(item => ({ value: item, label: item })),
  { value: '__uncategorized__', label: '未分类' },
])

const activeCategoryLabel = computed(() => {
  if (props.activeCategory === '__uncategorized__') return '未分类'
  return props.activeCategory || '全部题目'
})
</script>

<style scoped>
.question-bank-navigator { min-width:0; height:100%; display:grid; grid-template-rows:minmax(190px,34%) minmax(0,66%); border-right:1px solid var(--el-border-color-light); background:#fff; }
.question-bank-navigator__banks { min-height:0; border-bottom:1px solid var(--el-border-color-light); }
.question-bank-navigator__questions { min-height:0; display:flex; flex-direction:column; }
.question-bank-navigator__header { min-height:48px; display:flex; align-items:center; justify-content:space-between; gap:10px; padding:0 14px; border-bottom:1px solid #eef0f3; }
.question-bank-navigator__header strong { color:#303133; font-size:14px; }
.question-bank-navigator__header > span { color:#909399; font-size:12px; }
.question-bank-navigator__filters { display:grid; grid-template-columns:minmax(0,1fr) 112px; gap:8px; padding:10px; border-bottom:1px solid #eef0f3; }
.question-bank-navigator__list { min-height:0; overflow:auto; padding:8px; }
.question-bank-navigator__list--banks { height:calc(100% - 49px); box-sizing:border-box; }
.question-bank-navigator__list--questions { flex:1; }
.question-bank-navigator__bank,.question-bank-navigator__question { width:100%; border:0; border-radius:5px; background:transparent; cursor:pointer; color:#475467; text-align:left; transition:background-color .15s,color .15s; }
.question-bank-navigator__bank:hover,.question-bank-navigator__question:hover { background:#f5f7fa; }
.question-bank-navigator__bank.active,.question-bank-navigator__question.active { background:#ecf5ff; color:#2563eb; }
.question-bank-navigator__bank { display:flex; align-items:center; justify-content:space-between; min-height:34px; padding:7px 9px; }
.question-bank-navigator__bank small { color:#98a2b3; }
.question-bank-navigator__question { display:grid; grid-template-columns:24px minmax(0,1fr) 16px; align-items:center; gap:8px; min-height:54px; margin-bottom:4px; padding:8px; }
.question-bank-navigator__question-icon { display:flex; align-items:center; justify-content:center; width:24px; height:24px; border-radius:5px; background:#f1f5f9; color:#64748b; }
.question-bank-navigator__question-icon :deep(svg) { width:15px; height:15px; }
.question-bank-navigator__question-copy { min-width:0; }
.question-bank-navigator__question-copy strong,.question-bank-navigator__question-copy small { display:block; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.question-bank-navigator__question-copy strong { font-size:13px; font-weight:600; }
.question-bank-navigator__question-copy small { margin-top:4px; color:#98a2b3; font-size:11px; }
.question-bank-navigator__state { flex:1; min-height:160px; display:flex; flex-direction:column; align-items:center; justify-content:center; gap:10px; padding:18px; color:#909399; font-size:13px; text-align:center; }
.question-bank-navigator__state--error { color:#f56c6c; }
.question-bank-navigator__pagination { display:flex; justify-content:center; padding:8px; border-top:1px solid #eef0f3; }
</style>
