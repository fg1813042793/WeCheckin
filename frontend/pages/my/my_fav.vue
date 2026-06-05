<template>
  <view class="container">
    <view class="fav-list" v-if="list.length > 0">
      <view class="fav-card" v-for="(item, index) in list" :key="index" @click="goDetail(item.id)">
        <image v-if="item.img" :src="item.img" mode="aspectFill" class="card-img"></image>
        <view v-else class="card-img card-placeholder" :style="{ background: getPlaceholderBg(index) }">
          <text class="card-placeholder-text">{{ item.title }}</text>
        </view>
        <view class="card-body">
          <text class="card-title">{{ item.title }}</text>
          <text class="card-desc">{{ item.desc || '快来参与打卡吧！' }}</text>
          <view class="card-footer">
            <text class="footer-item">{{ item.joinCount || 0 }}人参与</text>
            <text class="footer-item">{{ item.checkinCount || 0 }}次打卡</text>
          </view>
        </view>
      </view>
    </view>

    <view class="empty" v-else>
      <image src="/static/empty.png" mode="aspectFit" class="empty-img"></image>
      <text class="empty-text">暂无收藏</text>
    </view>
  </view>
</template>

<script>
import { favApi } from '../../api/index'

export default {
  data() {
    return {
      list: []
    }
  },

  onLoad() {
    this.loadData()
  },

  onShow() {
    this.loadData()
  },

  onPullDownRefresh() {
    this.loadData().then(() => {
      uni.stopPullDownRefresh()
    })
  },

  methods: {
    getUserId() {
      const userInfo = uni.getStorageSync('userInfo')
      const token = uni.getStorageSync('token')
      return (userInfo && (userInfo.miniOpenID || userInfo.id)) || token || ''
    },

    async loadData() {
      try {
        const uid = this.getUserId()
        const res = await favApi.list({ user_id: uid })
        const data = Array.isArray(res.data) ? res.data : (res.data.list || [])
        this.list = data
      } catch (e) {
        console.error('加载收藏列表失败', e)
      }
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
      uni.navigateTo({ url: `/pages/enroll/enroll_detail?id=${id}` })
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.fav-list {
  padding: 20rpx;
  display: flex;
  flex-direction: column;
}
.fav-list .fav-card {
  margin-bottom: 20rpx;
}
.fav-list .fav-card:last-child {
  margin-bottom: 0;
}

.fav-card {
  background-color: #fff;
  border-radius: 16rpx;
  overflow: hidden;
  display: flex;
  padding: 24rpx;
}

.card-img {
  width: 180rpx;
  height: 180rpx;
  border-radius: 12rpx;
  margin-right: 24rpx;
  flex-shrink: 0;
}
.card-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
}
.card-placeholder-text {
  font-size: 26rpx;
  color: #fff;
  font-weight: bold;
  text-align: center;
  word-break: break-all;
  line-height: 1.3;
  padding: 0 10rpx;
}

.card-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-width: 0;
}

.card-title {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  margin-bottom: 8rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.card-desc {
  font-size: 26rpx;
  color: #999;
  margin-bottom: 12rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.card-footer {
  display: flex;
}
.card-footer .footer-item {
  margin-right: 20rpx;
}
.card-footer .footer-item:last-child {
  margin-right: 0;
}

.footer-item {
  font-size: 22rpx;
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
</style>
