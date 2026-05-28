<template>
  <view class="container">
    <view class="user-header" :style="{ paddingTop: statusBarHeight + 'px' }">
      <view class="user-info" v-if="userInfo">
        <image :src="userInfo.avatar || '/static/default-avatar.png'" mode="aspectFill" class="avatar"></image>
        <view class="info-content">
          <text class="user-name">{{ userInfo.name || '用户' }}</text>
          <text class="user-desc">{{ userInfo.desc || '点击完善个人信息' }}</text>
        </view>
      </view>
      <view class="user-info" v-else @click="goLogin">
        <image src="/static/default-avatar.png" mode="aspectFill" class="avatar"></image>
        <view class="info-content">
          <text class="user-name">登录/注册</text>
          <text class="user-desc">登录后享受更多功能</text>
        </view>
      </view>
    </view>

    <view class="menu-list">
      <view class="menu-item" @click="goPage('/pages/my/foot')">
        <image src="/static/icons/foot.png" mode="aspectFit" class="menu-icon"></image>
        <text class="menu-text">我的足迹</text>
        <text class="menu-arrow">></text>
      </view>
      <view class="menu-item" @click="goPage('/pages/my/fav')">
        <image src="/static/icons/fav.png" mode="aspectFit" class="menu-icon"></image>
        <text class="menu-text">我的收藏</text>
        <text class="menu-arrow">></text>
      </view>
      <view class="menu-item" @click="goPage('/pages/my/edit')">
        <image src="/static/icons/edit.png" mode="aspectFit" class="menu-icon"></image>
        <text class="menu-text">编辑资料</text>
        <text class="menu-arrow">></text>
      </view>
    </view>

    <view class="menu-list">
      <view class="menu-item" @click="goAbout">
        <image src="/static/icons/about.png" mode="aspectFit" class="menu-icon"></image>
        <text class="menu-text">关于我们</text>
        <text class="menu-arrow">></text>
      </view>
      <view class="menu-item" @click="showVersion">
        <image src="/static/icons/version.png" mode="aspectFit" class="menu-icon"></image>
        <text class="menu-text">版本信息</text>
        <text class="menu-version">{{ version }}</text>
      </view>
    </view>

    <view class="logout-btn" v-if="userInfo" @click="logout">
      退出登录
    </view>
  </view>
</template>

<script>
import config from '../../config/index'

export default {
  data() {
    return {
      userInfo: null,
      statusBarHeight: 0,
      version: config.VER
    }
  },

  onLoad() {
    const systemInfo = uni.getSystemInfoSync()
    this.statusBarHeight = systemInfo.statusBarHeight || 0
  },

  onShow() {
    this.userInfo = uni.getStorageSync('userInfo')
  },

  methods: {
    goLogin() {
      uni.navigateTo({ url: '/pages/login/login' })
    },

    goPage(url) {
      if (!this.userInfo) {
        this.goLogin()
        return
      }
      uni.navigateTo({ url })
    },

    goAbout() {
      uni.navigateTo({ url: '/pages/about/about_index' })
    },

    showVersion() {
      uni.showModal({
        title: '版本信息',
        content: `CC打卡\n当前版本：${this.version}`,
        showCancel: false
      })
    },

    logout() {
      uni.showModal({
        title: '提示',
        content: '确定要退出登录吗？',
        success: (res) => {
          if (res.confirm) {
            uni.removeStorageSync('userInfo')
            uni.removeStorageSync('token')
            this.userInfo = null
            uni.showToast({ title: '已退出登录', icon: 'success' })
          }
        }
      })
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.user-header {
  background: linear-gradient(135deg, #fb454c 0%, #ff6b6b 100%);
  padding: 160rpx 30rpx 60rpx;
}

.user-info {
  display: flex;
  align-items: center;
  padding-top: 40rpx;
}

.avatar {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  border: 4rpx solid rgba(255, 255, 255, 0.5);
}

.info-content {
  margin-left: 30rpx;
}

.user-name {
  display: block;
  font-size: 36rpx;
  color: #fff;
  font-weight: bold;
  margin-bottom: 8rpx;
}

.user-desc {
  display: block;
  font-size: 26rpx;
  color: rgba(255, 255, 255, 0.8);
}

.menu-list {
  margin: 20rpx;
  background-color: #fff;
  border-radius: 16rpx;
  overflow: hidden;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 30rpx;
  border-bottom: 1rpx solid #f5f5f5;
}

.menu-item:last-child {
  border-bottom: none;
}

.menu-icon {
  width: 44rpx;
  height: 44rpx;
  margin-right: 24rpx;
}

.menu-text {
  flex: 1;
  font-size: 30rpx;
  color: #333;
}

.menu-arrow {
  font-size: 28rpx;
  color: #ccc;
}

.menu-version {
  font-size: 26rpx;
  color: #999;
}

.logout-btn {
  margin: 60rpx 20rpx 20rpx;
  background-color: #fff;
  color: #fb454c;
  text-align: center;
  padding: 30rpx;
  border-radius: 16rpx;
  font-size: 30rpx;
}
</style>