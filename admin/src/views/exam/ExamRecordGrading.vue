<script setup lang="ts">
import { computed, nextTick, onActivated, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { AdminPageShell, AdminPageHeader, AdminTablePanel } from '@/components/admin-ui'
import { examGradingApi, hasExamGradingPermission } from '@/api/examGrading'
import type { ExamGradingContext, ExamGradingDraft, ExamGradingSubmission } from '@/types/examGrading'
import ExamGradingQuestion from './components/ExamGradingQuestion.vue'
import { buildSubmission, gradingErrorMessage, gradingPreview, retainGradingDraft, scoreError } from './components/examGradingModel'

const route = useRoute()
const router = useRouter()
// 页签按 fullPath 缓存；固定本实例参数，避免其他页签的路由污染草稿。
const examId = Number(route.query.examId)
const recordId = Number(route.params.recordId)
const returnTo = typeof route.query.returnTo === 'string' && /^\/exam\/responses(?:\?|$)/.test(route.query.returnTo)
  ? route.query.returnTo : `/exam/responses?examId=${examId}`
const permitted = ref(hasExamGradingPermission())
const context = ref<ExamGradingContext | null>(null)
const draft = ref<ExamGradingDraft>({})
const comment = ref('')
const loading = ref(false)
const submitting = ref(false)
const loadError = ref('')
const submitError = ref('')
const attempted = ref(false)
const requiresReview = ref(false)
const reviewConfirmed = ref(false)
const errorSummary = ref<HTMLElement>()
const editable = computed(() => permitted.value && context.value?.record.status === 1 && context.value.canGrade)
const manualCount = computed(() => context.value?.questions.filter(question => question.needManual).length || 0)
const preview = computed(() => context.value ? gradingPreview(context.value, draft.value) : { score: 0, remaining: 0 })
const statusText = computed(() => context.value?.record.status === 2 ? '已批改' : context.value?.record.status === 1 ? '待批改' : '进行中')
const readOnlyNotice = computed(() => {
  if (context.value?.record.status === 2) return '本答卷已批改，当前仅查看批改结果，不可重复改分。'
  if (context.value?.record.status === 0) return '考生尚未交卷，当前答卷不可批改。'
  if (!context.value?.canGrade) return '当前答卷不可批改，请查看提示或联系管理员。'
  if (!manualCount.value) return '没有需要人工评分的题目。'
  return ''
})

function formatTime(ms: number) { return ms ? new Date(ms).toLocaleString('zh-CN') : '—' }

async function load() {
  if (loading.value || submitting.value) return
  permitted.value = hasExamGradingPermission()
  if (!permitted.value) return
  loading.value = true
  loadError.value = ''
  try {
    const response = await examGradingApi.get(examId, recordId)
    const previous = context.value
    const latest = response.data
    const canContinue = latest.record.status === 1 && latest.canGrade
    const versionChanged = !!previous && previous.version !== latest.version
    // 仅在成功取得最新答卷后恢复仍有效的草稿，未评题不从默认 0 分初始化。
    draft.value = previous && canContinue ? retainGradingDraft(latest, draft.value) : {}
    if (!previous || !canContinue) comment.value = latest.record.gradingComment || ''
    requiresReview.value = canContinue && (requiresReview.value || versionChanged)
    if (!requiresReview.value || versionChanged) reviewConfirmed.value = false
    context.value = latest
    attempted.value = false
    submitError.value = ''
  } catch (error) {
    loadError.value = gradingErrorMessage(error, '加载答卷失败，请重试')
  } finally { loading.value = false }
}

function updateScore(questionId: number, value: number | null | undefined) {
  if (editable.value && !loading.value && !submitting.value) draft.value[questionId] = value
}

async function showSubmitError(message: string) {
  submitError.value = message
  await nextTick()
  errorSummary.value?.focus()
}

async function submit() {
  if (loading.value || submitting.value) return
  permitted.value = hasExamGradingPermission()
  if (!permitted.value || !editable.value || !context.value) return
  attempted.value = true
  submitError.value = ''
  if (requiresReview.value && !reviewConfirmed.value) {
    await showSubmitError('答卷已更新，请核对最新题目、答案和分值，并勾选确认后再提交')
    return
  }
  let payload: ExamGradingSubmission
  try { payload = buildSubmission(context.value, draft.value, comment.value) }
  catch (error) { await showSubmitError(gradingErrorMessage(error, '请检查评分')); return }
  submitting.value = true
  try {
    try {
      await ElMessageBox.confirm(
        `本次总分为 ${preview.value.score} / ${context.value.record.totalScore} 分，及格分为 ${context.value.passScore} 分。提交后不可重复改分，确认提交？`,
        '确认提交批改', { type: 'warning', confirmButtonText: '确认提交', cancelButtonText: '继续修改', closeOnClickModal: false },
      )
    } catch { return }
    permitted.value = hasExamGradingPermission()
    if (!permitted.value || !editable.value) return
    const response = await examGradingApi.submit(examId, recordId, payload)
    context.value = response.data
    draft.value = {}
    comment.value = response.data.record.gradingComment
    attempted.value = false
    requiresReview.value = false
    reviewConfirmed.value = false
    ElMessage.success('批改已提交')
  } catch (error) {
    await showSubmitError(gradingErrorMessage(error, '提交失败，评分和评语已保留，请重试'))
  } finally { submitting.value = false }
}

function goBack() { if (!submitting.value) void router.push(returnTo) }
onMounted(load)
onActivated(() => { permitted.value = hasExamGradingPermission() })
</script>

<template>
  <AdminPageShell width="form">
    <AdminPageHeader title="考试答卷批改" :description="context?.examTitle || '加载答卷并核对逐题评分'" back @back="goBack">
      <template #status>
        <el-tag v-if="context && permitted" :type="context.record.status === 2 ? 'success' : 'info'">{{ statusText }}</el-tag>
      </template>
      <template #meta><span>考试 #{{ examId }} · 答卷 #{{ recordId }}</span></template>
      <template #actions>
        <el-button v-if="permitted && context" :loading="loading" :disabled="loading || submitting" @click="load">重新加载答卷</el-button>
        <el-button :disabled="submitting" @click="goBack">返回考试记录</el-button>
      </template>
    </AdminPageHeader>

    <AdminTablePanel v-if="!permitted" empty empty-text="无考试答卷批改权限，请联系管理员授权" />
    <AdminTablePanel v-else-if="!context" title="答卷内容" :loading="loading" :error="loadError" :empty="!context" empty-text="未找到考试答卷">
      <template #actions><el-button :disabled="loading" @click="load">重新加载答卷</el-button></template>
    </AdminTablePanel>
    <template v-else>
      <el-alert v-if="loadError" :title="loadError" type="error" show-icon :closable="false" description="重新加载失败，当前仍显示原答卷，评分和评语已保留，请重试重新加载。" />
      <el-alert v-if="context.notice" :title="context.notice" type="warning" show-icon :closable="false" />
      <el-alert v-if="readOnlyNotice" :title="readOnlyNotice" type="info" show-icon :closable="false" />
      <AdminTablePanel title="答卷信息" :min-height="0">
        <div class="grading-content">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="用户ID">{{ context.record.userId && context.record.userId !== '0' ? context.record.userId : '匿名' }}</el-descriptions-item>
            <el-descriptions-item label="提交时间">{{ formatTime(context.record.submitTime) }}</el-descriptions-item>
            <el-descriptions-item label="当前得分">{{ context.record.score }} / {{ context.record.totalScore }}</el-descriptions-item>
            <el-descriptions-item label="及格分">{{ context.passScore }}</el-descriptions-item>
            <el-descriptions-item label="通过结果">{{ context.record.status === 2 ? (context.record.pass === 1 ? '通过' : '未通过') : '待判定' }}</el-descriptions-item>
            <el-descriptions-item label="批改人ID">{{ context.record.gradedBy || '—' }}</el-descriptions-item>
            <el-descriptions-item label="批改时间" :span="2">{{ formatTime(context.record.gradedAt) }}</el-descriptions-item>
          </el-descriptions>
        </div>
      </AdminTablePanel>
      <div v-if="submitError" ref="errorSummary" role="alert" tabindex="-1">
        <el-alert :title="submitError" type="error" show-icon :closable="false" description="评分和评语已保留；如提示答卷变化，请先重新加载答卷。" />
      </div>
      <el-form label-position="top" @submit.prevent="submit">
        <AdminTablePanel title="逐题批改" :count="context.questions.length" :empty="!context.questions.length" empty-text="此答卷没有可展示的题目">
          <ExamGradingQuestion
            v-for="(question, index) in context.questions" :key="question.questionId"
            :question="question" :index="index" :editable="editable" :busy="loading || submitting"
            :pending="context.record.status !== 2 && question.needManual"
            :score="draft[question.questionId]"
            :error="editable && question.needManual && (attempted || draft[question.questionId] !== undefined) ? scoreError(draft[question.questionId], question.fullScore) : ''"
            @score="updateScore(question.questionId, $event)"
          />
        </AdminTablePanel>
        <AdminTablePanel title="批改结论" class="grading-conclusion" :min-height="0">
          <div class="grading-content">
            <div class="grading-total" aria-live="polite">
              <strong>{{ editable ? (preview.remaining ? '已评小计' : '总分预览') : '答卷总分' }}：{{ editable ? preview.score : context.record.score }} / {{ context.record.totalScore }} 分</strong>
              <span v-if="editable && preview.remaining">还有 {{ preview.remaining }} 道人工题未完成有效评分</span>
              <el-tag v-else-if="editable" :type="preview.score >= context.passScore ? 'success' : 'warning'">预计{{ preview.score >= context.passScore ? '通过' : '未通过' }}（以提交结果为准）</el-tag>
            </div>
            <el-form-item v-if="editable && manualCount" label="人工评语（选填）">
              <el-input v-model="comment" type="textarea" :rows="4" :disabled="loading || submitting" :maxlength="1000" show-word-limit placeholder="填写对本次答卷的评语" />
            </el-form-item>
            <template v-else><h3 class="grading-comment-title">人工评语</h3><p class="grading-comment">{{ context.record.gradingComment || '暂无人工评语' }}</p></template>
            <template v-if="editable && requiresReview">
              <el-alert title="答卷已更新，请核对后提交" type="warning" show-icon :closable="false" description="已保留仍适用的评分和评语。请逐题核对最新题目、考生答案、参考答案及满分；已移除题目、不再需要人工评分或超出新满分的评分已丢弃，请补全并调整评分。" />
              <el-checkbox v-model="reviewConfirmed" class="grading-review" :disabled="loading || submitting">我已核对最新题目、答案和分值，并确认当前评分和评语适用</el-checkbox>
            </template>
          </div>
          <template v-if="editable && manualCount" #footer>
            <span class="grading-hint">全部人工题必填整数分数；提交后不可重复改分。</span>
            <el-button type="primary" native-type="submit" :loading="submitting" :disabled="loading || submitting || (requiresReview && !reviewConfirmed)">{{ submitError ? '重试提交批改' : '提交批改' }}</el-button>
          </template>
        </AdminTablePanel>
      </el-form>
    </template>
  </AdminPageShell>
</template>

<style scoped>
.grading-content { padding: var(--admin-ui-space-4); overflow-wrap: anywhere; }
.grading-conclusion { margin-top: var(--admin-ui-section-gap); }
.grading-total { display: flex; flex-wrap: wrap; align-items: center; gap: var(--admin-ui-space-3); margin-bottom: var(--admin-ui-space-4); }
.grading-hint { color: var(--admin-ui-color-text-muted); }
.grading-comment-title { margin: 0; font-size: var(--admin-ui-font-size-body); }
.grading-comment { margin-bottom: 0; white-space: pre-wrap; overflow-wrap: anywhere; }
.grading-review { margin-top: var(--admin-ui-space-3); height: auto; align-items: flex-start; }
.grading-review :deep(.el-checkbox__label) { white-space: normal; overflow-wrap: anywhere; }
</style>
