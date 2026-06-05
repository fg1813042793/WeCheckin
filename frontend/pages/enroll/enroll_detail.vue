<template>
  <view class="container">
    <view v-if="isLoad === null" class="load-status">内容不存在</view>
    <view v-else-if="!isLoad" class="load-status">加载中...</view>

    <block v-else>
      <view class="top">
        <image v-if="item.img" :src="item.img" mode="aspectFill" class="cover"></image>
        <view v-else class="cover cover-placeholder">
          <text class="cover-placeholder-text">{{ item.title }}</text>
        </view>
        <view class="top-right">
          <text class="title">{{ item.title }}</text>
          <text class="desc">{{ item.desc }}</text>
        </view>
      </view>

      <view class="detail-card">
        <view class="tab-bar">
          <view class="tab-item" :class="{ cur: cur === 'content' }" @click="switchTab('content')">详情介绍</view>
          <view class="tab-item" :class="{ cur: cur === 'rank' }" @click="switchTab('rank')">排行榜</view>
          <view class="tab-item" :class="{ cur: cur === 'activity' }" @click="switchTab('activity')">打卡动态</view>
        </view>

        <view v-if="cur === 'content'" class="content-tab">
          <view class="date-line">时间：{{ item.start }} ~ {{ item.end }}</view>
          <view class="desc-block" v-if="item.desc">
            <text>{{ item.desc }}</text>
          </view>
          <view v-for="(block, idx) in item.content" :key="idx">
            <view v-if="block.type === 'text'" class="text-block">
              <text>{{ block.val }}</text>
            </view>
            <view v-else-if="block.type === 'img'" class="img-block">
              <image :src="block.val" mode="widthFix" class="content-img"></image>
            </view>
          </view>
        </view>

        <view v-else-if="cur === 'rank'" class="rank-tab">
          <view v-if="!item.rankList || item.rankList.length === 0" class="empty-tip">暂无打卡记录与排行~</view>
          <view v-for="(rank, idx) in item.rankList" :key="idx" class="rank-item">
            <text class="rank-no" :class="{ top3: idx < 3 }">{{ idx + 1 < 10 ? '0' + (idx + 1) : idx + 1 }}</text>
            <image v-if="rank.userAvatar || rank.avatar" :src="rank.userAvatar || rank.avatar" mode="aspectFill" class="rank-avatar"></image>
            <text v-else class="rank-avatar-text">{{ (rank.userName || rank.name || '?').charAt(0) }}</text>
            <view class="rank-info">
              <text class="rank-name">{{ rank.userName || rank.name }}</text>
              <text class="rank-desc">打卡{{ rank.joinCount }}次 最近{{ rank.lastDay }}</text>
            </view>
          </view>
        </view>

        <view v-else class="activity-tab">
          <view class="calendar">
            <view class="calendar-header">
              <view class="cal-nav" @click="prevMonth">‹</view>
              <text class="cal-title">{{ calendarYear }}年{{ calendarMonth }}月</text>
              <view class="cal-nav" @click="nextMonth">›</view>
            </view>
            <view class="calendar-weekdays">
              <text class="weekday" v-for="(w, wi) in weekdays" :key="wi">{{ w }}</text>
            </view>
            <view class="calendar-grid">
              <view
                v-for="(cell, ci) in calendarCells"
                :key="ci"
                class="cal-cell"
                :class="{
                  'cal-empty': !cell.day,
                  'cal-checkin': cell.day && checkinSet.has(cell.day),
                  'cal-today': cell.day === todayStr,
                  'cal-selected': cell.day === activityDay
                }"
                @click="cell.day && selectDay(cell.day)"
              >
                <text v-if="cell.day" class="cal-day">{{ cell.date }}</text>
                <text v-if="cell.day && checkinSet.has(cell.day)" class="cal-dot">●</text>
              </view>
            </view>
          </view>
          <view class="timeline" v-if="activityList.length > 0">
            <view v-for="(u, ui) in activityList" :key="ui" class="timeline-item">
              <text class="timeline-dot">{{ formatTimestamp(u.addTime) }}</text>
              <view class="timeline-content">
                <view class="timeline-user">
                  <image v-if="u.userAvatar" :src="u.userAvatar" mode="aspectFill" class="timeline-avatar"></image>
                  <text v-else class="timeline-avatar-text">{{ (u.userName || u.name || '?').charAt(0) }}</text>
                  <text class="timeline-name">{{ u.userName || u.name || u.userId }}</text>
                </view>
                <view class="timeline-forms" v-if="u.formsArr && u.formsArr.length > 0">
                  <template v-for="(f, fi) in u.formsArr" :key="fi">
                    <view v-if="!isImgField(f)" class="form-row">
                      <text class="form-label">{{ f.label }}：</text>
                      <text class="form-value">{{ f.value }}</text>
                    </view>
                  </template>
                </view>
<view class="timeline-imgs-scroll" v-if="imgList(u.formsArr).length > 0">
  <image
    v-for="(img, ii) in imgList(u.formsArr)"
    :key="ii"
    :src="img"
    mode="aspectFill"
    class="timeline-img-h"
    @click="previewImg(imgList(u.formsArr), ii)"
  ></image>
</view>
<text class="timeline-location" v-if="getLocation(u.formsArr)">位置：{{ getLocation(u.formsArr) }}</text>
              </view>
            </view>
          </view>
          <view v-else class="empty-tip">暂无打卡记录~</view>
        </view>
      </view>

      <view class="bottom-bar">
        <view
          class="btn"
          :class="{ over: item.statusDesc !== '进行中', joinBtn: item.statusDesc === '进行中' && !item.isJoin }"
          @click="handleJoin"
        >
          {{ detailBtnText }}
        </view>
      </view>
    </block>
  </view>
</template>

<script>
import { enrollApi } from '../../api/index'

export default {
  data() {
    return {
      isLoad: false,
      item: {},
      cur: 'content',
      activityDay: '',
      activityList: [],
      calendarYear: 0,
      calendarMonth: 0,
      weekdays: ['日', '一', '二', '三', '四', '五', '六']
    }
  },

  onLoad(options) {
    if (options && options.id) {
      this.id = options.id
      this.loadDetail()
    }
  },

  onShow() {
    if (this.id) {
      this.loadDetail()
    }
  },

  onPullDownRefresh() {
    this.loadDetail().then(() => {
      uni.stopPullDownRefresh()
    })
  },

  computed: {
    todayStr() {
      const d = new Date()
      return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0')
    },
    checkinSet() {
      const s = new Set()
      if (this.item && this.item.dayList) {
        for (const d of this.item.dayList) {
          s.add(d.day)
        }
      }
      return s
    },
    calendarCells() {
      const year = this.calendarYear
      const month = this.calendarMonth
      const firstDay = new Date(year, month - 1, 1).getDay()
      const daysInMonth = new Date(year, month, 0).getDate()
      const cells = []
      for (let i = 0; i < firstDay; i++) {
        cells.push({ day: null, date: '' })
      }
      for (let d = 1; d <= daysInMonth; d++) {
        const dayStr = year + '-' + String(month).padStart(2, '0') + '-' + String(d).padStart(2, '0')
        cells.push({ day: dayStr, date: d })
      }
      return cells
    },
    detailBtnText() {
      if (this.item.statusDesc !== '进行中') return this.item.statusDesc
      if (!this.item.isJoin) return '立即参与'
      return '去打卡'
    },
    detailBtnDisabled() {
      return this.item.statusDesc !== '进行中'
    }
  },

  methods: {
    async loadDetail() {
      try {
        const userInfo = uni.getStorageSync('userInfo')
        const token = uni.getStorageSync('token')
        const uid = (userInfo && (userInfo.miniOpenID || userInfo.id)) || token || ''
        const res = await enrollApi.detail({ id: this.id, user_id: uid })
        if (!res.data) {
          this.isLoad = null
          return
        }
        this.item = res.data
        this.isLoad = true
        const now = new Date()
        this.calendarYear = now.getFullYear()
        this.calendarMonth = now.getMonth() + 1
        this.activityDay = this.todayStr
        this.selectDay(this.activityDay)
      } catch (e) {
        console.error('加载打卡详情失败', e)
        this.isLoad = null
      }
    },

    switchTab(tab) {
      this.cur = tab
    },

    prevMonth() {
      if (this.calendarMonth === 1) {
        this.calendarYear--
        this.calendarMonth = 12
      } else {
        this.calendarMonth--
      }
    },

    nextMonth() {
      if (this.calendarMonth === 12) {
        this.calendarYear++
        this.calendarMonth = 1
      } else {
        this.calendarMonth++
      }
    },

    async selectDay(day) {
      this.activityDay = day
      try {
        const res = await enrollApi.joinDay({ id: this.id, day })
        this.activityList = Array.isArray(res.data) ? res.data : []
      } catch (e) {
        console.error('加载打卡动态失败', e)
      }
    },

    isImgField(f) {
      if (f.type === 'image') return true
      if (f.locField) return true
      const label = (f.label || '').toLowerCase()
      const val = (f.value || '')
      if (label === '位置' || label.includes('纬度') || label.includes('经度')) return true
      return val.startsWith('http') && (label.includes('图') || label.includes('照片') || label.includes('img') || label.includes('pic') || label.includes('image'))
    },

    imgList(formsArr) {
      if (!formsArr) return []
      return formsArr.filter(f => {
        if (f.locField) return false
        if (f.type === 'image') return true
        return this.isImgField(f)
      }).map(f => f.value)
    },

    previewImg(images, index) {
      uni.previewImage({
        urls: images,
        current: images[index],
        indicator: 'number'
      })
    },

    getLocation(formsArr) {
      if (!formsArr) return ''
      const addr = formsArr.find(f => f.locField === '地址')
      if (addr) return addr.value
      const f = formsArr.find(f => f.label === '位置')
      return f ? f.value : ''
    },

    formatTimestamp(ts) {
      if (!ts) return ''
      const d = new Date(ts)
      return String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0')
    },

    handleJoin() {
      if (this.item.statusDesc !== '进行中') return
      if (this.item.isJoin) {
        uni.navigateTo({ url: '/pages/enroll/enroll_join_form?id=' + this.id })
        return
      }
      // Not joined yet — show enrollment form if has forms, else join directly
      let enrollForms = []
      try {
        enrollForms = typeof this.item.forms === 'string' ? JSON.parse(this.item.forms || '[]') : (this.item.forms || [])
      } catch (e) {}
      if (enrollForms.length > 0) {
        uni.navigateTo({ url: '/pages/enroll/enroll_join_form?id=' + this.id + '&mode=enroll' })
      } else {
        uni.navigateTo({ url: '/pages/enroll/enroll_join_form?id=' + this.id })
      }
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.load-status {
  text-align: center;
  padding-top: 200rpx;
  font-size: 28rpx;
  color: #999;
}

.top {
  background-color: #fb454c;
  padding: 60rpx 35rpx 110rpx;
  display: flex;
  align-items: flex-start;
  border-radius: 0 0 72% 39% / 10% 11% 16% 0;
}

.cover {
  width: 160rpx;
  height: 160rpx;
  border-radius: 30rpx;
  border: 10rpx solid #eb9a9a;
  margin-right: 30rpx;
  flex-shrink: 0;
}
.cover-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-color: transparent;
}
.cover-placeholder-text {
  font-size: 28rpx;
  color: #fff;
  font-weight: bold;
  text-align: center;
  word-break: break-all;
  line-height: 1.3;
  padding: 0 10rpx;
}

.top-right {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.title {
  font-size: 36rpx;
  font-weight: bold;
  color: #fff;
}

.desc {
  font-size: 24rpx;
  color: rgba(245, 196, 196, 0.9);
  padding-top: 10rpx;
}

.detail-card {
  margin: -70rpx 16rpx 150rpx;
  background-color: #fff;
  border-radius: 30rpx;
  min-height: 400rpx;
  overflow: hidden;
}

.tab-bar {
  display: flex;
  background-image: linear-gradient(0deg, #ffffff, #eee);
  height: 100rpx;
}

.tab-item {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  color: #333;
}

.tab-item.cur {
  font-weight: bold;
  background-color: #fff;
  border-radius: 20rpx;
}

.content-tab {
  padding: 20rpx;
}

.date-line {
  padding: 15rpx 20rpx;
  font-size: 26rpx;
  background-color: #f2f2f2;
  border-radius: 10rpx;
  color: #666;
  margin-bottom: 20rpx;
}

.desc-block {
  padding: 10rpx 0;
  font-size: 28rpx;
  color: #666;
  line-height: 1.8;
  margin-bottom: 10rpx;
}

.text-block {
  padding: 10rpx 0;
  font-size: 28rpx;
  color: #333;
  line-height: 1.8;
}

.img-block {
  margin: 20rpx 0;
}

.content-img {
  width: 100%;
  border-radius: 12rpx;
}

.rank-tab {
  padding: 40rpx 30rpx;
}

.rank-item {
  display: flex;
  align-items: center;
  margin-bottom: 35rpx;
}

.rank-no {
  width: 60rpx;
  font-weight: bold;
  font-size: 30rpx;
  color: #999;
  font-family: 'din';
}

.rank-no.top3 {
  color: #fb454c;
}

.rank-avatar {
  width: 90rpx;
  height: 90rpx;
  border-radius: 15rpx;
  margin-right: 20rpx;
  flex-shrink: 0;
}
.rank-avatar-text {
  width: 90rpx;
  height: 90rpx;
  border-radius: 15rpx;
  background-color: #fb454c;
  color: #fff;
  font-size: 36rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-right: 20rpx;
}

.rank-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.rank-name {
  font-size: 30rpx;
  color: #333;
  font-weight: bold;
  padding-bottom: 14rpx;
}

.rank-desc {
  font-size: 24rpx;
  color: #999;
}

.activity-tab {
  padding: 20rpx;
}

.calendar {
  padding: 20rpx;
}

.calendar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 10rpx 20rpx;
}

.cal-nav {
  font-size: 40rpx;
  color: #fb454c;
  padding: 10rpx 20rpx;
  font-weight: bold;
}

.cal-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
}

.calendar-weekdays {
  display: flex;
}

.weekday {
  flex: 1;
  text-align: center;
  font-size: 24rpx;
  color: #999;
  padding-bottom: 16rpx;
}

.calendar-grid {
  display: flex;
  flex-wrap: wrap;
}

.cal-cell {
  width: calc(100% / 7);
  height: 90rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  position: relative;
}

.cal-empty {
  visibility: hidden;
}

.cal-day {
  font-size: 28rpx;
  color: #333;
}

.cal-today .cal-day {
  background-color: #fb454c;
  color: #fff !important;
  width: 60rpx;
  height: 60rpx;
  border-radius: 50%;
  text-align: center;
  line-height: 60rpx;
}

.cal-selected .cal-day {
  border: 2rpx solid #fb454c;
  width: 60rpx;
  height: 60rpx;
  border-radius: 50%;
  text-align: center;
  line-height: 56rpx;
}

.cal-today.cal-selected .cal-day {
  border: none;
  line-height: 60rpx;
}

.cal-checkin .cal-day {
  color: #fb454c;
  font-weight: bold;
}

.cal-dot {
  font-size: 14rpx;
  color: #fb454c;
  position: absolute;
  bottom: -2rpx;
  line-height: 1;
}

.timeline {
  padding: 30rpx 20rpx;
}

.timeline-item {
  display: flex;
  padding-left: 20rpx;
  padding-bottom: 40rpx;
  position: relative;
  border-left: 4rpx solid #fb454c;
  margin-left: 20rpx;
}

.timeline-item:last-child {
  border-left-color: transparent;
}

.timeline-dot {
  position: absolute;
  left: -2rpx;
  top: 8rpx;
  font-size: 22rpx;
  color: #999;
  white-space: nowrap;
  line-height: 1;
  padding: 0 6rpx;
  background-color: #fff;
  transform: translateX(-50%);
  z-index: 1;
}

.timeline-content {
  flex: 1;
  min-width: 0;
  padding-left: 30rpx;
}

.timeline-user {
  display: flex;
  align-items: center;
  margin-bottom: 16rpx;
}

.timeline-avatar {
  width: 60rpx;
  height: 60rpx;
  border-radius: 50%;
  margin-right: 16rpx;
  flex-shrink: 0;
}
.timeline-avatar-text {
  width: 60rpx;
  height: 60rpx;
  border-radius: 50%;
  background-color: #fb454c;
  color: #fff;
  font-size: 24rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-right: 16rpx;
}

.timeline-name {
  font-size: 28rpx;
  font-weight: bold;
  color: #333;
}

.timeline-forms {
  padding: 10rpx 0;
}

.form-row {
  font-size: 26rpx;
  color: #666;
  line-height: 1.8;
}

.form-label {
  color: #999;
}

.timeline-imgs-scroll {
  width: 100%;
  margin-top: 15rpx;
  overflow-x: auto;
  overflow-y: hidden;
  white-space: nowrap;
  scrollbar-width: none;
}
.timeline-imgs-scroll::-webkit-scrollbar {
  display: none;
}

.timeline-img-h {
  width: 200rpx;
  height: 200rpx;
  border-radius: 12rpx;
  margin-right: 15rpx;
  display: inline-block;
  vertical-align: top;
}

.timeline-img-h:last-child {
  margin-right: 0;
}

.timeline-location {
  font-size: 22rpx;
  color: #999;
  margin-top: 10rpx;
  padding-left: 4rpx;
}

.empty-tip {
  width: 100%;
  text-align: center;
  font-size: 26rpx;
  color: #999;
  padding: 40rpx 0;
}

.bottom-bar {
  position: fixed;
  bottom: 50rpx;
  left: 0;
  right: 0;
  display: flex;
  justify-content: center;
  z-index: 9999;
}

.btn {
  background-image: linear-gradient(180deg, #df7b7b, #fb454c);
  padding: 20rpx 100rpx;
  border-radius: 40rpx;
  color: #fff;
  font-size: 30rpx;
}

.btn.over {
  background-image: linear-gradient(180deg, #f8f8f8, #f8f8f8);
  color: #999;
}
.joinBtn {
  background-image: linear-gradient(180deg, #ff6b6b, #fb454c);
  box-shadow: 0 8rpx 24rpx rgba(251, 69, 76, 0.4);
  transform: scale(1.05);
}
</style>
