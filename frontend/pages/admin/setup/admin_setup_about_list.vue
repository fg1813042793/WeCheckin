<template>
  <view class="main-admin">
    <view class="admin-comm-list">
      <view class="item" v-for="(item, idx) in list" :key="idx" @click="goEdit(item)">
        <view class="header">
          <view class="left text-cut">《{{ item.title }}》</view>
          <view class="right">
            <view class="btn round margin-right-s" @click.stop="goEdit(item)"><text class="icon-edit margin-right-xxs"></text>编辑</view>
            <view class="btn round margin-right-s" @click.stop="preview(item)"><text class="icon-attention margin-right-xxs"></text>预览</view>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
export default {
  data() {
    return {
      list: [
        { title: '关于我们', key: 'SETUP_CONTENT_ABOUT' },
        { title: '用户协议', key: 'SETUP_CONTENT_AGREEMENT' },
        { title: '隐私政策', key: 'SETUP_CONTENT_PRIVACY' }
      ]
    }
  },

  methods: {
    goEdit(item) {
      uni.navigateTo({
        url: `/pages/admin/setup/admin_setup_about?key=${item.key}&title=${encodeURIComponent(item.title)}`
      })
    },

    preview(item) {
      uni.navigateTo({
        url: `/pages/about/about_index?key=${item.key}`
      })
    }
  }
}
</script>

<style scoped>
.main-admin {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding: 20rpx;
}
.admin-comm-list {
  background: #fff;
  border-radius: 16rpx;
  overflow: hidden;
}
.item {
  padding: 30rpx;
  border-bottom: 1rpx solid #f5f5f5;
}
.item:last-child {
  border-bottom: none;
}
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.left {
  font-size: 30rpx;
  color: #333;
  font-weight: bold;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.right {
  display: flex;
  flex-shrink: 0;
}
.right .btn {
  margin-left: 12rpx;
}
.btn {
  font-size: 24rpx;
  color: #2499f2;
  padding: 8rpx 20rpx;
  border: 1rpx solid #2499f2;
  border-radius: 30rpx;
  display: flex;
  align-items: center;
}
.text-cut {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
