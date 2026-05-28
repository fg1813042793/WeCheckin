<template>
  <view class="container">
    <view class="tabs">
      <view 
        class="tab-item" 
        :class="{ active: cur === 'all' }" 
        @click="switchTab('all')"
      >全部</view>
      <view 
        class="tab-item" 
        :class="{ active: cur === 'join' }" 
        @click="switchTab('join')"
      >我参与的</view>
    </view>

    <scroll-view scroll-y class="content" @scrolltolower="loadMore">
      <view class="enroll-list" v-if="list.length > 0">
        <view class="enroll-card" v-for="(item, index) in list" :key="index" @click="goDetail(item.id)">
          <image :src="item.img || '/static/default.png'" mode="aspectFill" class="card-img"></image>
          <view class="card-body">
            <text class="card-title">{{ item.title }}</text>
            <text class="card-desc">{{ item.desc || '快来参与打卡吧！' }}</text>
            <view class="card-footer">
              <view class="card-info">
                <text class="info-item">{{ item.joinCount || 0 }}人参与</text>
                <text class="info-item">{{ item.checkinCount || 0 }}次打卡</text>
              </view>
              <view class="card-btn" v-if="!item.isJoin">立即参与</view>
              <view class="card-btn joined" v-else>已参与</view>
            </view>
          </view>
        </view>
      </view>

      <view class="empty" v-else>
        <image src="/static/empty.png" mode="aspectFit" class="empty-img"></image>
        <text class="empty-text">暂无打卡任务</text>
      </view>
    </scroll-view>

    <view class="fab" @click="goAdmin" v-if="isAdmin">
      <text class="fab-text">+</text>
    </view>
  </view>
</template>

<script>
import { enrollApi } from '../../api/index'

export default {
  data() {
    return {
      cur: 'all',
      list: [],
      page: 1,
      pageSize: 10,
      isAdmin: false
    }
  },

  onLoad() {
    this.checkAdmin()
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
    checkAdmin() {
      const userInfo = uni.getStorageSync('userInfo')
      this.isAdmin = userInfo && userInfo.role === 'admin'
    },

    switchTab(tab) {
      this.cur = tab
      this.page = 1
      this.list = []
      this.loadData()
    },

    async loadData() {
      try {
        const params = { page: this.page, pageSize: this.pageSize }
        if (this.cur === 'join') {
          const res = await enrollApi.myJoinList(params)
          this.handleListRes(res)
        } else {
          const res = await enrollApi.getList()
          this.handleListRes(res)
        }
      } catch (e) {
        console.error('加载打卡任务失败', e)
      }
    },

    handleListRes(res) {
      const data = Array.isArray(res.data) ? res.data : (res.data.list || [])
      if (this.page === 1) {
        this.list = data
      } else {
        this.list = [...this.list, ...data]
      }
    },

    loadMore() {
      this.page++
      this.loadData()
    },

    goDetail(id) {
      uni.navigateTo({ url: `/pages/enroll/enroll_detail?id=${id}` })
    },

    goAdmin() {
      uni.navigateTo({ url: '/pages/admin/admin_home' })
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.tabs {
  display: flex;
  background-color: #fff;
  padding: 20rpx;
}

.tab-item {
  flex: 1;
  text-align: center;
  padding: 16rpx 0;
  font-size: 28rpx;
  color: #666;
  position: relative;
}

.tab-item.active {
  color: #fb454c;
  font-weight: bold;
}

.tab-item.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 60rpx;
  height: 4rpx;
  background-color: #fb454c;
  border-radius: 2rpx;
}

.content {
  height: calc(100vh - 100rpx);
}

.enroll-list {
  padding: 20rpx;
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.enroll-card {
  background-color: #fff;
  border-radius: 16rpx;
  overflow: hidden;
}

.card-img {
  width: 100%;
  height: 300rpx;
}

.card-body {
  padding: 24rpx;
}

.card-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 12rpx;
}

.card-desc {
  font-size: 26rpx;
  color: #666;
  display: block;
  margin-bottom: 20rpx;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-info {
  display: flex;
  gap: 20rpx;
}

.info-item {
  font-size: 24rpx;
  color: #999;
}

.card-btn {
  background-color: #fb454c;
  color: #fff;
  padding: 12rpx 32rpx;
  border-radius: 30rpx;
  font-size: 26rpx;
}

.card-btn.joined {
  background-color: #eee;
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

.fab {
  position: fixed;
  right: 40rpx;
  bottom: 200rpx;
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  background-color: #fb454c;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4rpx 20rpx rgba(251, 69, 76, 0.4);
}

.fab-text {
  color: #fff;
  font-size: 48rpx;
  line-height: 1;
}
</style>