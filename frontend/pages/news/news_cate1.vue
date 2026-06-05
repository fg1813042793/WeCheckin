<template>
  <view class="container">
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
            <text class="news-time">{{ item._createTime }}</text>
            <text class="news-cate">{{ item.cateName || '公告' }}</text>
          </view>
        </view>
      </view>
    </view>

    <view class="empty" v-else>
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
      id: ''
    }
  },

  onLoad(options) {
    if (options && options.id) {
      this.id = options.id
      this.loadData()
    }
  },

  onReachBottom() {
    this.loadMore()
  },

  methods: {
    getUserId() {
      const userInfo = uni.getStorageSync('userInfo')
      const token = uni.getStorageSync('token')
      return (userInfo && (userInfo.miniOpenID || userInfo.id)) || token || ''
    },

    async loadData() {
      try {
        const res = await newsApi.getList({ cateId: this.id, user_id: this.getUserId() })
        const data = Array.isArray(res.data) ? res.data : (res.data.list || [])
        if (this.page === 1) {
          this.list = data
        } else {
          this.list = [...this.list, ...data]
        }
      } catch (e) {
        console.error('加载通知列表失败', e)
      }
    },

    loadMore() {
      this.page++
      this.loadData()
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

    goDetail(id) {
      uni.navigateTo({ url: `/pages/news/news_detail?id=${id}` })
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding: 20rpx;
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
