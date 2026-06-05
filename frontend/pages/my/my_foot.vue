<template>
  <view class="container">
    <view class="foot-list" v-if="list.length > 0">
      <view class="foot-item" v-for="(item, index) in list" :key="index" @click="goDetail(item.id)">
        <view class="foot-info">
          <text class="foot-title">{{ item.title }}</text>
          <text class="foot-time">{{ item._createTime || item.time || '' }}</text>
        </view>
        <text class="foot-arrow">></text>
      </view>
    </view>

    <view class="empty" v-else>
      <image src="/static/empty.png" mode="aspectFit" class="empty-img"></image>
      <text class="empty-text">暂无浏览记录</text>
    </view>
  </view>
</template>

<script>
import { enrollApi } from '../../api/index'

export default {
  data() {
    return {
      list: []
    }
  },

  onLoad() {
    this.loadData()
  },

  onPullDownRefresh() {
    this.loadData().then(() => {
      uni.stopPullDownRefresh()
    })
  },

  methods: {
    async loadData() {
      // 浏览记录功能暂不开放
      this.list = []
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

.foot-list {
  margin: 20rpx;
  background-color: #fff;
  border-radius: 16rpx;
  overflow: hidden;
}

.foot-item {
  display: flex;
  align-items: center;
  padding: 30rpx;
  border-bottom: 1rpx solid #f5f5f5;
}

.foot-item:last-child {
  border-bottom: none;
}

.foot-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.foot-info > * {
  margin-bottom: 8rpx;
}
.foot-info > *:last-child {
  margin-bottom: 0;
}

.foot-title {
  font-size: 30rpx;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.foot-time {
  font-size: 24rpx;
  color: #999;
}

.foot-arrow {
  font-size: 28rpx;
  color: #ccc;
  margin-left: 20rpx;
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
