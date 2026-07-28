<template>
  <view class="survey" :class="{ 'survey--with-progress': settings.progressBar && !loading && survey }">
    <view v-if="settings.progressBar && !loading && survey" class="survey-progress">
      <view class="survey-progress-inner">
        <view class="survey-progress-track"><view class="survey-progress-fill" :style="{ width: progressPct + '%' }" /></view>
        <text class="survey-progress-label">{{ answeredCount }}/{{ totalQuestions }}</text>
      </view>
    </view>

    <view v-if="loading" class="survey-loading">
      <view class="spinner"><view v-for="i in 5" :key="i" :class="'rect' + i" /></view>
    </view>

    <view v-else-if="error" class="survey-error">{{ error }}</view>

    <view v-else-if="submitted" class="survey-end">
      <view class="survey-end-card">
        <text class="survey-end-title">{{ endContent ? '' : '已提交，感谢填写' }}</text>
        <view v-if="endContent" class="survey-end-desc"><rich-text :nodes="endContent" /></view>
        <view v-if="settings.redirectUrl" class="survey-end-action">
          <button class="survey-btn survey-btn--primary" @click="goRedirect">查看结果</button>
        </view>
        <view v-else class="survey-end-action">
          <button class="survey-btn survey-btn--primary" @click="goBack">返回</button>
        </view>
      </view>
    </view>

    <view v-else-if="survey" class="survey-card">
      <view v-if="headerImage" class="survey-card-img"><image :src="headerImage" mode="aspectFill" /></view>
      <view class="survey-header">
        <text class="survey-title">{{ survey.title }}</text>
        <text v-if="survey.description" class="survey-desc">{{ survey.description }}</text>
        <view class="survey-meta">
          <text v-if="survey.anonymous === 1" class="survey-tag survey-tag--anon">匿名收集</text>
          <text v-if="survey.showResult === 1" class="survey-tag survey-tag--result">提交后查看结果</text>
          <text class="survey-tag">{{ totalQuestions }} 道题</text>
          <text v-if="remaining > 0" class="survey-tag" :class="remaining < 60000 ? 'survey-tag--danger' : 'survey-tag--warn'">⏱ {{ formatRemaining(remaining) }}</text>
        </view>
      </view>

      <view v-if="settings.onePageOneQuestion && currentQuestion" class="survey-form survey-form--single" @touchstart="onSwipeStart" @touchend="onSwipeEnd">
        <view class="survey-form-scroll">
          <QuestionField
            :q="currentQuestion"
            :index="currentQIndex"
            :value="answers[currentQuestion.id]"
            :file-list="fileLists[currentQuestion.id] || []"
            :show-number="settings.questionNumber !== false"
            @input="(v) => setAnswer(currentQuestion.id, v)"
            @update:fileList="(v) => $set(fileLists, currentQuestion.id, v)"
            @sig-open="isSigOpen = true" @sig-close="isSigOpen = false"
          />
        </view>
        <view class="survey-nav">
          <view :class="['survey-nav-btn', { disabled: currentQIndex <= 0 }]" @click="goPrev">上一题</view>
          <text class="survey-nav-index">{{ currentQIndex + 1 }} / {{ totalQuestions }}</text>
          <view v-if="!isLast" class="survey-nav-btn survey-nav-btn--primary" @click="goNext">下一题</view>
          <view v-else class="survey-nav-btn survey-nav-btn--primary" @click="onSubmit">提交</view>
        </view>
      </view>

      <view v-else class="survey-form">
        <view class="survey-form-scroll survey-form-scroll--all">
          <QuestionField
            v-for="(q, i) in realQuestions" :key="q.id"
            :q="q"
            :index="i"
            :value="answers[q.id]"
            :file-list="fileLists[q.id] || []"
            :show-number="settings.questionNumber !== false"
            @input="(v) => setAnswer(q.id, v)"
            @update:fileList="(v) => $set(fileLists, q.id, v)"
            @sig-open="isSigOpen = true" @sig-close="isSigOpen = false"
          />
        </view>
        <view class="survey-footer">
          <button class="survey-btn survey-btn--primary survey-btn--lg" :loading="submitting" @click="onSubmit">提交</button>
        </view>
      </view>
    </view>

    <view v-if="!loading && survey && settings.answerSheetVisible !== false">
      <view class="survey-sheet-btn" :style="sheetBtnStyle" @touchstart="onSheetDragStart" @touchmove.prevent="onSheetDragMove" @touchend="onSheetDragEnd" @click="openSheet">
        <text class="survey-sheet-btn-label">{{ answeredCount }}/{{ totalQuestions }}</text>
      </view>
      <view v-if="showSheet" class="survey-sheet-overlay" @click="showSheet = false" />
      <view v-if="showSheet" class="survey-sheet-panel">
        <text class="survey-sheet-title">答题卡</text>
        <view class="survey-sheet-grid">
          <view v-for="(q, i) in realQuestions" :key="q.id"
            class="survey-sheet-item"
            :class="{
              'survey-sheet-item--done': isAnswered(q, answers[q.id]),
              'survey-sheet-item--cur': settings.onePageOneQuestion && q === currentQuestion
            }"
            @click="jumpToQuestion(q)"
          >{{ i + 1 }}</view>
        </view>
        <view class="survey-sheet-stat">{{ answeredCount }}/{{ totalQuestions }}</view>
      </view>
    </view>

    <view v-if="backLeaveDialogVisible" class="survey-leave-mask" @click.stop>
      <view class="survey-leave-dialog" @click.stop>
        <text class="survey-leave-title">是否提交问卷？</text>
        <text class="survey-leave-desc">倒计时仍在进行，当前已完成 {{ answeredCount }}/{{ totalQuestions }} 道题。你可以继续填写、立即提交，或放弃本次填写。</text>
        <view class="survey-leave-actions">
          <view class="survey-leave-btn" @click="closeBackLeaveDialog">继续填写</view>
          <view class="survey-leave-btn survey-leave-btn--primary" @click="submitFromBackPrompt">提交问卷</view>
          <view class="survey-leave-btn survey-leave-btn--danger" @click="abandonTimedSurvey">放弃填写</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { surveyApi } from '../../api/index'
import QuestionField from '../../components/survey/QuestionField.vue'
import { evaluateFrontendRules } from '../../utils/logicEngine'
import {
  LAYOUT_TYPES,
  formatRemainingTime,
  getAnswerProgress,
  getQuestionInitialValue,
  isQuestionAnswered,
  parseQuestions,
  parseSettings
} from '../../utils/formFill'

export default {
  components: { QuestionField },
  data() {
    return {
      survey: null,
      settings: {},
      questions: [],
      answers: {},
      loading: true,
      error: '',
      submitting: false,
      submitted: false,
      endContent: '',
      currentQIndex: 0,
      fileLists: {},
      remaining: 0,
      startAt: 0,
      timer: null,
      showSheet: false,
      sheetX: 0,
      sheetY: 0,
      sheetDragStartX: 0,
      sheetDragStartY: 0,
      sheetOrigX: 0,
      sheetOrigY: 0,
      sheetDragging: false,
      autoSaveTimer: null,
      session: '',
      swipeStartX: 0,
      swipeStartY: 0,
      isSigOpen: false,
      backLeaveDialogVisible: false,
      abandoning: false,
      hiddenIds: [],
      logicRules: []
    }
  },
  computed: {
    answerProgress() { return getAnswerProgress(this.questions, this.answers, { hiddenIds: this.hiddenIds }) },
    totalQuestions() { return this.answerProgress.total },
    answeredCount() { return this.answerProgress.answered },
    progressPct() { return this.answerProgress.percent },
    realQuestions() { return (this.questions || []).filter(q => !LAYOUT_TYPES.includes(q.type) && this.hiddenIds.indexOf(q.id) < 0) },
    isLast() { return this.currentQIndex >= this.totalQuestions - 1 },
    currentQuestion() { return this.settings.onePageOneQuestion ? (this.realQuestions[this.currentQIndex] || null) : null },
    headerImage() {
      const imgs = this.settings.headerImages
      if (!imgs?.length) return ''
      return typeof imgs[0] === 'string' ? imgs[0] : (imgs[0]?.url || '')
    },
    sheetBtnStyle() {
      const s = { transform: `translate(${this.sheetX}px, ${this.sheetY}px)` }
      if (this.sheetDragging) s.opacity = 0.8
      return s
    }
  },
  onLoad(query) {
    this.surveyId = Number(query.id)
    this.load()
  },
  onUnload() {
    this.stopCountdown()
    this.stopAutoSave(!this.abandoning)
  },
  onPullDownRefresh() {
    if (this.surveyId) {
      this.load().finally(() => { uni.stopPullDownRefresh() })
    } else {
      uni.stopPullDownRefresh()
    }
  },
  onBackPress() {
    if (this.backLeaveDialogVisible) return true
    if (this.shouldConfirmTimedLeave()) {
      this.backLeaveDialogVisible = true
      return true
    }
    return false
  },
  methods: {
    async load() {
      if (!this.surveyId) { this.error = '参数错误'; this.loading = false; return }
      try {
        this.session = this.loadSession() || ''
        const res = await surveyApi.getDetail({ id: this.surveyId, session: this.session })
        if (res.code !== 0) { this.error = res.msg || '加载失败'; this.loading = false; return }
        this.survey = res.data
        this.settings = parseSettings(res.data?.settings)
        const rawRules = this.settings.logicRules
        if (rawRules) {
          if (typeof rawRules === 'string') { try { this.logicRules = JSON.parse(rawRules) } catch { this.logicRules = [] } }
          else if (Array.isArray(rawRules)) this.logicRules = rawRules
        }
        this.questions = parseQuestions(res.data?.schema)
        const init = {}
        this.questions.forEach(q => { init[q.id] = this.getInitVal(q) })
        if (this.settings.autoSave) {
          const saved = this.loadFromStorage()
          if (saved) {
            if (saved.answers) Object.keys(saved.answers).forEach(k => { if (k in init) init[k] = saved.answers[k] })
          }
        }
        this.answers = init
        this.reevaluateRules()
        if (res.data?.session) { this.session = res.data.session; this.saveSession(this.session) }
        if (res.data?.startAt) this.startAt = res.data.startAt
        if (this.settings.timeLimit) this.startCountdown()
      } catch (e) {
        this.error = e.msg || '加载失败'
        uni.showToast({ title: '加载失败', icon: 'none' })
      } finally { this.loading = false }
    },

    shouldConfirmTimedLeave() {
      return !!(
        this.survey &&
        !this.loading &&
        !this.submitted &&
        !this.submitting &&
        !this.abandoning &&
        this.settings?.timeLimit &&
        this.remaining > 0
      )
    },

    closeBackLeaveDialog() {
      this.backLeaveDialogVisible = false
    },

    getInitVal(q) {
      return getQuestionInitialValue(q)
    },

    isAnswered(q, val) {
      return isQuestionAnswered(q, val)
    },

    reevaluateRules() {
      if (!this.logicRules || !this.logicRules.length) { this.hiddenIds = []; return }
      const frontendRules = this.logicRules.filter(r => r.scope !== 'backend')
      if (!frontendRules.length) { this.hiddenIds = []; return }
      const result = evaluateFrontendRules(frontendRules, this.questions, this.answers)
      this.hiddenIds = Array.from(result.hiddenIds)
    },
    setAnswer(qid, val) {
      this.$set(this.answers, qid, val)
      this.reevaluateRules()
      this.autoSave()
    },

    autoSave() {
      if (!this.settings.autoSave) return
      if (this.autoSaveTimer) { clearTimeout(this.autoSaveTimer); this.autoSaveTimer = null }
      this.autoSaveTimer = setTimeout(() => {
        this.autoSaveTimer = null
        if (this.abandoning) return
        this.saveToStorage()
      }, 800)
    },
    stopAutoSave(flushPending = false) {
      if (!this.autoSaveTimer) return
      clearTimeout(this.autoSaveTimer)
      this.autoSaveTimer = null
      if (flushPending && !this.abandoning) this.saveToStorage()
    },
    saveToStorage() {
      try {
        uni.setStorageSync('survey_' + this.surveyId, JSON.stringify({ answers: this.answers }))
      } catch (e) {}
    },
    loadFromStorage() {
      try {
        const raw = uni.getStorageSync('survey_' + this.surveyId)
        return raw ? JSON.parse(raw) : null
      } catch (e) { return null }
    },
    clearStorage() {
      try { uni.removeStorageSync('survey_' + this.surveyId) } catch (e) {}
      try { uni.removeStorageSync('survey_session_' + this.surveyId) } catch (e) {}
    },
    saveSession(s) {
      try { uni.setStorageSync('survey_session_' + this.surveyId, s) } catch (e) {}
    },
    loadSession() {
      try { return uni.getStorageSync('survey_session_' + this.surveyId) } catch (e) { return '' }
    },

    goNext() { if (this.currentQIndex < this.totalQuestions - 1) this.currentQIndex++ },
    goPrev() { if (this.currentQIndex > 0) this.currentQIndex-- },
    isSigEvent(e) {
      let el = e.target
      while (el) {
        if (el.className && typeof el.className === 'string' && el.className.includes('sig-overlay')) return true
        el = el.parentElement || el.parentNode
      }
      return false
    },

    formatRemaining(ms) {
      return formatRemainingTime(ms)
    },

    startCountdown() {
      if (this.timer) { clearInterval(this.timer); this.timer = null }
      const limit = this.settings.timeLimit
      if (!limit || limit <= 0 || !this.startAt) return
      const tick = () => {
        if (this.abandoning) {
          this.stopCountdown()
          return
        }
        const left = limit * 60 * 1000 - (Date.now() - this.startAt)
        this.remaining = Math.max(0, left)
        if (left <= 0) {
          this.stopCountdown()
          uni.showToast({ title: '作答时间已到，自动提交', icon: 'none' })
          this.forceSubmit()
        }
      }
      tick()
      this.timer = setInterval(tick, 1000)
    },
    stopCountdown() {
      if (!this.timer) return
      clearInterval(this.timer)
      this.timer = null
    },

    openSheet() { if (!this.sheetDragging) this.showSheet = true },
    jumpToQuestion(q) { const idx = this.realQuestions.indexOf(q); if (idx >= 0) this.currentQIndex = idx; this.showSheet = false },

    onSheetDragStart(e) {
      this.sheetDragging = true
      this.sheetDragStartX = e.touches[0].clientX
      this.sheetDragStartY = e.touches[0].clientY
      this.sheetOrigX = this.sheetX
      this.sheetOrigY = this.sheetY
    },
    onSheetDragMove(e) {
      if (!this.sheetDragging) return
      this.sheetX = this.sheetOrigX + e.touches[0].clientX - this.sheetDragStartX
      this.sheetY = this.sheetOrigY + e.touches[0].clientY - this.sheetDragStartY
    },
    onSheetDragEnd() { this.sheetDragging = false },

    onSwipeStart(e) {
      if (this.isSigOpen) return
      this.swipeStartX = e.touches[0].clientX
      this.swipeStartY = e.touches[0].clientY
    },
    goNext() { if (this.currentQIndex < this.totalQuestions - 1) this.currentQIndex++ },
    goPrev() { if (this.currentQIndex > 0) this.currentQIndex-- },
    onSwipeEnd(e) {
      if (!this.settings.onePageOneQuestion) return
      if (this.isSigOpen) return
      if (this.isSigEvent(e)) return
      const dx = e.changedTouches[0].clientX - this.swipeStartX
      const dy = e.changedTouches[0].clientY - this.swipeStartY
      if (Math.abs(dx) < 30 || Math.abs(dy) > Math.abs(dx)) return
      if (dx < 0) this.goNext()
      else this.goPrev()
    },

    validateRequired() {
      const msgs = []
      for (const q of (this.questions || [])) {
        if (LAYOUT_TYPES.includes(q.type)) continue
        if (this.hiddenIds.indexOf(q.id) >= 0) continue
        if (q.required && !this.isAnswered(q, this.answers[q.id])) {
          msgs.push((q.title ? q.title.replace(/<[^>]+>/g, '').slice(0, 20) : '未命名') + ' 为必填项')
        }
      }
      if (msgs.length) { uni.showModal({ title: '请检查', content: msgs.join('; '), showCancel: false }); return false }
      return true
    },
    async onSubmit() {
      if (!this.validateRequired()) return
      uni.showModal({
        title: '确认提交',
        content: '提交后不可修改',
        success: async (r) => {
          if (!r.confirm) return
          this.submitting = true
          const deviceInfo = this.getDeviceUA()
          try {
            const vr = await surveyApi.validate({ surveyId: this.surveyId, answers: this.answers, device: deviceInfo, deviceId: this.getDeviceId() })
            if (vr.data && !vr.data.valid) {
              const msgs = (vr.data.errors || []).map(e => e.message).join('; ')
              uni.showModal({ title: '请检查', content: msgs, showCancel: false })
              this.submitting = false
              return
            }
          } catch {}
          await this.doSubmit(deviceInfo, false)
        }
      })
    },
    async submitFromBackPrompt() {
      if (this.submitting) return
      this.backLeaveDialogVisible = false
      if (!this.validateRequired()) return
      this.submitting = true
      const deviceInfo = this.getDeviceUA()
      try {
        const vr = await surveyApi.validate({ surveyId: this.surveyId, answers: this.answers, device: deviceInfo, deviceId: this.getDeviceId() })
        if (vr.data && !vr.data.valid) {
          const msgs = (vr.data.errors || []).map(e => e.message).join('; ')
          uni.showModal({ title: '请检查', content: msgs, showCancel: false })
          this.submitting = false
          return
        }
      } catch {}
      await this.doSubmit(deviceInfo, false)
    },
    abandonTimedSurvey() {
      this.abandoning = true
      this.backLeaveDialogVisible = false
      this.stopCountdown()
      this.stopAutoSave(false)
      this.clearStorage()
      uni.navigateBack()
    },
    async forceSubmit() {
      if (this.abandoning || this.submitting || this.submitted) return
      this.submitting = true
      await this.doSubmit(this.getDeviceUA(), true)
    },
    async doSubmit(deviceInfo, isAuto) {
      if (this.abandoning) return
      try {
        const res = await surveyApi.submit({ surveyId: this.surveyId, answers: this.answers, session: this.session, device: deviceInfo, autoSubmit: !!isAuto, deviceId: this.getDeviceId() })
        if (res.code !== 0) { uni.showToast({ title: res.msg || '提交失败', icon: 'none' }); this.submitting = false; return }
        if (this.timer) { clearInterval(this.timer); this.timer = null }
        this.handleSubmitSuccess(res.data)
      } catch (e) { uni.showToast({ title: e.msg || '提交失败', icon: 'none' }); this.submitting = false }
    },

    handleSubmitSuccess(data) {
      this.clearStorage()
      const url = this.settings.redirectUrl
      const content = this.settings.endContent
      if (url) { uni.showToast({ title: '已提交', icon: 'success' }); setTimeout(() => { window.location.href = url }, 500); return }
      if (content) { this.endContent = content; this.submitted = true; setTimeout(() => { this.goBack() }, 2000); return }
      uni.showToast({ title: '已提交', icon: 'success' })
      setTimeout(() => {
        if (this.survey && this.survey.showResult === 1) uni.redirectTo({ url: `/pages/survey/result?surveyId=${this.surveyId}` })
        else uni.navigateBack()
      }, 1000)
    },

    getDeviceUA() {
      if (typeof navigator !== 'undefined' && navigator.userAgent) return navigator.userAgent
      try {
        const sys = uni.getSystemInfoSync()
        const parts = [sys.platform, sys.model, sys.brand, sys.language].filter(Boolean)
        return 'uni-app/' + (sys.appVersion || '') + ' (' + parts.join('; ') + ')'
      } catch (e) {
        return 'uni-app'
      }
    },
    getDeviceId() {
      const key = '_device_id'
      let id = uni.getStorageSync(key)
      if (!id) {
        id = 'd_' + Date.now().toString(36) + '_' + Math.random().toString(36).slice(2, 10)
        uni.setStorageSync(key, id)
      }
      return id
    },
    goRedirect() { if (this.settings.redirectUrl) window.location.href = this.settings.redirectUrl },
    goBack() { uni.navigateBack() }
  }
}
</script>

<style scoped>
.survey { min-height: 100vh; background: #f5f5f5; padding: 24rpx 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; }
.survey--with-progress { padding-top: 72rpx; }
.survey-progress { position: fixed; top: var(--window-top, 0px); left: 0; right: 0; z-index: 1000; background: #fff; border-bottom: 1px solid #e8e8e8; padding: 8rpx 24rpx; }
.survey-progress-inner { max-width: 720px; margin: 0 auto; display: flex; align-items: center; gap: 12rpx; }
.survey-progress-track { flex: 1; height: 6px; background: #e8e8e8; border-radius: 3px; overflow: hidden; }
.survey-progress-fill { height: 100%; background: #3873f6; border-radius: 3px; transition: width .4s ease; }
.survey-progress-label { font-size: 26rpx; color: #909399; white-space: nowrap; }
.survey-card { max-width: 720px; margin: 0 auto; background: #fff; border-radius: 16rpx; box-shadow: 0 2rpx 6rpx rgba(0,0,0,.08); overflow: hidden; }
.survey-card-img { width: 100%; overflow: hidden; line-height: 0; }
.survey-card-img image { width: 100%; height: 200rpx; object-fit: cover; display: block; }
.survey-header { padding: 32rpx 36rpx 20rpx; border-bottom: 1px solid #f0f0f0; }
.survey-title { font-size: 40rpx; font-weight: 600; color: #1a1a1a; line-height: 1.4; display: block; margin-bottom: 10rpx; }
.survey-desc { font-size: 28rpx; color: #666; line-height: 1.6; display: block; margin-bottom: 14rpx; }
.survey-meta { display: flex; gap: 14rpx; flex-wrap: wrap; }
.survey-tag { display: inline-block; padding: 2rpx 16rpx; border-radius: 6rpx; font-size: 24rpx; line-height: 44rpx; background: #f0f2f5; color: #606266; }
.survey-tag--anon { background: #e6f7ff; color: #1890ff; }
.survey-tag--result { background: #f6ffed; color: #52c41a; }
.survey-tag--warn { background: #fff7e6; color: #fa8c16; }
.survey-tag--danger { background: #fff1f0; color: #f5222d; }
.survey-form { width: 100%; }
.survey-form--single { display: flex; flex-direction: column; min-height: calc(100vh - 120rpx); padding-bottom: 120rpx; }
.survey-form-scroll { padding: 28rpx 36rpx; flex: 1; overflow-y: auto; }
.survey-form-scroll--all { padding: 28rpx 36rpx 120rpx; }
.survey-nav { display: flex; align-items: center; justify-content: center; gap: 24rpx; padding: 24rpx 36rpx; padding-bottom: calc(24rpx + env(safe-area-inset-bottom)); background: #fff; border-top: 1px solid #f0f0f0; position: fixed; bottom: 0; left: 0; right: 0; z-index: 100; box-shadow: 0 -2rpx 8rpx rgba(0,0,0,.06); }
.survey-nav-btn { padding: 16rpx 40rpx; border-radius: 8rpx; font-size: 28rpx; background: #f0f2f5; color: #606266; text-align: center; }
.survey-nav-btn--primary { background: #3873f6; color: #fff; }
.survey-nav-btn.disabled { opacity: .4; pointer-events: none; }
.survey-nav-index { font-size: 26rpx; color: #909399; }
.survey-footer { text-align: center; padding: 24rpx 36rpx 48rpx; }
.survey-btn { background: #3873f6; color: #fff; border: none; border-radius: 12rpx; font-size: 30rpx; padding: 20rpx 48rpx; text-align: center; }
.survey-btn--lg { width: 100%; height: 88rpx; line-height: 88rpx; padding: 0; }
.survey-loading { display: flex; align-items: center; justify-content: center; padding: 200rpx 0; }
.spinner { width: 100rpx; height: 80rpx; font-size: 10px; text-align: center; }
.spinner > view { display: inline-block; width: 12rpx; height: 100%; background: #3873f6; margin: 0 4rpx; animation: sk-stretchdelay 1.2s infinite ease-in-out; }
.spinner .rect2 { animation-delay: -1.1s; }
.spinner .rect3 { animation-delay: -1s; }
.spinner .rect4 { animation-delay: -.9s; }
.spinner .rect5 { animation-delay: -.8s; }
@keyframes sk-stretchdelay { 0%, 40%, 100% { transform: scaleY(.4); } 20% { transform: scaleY(1); } }
.survey-error { max-width: 720px; margin: 0 auto; padding: 200rpx 0; text-align: center; color: #909399; font-size: 32rpx; }
.survey-end { max-width: 650px; margin: 0 auto; padding: 200rpx 40rpx; text-align: center; }
.survey-end-card { background: #fff; border-radius: 16rpx; box-shadow: 0 2rpx 6rpx rgba(0,0,0,.08); padding: 120rpx 80rpx; }
.survey-end-title { font-size: 40rpx; font-weight: 400; color: #333; display: block; margin-bottom: 24rpx; }
.survey-end-desc { font-size: 28rpx; color: #666; word-break: break-word; }
.survey-end-action { margin-top: 48rpx; }
.survey-end-action .survey-btn { min-width: 300rpx; }

.survey-sheet-btn { position: fixed; right: 24rpx; bottom: 120rpx; z-index: 900; width: 88rpx; height: 88rpx; border-radius: 50%; background: #3873f6; color: #fff; display: flex; align-items: center; justify-content: center; box-shadow: 0 4rpx 12rpx rgba(56,115,246,.4); }
.survey-sheet-btn-label { font-size: 24rpx; font-weight: 500; }
.survey-sheet-overlay { position: fixed; inset: 0; z-index: 900; background: rgba(0,0,0,.5); }
.survey-sheet-panel { position: fixed; top: 50%; left: 50%; z-index: 901; transform: translate(-50%, -50%); width: 560rpx; max-height: 70vh; background: #fff; border-radius: 16rpx; padding: 32rpx 28rpx 24rpx; display: flex; flex-direction: column; box-shadow: 0 8rpx 32rpx rgba(0,0,0,.15); }
.survey-sheet-title { font-size: 30rpx; font-weight: 600; color: #333; text-align: center; margin-bottom: 20rpx; }
.survey-sheet-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 12rpx; overflow-y: auto; flex: 1; }
.survey-sheet-item { display: flex; align-items: center; justify-content: center; height: 60rpx; border-radius: 8rpx; font-size: 26rpx; color: #606266; background: #f5f6f8; }
.survey-sheet-item--done { background: #eef2ff; color: #3873f6; font-weight: 500; }
.survey-sheet-item--cur { border: 2px solid #3873f6; font-weight: 600; }
.survey-sheet-stat { text-align: center; font-size: 24rpx; color: #909399; padding-top: 16rpx; margin-top: 16rpx; border-top: 1px solid #f0f0f0; }

.survey-leave-mask { position: fixed; inset: 0; z-index: 1300; display: flex; align-items: center; justify-content: center; padding: 40rpx; background: rgba(0,0,0,.42); }
.survey-leave-dialog { width: 620rpx; max-width: 100%; border-radius: 20rpx; background: #fff; padding: 36rpx 32rpx 28rpx; box-shadow: 0 16rpx 48rpx rgba(0,0,0,.18); }
.survey-leave-title { display: block; font-size: 34rpx; font-weight: 600; color: #1f2937; text-align: center; margin-bottom: 18rpx; }
.survey-leave-desc { display: block; font-size: 28rpx; line-height: 1.6; color: #606266; text-align: center; margin-bottom: 28rpx; }
.survey-leave-actions { display: flex; gap: 16rpx; }
.survey-leave-btn { flex: 1; min-width: 0; height: 72rpx; line-height: 72rpx; border-radius: 12rpx; background: #f3f4f6; color: #4b5563; font-size: 26rpx; text-align: center; }
.survey-leave-btn--primary { background: #3873f6; color: #fff; }
.survey-leave-btn--danger { background: #fff1f0; color: #f56c6c; }
</style>
