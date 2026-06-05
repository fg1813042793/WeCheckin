<template>
  <view class="container" :style="{ paddingTop: containerPad }">
    <view class="toolbar" :style="{ top: fixedTop }">
      <view class="search-bar">
        <input v-model="keyword" placeholder="搜索赛事活动" class="search-input" @confirm="handleSearch" />
        <text class="search-btn" @click="handleSearch">搜索</text>
      </view>
      <view class="add-btn" v-if="hasPerm('event:add')" @click="goAdd">+ 新增</view>
    </view>
    <view class="tabs" :style="{ top: tabsTop }">
      <view class="tab-item" :class="{ active: type === '' }" @click="switchType('')">全部</view>
      <view class="tab-item" :class="{ active: type === '1' }" @click="switchType('1')">活动</view>
      <view class="tab-item" :class="{ active: type === '2' }" @click="switchType('2')">赛事</view>
    </view>
    <scroll-view scroll-y class="scroll-area" @scrolltolower="loadMore">
      <view class="list" v-if="list.length > 0">
        <view class="card" v-for="(item, i) in list" :key="i">
        <view class="card-title-row">
          <text class="card-title">{{ item.title }}</text>
          <view class="tag-group">
            <text class="badge-tag badge-vouch" v-if="item.vouch">推荐</text>
            <text class="badge-tag badge-top" v-if="item.isTop">置顶</text>
            <text class="type-tag" :class="item.type === 1 ? 'tag-activity' : 'tag-competition'">{{ item.type === 1 ? '活动' : '赛事' }}</text>
          </view>
        </view>
        <text class="card-desc">{{ item.desc || '' }}</text>
        <view class="card-meta">
          <text>{{ item.userCnt || 0 }}人参与</text>
          <text v-if="item.regStartStr || item.regEndStr">{{ item.regStartStr || '-' }} ~ {{ item.regEndStr || '-' }}</text>
        </view>
        <view class="card-status">
          <text class="status-text" :class="'status-' + (item.status || 0)">{{ item.statusDesc || (item.status === 1 ? '进行中' : item.status === 2 ? '已结束' : '待开始') }}</text>
        </view>
        <view class="card-actions">
          <view v-if="hasPerm('event:edit')" class="action-btn" @click="goEdit(item)">编辑</view>
          <view v-if="hasPerm('event:list')" class="action-btn" @click="goUsers(item)">参与用户</view>
          <view v-if="hasPerm('event:edit')" class="action-btn" :class="item.vouch ? 'action-vouch-on' : 'action-vouch-off'" @click="toggleVouch(item)">{{ item.vouch ? '取消推荐' : '推荐' }}</view>
          <view v-if="hasPerm('event:edit')" class="action-btn" :class="item.isTop ? 'action-top-on' : 'action-top-off'" @click="toggleTop(item)">{{ item.isTop ? '取消置顶' : '置顶' }}</view>
          <view v-if="item.status === 0 && hasPerm('event:edit')" class="action-btn action-start" @click="toggleStatus(item, 1)">开始</view>
          <view v-if="item.status === 1 && hasPerm('event:edit')" class="action-btn action-end" @click="toggleStatus(item, 2)">结束</view>
          <view v-if="item.status === 0 && hasPerm('event:del')" class="action-btn action-del" @click="handleDel(item)">删除</view>
        </view>
      </view>
        <view class="load-more" v-if="hasMore && !loading"><text>加载更多...</text></view>
      </view>
      <view class="empty" v-else-if="!loading">
      <text class="empty-text">暂无赛事活动</text>
    </view>
    <view class="loading-more" v-if="loading"><text>加载中...</text></view>
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
      keyword: '',
      type: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      loading: false,
      containerPad: '0px',
      fixedTop: '0px',
      tabsTop: '0px'
    }
  },
  onReady() {
    const sys = uni.getSystemInfoSync()
    if (sys.platform === 'android') {
      this.fixedTop = '0px'
      this.tabsTop = '112rpx'
      this.containerPad = '200rpx'
    } else {
      const navOffset = (sys.statusBarHeight || 0) + 44
      this.fixedTop = navOffset + 'px'
      const pxScale = 750 / sys.windowWidth
      const toolbarH = Math.round(112 / pxScale)
      const tabsH = Math.round(88 / pxScale)
      this.tabsTop = (navOffset + toolbarH) + 'px'
      this.containerPad = (navOffset + toolbarH + tabsH) + 'px'
    }
  },
  onLoad() { this.loadData() },
  onPullDownRefresh() {
    this.loadData().then(() => { uni.stopPullDownRefresh() })
  },
  methods: {
    handleSearch() { this.page = 1; this.list = []; this.hasMore = true; this.loadData() },
    switchType(t) { this.type = t; this.keyword = ''; this.page = 1; this.list = []; this.hasMore = true; this.loadData() },
    async loadData() {
      if (this.loading) return
      this.loading = true
      try {
        const params = { page: this.page, pageSize: this.pageSize, keyword: this.keyword }
        if (this.type !== '') params.type = this.type
        const res = await adminApi.eventList(params)
        const data = Array.isArray(res.data) ? res.data : (res.data.list || [])
        if (this.page === 1) { this.list = data } else { this.list = [...this.list, ...data] }
        this.hasMore = data.length >= this.pageSize
      } catch (e) { console.error(e) }
      this.loading = false
    },
    loadMore() { if (this.hasMore && !this.loading) { this.page++; this.loadData() } },
    goAdd() { uni.navigateTo({ url: '/pages/admin/event/admin_event_add' }) },
    goEdit(item) { uni.navigateTo({ url: '/pages/admin/event/admin_event_edit?id=' + item.id }) },
    goUsers(item) { uni.navigateTo({ url: '/pages/admin/event/admin_event_user_list?event_id=' + item.id + '&title=' + encodeURIComponent(item.title) }) },
    async toggleStatus(item, status) {
      try {
        await adminApi.eventStatus({ id: item.id, status })
        item.status = status
        uni.showToast({ title: status === 1 ? '已开始' : '已结束', icon: 'success' })
      } catch (e) { uni.showToast({ title: '操作失败', icon: 'none' }) }
    },
    handleDel(item) {
      uni.showModal({ title: '提示', content: '确定删除该赛事活动？', success: async (r) => {
        if (r.confirm) {
          try {
            await adminApi.eventDel({ id: item.id })
            this.list = this.list.filter(v => v.id !== item.id)
            uni.showToast({ title: '已删除', icon: 'success' })
          } catch (e) { uni.showToast({ title: '删除失败', icon: 'none' }) }
        }
      }})
    },
    async toggleVouch(item) {
      try {
        const vouch = item.vouch ? '0' : '1'
        await adminApi.eventVouch({ id: item.id, vouch })
        item.vouch = vouch === '1' ? 1 : 0
        uni.showToast({ title: vouch === '1' ? '已推荐' : '已取消推荐', icon: 'success' })
      } catch (e) { uni.showToast({ title: '操作失败', icon: 'none' }) }
    },
    async toggleTop(item) {
      try {
        const top = item.isTop ? '0' : '1'
        await adminApi.eventTop({ id: item.id, top })
        item.isTop = top === '1' ? 1 : 0
        uni.showToast({ title: top === '1' ? '已置顶' : '已取消置顶', icon: 'success' })
      } catch (e) { uni.showToast({ title: '操作失败', icon: 'none' }) }
    }
  }
}
</script>

<style scoped>
.container { min-height: 100vh; background-color: #f5f5f5; display: flex; flex-direction: column; }
.toolbar { position: fixed; left: 0; right: 0; z-index: 10; display: flex; align-items: center; padding: 24rpx 20rpx; background-color: #fff; }
.toolbar::before { content: ''; position: absolute; left: 0; right: 0; top: -12rpx; height: 12rpx; background-color: #f5f5f5; }
.scroll-area { flex: 1; overflow-y: auto; }
.tabs { position: fixed; left: 0; right: 0; z-index: 9; display: flex; background-color: #fff; padding: 0 20rpx 20rpx; }
.search-bar { flex: 1; display: flex; align-items: center; }
.search-input { flex: 1; height: 64rpx; background-color: #f5f5f5; border-radius: 32rpx; padding: 0 24rpx; font-size: 26rpx; }
.search-btn { font-size: 26rpx; color: #fb454c; flex-shrink: 0; margin-left: 16rpx; }
.add-btn { background-color: #fb454c; color: #fff; padding: 14rpx 28rpx; border-radius: 32rpx; font-size: 26rpx; flex-shrink: 0; margin-left: 16rpx; }
.tab-item { flex: 1; text-align: center; font-size: 26rpx; color: #666; padding-bottom: 12rpx; }
.tab-item.active { color: #fb454c; font-weight: bold; border-bottom: 3rpx solid #fb454c; }
.list { padding: 20rpx; }
.card { background-color: #fff; border-radius: 12rpx; padding: 20rpx; margin-bottom: 16rpx; }
.card-title-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8rpx; }
.card-title { font-size: 30rpx; font-weight: bold; color: #333; flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.tag-group { display: flex; align-items: center; flex-shrink: 0; }
.type-tag { font-size: 22rpx; padding: 4rpx 12rpx; border-radius: 6rpx; margin-left: 8rpx; }
.tag-activity { background-color: #fff7e6; color: #fa8c16; }
.tag-competition { background-color: #fff1f0; color: #f5222d; }
.badge-tag { font-size: 20rpx; padding: 2rpx 10rpx; border-radius: 6rpx; margin-left: 6rpx; }
.badge-vouch { background-color: #e8f5e9; color: #2e7d32; }
.badge-top { background-color: #fff3e0; color: #e65100; }
.card-desc { font-size: 24rpx; color: #666; display: block; margin-bottom: 8rpx; display: -webkit-box; -webkit-line-clamp: 1; -webkit-box-orient: vertical; overflow: hidden; }
.card-meta { font-size: 22rpx; color: #999; display: flex; gap: 16rpx; margin-bottom: 8rpx; }
.card-status { margin-bottom: 12rpx; }
.status-text { font-size: 22rpx; padding: 4rpx 12rpx; border-radius: 6rpx; }
.status-0 { background-color: #f0f5ff; color: #2b7ef5; }
.status-1 { background-color: #fff7e6; color: #fa8c16; }
.status-2 { background-color: #f5f5f5; color: #999; }
.card-actions { display: flex; gap: 12rpx; justify-content: flex-end; flex-wrap: nowrap; overflow-x: auto; white-space: nowrap; }
.action-btn { background-color: #f5f5f5; padding: 8rpx 20rpx; border-radius: 20rpx; font-size: 24rpx; color: #333; flex-shrink: 0; }
.action-start { background-color: #f0f5ff; color: #2b7ef5; }
.action-end { background-color: #fff7e6; color: #fa8c16; }
.action-vouch-on { background-color: #e8f5e9; color: #2e7d32; }
.action-vouch-off { background-color: #f5f5f5; color: #666; }
.action-top-on { background-color: #fff3e0; color: #e65100; }
.action-top-off { background-color: #f5f5f5; color: #666; }
.action-del { color: #fb454c; }
.load-more { text-align: center; padding: 30rpx; font-size: 24rpx; color: #999; }
.empty, .loading-more { display: flex; align-items: center; justify-content: center; padding-top: 200rpx; font-size: 28rpx; color: #999; }
</style>
