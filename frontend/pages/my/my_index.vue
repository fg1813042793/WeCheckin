<template>
  <view class="main">
    <view class="upside upside-shadow">
      <image mode="aspectFill" class="upImg" src="/static/images/my.jpg" />
      <view class="user-bar">
        <view class="detail">
          <view class="name text-cut">{{ userInfo ? userInfo.name : '欢迎回来~~~' }}</view>
          <view class="desc">
            <text v-if="!userInfo">马上注册，使用更多功能</text>
            <text v-else class="text-cut">欢迎回来~~~</text>
          </view>
        </view>
        <view class="avatar-wrap" @click.stop="logout">
          <image v-if="userInfo && userInfo.avatar" :src="userInfo.avatar" class="avatar" mode="aspectFill" />
          <text v-else class="avatar-text">{{ (userInfo && userInfo.name || '?').charAt(0) }}</text>
        </view>
      </view>
    </view>

    <view class="down padding-project">
      <view class="comm-list menu card-project shadow-project">


        <view class="item arrow" @click="goMyCheckin">
          <view class="content">
            <text class="icon-appreciate my-icon-project"></text>
            <text class="text-black">我的打卡</text>
          </view>
        </view>

        <view class="item arrow" @click="goEvent('activity')">
          <view class="content">
            <text class="icon-activity my-icon-project"></text>
            <text class="text-black">我的活动</text>
          </view>
        </view>

        <view class="item arrow" @click="goEvent('competition')">
          <view class="content">
            <text class="icon-medal my-icon-project"></text>
            <text class="text-black">我的赛事</text>
          </view>
        </view>

        <view class="item arrow" @click="goEvent('manage')" v-if="hasEventRole">
          <view class="content">
            <text class="icon-crown my-icon-project"></text>
            <text class="text-black">赛事活动管理</text>
          </view>
        </view>

        <view class="item arrow" @click="goPage('/pages/my/my_fav')">
          <view class="content">
            <text class="icon-favor my-icon-project"></text>
            <text class="text-black">我的收藏</text>
          </view>
        </view>

        <!--<view class="item arrow" @click="goPage('/pages/my/my_foot')">
          <view class="content">
            <text class="icon-footprint my-icon-project"></text>
            <text class="text-black">历史浏览</text>
          </view>
        </view>-->
		<view class="item arrow" @click="goPage('/pages/my/my_edit')" v-if="userInfo">
		  <view class="content">
		    <text class="icon-edit my-icon-project"></text>
		    <text class="text-black">个人资料</text>
		  </view>
		</view>
      </view>

      <view class="comm-list menu card-project shadow-project">
        <view class="item arrow" @click="goAbout">
          <view class="content">
            <text class="icon-service my-icon-project"></text>
            <text class="text-black">关于我们</text>
          </view>
        </view>

        <view class="item arrow" @click="clearCache">
          <view class="content">
            <text class="icon-delete my-icon-project"></text>
            <text class="text-black">清除缓存</text>
          </view>
        </view>
      </view>

      <view class="comm-list menu card-project shadow-project" v-if="adminInfo">
        <view class="item arrow" @click="goAdmin">
          <view class="content">
            <text class="icon-settings my-icon-project"></text>
            <text class="text-red text-bold">后台管理</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import config from '../../config/index'
import { passportApi, eventApi } from '../../api/index'

export default {
  data() {
    return {
      userInfo: null,
      adminInfo: null,
      version: config.VER,
      hasEventRole: false
    }
  },

  onShow() {
    this.loadUserInfo()
    this.loadAdminInfo()
    this.loadEventRole()
  },

  methods: {
    loadAdminInfo() {
      const token = uni.getStorageSync('admin_token')
      const info = uni.getStorageSync('admin_info')
      this.adminInfo = token && info ? info : null
    },
    async loadUserInfo() {
      const local = uni.getStorageSync('userInfo')
      const token = uni.getStorageSync('token')
      if (token && local && local.id) {
        try {
          const uid = (local && (local.miniOpenID || local.id)) || token
          const res = await passportApi.getMyDetail({ user_id: uid })
          if (res.data && res.data.id) {
            this.userInfo = res.data
            uni.setStorageSync('userInfo', res.data)
            return
          }
        } catch (e) {
          // fallback to local
        }
      }
      this.userInfo = local
    },

    async loadEventRole() {
      this.hasEventRole = false
      const local = uni.getStorageSync('userInfo')
      const token = uni.getStorageSync('token')
      if (!token || !local) return
      try {
        const uid = (local && (local.miniOpenID || local.id)) || token
        const res = await eventApi.myRoles({ user_id: uid })
        if (res.data) {
          this.hasEventRole = res.data.hasOrganizer || res.data.hasAssistant || res.data.hasReferee
        }
      } catch (e) {
        console.error('加载角色失败', e)
      }
    },

    goProfile() {
      if (this.userInfo) {
        uni.navigateTo({ url: '/pages/my/my_reg' })
      } else {
        uni.navigateTo({ url: '/pages/login/login' })
      }
    },

    goMyCheckin() {
      if (!this.userInfo) {
        uni.navigateTo({ url: '/pages/login/login' })
        return
      }
      uni.navigateTo({ url: '/pages/enroll/my_user_list' })
    },

    goPage(url) {
      if (!this.userInfo) {
        uni.navigateTo({ url: '/pages/login/login' })
        return
      }
      uni.navigateTo({ url })
    },

    goEvent(type) {
      if (!this.userInfo) {
        uni.navigateTo({ url: '/pages/login/login' })
        return
      }
      if (type === 'manage') {
        uni.navigateTo({ url: '/pages/event/my_event_manage' })
      } else if (type === 'activity') {
        uni.navigateTo({ url: '/pages/my/my_activity' })
      } else if (type === 'competition') {
        uni.navigateTo({ url: '/pages/my/my_competition' })
      }
    },

    goAbout() {
      uni.navigateTo({ url: '/pages/about/about_index?key=SETUP_CONTENT_ABOUT' })
    },

    clearCache() {
      uni.showModal({
        title: '提示',
        content: '确定要清除缓存吗？',
        success: (res) => {
          if (res.confirm) {
            try {
              const keep = ['token', 'userInfo', 'admin_token', 'admin_info']
              const saved = {}
              for (const key of keep) {
                try { saved[key] = uni.getStorageSync(key) } catch (e) {}
              }
              uni.clearStorageSync()
              for (const key of keep) {
                if (saved[key] !== undefined) {
                  try { uni.setStorageSync(key, saved[key]) } catch (e) {}
                }
              }
              uni.showToast({ title: '清除成功', icon: 'success' })
            } catch (e) {
              console.error(e)
            }
          }
        }
      })
    },

    goAdmin() {
      uni.navigateTo({ url: '/pages/admin/admin_home' })
    },

    logout() {
      uni.showModal({
        title: '提示',
        content: '确定要退出登录吗？',
        success: (res) => {
          if (res.confirm) {
            uni.removeStorageSync('userInfo')
            uni.removeStorageSync('token')
            uni.showToast({ title: '已退出登录', icon: 'success' })
            setTimeout(() => {
              uni.reLaunch({ url: '/pages/login/login' })
            }, 1500)
          }
        }
      })
    }
  }
}
</script>

<style scoped>
page {
  background-color: #f5f5f5;
}

.main {
  padding-bottom: 100rpx;
}

.upside {
  width: 100%;
  height: 360rpx;
  position: relative;
  overflow: hidden;
  border-radius: 0 0 40rpx 40rpx;
}

.upside-shadow {
  box-shadow: 0 4rpx 20rpx rgba(0, 0, 0, 0.1);
}

.upImg {
  width: 100%;
  height: 100%;
}

.user-bar {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  padding: 30rpx;
  background: linear-gradient(transparent, rgba(0, 0, 0, 0.4));
}

.detail {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.name {
  font-size: 36rpx;
  color: #fff;
  font-weight: bold;
  margin-bottom: 8rpx;
}

.desc {
  font-size: 26rpx;
  color: rgba(255, 255, 255, 0.85);
}

.avatar-wrap {
  position: relative;
}

.avatar {
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  border: 4rpx solid rgba(255, 255, 255, 0.6);
  flex-shrink: 0;
}
.avatar-text {
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  background-color: #fb454c;
  color: #fff;
  font-size: 40rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border: 4rpx solid rgba(255, 255, 255, 0.6);
}

.down {
  padding: 30rpx;
}

.comm-list {
  background-color: #fff;
  border-radius: 16rpx;
  overflow: hidden;
}

.menu + .menu {
  margin-top: 20rpx;
}

.card-project {
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
}

.item {
  display: flex;
  align-items: center;
  padding: 28rpx 30rpx;
  border-bottom: 1rpx solid #f5f5f5;
}

.item:last-child {
  border-bottom: none;
}

.arrow {
  position: relative;
}

.arrow::after {
  content: '>';
  position: absolute;
  right: 30rpx;
  color: #ccc;
  font-size: 28rpx;
}

.content {
  display: flex;
  align-items: center;
}
.content .my-icon-project {
  margin-right: 20rpx;
}

.my-icon-project {
  font-size: 40rpx;
  color: #fb454c;
}

.text-black {
  font-size: 28rpx;
  color: #333;
}

.text-red {
  color: #fb454c;
}

.text-bold {
  font-weight: bold;
}

.text-grey {
  color: #999;
}

.text-normal {
  font-weight: normal;
  font-size: 24rpx;
}

.text-cut {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.shadow-project {
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
}
</style>