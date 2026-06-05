<template>
  <view class="container">
    <view v-if="isLoad === null" class="load-status">内容不存在</view>
    <view v-else-if="!isLoad" class="load-status">加载中...</view>

    <view class="article" v-else>
      <text class="title">{{ item.title }}</text>
      <view class="meta">
        <text class="time">{{ item._createTime }}</text>
        <text class="cate">{{ item.cateName }}</text>
      </view>

      <view class="content-text">{{ item.content }}</view>
    </view>
  </view>
</template>

<script>
import { newsApi } from '../../api/index'

export default {
  data() {
    return {
      isLoad: false,
      item: {}
    }
  },

  onLoad(options) {
    if (options && options.id) {
      this.id = options.id
      this.loadDetail()
    }
  },

  onPullDownRefresh() {
    this.loadDetail().then(() => {
      uni.stopPullDownRefresh()
    })
  },

  methods: {
    async loadDetail() {
      try {
        const res = await newsApi.detail(this.id)
        if (!res.data) {
          this.isLoad = null
          return
        }
        this.item = res.data
        this.isLoad = true
      } catch (e) {
        console.error('加载通知详情失败', e)
        this.isLoad = null
      }
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding: 30rpx;
}

.load-status {
  text-align: center;
  padding-top: 200rpx;
  font-size: 28rpx;
  color: #999;
}

.article {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 30rpx;
}

.title {
  font-size: 36rpx;
  font-weight: bold;
  color: #333;
  display: block;
  line-height: 1.5;
}

.meta {
  display: flex;
  align-items: center;
  padding: 20rpx 0;
}
.meta > * {
  margin-right: 20rpx;
}
.meta > *:last-child {
  margin-right: 0;
}

.time {
  font-size: 24rpx;
  color: #999;
}

.cate {
  font-size: 22rpx;
  color: #fb454c;
  background-color: rgba(251, 69, 76, 0.1);
  padding: 4rpx 16rpx;
  border-radius: 20rpx;
}

.text-block {
  padding: 10rpx 0;
  font-size: 28rpx;
  color: #333;
  line-height: 1.8;
}

.img-block {
  margin: 20rpx 0;
}

.content-text {
  font-size: 28rpx;
  color: #333;
  line-height: 1.8;
  word-wrap: break-word;
}
</style>
