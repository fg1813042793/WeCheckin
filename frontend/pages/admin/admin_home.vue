<template>
  <view class="container">
    <view class="header" v-if="adminInfo">
      <view class="admin-top">
        <view class="pic">
          <image v-if="adminInfo.pic" :src="adminInfo.pic" mode="aspectFill" class="pic-img"></image>
          <text v-else class="icon-profilefill"></text>
        </view>
        <view class="right">
          <view class="name">
            <text class="name-text">{{ adminInfo.name }}</text>
            <text v-if="adminInfo.type == 1" class="tag super">超级管理员</text>
            <text v-else class="tag normal">一般管理员</text>
          </view>
          <text class="desc">共登录{{ adminInfo.loginCnt || 0 }}次</text>
        </view>
        <view class="exit" @click="exitAdmin"><text class="icon-exit"></text></view>
      </view>
    </view>

    <view class="stats-grid">
      <view class="stat-item" v-for="(item, idx) in stats" :key="idx" @click="goStatPage(item.key)">
        <text class="stat-num">{{ item.cnt }}</text>
        <text class="stat-label">{{ item.title }}</text>
      </view>
    </view>

    <view class="section-title">
      <text class="icon-title"></text> 功能管理
    </view>

    <view class="menu-grid">
      <view class="menu-item" v-if="hasPerm('user:list')" @click="goPage('/pages/admin/user/admin_user_list')">
        <text class="icon-group_fill menu-icon green"></text>
        <text>用户管理</text>
      </view>
      <view class="menu-item" v-if="hasPerm('news:list')" @click="goPage('/pages/admin/news/admin_news_list')">
        <text class="icon-form menu-icon darkgreen"></text>
        <text>内容管理</text>
      </view>
      <view class="menu-item" v-if="hasPerm('enroll:list')" @click="goPage('/pages/admin/enroll/admin_enroll_list')">
        <text class="icon-friendadd menu-icon purple"></text>
        <text>打卡管理</text>
      </view>
      <view class="menu-item" v-if="hasPerm('event:list')" @click="goPage('/pages/admin/event/admin_event_list')">
        <text class="icon-activity menu-icon orange"></text>
        <text>赛事活动</text>
      </view>
    </view>

    <view class="comm-list">
      <view class="list-item arrow" v-if="hasPerm('setup:edit')" @click="goPage('/pages/admin/setup/admin_setup_about_list')">
        <view class="content">
          <text class="icon-edit icon-left darkgreen"></text>
          <text>编辑 - 关于我们</text>
        </view>
      </view>
      <view class="list-item arrow" v-if="hasPerm('setup:edit')" @click="goPage('/pages/admin/setup/admin_setup_qr')">
        <view class="content">
          <text class="icon-qr_code icon-left mauve"></text>
          <text>小程序二维码</text>
        </view>
      </view>
    </view>

    <view class="comm-list">
      <view v-if="hasPerm('mgr:list')" class="list-item arrow" @click="goPage('/pages/admin/mgr/admin_mgr_list')">
        <view class="content">
          <text class="icon-profile icon-left red"></text>
          <text>系统管理员管理</text>
        </view>
      </view>
      <view class="list-item arrow" @click="goPage('/pages/admin/mgr/admin_mgr_pwd')">
        <view class="content">
          <text class="icon-lock icon-left orange"></text>
          <text>修改我的管理员密码</text>
        </view>
      </view>
      <view class="list-item arrow" v-if="hasPerm('log:list')" @click="goPage('/pages/admin/mgr/admin_log_list')">
        <view class="content">
          <text class="icon-footprint icon-left brown"></text>
          <text>管理员操作日志</text>
        </view>
      </view>
    </view>

    <view class="comm-list">
      <view class="list-item arrow" v-if="hasPerm('setup:edit')" @click="showMoreSettings">
        <view class="content">
          <text class="icon-settings icon-left grey"></text>
          <text>更多设置</text>
        </view>
      </view>
    </view>

    <!--<button class="logout-btn" @click="exitAdmin">退出登录</button>-->
  </view>
</template>

<script>
import { adminApi } from '../../api/admin'

export default {
  data() {
    return {
      adminInfo: null,
      perms: [],
      stats: [
        { key: 'userCnt', title: '用户数', cnt: 0 },
        { key: 'enrollCnt', title: '打卡项目', cnt: 0 },
        { key: 'newsCnt', title: '通知', cnt: 0 },
        { key: 'joinCnt', title: '今日打卡', cnt: 0 }
      ]
    }
  },

  onLoad() {
    this.checkAuth()
  },

  onShow() {
    if (!this.adminInfo) {
      this.checkAuth()
    }
  },

  onPullDownRefresh() {
    this.loadHome().then(() => {
      uni.stopPullDownRefresh()
    })
  },

  methods: {
    checkAuth() {
      const token = uni.getStorageSync('admin_token')
      if (!token) {
        uni.redirectTo({ url: '/pages/admin/admin_login' })
        return
      }
      const info = uni.getStorageSync('admin_info')
      if (info) {
        this.adminInfo = info
      }
      this.loadHome()
      this.loadPerms()
    },

    hasPerm(perm) {
      if (this.adminInfo && this.adminInfo.type == 1) return true
      return this.perms.indexOf(perm) !== -1
    },

    async loadPerms() {
      try {
        const res = await adminApi.adminPerms()
        this.perms = res.data || []
        uni.setStorageSync('admin_perms', this.perms)
      } catch (e) {
        this.perms = []
        uni.setStorageSync('admin_perms', [])
      }
    },

    async loadHome() {
      try {
        const res = await adminApi.home()
        if (res.data) {
          this.stats.forEach(s => {
            s.cnt = res.data[s.key] || 0
          })
        }
      } catch (e) {
        console.error('加载管理后台失败', e)
      }
    },

    goPage(url) {
      uni.navigateTo({ url })
    },

    goStatPage(key) {
      const map = {
        userCnt: '/pages/admin/user/admin_user_list',
        enrollCnt: '/pages/admin/enroll/admin_enroll_list',
        newsCnt: '/pages/admin/news/admin_news_list',
        joinCnt: '/pages/admin/enroll/admin_enroll_list'
      }
      if (map[key]) uni.navigateTo({ url: map[key] })
    },

    showMoreSettings() {
      uni.showActionSheet({
        itemList: ['取消所有首页推荐'],
        success: async (res) => {
          if (res.tapIndex === 0) {
            try {
              await adminApi.clearVouch()
              uni.showToast({ title: '操作成功', icon: 'success' })
            } catch (e) {
              console.error('清除推荐失败', e)
            }
          }
        }
      })
    },

    exitAdmin() {
      uni.showModal({
        title: '提示',
        content: '您确认退出?',
        success: (res) => {
          if (res.confirm) {
            uni.removeStorageSync('admin_token')
            uni.removeStorageSync('admin_info')
            const clientToken = uni.getStorageSync('token')
            if (clientToken) {
              uni.reLaunch({ url: '/pages/my/my_index' })
            } else {
              uni.reLaunch({ url: '/pages/login/login' })
            }
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
  padding-bottom: 40rpx;
}
.header {
  background-color: #2499f2;
  padding: 60rpx 30rpx 30rpx;
}
.admin-top {
  display: flex;
  align-items: center;
}
.pic {
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  background: rgba(255,255,255,0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 20rpx;
}
.pic-img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
}

.pic text {
  font-size: 56rpx;
  color: #fff;
}
.right {
  flex: 1;
}
.name {
  display: flex;
  align-items: center;
  margin-bottom: 8rpx;
}
.name-text {
  font-size: 36rpx;
  font-weight: bold;
  color: #fff;
  margin-right: 16rpx;
}
.tag {
  font-size: 20rpx;
  color: #fff;
  padding: 2rpx 16rpx;
  border-radius: 20rpx;
}
.tag.super {
  background-color: #ff9800;
}
.tag.normal {
  background-color: #4caf50;
}
.desc {
  font-size: 24rpx;
  color: rgba(255,255,255,0.8);
}
.exit {
  width: 60rpx;
  height: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
.exit text {
  font-size: 40rpx;
  color: #fff;
}
.stats-grid {
  display: flex;
  background: #fff;
  margin: 20rpx;
  border-radius: 16rpx;
  padding: 30rpx 10rpx;
}
.stat-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  border-right: 1rpx solid #f0f0f0;
}
.stat-item:last-child {
  border-right: none;
}
.stat-num {
  font-size: 40rpx;
  font-weight: bold;
  color: #333;
  margin-bottom: 8rpx;
}
.stat-label {
  font-size: 24rpx;
  color: #999;
}
.section-title {
  padding: 20rpx 20rpx 10rpx;
  font-size: 30rpx;
  color: #333;
  font-weight: bold;
}
.section-title text {
  margin-right: 8rpx;
}
.menu-grid {
  display: flex;
  background: #fff;
  margin: 0 20rpx 20rpx;
  border-radius: 16rpx;
  padding: 30rpx 10rpx;
}
.menu-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  font-size: 26rpx;
  color: #333;
}
.menu-icon {
  font-size: 50rpx;
  margin-bottom: 12rpx;
}
.menu-icon.green { color: #4caf50; }
.menu-icon.darkgreen { color: #388e3c; }
.menu-icon.purple { color: #9c27b0; }
.menu-icon.orange { color: #ff9800; }
.comm-list {
  background: #fff;
  margin: 0 20rpx 20rpx;
  border-radius: 16rpx;
  overflow: hidden;
}
.list-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 30rpx;
  border-bottom: 1rpx solid #f5f5f5;
}
.list-item:last-child {
  border-bottom: none;
}
.list-item .content {
  display: flex;
  align-items: center;
  font-size: 30rpx;
  color: #333;
}
.list-item.arrow::after {
  content: '>';
  font-size: 28rpx;
  color: #ccc;
}
.icon-left {
  font-size: 40rpx;
  margin-right: 20rpx;
}
.icon-left.darkgreen { color: #388e3c; }
.icon-left.mauve { color: #9c27b0; }
.icon-left.red { color: #f44336; }
.icon-left.orange { color: #ff9800; }
.icon-left.brown { color: #795548; }
.icon-left.grey { color: #9e9e9e; }
.logout-btn {
  margin: 40rpx 20rpx;
  height: 88rpx;
  line-height: 88rpx;
  background: #fff;
  color: #2499f2;
  text-align: center;
  border-radius: 16rpx;
  font-size: 30rpx;
  border: 2rpx solid #2499f2;
}
</style>
