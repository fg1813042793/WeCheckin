<template>
  <view class="container" :style="{ paddingTop: containerPad }">
    <view class="toolbar" :style="{ top: fixedTop }">
      <view class="search-bar">
        <input class="search-input" v-model="keyword" placeholder="搜索用户名/手机号" @confirm="onSearch" />
        <text class="search-btn" @click="onSearch">搜索</text>
      </view>
      <view class="add-btn" v-if="hasPerm('user:add')" @click="goAddUser">+ 增加用户</view>
    </view>

    <scroll-view scroll-y class="scroll-area">
      <view class="list" v-if="list.length > 0">
        <view class="card" v-for="(item, index) in list" :key="index">
          <view class="card-header" @click="goDetail(item.id)">
            <view class="user-info">
              <image v-if="item.avatar" :src="item.avatar" mode="aspectFill" class="avatar"></image>
              <text v-else class="avatar-text">{{ (item.name || '?').charAt(0) }}</text>
              <view class="user-meta">
                <text class="user-name">{{ item.name || '未知' }}</text>
                <text class="user-phone">{{ item.mobile || '-' }}</text>
              </view>
            </view>
            <text class="tag" :class="statusClass(item.status)">{{ statusText(item.status) }}</text>
          </view>
          <view class="card-body">
            <view class="info-row" v-if="item.loginCnt">
              <text class="info-label">登录</text>
              <text class="info-value">{{ item.loginCnt }} 次 / 最后 {{ item.loginTime ? formatTime(item.loginTime) : '-' }}</text>
            </view>
            <view class="info-row">
              <text class="info-label">注册时间</text>
              <text class="info-value">{{ item.addTime ? formatTime(item.addTime) : '-' }}</text>
            </view>
            <view class="info-row" v-if="item.checkReason">
              <text class="info-label">原因</text>
              <text class="info-value red">{{ item.checkReason }}</text>
            </view>
          </view>
          <view class="card-actions" v-if="item.status !== 9">
            <view v-if="hasPerm('user:edit')" class="action-btn btn-edit" @click="goEdit(item.id)">编辑</view>
            <view v-if="item.status === 0 && hasPerm('user:edit')" class="action-group">
              <view class="action-btn btn-success" @click="changeStatus(item, 1, '审核通过')">审核通过</view>
              <view class="action-btn btn-warning" @click="showReasonModal(item, 2, '审核不过')">审核不过</view>
            </view>
            <view v-if="item.status !== 2 && hasPerm('user:edit')" class="action-btn btn-disabled" @click="showReasonModal(item, 2, '禁用')">禁用</view>
            <view v-if="item.status === 2 && hasPerm('user:edit')" class="action-btn btn-success" @click="changeStatus(item, 1, '恢复正常')">恢复正常</view>
            <view v-if="hasPerm('user:del')" class="action-btn btn-danger" @click="delItem(item)">删除</view>
          </view>
        </view>

        <view class="load-more" v-if="hasMore">
          <text>加载更多...</text>
        </view>
      </view>

      <view class="empty" v-else-if="!loading">
        <text class="empty-text">暂无用户</text>
      </view>
    </scroll-view>

    <view class="modal-mask" v-if="showModal" @click="closeModal">
      <view class="modal-content" @click.stop>
        <text class="modal-title">{{ modalTitle }}</text>
        <textarea class="modal-textarea" v-model="modalReason" placeholder="请输入原因（可选）" />
        <view class="modal-actions">
          <view class="modal-btn btn-cancel" @click="closeModal">取消</view>
          <view class="modal-btn btn-confirm" @click="confirmModal">确定</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { adminApi } from '../../../api/admin'
import adminPerms from '../../../mixins/adminPerms'

export default {
  mixins: [adminPerms],
  data() {
    return {
      list: [],
      total: 0,
      page: 1,
      pageSize: 20,
      keyword: '',
      loading: false,
      showModal: false,
      modalTitle: '',
      modalReason: '',
      modalItem: null,
      modalStatus: null,
      fixedTop: '0px',
      containerPad: '0px'
    }
  },

  computed: {
    hasMore() {
      return this.list.length < this.total
    }
  },

  onReady() {
    const sys = uni.getSystemInfoSync()
    if (sys.platform === 'android') {
      this.fixedTop = '0px'
      this.containerPad = '124rpx'
    } else {
      const navOffset = (sys.statusBarHeight || 0) + 44
      this.fixedTop = navOffset + 'px'
      const pxScale = 750 / sys.windowWidth
      this.containerPad = (navOffset + Math.round(120 / pxScale)) + 'px'
    }
  },

  onShow() {
    this.loadData()
  },

  onPullDownRefresh() {
    this.page = 1
    this.loadData().finally(() => {
      uni.stopPullDownRefresh()
    })
  },

  onReachBottom() {
    if (this.hasMore && !this.loading) {
      this.page++
      this.loadData()
    }
  },

  methods: {
    onSearch() {
      this.page = 1
      this.loadData()
    },

    async loadData() {
      if (this.loading) return
      this.loading = true
      try {
        const res = await adminApi.userList({ page: this.page, pageSize: this.pageSize, keyword: this.keyword })
        const data = Array.isArray(res.data) ? res.data : (res.data.list || [])
        this.total = res.data.total || data.length
        if (this.page === 1) {
          this.list = data
        } else {
          this.list = [...this.list, ...data]
        }
      } catch (e) {
        console.error('加载用户列表失败', e)
      } finally {
        this.loading = false
      }
    },

    formatTime(ts) {
      if (!ts) return '-'
      const d = new Date(ts)
      return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0') + ' ' + String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0')
    },

    statusText(status) {
      const map = { 0: '待审核', 1: '正常', 2: '禁用', 9: '管理员' }
      return map[status] || '未知'
    },

    statusClass(status) {
      const map = { 0: 'tag-pending', 1: 'tag-active', 2: 'tag-disabled', 9: 'tag-admin' }
      return map[status] || ''
    },

    goEdit(id) {
      uni.navigateTo({ url: `/pages/admin/user/admin_user_edit?id=${id}` })
    },

    goDetail(id) {
      uni.navigateTo({ url: `/pages/admin/user/admin_user_detail?id=${id}` })
    },

    goAddUser() {
      uni.navigateTo({ url: '/pages/admin/user/admin_user_edit' })
    },

    showReasonModal(item, status, title) {
      this.modalItem = item
      this.modalStatus = status
      this.modalTitle = title
      this.modalReason = ''
      this.showModal = true
    },

    closeModal() {
      this.showModal = false
      this.modalItem = null
      this.modalStatus = null
      this.modalReason = ''
    },

    confirmModal() {
      this.changeStatus(this.modalItem, this.modalStatus, this.modalTitle, this.modalReason)
      this.closeModal()
    },

    async changeStatus(item, status, label, reason) {
      try {
        await adminApi.userStatus({ id: item.id, status, reason: reason || '' })
        uni.showToast({ title: `${label}成功`, icon: 'success' })
        this.loadData()
      } catch (e) {
        console.error('操作失败', e)
      }
    },

    delItem(item) {
      uni.showModal({
        title: '提示',
        content: `确定要删除用户 "${item.name}" 吗？`,
        success: async (res) => {
          if (res.confirm) {
            try {
              await adminApi.userDel({ id: item.id })
              uni.showToast({ title: '已删除', icon: 'success' })
              this.loadData()
            } catch (e) {
              console.error('删除失败', e)
            }
          }
        }
      })
    },

    loadMore() {
      if (this.hasMore && !this.loading) {
        this.page++
        this.loadData()
      }
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
  display: flex;
  flex-direction: column;
}

.toolbar { position: fixed; left: 0; right: 0; z-index: 10; display: flex; align-items: center; padding: 24rpx 20rpx; background-color: #fff; }
.toolbar::before { content: ''; position: absolute; left: 0; right: 0; top: -12rpx; height: 12rpx; background-color: #f5f5f5; }
.search-bar { flex: 1; display: flex; align-items: center; }
.search-input { flex: 1; height: 64rpx; background-color: #f5f5f5; border-radius: 32rpx; padding: 0 24rpx; font-size: 26rpx; }
.search-btn { font-size: 26rpx; color: #fb454c; flex-shrink: 0; margin-left: 16rpx; }
.add-btn { background-color: #fb454c; color: #fff; padding: 14rpx 28rpx; border-radius: 32rpx; font-size: 26rpx; flex-shrink: 0; margin-left: 16rpx; }

.scroll-area {
  flex: 1;
  overflow-y: auto;
}
.list {
  padding: 20rpx;
  max-width: 750rpx;
  margin: 0 auto;
}
.card {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}

.user-info {
  display: flex;
  align-items: center;
  flex: 1;
}

.avatar {
  width: 72rpx;
  height: 72rpx;
  border-radius: 50%;
  margin-right: 16rpx;
  background-color: #f0f0f0;
}
.avatar-text {
  width: 72rpx;
  height: 72rpx;
  border-radius: 50%;
  background-color: #fb454c;
  color: #fff;
  font-size: 28rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16rpx;
  flex-shrink: 0;
}

.user-meta {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
}

.user-name {
  font-size: 30rpx;
  color: #333;
  font-weight: bold;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-phone {
  font-size: 24rpx;
  color: #999;
  margin-top: 4rpx;
}

.tag {
  font-size: 22rpx;
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
  flex-shrink: 0;
  margin-left: 12rpx;
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

.tag-admin {
  background-color: #e3f2fd;
  color: #1565c0;
}

.card-body {
  border-top: 2rpx solid #f5f5f5;
  padding-top: 16rpx;
}

.info-row {
  display: flex;
  align-items: center;
  font-size: 24rpx;
  margin-bottom: 8rpx;
}

.info-row:last-child {
  margin-bottom: 0;
}

.info-label {
  color: #999;
  width: 130rpx;
  flex-shrink: 0;
}

.info-value {
  color: #666;
  flex: 1;
  min-width: 0;
}

.info-value.red {
  color: #c62828;
}

.card-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  margin-top: 16rpx;
  padding-top: 16rpx;
  border-top: 2rpx solid #f5f5f5;
}
.card-actions .action-btn {
  margin: 4rpx 0 4rpx 8rpx;
}

.action-group {
  display: flex;
}

.action-btn {
  font-size: 22rpx;
  padding: 8rpx 20rpx;
  border-radius: 8rpx;
  text-align: center;
  color: #fff;
}

.btn-edit {
  background-color: #2499f2;
}

.btn-success {
  background-color: #2e7d32;
}

.btn-warning {
  background-color: #ff9800;
}

.btn-disabled {
  background-color: #c62828;
}

.btn-danger {
  background-color: #fb454c;
}

.empty {
  display: flex;
  justify-content: center;
  padding-top: 200rpx;
}

.empty-text {
  font-size: 28rpx;
  color: #999;
}

.load-more {
  text-align: center;
  padding: 20rpx 0;
  font-size: 24rpx;
  color: #999;
}

.modal-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
}

.modal-content {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 32rpx;
  width: 600rpx;
}

.modal-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 20rpx;
}

.modal-textarea {
  width: 100%;
  height: 200rpx;
  border: 2rpx solid #eee;
  border-radius: 8rpx;
  padding: 16rpx;
  font-size: 26rpx;
  color: #333;
  box-sizing: border-box;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 24rpx;
}
.modal-actions .modal-btn {
  margin-left: 20rpx;
}
.modal-actions .modal-btn:first-child {
  margin-left: 0;
}

.modal-btn {
  font-size: 26rpx;
  padding: 16rpx 40rpx;
  border-radius: 8rpx;
  text-align: center;
}

.btn-cancel {
  background-color: #eee;
  color: #666;
}

.btn-confirm {
  background-color: #fb454c;
  color: #fff;
}
</style>
