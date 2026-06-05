<template>
  <view class="container">
    <view class="header">
      <text class="header-title">参与用户（{{ userList.length }}人）</text>
    </view>

    <scroll-view scroll-y class="content" @scrolltolower="loadMore">
      <view class="list" v-if="userList.length > 0">
        <view class="user-card" v-for="item in userList" :key="item.userId" @click="showUserDetail(item)">
          <view class="user-info">
            <image v-if="item.avatar" :src="item.avatar" mode="aspectFill" class="user-avatar-img"></image>
            <text v-else class="user-avatar-text">{{ (item.userName || '?').charAt(0) }}</text>
            <view class="user-detail">
              <text class="user-name">{{ item.userName || '未知用户' }}</text>
              <text class="user-meta">打卡 {{ item.joinCnt }} 次</text>
            </view>
            <text class="user-arrow">></text>
          </view>
        </view>
      </view>

      <view class="empty" v-else-if="!loading">
        <text class="empty-text">暂无用户参与</text>
      </view>

      <view class="loading" v-if="loading">
        <text>加载中...</text>
      </view>
    </scroll-view>

    <uni-popup ref="confirmPopup" type="center">
      <view class="confirm-card">
        <text class="confirm-title">确认删除</text>
        <text class="confirm-desc">确定将 {{ detailUser?.userName || '该用户' }} 从项目中移除吗？将同时删除其所有打卡记录。</text>
        <view class="confirm-actions">
          <view class="confirm-btn btn-cancel" @click="cancelRemove">取消</view>
          <view class="confirm-btn btn-confirm" @click="doRemove">确定</view>
        </view>
      </view>
    </uni-popup>

    <uni-popup ref="detailPopup" type="center">
      <view class="detail-card" v-if="detailUser">
        <image v-if="detailUser.avatar" :src="detailUser.avatar" mode="aspectFill" class="detail-avatar"></image>
        <text v-else class="detail-avatar-text">{{ (detailUser.userName || '?').charAt(0) }}</text>
        <text class="detail-name">{{ detailUser.userName || '未知用户' }}</text>
        <view class="detail-row">
          <text class="detail-label">手机号</text>
          <text class="detail-value">{{ detailUser.mobile || '未绑定' }}</text>
        </view>
        <view class="detail-row">
          <text class="detail-label">打卡次数</text>
          <text class="detail-value">{{ detailUser.joinCnt }} 次</text>
        </view>
        <view class="detail-row">
          <text class="detail-label">注册时间</text>
          <text class="detail-value">{{ detailUser.addTime || '-' }}</text>
        </view>
        <view class="detail-row">
          <text class="detail-label">参与时间</text>
          <text class="detail-value">{{ detailUser.firstTime ? formatTime(detailUser.firstTime) : '-' }}</text>
        </view>
        <view class="detail-section" v-if="detailUser.forms">
          <text class="detail-section-title">报名表单</text>
          <view class="form-row" v-for="(f, fi) in formDataArr(detailUser.forms)" :key="fi">
            <text class="form-label">{{ f.label }}：</text>
            <text class="form-value">{{ f.value || '-' }}</text>
          </view>
        </view>
        <view class="detail-actions">
          <view class="detail-btn btn-remove" @click="confirmRemove">删除</view>
          <view class="detail-btn btn-close" @click="closeDetail">关闭</view>
        </view>
      </view>
    </uni-popup>
  </view>
</template>

<script>
import { adminApi } from '../../../api/admin'

export default {
  data() {
    return {
      enrollId: '',
      allRecords: [],
      userList: [],
      page: 1,
      pageSize: 200,
      loading: false,
      hasMore: true,
      detailUser: null
    }
  },

  onLoad(options) {
    this.enrollId = options.enrollId
    this.loadAll()
  },

  methods: {
    formatTime(ts) {
      if (!ts || ts === 0) return '-'
      const d = new Date(ts)
      return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0') + ' ' + String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0')
    },

    async loadAll() {
      this.loading = true
      try {
        const [joinRes, userRes] = await Promise.all([
          adminApi.enrollJoinList({ enrollId: this.enrollId, page: 1, pageSize: 9999 }),
          adminApi.enrollUserList({ enrollId: this.enrollId })
        ])
        const records = Array.isArray(joinRes.data) ? joinRes.data : (joinRes.data.list || [])
        const enrollUserList = Array.isArray(userRes.data) ? userRes.data : []
        // Build enrollment forms lookup by userId
        const formsMap = {}
        const enrolledUserSet = new Set()
        for (const eu of enrollUserList) {
          formsMap[eu.miniOpenId] = eu.forms || ''
          enrolledUserSet.add(eu.miniOpenId)
        }
        const userMap = {}
        for (const r of records) {
          if (!userMap[r.userId]) {
            userMap[r.userId] = { userId: r.userId, userName: r.enrollTitle || '', joinCnt: 0, firstTime: r._createTime || '', forms: '' }
          }
          userMap[r.userId].joinCnt++
          enrolledUserSet.delete(r.userId)
        }
        // Add users who only enrolled (no check-in records)
        for (const eu of enrollUserList) {
          if (enrolledUserSet.has(eu.miniOpenId)) {
            userMap[eu.miniOpenId] = { userId: eu.miniOpenId, userName: '', joinCnt: 0, firstTime: eu._createTime || '', forms: eu.forms || '' }
          }
        }
        const users = Object.values(userMap)
        for (const u of users) {
          if (!u.forms) u.forms = formsMap[u.userId] || ''
          try {
            const detail = await adminApi.userDetail(u.userId)
            if (detail.data) {
              u.avatar = detail.data.avatar || ''
              u.mobile = detail.data.mobile || ''
              u.addTime = detail.data.addTime || ''
              u.userName = detail.data.name || ''
            }
          } catch (e) {}
        }
        this.userList = users
      } catch (e) {
        console.error('加载用户列表失败', e)
      } finally {
        this.loading = false
      }
    },

    loadMore() {
    },

    showUserDetail(item) {
      this.detailUser = item
      this.$refs.detailPopup.open()
    },

    formDataArr(formsStr) {
      if (!formsStr) return []
      try {
        const arr = typeof formsStr === 'string' ? JSON.parse(formsStr) : formsStr
        return Array.isArray(arr) ? arr : []
      } catch (e) {
        return []
      }
    },

    closeDetail() {
      this.$refs.detailPopup.close()
    },

    confirmRemove() {
      this.$refs.detailPopup.close()
      setTimeout(() => {
        this.$refs.confirmPopup.open()
      }, 200)
    },

    cancelRemove() {
      this.$refs.confirmPopup.close()
    },

    async doRemove() {
      this.$refs.confirmPopup.close()
      try {
        await adminApi.enrollRemoveUser({ enrollId: this.enrollId, userId: this.detailUser.userId })
        uni.showToast({ title: '已移除', icon: 'success' })
        this.loadAll()
      } catch (e) {
        uni.showToast({ title: '移除失败', icon: 'none' })
      }
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.header {
  background-color: #fff;
  padding: 30rpx;
  border-bottom: 2rpx solid #eee;
}

.header-title {
  font-size: 36rpx;
  font-weight: bold;
  color: #333;
}

.content {
  height: calc(100vh - 100rpx);
}
.content::-webkit-scrollbar {
  display: none;
}

.list {
  padding: 20rpx;
  display: flex;
  flex-direction: column;
}
.list .user-card {
  margin-bottom: 16rpx;
}
.list .user-card:last-child {
  margin-bottom: 0;
}

.user-card {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
}

.user-card:active {
  opacity: 0.8;
}

.user-info {
  display: flex;
  align-items: center;
}
.user-info .user-avatar-img,
.user-info .avatar-text {
  margin-right: 20rpx;
}

.user-avatar-img {
  width: 80rpx;
  height: 80rpx;
  border-radius: 50%;
  flex-shrink: 0;
  background-color: #f0f0f0;
}

.user-avatar-text {
  width: 80rpx;
  height: 80rpx;
  border-radius: 50%;
  background-color: #fb454c;
  color: #fff;
  font-size: 32rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.user-detail {
  flex: 1;
}

.user-name {
  font-size: 28rpx;
  font-weight: 500;
  color: #333;
  display: block;
}

.user-meta {
  font-size: 24rpx;
  color: #999;
  margin-top: 8rpx;
  display: block;
}

.user-arrow {
  font-size: 28rpx;
  color: #ccc;
}

.detail-card {
  background-color: #fff;
  border-radius: 20rpx;
  padding: 40rpx;
  width: 580rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.detail-avatar {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  background-color: #f0f0f0;
  margin-bottom: 20rpx;
}

.detail-avatar-text {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  background-color: #fb454c;
  color: #fff;
  font-size: 48rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20rpx;
}

.detail-name {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
  margin-bottom: 30rpx;
}

.detail-row {
  width: 100%;
  display: flex;
  justify-content: space-between;
  padding: 16rpx 0;
  border-bottom: 2rpx solid #f5f5f5;
}

.detail-label {
  font-size: 26rpx;
  color: #999;
}

.detail-value {
  font-size: 26rpx;
  color: #333;
}

.detail-section {
  width: 100%;
  margin-top: 20rpx;
  padding-top: 16rpx;
  border-top: 2rpx solid #f5f5f5;
}

.detail-section-title {
  font-size: 28rpx;
  font-weight: bold;
  color: #333;
  margin-bottom: 12rpx;
  display: block;
}

.form-row {
  display: flex;
  padding: 8rpx 0;
}

.form-label {
  font-size: 26rpx;
  color: #999;
  flex-shrink: 0;
}

.form-value {
  font-size: 26rpx;
  color: #333;
  margin-left: 8rpx;
}

.detail-actions {
  margin-top: 40rpx;
  display: flex;
}
.detail-actions .detail-btn {
  margin-right: 20rpx;
}
.detail-actions .detail-btn:last-child {
  margin-right: 0;
}

.detail-btn {
  padding: 16rpx 60rpx;
  border-radius: 40rpx;
  font-size: 28rpx;
  text-align: center;
}

.btn-remove {
  background-color: #fb454c;
  color: #fff;
}

.btn-close {
  background-color: #f0f0f0;
  color: #666;
}

.confirm-card {
  background-color: #fff;
  border-radius: 20rpx;
  padding: 40rpx;
  width: 580rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.confirm-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
  margin-bottom: 20rpx;
}

.confirm-desc {
  font-size: 26rpx;
  color: #666;
  text-align: center;
  line-height: 1.6;
  margin-bottom: 30rpx;
}

.confirm-actions {
  display: flex;
}
.confirm-actions .confirm-btn {
  margin-right: 20rpx;
}
.confirm-actions .confirm-btn:last-child {
  margin-right: 0;
}

.confirm-btn {
  padding: 16rpx 60rpx;
  border-radius: 40rpx;
  font-size: 28rpx;
  text-align: center;
}

.btn-cancel {
  background-color: #f0f0f0;
  color: #666;
}

.btn-confirm {
  background-color: #fb454c;
  color: #fff;
}

.empty {
  display: flex;
  align-items: center;
  justify-content: center;
  padding-top: 200rpx;
}

.empty-text {
  font-size: 28rpx;
  color: #999;
}

.loading {
  text-align: center;
  padding: 30rpx;
  font-size: 24rpx;
  color: #999;
}
</style>
