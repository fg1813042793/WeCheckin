<template>
  <view class="page" v-if="survey">
    <view class="topbar">
      <view class="back" @click="goBack">‹</view>
      <view class="title">{{ survey.title }}</view>
    </view>

    <view class="intro card">
      <view class="intro-title">{{ survey.title }}</view>
      <view class="intro-desc" v-if="survey.description">{{ survey.description }}</view>
      <view class="intro-meta">
        <text class="meta-tag" v-if="survey.anonymous===1">匿名收集</text>
        <text class="meta-tag" v-if="survey.showResult===1">提交后查看结果</text>
        <text class="meta-tag">{{ questions.length }} 道题</text>
      </view>
    </view>

    <view class="q-list">
      <view v-for="(q, i) in questions" :key="q.id" class="q-item card">
        <view class="q-title">
          <text class="q-num">{{ i + 1 }}.</text>
          <text>{{ q.title }}</text>
          <text v-if="q.required" class="q-req">*</text>
        </view>

        <view v-if="q.type === 'radio'" class="opt-list">
          <view v-for="opt in parseOpts(q.options || q.props)" :key="opt.value"
            class="opt" :class="{ active: answers[q.id] === opt.value }"
            @click="setAnswer(q.id, opt.value)">
            <view class="opt-circle">{{ opt.label }}</view>
            <text>{{ opt.labelText }}</text>
          </view>
        </view>

        <view v-else-if="q.type === 'checkbox'" class="opt-list">
          <view v-for="opt in parseOpts(q.options || q.props)" :key="opt.value"
            class="opt" :class="{ active: isChecked(q.id, opt.value) }"
            @click="toggleCheck(q.id, opt.value)">
            <view class="opt-square">{{ isChecked(q.id, opt.value) ? '✓' : '' }}</view>
            <text>{{ opt.labelText }}</text>
          </view>
        </view>

        <view v-else-if="['input','phone','email','idCard','number'].includes(q.type)">
          <input class="q-input" v-model="answers[q.id]" :placeholder="q.placeholder || '请输入'" />
        </view>

        <view v-else-if="q.type === 'textarea'">
          <textarea class="q-textarea" v-model="answers[q.id]" :placeholder="q.placeholder || '请作答'" :rows="4" />
        </view>

        <view v-else-if="q.type === 'select'">
          <picker mode="selector" :range="parseOpts(q.options || q.props).map(o => o.labelText)" @change="(e) => onSelectChange(q.id, parseOpts(q.options || q.props), e)">
            <view class="q-input">{{ answers[q.id] ? getOptLabel(parseOpts(q.options || q.props), answers[q.id]) : '请选择' }}</view>
          </picker>
        </view>

        <view v-else-if="q.type === 'date'">
          <picker mode="date" @change="(e) => setAnswer(q.id, e.detail.value)">
            <view class="q-input">{{ answers[q.id] || '请选择日期' }}</view>
          </picker>
        </view>

        <view v-else>
          <input class="q-input" v-model="answers[q.id]" placeholder="请输入" />
        </view>
      </view>
    </view>

    <view class="footer">
      <button class="btn-submit" :loading="submitting" @click="onSubmit">提交</button>
    </view>
  </view>

  <view class="loading" v-else>
    <text>加载中...</text>
  </view>
</template>

<script>
import { surveyApi } from '../../api/index'

export default {
  data() {
    return {
      survey: null,
      questions: [],
      answers: {},
      submitting: false
    }
  },
  onLoad(query) {
    this.surveyId = query.id
    this.load()
  },
  methods: {
    async load() {
      try {
        const res = await surveyApi.getDetail({ id: this.surveyId })
        this.survey = res.data
        const sch = (res.data && res.data.schema) || { questions: [] }
        this.questions = sch.questions || []
      } catch (e) {
        uni.showToast({ title: e.msg || '加载失败', icon: 'none' })
        setTimeout(() => uni.navigateBack(), 1500)
      }
    },
    parseOpts(opts) {
      if (!opts) return []
      if (typeof opts === 'string') {
        try { opts = JSON.parse(opts) } catch { return [] }
      }
      if (!Array.isArray(opts)) return []
      return opts.map((o, i) => {
        if (typeof o === 'string') return { label: String.fromCharCode(65 + i), value: o, labelText: o }
        return {
          label: o.label || String.fromCharCode(65 + i),
          value: o.value !== undefined ? o.value : o.label,
          labelText: o.label || o.value
        }
      })
    },
    setAnswer(qid, val) { this.answers[qid] = val },
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
      return o ? o.labelText : val
    },
    async onSubmit() {
      // 校验
      try {
        const vr = await surveyApi.validate({
          surveyId: this.surveyId,
          answers: this.answers
        })
        if (vr.data && !vr.data.valid) {
          const msgs = (vr.data.errors || []).map(e => e.message).join('; ')
          uni.showModal({ title: '请检查', content: msgs, showCancel: false })
          return
        }
      } catch (e) { /* allow submit even if validate fails */ }

      uni.showModal({
        title: '确认提交',
        content: '提交后不可修改',
        success: async (r) => {
          if (!r.confirm) return
          this.submitting = true
          try {
            await surveyApi.submit({
              surveyId: this.surveyId,
              answers: this.answers,
              device: 'uni-app'
            })
            uni.showToast({ title: '已提交', icon: 'success' })
            setTimeout(() => {
              if (this.survey && this.survey.showResult === 1) {
                uni.redirectTo({ url: `/pages/survey/result?surveyId=${this.surveyId}` })
              } else {
                uni.navigateBack()
              }
            }, 1000)
          } catch (e) {
            uni.showToast({ title: e.msg || '提交失败', icon: 'none' })
            this.submitting = false
          }
        }
      })
    },
    goBack() { uni.navigateBack() }
  }
}
</script>

<style scoped>
.page { min-height: 100vh; background: #f5f5f5; padding-bottom: 160rpx; }
.topbar { display: flex; align-items: center; height: 90rpx; padding: 0 30rpx; background: #fff; border-bottom: 1rpx solid #f0f0f0; }
.back { font-size: 50rpx; color: #333; width: 60rpx; }
.title { flex: 1; text-align: center; font-size: 30rpx; font-weight: 500; margin-right: 60rpx; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.intro { margin: 20rpx 30rpx; padding: 30rpx; background: linear-gradient(135deg, #fff5f5, #ffe0e0); border-radius: 16rpx; }
.card { box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06); }
.intro-title { font-size: 36rpx; font-weight: 500; color: #333; margin-bottom: 10rpx; }
.intro-desc { font-size: 26rpx; color: #666; line-height: 1.5; margin-bottom: 14rpx; }
.intro-meta { display: flex; gap: 14rpx; flex-wrap: wrap; }
.meta-tag { background: rgba(251, 69, 76, 0.1); color: #fb454c; font-size: 22rpx; padding: 4rpx 14rpx; border-radius: 6rpx; }
.q-list { padding: 0 30rpx 20rpx; }
.q-item { background: #fff; border-radius: 16rpx; padding: 30rpx; margin-bottom: 20rpx; }
.q-title { font-size: 30rpx; color: #333; margin-bottom: 24rpx; line-height: 1.5; }
.q-num { color: #fb454c; font-weight: 500; margin-right: 6rpx; }
.q-req { color: #fb454c; margin-left: 6rpx; }
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
