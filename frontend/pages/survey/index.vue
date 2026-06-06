<template>
  <view class="page">
    <view class="topbar">
      <view class="back" @click="goBack">‹</view>
      <view class="title">问卷中心</view>
    </view>

    <view class="search-bar">
      <input v-model="keyword" placeholder="搜索问卷" @confirm="load" />
    </view>

    <view v-if="list.length > 0" class="survey-list">
      <view v-for="s in list" :key="s.id" class="survey-item card" @click="goFill(s)">
        <view class="s-title">{{ s.title }}</view>
        <view class="s-desc" v-if="s.description">{{ s.description }}</view>
        <view class="s-meta">
          <text class="meta-tag" v-if="s.category">{{ s.category }}</text>
          <text class="meta-tag" v-if="s.anonymous===1">匿名</text>
          <text class="meta-tag" v-if="s.allowMulti===1">可多次</text>
        </view>
        <view class="s-time" v-if="s.startTime || s.endTime">
          <text>时间: {{ formatTime(s.startTime) }} ~ {{ formatTime(s.endTime) }}</text>
        </view>
        <button class="btn-fill" @click.stop="goFill(s)">立即填写</button>
      </view>
    </view>

    <view class="empty" v-else>
      <text>{{ loading ? '加载中...' : '暂无可填写的问卷' }}</text>
    </view>

    <view class="my-section" v-if="isLogged">
      <view class="divider"><text>我填过的</text></view>
      <view v-for="r in myList" :key="r.id" class="my-item card" @click="goMyResp(r)">
        <view class="my-left">
          <view class="my-title">问卷 #{{ r.surveyId }}</view>
          <view class="my-time">{{ formatTime(r.submitTime) }}</view>
        </view>
        <view class="my-status st-success">已完成</view>
      </view>
    </view>
  </view>
</template>

<script>
import { surveyApi } from '../../api/index'

export default {
  data() {
    return {
      keyword: '',
      list: [],
      myList: [],
      loading: false
    }
  },
  computed: {
    isLogged() {
      return !!uni.getStorageSync('token')
    }
  },
  onShow() {
    this.load()
    if (this.isLogged) this.loadMy()
  },
  methods: {
    async load() {
      this.loading = true
      try {
        const res = await surveyApi.getList({ page: 1, pageSize: 50, keyword: this.keyword })
        this.list = (res.data && res.data.list) || []
      } catch (e) {
        console.error(e)
      } finally { this.loading = false }
    },
    async loadMy() {
      try {
        const res = await surveyApi.myResponses()
        this.myList = (res.data && res.data.list) || []
      } catch (e) {}
    },
    goFill(s) {
      uni.navigateTo({ url: `/pages/survey/fill?id=${s.id}` })
    },
    goMyResp(r) {
      uni.navigateTo({ url: `/pages/survey/result?id=${r.id}` })
    },
    formatTime(ms) {
      if (!ms) return '-'
      const d = new Date(ms)
      return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
    },
    goBack() { uni.navigateBack() }
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
.survey-list { padding: 20rpx 30rpx; }
.survey-item { background: #fff; border-radius: 16rpx; padding: 30rpx; margin-bottom: 20rpx; }
.card { box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06); }
.s-title { font-size: 32rpx; font-weight: 500; color: #333; margin-bottom: 12rpx; }
.s-desc { font-size: 24rpx; color: #888; margin-bottom: 14rpx; line-height: 1.5; }
.s-meta { display: flex; gap: 14rpx; margin-bottom: 12rpx; flex-wrap: wrap; }
.meta-tag { background: #f0f5ff; color: #4a7af0; font-size: 22rpx; padding: 4rpx 14rpx; border-radius: 6rpx; }
.s-time { color: #888; font-size: 24rpx; margin-bottom: 18rpx; }
.btn-fill { background: linear-gradient(90deg, #fb454c, #ff6b6b); color: #fff; border-radius: 50rpx; font-size: 28rpx; height: 80rpx; line-height: 80rpx; }
.empty { text-align: center; padding: 100rpx 0; color: #aaa; font-size: 28rpx; }
.my-section { padding: 0 30rpx; }
.divider { text-align: center; color: #aaa; font-size: 24rpx; margin: 30rpx 0; position: relative; }
.divider::before, .divider::after { content: ''; position: absolute; top: 50%; width: 25%; height: 1rpx; background: #ddd; }
.divider::before { left: 10%; }
.divider::after { right: 10%; }
.my-item { background: #fff; border-radius: 12rpx; padding: 24rpx; margin-bottom: 14rpx; display: flex; align-items: center; }
.my-left { flex: 1; }
.my-title { font-size: 28rpx; color: #333; }
.my-time { font-size: 22rpx; color: #999; margin-top: 6rpx; }
.my-status { padding: 4rpx 12rpx; font-size: 20rpx; border-radius: 4rpx; }
.st-success { background: #e8f5e8; color: #4caf50; }
</style>
