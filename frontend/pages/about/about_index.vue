<template>
  <view class="container">
    <view class="about-header">
      <image src="/static/logo.png" mode="aspectFit" class="logo"></image>
      <text class="app-name">MY打卡</text>
      <text class="version">版本 {{ version }}</text>
    </view>

    <view class="info-list">
      <view class="info-item">
        <text class="info-label">应用名称</text>
        <text class="info-value">MY打卡</text>
      </view>
      <view class="info-item">
        <text class="info-label">当前版本</text>
        <text class="info-value">{{ version }}</text>
      </view>
      <view class="info-item">
        <text class="info-label">更新时间</text>
        <text class="info-value">2025-05-27</text>
      </view>
    </view>

    <view class="intro">
      <text class="intro-title">应用介绍</text>
      <text class="intro-content">
        MY打卡是一款集学习、兴趣、生活、运动于一身的打卡应用，主要面向学校学生、学生家长、老师、各种兴趣团体（跑步，运动，体育，艺术等）。
        
各团体为了促活会定期开展一些闯关打卡活动；找一群志同道合的人，一起去打卡。打卡是一种习惯的养成，是当下最流行、最受欢迎的工具。
      </text>
    </view>

    <view class="feature-list">
      <text class="intro-title">主要功能</text>
      <view class="feature-item" v-for="(item, index) in features" :key="index">
        <text class="feature-icon">{{ index + 1 }}</text>
        <text class="feature-text">{{ item }}</text>
      </view>
    </view>

    <view class="platform-info">
      <text class="platform-title">支持平台</text>
      <view class="platform-list">
        <view class="platform-item">
          <image src="/static/icons/android.png" mode="aspectFit" class="platform-icon"></image>
          <text class="platform-name">Android</text>
        </view>
        <view class="platform-item">
          <image src="/static/icons/ios.png" mode="aspectFit" class="platform-icon"></image>
          <text class="platform-name">iOS</text>
        </view>
        <view class="platform-item">
          <image src="/static/icons/harmony.png" mode="aspectFit" class="platform-icon"></image>
          <text class="platform-name">鸿蒙</text>
        </view>
      </view>
    </view>

    <view class="footer">
      <text class="copyright">© 2025 MY打卡 All Rights Reserved</text>
    </view>
  </view>
</template>

<script>
import config from '../../config/index'

export default {
  data() {
    return {
      version: config.VER,
      features: [
        '每日打卡记录',
        '打卡排行榜',
        '通知公告',
        '个人中心',
        '数据统计',
        '后台管理'
      ]
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 60rpx;
}

.about-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 80rpx 40rpx 40rpx;
  background: linear-gradient(135deg, #fb454c 0%, #ff6b6b 100%);
}

.logo {
  width: 160rpx;
  height: 160rpx;
  border-radius: 32rpx;
  margin-bottom: 24rpx;
}

.app-name {
  font-size: 44rpx;
  font-weight: bold;
  color: #fff;
  margin-bottom: 12rpx;
}

.version {
  font-size: 26rpx;
  color: rgba(255, 255, 255, 0.8);
}

.info-list {
  margin: -30rpx 30rpx 30rpx;
  background-color: #fff;
  border-radius: 16rpx;
  padding: 10rpx 0;
  position: relative;
  z-index: 1;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 28rpx 30rpx;
  border-bottom: 1rpx solid #f5f5f5;
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  font-size: 28rpx;
  color: #666;
}

.info-value {
  font-size: 28rpx;
  color: #333;
}

.intro, .feature-list, .platform-info {
  margin: 20rpx 30rpx;
  background-color: #fff;
  border-radius: 16rpx;
  padding: 30rpx;
}

.intro-title, .platform-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 24rpx;
}

.intro-content {
  font-size: 28rpx;
  color: #666;
  line-height: 1.8;
  display: block;
}

.feature-item {
  display: flex;
  align-items: center;
  padding: 16rpx 0;
}

.feature-icon {
  width: 48rpx;
  height: 48rpx;
  background-color: #fb454c;
  color: #fff;
  border-radius: 50%;
  text-align: center;
  line-height: 48rpx;
  font-size: 26rpx;
  margin-right: 20rpx;
  flex-shrink: 0;
}

.feature-text {
  font-size: 28rpx;
  color: #333;
}

.platform-list {
  display: flex;
  justify-content: space-around;
  padding-top: 20rpx;
}

.platform-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.platform-icon {
  width: 72rpx;
  height: 72rpx;
  margin-bottom: 12rpx;
}

.platform-name {
  font-size: 26rpx;
  color: #666;
}

.footer {
  text-align: center;
  padding: 40rpx;
}

.copyright {
  font-size: 24rpx;
  color: #999;
}
</style>