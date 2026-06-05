<template>
  <view class="main-admin">
    <view class="search-bar">
      <input v-model="search" placeholder="搜索内容，管理员账号，姓名" class="search-input" @confirm="handleSearch" />
      <button class="btn-clear" @click="clearLogs">清空日志</button>
    </view>

    <view v-if="total > 0" class="load-info">共{{ total }}条记录</view>

    <scroll-view scroll-y class="scroll-content" @scrolltolower="loadMore">
      <view class="admin-comm-list">
        <view class="item" v-for="(item, idx) in dataList" :key="item.id">
          <view class="item-header">
            <text class="no">{{ idx + 1 }}</text>
            <text class="left text-cut">{{ typeDesc(item.type) }}操作</text>
          </view>
          <view class="info">
            <view class="info-item">
              <text class="title">操作人</text>
              <text class="mao">：</text>
              <text class="content">{{ item.adminName }} ({{ item.adminDesc }})</text>
            </view>
            <view class="info-item">
              <text class="title">操作时间</text>
              <text class="mao">：</text>
              <text class="content">{{ formatTime(item._createTime) }}</text>
            </view>
            <view class="info-item">
              <text class="title">操作内容</text>
              <text class="mao">：</text>
              <text class="content">{{ item.content }}</text>
            </view>
            <view class="info-item">
              <text class="title">IP地址</text>
              <text class="mao">：</text>
              <text class="content">{{ item.LOG_ADD_IP }}</text>
            </view>
          </view>
        </view>
      </view>

      <view class="loading-more" v-if="loadingMore">
        <text>加载中...</text>
      </view>
      <view class="no-more" v-if="!hasMore && dataList.length > 0">
        <text>没有更多了</text>
      </view>
    </scroll-view>
  </view>
</template>

<script>
import { adminApi } from '../../../api/admin'

export default {
  data() {
    return {
      dataList: [],
      total: 0,
      search: '',
      page: 1,
      pageSize: 20,
      loadingMore: false
    }
  },

  computed: {
    hasMore() {
      return this.dataList.length < this.total
    }
  },

  onShow() {
    this.page = 1
    this.dataList = []
    this.total = 0
    this.loadData()
  },

  onPullDownRefresh() {
    this.page = 1
    this.dataList = []
    this.total = 0
    this.loadData().then(() => {
      uni.stopPullDownRefresh()
    })
  },

  methods: {
    handleSearch() {
      this.page = 1
      this.dataList = []
      this.total = 0
      this.loadData()
    },

    async loadData() {
      try {
        const res = await adminApi.logList({ search: this.search, page: this.page, pageSize: this.pageSize })
        const list = Array.isArray(res.data) ? res.data : (res.data.list || [])
        this.total = res.data.total || list.length
        if (this.page === 1) {
          this.dataList = list
        } else {
          this.dataList = [...this.dataList, ...list]
        }
      } catch (e) {
        console.error('加载日志失败', e)
      }
    },

    loadMore() {
      if (this.loadingMore || !this.hasMore) return
      this.loadingMore = true
      this.page++
      this.loadData().finally(() => {
        this.loadingMore = false
      })
    },

    typeDesc(type) {
      const map = { 1: '登录', 2: '添加', 3: '删除', 4: '修改', 5: '其他' }
      return map[type] || '未知'
    },

    formatTime(ts) {
      if (!ts || ts === 0) return '-'
      const d = new Date(ts)
      return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0') + ' ' + String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0')
    },

    clearLogs() {
      uni.showModal({
        title: '警告',
        content: '确定要清空所有操作日志吗？此操作不可恢复！',
        success: async (res) => {
          if (res.confirm) {
            try {
              await adminApi.logClear()
              uni.showToast({ title: '日志已清空', icon: 'success' })
              this.dataList = []
              this.total = 0
            } catch (e) {
              console.error('清空日志失败', e)
            }
          }
        }
      })
    }
  }
}
</script>

<style scoped>
.main-admin {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding: 20rpx;
}
.search-bar {
  display: flex;
  margin-bottom: 20rpx;
}
.search-bar .search-input {
  margin-right: 16rpx;
}
.search-input {
  flex: 1;
  height: 72rpx;
  background: #fff;
  border-radius: 16rpx;
  padding: 0 20rpx;
  font-size: 26rpx;
  color: #333;
}
.btn-clear {
  height: 72rpx;
  line-height: 72rpx;
  background-color: #2499f2;
  color: #fff;
  font-size: 24rpx;
  padding: 0 24rpx;
  border-radius: 16rpx;
  border: none;
  flex-shrink: 0;
}
.load-info {
  font-size: 24rpx;
  color: #999;
  margin-bottom: 16rpx;
}
.scroll-content {
  height: calc(100vh - 180rpx);
}
.scroll-content::-webkit-scrollbar {
  display: none;
}
.admin-comm-list {
  background: #fff;
  border-radius: 16rpx;
  overflow: hidden;
}
.item {
  padding: 24rpx;
  border-bottom: 1rpx solid #f5f5f5;
}
.item:last-child {
  border-bottom: none;
}
.item-header {
  display: flex;
  align-items: center;
  margin-bottom: 12rpx;
}
.no {
  font-size: 30rpx;
  font-weight: bold;
  color: #999;
  width: 60rpx;
  flex-shrink: 0;
}
.left {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  flex: 1;
  min-width: 0;
}
.text-cut {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.info-item {
  display: flex;
  font-size: 26rpx;
  margin-bottom: 8rpx;
  color: #666;
}
.info-item .title {
  width: 120rpx;
  flex-shrink: 0;
  color: #999;
}
.info-item .mao {
  margin-right: 8rpx;
}
.info-item .content {
  flex: 1;
}
.loading-more, .no-more {
  text-align: center;
  padding: 24rpx;
  font-size: 24rpx;
  color: #999;
}
</style>
