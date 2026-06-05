<template>
  <view class="container">
    <view class="page-header">
      <text class="page-title">用户详情</text>
    </view>

    <view class="profile-card">
      <image v-if="detail.avatar" :src="detail.avatar" mode="aspectFill" class="avatar"></image>
      <text v-else class="avatar-text">{{ (detail.name || '?').charAt(0) }}</text>
      <text class="user-name">{{ detail.name || '未知' }}</text>
      <text class="user-status" :class="statusClass(detail.status)">{{ statusText(detail.status) }}</text>
    </view>

    <view class="info-card">
      <view class="info-row">
        <text class="info-label">顶级部门</text>
        <text class="info-value">{{ detail.topDeptNames && detail.topDeptNames.length > 0 ? detail.topDeptNames.join('、') : '-' }}</text>
      </view>
      <view class="info-row">
        <text class="info-label">所在部门</text>
        <text class="info-value">{{ detail.deptNames && detail.deptNames.length > 0 ? detail.deptNames.join('、') : '-' }}</text>
      </view>
      <view class="info-row">
        <text class="info-label">手机号</text>
        <text class="info-value">{{ detail.mobile || '-' }}</text>
      </view>
      <view class="info-row">
        <text class="info-label">注册时间</text>
        <text class="info-value">{{ detail.addTime ? formatTime(detail.addTime) : '-' }}</text>
      </view>
      <view class="info-row" v-if="detail.loginTime">
        <text class="info-label">最后登录</text>
        <text class="info-value">{{ formatTime(detail.loginTime) }}</text>
      </view>
    </view>

    <view class="section" v-if="formList.length > 0">
      <text class="section-title">用户表单数据</text>
      <view class="form-card">
        <view class="info-row" v-for="(item, index) in formList" :key="index">
          <text class="info-label">{{ item.label || item.key }}</text>
          <text class="info-value">{{ item.value || '-' }}</text>
        </view>
      </view>
    </view>

    <view class="section" v-else>
      <text class="section-title">用户表单数据</text>
      <view class="form-card">
        <text class="no-data">暂无表单数据</text>
      </view>
    </view>
  </view>
</template>

<script>
import { adminApi } from '../../../api/admin'

export default {
  data() {
    return {
      detail: {},
      formList: []
    }
  },

  onLoad(options) {
    if (options.id) {
      this.loadDetail(options.id)
    }
  },

  methods: {
    parseUserForms(raw) {
      if (!raw) return {}
      try {
        const arr = typeof raw === 'string' ? JSON.parse(raw) : raw
        if (Array.isArray(arr)) {
          const map = {}
          for (const item of arr) {
            map[item.label] = item.value
          }
          return map
        }
        return {}
      } catch (e) {
        return {}
      }
    },

    async loadDetail(id) {
      try {
        const [userRes, configRes] = await Promise.all([
          adminApi.userDetailById(id),
          adminApi.userFormFields()
        ])
        const data = userRes.data || {}
        this.detail = data

        // Build form list from config + user data
        const list = Array.isArray(configRes.data) ? configRes.data : []

        const userForms = this.parseUserForms(data.forms || data.formList || [])
        if (list.length > 0) {
          this.formList = list.map(f => ({
            label: f.label,
            value: userForms[f.label] || '-'
          }))
        } else {
          // Fallback: show raw forms
          this.formList = Array.isArray(data.forms) ? data.forms : (data.formList || [])
        }
      } catch (e) {
        console.error('加载用户详情失败', e)
      }
    },

    statusText(status) {
      const map = { 0: '待审核', 1: '正常', 2: '禁用' }
      return map[status] || '未知'
    },

    statusClass(status) {
      const map = { 0: 'tag-pending', 1: 'tag-active', 2: 'tag-disabled' }
      return map[status] || ''
    },
    formatTime(ts) {
      if (!ts || ts === 0 || ts === '0') return '-'
      const d = new Date(Number(ts))
      return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0') + ' ' + String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0')
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding: 20rpx;
}

.page-header {
  margin-bottom: 20rpx;
}

.page-title {
  font-size: 36rpx;
  font-weight: bold;
  color: #333;
}

.profile-card {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 40rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 20rpx;
}

.avatar {
  width: 160rpx;
  height: 160rpx;
  border-radius: 50%;
  margin-bottom: 20rpx;
}
.avatar-text {
  width: 160rpx;
  height: 160rpx;
  border-radius: 50%;
  background-color: #fb454c;
  color: #fff;
  font-size: 64rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20rpx;
}

.user-name {
  font-size: 36rpx;
  font-weight: bold;
  color: #333;
  margin-bottom: 12rpx;
}

.user-status {
  font-size: 24rpx;
  padding: 6rpx 20rpx;
  border-radius: 8rpx;
}

.tag-active {
  background-color: #e8f5e9;
  color: #2e7d32;
}

.tag-pending {
  background-color: #fff3e0;
  color: #e65100;
}

.tag-disabled {
  background-color: #fbe9e7;
  color: #c62828;
}

.info-card {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}

.info-row {
  display: flex;
  font-size: 28rpx;
  padding: 16rpx 0;
  border-bottom: 1rpx solid #f5f5f5;
}

.info-row:last-child {
  border-bottom: none;
}

.info-label {
  color: #999;
  width: 160rpx;
  flex-shrink: 0;
}

.info-value {
  color: #333;
  flex: 1;
}

.section {
  margin-bottom: 20rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 16rpx;
}

.form-card {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
}

.no-data {
  font-size: 26rpx;
  color: #999;
  text-align: center;
  display: block;
  padding: 40rpx 0;
}
</style>
