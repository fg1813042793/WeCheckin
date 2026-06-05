<template>
  <view class="container" :style="{ paddingTop: containerPad }">
    <view class="tabs" :style="{ top: fixedTop }">
      <view class="tab" :class="{ active: tab === 'today' }" @click="switchTab('today')">今日活动</view>
      <view class="tab" :class="{ active: tab === 'mine' }" @click="switchTab('mine')">我参与的</view>
    </view>
    <scroll-view scroll-y class="content" :style="{ height: contentHeight }">
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
            </view>
            <text class="card-desc">{{ item.desc || '快来报名参与吧！' }}</text>
            <view class="card-footer">
              <view class="card-info">
                <text class="info-item">{{ item.userCnt || 0 }}人参与</text>
                <text class="info-item" v-if="item.regStartStr || item.regEndStr">报名时间：{{ fmtTime(item.regStartStr) }} ~ {{ fmtTime(item.regEndStr) }}</text>
              </view>
            </view>
            <text class="status-tag" v-if="item.statusDesc">{{ item.statusDesc }}</text>
          </view>
        </view>
      </view>
      <view class="empty" v-else-if="!loading">
        <image src="/static/empty.png" mode="aspectFit" class="empty-img"></image>
        <text class="empty-text">暂无活动</text>
      </view>
      <view class="loading-more" v-if="loading">
        <text class="loading-text">加载中...</text>
      </view>
      <view class="load-more" v-else-if="hasMore" @click="loadMore">加载更多</view>
    </scroll-view>
  </view>
</template>

<script>
import { eventApi } from '../../api/index'

export default {
  data() {
    return {
      tab: 'today',
      list: [],
      page: 1,
      pageSize: 10,
      loading: false,
      hasMore: true,
      fixedTop: '0px',
      containerPad: '0px',
      contentHeight: '100vh'
    }
  },
  onLoad() {
    const sys = uni.getSystemInfoSync()
    const pxScale = 750 / sys.windowWidth
    if (sys.platform === 'android') {
      this.fixedTop = '12rpx'
      this.containerPad = '92rpx'
      this.contentHeight = (sys.windowHeight - Math.round(80 / pxScale)) + 'px'
    } else {
      const navOffset = (sys.statusBarHeight || 0) + 44
      this.fixedTop = (navOffset + 6) + 'px'
      this.containerPad = (navOffset + 6 + Math.round(80 / pxScale)) + 'px'
      this.contentHeight = (sys.windowHeight - Math.round(80 / pxScale)) + 'px'
    }
    this.loadData()
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
    switchTab(t) {
      this.tab = t
      this.page = 1
      this.list = []
      this.hasMore = true
      this.loadData()
    },
    async loadData() {
      if ((!this.hasMore && this.page > 1) || this.loading) return
      this.loading = true
      try {
        const uid = this.getUserId()
        let res
        if (this.tab === 'today') {
          res = await eventApi.myList({ page: this.page, pageSize: this.pageSize, user_id: uid, type: '1' })
        } else {
          res = await eventApi.myList({ page: this.page, pageSize: this.pageSize, user_id: uid, type: '1' })
        }
        const data = Array.isArray(res.data) ? res.data : (res.data.list || [])
        let filtered = data
        if (this.tab === 'today') {
          filtered = data.filter((item) => {
            const now = Date.now()
            return item.eventStart <= now && item.eventEnd >= now
          })
        }
        if (this.page === 1) { this.list = filtered } else { this.list = [...this.list, ...filtered] }
        this.hasMore = data.length >= this.pageSize
      } catch (e) { console.error('加载失败', e) }
      this.loading = false
    },
    loadMore() { if (this.hasMore && !this.loading) { this.page++; this.loadData() } },
    goDetail(item) {
      uni.navigateTo({ url: '/pages/event/event_detail?id=' + item.id })
    },
    fmtTime(str) {
      if (!str) return '-'
      const parts = str.split(' ')
      return parts[0].replace(/-/g, '/')
    },
    getPlaceholderBg(index) {
      const colors = [
        'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
        'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
        'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)',
        'linear-gradient(135deg, #fa709a 0%, #fee140 100%)',
      ]
      return colors[index % colors.length]
    }
  }
}
</script>

<style scoped>
page { background-color: #f5f5f5; }
.container { min-height: 100vh; background: #f5f5f5; }
.tabs {
  display: flex;
  background: #fff;
  border-bottom: 1rpx solid #eee;
  position: fixed;
  left: 0;
  right: 0;
  z-index: 10;
}
.tab {
  flex: 1;
  text-align: center;
  padding: 24rpx 0;
  font-size: 28rpx;
  color: #666;
  position: relative;
}
.tab.active {
  color: #fb454c;
  font-weight: bold;
}
.tab.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 60rpx;
  height: 4rpx;
  background: #fb454c;
  border-radius: 2rpx;
}
.content { padding: 20rpx; box-sizing: border-box; }
.event-card {
  background: #fff;
  border-radius: 16rpx;
  overflow: hidden;
  margin-bottom: 20rpx;
  box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.06);
}
.card-img { width: 100%; height: 300rpx; }
.card-img-inner { width: 100%; height: 100%; }
.card-img-placeholder {
  width: 100%; height: 300rpx;
  display: flex; align-items: center; justify-content: center;
}
.placeholder-text { color: #fff; font-size: 40rpx; font-weight: bold; }
.card-body { padding: 24rpx; position: relative; }
.card-title-row { display: flex; align-items: center; margin-bottom: 12rpx; }
.card-title { font-size: 30rpx; font-weight: bold; color: #333; flex: 1; }
.card-desc { font-size: 26rpx; color: #999; line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.card-footer { display: flex; align-items: center; justify-content: space-between; margin-top: 16rpx; }
.card-info { display: flex; flex-direction: column; gap: 6rpx; }
.info-item { font-size: 24rpx; color: #999; }
.status-tag {
  position: absolute; top: 24rpx; right: 24rpx;
  font-size: 22rpx; padding: 4rpx 12rpx; border-radius: 8rpx;
  background: #fb454c; color: #fff;
}
.empty { display: flex; flex-direction: column; align-items: center; padding-top: 120rpx; }
.empty-img { width: 200rpx; height: 200rpx; }
.empty-text { font-size: 28rpx; color: #999; margin-top: 20rpx; }
.loading-more { text-align: center; padding: 30rpx; }
.loading-text { font-size: 26rpx; color: #999; }
.load-more { text-align: center; padding: 20rpx; font-size: 26rpx; color: #999; }
</style>
