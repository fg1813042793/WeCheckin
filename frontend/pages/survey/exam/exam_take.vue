<template>
  <view class="page" v-if="loaded">
    <view class="topbar">
      <view class="back" @click="confirmExit">‹</view>
      <view class="title">{{ paper.title || '考试中' }}</view>
      <view class="timer" :class="{ danger: remain < 60 }">{{ formatRemain }}</view>
    </view>

    <view class="info">
      <text>总分: {{ paper.totalScore }} | 题目: {{ questions.length }} | 进度: {{ answeredCount }}/{{ questions.length }}</text>
    </view>

    <view class="q-list">
      <view v-for="(q, i) in questions" :key="q.id" class="q-item card">
        <view class="q-title">
          <text class="q-num">{{ i + 1 }}.</text>
          <text>{{ q.title }}</text>
          <text class="q-score">({{ q.score }}分)</text>
        </view>

        <view v-if="q.type === 'radio'" class="opt-list">
          <view v-for="opt in parseOpts(q.options)" :key="opt.value"
            class="opt" :class="{ active: answers[q.id] === opt.value }"
            @click="setAnswer(q.id, opt.value)">
            <view class="opt-circle">{{ opt.label }}</view>
            <text>{{ opt.label }}. {{ opt.labelText || opt.label }}</text>
          </view>
        </view>

        <view v-else-if="q.type === 'checkbox'" class="opt-list">
          <view v-for="opt in parseOpts(q.options)" :key="opt.value"
            class="opt" :class="{ active: isChecked(q.id, opt.value) }"
            @click="toggleCheck(q.id, opt.value)">
            <view class="opt-square">{{ isChecked(q.id, opt.value) ? '✓' : '' }}</view>
            <text>{{ opt.label }}. {{ opt.labelText || opt.label }}</text>
          </view>
        </view>

        <view v-else-if="q.type === 'input' || q.type === 'phone' || q.type === 'email' || q.type === 'idCard' || q.type === 'number'">
          <input class="q-input" v-model="answers[q.id]" :placeholder="'请输入' + (q.title || q.type)" />
        </view>

        <view v-else-if="q.type === 'textarea'">
          <textarea class="q-textarea" v-model="answers[q.id]" placeholder="请作答" :rows="4" />
        </view>

        <view v-else-if="q.type === 'select'">
          <picker mode="selector" :range="parseOpts(q.options).map(o => o.labelText || o.label)" @change="(e) => onSelectChange(q.id, parseOpts(q.options), e)">
            <view class="q-input">{{ answers[q.id] ? getOptLabel(parseOpts(q.options), answers[q.id]) : '请选择' }}</view>
          </picker>
        </view>

        <view v-else-if="q.type === 'date'">
          <picker mode="date" @change="(e) => setAnswer(q.id, e.detail.value)">
            <view class="q-input">{{ answers[q.id] || '请选择日期' }}</view>
          </picker>
        </view>

        <view v-else>
          <input class="q-input" v-model="answers[q.id]" :placeholder="'请输入'" />
        </view>
      </view>
    </view>

    <view class="footer">
      <button class="btn-submit" :loading="submitting" @click="onSubmit">交卷</button>
    </view>
  </view>

  <view class="loading" v-else>
    <text>加载中...</text>
  </view>
</template>

<script>
import { examApi } from '../../api/index'

export default {
  data() {
    return {
      loaded: false,
      exam: {},
      paper: {},
      record: {},
      questions: [],
      answers: {},
      remain: 0,
      timer: null,
      submitting: false
    }
  },
  computed: {
    formatRemain() {
      const m = Math.floor(this.remain / 60)
      const s = this.remain % 60
      return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
    },
    answeredCount() {
      return Object.values(this.answers).filter(v => {
        if (v == null) return false
        if (typeof v === 'string') return v.trim() !== ''
        if (Array.isArray(v)) return v.length > 0
        return true
      }).length
    }
  },
  onUnload() {
    if (this.timer) clearInterval(this.timer)
  },
  onLoad(query) {
    this.examId = query.examId
    this.start()
  },
  methods: {
    async start() {
      try {
        const res = await examApi.start({ examId: this.examId })
        const d = res.data
        this.record = d.record
        this.paper = d.paper
        this.exam = d.exam
        this.questions = d.questions || []
        this.answers = d.answers || {}
        this.loaded = true
        // 启动倒计时
        if (d.exam && d.exam.duration > 0) {
          this.remain = d.exam.duration * 60
          this.timer = setInterval(() => {
            this.remain--
            if (this.remain <= 0) {
              clearInterval(this.timer)
              this.onSubmit()
            }
          }, 1000)
        }
      } catch (e) {
        uni.showToast({ title: e.msg || '加载失败', icon: 'none' })
        setTimeout(() => uni.navigateBack(), 1500)
      }
    },
    parseOpts(options) {
      if (!options) return []
      if (typeof options === 'string') {
        try { options = JSON.parse(options) } catch { return [] }
      }
      if (!Array.isArray(options)) return []
      return options.map((o, i) => {
        if (typeof o === 'string') return { label: String.fromCharCode(65 + i), value: o, labelText: o }
        return { label: o.label || String.fromCharCode(65 + i), value: o.value || o.label, labelText: o.label || o.value }
      })
    },
    setAnswer(qid, val) {
      this.answers[qid] = val
    },
    isChecked(qid, val) {
      const v = this.answers[qid]
      return Array.isArray(v) ? v.includes(val) : false
    },
    toggleCheck(qid, val) {
      if (!Array.isArray(this.answers[qid])) this.answers[qid] = []
      const arr = this.answers[qid]
      const i = arr.indexOf(val)
      if (i >= 0) arr.splice(i, 1)
      else arr.push(val)
    },
    onSelectChange(qid, opts, e) {
      const i = Number(e.detail.value)
      this.answers[qid] = opts[i] ? opts[i].value : ''
    },
    getOptLabel(opts, val) {
      const o = opts.find(x => x.value === val)
      return o ? (o.labelText || o.label) : val
    },
    async onSubmit() {
      if (this.submitting) return
      uni.showModal({
        title: '提示',
        content: '确认交卷？',
        success: async (r) => {
          if (!r.confirm) return
          this.submitting = true
          if (this.timer) clearInterval(this.timer)
          try {
            const res = await examApi.submit({
              recordId: this.record.id,
              answers: JSON.stringify(this.answers)
            })
            uni.showToast({ title: '已交卷', icon: 'success' })
            setTimeout(() => {
              uni.redirectTo({ url: `/pages/survey/exam/exam_result?id=${this.record.id}` })
            }, 1000)
          } catch (e) {
            uni.showToast({ title: e.msg || '交卷失败', icon: 'none' })
            this.submitting = false
          }
        }
      })
    },
    confirmExit() {
      uni.showModal({
        title: '提示',
        content: '退出后当前未提交答案将丢失，确认退出？',
        success: (r) => {
          if (r.confirm) {
            if (this.timer) clearInterval(this.timer)
            uni.navigateBack()
          }
        }
      })
    }
  }
}
</script>

<style scoped>
.page { min-height: 100vh; background: #f5f5f5; padding-bottom: 160rpx; }
.topbar { display: flex; align-items: center; height: 90rpx; padding: 0 30rpx; background: #fff; border-bottom: 1rpx solid #f0f0f0; }
.back { font-size: 50rpx; color: #333; width: 60rpx; }
.title { flex: 1; text-align: center; font-size: 30rpx; font-weight: 500; }
.timer { font-size: 28rpx; color: #4a7af0; font-weight: 500; min-width: 100rpx; text-align: right; }
.timer.danger { color: #fb454c; }
.info { padding: 16rpx 30rpx; font-size: 24rpx; color: #888; background: #fff; border-bottom: 1rpx solid #f0f0f0; }
.q-list { padding: 20rpx 30rpx; }
.q-item { background: #fff; border-radius: 16rpx; padding: 30rpx; margin-bottom: 20rpx; }
.card { box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06); }
.q-title { font-size: 30rpx; color: #333; margin-bottom: 24rpx; line-height: 1.5; }
.q-num { color: #fb454c; font-weight: 500; margin-right: 6rpx; }
.q-score { color: #888; font-size: 24rpx; margin-left: 8rpx; }
.opt-list { display: flex; flex-direction: column; gap: 16rpx; }
.opt { display: flex; align-items: center; padding: 18rpx 20rpx; background: #f7f7f7; border-radius: 12rpx; border: 2rpx solid transparent; }
.opt.active { background: #fff5f5; border-color: #fb454c; }
.opt-circle { width: 44rpx; height: 44rpx; border-radius: 50%; background: #fff; border: 2rpx solid #ddd; text-align: center; line-height: 40rpx; font-size: 24rpx; margin-right: 16rpx; flex-shrink: 0; }
.opt.active .opt-circle { background: #fb454c; border-color: #fb454c; color: #fff; }
.opt-square { width: 44rpx; height: 44rpx; border-radius: 8rpx; background: #fff; border: 2rpx solid #ddd; text-align: center; line-height: 40rpx; font-size: 28rpx; color: #fb454c; margin-right: 16rpx; flex-shrink: 0; }
.q-input { background: #f7f7f7; border-radius: 12rpx; padding: 18rpx 20rpx; font-size: 28rpx; min-height: 40rpx; }
.q-textarea { background: #f7f7f7; border-radius: 12rpx; padding: 18rpx 20rpx; font-size: 28rpx; width: 100%; box-sizing: border-box; min-height: 200rpx; }
.footer { position: fixed; bottom: 0; left: 0; right: 0; padding: 20rpx 30rpx; background: #fff; border-top: 1rpx solid #f0f0f0; }
.btn-submit { background: linear-gradient(90deg, #fb454c, #ff6b6b); color: #fff; border-radius: 50rpx; font-size: 30rpx; height: 88rpx; line-height: 88rpx; }
.loading { padding: 200rpx 0; text-align: center; color: #aaa; }
</style>
