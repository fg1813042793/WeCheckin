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
        <view class="event-card" v-for="(item, index) in list" :key="index" @click="goDetail(item)">
          <view v-if="item.img" class="card-img">
            <image :src="item.img" mode="aspectFill" class="card-img-inner" />
          </view>
          <view v-else class="card-img-placeholder" :style="{ background: getPlaceholderBg(index) }">
            <text class="placeholder-text">{{ item.title }}</text>
          </view>
          <view class="card-body">
            <view class="card-title-row">
              <text class="card-title">{{ item.title }}</text>
              <text class="type-tag" :class="item.type === 1 ? 'tag-activity' : 'tag-competition'">{{ item.type === 1 ? '活动' : '赛事' }}</text>
            </view>
            <text class="card-desc">{{ item.desc || '快来报名参与吧！' }}</text>
            <view class="card-footer">
              <view class="card-info">
                <text class="info-item">{{ item.userCnt || 0 }}人参与</text>
                <text class="info-item" v-if="item.regStartStr || item.regEndStr">报名时间：{{ formatTime(item.regStartStr) }} ~ {{ formatTime(item.regEndStr) }}</text>
                <text class="info-item" v-if="item.eventStartStr || item.eventEndStr">{{ item.type === 1 ? '活动时间' : '比赛时间' }}：{{ formatTime(item.eventStartStr) }} ~ {{ formatTime(item.eventEndStr) }}</text>
              </view>
              <view v-if="!item.isJoin" class="card-btn" @click.stop="handleJoin(item)">立即报名</view>
              <view v-else class="card-btn joined">已报名</view>
            </view>
            <text class="status-tag" v-if="item.statusDesc">{{ item.statusDesc }}</text>
          </view>
        </view>
      </view>
      <view class="empty" v-else-if="!loading">
        <image src="/static/empty.png" mode="aspectFit" class="empty-img"></image>
        <text class="empty-text">暂无赛事活动</text>
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
    const typeFilter = uni.getStorageSync('eventTypeFilter')
    if (typeFilter) {
      this.cur = typeFilter
      uni.removeStorageSync('eventTypeFilter')
    }
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
        const params = { page: this.page, pageSize: this.pageSize, user_id: uid, keyword: this.keyword }
        if (this.cur !== 'all') params.type = this.cur
        const res = await eventApi.getList(params)
        const data = Array.isArray(res.data) ? res.data : (res.data.list || [])
        if (this.page === 1) { this.list = data } else { this.list = [...this.list, ...data] }
        this.hasMore = data.length >= this.pageSize
      } catch (e) { console.error('加载失败', e) }
      this.loading = false
    },
    loadMore() { if (this.hasMore && !this.loading) { this.page++; this.loadData() } },
    goDetail(item) { uni.navigateTo({ url: '/pages/event/event_detail?id=' + item.id }) },
    async handleJoin(item) {
      const uid = this.getUserId()
      if (!uid) { uni.showToast({ title: '请先登录', icon: 'none' }); return }
      let forms = []
      try { forms = typeof item.forms === 'string' ? JSON.parse(item.forms || '[]') : (item.forms || []) } catch (e) {}
      if (forms.length > 0) {
        uni.navigateTo({ url: '/pages/event/event_join_form?id=' + item.id })
        return
      }
      try {
        await eventApi.participate({ event_id: item.id, user_id: uid, forms: '[]' })
        item.isJoin = true
        item.userCnt = (item.userCnt || 0) + 1
        uni.showToast({ title: '报名成功', icon: 'success' })
      } catch (e) { console.error('报名失败', e) }
    },
    formatTime(str) {
      if (!str) return '-'
      const parts = str.split(' ')
      if (parts.length > 0) {
        return parts[0].replace(/-/g, '/')
      }
      return str
    },
    getPlaceholderBg(index) {
      const colors = [
        'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
        'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
        'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)',
        'linear-gradient(135deg, #fa709a 0%, #fee140 100%)',
        'linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%)',
        'linear-gradient(135deg, #fccb90 0%, #d57eeb 100%)',
        'linear-gradient(135deg, #e0c3fc 0%, #8ec5fc 100%)',
        'linear-gradient(135deg, #f5576c 0%, #ff758c 100%)',
        'linear-gradient(135deg, #3b82f6 0%, #2dd4bf 100%)',
      ]
      return colors[index % colors.length]
    }
  }
}
</script>

<style scoped>
.container { min-height: 100vh; background-color: #f5f5f5; }
.header-sticky { position: fixed; left: 0; right: 0; z-index: 10; background-color: #fff; }
.header-sticky::before { content: ''; position: absolute; top: -12rpx; left: 0; right: 0; height: 12rpx; background-color: #f5f5f5; }
.search-bar { display: flex; align-items: center; padding: 20rpx; }
.search-input { flex: 1; height: 64rpx; background-color: #f5f5f5; border-radius: 32rpx; padding: 0 24rpx; font-size: 26rpx; color: #333; }
.search-btn { font-size: 26rpx; color: #fb454c; flex-shrink: 0; margin-left: 16rpx; }
.tabs { display: flex; background-color: #fff; padding: 0 20rpx 20rpx; }
.tab-item { flex: 1; text-align: center; padding: 16rpx 0; font-size: 28rpx; color: #666; position: relative; }
.tab-item.active { color: #fb454c; font-weight: bold; }
.tab-item.active::after { content: ''; position: absolute; bottom: 0; left: 50%; transform: translateX(-50%); width: 60rpx; height: 4rpx; background-color: #fb454c; border-radius: 2rpx; }
.event-list { padding: 16rpx 0 0; }
.event-card { background-color: #fff; border-radius: 16rpx; overflow: hidden; margin: 0 4rpx 20rpx; position: relative; }
.card-img { width: 100%; height: 300rpx; overflow: hidden; }
.card-img-inner { width: 100%; height: 100%; }
.card-img-placeholder { width: 100%; height: 300rpx; display: flex; align-items: center; justify-content: center; }
.placeholder-text { font-size: 40rpx; color: #fff; font-weight: bold; text-align: center; word-break: break-all; line-height: 1.4; padding: 0 20rpx; }
.card-body { padding: 24rpx; }
.card-title-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12rpx; }
.card-title { font-size: 32rpx; font-weight: bold; color: #333; flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; margin-right: 12rpx; }
.type-tag { font-size: 22rpx; padding: 4rpx 12rpx; border-radius: 8rpx; flex-shrink: 0; }
.tag-activity { background-color: #fff7e6; color: #fa8c16; }
.tag-competition { background-color: #fff1f0; color: #f5222d; }
.card-desc { font-size: 26rpx; color: #666; display: block; margin-bottom: 20rpx; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.card-footer { display: flex; justify-content: space-between; align-items: flex-end; }
.card-info { display: flex; flex-direction: column; gap: 8rpx; flex: 1; }
.card-btn { background-color: #fb454c; color: #fff; padding: 12rpx 32rpx; border-radius: 30rpx; font-size: 26rpx; flex-shrink: 0; margin-left: 16rpx; }
.card-btn.joined { background-color: #eee; color: #999; padding: 12rpx 32rpx; border-radius: 30rpx; font-size: 26rpx; flex-shrink: 0; margin-left: 16rpx; }
.info-item { font-size: 24rpx; color: #999; }
.status-tag { position: absolute; top: 16rpx; right: 16rpx; background-color: rgba(0,0,0,0.5); color: #fff; font-size: 22rpx; padding: 4rpx 12rpx; border-radius: 8rpx; }
.empty { display: flex; flex-direction: column; align-items: center; justify-content: center; padding-top: 200rpx; }
.empty-img { width: 240rpx; height: 240rpx; }
.empty-text { font-size: 28rpx; color: #999; margin-top: 30rpx; }
.loading-more { text-align: center; padding: 20rpx; }
.loading-text { font-size: 24rpx; color: #999; }
</style>
