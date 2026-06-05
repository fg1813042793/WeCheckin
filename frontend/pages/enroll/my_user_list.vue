<template>
  <view class="container" :style="{ paddingTop: containerPad }">
    <view class="tabs" :style="{ top: fixedTop }">
      <view class="tab-item" :class="{ active: cur === 'task' }" @click="switchTab('task')">今日任务</view>
      <view class="tab-item" :class="{ active: cur === 'stats' }" @click="switchTab('stats')">打卡统计</view>
    </view>

    <scroll-view scroll-y class="content">
      <view v-if="cur === 'task'">
        <view class="task-list" v-if="tasks.length > 0">
          <view class="task-card" v-for="(item, index) in tasks" :key="index" @click="goDetail(item.enrollId)">
            <view class="task-info">
              <text class="task-name">{{ item.title || '打卡任务' }}</text>
              <text class="task-meta">已打卡 {{ item.dayCnt || 0 }} 天<text v-if="item.todayJoinCnt > 0"> / 今日{{ item.todayJoinCnt }}次</text></text>
            </view>
            <text class="task-status" :class="item.checkedInToday ? 'done' : 'todo'">
              {{ item.checkedInToday ? '已打卡' : '去打卡' }}
            </text>
          </view>
        </view>
        <view class="empty" v-else>
          <text class="empty-text">暂无打卡任务</text>
        </view>
      </view>

      <view v-else class="stats-page">
        <view class="stats-card" v-if="stats">
          <view class="stat-item">
            <text class="stat-num">{{ stats.totalEnroll }}</text>
            <text class="stat-label">参与项目</text>
          </view>
          <view class="stat-item">
            <text class="stat-num">{{ stats.totalDays }}</text>
            <text class="stat-label">总打卡天数</text>
          </view>
          <view class="stat-item">
            <text class="stat-num">{{ stats.totalJoins }}</text>
            <text class="stat-label">总打卡次数</text>
          </view>
        </view>

        <view class="calendar-card">
          <view class="calendar-header">
            <view class="cal-nav" @click="prevMonth">‹</view>
            <text class="cal-title">{{ calendarYear }}年{{ calendarMonth }}月</text>
            <view class="cal-nav" @click="nextMonth">›</view>
          </view>
          <view class="calendar-weekdays">
            <text class="weekday" v-for="(w, wi) in weekdays" :key="wi">{{ w }}</text>
          </view>
          <view class="calendar-grid">
            <view v-for="(cell, ci) in calendarCells" :key="ci" class="cal-cell" :class="{ 'cal-empty': !cell.day, 'cal-checkin': cell.day && checkinDays.has(cell.day), 'cal-today': cell.day === todayStr, 'cal-selected': cell.day === selectedDay }" @click="cell.day && selectDay(cell.day)">
              <text v-if="cell.day" class="cal-day">{{ cell.date }}</text>
            </view>
          </view>
        </view>

        <view class="day-records" v-if="selectedDay && dayRecords.length > 0">
          <view class="day-records-header">{{ selectedDay }} 打卡记录</view>
          <view class="day-record-item" v-for="(rec, ri) in dayRecords" :key="ri">
            <view class="day-record-title">{{ rec.enrollTitle || '打卡' }}</view>
            <text class="day-record-time">{{ formatTime(rec.addTime) }}</text>
            <view class="day-record-images" v-if="rec.images && rec.images.length > 0">
              <image v-for="(img, ii) in rec.images" :key="ii" :src="img" mode="aspectFill" class="day-record-img" @click="previewImage(rec.images, ii)"></image>
            </view>
            <text class="day-record-location" v-if="rec.location">位置：{{ rec.location }}</text>
          </view>
        </view>

        <view class="stats-detail" v-if="tasks.length > 0">
          <view class="detail-item" v-for="(item, index) in tasks" :key="index" @click="goMyJoin(item.enrollId)">
            <text class="detail-name">{{ item.title || '打卡任务' }}</text>
            <view class="detail-numbers">
              <text class="detail-num">天数 <text class="num">{{ item.dayCnt || 0 }}</text></text>
              <text class="detail-num">次数 <text class="num">{{ item.joinCnt || 0 }}</text></text>
              <text class="detail-num">今日 <text class="num">{{ item.todayJoinCnt || 0 }}</text></text>
            </view>
          </view>
        </view>

        <view class="empty" v-else>
          <text class="empty-text">暂无统计数据</text>
        </view>
      </view>
    </scroll-view>
  </view>
</template>

<script>
import { enrollApi } from '../../api/index'

export default {
  data() {
    return {
      cur: 'task',
      tasks: [],
      stats: null,
      calendarYear: 0,
      calendarMonth: 0,
      checkinDays: new Set(),
      selectedDay: '',
      dayRecords: [],
      weekdays: ['日', '一', '二', '三', '四', '五', '六'],
      fixedTop: '0px',
      containerPad: '0px'
    }
  },

  computed: {
    todayStr() {
      const d = new Date()
      return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0')
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
    }
  },

  onLoad() {
    const now = new Date()
    this.calendarYear = now.getFullYear()
    this.calendarMonth = now.getMonth() + 1
    this.loadData()
    const sys = uni.getSystemInfoSync()
    const pxScale = 750 / sys.windowWidth
    if (sys.platform === 'android') {
      this.fixedTop = '12rpx'
      this.containerPad = '92rpx'
    } else {
      const navOffset = (sys.statusBarHeight || 0) + 44
      this.fixedTop = (navOffset + 6) + 'px'
      this.containerPad = (navOffset + 6 + Math.round(80 / pxScale)) + 'px'
    }
  },

  onShow() {
    this.loadData()
  },

  onPullDownRefresh() {
    this.loadData().then(() => {
      uni.stopPullDownRefresh()
    })
  },

  methods: {
    switchTab(tab) {
      this.cur = tab
      if (tab === 'stats' && this.checkinDays.size === 0) {
        this.loadCalendar()
      }
    },

    getUserId() {
      const userInfo = uni.getStorageSync('userInfo')
      const token = uni.getStorageSync('token')
      return (userInfo && (userInfo.miniOpenID || userInfo.id)) || token || ''
    },

    async loadData() {
      try {
        const uid = this.getUserId()
        const taskRes = await enrollApi.myUserList({ user_id: uid })
        this.tasks = Array.isArray(taskRes.data) ? taskRes.data : (taskRes.data.list || [])
        this.calcStats()
        this.loadCalendar()
      } catch (e) {
        console.error('加载打卡数据失败', e)
      }
    },

    async loadCalendar() {
      try {
        const uid = this.getUserId()
        const month = this.calendarYear + '-' + String(this.calendarMonth).padStart(2, '0')
        const res = await enrollApi.myCalendar({ user_id: uid, month })
        const days = new Set()
        if (res.data) {
          for (const enrollId in res.data) {
            for (const d of res.data[enrollId]) {
              days.add(d)
            }
          }
        }
        this.checkinDays = days
      } catch (e) {
        console.error('加载日历数据失败', e)
      }
    },

    prevMonth() {
      if (this.calendarMonth === 1) {
        this.calendarYear--
        this.calendarMonth = 12
      } else {
        this.calendarMonth--
      }
      this.loadCalendar()
    },

    nextMonth() {
      if (this.calendarMonth === 12) {
        this.calendarYear++
        this.calendarMonth = 1
      } else {
        this.calendarMonth++
      }
      this.loadCalendar()
    },

    async selectDay(day) {
      this.selectedDay = day
      this.dayRecords = []
      try {
        const uid = this.getUserId()
        const res = await enrollApi.myDayRecords({ user_id: uid, day })
        this.dayRecords = Array.isArray(res.data) ? res.data : []
      } catch (e) {
        console.error('加载当日打卡记录失败', e)
      }
    },

    formatTime(ts) {
      if (!ts) return ''
      const d = new Date(ts)
      return String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0')
    },

    previewImage(urls, current) {
      uni.previewImage({ urls, current: urls[current] })
    },

    calcStats() {
      let totalDays = 0
      let totalJoins = 0
      for (const t of this.tasks) {
        totalDays += t.dayCnt || 0
        totalJoins += t.joinCnt || 0
      }
      this.stats = {
        totalEnroll: this.tasks.length,
        totalDays,
        totalJoins
      }
    },

    goDetail(enrollId) {
      uni.navigateTo({ url: `/pages/enroll/enroll_detail?id=${enrollId}` })
    },
    goMyJoin(enrollId) {
      uni.navigateTo({ url: `/pages/enroll/enroll_my_join_list?id=${enrollId}` })
    }
  }
}
</script>

<style scoped>
.container {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #f5f5f5;
  overflow: hidden;
}

.tabs {
  position: fixed;
  left: 0;
  right: 0;
  z-index: 10;
  display: flex;
  background-color: #fff;
  padding: 20rpx 20rpx 0;
}

.tabs::before {
  content: '';
  position: absolute;
  top: -12rpx;
  left: 0;
  right: 0;
  height: 12rpx;
  background-color: #f5f5f5;
}

.tab-item {
  flex: 1;
  text-align: center;
  font-size: 28rpx;
  color: #666;
  padding: 16rpx 0;
  position: relative;
}

.tab-item.active {
  color: #fb454c;
  font-weight: bold;
}

.tab-item.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 60rpx;
  height: 4rpx;
  background-color: #fb454c;
  border-radius: 2rpx;
}

.content {
  flex: 1;
  overflow-y: auto;
}

.task-list, .stats-page {
  padding: 20rpx;
  max-width: 750rpx;
  margin: 0 auto;
}

.task-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background-color: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}

.task-info {
  flex: 1;
  min-width: 0;
  margin-right: 16rpx;
}

.task-name {
  font-size: 28rpx;
  color: #333;
  font-weight: 500;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-meta {
  font-size: 22rpx;
  color: #999;
  margin-top: 6rpx;
}

.task-status {
  font-size: 24rpx;
  padding: 8rpx 24rpx;
  border-radius: 24rpx;
  flex-shrink: 0;
}

.task-status.todo {
  background-color: #fb454c;
  color: #fff;
}

.task-status.done {
  background-color: #e8f5e9;
  color: #2e7d32;
}

.empty {
  text-align: center;
  padding-top: 200rpx;
}

.empty-text {
  font-size: 28rpx;
  color: #999;
}

.stats-card {
  display: flex;
  background-color: #fff;
  border-radius: 16rpx;
  padding: 30rpx;
  margin-bottom: 24rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
}

.stat-item {
  flex: 1;
  text-align: center;
}

.stat-num {
  display: block;
  font-size: 40rpx;
  font-weight: bold;
  color: #fb454c;
}

.stat-label {
  display: block;
  font-size: 24rpx;
  color: #999;
  margin-top: 8rpx;
}

.calendar-card {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 24rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
}

.calendar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 20rpx;
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
  height: 80rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.cal-empty {
  visibility: hidden;
}

.cal-day {
  font-size: 26rpx;
  color: #333;
}

.cal-checkin .cal-day {
  color: #fff;
  background-color: #fb454c;
  width: 56rpx;
  height: 56rpx;
  border-radius: 50%;
  text-align: center;
  line-height: 56rpx;
}

.cal-selected .cal-day {
  border: 2rpx solid #2499f2 !important;
  width: 56rpx;
  height: 56rpx;
  border-radius: 50%;
  text-align: center;
  line-height: 52rpx;
}

.cal-checkin.cal-selected .cal-day {
  border: none !important;
  line-height: 56rpx;
  background-color: #2499f2;
}

.cal-today .cal-day {
  border: 2rpx solid #fb454c;
  width: 56rpx;
  height: 56rpx;
  border-radius: 50%;
  text-align: center;
  line-height: 52rpx;
}

.cal-checkin.cal-today .cal-day {
  border: none;
  line-height: 56rpx;
}

.day-records {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 24rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
}

.day-records-header {
  font-size: 28rpx;
  font-weight: bold;
  color: #333;
  margin-bottom: 16rpx;
}

.day-record-item {
  padding: 16rpx 0;
  border-bottom: 2rpx solid #f5f5f5;
}

.day-record-item:last-child {
  border-bottom: none;
}

.day-record-title {
  font-size: 26rpx;
  color: #333;
  font-weight: 500;
}

.day-record-time {
  font-size: 22rpx;
  color: #999;
  margin-left: 12rpx;
}

.day-record-images {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
  margin-top: 8rpx;
}

.day-record-img {
  width: 140rpx;
  height: 140rpx;
  border-radius: 8rpx;
}

.day-record-location {
  display: block;
  font-size: 22rpx;
  color: #999;
  margin-top: 6rpx;
}

.day-record-empty {
  text-align: center;
  padding: 20rpx 0;
  font-size: 24rpx;
  color: #999;
}

.stats-detail {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20rpx 0;
  border-bottom: 2rpx solid #f5f5f5;
}

.detail-item:last-child {
  border-bottom: none;
}

.detail-name {
  font-size: 26rpx;
  color: #333;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 16rpx;
}

.detail-numbers {
  display: flex;
  flex-shrink: 0;
}
.detail-numbers .detail-num {
  margin-right: 20rpx;
}
.detail-numbers .detail-num:last-child {
  margin-right: 0;
}

.detail-num {
  font-size: 22rpx;
  color: #999;
}

.detail-num .num {
  color: #fb454c;
  font-weight: bold;
  font-size: 26rpx;
}
</style>
