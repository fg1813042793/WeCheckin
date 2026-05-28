<template>
  <view class="container">
    <view class="news-list" v-if="list.length > 0">
      <view class="news-item" v-for="(item, index) in list" :key="index" @click="goDetail(item.id)">
        <image :src="item.img || '/static/default.png'" mode="aspectFill" class="news-img"></image>
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
      pageSize: 10
    }
  },

  onLoad() {
    this.loadData()
  },

  onPullDownRefresh() {
    this.page = 1
    this.loadData().then(() => {
      uni.stopPullDownRefresh()
    })
  },

  onReachBottom() {
    this.loadMore()
  },

  methods: {
    async loadData() {
      try {
        const res = await newsApi.getList()
        const data = Array.isArray(res.data) ? res.data : (res.data.list || [])
        if (this.page === 1) {
          this.list = data
        } else {
          this.list = [...this.list, ...data]
        }
      } catch (e) {
        console.error('加载通知失败', e)
      }
    },

    loadMore() {
      this.page++
      this.loadData()
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
  gap: 20rpx;
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