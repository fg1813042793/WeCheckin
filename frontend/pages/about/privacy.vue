<template>
  <view class="container">
    <view class="content">
      <text class="title">隐私政策</text>
      <text class="text" v-if="!content">加载中...</text>
      <text class="text" v-else>{{ content }}</text>
    </view>
  </view>
</template>

<script>
import { get } from '../../utils/request'

export default {
  data() {
    return {
      content: ''
    }
  },
  onLoad() {
    this.loadContent()
  },
  methods: {
    async loadContent() {
      try {
        const res = await get('/home/setup_get', { params: { key: 'SETUP_CONTENT_PRIVACY' } })
        this.content = res.data || '暂无内容'
      } catch (e) {
        this.content = '加载失败'
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
.content {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 40rpx;
}
.title {
  display: block;
  font-size: 36rpx;
  font-weight: bold;
  color: #333;
  text-align: center;
  margin-bottom: 40rpx;
}
.text {
  display: block;
  font-size: 28rpx;
  color: #666;
  line-height: 1.8;
  margin-bottom: 20rpx;
}
</style>
