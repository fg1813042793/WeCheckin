<script setup lang="ts">
import { computed } from 'vue'
import { ElImage } from 'element-plus'
import type { ExamGradingQuestion } from '@/types/examGrading'
import { gradingImageUrl } from './examGradingModel'

const props = defineProps<{
  question: ExamGradingQuestion
  index: number
  editable: boolean
  pending: boolean
  busy: boolean
  score: number | null | undefined
  error: string
}>()
const emit = defineEmits<{ score: [value: number | null | undefined] }>()
const images = computed(() => (props.question.images || []).map(image => ({
  url: gradingImageUrl(image.url), alt: image.alt || '题目配图',
})))
const previewUrls = computed(() => images.value.map(image => image.url).filter(Boolean))
</script>

<template>
  <article class="grading-question" :aria-labelledby="`question-title-${question.questionId}`">
    <div class="grading-question__heading">
      <h3 :id="`question-title-${question.questionId}`">{{ index + 1 }}. {{ question.title || '未命名题目' }}</h3>
      <el-tag :type="pending && !question.unscored ? 'warning' : 'info'" size="small">{{ question.unscored ? '不计分' : pending ? '人工评分' : '只读' }}</el-tag>
    </div>
    <dl v-if="question.materials?.length" class="grading-question__materials" aria-label="题目材料">
      <div v-for="(material, materialIndex) in question.materials" :key="materialIndex">
        <dt>{{ material.label || '补充说明' }}</dt>
        <dd>{{ material.content || '（无内容）' }}</dd>
      </div>
    </dl>
    <section v-if="question.images" class="grading-question__pictures" aria-label="题目配图">
      <figure v-for="(image, imageIndex) in images" :key="imageIndex">
        <ElImage
          v-if="image.url" class="grading-question__image" :src="image.url" :alt="image.alt" fit="contain"
          :preview-src-list="previewUrls" :initial-index="previewUrls.indexOf(image.url)" preview-teleported
        >
          <template #placeholder><span class="grading-question__image-feedback" role="status">图片加载中…</span></template>
          <template #error><span class="grading-question__image-feedback" role="status">图片加载失败</span></template>
        </ElImage>
        <div v-else class="grading-question__image grading-question__image-feedback" role="status">图片地址为空或不受支持</div>
        <figcaption>{{ image.alt }}</figcaption>
      </figure>
      <p v-if="!images.length" class="grading-question__hint">暂无题目配图</p>
    </section>
    <div class="grading-question__answers">
      <div><h4>考生答案</h4><p>{{ question.answer || '（未作答）' }}</p></div>
      <div v-if="!question.unscored"><h4 class="grading-question__reference-title">参考答案</h4><p>{{ question.referenceAnswer || '（未提供参考答案）' }}</p></div>
    </div>
    <el-form-item v-if="!question.unscored && editable && question.needManual" :label="`评分（满分 ${question.fullScore} 分）`" :error="error" required>
      <el-input-number
        :id="`score-${question.questionId}`" :model-value="score" :min="0" :max="question.fullScore"
        :step="1" :controls="false" :disabled="busy" placeholder="请输入整数分数"
        :aria-label="`第 ${index + 1} 题评分，满分 ${question.fullScore} 分`"
        :aria-describedby="`score-help-${question.questionId}`"
        @update:model-value="emit('score', $event)"
      />
      <span :id="`score-help-${question.questionId}`" class="grading-question__hint">必填，0 分有效</span>
    </el-form-item>
    <p v-else-if="!question.unscored" class="grading-question__score">
      得分：<strong>{{ pending ? '待批改' : question.gotScore }}</strong> / {{ question.fullScore }} 分
      <el-tag v-if="!pending && question.gotScore > 0 && question.gotScore < question.fullScore" type="warning" size="small">部分得分</el-tag>
    </p>
  </article>
</template>

<style scoped>
.grading-question { padding: var(--admin-ui-space-4); border-bottom: 1px solid var(--admin-ui-color-border); }
.grading-question:last-child { border-bottom: 0; }
.grading-question__heading { display: flex; align-items: flex-start; gap: var(--admin-ui-space-3); }
.grading-question__heading h3 { flex: 1; min-width: 0; margin: 0; white-space: pre-wrap; overflow-wrap: anywhere; font-size: var(--admin-ui-font-size-section); }
.grading-question__heading .el-tag { flex-shrink: 0; }
.grading-question__materials { display: grid; gap: var(--admin-ui-space-3); margin: var(--admin-ui-space-4) 0; }
.grading-question__materials dt { font-weight: 600; color: var(--admin-ui-color-text-secondary); white-space: pre-wrap; overflow-wrap: anywhere; }
.grading-question__materials dd { margin: var(--admin-ui-space-2) 0 0; white-space: pre-wrap; overflow-wrap: anywhere; line-height: var(--admin-ui-line-height-body); }
.grading-question__pictures { display: flex; flex-wrap: wrap; gap: var(--admin-ui-space-3); margin: var(--admin-ui-space-4) 0; }
.grading-question__pictures figure { width: 240px; max-width: 100%; min-width: 0; margin: 0; }
.grading-question__image { width: 100%; height: 160px; border-radius: var(--admin-ui-radius-sm); background: var(--el-fill-color-light); }
.grading-question__image-feedback { display: flex; align-items: center; justify-content: center; height: 100%; padding: var(--admin-ui-space-2); box-sizing: border-box; color: var(--admin-ui-color-text-muted); }
div.grading-question__image-feedback { height: 160px; }
.grading-question__pictures figcaption { margin-top: var(--admin-ui-space-2); white-space: pre-wrap; overflow-wrap: anywhere; }
.grading-question__answers { display: grid; grid-template-columns: minmax(0, 1fr); gap: var(--admin-ui-space-4); margin: var(--admin-ui-space-4) 0; }
.grading-question__answers h4 { margin: 0 0 var(--admin-ui-space-2); font-size: var(--admin-ui-font-size-body); color: var(--admin-ui-color-text-secondary); }
.grading-question__answers .grading-question__reference-title { color: var(--admin-ui-color-success); }
.grading-question__answers p { margin: 0; white-space: pre-wrap; overflow-wrap: anywhere; line-height: var(--admin-ui-line-height-body); }
.grading-question__hint { margin-left: var(--admin-ui-space-2); color: var(--admin-ui-color-text-muted); }
.grading-question__score { display: flex; flex-wrap: wrap; align-items: center; gap: var(--admin-ui-space-2); margin: 0; }
@media (max-width: 767px) { .grading-question__answers { grid-template-columns: minmax(0, 1fr); } }
</style>
