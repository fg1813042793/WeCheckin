<template>
  <view class="container">
    <view class="toolbar">
      <text class="count" v-if="dataList.length">共 {{ dataList.length }} 位管理员</text>
      <view class="add-btn" @click="goAdd">+ 添加</view>
    </view>

    <view class="card-list">
      <view class="card" v-for="(item, idx) in dataList" :key="item.id">
        <view class="card-header">
          <image v-if="item.pic" :src="item.pic" mode="aspectFill" class="avatar"></image>
          <text v-else class="avatar-text">{{ (item.name || '?').charAt(0) }}</text>
          <view class="header-info">
            <text class="card-name">{{ item.name }}</text>
            <text class="card-role" :class="item.type == 1 ? 'super' : ''">{{ item.type == 1 ? '超级管理员' : '普通管理员' }}</text>
          </view>
          <text class="status-badge" :class="item.status == 1 ? 'active' : 'inactive'">{{ item.status == 1 ? '正常' : '停用' }}</text>
        </view>

        <view class="card-body">
          <view class="info-row">
            <text class="info-label">姓名</text>
            <text class="info-value">{{ item.desc || '未填写' }}</text>
          </view>
          <view class="info-row">
            <text class="info-label">手机</text>
            <text class="info-value">{{ item.phone || '-' }}</text>
          </view>
          <view class="info-row">
            <text class="info-label">登录</text>
            <text class="info-value">{{ item.loginCnt || 0 }}次</text>
          </view>
        </view>

        <view class="card-actions">
          <view class="action-btn detail" @click="goDetail(item.id)">查看详情</view>
          <view class="action-btn edit" @click="goEdit(item.id)">编辑</view>
          <template v-if="item.type != 1">
            <view v-if="item.status == 0" class="action-btn enable" @click="toggleStatus(item, 1)">启用</view>
            <view v-else class="action-btn disable" @click="toggleStatus(item, 0)">停用</view>
            <view class="action-btn delete" @click="delItem(item)">删除</view>
          </template>
        </view>
      </view>
    </view>

    <view class="empty" v-if="dataList.length === 0">
      <text>暂无管理员</text>
    </view>
  </view>
</template>

<script>
import { adminApi } from '../../../api/admin'

export default {
  data() {
    return {
      dataList: [],
      search: ''
    }
  },

  onShow() {
    this.loadData()
  },

  methods: {
    async loadData() {
      try {
        const res = await adminApi.mgrList({ search: this.search })
        this.dataList = Array.isArray(res.data) ? res.data : (res.data.list || [])
      } catch (e) {
        console.error('加载管理员列表失败', e)
      }
    },

    goAdd() {
      uni.navigateTo({ url: '/pages/admin/mgr/admin_mgr_add' })
    },

    goDetail(id) {
      uni.navigateTo({ url: `/pages/admin/mgr/admin_mgr_detail?id=${id}` })
    },

    goEdit(id) {
      uni.navigateTo({ url: `/pages/admin/mgr/admin_mgr_edit?id=${id}` })
    },

    async toggleStatus(item, status) {
      try {
        await adminApi.mgrStatus({ id: item.id, status })
        uni.showToast({ title: status === 1 ? '已启用' : '已停用', icon: 'success' })
        this.loadData()
      } catch (e) {
        console.error('操作失败', e)
      }
    },

    delItem(item) {
      uni.showModal({
        title: '提示',
        content: `确定要删除管理员 "${item.name}" 吗？`,
        success: async (res) => {
          if (res.confirm) {
            try {
              await adminApi.mgrDel({ id: item.id })
              uni.showToast({ title: '已删除', icon: 'success' })
              this.loadData()
            } catch (e) {
              console.error('删除失败', e)
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
  padding: 20rpx;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}

.count {
  font-size: 24rpx;
  color: #999;
}

.add-btn {
  background-color: #fb454c;
  color: #fff;
  padding: 14rpx 32rpx;
  border-radius: 36rpx;
  font-size: 26rpx;
}

.card-list {
  display: flex;
  flex-direction: column;
}
.card-list > .card {
  margin-bottom: 20rpx;
}
.card-list > .card:last-child {
  margin-bottom: 0;
}

.card {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 28rpx;
}

.card-header {
  display: flex;
  align-items: center;
  margin-bottom: 20rpx;
}

.avatar {
  width: 80rpx;
  height: 80rpx;
  border-radius: 50%;
  background-color: #f0f0f0;
  flex-shrink: 0;
}

.avatar-text {
  width: 80rpx;
  height: 80rpx;
  border-radius: 50%;
  background-color: #2499f2;
  color: #fff;
  font-size: 32rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.header-info {
  flex: 1;
  margin-left: 20rpx;
}

.card-name {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  display: block;
}

.card-role {
  font-size: 22rpx;
  color: #999;
  margin-top: 4rpx;
  display: inline-block;
}

.card-role.super {
  color: #f44336;
  font-weight: 500;
}

.status-badge {
  font-size: 22rpx;
  padding: 6rpx 18rpx;
  border-radius: 20rpx;
  flex-shrink: 0;
}

.status-badge.active {
  color: #52c41a;
  background-color: #f6ffed;
}

.status-badge.inactive {
  color: #999;
  background-color: #f5f5f5;
}

.card-body {
  padding: 16rpx 0;
  border-top: 2rpx solid #f5f5f5;
  border-bottom: 2rpx solid #f5f5f5;
}

.info-row {
  display: flex;
  padding: 10rpx 0;
  font-size: 26rpx;
}

.info-label {
  width: 100rpx;
  color: #999;
  flex-shrink: 0;
}

.info-value {
  color: #333;
}

.card-actions {
  display: flex;
  margin-top: 20rpx;
}
.card-actions .action-btn {
  margin-right: 16rpx;
}
.card-actions .action-btn:last-child {
  margin-right: 0;
}

.action-btn {
  font-size: 24rpx;
  padding: 10rpx 28rpx;
  border-radius: 24rpx;
}

.action-btn.detail {
  color: #666;
  border: 2rpx solid #ccc;
}

.action-btn.edit {
  color: #2499f2;
  border: 2rpx solid #2499f2;
}

.action-btn.enable {
  color: #52c41a;
  border: 2rpx solid #52c41a;
}

.action-btn.disable {
  color: #ff9800;
  border: 2rpx solid #ff9800;
}

.action-btn.delete {
  color: #f44336;
  border: 2rpx solid #f44336;
}

.empty {
  text-align: center;
  padding-top: 200rpx;
  color: #999;
  font-size: 28rpx;
}
</style>
