<template>
  <view class="container">
    <view class="card" v-if="info">
      <view class="avatar-section">
        <image v-if="info.pic" :src="info.pic" mode="aspectFill" class="avatar"></image>
        <text v-else class="avatar-text">{{ (info.name || '?').charAt(0) }}</text>
        <text class="avatar-name">{{ info.name }}</text>
        <text class="avatar-role" :class="info.type == 1 ? 'super' : ''">{{ info.type == 1 ? '超级管理员' : '普通管理员' }}</text>
      </view>

      <view class="info-list">
        <view class="info-row">
          <text class="label">状态</text>
          <text class="value" :class="info.status == 1 ? 'green' : 'red'">{{ info.status == 1 ? '正常' : '停用' }}</text>
        </view>
        <view class="info-row">
          <text class="label">姓名</text>
          <text class="value">{{ info.desc || '未填写' }}</text>
        </view>
        <view class="info-row">
          <text class="label">手机号</text>
          <text class="value">{{ info.phone || '未绑定' }}</text>
        </view>
        <view class="info-row">
          <text class="label">登录次数</text>
          <text class="value">{{ info.loginCnt || 0 }} 次</text>
        </view>
        <view class="info-row">
          <text class="label">最后登录</text>
          <text class="value">{{ info.loginTime ? formatTime(info.loginTime) : '-' }}</text>
        </view>
        <view class="info-row">
          <text class="label">创建时间</text>
          <text class="value">{{ info._createTime ? formatTime(info._createTime) : '-' }}</text>
        </view>
        <view class="info-row">
          <text class="label">创建IP</text>
          <text class="value">{{ info.ADMIN_ADD_IP || '-' }}</text>
        </view>
        <view class="info-row">
          <text class="label">编辑时间</text>
          <text class="value">{{ info.editTime ? formatTime(info.editTime) : '-' }}</text>
        </view>
        <view class="info-row">
          <text class="label">编辑IP</text>
          <text class="value">{{ info.ADMIN_EDIT_IP || '-' }}</text>
        </view>
      </view>
    </view>

    <view class="loading" v-else>
      <text>加载中...</text>
    </view>
  </view>
</template>

<script>
import { adminApi } from '../../../api/admin'

export default {
  data() {
    return {
      info: null
    }
  },

  onLoad(options) {
    if (options.id) {
      this.loadDetail(options.id)
    }
  },

  methods: {
    async loadDetail(id) {
      try {
        const res = await adminApi.mgrDetail(id)
        this.info = res.data
      } catch (e) {
        console.error('加载管理员详情失败', e)
      }
    },

    formatTime(ts) {
      if (!ts || ts === 0) return '-'
      const d = new Date(ts)
      return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0') + ' ' + String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0')
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

.card {
  background-color: #fff;
  border-radius: 20rpx;
  overflow: hidden;
}

.avatar-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 60rpx 0 40rpx;
  background: linear-gradient(135deg, #2499f2, #1a7ad9);
}

.avatar {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  border: 4rpx solid rgba(255,255,255,0.4);
  background-color: #f0f0f0;
}

.avatar-text {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  border: 4rpx solid rgba(255,255,255,0.4);
  background-color: rgba(255,255,255,0.2);
  color: #fff;
  font-size: 48rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-name {
  font-size: 34rpx;
  font-weight: bold;
  color: #fff;
  margin-top: 20rpx;
}

.avatar-role {
  font-size: 24rpx;
  color: rgba(255,255,255,0.8);
  margin-top: 8rpx;
  padding: 4rpx 20rpx;
  border-radius: 20rpx;
  background-color: rgba(255,255,255,0.15);
}

.avatar-role.super {
  background-color: rgba(255, 152, 0, 0.4);
  color: #fff;
}

.info-list {
  padding: 30rpx;
}

.info-row {
  display: flex;
  padding: 20rpx 0;
  border-bottom: 2rpx solid #f5f5f5;
  font-size: 28rpx;
}

.info-row:last-child {
  border-bottom: none;
}

.label {
  width: 160rpx;
  color: #999;
  flex-shrink: 0;
}

.value {
  color: #333;
  flex: 1;
}

.value.green {
  color: #52c41a;
}

.value.red {
  color: #f44336;
}

.loading {
  text-align: center;
  padding-top: 300rpx;
  font-size: 28rpx;
  color: #999;
}
</style>
