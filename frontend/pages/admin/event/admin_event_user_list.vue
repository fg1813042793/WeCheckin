<template>
  <view class="container">
    <view class="header">
      <view class="header-text">
        <text class="header-title">{{ title }}</text>
        <text class="header-count">共 {{ list.length }} 人参与</text>
      </view>
      <view class="export-btn" @click="exportUsers">导出用户信息</view>
    </view>
    <view class="list" v-if="list.length > 0">
      <view class="card" v-for="(item, i) in list" :key="i">
        <view class="card-top">
          <image v-if="item.avatar" :src="item.avatar" mode="aspectFill" class="avatar" />
          <view v-else class="avatar-placeholder">{{ (item.userName || '匿')[0] }}</view>
          <view class="card-name-row">
            <text class="user-name">{{ item.userName || item.nickname || '匿名' }}</text>
            <text class="user-dept" v-if="item.deptName">（{{ item.deptName }}）</text>
          </view>
        </view>
        <view class="card-body">
          <text class="user-time" v-if="item._createTime || item.addTime">报名时间：{{ formatTime(item._createTime || item.addTime) }}</text>
          <view class="user-forms" v-if="getForms(item.forms).length">
            <text class="forms-label">报名表单：</text>
            <view class="forms-list">
              <view class="forms-item" v-for="(f, k) in getForms(item.forms)" :key="k">
                <text class="forms-val">{{ f }}</text>
              </view>
            </view>
          </view>
        </view>
        <view class="card-actions">
          <view class="action-btn edit" @click="openEdit(item)">编辑</view>
          <view class="action-btn del" @click="delItem(item)">删除</view>
        </view>
      </view>
    </view>
    <view class="empty" v-else-if="!loading">
      <text class="empty-text">暂无参与者</text>
    </view>
    <view class="loading" v-if="loading"><text>加载中...</text></view>

    <view class="overlay" v-if="editVisible" @click="editVisible = false">
      <view class="overlay-content" @click.stop>
        <text class="overlay-title">编辑报名表单</text>
        <view v-for="(field, idx) in formSchema" :key="idx" class="overlay-item">
          <text class="ol-label">{{ field.label }}<text v-if="field.required" class="required">*</text>：</text>
          <input v-if="field.type === 'input' || field.type === 'text' || !field.type" v-model="editData[idx]" :placeholder="field.placeholder || '请输入'" class="ol-input" />
          <textarea v-else-if="field.type === 'textarea'" v-model="editData[idx]" :placeholder="field.placeholder || '请输入'" class="ol-textarea" />
          <picker v-else-if="field.type === 'select' || field.type === 'picker'" :range="field.options || []" @change="(e) => { editData[idx] = (field.options || [])[e.detail.value] }">
            <view class="ol-picker">{{ editData[idx] || (field.placeholder || '请选择') }}</view>
          </picker>
          <input v-else-if="field.type === 'number'" v-model="editData[idx]" type="number" :placeholder="field.placeholder || '请输入'" class="ol-input" />
          <input v-else v-model="editData[idx]" :placeholder="field.placeholder || '请输入'" class="ol-input" />
        </view>
        <view v-if="!formSchema.length" class="overlay-item">
          <text class="ol-label">表单数据：</text>
          <textarea v-model="editRawForms" class="ol-textarea" placeholder="JSON 数组" />
        </view>
        <view class="overlay-btns">
          <view class="ol-btn cancel" @click="editVisible = false">取消</view>
          <view class="ol-btn save" @click="saveEdit">保存</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { adminApi } from '../../../api/admin'
import { eventApi } from '../../../api/index'
export default {
  data() {
    return {
      eventId: '',
      title: '',
      list: [],
      loading: false,
      formSchema: [],
      editVisible: false,
      editId: '',
      editData: [],
      editRawForms: ''
    }
  },
  onLoad(opts) {
    if (opts.event_id) { this.eventId = opts.event_id }
    if (opts.title) { this.title = decodeURIComponent(opts.title) }
    this.loadData()
    this.loadFormSchema()
  },
  methods: {
    async loadData() {
      this.loading = true
      try {
        const res = await adminApi.eventParticipantList({ eventId: this.eventId })
        const data = Array.isArray(res.data) ? res.data : (res.data.list || [])
        this.list = data
      } catch (e) { console.error(e) }
      this.loading = false
    },
    async loadFormSchema() {
      try {
        const res = await eventApi.getDetail({ id: this.eventId })
        const event = res.data || {}
        let schema = []
        try { schema = typeof event.forms === 'string' ? JSON.parse(event.forms || '[]') : (event.forms || []) } catch (e) {}
        this.formSchema = Array.isArray(schema) ? schema : []
      } catch (e) {
        this.formSchema = []
      }
    },
    openEdit(item) {
      this.editId = item.id
      let values = []
      try { values = typeof item.forms === 'string' ? JSON.parse(item.forms || '[]') : (item.forms || []) } catch (e) {}
      this.editData = this.formSchema.map((_, i) => values[i] !== undefined && values[i] !== null ? String(values[i]) : '')
      this.editRawForms = !this.formSchema.length && item.forms ? item.forms : ''
      this.editVisible = true
    },
    async saveEdit() {
      try {
        let formsStr
        if (this.formSchema.length) {
          formsStr = JSON.stringify(this.editData)
        } else {
          try { JSON.parse(this.editRawForms || '[]'); formsStr = this.editRawForms || '[]' } catch (e) {
            uni.showToast({ title: '表单数据格式错误', icon: 'none' })
            return
          }
        }
        await adminApi.eventParticipantEdit({ id: this.editId, forms: formsStr })
        uni.showToast({ title: '已保存', icon: 'success' })
        this.editVisible = false
        this.loadData()
      } catch (e) { console.error(e) }
    },
    delItem(item) {
      uni.showModal({
        title: '提示',
        content: '确定删除该参与者？',
        success: async (r) => {
          if (!r.confirm) return
          try {
            await adminApi.eventParticipantDel({ id: item.id })
            uni.showToast({ title: '已删除', icon: 'success' })
            const idx = this.list.findIndex(x => x.id === item.id)
            if (idx > -1) this.list.splice(idx, 1)
          } catch (e) { console.error(e) }
        }
      })
    },
    parseForms(forms) {
      if (!forms) return ''
      try {
        const parsed = typeof forms === 'string' ? JSON.parse(forms) : forms
        if (Array.isArray(parsed)) {
          return parsed.map(v => v !== null && v !== undefined ? String(v) : '').join(' | ')
        }
        if (typeof parsed === 'object' && parsed) {
          return Object.entries(parsed).map(([k, v]) => k + ':' + (v ?? '')).join(' | ')
        }
        return String(parsed)
      } catch (e) {
        return String(forms)
      }
    },
    getForms(forms) {
      if (!forms) return []
      try {
        const parsed = typeof forms === 'string' ? JSON.parse(forms) : forms
        if (Array.isArray(parsed)) {
          return parsed.map((v, i) => {
            const lbl = this.formSchema[i] && this.formSchema[i].label
            const val = v !== null && v !== undefined ? String(v) : ''
            return lbl ? (lbl + '：' + val) : val
          }).filter(v => v !== '')
        }
        if (typeof parsed === 'object' && parsed) {
          return Object.entries(parsed).map(([k, v]) => k + '：' + (v ?? '')).filter(v => v)
        }
        return [String(parsed)]
      } catch (e) {
        return [String(forms)]
      }
    },
    buildCsv() {
      const header = ['用户ID(openid)', '名称', '电话', '部门', '顶级部门', '报名时间']
      this.formSchema.forEach(f => { if (f && f.label) header.push(f.label) })
      if (header.length === 6) header.push('报名表单')
      const rows = [header]
      this.list.forEach(p => {
        let values = []
        try { values = typeof p.forms === 'string' ? JSON.parse(p.forms || '[]') : (p.forms || []) } catch (e) {}
        const cells = [
          p.miniOpenId || '',
          p.userName || p.nickname || '',
          p.mobile || '',
          p.deptName || '',
          p.topDeptName || '',
          this.formatTime(p._createTime || p.addTime)
        ]
        for (let i = 0; i < this.formSchema.length; i++) {
          cells.push(values[i] !== undefined && values[i] !== null ? String(values[i]) : '')
        }
        if (!this.formSchema.length) cells.push(this.parseForms(p.forms))
        rows.push(cells)
      })
      return '\uFEFF' + rows.map(r => r.map(v => '"' + String(v ?? '').replace(/"/g, '""') + '"').join(',')).join('\n')
    },
    formatTime(ts) {
      if (!ts) return ''
      const n = typeof ts === 'string' ? Number(ts) : ts
      if (!n) return ''
      const d = new Date(n)
      if (isNaN(d.getTime())) return String(ts)
      return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0') + ' ' + String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0') + ':' + String(d.getSeconds()).padStart(2, '0')
    },
    exportUsers() {
      if (this.list.length === 0) {
        uni.showToast({ title: '暂无数据可导出', icon: 'none' })
        return
      }
      const csv = this.buildCsv()
      const filename = (this.title || '报名名单') + '-用户信息.csv'

      // #ifdef H5
      const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = filename
      a.click()
      URL.revokeObjectURL(url)
      uni.showToast({ title: '已导出', icon: 'success' })
      // #endif

      // #ifdef MP-WEIXIN
      const filePath = wx.env ? wx.env.USER_DATA_PATH + '/' + filename : ''
      const fs = wx.getFileSystemManager()
      fs.writeFile({
        filePath: filePath,
        data: csv,
        encoding: 'utf8',
        success: () => {
          wx.openDocument({ filePath: filePath, showMenu: true, success: () => {
            uni.showToast({ title: '已导出', icon: 'success' })
          }, fail: () => {
            uni.setClipboardData({ data: csv, success: () => {
              uni.showToast({ title: 'CSV已复制到剪贴板，请粘贴到 Excel', icon: 'none', duration: 2500 })
            } })
          } })
        },
        fail: () => {
          uni.setClipboardData({ data: csv, success: () => {
            uni.showToast({ title: 'CSV已复制到剪贴板', icon: 'none' })
          } })
        }
      })
      // #endif

      // #ifdef APP-PLUS
      const path = '_doc/' + filename
      const plusFS = plus.io.requestFileSystem(plus.io.PRIVATE_DOC, 0, (fs) => {
        fs.root.getFile(filename, { create: true }, (fileEntry) => {
          fileEntry.createWriter((writer) => {
            writer.onwrite = () => {
              plus.share.sendWithSystem({ type: 'file', files: [fileEntry.toLocalURL()] }, () => {}, () => {})
              uni.showToast({ title: '已保存到本地', icon: 'success' })
            }
            writer.onerror = (e) => {
              uni.setClipboardData({ data: csv, success: () => {
                uni.showToast({ title: 'CSV已复制到剪贴板', icon: 'none' })
              } })
            }
            writer.write(csv)
          }, (e) => {
            uni.setClipboardData({ data: csv, success: () => {
              uni.showToast({ title: 'CSV已复制到剪贴板', icon: 'none' })
            } })
          })
        }, (e) => {
          uni.setClipboardData({ data: csv, success: () => {
            uni.showToast({ title: 'CSV已复制到剪贴板', icon: 'none' })
          } })
        })
      }, (e) => {
        uni.setClipboardData({ data: csv, success: () => {
          uni.showToast({ title: 'CSV已复制到剪贴板', icon: 'none' })
        } })
      })
      // #endif
    }
  }
}
</script>

<style scoped>
.container { min-height: 100vh; background-color: #f5f5f5; }
.header { padding: 20rpx; background-color: #fff; margin-bottom: 20rpx; display: flex; align-items: center; justify-content: space-between; }
.header-text { flex: 1; min-width: 0; }
.header-title { font-size: 32rpx; font-weight: bold; color: #333; display: block; }
.header-count { font-size: 24rpx; color: #999; margin-top: 8rpx; display: block; }
.export-btn { flex-shrink: 0; background-color: #2b7ef5; color: #fff; padding: 12rpx 24rpx; border-radius: 8rpx; font-size: 26rpx; }
.list { padding: 0 20rpx; }
.card { background-color: #fff; border-radius: 12rpx; padding: 20rpx; margin-bottom: 12rpx; }
.card-top { display: flex; align-items: flex-start; }
.avatar { width: 64rpx; height: 64rpx; border-radius: 50%; flex-shrink: 0; }
.avatar-placeholder { width: 64rpx; height: 64rpx; border-radius: 50%; background-color: #fb454c; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 28rpx; flex-shrink: 0; }
.card-name-row { margin-left: 16rpx; padding-top: 8rpx; flex: 1; min-width: 0; display: flex; align-items: center; flex-wrap: wrap; }
.user-name { font-size: 30rpx; color: #333; font-weight: bold; }
.user-dept { font-size: 24rpx; color: #999; margin-left: 4rpx; }
.card-body { margin-top: 12rpx; padding-left: 80rpx; }
.user-time { font-size: 22rpx; color: #999; display: block; }
.user-forms { margin-top: 8rpx; display: flex; align-items: flex-start; }
.forms-label { font-size: 22rpx; color: #999; flex-shrink: 0; }
.forms-list { display: flex; flex-wrap: wrap; gap: 8rpx; flex: 1; }
.forms-item { background-color: #f0f5ff; color: #2b7ef5; padding: 4rpx 14rpx; border-radius: 6rpx; font-size: 22rpx; }
.card-actions { display: flex; justify-content: flex-end; gap: 16rpx; margin-top: 16rpx; padding-top: 16rpx; border-top: 1rpx solid #f0f0f0; }
.action-btn { padding: 10rpx 24rpx; border-radius: 8rpx; font-size: 24rpx; text-align: center; border: 1rpx solid transparent; }
.action-btn.edit { color: #2b7ef5; background-color: #f0f5ff; border-color: #d6e4ff; }
.action-btn.del { color: #fb454c; background-color: #fff1f0; border-color: #ffccc7; }
.empty, .loading { display: flex; align-items: center; justify-content: center; padding-top: 200rpx; font-size: 28rpx; color: #999; }
.overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background-color: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 100; }
.overlay-content { background-color: #fff; border-radius: 16rpx; padding: 32rpx; width: 80%; max-height: 70vh; overflow-y: auto; }
.overlay-title { font-size: 30rpx; font-weight: bold; color: #333; display: block; margin-bottom: 20rpx; text-align: center; }
.overlay-item { margin-bottom: 20rpx; }
.ol-label { font-size: 26rpx; color: #999; display: block; margin-bottom: 8rpx; }
.ol-val { font-size: 26rpx; color: #333; }
.ol-input { width: 100%; height: 72rpx; background-color: #f5f5f5; border-radius: 8rpx; padding: 0 16rpx; font-size: 26rpx; box-sizing: border-box; }
.ol-textarea { width: 100%; height: 140rpx; background-color: #f5f5f5; border-radius: 8rpx; padding: 12rpx 16rpx; font-size: 26rpx; box-sizing: border-box; }
.ol-picker { height: 72rpx; background-color: #f5f5f5; border-radius: 8rpx; padding: 0 16rpx; line-height: 72rpx; font-size: 26rpx; color: #999; }
.overlay-btns { display: flex; gap: 20rpx; margin-top: 24rpx; }
.ol-btn { flex: 1; height: 80rpx; line-height: 80rpx; text-align: center; border-radius: 8rpx; font-size: 28rpx; }
.ol-btn.cancel { background-color: #f5f5f5; color: #666; }
.ol-btn.save { background-color: #2b7ef5; color: #fff; }
.required { color: #fb454c; }
</style>
