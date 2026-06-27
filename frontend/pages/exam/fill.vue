<template>
  <view class="exam">
    <view v-if="settings.progressBar && !loading && exam" class="exam-progress">
      <view class="exam-progress-inner">
        <view class="exam-progress-track"><view class="exam-progress-fill" :style="{ width: progressPct + '%' }" /></view>
        <text class="exam-progress-label">{{ answeredCount }}/{{ totalQuestions }}</text>
      </view>
    </view>

    <view v-if="loading" class="exam-loading">
      <view class="spinner"><view v-for="i in 5" :key="i" :class="'rect' + i" /></view>
    </view>

    <view v-else-if="showLogin" class="exam-login">
      <view class="exam-card exam-card--narrow">
        <view class="exam-login-header">
          <text class="exam-login-title">登录后参加考试</text>
          <text class="exam-login-desc" v-if="exam && exam.title">{{ exam.title }}</text>
        </view>
        <view class="exam-login-form">
          <input v-model="loginForm.name" class="login-input" placeholder="用户名 / 手机号" @confirm="doLogin" />
          <input v-model="loginForm.password" class="login-input" type="password" placeholder="请输入密码" @confirm="doLogin" />
          <button class="exam-btn exam-btn--primary exam-btn--lg" :loading="loginLoading" @click="doLogin">登录</button>
        </view>
        <view class="exam-login-footer">
          <text>还没有账号？</text>
          <text class="exam-login-link" @click="goRegister">去注册</text>
        </view>
      </view>
    </view>

    <view v-else-if="error" class="exam-error">{{ error }}</view>

    <view v-else-if="submitted && !showResultView" class="exam-end">
      <view class="exam-end-card">
        <text class="exam-end-title">{{ endContent ? '' : '已交卷，感谢作答' }}</text>
        <view v-if="endContent" class="exam-end-desc"><rich-text :nodes="endContent" /></view>
        <view v-if="result" class="exam-end-result">
          <button v-if="settings.showAnalysis && settings.transcriptVisible !== false" class="exam-btn exam-btn--primary" @click="showResultView = true">查看结果</button>
          <button v-else-if="settings.showAnalysis" class="exam-btn exam-btn--primary" @click="showResultView = true">查看解析</button>
          <button v-else class="exam-btn exam-btn--primary" @click="showResultView = true">查看成绩</button>
          <button v-if="settings.redirectUrl" class="exam-btn" @click="goRedirect">查看结果</button>
        </view>
        <view v-else class="exam-end-action">
          <button class="exam-btn exam-btn--primary" @click="goBack">返回</button>
        </view>
      </view>
    </view>

    <!-- 成绩单 -->
    <view v-else-if="showResultView && result && (settings.transcriptVisible !== false || settings.showAnalysis)" class="result-container">
      <view v-if="settings.transcriptVisible !== false" class="result-header">
        <text class="result-title">{{ exam ? exam.title : '' }}</text>
        <view class="result-score">
          <text class="result-score-num">{{ result.score }}</text>
          <text class="result-score-total">/{{ result.fullScore }}</text>
        </view>
        <text v-if="result.correctCnt !== undefined" class="result-correct">正确 {{ result.correctCnt }}/{{ totalQuestions }}</text>
      </view>
      <scroll-view class="result-scroll" scroll-y>
        <view v-for="(q, i) in questions" :key="q.id" class="q-item" :data-qid="q.id">
          <view class="q-title"><rich-text :nodes="resultFullTitle(q, i)" /></view>
          <view class="q-answer" v-if="result && result.results && result.results[i]">
            <text :style="{ color: result.results[i].correct ? '#67c23a' : '#f56c6c' }">{{ result.results[i].correct ? '✓' : '✗' }}</text>
            <text style="margin-left:8rpx;color:#606266">你的答案：{{ formatResultAnswer(q, answers[q.id]) }}</text>
          </view>
          <view v-if="q.examCorrectAnswer && (!result || !result.results || !result.results[i] || !result.results[i].correct)" class="q-correct">
            <text>✓ 正确答案：</text>
            <rich-text :nodes="displayAnswer(q, q.examCorrectAnswer)" />
          </view>
          <view v-if="q.examAnalysis" class="q-analysis">
            <text>解析：{{ q.examAnalysis }}</text>
          </view>
        </view>
      </scroll-view>
    </view>

    <view v-else-if="exam" class="exam-card">
      <view v-if="headerImage" class="exam-card-img"><image :src="headerImage" mode="aspectFill" /></view>
      <view class="exam-header">
        <text class="exam-title">{{ exam.title }}</text>
        <text v-if="exam.description" class="exam-desc">{{ exam.description }}</text>
        <view class="exam-meta">
          <text v-if="exam.showScore && totalScore" class="exam-tag">满分 {{ totalScore }} 分</text>
          <text class="exam-tag">{{ totalQuestions }} 道题</text>
          <text v-if="examTimeRange" class="exam-tag exam-tag--time">{{ examTimeRange }}</text>
          <text v-if="remaining > 0" class="exam-tag" :class="remaining < 60000 ? 'exam-tag--danger' : 'exam-tag--warn'">⏱ {{ formatRemaining(remaining) }}</text>
        </view>
      </view>

      <view v-if="settings.onePageOneQuestion && currentQuestion" class="exam-form exam-form--single" @touchstart="onSwipeStart" @touchend="onSwipeEnd">
        <view class="exam-form-scroll">
            <view class="q-wrap">
              <QuestionField
                :q="currentQuestion"
                :index="currentQIndex"
                :q-score="exam.showScore && currentQuestion.examScore ? currentQuestion.examScore : 0"
                :value="answers[currentQuestion.id]"
                :file-list="fileLists[currentQuestion.id] || []"
                :show-number="settings.questionNumber !== false"
                @input="(v) => setAnswer(currentQuestion.id, v)"
                @update:fileList="(v) => $set(fileLists, currentQuestion.id, v)"
                @sig-open="isSigOpen = true" @sig-close="isSigOpen = false"
              />
            </view>
          <view v-if="settings.answerVisible && !LAYOUT_TYPES.includes(currentQuestion.type) && (currentQuestion.examCorrectAnswer || currentQuestion.examAnalysis)" class="exam-answer-box">
            <view v-if="currentQuestion.examCorrectAnswer" style="margin-bottom:4rpx">
              <text style="color:#67c23a">✓ 正确答案：</text>
              <rich-text :nodes="displayAnswer(currentQuestion, currentQuestion.examCorrectAnswer)" />
            </view>
            <view v-if="currentQuestion.examAnalysis" style="color:#909399">
              <text style="font-weight:500">解析：</text>
              <text>{{ currentQuestion.examAnalysis }}</text>
            </view>
          </view>
        </view>
        <view class="exam-nav">
          <view :class="['exam-nav-btn', { disabled: currentQIndex <= 0 }]" @click="goPrev">上一题</view>
          <text class="exam-nav-index">{{ currentQIndex + 1 }} / {{ totalQuestions }}</text>
          <view v-if="!isLast" class="exam-nav-btn exam-nav-btn--primary" @click="goNext">下一题</view>
          <view v-else class="exam-nav-btn exam-nav-btn--primary" @click="onSubmit">交卷</view>
        </view>
      </view>

      <view v-else class="exam-form">
        <view class="exam-form-scroll exam-form-scroll--all">
          <view v-for="(q, i) in realQuestions" :key="q.id" class="q-wrap">
            <QuestionField
              :q="q"
              :index="i"
              :q-score="exam.showScore && q.examScore ? q.examScore : 0"
              :value="answers[q.id]"
              :file-list="fileLists[q.id] || []"
              :show-number="settings.questionNumber !== false"
              @input="(v) => setAnswer(q.id, v)"
              @update:fileList="(v) => $set(fileLists, q.id, v)"
              @sig-open="isSigOpen = true" @sig-close="isSigOpen = false"
            />
          </view>
        </view>
        <view class="exam-footer">
          <button class="exam-btn exam-btn--primary exam-btn--lg" :loading="submitting" @click="onSubmit">交卷</button>
        </view>
      </view>
    </view>

    <view v-if="!loading && exam && settings.answerSheetVisible !== false">
      <view class="exam-sheet-btn" :style="sheetBtnStyle" @touchstart="onSheetDragStart" @touchmove.prevent="onSheetDragMove" @touchend="onSheetDragEnd" @click="openSheet">
        <text class="exam-sheet-btn-label">{{ answeredCount }}/{{ totalQuestions }}</text>
      </view>
      <view v-if="showSheet" class="exam-sheet-overlay" @click="showSheet = false" />
      <view v-if="showSheet" class="exam-sheet-panel">
        <text class="exam-sheet-title">答题卡</text>
        <view class="exam-sheet-grid">
          <view v-for="(q, i) in realQuestions" :key="q.id"
            class="exam-sheet-item"
            :class="{
              'exam-sheet-item--done': isAnswered(q, answers[q.id]),
              'exam-sheet-item--cur': settings.onePageOneQuestion && q === currentQuestion
            }"
            @click="jumpToQuestion(q)"
          >{{ i + 1 }}</view>
        </view>
        <view class="exam-sheet-stat">{{ answeredCount }}/{{ totalQuestions }}</view>
      </view>
    </view>
  </view>
</template>

<script>
import { examApi } from '../../api/index'
import QuestionField from '../../components/survey/QuestionField.vue'
import CONFIG from '../../config/index'

const LAYOUT_TYPES = ['description', 'divider', 'pagination']

export default {
  components: { QuestionField },
  data() {
    return {
      exam: null,
      settings: {},
      questions: [],
      answers: {},
      loading: true,
      error: '',
      submitting: false,
      submitted: false,
      showLogin: false,
      loginLoading: false,
      loginForm: { name: '', password: '' },
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
      result: null,
      showResultView: false,
      startTimer: null,
      countdownToStart: 0
    }
  },
  computed: {
    LAYOUT_TYPES() { return LAYOUT_TYPES },
    totalQuestions() { return this.realQuestions.length },
    totalScore() { return (this.questions || []).reduce((s, q) => s + (Number(q.examScore) || 0), 0) },
    answeredCount() { return this.realQuestions.filter(q => this.isAnswered(q, this.answers[q.id])).length },
    progressPct() { return this.totalQuestions ? Math.round(this.answeredCount / this.totalQuestions * 100) : 0 },
    realQuestions() { return (this.questions || []).filter(q => !LAYOUT_TYPES.includes(q.type) && !q.defaultHidden) },
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
    },
    examTimeRange() {
      if (!this.exam) return ''
      const st = Number(this.exam.startTime)
      const et = Number(this.exam.endTime)
      if (!st && !et) return ''
      const fmt = (t) => {
        const d = new Date(t)
        return `${d.getFullYear()}/${d.getMonth() + 1}/${d.getDate()} ${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`
      }
      if (st && et) return `${fmt(st)} ~ ${fmt(et)}`
      if (st) return `开始：${fmt(st)}`
      return `截止：${fmt(et)}`
    }
  },
  onLoad(query) {
    this.examId = Number(query.id)
    this.load()
  },
  onUnload() {
    if (this.timer) { clearInterval(this.timer); this.timer = null }
    if (this.startTimer) { clearInterval(this.startTimer); this.startTimer = null }
    if (this.autoSaveTimer) { clearTimeout(this.autoSaveTimer); this.saveToStorage() }
  },
  methods: {
    async load() {
      if (!this.examId) { this.error = '参数错误'; this.loading = false; return }
      // 检查结果模式
      const resultSession = this.$route?.query?.exSession
      if (resultSession) {
        try {
          const res = await examApi.getDetail({ session: resultSession })
          if (res.code === 0 && res.data) {
            this.exam = res.data.exam
            this.questions = res.data.questions || []
            this.answers = res.data.answers || {}
            this.settings = res.data.settings || {}
            if (res.data.record) {
              const d = res.data
              this.result = {
                score: d.record?.score,
                fullScore: d.record?.totalScore,
                correctCnt: d.results?.filter(r => r.correct).length,
                results: d.results
              }
            }
            this.submitted = true
            this.showResultView = true
            this.loading = false
            return
          }
        } catch {}
      }

      try {
        this.session = this.loadSession() || ''
        const res = await examApi.getDetail({ id: this.examId, session: this.session })
        if (res.code !== 0) {
          if (res.msg === '请先登录') { this.showLogin = true; this.loading = false; return }
          if (res.msg === '考试未开始' && res.data) {
            const startMs = Number(res.data.startTime)
            if (startMs > 0) {
              this.error = '考试未开始'
              const tick = () => {
                const left = startMs - Date.now()
                if (left <= 0) { if (this.startTimer) clearInterval(this.startTimer); this.load() }
                this.countdownToStart = Math.max(0, left)
              }
              tick()
              this.startTimer = setInterval(tick, 1000)
              this.loading = false
              return
            }
          }
          this.error = res.msg || '加载失败'; this.loading = false; return
        }
        this.exam = res.data
        if (this.exam.endTime > 0 && Date.now() > this.exam.endTime) {
          this.error = '考试已结束'; this.loading = false; return
        }
        const rawSettings = res.data?.settings
        this.settings = rawSettings ? (typeof rawSettings === 'string' ? JSON.parse(rawSettings) : rawSettings) : {}
        const raw = res.data?.schema
        const sch = raw ? (typeof raw === 'string' ? JSON.parse(raw) : raw) : { questions: [] }
        this.questions = sch.questions || []

        if (this.settings.randomOrder) {
          const layout = this.questions.filter(q => LAYOUT_TYPES.includes(q.type))
          const nonLayout = this.questions.filter(q => !LAYOUT_TYPES.includes(q.type))
          for (let i = nonLayout.length - 1; i > 0; i--) {
            const j = Math.floor(Math.random() * (i + 1));
            [nonLayout[i], nonLayout[j]] = [nonLayout[j], nonLayout[i]]
          }
          this.questions = [...layout, ...nonLayout]
        }

        const init = {}
        this.questions.forEach(q => { init[q.id] = this.getInitVal(q) })
        if (this.settings.autoSave) {
          const saved = this.loadFromStorage()
          if (saved) {
            if (saved.answers) Object.keys(saved.answers).forEach(k => { if (k in init) init[k] = saved.answers[k] })
          }
        }
        this.answers = init
        if (res.data?.session) { this.session = res.data.session; this.saveSession(this.session) }
        if (res.data?.startAt) this.startAt = res.data.startAt
        this.startCountdown()
      } catch (e) {
        this.error = e.msg || '加载失败'
        uni.showToast({ title: '加载失败', icon: 'none' })
      } finally { this.loading = false }
    },

    async doLogin() {
      this.loginLoading = true
      try {
        const res = await this.apiPost('/passport/login_pwd', { name: this.loginForm.name, pwd: this.loginForm.password })
        if (res.code === 0) {
          uni.setStorageSync('token', res.data.token)
          if (res.data.userInfo) uni.setStorageSync('userInfo', res.data.userInfo)
          this.showLogin = false; this.loading = true; this.load()
        } else {
          uni.showToast({ title: res.msg || '登录失败', icon: 'none' })
        }
      } catch { uni.showToast({ title: '登录失败', icon: 'none' }) }
      finally { this.loginLoading = false }
    },

    async apiPost(path, data) {
      const token = uni.getStorageSync('token')
      const BASE_URL = CONFIG.BASE_URL
      return new Promise((resolve) => {
        uni.request({
          url: BASE_URL + path,
          method: 'POST',
          data: JSON.stringify(data),
          header: { 'Content-Type': 'application/json', Authorization: token || '' },
          success: (res) => resolve(res.data),
          fail: () => resolve({ code: 1, msg: '网络错误' })
        })
      })
    },

    getInitVal(q) {
      const type = q.type
      if (type === 'checkbox') return []
      if (type === 'switch') return false
      if (type === 'rating') return 0
      if (type === 'nps') return 0
      if (type === 'dateRange') return ['', '']
      if (['matrixRadio','matrixCheckbox','matrixFillBlank'].includes(type)) return {}
      if (type === 'matrixAuto') return []
      if (['multiInput','hInput'].includes(type)) return (q.props?.fields||[]).map(() => '')
      if (['user','dept'].includes(type)) return q.multiple ? [] : ''
      if (['cascade'].includes(type)) return []
      if (type === 'picker') return ''
      if (type === 'file') return []
      return ''
    },

    isAnswered(q, val) {
      const type = q.type
      if (val === undefined || val === null) return false
      if (type === 'checkbox') return Array.isArray(val) && val.length > 0
      if (type === 'file') return typeof val === 'string' ? !!val : (Array.isArray(val) && val.length > 0)
      if (['rating', 'nps'].includes(type)) return val > 0
      if (['multiInput', 'hInput'].includes(type)) return Array.isArray(val) && val.some(v => !!v)
      if (type === 'matrixRadio') return Object.keys(val).length > 0
      if (type === 'matrixCheckbox') return Object.values(val).some(v => Array.isArray(v) && v.length > 0)
      if (type === 'matrixFillBlank') return Object.values(val).some(v => !!v)
      if (type === 'matrixAuto') return Array.isArray(val) && val.some(row => row.some(v => !!v))
      if (type === 'dateRange') return Array.isArray(val) && val.some(v => !!v)
      if (type === 'switch') return val === true
      if (type === 'cascade') return Array.isArray(val) && val.length > 0
      if (['user', 'dept'].includes(type)) return q.multiple ? (Array.isArray(val) && val.length > 0) : !!val
      if (type === 'picker') return !!val
      return !!val
    },

    setAnswer(qid, val) {
      this.$set(this.answers, qid, val)
      this.autoSave()
    },

    autoSave() {
      if (!this.settings.autoSave) return
      if (this.autoSaveTimer) { clearTimeout(this.autoSaveTimer); this.autoSaveTimer = null }
      this.autoSaveTimer = setTimeout(() => {
        this.saveToStorage()
        this.autoSaveTimer = null
      }, 800)
    },
    saveToStorage() {
      try { uni.setStorageSync('exam_' + this.examId, JSON.stringify({ answers: this.answers })) } catch (e) {}
    },
    loadFromStorage() {
      try { const raw = uni.getStorageSync('exam_' + this.examId); return raw ? JSON.parse(raw) : null } catch (e) { return null }
    },
    clearStorage() {
      try { uni.removeStorageSync('exam_' + this.examId) } catch (e) {}
      try { uni.removeStorageSync('exam_session_' + this.examId) } catch (e) {}
    },
    saveSession(s) { try { uni.setStorageSync('exam_session_' + this.examId, s) } catch (e) {} },
    loadSession() { try { return uni.getStorageSync('exam_session_' + this.examId) } catch (e) { return '' } },

    goNext() { if (this.currentQIndex < this.totalQuestions - 1) this.currentQIndex++ },
    goPrev() { if (this.currentQIndex > 0) this.currentQIndex-- },

    formatRemaining(ms) {
      if (ms <= 0) return '已超时'
      const t = Math.ceil(ms / 1000)
      return `${Math.floor(t / 60)}:${(t % 60).toString().padStart(2, '0')}`
    },

    startCountdown() {
      if (this.timer) { clearInterval(this.timer); this.timer = null }
      if (!this.startAt) return
      const limit = Number(this.settings.timeLimit)
      const maxSubmit = Number(this.settings.maxSubmitMinutes)
      const endTime = Number(this.exam?.endTime)
      const tick = () => {
        const now = Date.now()
        let deadline = Infinity
        if (limit > 0) deadline = Math.min(deadline, this.startAt + limit * 60 * 1000)
        if (maxSubmit > 0) deadline = Math.min(deadline, this.startAt + maxSubmit * 60 * 1000)
        if (endTime > 0) deadline = Math.min(deadline, endTime)
        if (deadline === Infinity) { this.remaining = 0; return }
        const left = deadline - now
        this.remaining = Math.max(0, left)
        if (left <= 0) {
          clearInterval(this.timer); this.timer = null
          uni.showToast({ title: '作答时间已到，自动交卷', icon: 'none' })
          this.forceSubmit()
        }
      }
      tick()
      this.timer = setInterval(tick, 1000)
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
    onSwipeEnd(e) {
      if (!this.settings.onePageOneQuestion) return
      if (this.isSigOpen) return
      const dx = e.changedTouches[0].clientX - this.swipeStartX
      const dy = e.changedTouches[0].clientY - this.swipeStartY
      if (Math.abs(dx) < 30 || Math.abs(dy) > Math.abs(dx)) return
      if (dx < 0) this.goNext()
      else this.goPrev()
    },

    displayAnswer(q, val) {
      if (val === undefined || val === null) return ''
      if (q.type === 'judge') return val === 'true' ? '对' : '错'
      if (Array.isArray(val)) {
        return val.map(v => {
          const opts = q.props?.options || []
          const opt = opts.find(o => String(o.value) === String(v))
          return opt?.label || String(v)
        }).join('、')
      }
      const opts = q.props?.options || []
      const opt = opts.find(o => String(o.value) === String(val))
      return opt?.label || String(val)
    },

    formatResultAnswer(q, val) {
      if (val === undefined || val === null) return '-'
      if (typeof val === 'string') return val
      if (Array.isArray(val)) return val.join(', ')
      return String(val)
    },

    async onSubmit() {
      // 最短交卷时间检查
      const minSubmit = Number(this.settings.minSubmitMinutes)
      if (minSubmit > 0 && this.startAt > 0) {
        const elapsed = (Date.now() - this.startAt) / 60000
        if (elapsed < minSubmit) {
          uni.showToast({ title: `距最短交卷时间还有 ${Math.ceil(minSubmit - elapsed)} 分钟`, icon: 'none' })
          return
        }
      }
      uni.showModal({
        title: '确认交卷',
        content: '交卷后不可修改',
        success: async (r) => {
          if (!r.confirm) return
          this.submitting = true
          const deviceInfo = this.getDeviceUA()
          try {
            const vr = await examApi.validate({ examId: this.examId, answers: this.answers, device: deviceInfo, deviceId: this.getDeviceId() })
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
    async forceSubmit() {
      this.submitting = true
      await this.doSubmit(this.getDeviceUA(), true)
    },
    async doSubmit(deviceInfo, isAuto) {
      try {
        const res = await examApi.submit({ examId: this.examId, answers: this.answers, session: this.session, device: deviceInfo, autoSubmit: !!isAuto, deviceId: this.getDeviceId() })
        if (res.code !== 0) { uni.showToast({ title: res.msg || '交卷失败', icon: 'none' }); this.submitting = false; return }
        if (this.timer) { clearInterval(this.timer); this.timer = null }
        this.handleSubmitSuccess(res.data)
      } catch (e) { uni.showToast({ title: e.msg || '交卷失败', icon: 'none' }); this.submitting = false }
    },

    handleSubmitSuccess(data) {
      this.clearStorage()
      if (data?.record) {
        this.result = {
          score: data.record?.score,
          fullScore: data.record?.totalScore,
          correctCnt: data.results?.filter(r => r.correct).length,
          results: data.results
        }
      }
      this.submitted = true
      if (this.settings.transcriptVisible !== false && !this.settings.showAnalysis) {
        this.showResultView = true
        return
      }
      if (!this.settings.transcriptVisible && !this.settings.showAnalysis) {
        const url = this.settings?.redirectUrl
        const content = this.settings?.endContent
        if (url) { uni.showToast({ title: '已交卷', icon: 'success' }); setTimeout(() => { window.location.href = url }, 500); return }
        if (content) { this.endContent = content; this.submitted = true; setTimeout(() => { this.goBack() }, 2000); return }
      }
      uni.showToast({ title: '已交卷', icon: 'success' })
      setTimeout(() => { uni.navigateBack() }, 1000)
    },

    getDeviceUA() {
      if (typeof navigator !== 'undefined' && navigator.userAgent) return navigator.userAgent
      try {
        const sys = uni.getSystemInfoSync()
        const parts = [sys.platform, sys.model, sys.brand, sys.language].filter(Boolean)
        return 'uni-app/' + (sys.appVersion || '') + ' (' + parts.join('; ') + ')'
      } catch (e) { return 'uni-app' }
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
    goBack() { uni.navigateBack() },
    goRegister() {
      uni.navigateTo({ url: '/pages/login/login' })
    },
    unwrapOuterP(html) {
      if (!html) return ''
      const trimmed = html.trim()
      const match = trimmed.match(/^<p([^>]*)>([\s\S]*)<\/p>$/)
      if (!match) return html
      const attrs = match[1]
      const content = match[2]
      if (/\bql-align-\w+\b/.test(attrs)) {
        return `<span${attrs} style="display:inline-block;width:100%;text-align:inherit">${content}</span>`
      }
      return content
    },
    resultFullTitle(q, i) {
      let html = `<span style="color:#3873f6;font-weight:600;margin-right:4rpx">${i + 1}.</span>`
      html += this.unwrapOuterP(q.title || '')
      return html
    }
  }
}
</script>

<style scoped>
.exam { min-height: 100vh; background: #f5f5f5; padding: 24rpx 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; }
.exam-loading { display: flex; align-items: center; justify-content: center; padding: 200rpx 0; }
.spinner { width: 100rpx; height: 80rpx; font-size: 10px; text-align: center; }
.spinner > view { display: inline-block; width: 12rpx; height: 100%; background: #3873f6; margin: 0 4rpx; animation: sk-stretchdelay 1.2s infinite ease-in-out; }
.spinner .rect2 { animation-delay: -1.1s; }
.spinner .rect3 { animation-delay: -1s; }
.spinner .rect4 { animation-delay: -.9s; }
.spinner .rect5 { animation-delay: -.8s; }
@keyframes sk-stretchdelay { 0%, 40%, 100% { transform: scaleY(.4); } 20% { transform: scaleY(1); } }

.exam-progress { position: fixed; top: 0; left: 0; right: 0; z-index: 1000; background: #fff; border-bottom: 1px solid #e8e8e8; padding: 8rpx 24rpx; }
/* #ifdef H5 */
.exam-progress { top: var(--window-top); }
/* #endif */
.exam-progress-inner { max-width: 720px; margin: 0 auto; display: flex; align-items: center; gap: 12rpx; }
.exam-progress-track { flex: 1; height: 6px; background: #e8e8e8; border-radius: 3px; overflow: hidden; }
.exam-progress-fill { height: 100%; background: #3873f6; border-radius: 3px; transition: width .4s ease; }
.exam-progress-label { font-size: 26rpx; color: #909399; white-space: nowrap; }

.exam-card { max-width: 720px; margin: 0 auto; background: #fff; border-radius: 16rpx; box-shadow: 0 2rpx 6rpx rgba(0,0,0,.08); overflow: hidden; }
.exam-card--narrow { width: 98%; }
.exam-card-img { width: 100%; overflow: hidden; line-height: 0; }
.exam-card-img image { width: 100%; height: 200rpx; object-fit: cover; display: block; }
.exam-header { padding: 32rpx 36rpx 20rpx; border-bottom: 1px solid #f0f0f0; }
.exam-title { font-size: 40rpx; font-weight: 600; color: #1a1a1a; line-height: 1.4; display: block; margin-bottom: 10rpx; }
.exam-desc { font-size: 28rpx; color: #666; line-height: 1.6; display: block; margin-bottom: 14rpx; }
.exam-meta { display: flex; gap: 14rpx; flex-wrap: wrap; }
.exam-tag { display: inline-block; padding: 2rpx 16rpx; border-radius: 6rpx; font-size: 24rpx; line-height: 44rpx; background: #f0f2f5; color: #606266; }
.exam-tag--time { background: #e6f7ff; color: #1890ff; }
.exam-tag--warn { background: #fff7e6; color: #fa8c16; }
.exam-tag--danger { background: #fff1f0; color: #f5222d; }

.exam-form { width: 100%; }
.exam-form--single { display: flex; flex-direction: column; min-height: calc(100vh - 120rpx); padding-bottom: 120rpx; }
.exam-form-scroll { padding: 28rpx 36rpx; flex: 1; overflow-y: auto; }
.exam-form-scroll--all { padding: 28rpx 36rpx 120rpx; }

.exam-nav { display: flex; align-items: center; justify-content: center; gap: 24rpx; padding: 24rpx 36rpx; padding-bottom: calc(24rpx + env(safe-area-inset-bottom)); background: #fff; border-top: 1px solid #f0f0f0; position: fixed; bottom: 0; left: 0; right: 0; z-index: 100; box-shadow: 0 -2rpx 8rpx rgba(0,0,0,.06); }
.exam-nav-btn { padding: 16rpx 40rpx; border-radius: 8rpx; font-size: 28rpx; background: #f0f2f5; color: #606266; text-align: center; }
.exam-nav-btn--primary { background: #3873f6; color: #fff; }
.exam-nav-btn.disabled { opacity: .4; pointer-events: none; }
.exam-nav-index { font-size: 26rpx; color: #909399; }

.exam-footer { text-align: center; padding: 24rpx 36rpx 48rpx; }
.exam-btn { background: #3873f6; color: #fff; border: none; border-radius: 12rpx; font-size: 30rpx; padding: 20rpx 48rpx; text-align: center; }
.exam-btn--lg { width: 100%; height: 88rpx; line-height: 88rpx; padding: 0; }

.exam-error { max-width: 720px; margin: 0 auto; padding: 200rpx 0; text-align: center; color: #909399; font-size: 32rpx; }

.exam-login { max-width: 720px; margin: 0 auto; padding: 100rpx 0; }
.exam-login-header { text-align: center; margin-bottom: 40rpx; }
.exam-login-title { font-size: 36rpx; font-weight: 600; color: #333; display: block; margin-bottom: 12rpx; }
.exam-login-desc { font-size: 28rpx; color: #909399; display: block; }
.exam-login-form { padding: 0 30rpx; }
.login-input { width: 100%; height: 88rpx; border: 1px solid #dcdfe6; border-radius: 12rpx; font-size: 28rpx; padding: 0 24rpx; margin-bottom: 24rpx; box-sizing: border-box; background: #fff; }
.exam-login-footer { text-align: center; margin-top: 16rpx; font-size: 26rpx; color: #909399; }
.exam-login-link { color: #3873f6; margin-left: 8rpx; }

.exam-end { max-width: 650px; margin: 0 auto; padding: 200rpx 40rpx; text-align: center; }
.exam-end-card { background: #fff; border-radius: 16rpx; box-shadow: 0 2rpx 6rpx rgba(0,0,0,.08); padding: 120rpx 80rpx; }
.exam-end-title { font-size: 40rpx; font-weight: 400; color: #333; display: block; margin-bottom: 24rpx; }
.exam-end-desc { font-size: 28rpx; color: #666; word-break: break-word; }
.exam-end-result { margin-top: 48rpx; display: flex; flex-direction: column; gap: 20rpx; align-items: center; }
.exam-end-action { margin-top: 48rpx; }
.exam-end-action .exam-btn { min-width: 300rpx; }

.exam-answer-box { margin: 12rpx 0 0; font-size: 26rpx; border: 1px dashed #ccc; border-radius: 8rpx; padding: 16rpx 20rpx; background: #fafafa; }
.q-wrap { position: relative; }

.exam-sheet-btn { position: fixed; right: 24rpx; bottom: 120rpx; z-index: 900; width: 88rpx; height: 88rpx; border-radius: 50%; background: #3873f6; color: #fff; display: flex; align-items: center; justify-content: center; box-shadow: 0 4rpx 12rpx rgba(56,115,246,.4); }
.exam-sheet-btn-label { font-size: 24rpx; font-weight: 500; }
.exam-sheet-overlay { position: fixed; inset: 0; z-index: 900; background: rgba(0,0,0,.5); }
.exam-sheet-panel { position: fixed; top: 50%; left: 50%; z-index: 901; transform: translate(-50%, -50%); width: 560rpx; max-height: 70vh; background: #fff; border-radius: 16rpx; padding: 32rpx 28rpx 24rpx; display: flex; flex-direction: column; box-shadow: 0 8rpx 32rpx rgba(0,0,0,.15); }
.exam-sheet-title { font-size: 30rpx; font-weight: 600; color: #333; text-align: center; margin-bottom: 20rpx; }
.exam-sheet-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 12rpx; overflow-y: auto; flex: 1; }
.exam-sheet-item { display: flex; align-items: center; justify-content: center; height: 60rpx; border-radius: 8rpx; font-size: 26rpx; color: #606266; background: #f5f6f8; }
.exam-sheet-item--done { background: #eef2ff; color: #3873f6; font-weight: 500; }
.exam-sheet-item--cur { border: 2px solid #3873f6; font-weight: 600; }
.exam-sheet-stat { text-align: center; font-size: 24rpx; color: #909399; padding-top: 16rpx; margin-top: 16rpx; border-top: 1px solid #f0f0f0; }

/* 成绩单 */
.result-container { max-width: 720px; margin: 0 auto; background: #fff; min-height: 100vh; }
.result-header { text-align: center; padding: 40rpx 30rpx 24rpx; border-bottom: 1px solid #f0f0f0; }
.result-title { font-size: 36rpx; font-weight: 600; color: #333; display: block; margin-bottom: 16rpx; }
.result-score { display: flex; align-items: baseline; justify-content: center; gap: 4rpx; }
.result-score-num { font-size: 64rpx; font-weight: 700; color: #3873f6; }
.result-score-total { font-size: 32rpx; color: #909399; }
.result-correct { font-size: 26rpx; color: #67c23a; display: block; margin-top: 8rpx; }
.result-scroll { padding: 16rpx 0; max-height: calc(100vh - 300rpx); }
.result-scroll .q-item { padding: 24rpx 36rpx; border-bottom: 1px solid #f0f0f0; }
.result-scroll .q-title { font-size: 30rpx; color: #303133; font-weight: 500; margin-bottom: 8rpx; }
.result-scroll .q-title :deep(img) { max-width: 100%; height: auto; }
.result-scroll .q-answer { font-size: 26rpx; color: #606266; }
.result-scroll .q-correct { margin-top: 8rpx; font-size: 26rpx; color: #67c23a; }
.result-scroll .q-analysis { margin-top: 8rpx; font-size: 26rpx; color: #909399; background: #f5f7fa; padding: 12rpx 16rpx; border-radius: 8rpx; }
</style>
