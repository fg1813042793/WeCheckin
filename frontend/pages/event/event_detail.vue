<template>
  <view class="container">
    <scroll-view scroll-y class="main-scroll" @scrolltolower="loadMoreDynamics" @scroll="onScroll" :scroll-top="scrollTop">
      <view v-if="info.id">
        <view v-if="info.img" class="banner"><image :src="info.img" mode="aspectFill" class="banner-img" /></view>
        <view v-else class="banner-placeholder" :style="{ background: placeholderBg }">
          <text class="banner-placeholder-text">{{ info.title }}</text>
        </view>
        <view class="detail-body">
          <view class="title-row">
            <text class="title">{{ info.title }}</text>
            <text class="type-tag" :class="info.type === 1 ? 'tag-activity' : 'tag-competition'">{{ info.type === 1 ? '活动' : '赛事' }}</text>
          </view>
          <view class="meta-list">
            <view class="meta-item"><text class="meta-label">报名时间：</text><text>{{ info.regStartStr || '-' }} ~ {{ info.regEndStr || '-' }}</text></view>
            <view class="meta-item"><text class="meta-label">活动时间：</text><text>{{ info.eventStartStr || '-' }} ~ {{ info.eventEndStr || '-' }}</text></view>
            <view class="meta-item"><text class="meta-label">参与人数：</text><text>{{ info.userCnt || 0 }}人</text></view>
            <view class="meta-item" v-if="info.location"><text class="meta-label">地点：</text><text>{{ info.location }}</text></view>
          </view>
          <view class="section"><text class="section-title">活动简介</text><text class="section-content">{{ info.desc || '暂无介绍' }}</text></view>
          <view v-if="info.rules" class="section" @click="viewRules">
            <view class="section-title-row">
              <text class="section-title">{{ info.type === 2 ? '赛事规则' : '活动规则' }}</text>
              <text class="rules-detail">详情</text>
            </view>
          </view>
          <view v-if="info.content" class="section"><text class="section-title">详细内容</text><rich-text :nodes="info.content" class="rich-content" /></view>
        </view>
        <view class="tabs">
          <view class="tab-item" :class="{ active: tab === 'dynamics' }" @click="tab = 'dynamics'">动态({{ dynamicCnt }})</view>
          <view class="tab-item" :class="{ active: tab === 'scores' }" @click="tab = 'scores'" v-if="info.type === 2">成绩</view>
        </view>
        <view v-if="tab === 'dynamics'" class="dynamics-moments">
          <view class="moment-item" v-for="(d, i) in dynamics" :key="i">
            <image v-if="d.userAvatar" :src="d.userAvatar" mode="aspectFill" class="moment-avatar" />
            <view v-else class="moment-avatar-text">{{ (d.userName || '匿')[0] }}</view>
            <view class="moment-body">
              <view class="moment-name-row">
                <text class="moment-name">{{ d.userName || '匿名' }}</text>
                <text v-if="d.title" class="moment-title">{{ d.title }}</text>
              </view>
              <text v-if="d.content" class="moment-content">{{ d.content }}</text>
              <scroll-view scroll-x class="moment-scroll-x" :show-scrollbar="false" v-if="d.imageList && d.imageList.length > 0">
                <view class="moment-images">
                  <image v-for="(img, j) in d.imageList" :key="j" :src="img" mode="aspectFill" class="moment-img" @click="previewImg(d.imageList, j)" />
                </view>
              </scroll-view>
              <scroll-view scroll-x class="moment-scroll-x" :show-scrollbar="false" v-if="d.videoList && d.videoList.length > 0">
                <view class="moment-videos">
                  <view class="moment-video-wrap" v-for="(v, k) in d.videoList" :key="k" @click="playVideo(v)">
                    <view class="moment-video-thumb" :style="{ backgroundImage: 'url(' + getVideoThumb(v) + ')' }"></view>
                    <view class="video-play-overlay"></view>
                    <text class="video-play-icon">▶</text>
                  </view>
                </view>
              </scroll-view>
              <text class="moment-time">{{ formatTimestamp(d._createTime) }}</text>
            </view>
          </view>
          <text v-if="!loading && dynamics.length === 0" class="empty-tip">暂无动态</text>
          <text v-if="loading" class="loading-tip">加载中...</text>
        </view>
        <view v-if="tab === 'scores'" class="scores-list">
          <view class="score-header"><text class="score-col-name">用户</text><text class="score-col-dept">部门</text><text class="score-col-score">成绩</text><text class="score-col-time">时间</text></view>
          <view class="score-item" v-for="(s, i) in scores" :key="i">
            <text class="score-col-name">{{ s.participantName || s.userName || '匿名' }}</text>
            <text class="score-col-dept">{{ [s.participantDept, s.participantTopDept].filter(Boolean).join(' / ') }}</text>
            <text class="score-col-score">
              <template v-if="s._parsed && s._parsed.length > 0">
                <text v-for="(ps, j) in s._parsed" :key="j" class="multi-score">{{ ps.name }}:{{ ps.score }}</text>
              </template>
              <text v-else>{{ s.score }}</text>
            </text>
            <text class="score-col-time">{{ s._createTime ? formatTimestamp(s._createTime) : s.createdAt || '' }}</text>
          </view>
          <text v-if="!loading && scores.length === 0" class="empty-tip">暂无成绩</text>
          <text v-if="loading" class="loading-tip">加载中...</text>
        </view>
      </view>
      <view class="loading-full" v-else-if="loading"><text>加载中...</text></view>
      <view class="empty-full" v-else><text>活动不存在</text></view>
    </scroll-view>
    <view class="back-top" v-if="showBackTop" @click="scrollToTop">↑</view>
    <view class="bottom-bar" :class="{ 'bottom-bar-center': info.isJoin && !(info.type === 2 && info.hasScore) }" v-if="info.id">
      <view class="bottom-info">
        <text class="bottom-status">{{ info.statusDesc || (info.isJoin ? '已报名' : '未报名') }}</text>
      </view>
      <view v-if="!info.isJoin" class="bottom-btn" @click="handleJoin">立即报名</view>
      <view v-else-if="info.type === 2 && info.hasScore" class="bottom-btn" @click="goScore">查看成绩</view>
    </view>
  </view>
</template>

<script>
import { eventApi } from '../../api/index'
export default {
  data() {
    return {
      id: '',
      info: {},
      tab: 'dynamics',
      dynamics: [],
      scores: [],
      dynamicCnt: 0,
      dynamicPage: 1,
      scorePage: 1,
      pageSize: 20,
      hasMore: true,
      loading: false,
      placeholderBg: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      showBackTop: false,
      scrollTop: 0
    }
  },
  onLoad(opts) {
    if (opts.id) { this.id = opts.id; this.loadDetail() }
  },
  onPullDownRefresh() {
    this.loadDetail()
    setTimeout(() => uni.stopPullDownRefresh(), 500)
  },
  methods: {
    getUserId() {
      const userInfo = uni.getStorageSync('userInfo')
      const token = uni.getStorageSync('token')
      return (userInfo && (userInfo.miniOpenID || userInfo.id)) || token || ''
    },
    async loadDetail() {
      this.loading = true
      try {
        const uid = this.getUserId()
        const res = await eventApi.getDetail({ id: this.id, user_id: uid })
        this.info = res.data || {}
        this.placeholderBg = this.getPlaceholderBg(0)
        this.loadDynamics()
        if (this.info.type === 2) this.loadScores()
      } catch (e) { console.error(e) }
      this.loading = false
    },
    async loadDynamics() {
      this.loading = true
      try {
        const res = await eventApi.dynamicList({ event_id: this.id, page: this.dynamicPage, pageSize: this.pageSize })
        const list = Array.isArray(res.data) ? res.data : (res.data.list || [])
        if (this.dynamicPage === 1) { this.dynamics = list } else { this.dynamics = [...this.dynamics, ...list] }
        this.dynamicCnt = res.data.total || list.length
        this.hasMore = list.length >= this.pageSize
      } catch (e) { console.error(e) }
      this.loading = false
    },
    async loadScores() {
      try {
        const res = await eventApi.scoreList({ event_id: this.id, page: this.scorePage, pageSize: this.pageSize })
        const list = Array.isArray(res.data) ? res.data : (res.data.list || [])
        const parsed = list.map(s => {
          let _parsed = []
          try { const t = JSON.parse(s.score); if (Array.isArray(t)) _parsed = t } catch {}
          return { ...s, _parsed }
        })
        if (this.scorePage === 1) { this.scores = parsed } else { this.scores = [...this.scores, ...parsed] }
      } catch (e) { console.error(e) }
    },
    loadMoreDynamics() {
      if (this.tab === 'dynamics' && this.hasMore && !this.loading) { this.dynamicPage++; this.loadDynamics() }
    },
    onScroll(e) {
      this.showBackTop = e.detail.scrollTop > 600
    },
    scrollToTop() {
      this.scrollTop = 1
      this.$nextTick(() => { this.scrollTop = 0 })
    },
    playVideo(src) {
      uni.navigateTo({ url: '/pages/event/video_play?src=' + encodeURIComponent(src) })
    },
    async handleJoin() {
      const uid = this.getUserId()
      if (!uid) { uni.showToast({ title: '请先登录', icon: 'none' }); return }
      let forms = []
      try { forms = typeof this.info.forms === 'string' ? JSON.parse(this.info.forms || '[]') : (this.info.forms || []) } catch (e) {}
      if (forms.length > 0) {
        uni.navigateTo({ url: '/pages/event/event_join_form?id=' + this.id })
        return
      }
      try {
        await eventApi.participate({ event_id: this.id, user_id: uid, forms: '[]' })
        this.info.isJoin = true
        uni.showToast({ title: '报名成功', icon: 'success' })
      } catch (e) { uni.showToast({ title: '报名失败', icon: 'none' }) }
    },
    goScore() {
      if (this.scores.length > 0) {
        const item = this.scores.find(s => s.participantId === this.getUserId() || s.userName === (uni.getStorageSync('userInfo') || {}).nickname)
        if (item) {
          let content = ''
          if (item._parsed && item._parsed.length > 0) {
            content = item._parsed.map(p => p.name + '：' + p.score).join('\n')
          } else {
            content = '成绩：' + item.score
          }
          uni.showModal({ title: '我的成绩', content })
        }
      }
    },
    previewImg(imgs, index) { uni.previewImage({ urls: imgs, current: imgs[index] }) },
    formatTimestamp(ts) {
      if (!ts) return ''
      const d = new Date(ts)
      const y = d.getFullYear()
      const m = String(d.getMonth() + 1).padStart(2, '0')
      const day = String(d.getDate()).padStart(2, '0')
      const h = String(d.getHours()).padStart(2, '0')
      const min = String(d.getMinutes()).padStart(2, '0')
      return y + '/' + m + '/' + day + ' ' + h + ':' + min
    },
    viewRules() {
      uni.showModal({
        title: this.info.type === 2 ? '赛事规则' : '活动规则',
        content: this.info.rules || '',
        showCancel: false
      })
    },
    getPlaceholderBg(index) {
      const colors = ['linear-gradient(135deg, #667eea 0%, #764ba2 100%)','linear-gradient(135deg, #f093fb 0%, #f5576c 100%)','linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)','linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)','linear-gradient(135deg, #fa709a 0%, #fee140 100%)','linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%)','linear-gradient(135deg, #fccb90 0%, #d57eeb 100%)','linear-gradient(135deg, #e0c3fc 0%, #8ec5fc 100%)','linear-gradient(135deg, #f5576c 0%, #ff758c 100%)','linear-gradient(135deg, #3b82f6 0%, #2dd4bf 100%)']
      return colors[index % colors.length]
    },
    getVideoThumb(url) {
      return url.replace(/\.[^.]+$/, '_thumb.jpg')
    }
  }
}
</script>

<style scoped>
.container { height: 100vh; background-color: #f5f5f5; padding: 0; }
.main-scroll { height: 100vh; padding-bottom: 120rpx; }
.banner { width: 100%; height: 400rpx; }
.banner-img { width: 100%; height: 100%; }
.banner-placeholder { width: 100%; height: 400rpx; display: flex; align-items: center; justify-content: center; }
.banner-placeholder-text { font-size: 48rpx; color: #fff; font-weight: bold; padding: 0 40rpx; text-align: center; }
.detail-body { padding: 24rpx 20rpx; background-color: #fff; }
.title-row { display: flex; align-items: center; margin-bottom: 16rpx; }
.title { font-size: 36rpx; font-weight: bold; color: #333; flex: 1; margin-right: 12rpx; }
.type-tag { font-size: 22rpx; padding: 4rpx 12rpx; border-radius: 8rpx; flex-shrink: 0; }
.tag-activity { background-color: #fff7e6; color: #fa8c16; }
.tag-competition { background-color: #fff1f0; color: #f5222d; }
.meta-list { background-color: #f9f9f9; border-radius: 12rpx; padding: 20rpx; margin-bottom: 24rpx; }
.meta-item { font-size: 26rpx; color: #333; margin-bottom: 8rpx; }
.meta-label { color: #999; }
.section { margin-bottom: 24rpx; }
.section-title { font-size: 30rpx; font-weight: bold; color: #333; display: block; margin-bottom: 12rpx; }
.section-title-row { display: flex; justify-content: space-between; align-items: center; padding-right: 12rpx; }
.rules-detail { font-size: 24rpx; color: #999; }
.rich-content { font-size: 26rpx; color: #666; line-height: 1.8; }
.tabs { display: flex; background-color: #fff; padding: 0 20rpx; border-top: 1rpx solid #f0f0f0; }
.tab-item { padding: 20rpx 16rpx; font-size: 28rpx; color: #666; }

.dynamics-moments { padding: 20rpx 20rpx; background-color: #fff; }
.moment-item { display: flex; margin-bottom: 40rpx; }
.moment-avatar { width: 76rpx; height: 76rpx; border-radius: 6rpx; flex-shrink: 0; }
.moment-avatar-text { width: 76rpx; height: 76rpx; border-radius: 6rpx; background-color: #fb454c; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 32rpx; flex-shrink: 0; }
.moment-body { flex: 1; margin-left: 16rpx; min-width: 0; }
.moment-name-row { display: flex; align-items: baseline; flex-wrap: wrap; gap: 16rpx; }
.moment-name { font-size: 28rpx; color: #576b95; font-weight: 500; line-height: 1.4; }
.moment-title { font-size: 28rpx; color: #333; font-weight: bold; line-height: 1.4; }
.moment-content { font-size: 28rpx; color: #333; line-height: 1.6; display: block; margin-top: 6rpx; word-break: break-all; }
.moment-scroll-x { margin-top: 10rpx; overflow: hidden; white-space: nowrap; }
.moment-images, .moment-videos { display: flex; flex-wrap: nowrap; }
.moment-img { width: 200rpx; height: 200rpx; border-radius: 4rpx; background-color: #f5f5f5; flex-shrink: 0; margin-right: 4rpx; }
.moment-video-wrap { position: relative; width: 200rpx; height: 200rpx; border-radius: 4rpx; overflow: hidden; background-color: #222; flex-shrink: 0; margin-right: 4rpx; }
.moment-video-thumb { width: 100%; height: 100%; background-size: cover; background-position: center; }
.video-play-overlay { position: absolute; top: 0; left: 0; right: 0; bottom: 0; background-color: rgba(0,0,0,0.15); }
.video-play-icon { position: absolute; top: 50%; left: 50%; transform: translate(-50%,-50%); width: 64rpx; height: 64rpx; background-color: rgba(0,0,0,0.6); color: #fff; border-radius: 50%; text-align: center; line-height: 64rpx; font-size: 36rpx; border: 2rpx solid rgba(255,255,255,0.5); }
.moment-time { font-size: 22rpx; color: #b2b2b2; display: block; margin-top: 8rpx; }
.scores-list { padding: 20rpx; }
.score-header, .score-item { display: flex; background-color: #fff; padding: 16rpx 20rpx; font-size: 26rpx; }
.score-header { font-weight: bold; color: #999; border-bottom: 1rpx solid #f0f0f0; }
.score-item { border-bottom: 1rpx solid #f5f5f5; }
.score-col-name { flex: 1; }
.score-col-dept { flex: 0 0 180rpx; text-align: center; color: #999; font-size: 22rpx; }
.score-col-score { flex: 0 0 200rpx; text-align: center; color: #fb454c; font-size: 24rpx; }
.multi-score { display: block; line-height: 1.5; }
.score-col-time { flex: 0 0 160rpx; text-align: right; color: #999; font-size: 22rpx; }
.bottom-bar { position: fixed; left: 0; right: 0; bottom: 0; display: flex; align-items: center; justify-content: space-between; padding: 16rpx 20rpx; background-color: #fff; border-top: 1rpx solid #f0f0f0; padding-bottom: calc(16rpx + env(safe-area-inset-bottom)); z-index: 10; }
.bottom-bar-center { justify-content: center; }
.back-top { position: fixed; right: 30rpx; bottom: 140rpx; width: 80rpx; height: 80rpx; background-color: rgba(0,0,0,0.5); color: #fff; border-radius: 50%; text-align: center; line-height: 80rpx; font-size: 40rpx; z-index: 11; }
.bottom-info { }
.bottom-status { font-size: 28rpx; color: #666; }
.bottom-btn { background-color: #fb454c; color: #fff; padding: 16rpx 48rpx; border-radius: 40rpx; font-size: 28rpx; }
.empty-tip, .loading-tip { display: block; text-align: center; padding: 40rpx; font-size: 26rpx; color: #999; }
.loading-full, .empty-full { display: flex; align-items: center; justify-content: center; height: 60vh; font-size: 28rpx; color: #999; }
</style>
