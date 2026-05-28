<template>
  <view class="container">
    <view class="header">
      <text class="header-title">我的打卡记录</text>
    </view>

    <view class="stats" v-if="stats.total > 0">
      <view class="stat-item">
        <text class="stat-num">{{ stats.total }}</text>
        <text class="stat-label">总打卡</text>
      </view>
      <view class="stat-item">
        <text class="stat-num">{{ stats.week }}</text>
        <text class="stat-label">本周</text>
      </view>
      <view class="stat-item">
        <text class="stat-num">{{ stats.month }}</text>
        <text class="stat-label">本月</text>
      </view>
    </view>

    <scroll-view scroll-y class="content" @scrolltolower="loadMore">
      <view class="checkin-list" v-if="list.length > 0">
        <view class="checkin-card" v-for="(item, index) in list" :key="index">
          <view class="card-header">
            <text class="card-title">{{ item.enrollTitle || '打卡任务' }}</text>
            <text class="card-time">{{ item._createTime }}</text>
          </view>
          <view class="card-content" v-if="item.content">
            <text class="content-text">{{ item.content }}</text>
          </view>
          <view class="card-images" v-if="item.images && item.images.length > 0">
            <image 
              v-for="(img, imgIndex) in item.images" 
              :key="imgIndex"
              :src="img" 
              mode="aspectFill" 
              class="card-img"
              @click="previewImage(item.images, imgIndex)"
            ></image>
          </view>
          <view class="card-footer">
            <text class="card-location" v-if="item.location">{{ item.location }}</text>
          </view>
        </view>
      </view>

      <view class="empty" v-else>
        <image src="/static/empty.png" mode="aspectFit" class="empty-img"></image>
        <text class="empty-text">暂无打卡记录</text>
        <view class="go-enroll" @click="goEnroll">去打卡</view>
      </view>
    </scroll-view>
  </view>
</template>

<script>
import { enrollApi } from '../../api/index'

export default {
  data() {
    return {
      list: [],
      stats: {
        total: 0,
        week: 0,
        month: 0
      },
      page: 1,
      pageSize: 10
    }
  },

  onLoad() {
    this.loadData()
  },

  onShow() {
    this.loadData()
  },

  onPullDownRefresh() {
    this.page = 1
    this.loadData().then(() => {
      uni.stopPullDownRefresh()
    })
  },

  methods: {
    async loadData() {
      try {
        const res = await enrollApi.myUserList({ page: this.page, pageSize: this.pageSize })
        if (res.data) {
          const data = Array.isArray(res.data) ? res.data : (res.data.list || [])
          if (this.page === 1) {
            this.list = data
          } else {
            this.list = [...this.list, ...data]
          }
        }
      } catch (e) {
        console.error('加载打卡记录失败', e)
      }
    },

    loadMore() {
      this.page++
      this.loadData()
    },

    previewImage(urls, current) {
      uni.previewImage({
        urls,
        current: urls[current]
      })
    },

    goEnroll() {
      uni.switchTab({ url: '/pages/enroll/enroll_index' })
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.header {
  padding: 30rpx;
  background-color: #fb454c;
}

.header-title {
  font-size: 34rpx;
  color: #fff;
  font-weight: bold;
}

.stats {
  display: flex;
  background-color: #fff;
  margin: -40rpx 20rpx 20rpx;
  border-radius: 16rpx;
  padding: 30rpx;
  box-shadow: 0 4rpx 20rpx rgba(0, 0, 0, 0.05);
  position: relative;
  z-index: 1;
}

.stat-item {
  flex: 1;
  text-align: center;
}

.stat-num {
  display: block;
  font-size: 44rpx;
  font-weight: bold;
  color: #fb454c;
}

.stat-label {
  display: block;
  font-size: 24rpx;
  color: #999;
  margin-top: 8rpx;
}

.content {
  height: calc(100vh - 300rpx);
}

.checkin-list {
  padding: 20rpx;
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.checkin-card {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}

.card-title {
  font-size: 30rpx;
  color: #333;
  font-weight: 500;
}

.card-time {
  font-size: 24rpx;
  color: #999;
}

.card-content {
  margin-bottom: 16rpx;
}

.content-text {
  font-size: 28rpx;
  color: #666;
  line-height: 1.6;
}

.card-images {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-bottom: 16rpx;
}

.card-img {
  width: 180rpx;
  height: 180rpx;
  border-radius: 8rpx;
}

.card-footer {
  border-top: 1rpx solid #f0f0f0;
  padding-top: 16rpx;
}

.card-location {
  font-size: 24rpx;
  color: #999;
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

.go-enroll {
  margin-top: 30rpx;
  background-color: #fb454c;
  color: #fff;
  padding: 16rpx 60rpx;
  border-radius: 40rpx;
  font-size: 28rpx;
}
</style>