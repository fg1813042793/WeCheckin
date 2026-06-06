<template>
  <view class="page">
    <view class="topbar">
      <view class="back" @click="goBack">‹</view>
      <view class="title">在线考试</view>
    </view>

    <view class="search-bar">
      <input v-model="keyword" placeholder="搜索考试" @confirm="load" />
    </view>

    <view class="exam-list" v-if="list.length > 0">
      <view v-for="exam in list" :key="exam.id" class="exam-item card">
        <view class="exam-title">{{ exam.title }}</view>
        <view class="exam-meta">
          <text class="meta-tag">时长 {{ exam.duration || 60 }}分</text>
          <text class="meta-tag">最多 {{ exam.maxAttempts || 1 }}次</text>
        </view>
        <view class="exam-time" v-if="exam.startTime || exam.endTime">
          <text>时间: {{ formatTime(exam.startTime) }} ~ {{ formatTime(exam.endTime) }}</text>
        </view>
        <button class="btn-start" @click="onStart(exam)">开始考试</button>
      </view>
    </view>

    <view class="empty" v-else>
      <text>{{ loading ? '加载中...' : '暂无考试' }}</text>
    </view>

    <view class="my-records">
      <view class="divider"><text>我的考试记录</text></view>
      <view v-for="rec in records" :key="rec.id" class="record-item card" @click="onViewRecord(rec)">
        <view class="rec-left">
          <view class="rec-title">记录 #{{ rec.id }}</view>
          <view class="rec-time">{{ formatTime(rec.startTime) }}</view>
        </view>
        <view class="rec-right">
          <view class="rec-score">
            <text class="score-num">{{ rec.score }}</text>
            <text class="score-total">/{{ rec.totalScore }}</text>
          </view>
          <view class="rec-status" :class="statusClass(rec.status)">
            {{ statusText(rec.status) }}
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { examApi } from '../../api/index'

export default {
  data() {
    return {
      keyword: '',
      list: [],
      records: [],
      loading: false
    }
  },
  onShow() {
    this.load()
    this.loadRecords()
  },
  methods: {
    async load() {
      this.loading = true
      try {
        const res = await examApi.getList({ page: 1, pageSize: 50, keyword: this.keyword })
        this.list = (res.data && res.data.list) || []
      } catch (e) {
        console.error(e)
      } finally { this.loading = false }
    },
    async loadRecords() {
      try {
        const res = await examApi.myRecords()
        this.records = (res.data && res.data.list) || []
      } catch (e) {}
    },
    onStart(exam) {
      uni.navigateTo({ url: `/pages/survey/exam/exam_take?examId=${exam.id}` })
    },
    onViewRecord(rec) {
      uni.navigateTo({ url: `/pages/survey/exam/exam_result?id=${rec.id}` })
    },
    formatTime(ms) {
      if (!ms) return '-'
      const d = new Date(ms)
      return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
    },
    statusText(s) {
      return { 0: '进行中', 1: '待批改', 2: '已完成' }[s] || '未知'
    },
    statusClass(s) {
      return { 0: 'st-info', 1: 'st-warning', 2: 'st-success' }[s] || ''
    },
    goBack() {
      uni.navigateBack()
    }
  }
}
</script>

<style scoped>
.page { padding-bottom: 200rpx; }
.topbar { display: flex; align-items: center; height: 90rpx; padding: 0 30rpx; background: #fff; border-bottom: 1rpx solid #f0f0f0; }
.back { font-size: 50rpx; color: #333; width: 60rpx; }
.title { flex: 1; text-align: center; font-size: 32rpx; font-weight: 500; margin-right: 60rpx; }
.search-bar { padding: 20rpx 30rpx; background: #fff; }
.search-bar input { background: #f5f5f5; height: 70rpx; border-radius: 35rpx; padding: 0 24rpx; font-size: 28rpx; }
.exam-list { padding: 20rpx 30rpx; }
.exam-item { background: #fff; border-radius: 16rpx; padding: 30rpx; margin-bottom: 20rpx; }
.card { box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06); }
.exam-title { font-size: 32rpx; font-weight: 500; color: #333; margin-bottom: 14rpx; }
.exam-meta { display: flex; gap: 14rpx; margin-bottom: 12rpx; }
.meta-tag { background: #f0f5ff; color: #4a7af0; font-size: 22rpx; padding: 4rpx 14rpx; border-radius: 6rpx; }
.exam-time { color: #888; font-size: 24rpx; margin-bottom: 18rpx; }
.btn-start { background: linear-gradient(90deg, #fb454c, #ff6b6b); color: #fff; border-radius: 50rpx; font-size: 28rpx; height: 80rpx; line-height: 80rpx; }
.empty { text-align: center; padding: 100rpx 0; color: #aaa; font-size: 28rpx; }
.my-records { padding: 0 30rpx; }
.divider { text-align: center; color: #aaa; font-size: 24rpx; margin: 30rpx 0; position: relative; }
.divider::before, .divider::after { content: ''; position: absolute; top: 50%; width: 25%; height: 1rpx; background: #ddd; }
.divider::before { left: 10%; }
.divider::after { right: 10%; }
.record-item { background: #fff; border-radius: 12rpx; padding: 24rpx; margin-bottom: 14rpx; display: flex; align-items: center; }
.rec-left { flex: 1; }
.rec-title { font-size: 28rpx; color: #333; }
.rec-time { font-size: 22rpx; color: #999; margin-top: 6rpx; }
.rec-right { text-align: right; }
.rec-score { font-size: 28rpx; }
.score-num { color: #fb454c; font-weight: bold; font-size: 36rpx; }
.score-total { color: #888; font-size: 22rpx; }
.rec-status { display: inline-block; margin-top: 6rpx; padding: 2rpx 12rpx; font-size: 20rpx; border-radius: 4rpx; }
.st-info { background: #e8f4ff; color: #4a7af0; }
.st-warning { background: #fff3e0; color: #ff9800; }
.st-success { background: #e8f5e8; color: #4caf50; }
</style>
