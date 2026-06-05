<template>
  <view class="container" :style="{ paddingTop: containerPad }">
    <view class="toolbar" :style="{ top: fixedTop }">
      <view class="search-bar">
        <input v-model="keyword" placeholder="搜索标题" class="search-input" @confirm="handleSearch" />
        <text class="search-btn" @click="handleSearch">搜索</text>
      </view>
      <view class="add-btn" v-if="hasPerm('enroll:add')" @click="goAdd">+ 新增</view>
    </view>

    <scroll-view scroll-y class="scroll-area" @scrolltolower="loadMore">
      <view class="list" v-if="list.length > 0">
        <view class="card" v-for="(item, index) in list" :key="index">
          <view class="card-header">
            <text class="card-title">{{ item.title }}</text>
            <text class="card-status" :class="{ active: item.status === 1 }">{{ item.status === 1 ? '正常' : '停用' }}</text>
          </view>
          <view class="card-body">
            <view class="info-row">
              <text class="info-label">分类</text>
              <text class="info-value">{{ item.cateName || '未分类' }}</text>
            </view>
            <view class="info-row">
              <text class="info-label">时间范围</text>
              <text class="info-value">{{ formatTime(item.timeStart) }} ~ {{ formatTime(item.timeEnd) }}</text>
            </view>
            <view class="info-row">
              <text class="info-label">打卡天数</text>
                <text class="info-value">{{ item.dayCnt || 0 }}天</text>
            </view>
            <view class="info-row">
              <text class="info-label">参与人数</text>
              <text class="info-value">{{ item.userCnt || 0 }}人</text>
            </view>
          </view>
          <view class="card-actions">
            <view v-if="hasPerm('enroll:edit')" class="action-btn" @click="goEdit(item.id)">编辑</view>
            <view class="action-btn" @click="recordManage(item)">记录管理</view>
            <view class="action-btn" @click="statusManage(item)">状态管理</view>
            <view class="action-btn" @click="moreManage(item)">更多</view>
          </view>
        </view>
        <view class="load-more" v-if="hasMore">
          <text>加载更多...</text>
        </view>
      </view>

      <view class="empty" v-else>
        <text class="empty-text">暂无打卡项目</text>
      </view>
    </scroll-view>

    <view class="fab" v-if="hasPerm('enroll:add')" @click="goAdd">
      <text class="fab-text">+</text>
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
      keyword: '',
      page: 1,
      pageSize: 20,
      hasMore: true,
      containerPad: '0px',
      fixedTop: '0px'
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
      this.containerPad = (navOffset + Math.round(124 / pxScale)) + 'px'
    }
  },

  onLoad() {
    this.loadData()
  },

  onShow() {
    this.page = 1
    this.list = []
    this.hasMore = true
    this.loadData()
  },

  onPullDownRefresh() {
    this.page = 1
    this.list = []
    this.hasMore = true
    this.loadData().then(() => { uni.stopPullDownRefresh() })
  },

  methods: {
    handleSearch() {
      this.page = 1
      this.list = []
      this.hasMore = true
      this.loadData()
    },

    async loadData() {
      if (!this.hasMore && this.page > 1) return
      try {
        const params = { page: this.page, pageSize: this.pageSize }
        if (this.keyword) params.keyword = this.keyword
        const res = await adminApi.enrollList(params)
        const data = Array.isArray(res.data) ? res.data : (res.data.list || [])
        if (this.page === 1) {
          this.list = data
        } else {
          this.list = [...this.list, ...data]
        }
        this.hasMore = data.length >= this.pageSize
      } catch (e) {
        console.error('加载打卡项目失败', e)
      }
    },

    loadMore() {
      if (this.hasMore) {
        this.page++
        this.loadData()
      }
    },

    formatTime(ts) {
      if (!ts || ts === 0) return '-'
      const d = new Date(ts)
      return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0')
    },

    goAdd() {
      uni.navigateTo({ url: '/pages/admin/enroll/admin_enroll_add' })
    },

    goEdit(id) {
      uni.navigateTo({ url: '/pages/admin/enroll/admin_enroll_edit?id=' + id })
    },

    recordManage(item) {
      const actions = [
        { label: '打卡记录管理', perm: 'enroll:list', handler: () => uni.navigateTo({ url: '/pages/admin/enroll/admin_enroll_join_list?enrollId=' + item.id }) },
        { label: '导出Excel', perm: 'enroll:list', handler: () => uni.navigateTo({ url: '/pages/admin/enroll/admin_enroll_export?enrollId=' + item.id }) },
        { label: '清空', perm: 'enroll:del', handler: () => {
          uni.showModal({
            title: '提示',
            content: '确定要清空该项目的所有打卡记录吗？',
            success: async (confirmRes) => {
              if (confirmRes.confirm) {
                try {
                  await adminApi.enrollClear({ id: item.id })
                  uni.showToast({ title: '清空成功', icon: 'success' })
                } catch (e) {
                  console.error('清空失败', e)
                }
              }
            }
          })
        }}
      ]
      this.showFilteredActionSheet(actions)
    },

    statusManage(item) {
      const actions = [
        { label: '启用', perm: 'enroll:edit', handler: () => this.changeStatus(item, 1) },
        { label: '停用', perm: 'enroll:edit', handler: () => this.changeStatus(item, 0) },
        { label: '删除', perm: 'enroll:del', handler: () => this.deleteItem(item) }
      ]
      this.showFilteredActionSheet(actions)
    },

    showFilteredActionSheet(actions) {
      const filtered = actions.filter(a => this.hasPerm(a.perm))
      if (filtered.length === 0) {
        uni.showToast({ title: '无权限', icon: 'none' })
        return
      }
      uni.showActionSheet({
        itemList: filtered.map(a => a.label),
        success: (res) => filtered[res.tapIndex].handler()
      })
    },

    async changeStatus(item, status) {
      try {
        await adminApi.enrollStatus({ id: item.id, status })
        uni.showToast({ title: status === 1 ? '已启用' : '已停用', icon: 'success' })
        item.status = status
      } catch (e) {
        console.error('状态变更失败', e)
      }
    },

    deleteItem(item) {
      uni.showModal({
        title: '提示',
        content: '确定要删除该项目吗？',
        success: async (res) => {
          if (res.confirm) {
            try {
              await adminApi.enrollDel({ id: item.id })
              uni.showToast({ title: '删除成功', icon: 'success' })
              this.list = this.list.filter(v => v.id !== item.id)
            } catch (e) {
              console.error('删除失败', e)
            }
          }
        }
      })
    },

    moreManage(item) {
      const actions = [
        { label: '预览', perm: '', handler: () => uni.navigateTo({ url: '/pages/enroll/enroll_detail?id=' + item.id }) },
        { label: '置顶', perm: 'enroll:edit', handler: () => this.setTop(item) },
        { label: '推荐首页', perm: 'enroll:edit', handler: () => this.setVouch(item) },
        { label: '查看参与用户', perm: 'enroll:list', handler: () => uni.navigateTo({ url: '/pages/admin/enroll/admin_enroll_user_list?enrollId=' + item.id }) },
        { label: '二维码', perm: '', handler: () => this.showQr(item) }
      ]
      this.showFilteredActionSheet(actions)
    },

    async setTop(item) {
      try {
        await adminApi.enrollSort({ id: item.id })
        uni.showToast({ title: '置顶成功', icon: 'success' })
      } catch (e) {
        console.error('置顶失败', e)
      }
    },

    async setVouch(item) {
      try {
        await adminApi.enrollVouch({ id: item.id })
        uni.showToast({ title: '推荐成功', icon: 'success' })
      } catch (e) {
        console.error('推荐失败', e)
      }
    },

    showQr(item) {
      uni.navigateTo({ url: '/pages/admin/setup/admin_setup_qr?type=enroll&id=' + item.id })
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  padding-bottom: 200rpx;
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
  display: flex;
  flex-direction: column;
}
.list > .card {
  margin-bottom: 20rpx;
}
.list > .card:last-child {
  margin-bottom: 0;
}

.card {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}

.card-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
}

.card-status {
  font-size: 24rpx;
  color: #999;
  padding: 6rpx 16rpx;
  border-radius: 20rpx;
  background-color: #f5f5f5;
}

.card-status.active {
  color: #52c41a;
  background-color: #f6ffed;
}

.card-body {
  display: flex;
  flex-direction: column;
  margin-bottom: 20rpx;
}
.card-body .info-row {
  margin-bottom: 12rpx;
}
.card-body .info-row:last-child {
  margin-bottom: 0;
}

.info-row {
  display: flex;
  align-items: center;
}

.info-label {
  font-size: 26rpx;
  color: #999;
  width: 140rpx;
}

.info-value {
  font-size: 26rpx;
  color: #333;
}

.card-actions {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  border-top: 2rpx solid #f5f5f5;
  padding-top: 20rpx;
}
.card-actions .action-btn {
  margin-left: 12rpx;
  margin-bottom: 6rpx;
}

.action-btn {
  font-size: 22rpx;
  color: #fff;
  padding: 8rpx 20rpx;
  border-radius: 12rpx;
  background-color: #2499f2;
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

.load-more {
  text-align: center;
  padding: 30rpx;
  font-size: 24rpx;
  color: #999;
}

.fab {
  position: fixed;
  right: 40rpx;
  bottom: 100rpx;
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  background-color: #2e7d32;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4rpx 20rpx rgba(46, 125, 50, 0.4);
}

.fab-text {
  color: #fff;
  font-size: 76rpx;
  line-height: 1;
  margin-top: -6rpx;
}
</style>
