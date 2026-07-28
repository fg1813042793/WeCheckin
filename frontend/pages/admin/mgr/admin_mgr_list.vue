<template>
  <view class="container">
    <view class="toolbar">
      <view class="search-bar">
        <input v-model="search" placeholder="搜索管理员姓名/手机号" class="search-input" confirm-type="search" @confirm="handleSearch" />
      </view>
      <view class="action-row">
        <view class="add-btn" aria-label="创建管理员" @click="goAdd">
          <text class="add-text">创建</text>
        </view>
      </view>
    </view>

    <view class="card-list">
      <text class="list-summary" v-if="dataList.length">共 {{ dataList.length }} 位管理员</text>
      <view class="card" v-for="(item, idx) in dataList" :key="item.id">
        <view class="card-header">
          <image v-if="item.pic" :src="item.pic" mode="aspectFill" class="avatar"></image>
          <text v-else class="avatar-text">{{ (item.name || '?').charAt(0) }}</text>
          <view class="header-info">
            <text class="card-name">{{ item.name }}</text>
            <text class="card-role" :class="isSuperAdminRole(item) ? 'super' : ''">{{ item.roleName || '-' }}</text>
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
          <template v-if="!isSuperAdminRole(item)">
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
    handleSearch() {
      this.loadData()
    },

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

    isSuperAdminRole(item) {
      return item && item.roleName === '超级管理员'
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
  padding-top: 206rpx;
  box-sizing: border-box;
}

.toolbar {
  position: fixed;
  left: 0;
  right: 0;
  top: var(--window-top, 0px);
  z-index: 20;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 16rpx;
  padding: 20rpx;
  background-color: #fff;
  box-sizing: border-box;
}

.search-bar { display: flex; align-items: center; width: 100%; min-width: 0; }
.search-input { flex: 1; height: 70rpx; background-color: #f5f5f5; border-radius: 35rpx; padding: 0 24rpx; font-size: 28rpx; }
.action-row { display: flex; align-items: center; gap: 12rpx; min-height: 56rpx; }
.add-btn { height: 56rpx; line-height: 56rpx; display: flex; align-items: center; justify-content: center; background-color: #eaf3ff; color: #1677ff; padding: 0 28rpx; border-radius: 16rpx; box-shadow: none; }
.add-text { color: #1677ff; font-size: 24rpx; font-weight: 600; line-height: 56rpx; }

.card-list {
  display: flex;
  flex-direction: column;
  padding: 20rpx;
  max-width: 750rpx;
  margin: 0 auto;
  box-sizing: border-box;
}
.list-summary {
  font-size: 24rpx;
  color: #94a3b8;
  margin-bottom: 16rpx;
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
