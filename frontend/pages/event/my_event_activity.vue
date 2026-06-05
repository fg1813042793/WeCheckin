<template>
  <view class="container">
    <view class="tabs">
      <view class="tab-item" :class="{ active: status === '' }" @click="switchStatus('')">全部</view>
      <view class="tab-item" :class="{ active: status === '0' }" @click="switchStatus('0')">待开始</view>
      <view class="tab-item" :class="{ active: status === '1' }" @click="switchStatus('1')">进行中</view>
      <view class="tab-item" :class="{ active: status === '2' }" @click="switchStatus('2')">已结束</view>
    </view>
    <view class="list" v-if="list.length > 0">
      <view class="card" v-for="(item, i) in list" :key="i" @click="goDetail(item)">
        <view class="card-title-row">
          <text class="card-title">{{ item.title }}</text>
          <text class="status-tag" :class="'status-' + (item.status || 0)">{{ item.statusDesc || (item.status === 1 ? '进行中' : item.status === 2 ? '已结束' : '待开始') }}</text>
        </view>
        <text class="card-time" v-if="item.eventStartStr || item.eventEndStr">{{ item.eventStartStr || '' }} ~ {{ item.eventEndStr || '' }}</text>
        <text class="card-desc">{{ item.desc || '' }}</text>
      </view>
    </view>
    <view class="empty" v-else-if="!loading">
      <text class="empty-text">暂无活动</text>
    </view>
    <view class="loading" v-if="loading"><text>加载中...</text></view>
  </view>
</template>

<script>
import { eventApi } from '../../api/index'
export default {
  data() {
    return {
      status: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      loading: false
    }
  },
  onLoad() { this.loadData() },
  onReachBottom() { this.loadMore() },
  methods: {
    getUserId() {
      const userInfo = uni.getStorageSync('userInfo')
      const token = uni.getStorageSync('token')
      return (userInfo && (userInfo.miniOpenID || userInfo.id)) || token || ''
    },
    switchStatus(s) { this.status = s; this.page = 1; this.list = []; this.hasMore = true; this.loadData() },
    async loadData() {
      if (this.loading) return
      this.loading = true
      try {
        const uid = this.getUserId()
        const params = { page: this.page, pageSize: this.pageSize, user_id: uid, type: '1' }
        if (this.status !== '') params.status = this.status
        const res = await eventApi.myParticipate(params)
        const data = Array.isArray(res.data) ? res.data : (res.data.list || [])
        if (this.page === 1) { this.list = data } else { this.list = [...this.list, ...data] }
        this.hasMore = data.length >= this.pageSize
      } catch (e) { console.error(e) }
      this.loading = false
    },
    loadMore() { if (this.hasMore && !this.loading) { this.page++; this.loadData() } },
    goDetail(item) { uni.navigateTo({ url: '/pages/event/event_detail?id=' + item.id }) }
  }
}
</script>

<style scoped>
.container { min-height: 100vh; background-color: #f5f5f5; }
.tabs { display: flex; background-color: #fff; padding: 20rpx; }
.tab-item { flex: 1; text-align: center; font-size: 26rpx; color: #666; padding-bottom: 12rpx; }
.tab-item.active { color: #fb454c; font-weight: bold; border-bottom: 3rpx solid #fb454c; }
.list { padding: 20rpx; }
.card { background-color: #fff; border-radius: 12rpx; padding: 20rpx; margin-bottom: 16rpx; }
.card-title-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8rpx; }
.card-title { font-size: 30rpx; font-weight: bold; color: #333; flex: 1; }
.status-tag { font-size: 22rpx; padding: 4rpx 12rpx; border-radius: 6rpx; flex-shrink: 0; }
.status-0 { background-color: #f0f5ff; color: #2b7ef5; }
.status-1 { background-color: #fff7e6; color: #fa8c16; }
.status-2 { background-color: #f5f5f5; color: #999; }
.card-time { font-size: 24rpx; color: #999; display: block; margin-bottom: 8rpx; }
.card-desc { font-size: 26rpx; color: #666; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.empty, .loading { display: flex; align-items: center; justify-content: center; padding-top: 200rpx; font-size: 28rpx; color: #999; }
</style>
