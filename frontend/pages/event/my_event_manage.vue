<template>
  <view class="container" :style="{ paddingTop: containerPad }">
    <view class="header-sticky" :style="{ top: fixedTop }">
      <view class="search-bar">
        <input v-model="keyword" placeholder="搜索赛事活动" class="search-input" @confirm="handleSearch" />
        <text class="search-btn" @click="handleSearch">搜索</text>
      </view>
      <view class="tabs">
        <view class="tab-item" :class="{ active: cur === 'all' }" @click="switchTab('all')">全部</view>
        <view class="tab-item" :class="{ active: cur === '1' }" @click="switchTab('1')">活动</view>
        <view class="tab-item" :class="{ active: cur === '2' }" @click="switchTab('2')">赛事</view>
      </view>
    </view>
    <view class="content">
      <view class="event-list" v-if="list.length > 0">
        <view class="event-card" v-for="(item, index) in list" :key="index" @click="goManage(item)">
          <view v-if="item.img" class="card-img">
            <image :src="item.img" mode="aspectFill" class="card-img-inner" :key="item.img" />
          </view>
          <view v-else class="card-img-placeholder" :style="{ background: getPlaceholderBg(index) }">
            <text class="placeholder-text">{{ item.title }}</text>
          </view>
          <view class="card-body">
            <view class="card-title-row">
              <text class="card-title">{{ item.title }}</text>
              <text class="role-tag">{{ item.roleName || (item.role === 'organizer' ? '工作人员:主办人' : item.role === 'assistant' ? '工作人员:主办人助理' : item.role === 'referee' ? '工作人员:裁判' : '') }}</text>
            </view>
            <text class="card-desc">{{ item.desc || '' }}</text>
            <view class="card-footer">
              <text class="card-info">{{ item.userCnt || 0 }}人参与</text>
              <view class="card-actions">
                <view class="action-btn" @click.stop="goPublish(item)">发布动态</view>
                <view class="action-btn" @click.stop="goScoreEntry(item)" v-if="item.type === 2">成绩录入</view>
              </view>
            </view>
          </view>
        </view>
      </view>
      <view class="empty" v-else-if="!loading">
        <image src="/static/empty.png" mode="aspectFit" class="empty-img"></image>
        <text class="empty-text">暂无管理的赛事活动</text>
      </view>
      <view class="loading-more" v-if="loading">
        <text class="loading-text">加载中...</text>
      </view>
    </view>
  </view>
</template>

<script>
import { eventApi } from '../../api/index'
export default {
  data() {
    return {
      cur: 'all',
      list: [],
      page: 1,
      pageSize: 10,
      keyword: '',
      hasMore: true,
      loading: false,
      fixedTop: '0px',
      containerPad: '0px'
    }
  },
  onLoad() {
    this.loadData()
    const sys = uni.getSystemInfoSync()
    if (sys.platform === 'android') {
      this.fixedTop = '12rpx'
      this.containerPad = '192rpx'
    } else {
      const navOffset = (sys.statusBarHeight || 0) + 44
      this.fixedTop = (navOffset + 6) + 'px'
      this.containerPad = (navOffset + 6 + Math.round(180 / 750 * sys.windowWidth)) + 'px'
    }
  },
  onShow() {
    this.page = 1
    this.hasMore = true
    this.loadData()
  },
  onReachBottom() {
    this.loadMore()
  },
  onPullDownRefresh() {
    this.page = 1
    this.hasMore = true
    this.loadData().then(() => { uni.stopPullDownRefresh() })
  },
  methods: {
    getUserId() {
      const userInfo = uni.getStorageSync('userInfo')
      const token = uni.getStorageSync('token')
      return (userInfo && (userInfo.miniOpenID || userInfo.id)) || token || ''
    },
    handleSearch() { this.page = 1; this.list = []; this.hasMore = true; this.loadData() },
    switchTab(tab) { this.cur = tab; this.keyword = ''; this.page = 1; this.list = []; this.hasMore = true; this.loadData() },
    async loadData() {
      if ((!this.hasMore && this.page > 1) || this.loading) return
      this.loading = true
      try {
        const uid = this.getUserId()
        const params = { user_id: uid, page: this.page, pageSize: this.pageSize, keyword: this.keyword }
        if (this.cur !== 'all') params.type = this.cur
        const res = await eventApi.myManaged(params)
        const data = Array.isArray(res.data) ? res.data : (res.data.list || [])
        if (this.page === 1) { this.list = data } else { this.list = [...this.list, ...data] }
        this.hasMore = data.length >= this.pageSize
      } catch (e) { console.error('加载失败', e) }
      this.loading = false
    },
    loadMore() { if (this.hasMore && !this.loading) { this.page++; this.loadData() } },
    goManage(item) { uni.navigateTo({ url: '/pages/event/event_detail?id=' + item.id }) },
    goPublish(item) {
      uni.navigateTo({ url: '/pages/event/event_publish_dynamic?event_id=' + item.id })
    },
    goScoreEntry(item) {
      uni.navigateTo({ url: '/pages/event/my_event_score_entry?event_id=' + item.id + '&title=' + encodeURIComponent(item.title) })
    },
    getPlaceholderBg(index) {
      const colors = ['#ff6b6b', '#f8a5c2', '#74b9ff', '#55efc4', '#fdcb6e', '#a29bfe']
      return colors[index % colors.length]
    }
  }
}
</script>

<style scoped>
.container { min-height: 100vh; background-color: #f5f5f5; }
.header-sticky { position: fixed; left: 0; right: 0; z-index: 10; background-color: #f5f5f5; }
.search-bar { display: flex; padding: 16rpx 20rpx; background-color: #fff; align-items: center; gap: 16rpx; }
.search-input { flex: 1; height: 64rpx; background-color: #f5f5f5; border-radius: 32rpx; padding: 0 28rpx; font-size: 26rpx; }
.search-btn { font-size: 28rpx; color: #fb454c; flex-shrink: 0; }
.tabs { display: flex; background-color: #fff; padding: 0 20rpx; border-bottom: 1rpx solid #eee; }
.tab-item { padding: 20rpx 24rpx; font-size: 28rpx; color: #666; position: relative; }
.tab-item.active { color: #fb454c; font-weight: bold; }
.tab-item.active::after { content: ''; position: absolute; bottom: 0; left: 24rpx; right: 24rpx; height: 4rpx; background-color: #fb454c; border-radius: 2rpx; }
.content { padding: 20rpx; }
.event-list { display: flex; flex-direction: column; }
.event-card { background-color: #fff; border-radius: 16rpx; overflow: hidden; box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.06); margin-bottom: 16rpx; }
.event-card:last-child { margin-bottom: 0; }
.card-img { width: 100%; height: 300rpx; display: flex; align-items: center; justify-content: center; overflow: hidden; }
.card-img-inner { width: 100%; height: 100%; flex-shrink: 0; object-fit: cover; }
.card-img-placeholder { width: 100%; height: 300rpx; display: flex; align-items: center; justify-content: center; }
.placeholder-text { font-size: 48rpx; color: rgba(255,255,255,0.8); font-weight: bold; }
.card-body { padding: 20rpx; }
.card-title-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8rpx; }
.card-title { font-size: 30rpx; font-weight: bold; color: #333; flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.role-tag { font-size: 22rpx; background-color: #f0f5ff; color: #2b7ef5; padding: 4rpx 16rpx; border-radius: 6rpx; flex-shrink: 0; margin-left: 12rpx; }
.card-desc { font-size: 26rpx; color: #666; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; margin-bottom: 12rpx; }
.card-footer { display: flex; justify-content: space-between; align-items: center; }
.card-info { font-size: 24rpx; color: #999; }
.card-actions { display: flex; }
.action-btn { background-color: #f5f5f5; padding: 8rpx 24rpx; border-radius: 20rpx; font-size: 24rpx; color: #333; margin-left: 12rpx; }
.action-btn:first-child { margin-left: 0; }
.empty { display: flex; flex-direction: column; align-items: center; justify-content: center; padding-top: 200rpx; }
.empty-img { width: 240rpx; height: 240rpx; }
.empty-text { font-size: 28rpx; color: #999; margin-top: 20rpx; }
.loading-more { text-align: center; padding: 30rpx; }
.loading-text { font-size: 26rpx; color: #999; }
</style>
