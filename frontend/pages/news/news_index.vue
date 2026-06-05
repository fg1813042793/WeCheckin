<template>
  <view class="container" :style="{ paddingTop: containerPad }">
    <view class="search-bar" :style="{ top: fixedTop }">
      <input class="search-input" v-model="keyword" placeholder="搜索通知标题" @confirm="onSearch" confirm-type="search" />
      <text class="search-btn" @click="onSearch">搜索</text>
    </view>

    <view class="news-list" v-if="list.length > 0">
      <view class="news-item" v-for="(item, index) in list" :key="index" @click="goDetail(item.id)">
        <view v-if="item.img" class="news-img">
          <image :src="item.img" mode="aspectFill" class="news-img-inner" />
        </view>
        <view v-else class="news-img-placeholder" :style="{ background: getPlaceholderBg(index) }">
          <text class="placeholder-text">{{ item.title }}</text>
        </view>
        <view class="news-info">
          <text class="news-title">{{ item.title }}</text>
          <text class="news-desc">{{ item.desc || '' }}</text>
          <view class="news-meta">
            <text class="news-time">{{ formatTime(item._createTime) }}</text>
            <text class="news-cate">{{ item.cateName || '公告' }}</text>
          </view>
        </view>
      </view>
    </view>

    <view class="empty" v-else-if="!loading">
      <image src="/static/empty.png" mode="aspectFit" class="empty-img"></image>
      <text class="empty-text">暂无通知</text>
    </view>
  </view>
</template>

<script>
import { newsApi } from '../../api/index'

export default {
  data() {
    return {
      list: [],
      page: 1,
      pageSize: 10,
      hasMore: true,
      keyword: '',
      loading: false,
      fixedTop: '0px',
      containerPad: '0px'
    }
  },

  onLoad() {
    this.loadData()
    const sys = uni.getSystemInfoSync()
    const pxScale = 750 / sys.windowWidth
    if (sys.platform === 'android') {
      this.fixedTop = '0px'
      this.containerPad = '112rpx'
    } else {
      const navOffset = (sys.statusBarHeight || 0) + 44
      this.fixedTop = navOffset + 'px'
      this.containerPad = (navOffset + Math.round(112 / pxScale)) + 'px'
    }
  },

  onPullDownRefresh() {
    this.page = 1
    this.hasMore = true
    this.loadData().then(() => {
      uni.stopPullDownRefresh()
    })
  },

  onReachBottom() {
    this.loadMore()
  },

  methods: {
    onSearch() {
      this.page = 1
      this.hasMore = true
      this.loadData()
    },

    getUserId() {
      const userInfo = uni.getStorageSync('userInfo')
      const token = uni.getStorageSync('token')
      return (userInfo && (userInfo.miniOpenID || userInfo.id)) || token || ''
    },

    async loadData() {
      this.loading = true
      try {
        const params = { page: this.page, pageSize: this.pageSize, user_id: this.getUserId() }
        if (this.keyword) params.keyword = this.keyword
        const res = await newsApi.getList(params)
        const data = Array.isArray(res.data) ? res.data : (res.data.list || [])
        if (this.page === 1) {
          this.list = data
        } else {
          this.list = [...this.list, ...data]
        }
        if (data.length < this.pageSize) {
          this.hasMore = false
        }
      } catch (e) {
        console.error('加载通知失败', e)
      }
      this.loading = false
    },

    loadMore() {
      if (!this.hasMore) return
      this.page++
      this.loadData()
    },

    goDetail(id) {
      uni.navigateTo({ url: `/pages/news/news_detail?id=${id}` })
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
    },

    formatTime(ts) {
      if (!ts) return ''
      const d = new Date(ts)
      const y = d.getFullYear()
      const m = String(d.getMonth() + 1).padStart(2, '0')
      const day = String(d.getDate()).padStart(2, '0')
      return y + '-' + m + '-' + day
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.search-bar {
  position: fixed;
  left: 0;
  right: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  background-color: #f5f5f5;
  padding: 20rpx;
}

.search-input {
  flex: 1;
  height: 72rpx;
  font-size: 28rpx;
  background-color: #fff;
  border-radius: 50rpx;
  padding: 0 20rpx;
}

.search-btn {
  font-size: 28rpx;
  color: #409eff;
  padding-left: 20rpx;
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  height: 72rpx;
  font-size: 28rpx;
}

.search-btn {
  font-size: 28rpx;
  color: #409eff;
  padding-left: 20rpx;
  flex-shrink: 0;
}

.news-list {
  display: flex;
  flex-direction: column;
}
.news-list .news-item {
  margin-bottom: 20rpx;
}
.news-list .news-item:last-child {
  margin-bottom: 0;
}

.news-item {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  display: flex;
}

.news-img {
  width: 200rpx;
  height: 160rpx;
  border-radius: 12rpx;
  margin-right: 24rpx;
  flex-shrink: 0;
  overflow: hidden;
}
.news-img-inner {
  width: 100%;
  height: 100%;
}
.news-img-placeholder {
  width: 200rpx;
  height: 160rpx;
  border-radius: 12rpx;
  margin-right: 24rpx;
  flex-shrink: 0;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}
.placeholder-text {
  font-size: 32rpx;
  color: #fff;
  font-weight: bold;
  text-align: center;
  word-break: break-all;
  line-height: 1.4;
  padding: 0 10rpx;
}

.news-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.news-title {
  font-size: 30rpx;
  color: #333;
  font-weight: 500;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.news-desc {
  font-size: 26rpx;
  color: #666;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-top: 10rpx;
}

.news-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.news-time {
  font-size: 24rpx;
  color: #999;
}

.news-cate {
  font-size: 22rpx;
  color: #fb454c;
  background-color: rgba(251, 69, 76, 0.1);
  padding: 4rpx 16rpx;
  border-radius: 20rpx;
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding-top: 200rpx;
}

.empty-img {
  width: 240rpx;
  height: 240rpx;
}

.empty-text {
  font-size: 28rpx;
  color: #999;
  margin-top: 30rpx;
}
</style>
