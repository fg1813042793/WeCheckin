<template>
  <div class="draggable-list">
    <div
      v-for="(q, idx) in questions"
      :key="q.id"
      class="question-card"
      :class="{ selected: q.id === selectedId }"
      @click="$emit('select', q.id)"
    >
      <div
        class="card-handle"
        draggable="true"
        @dragstart="onDragStart(idx, $event)"
        @dragover.prevent
        @drop="onDrop(idx, $event)"
      >⠿</div>
      <div class="card-index">{{ idx + 1 }}.</div>
      <div class="card-body" @click.stop="$emit('select', q.id)">
        <div class="card-title">
          <el-tag size="small" :type="tagType(q.type)">{{ typeName(q.type) }}</el-tag>
          <span class="title-text">{{ q.title || '未命名题目' }}</span>
          <el-tag v-if="q.required" type="danger" size="small" effect="plain">必填</el-tag>
        </div>
        <div v-if="q.description" class="card-desc">{{ q.description }}</div>
      </div>
      <div class="card-actions">
        <el-tooltip content="删除" placement="top">
          <el-button size="small" text type="danger" @click.stop="removeAt(idx)">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/></svg>
          </el-button>
        </el-tooltip>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ questions: any[]; selectedId: string | null }>()
const emit = defineEmits<{ select: [id: string]; 'update:questions': [q: any[]] }>()

const typeName = (t: string) => {
  const map: Record<string, string> = {
    input: '单行文本',
    text: '文本',
    textarea: '多行',
    number: '数字',
    select: '下拉',
    radio: '单选',
    checkbox: '多选',
    picker: '选择器',
    rating: '评分',
    date: '日期',
    time: '时间',
    dateRange: '日期范围',
    file: '文件',
    signature: '签名',
    location: '位置',
    phone: '手机',
    email: '邮箱',
    idCard: '身份证',
    password: '密码',
    switch: '开关',
    matrixRadio: '矩阵单选',
    autopop: '自动填充',
    divider: '分割线',
    description: '说明',
    judge: '判断',
    nps: 'NPS评分'
  }
  return map[t] || t
}
const tagType = (t: string) => {
  const map: Record<string, string> = {
    input: '', text: '', textarea: '', number: '',
    select: 'warning', radio: 'warning', checkbox: 'warning', picker: 'warning', judge: 'warning',
    rating: 'success', date: '', time: '', dateRange: '',
    file: '', signature: '', location: '',
    phone: '', email: '', idCard: '', password: '', switch: '',
    matrixRadio: '', autopop: '',
    divider: 'info', description: 'info',
    nps: 'success'
  }
  return map[t] as any || ''
}

let dragIndex = -1

const onDragStart = (idx: number, e: DragEvent) => {
  dragIndex = idx
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
  }
}

const onDrop = (idx: number, _e: DragEvent) => {
  if (dragIndex < 0 || dragIndex === idx) return
  const arr = [...props.questions]
  const moved = arr.splice(dragIndex, 1)[0]
  arr.splice(idx, 0, moved)
  emit('update:questions', arr)
  dragIndex = -1
}

const removeAt = (idx: number) => {
  const arr = props.questions.filter((_, i) => i !== idx)
  emit('update:questions', arr)
}
</script>

<style scoped>
.draggable-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.question-card {
  display: flex;
  align-items: flex-start;
  background: #fff;
  border-radius: 8px;
  padding: 10px 12px;
  cursor: pointer;
  transition: all 0.15s;
  position: relative;
}
.question-card:hover {
  background: #fafafa;
}
.question-card.selected {
  background: #fff5f5;
}
.question-card::before {
  content: ''; position:absolute; left:0; top:4px; bottom:4px; width:3px;
  border-radius:0 2px 2px 0; background:transparent; transition:all 0.15s;
}
.question-card.selected::before { background:#fb454c; }
.card-handle {
  color: #d0d0d0;
  font-size: 16px;
  margin-right: 6px;
  cursor: grab;
  user-select: none;
  line-height: 24px;
  flex-shrink: 0;
  opacity: 0;
  transition: opacity 0.15s;
}
.question-card:hover .card-handle { opacity: 1; }
.card-index {
  font-size: 15px;
  font-weight: 600;
  color: #fb454c;
  margin-right: 8px;
  line-height: 24px;
  min-width: 24px;
  flex-shrink: 0;
}
.card-body {
  flex: 1;
  min-width: 0;
}
.card-title {
  display: flex;
  align-items: center;
  gap: 6px;
}
.title-text {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.card-desc {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.card-actions {
  display: flex;
  gap: 2px;
  opacity: 0;
  transition: opacity 0.15s;
  flex-shrink: 0;
}
.question-card:hover .card-actions,
.question-card.selected .card-actions {
  opacity: 1;
}
</style>
