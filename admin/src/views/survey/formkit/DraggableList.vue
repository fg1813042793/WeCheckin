<template>
  <TransitionGroup name="drag" tag="div" class="draggable-list" :style="{ '--ph-h': phHeight + 'px' }">
    <div
      v-for="(q, idx) in questions"
      :key="q.id"
      class="question-card"
      :class="{ selected: q.id === selectedId, dragging: dragIndex === idx, 'drop-before': overIndex === idx && dragIndex !== idx }"
      @click="$emit('select', q.id)"
      @dragover.prevent
      @dragenter="onDragEnter(idx)"
      @drop="onDrop(idx, $event)"
    >
      <div
        class="card-handle"
        draggable="true"
        @dragstart="onDragStart(idx, $event)"
        @dragend="onDragEnd"
      >⠿</div>
      <div v-if="q.type !== 'description' && q.type !== 'divider' && q.type !== 'pagination' && q.type !== 'questionSet'" class="card-index">{{ idx + 1 }}.</div>
      <div class="card-body" @click.stop="$emit('select', q.id)">
        <QuestionPreview
          :q="q"
          :editing="q.id === selectedId"
          @update:title="v => patchQuestion(q.id, 'title', v)"
          @update:description="v => patchQuestion(q.id, 'description', v)"
          @remove="$emit('remove', q.id)"
          @open-logic="$emit('open-logic', q.id)"
          @copy="onCopy(q.id)"
          @upload-bank="$emit('upload-bank', q.id)"
          @select-option="(idx:number) => $emit('select-option', q.id, idx)"
        />
      </div>
    </div>
    <div
      v-if="dragIndex >= 0" key="__tail__"
      class="list-tail-zone"
      :class="{ active: overIndex === questions.length }"
      @dragover.prevent
      @dragenter="onDragEnter(questions.length)"
      @drop="onDrop(questions.length, $event)"
    >
      <span>拖到此处置底</span>
    </div>
  </TransitionGroup>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import QuestionPreview from './QuestionPreview.vue'

const props = defineProps<{ questions: any[]; selectedId: string | null; editing?: boolean }>()
const emit = defineEmits<{
  select: [id: string]
  remove: [id: string]
  'update:questions': [q: any[]]
  'open-logic': [id: string]
  'upload-bank': [id: string]
  'select-option': [qId: string, optIdx: number]
}>()

let dragIndex = -1
const overIndex = ref(-1)
const phHeight = ref(54)

const onDragStart = (idx: number, e: DragEvent) => {
  dragIndex = idx
  overIndex.value = -1
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
    const card = (e.target as HTMLElement)?.closest('.question-card') as HTMLElement
    if (card) {
      const rect = card.getBoundingClientRect()
      phHeight.value = Math.max(rect.height - 4, 32)
      e.dataTransfer.setDragImage(card, e.clientX - rect.left, e.clientY - rect.top)
    }
  }
}
const onDragEnd = () => {
  dragIndex = -1
  overIndex.value = -1
}
const onDragEnter = (idx: number) => {
  if (dragIndex >= 0 && idx > dragIndex && idx === dragIndex + 1) {
    overIndex.value = -1; return
  }
  overIndex.value = idx
}
const onDrop = (idx: number, _e: DragEvent) => {
  if (dragIndex < 0 || dragIndex === idx) {
    dragIndex = -1; overIndex.value = -1; return
  }
  // 拖到自身下方一位等于没移动
  if (idx > dragIndex && idx === dragIndex + 1) {
    dragIndex = -1; overIndex.value = -1; return
  }
  const arr = [...props.questions]
  const moved = arr.splice(dragIndex, 1)[0]
  const target = idx > dragIndex ? idx - 1 : idx
  arr.splice(target, 0, moved)
  emit('update:questions', arr)
  dragIndex = -1
  overIndex.value = -1
}

const removeAt = (idx: number) => {
  const arr = props.questions.filter((_, i) => i !== idx)
  emit('update:questions', arr)
}

let copyCounter = 0
const onCopy = (id: string) => {
  const src = props.questions.find(q => q.id === id)
  if (!src) return
  copyCounter++
  const clone = JSON.parse(JSON.stringify(src))
  clone.id = id + '_copy_' + copyCounter
  clone.title = (clone.title || '题目') + ' (副本)'
  const arr = [...props.questions]
  const srcIdx = arr.findIndex(q => q.id === id)
  arr.splice(srcIdx + 1, 0, clone)
  emit('update:questions', arr)
  emit('select', clone.id)
}

function patchQuestion(id: string, key: string, val: any) {
  const arr = props.questions.map(q => q.id === id ? { ...q, [key]: val } : q)
  emit('update:questions', arr)
}
</script>

<style scoped>
.draggable-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  position: relative;
}
/* TransitionGroup move animation */
.drag-move,
.drag-enter-active,
.drag-leave-active {
  transition: all 0.25s ease;
}
.drag-enter-from,
.drag-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}
.drag-leave-active { position: absolute; }
.question-card {
  display: flex;
  align-items: flex-start;
  background: #fff;
  border-radius: 8px;
  padding: 10px 12px;
  cursor: pointer;
  transition: all 0.25s ease;
  position: relative;
}
.question-card:hover {
  background: #fafafa;
}
.question-card.selected {
  background: #fff5f5;
}
.question-card.dragging {
  opacity: 0.3;
  outline: 2px dashed #ddd;
  outline-offset: -2px;
}
.question-card.drop-before {
  margin-top: calc(var(--ph-h, 54px) + 10px);
  position: relative;
}
.question-card.drop-before::after {
  content: '释 放 到 此 位 置';
  position: absolute;
  left: 0;
  right: 0;
  top: calc(-1 * var(--ph-h, 54px) - 8px);
  height: var(--ph-h, 54px);
  border: 2px dashed #fb454c;
  border-radius: 8px;
  background: rgba(251, 69, 76, 0.04);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fb454c;
  font-size: 12px;
  letter-spacing: 2px;
  box-sizing: border-box;
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
.list-tail-zone {
  margin-top: 4px; border:2px dashed transparent; border-radius:8px; padding:12px;
  text-align:center; color:#ccc; font-size:13px; transition:all 0.2s;
}
.list-tail-zone.active {
  border-color:#fb454c; background:rgba(251,69,76,0.04); color:#fb454c; padding:32px 12px;
}
.tail-indicator { display:none; }
</style>
