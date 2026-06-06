<template>
  <view class="page" v-if="survey">
    <view class="topbar">
      <view class="back" @click="goBack">‹</view>
      <view class="title">{{ survey.title }} - 我的答卷</view>
    </view>

    <view class="intro card">
      <view class="intro-title">提交成功</view>
      <view class="intro-desc">感谢您的参与！</view>
    </view>

    <view class="q-list">
      <view v-for="(q, i) in questions" :key="q.id" class="q-item card">
        <view class="q-title">
          <text class="q-num">{{ i + 1 }}.</text>
          <text>{{ q.title }}</text>
        </view>
        <view class="q-answer">{{ formatAnswer(answers[q.id]) }}</view>
      </view>
    </view>

    <view class="footer">
      <button class="btn-back" @click="goBack">返回问卷中心</button>
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
      answers: {}
    }
  },
  onLoad(query) {
    this.respId = query.id
    this.surveyId = query.surveyId
    this.load()
  },
  methods: {
    async load() {
      try {
        let res
        if (this.respId) {
          res = await surveyApi.myResponse({ id: this.respId })
        } else {
          res = await surveyApi.getDetail({ id: this.surveyId })
        }
        const d = res.data
        this.answers = d.answers || {}
        if (d.survey) {
          this.survey = d.survey
          try {
            const sch = d.survey.schema ? JSON.parse(d.survey.schema) : { questions: [] }
            this.questions = sch.questions || []
          } catch { this.questions = [] }
        }
      } catch (e) {
        uni.showToast({ title: e.msg || '加载失败', icon: 'none' })
      }
    },
    formatAnswer(v) {
      if (v == null) return '(未填)'
      if (Array.isArray(v)) return v.join(', ')
      if (typeof v === 'object') return JSON.stringify(v)
      return v
    },
    goBack() {
      uni.reLaunch({ url: '/pages/survey/index' })
    }
  }
}
</script>

<style scoped>
.page { min-height: 100vh; background: #f5f5f5; padding-bottom: 160rpx; }
.topbar { display: flex; align-items: center; height: 90rpx; padding: 0 30rpx; background: #fff; border-bottom: 1rpx solid #f0f0f0; }
.back { font-size: 50rpx; color: #333; width: 60rpx; }
.title { flex: 1; text-align: center; font-size: 30rpx; font-weight: 500; margin-right: 60rpx; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.intro { margin: 20rpx 30rpx; padding: 30rpx; background: linear-gradient(135deg, #e8f5e8, #c8e6c9); border-radius: 16rpx; text-align: center; }
.card { box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06); }
.intro-title { font-size: 36rpx; font-weight: 500; color: #4caf50; margin-bottom: 10rpx; }
.intro-desc { font-size: 26rpx; color: #666; }
.q-list { padding: 0 30rpx 20rpx; }
.q-item { background: #fff; border-radius: 16rpx; padding: 30rpx; margin-bottom: 20rpx; }
.q-title { font-size: 30rpx; color: #333; margin-bottom: 16rpx; line-height: 1.5; }
.q-num { color: #fb454c; font-weight: 500; margin-right: 6rpx; }
.q-answer { background: #f7f7f7; border-radius: 8rpx; padding: 16rpx 20rpx; color: #333; font-size: 28rpx; }
.footer { position: fixed; bottom: 0; left: 0; right: 0; padding: 20rpx 30rpx; background: #fff; border-top: 1rpx solid #f0f0f0; }
.btn-back { background: #fff; color: #fb454c; border: 2rpx solid #fb454c; border-radius: 50rpx; font-size: 30rpx; height: 88rpx; line-height: 84rpx; }
.loading { padding: 200rpx 0; text-align: center; color: #aaa; }
</style>
