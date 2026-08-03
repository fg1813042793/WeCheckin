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
        <view class="item arrow" v-for="menu in visiblePrimaryMenus" :key="menu.permissionKey" @click="openPrimaryMenu(menu)">
          <view class="content">
            <text :class="[menu.icon, 'my-icon-project']"></text>
            <text class="text-black">{{ menu.title }}</text>
          </view>
        </view>
        <view v-if="visiblePrimaryMenus.length === 0" class="empty-menu-item">暂无可用功能</view>
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
import {
  AUTH_STORAGE_KEYS,
  clearClientAuth,
  getAdminAuth,
  getClientAuth,
  getClientUserId,
  setClientUserInfo
} from '../../utils/auth'
import {
  ensureClientPermissionSnapshot,
  hasClientMenuPermission,
  openClientMenu
} from '../../utils/clientPermission'

export default {
  data() {
    return {
      userInfo: null,
      adminInfo: null,
      version: config.VER,
      hasEventRole: false,
      primaryMenus: [
        { permissionKey: 'client:menu:my_checkin', title: '我的打卡', url: '/pages/enroll/my_user_list', icon: 'icon-appreciate', requiresLogin: true },
        { permissionKey: 'client:menu:my_activity', title: '我的活动', url: '/pages/my/my_activity', icon: 'icon-activity', requiresLogin: true },
        { permissionKey: 'client:menu:my_competition', title: '我的赛事', url: '/pages/my/my_competition', icon: 'icon-medal', requiresLogin: true },
        { permissionKey: 'client:menu:event_manage', title: '赛事活动管理', url: '/pages/event/my_event_manage', icon: 'icon-crown', requiresLogin: true, requiresEventRole: true },
        { permissionKey: 'client:menu:exam', title: '在线考试', url: '/pages/exam/index', icon: 'icon-write', requiresLogin: true },
        { permissionKey: 'client:menu:survey', title: '问卷中心', url: '/pages/survey/index', icon: 'icon-form', requiresLogin: true },
        { permissionKey: 'client:menu:favorite', title: '我的收藏', url: '/pages/my/my_fav', icon: 'icon-favor', requiresLogin: true },
        { permissionKey: 'client:menu:profile', title: '个人资料', url: '/pages/my/my_edit', icon: 'icon-edit', requiresLogin: true }
      ]
    }
  },

  computed: {
    visiblePrimaryMenus() {
      return this.primaryMenus.filter((menu) => {
        if (menu.requiresLogin && !this.userInfo) return false
        if (!hasClientMenuPermission(menu.permissionKey)) return false
        if (menu.requiresEventRole && !this.hasEventRole) return false
        return true
      })
    }
  },

  async onShow() {
    await this.loadUserInfo()
    this.loadAdminInfo()
    await this.loadClientPermissions()
    this.loadEventRole()
  },

  methods: {
    loadAdminInfo() {
      const { token, info } = getAdminAuth()
      this.adminInfo = token && info ? info : null
    },
    async loadClientPermissions() {
      const { token, info } = getClientAuth()
      if (!token || !info) return
      try {
        await ensureClientPermissionSnapshot()
      } catch (e) {
        console.error('加载客户端菜单权限失败', e)
      }
    },
    async loadUserInfo() {
      const { token, info: local } = getClientAuth()
      if (token && local && local.id) {
        try {
          const uid = getClientUserId()
          const res = await passportApi.getMyDetail({ user_id: uid })
          const user = res.data && res.data.user
          if (user && user.id) {
            const domain = res.data.domain || ''
            if (user.avatar && !user.avatar.startsWith('http')) {
              user.avatar = domain + user.avatar
            }
            this.userInfo = user
            setClientUserInfo(user)
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
      const { token, info: local } = getClientAuth()
      if (!token || !local) return
      if (!hasClientMenuPermission('client:menu:event_manage')) return
      try {
        const uid = getClientUserId()
        const res = await eventApi.myRoles({ user_id: uid })
        if (res.data) {
          this.hasEventRole = res.data.hasOrganizer || res.data.hasAssistant || res.data.hasReferee
        }
      } catch (e) {
        console.error('加载角色失败', e)
      }
    },

    openPrimaryMenu(menu) {
      if (!this.userInfo) {
        uni.navigateTo({ url: '/pages/login/login' })
        return
      }
      openClientMenu(menu)
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

    goExam() {
      if (!this.userInfo) {
        uni.navigateTo({ url: '/pages/login/login' })
        return
      }
      uni.navigateTo({ url: '/pages/exam/index' })
    },

    goSurvey() {
      uni.navigateTo({ url: '/pages/survey/index' })
    },

    clearCache() {
      uni.showModal({
        title: '提示',
        content: '确定要清除缓存吗？',
        success: (res) => {
          if (res.confirm) {
            try {
              const keep = AUTH_STORAGE_KEYS
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
            clearClientAuth()
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

.empty-menu-item {
  padding: 34rpx 30rpx;
  text-align: center;
  color: #999;
  font-size: 26rpx;
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
