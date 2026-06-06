<template>
  <view class="page" v-if="loaded">
    <view class="topbar">
      <view class="back" @click="goBack">‹</view>
      <view class="title">考试结果</view>
    </view>

    <view class="score-card">
      <view class="score-num">
        <text class="big">{{ record.score }}</text>
        <text class="total">/{{ record.totalScore }}</text>
      </view>
      <view class="score-status" :class="record.pass===1 ? 'pass' : 'fail'">
        {{ record.pass===1 ? '通过' : '未通过' }}
      </view>
      <view class="score-info">
        <text>用时: {{ timeSpent }}</text>
        <text> | 提交: {{ formatTime(record.submitTime) }}</text>
      </view>
    </view>

    <view class="result-list">
      <view v-for="(r, i) in results" :key="r.questionId" class="r-item card">
        <view class="r-header">
          <text class="r-num">{{ i + 1 }}.</text>
          <text class="r-status" :class="resultClass(r)">{{ resultText(r) }}</text>
          <text class="r-score">{{ r.gotScore }}/{{ r.fullScore }}</text>
        </view>
        <view v-if="r.reason" class="r-reason">{{ r.reason }}</view>
        <view v-if="questionMap[r.questionId] && questionMap[r.questionId].answer" class="r-answer">
          <text class="label">正确答案: </text>
          <text>{{ questionMap[r.questionId].answer }}</text>
        </view>
        <view v-if="answers[r.questionId] !== undefined" class="r-my">
          <text class="label">我的答案: </text>
          <text>{{ formatAnswer(answers[r.questionId]) }}</text>
        </view>
      </view>
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
      record: {},
      results: [],
      answers: {},
      questionMap: {}
    }
  },
  computed: {
    timeSpent() {
      if (!this.record.startTime || !this.record.submitTime) return '-'
      const sec = Math.floor((this.record.submitTime - this.record.startTime) / 1000)
      const m = Math.floor(sec / 60)
      const s = sec % 60
      return `${m}分${s}秒`
    }
  },
  onLoad(query) {
    this.recordId = query.id
    this.load()
  },
  methods: {
    async load() {
      try {
        const res = await examApi.getRecord({ id: this.recordId })
        const d = res.data
        this.record = d.record
        this.results = d.results || []
        this.answers = d.answers || {}
        const map = {}
        for (const q of (d.questions || [])) map[q.id] = q
        this.questionMap = map
        this.loaded = true
      } catch (e) {
        uni.showToast({ title: e.msg || '加载失败', icon: 'none' })
      }
    },
    formatTime(ms) {
      if (!ms) return '-'
      return new Date(ms).toLocaleString()
    },
    formatAnswer(v) {
      if (Array.isArray(v)) return v.join(', ')
      return v
    },
    resultText(r) {
      if (r.needManual) return '待批改'
      return r.correct ? '正确' : '错误'
    },
    resultClass(r) {
      if (r.needManual) return 'st-warning'
      return r.correct ? 'st-success' : 'st-error'
    },
    goBack() {
      uni.navigateBack()
    }
  }
}
</script>

<style scoped>
.page { min-height: 100vh; background: #f5f5f5; padding-bottom: 60rpx; }
.topbar { display: flex; align-items: center; height: 90rpx; padding: 0 30rpx; background: #fff; border-bottom: 1rpx solid #f0f0f0; }
.back { font-size: 50rpx; color: #333; width: 60rpx; }
.title { flex: 1; text-align: center; font-size: 32rpx; font-weight: 500; margin-right: 60rpx; }
.score-card { background: linear-gradient(135deg, #fb454c, #ff6b6b); padding: 50rpx 30rpx; text-align: center; color: #fff; }
.score-num { margin-bottom: 16rpx; }
.big { font-size: 100rpx; font-weight: bold; }
.total { font-size: 36rpx; opacity: 0.8; }
.score-status { display: inline-block; padding: 8rpx 30rpx; border-radius: 30rpx; font-size: 26rpx; margin-bottom: 16rpx; }
.score-status.pass { background: rgba(255, 255, 255, 0.25); }
.score-status.fail { background: rgba(0, 0, 0, 0.25); }
.score-info { font-size: 24rpx; opacity: 0.9; }
.result-list { padding: 20rpx 30rpx; }
.r-item { background: #fff; border-radius: 12rpx; padding: 24rpx; margin-bottom: 16rpx; }
.card { box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06); }
.r-header { display: flex; align-items: center; }
.r-num { color: #fb454c; font-weight: 500; margin-right: 8rpx; }
.r-status { flex: 1; font-size: 26rpx; padding: 4rpx 14rpx; border-radius: 6rpx; display: inline-block; width: max-content; }
.st-success { background: #e8f5e8; color: #4caf50; }
.st-error { background: #ffe8e8; color: #fb454c; }
.st-warning { background: #fff3e0; color: #ff9800; }
.r-score { font-size: 26rpx; color: #888; }
.r-reason { color: #888; font-size: 24rpx; margin-top: 8rpx; }
.r-answer, .r-my { font-size: 26rpx; color: #555; margin-top: 8rpx; line-height: 1.5; }
.r-answer .label, .r-my .label { color: #888; }
.loading { padding: 200rpx 0; text-align: center; color: #aaa; }
</style>
