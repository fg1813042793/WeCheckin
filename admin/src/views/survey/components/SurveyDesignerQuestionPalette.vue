<template>
  <section class="question-panel" aria-label="题型选择">
    <div class="type-tabs-bar">
      <button class="tab-scroll-btn" type="button" aria-label="上一个题型分类" :disabled="atTabStart" @click="scrollTabs(-1)">
        <el-icon><ArrowLeft /></el-icon>
      </button>
      <div ref="tabViewportRef" class="tab-scroll-viewport">
        <div ref="tabTrackRef" class="tab-scroll-track">
          <button
            v-for="category in categories"
            :key="category.name"
            type="button"
            class="type-tab-btn"
            :class="{ active: activeCategory === category.name }"
            @click="activeCategory = category.name"
          >
            {{ category.label }}
          </button>
        </div>
      </div>
      <button class="tab-scroll-btn" type="button" aria-label="下一个题型分类" :disabled="atTabEnd" @click="scrollTabs(1)">
        <el-icon><ArrowRight /></el-icon>
      </button>
    </div>

    <div class="question-type">
      <button v-for="type in activeTypes" :key="type.type" type="button" class="menu-group-item" @click="addQuestion(type)">
        <span class="item-icon"><QuestionIcon :type="type.type" /></span>
        <span class="item-label">{{ type.displayName }}</span>
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import QuestionIcon from '../formkit/QuestionIcon.vue'
import { categoryDefs } from '../survey-designer-question-types'

interface SurveyQuestionTypeDefinition {
  type: string
  displayName: string
  category?: string
  [key: string]: unknown
}

const props = defineProps<{ types: SurveyQuestionTypeDefinition[] }>()
const emit = defineEmits<{ add: [type: SurveyQuestionTypeDefinition] }>()

const activeCategory = ref('')
const tabViewportRef = ref<HTMLElement | null>(null)
const tabTrackRef = ref<HTMLElement | null>(null)
const scrollPosition = ref(0)
const atTabStart = computed(() => scrollPosition.value <= 0)
const atTabEnd = ref(true)
const categories = computed(() => categoryDefs.filter(category => props.types.some(type => category.types.includes(type.type))))
const activeTypes = computed(() => {
  const category = categoryDefs.find(item => item.name === activeCategory.value)
  return category ? props.types.filter(type => category.types.includes(type.type)) : []
})

watch(categories, availableCategories => {
  if (!availableCategories.some(category => category.name === activeCategory.value)) {
    activeCategory.value = availableCategories[0]?.name || ''
  }
  void nextTick(updateTabScroll)
}, { immediate: true })

onMounted(() => {
  window.addEventListener('resize', updateTabScroll)
  updateTabScroll()
})

onBeforeUnmount(() => window.removeEventListener('resize', updateTabScroll))

function scrollTabs(direction: number) {
  const viewport = tabViewportRef.value
  const track = tabTrackRef.value
  if (!viewport || !track) return
  const maxPosition = Math.max(0, track.scrollWidth - viewport.clientWidth)
  scrollPosition.value = Math.max(0, Math.min(scrollPosition.value + direction * viewport.clientWidth * 0.6, maxPosition))
  track.style.transform = `translateX(${-scrollPosition.value}px)`
  atTabEnd.value = scrollPosition.value >= maxPosition - 1
}

function updateTabScroll() {
  const viewport = tabViewportRef.value
  const track = tabTrackRef.value
  if (!viewport || !track) return
  const maxPosition = Math.max(0, track.scrollWidth - viewport.clientWidth)
  scrollPosition.value = Math.min(scrollPosition.value, maxPosition)
  track.style.transform = `translateX(${-scrollPosition.value}px)`
  atTabEnd.value = scrollPosition.value >= maxPosition - 1
}

function addQuestion(type: SurveyQuestionTypeDefinition) {
  emit('add', type)
}
</script>

<style scoped>
.question-panel { padding:10px; }
.type-tabs-bar { display:flex; align-items:stretch; gap:4px; margin-bottom:12px; }
.tab-scroll-viewport { flex:1; overflow:hidden; }
.tab-scroll-track { display:flex; gap:4px; transition:transform .2s ease; }
.tab-scroll-btn {
  flex:0 0 24px; min-width:24px; padding:0; border:1px solid var(--designer-border); border-radius:6px;
  background:#fff; color:#98a2b3; cursor:pointer; display:flex; align-items:center; justify-content:center;
  transition:color .15s, border-color .15s;
}
.tab-scroll-btn:hover:not(:disabled) { color:var(--designer-accent); border-color:var(--designer-accent); }
.tab-scroll-btn:disabled { opacity:.3; cursor:default; }
.type-tab-btn {
  flex:0 0 auto; height:30px; padding:4px 8px; border:1px solid var(--designer-border); border-radius:7px;
  background:#fff; color:#667085; cursor:pointer; font-size:12px; line-height:20px; white-space:nowrap;
  transition:color .15s, border-color .15s, background .15s, box-shadow .15s;
}
.type-tab-btn:hover { border-color:var(--designer-accent-border); color:var(--designer-accent); background:var(--designer-accent-soft); }
.type-tab-btn.active {
  color:#fff; background:var(--designer-accent); border-color:var(--designer-accent); font-weight:600;
  box-shadow:0 6px 14px rgba(37,99,235,.16);
}
.question-type { padding:2px 0 8px; display:flex; flex-direction:column; gap:6px; }
.menu-group-item {
  width:100%; display:flex; align-items:center; gap:8px; padding:8px 10px; border:1px solid #edf0f5;
  border-radius:8px; background:#fff; color:inherit; cursor:pointer; line-height:20px; text-align:left;
  transition:color .14s, background .14s, border-color .14s, box-shadow .14s, transform .14s;
}
.menu-group-item:hover {
  background:var(--designer-accent-soft); color:var(--designer-accent); border-color:var(--designer-accent-border);
  box-shadow:0 8px 18px rgba(37,99,235,.08); transform:translateY(-1px);
}
.item-icon {
  flex:0 0 20px; width:20px; height:20px; display:inline-flex; align-items:center; justify-content:center;
  color:#98a2b3; font-size:15px; line-height:1;
}
.item-icon :deep(svg) { display:block; width:1em; height:1em; }
.menu-group-item:hover .item-icon { color:var(--designer-accent); }
.item-label { min-width:0; flex:1; font-size:13px; }
</style>
