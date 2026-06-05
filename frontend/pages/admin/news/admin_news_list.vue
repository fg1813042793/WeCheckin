<template>
  <view class="container" :style="{ paddingTop: containerPad }">
    <view class="toolbar" :style="{ top: fixedTop }">
      <view class="search-bar">
        <input v-model="keyword" placeholder="搜索通知标题" class="search-input" @confirm="handleSearch" />
        <text class="search-btn" @click="handleSearch">搜索</text>
      </view>
      <view class="add-btn" v-if="hasPerm('news:add')" @click="goAdd">+ 添加通知</view>
    </view>

    <scroll-view scroll-y class="scroll-area" @scrolltolower="loadMore">
      <view class="list">
        <view class="news-item" v-for="(item, index) in list" :key="item.id">
        <view class="item-header">
          <text class="item-title">{{ item.title }}</text>
          <text class="item-status" :class="item.status === 1 ? 'status-on' : 'status-off'">
            {{ item.status === 1 ? '启用' : '停用' }}
          </text>
        </view>
        <view class="item-meta">
          <text class="meta-cate">{{ item.cateName || '未分类' }}</text>
          <text class="meta-order">排序: {{ item.sortOrder || 0 }}</text>
        </view>
        <view class="item-actions">
          <text v-if="hasPerm('news:edit')" class="action-btn" @click="goEdit(item.id)">编辑</text>
          <text v-if="hasPerm('news:edit')" class="action-btn" @click="toggleStatus(item)">{{ item.status === 1 ? '停用' : '启用' }}</text>
          <text v-if="hasPerm('news:del')" class="action-btn danger" @click="handleDelete(item)">删除</text>
          <view class="more-wrap">
            <text class="action-btn more-trigger" @click="toggleMore(index)">更多</text>
            <view class="more-dropdown" v-if="item._showMore">
              <text class="dropdown-item" @click="handlePreview(item)">预览</text>
              <text v-if="hasPerm('news:edit')" class="dropdown-item" @click="toggleVouch(item)">{{ item.isVouch ? '取消置顶' : '置顶' }}</text>
              <text class="dropdown-item" @click="handleQr(item)">二维码</text>
            </view>
          </view>
        </view>
        </view>
      </view>

      <view class="empty" v-if="list.length === 0 && !loading">
        <text class="empty-text">暂无通知</text>
      </view>
      <view class="loading-more" v-if="loadingMore">
        <text>加载中...</text>
      </view>
    </scroll-view>
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
      loading: false,
      loadingMore: false,
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
    this.loadData(true)
  },

  onShow() {
    this.loadData(true)
  },

  onPullDownRefresh() {
    this.loadData(true).then(() => { uni.stopPullDownRefresh() })
  },

  methods: {
    async loadData(isRefresh) {
      if (isRefresh) {
        this.page = 1
        this.loading = true
      } else {
        this.loadingMore = true
      }
      try {
        const params = { page: this.page, pageSize: this.pageSize, keyword: this.keyword }
        const res = await adminApi.newsList(params)
        const data = Array.isArray(res.data) ? res.data : (res.data.list || [])
        const list = data.map(item => {
          item._showMore = false
          return item
        })
        if (isRefresh || this.page === 1) {
          this.list = list
        } else {
          this.list = [...this.list, ...list]
        }
      } catch (e) {
        console.error('加载通知列表失败', e)
      } finally {
        this.loading = false
        this.loadingMore = false
      }
    },

    handleSearch() {
      this.loadData(true)
    },

    loadMore() {
      this.page++
      this.loadData(false)
    },

    goAdd() {
      uni.navigateTo({ url: '/pages/admin/news/admin_news_add' })
    },

    goEdit(id) {
      uni.navigateTo({ url: `/pages/admin/news/admin_news_edit?id=${id}` })
    },

    async toggleStatus(item) {
      try {
        await adminApi.newsStatus({ id: item.id, status: item.status === 1 ? 0 : 1 })
        item.status = item.status === 1 ? 0 : 1
        uni.showToast({ title: '操作成功', icon: 'success' })
      } catch (e) {
        console.error('状态更新失败', e)
      }
    },

    handleDelete(item) {
      uni.showModal({
        title: '确认删除',
        content: `确定要删除通知「${item.title}」吗？`,
        success: async (res) => {
          if (res.confirm) {
            try {
              await adminApi.newsDel({ id: item.id })
              this.list = this.list.filter(i => i.id !== item.id)
              uni.showToast({ title: '删除成功', icon: 'success' })
            } catch (e) {
              console.error('删除失败', e)
            }
          }
        }
      })
    },

    toggleMore(index) {
      this.list[index]._showMore = !this.list[index]._showMore
    },

    handlePreview(item) {
      uni.navigateTo({ url: `/pages/news/news_detail?id=${item.id}` })
    },

    async toggleVouch(item) {
      try {
        await adminApi.newsVouch({ id: item.id, isVouch: item.isVouch ? 0 : 1 })
        item.isVouch = item.isVouch ? 0 : 1
        uni.showToast({ title: item.isVouch ? '已置顶' : '已取消置顶', icon: 'success' })
      } catch (e) {
        console.error('置顶操作失败', e)
      }
    },

    handleQr(item) {
      uni.navigateTo({ url: `/pages/admin/setup/admin_setup_qr?id=${item.id}&type=news` })
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

.scroll-area {
  flex: 1;
  overflow-y: auto;
}

.list {
  padding: 20rpx;
  max-width: 750rpx;
  margin: 0 auto;
}
.search-bar { flex: 1; display: flex; align-items: center; }
.search-input { flex: 1; height: 64rpx; background-color: #f5f5f5; border-radius: 32rpx; padding: 0 24rpx; font-size: 26rpx; }
.search-btn { font-size: 26rpx; color: #fb454c; flex-shrink: 0; margin-left: 16rpx; }
.add-btn { background-color: #fb454c; color: #fff; padding: 14rpx 28rpx; border-radius: 32rpx; font-size: 26rpx; flex-shrink: 0; margin-left: 16rpx; }

.news-item {
  background-color: #fff;
  margin: 16rpx 0;
  border-radius: 16rpx;
  padding: 24rpx;
}

.item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}

.item-title {
  font-size: 30rpx;
  font-weight: 500;
  color: #333;
  flex: 1;
  margin-right: 16rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-status {
  font-size: 22rpx;
  padding: 4rpx 16rpx;
  border-radius: 20rpx;
  flex-shrink: 0;
}

.status-on {
  color: #52c41a;
  background-color: rgba(82, 196, 26, 0.1);
}

.status-off {
  color: #999;
  background-color: rgba(153, 153, 153, 0.1);
}

.item-meta {
  display: flex;
  font-size: 24rpx;
  color: #999;
  margin-bottom: 20rpx;
}
.item-meta > * {
  margin-right: 24rpx;
}
.item-meta > *:last-child {
  margin-right: 0;
}

.item-actions {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  position: relative;
}
.item-actions .action-btn {
  margin-left: 12rpx;
}

.action-btn {
  font-size: 22rpx;
  color: #fff;
  padding: 8rpx 20rpx;
  border-radius: 12rpx;
  background-color: #2499f2;
}

.action-btn.danger {
  background-color: #fb454c;
}

.more-wrap {
  position: relative;
}

.more-trigger {
  background-color: #999;
}

.more-dropdown {
  position: absolute;
  top: 56rpx;
  right: 0;
  background-color: #fff;
  border-radius: 12rpx;
  box-shadow: 0 4rpx 20rpx rgba(0, 0, 0, 0.15);
  z-index: 10;
  overflow: hidden;
}

.dropdown-item {
  display: block;
  padding: 20rpx 40rpx;
  font-size: 26rpx;
  color: #333;
  border-bottom: 1rpx solid #f0f0f0;
  white-space: nowrap;
}

.dropdown-item:last-child {
  border-bottom: none;
}

.dropdown-item:active {
  background-color: #f5f5f5;
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
  padding: 30rpx;
  font-size: 24rpx;
  color: #999;
}

.loading-more {
  text-align: center;
  padding: 30rpx;
  font-size: 24rpx;
  color: #999;
}
</style>
