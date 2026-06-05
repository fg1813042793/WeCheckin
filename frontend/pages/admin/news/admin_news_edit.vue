<template>
  <view class="container">
    <scroll-view scroll-y class="form-scroll">
      <view class="form-group">
        <text class="form-label">标题 <text class="required">*</text></text>
        <input v-model="form.title" placeholder="请输入通知标题" class="form-input" />
      </view>

      <view class="form-group">
        <text class="form-label">分类</text>
        <picker :range="categories" range-key="name" @change="onCateChange">
          <view class="picker-value" :class="{ placeholder: !form.cateId }">{{ form.cateName || '请选择分类' }}</view>
        </picker>
      </view>

      <view class="form-group">
        <text class="form-label">排序</text>
        <input v-model="form.order" type="digit" placeholder="输入排序值（越小越靠前）" class="form-input" />
      </view>

      <view class="form-group">
        <text class="form-label">描述</text>
        <textarea v-model="form.desc" placeholder="简短描述（可选）" class="form-textarea" />
      </view>

      <view class="form-group">
        <text class="form-label">内容 <text class="required">*</text></text>
        <textarea v-model="form.content" placeholder="请输入通知内容" class="form-textarea content-area" />
      </view>

      <view class="form-group">
        <text class="form-label">封面图片</text>
        <view class="cover-upload" @click="chooseCover">
          <image v-if="form.img" :src="form.img" mode="aspectFill" class="cover-preview"></image>
          <view v-else class="cover-placeholder">
            <text class="cover-plus">+</text>
            <text class="cover-hint">点击上传封面</text>
          </view>
        </view>
      </view>

      <view class="form-group" @click="toggleDeptPicker">
        <text class="form-label">发布部门</text>
        <view class="picker-value" :class="{ placeholder: !form.publishDeptIds }">
          {{ form.publishDeptNames || '全部部门可见（点击选择）' }}
        </view>
      </view>

      <view class="form-actions">
        <view class="btn-submit" @click="handleSubmit">保存</view>
        <view class="btn-cancel" @click="goBack">取消</view>
      </view>
    </scroll-view>

    <!-- 部门选择弹窗 -->
    <view class="modal-mask" v-if="showDeptPicker" @click="showDeptPicker = false">
      <view class="modal-content" @click.stop style="max-height:70vh">
        <text class="modal-title">选择发布部门</text>
        <view v-for="(d, i) in visibleDepts" :key="d.id" class="dept-item" :style="{ paddingLeft: (d.depth * 40 + 20) + 'rpx' }">
          <view class="dept-left" @click="toggleDeptExpand(d.id)">
            <text v-if="d.hasChildren" class="dept-arrow">{{ d.expanded ? '▼' : '▶' }}</text>
            <text v-else class="dept-arrow dept-arrow-placeholder">　</text>
            <text class="dept-name">{{ d.name }}</text>
          </view>
          <text class="dept-check" v-if="selectedDeptIds.includes(d.id)" @click="toggleDeptSelect(d.id)">✓</text>
          <text v-else class="dept-check dept-uncheck" @click="toggleDeptSelect(d.id)">○</text>
        </view>
        <view class="modal-actions">
          <view class="modal-btn btn-cancel" @click="showDeptPicker = false">取消</view>
          <view class="modal-btn btn-confirm" @click="confirmDeptPicker">确定</view>
        </view>
      </view>
    </view>

  </view>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { adminApi, dictApi } from '../../../api/admin'

const form = reactive({
  id: 0,
  title: '',
  cateId: 0,
  cateName: '',
  order: 0,
  desc: '',
  content: '',
  img: '',
  publishDeptIds: '',
  publishDeptNames: ''
})
const categories = ref([])
const loaded = ref(false)
const deptTree = ref([])
const expandedDeptIds = ref([])
const showDeptPicker = ref(false)
const selectedDeptIds = ref([])

const visibleDepts = computed(() => {
  const result = []
  const walk = (nodes, depth) => {
    for (const n of nodes) {
      const hasChildren = n.children && n.children.length > 0
      const expanded = expandedDeptIds.value.includes(n.id)
      result.push({ id: n.id, name: n.name, depth, hasChildren, expanded })
      if (hasChildren && expanded) {
        walk(n.children, depth + 1)
      }
    }
  }
  walk(deptTree.value, 0)
  return result
})

async function loadDetail(id) {
  try {
    const [res, deptRes] = await Promise.all([
      adminApi.newsDetail(id),
      adminApi.deptTree()
    ])
    deptTree.value = deptRes.data || []
    expandedDeptIds.value = deptTree.value.map(n => n.id)
    if (res.data) {
      form.id = res.data.id
      form.title = res.data.title || ''
      form.cateId = res.data.cateId || 0
      form.cateName = res.data.cateName || ''
      form.order = res.data.order || 0
      form.desc = res.data.desc || ''
      form.content = res.data.content || ''
      form.img = res.data.img || ''
      form.publishDeptIds = res.data.publishDeptIds || ''
      form.publishDeptNames = getDeptNames(res.data.publishDeptIds || '')
      loaded.value = true
    }
  } catch (e) {
    console.error('加载通知详情失败', e)
  }
}

async function loadCategories() {
  try {
    const res = await dictApi.items('content_type')
    const list = Array.isArray(res.data) ? res.data : (res.data.list || [])
    categories.value = list.map(d => ({ id: d.value, name: d.label }))
  } catch (e) {
    console.error('加载分类失败', e)
  }
}

async function loadDeptTree() {
  try {
    const res = await adminApi.deptTree()
    deptTree.value = res.data || []
    expandedDeptIds.value = deptTree.value.map(n => n.id)
  } catch (e) {
    console.error('加载部门失败', e)
  }
}

function onCateChange(e) {
  const index = e.detail.value
  const c = categories.value[index]
  if (c) {
    form.cateId = c.id
    form.cateName = c.name
  }
}

async function handleSubmit() {
  if (!form.title.trim()) {
    uni.showToast({ title: '请输入标题', icon: 'none' })
    return
  }
  if (!form.content.trim()) {
    uni.showToast({ title: '请输入内容', icon: 'none' })
    return
  }
  try {
    const data = { ...form }
    if (!data.cateId) {
      data.cateId = 0
      delete data.cateName
    }
    delete data.publishDeptNames
    await adminApi.newsEdit(data)
    uni.showToast({ title: '保存成功', icon: 'success' })
    setTimeout(() => {
      uni.navigateBack()
    }, 500)
  } catch (e) {
    console.error('编辑通知失败', e)
  }
}

function chooseCover() {
  uni.chooseImage({
    count: 1,
    success: (res) => {
      const tempPath = res.tempFilePaths[0]
      uni.uploadFile({
        url: 'http://localhost:8080/upload',
        filePath: tempPath,
        name: 'file',
        header: {
          Authorization: uni.getStorageSync('admin_token') || ''
        },
        success: (uploadRes) => {
          const data = JSON.parse(uploadRes.data)
          if (data && data.data && data.data.url) {
            form.img = data.data.url
          } else {
            uni.showToast({ title: '上传失败', icon: 'none' })
          }
        },
        fail: () => {
          uni.showToast({ title: '上传失败', icon: 'none' })
        }
      })
    }
  })
}

function goBack() {
  uni.navigateBack()
}

function findDeptInTree(nodes, id) {
  for (const n of nodes) {
    if (n.id === id) return n
    if (n.children && n.children.length) {
      const found = findDeptInTree(n.children, id)
      if (found) return found
    }
  }
  return null
}

function getDeptName(id) {
  const n = findDeptInTree(deptTree.value, Number(id))
  return n ? n.name : ''
}

function getDeptNames(ids) {
  if (!ids) return ''
  return ids.split(',').map(id => getDeptName(id)).filter(Boolean).join('、')
}

function toggleDeptExpand(id) {
  const idx = expandedDeptIds.value.indexOf(id)
  if (idx >= 0) {
    expandedDeptIds.value.splice(idx, 1)
  } else {
    expandedDeptIds.value.push(id)
  }
}

function toggleDeptSelect(id) {
  const idx = selectedDeptIds.value.indexOf(id)
  if (idx >= 0) {
    selectedDeptIds.value.splice(idx, 1)
  } else {
    selectedDeptIds.value.push(id)
  }
}

function toggleDeptPicker() {
  selectedDeptIds.value = form.publishDeptIds ? form.publishDeptIds.split(',').map(Number) : []
  showDeptPicker.value = true
}

function confirmDeptPicker() {
  form.publishDeptIds = selectedDeptIds.value.join(',')
  form.publishDeptNames = selectedDeptIds.value.map(id => getDeptName(id)).filter(Boolean).join('、')
  showDeptPicker.value = false
}

onLoad((options) => {
  const id = Number(options.id)
  if (id) {
    loadDetail(id)
  } else {
    loadDeptTree()
  }
  loadCategories()
})

onPullDownRefresh(() => {
  if (form.id) {
    loadDetail(form.id).then(() => {
      uni.stopPullDownRefresh()
    })
  } else {
    uni.stopPullDownRefresh()
  }
})
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.form-scroll {
  padding: 30rpx;
  max-width: 750rpx;
  margin: 0 auto;
  box-sizing: border-box;
}

.form-group {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}

.form-label {
  display: block;
  font-size: 28rpx;
  color: #333;
  font-weight: 500;
  margin-bottom: 16rpx;
}

.required {
  color: #ff4d4f;
}

.form-input {
  height: 72rpx;
  border: 1rpx solid #e8e8e8;
  border-radius: 12rpx;
  padding: 0 20rpx;
  font-size: 26rpx;
  color: #333;
}

.form-textarea {
  width: 100%;
  min-height: 120rpx;
  border: 1rpx solid #e8e8e8;
  border-radius: 12rpx;
  padding: 20rpx;
  font-size: 26rpx;
  color: #333;
  box-sizing: border-box;
}

.content-area {
  min-height: 300rpx;
}

.picker-value {
  font-size: 28rpx;
  color: #333;
  height: 60rpx;
  line-height: 60rpx;
}

.picker-value.placeholder {
  color: #999;
}

.cover-upload {
  width: 100%;
  height: 300rpx;
  border-radius: 12rpx;
  overflow: hidden;
}

.cover-preview {
  width: 100%;
  height: 100%;
  background-color: #f0f0f0;
}

.cover-placeholder {
  width: 100%;
  height: 100%;
  border: 2rpx dashed #ddd;
  border-radius: 12rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background-color: #fafafa;
}

.cover-plus {
  font-size: 60rpx;
  color: #ccc;
  line-height: 1;
}

.cover-hint {
  font-size: 26rpx;
  color: #999;
  margin-top: 12rpx;
}

.form-actions {
  display: flex;
  margin-top: 40rpx;
  padding-bottom: 60rpx;
}
.form-actions .btn-submit {
  margin-right: 20rpx;
}
.form-actions .btn-submit:last-child {
  margin-right: 0;
}

.btn-submit {
  flex: 1;
  background-color: #fb454c;
  color: #fff;
  text-align: center;
  padding: 24rpx 0;
  border-radius: 40rpx;
  font-size: 30rpx;
}

.btn-cancel {
  flex: 1;
  text-align: center;
  padding: 24rpx 0;
  border-radius: 40rpx;
  font-size: 30rpx;
  color: #666;
  background-color: #fff;
  border: 1rpx solid #ddd;
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
  width: 620rpx;
  max-height: 80vh;
  overflow-y: auto;
}
.modal-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 24rpx;
}
.modal-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 32rpx;
}
.modal-actions .modal-btn {
  margin-left: 20rpx;
}
.modal-actions .modal-btn:first-child {
  margin-left: 0;
}
.modal-btn {
  font-size: 28rpx;
  padding: 16rpx 48rpx;
  border-radius: 8rpx;
  text-align: center;
}
.modal-btn.btn-cancel {
  background-color: #eee;
  color: #666;
}
.modal-btn.btn-confirm {
  background-color: #fb454c;
  color: #fff;
}

.dept-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20rpx 0;
  border-bottom: 2rpx solid #f5f5f5;
}
.dept-left {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
}
.dept-arrow {
  font-size: 20rpx;
  color: #999;
  width: 32rpx;
  text-align: center;
  flex-shrink: 0;
}
.dept-arrow-placeholder {
  visibility: hidden;
}
.dept-name {
  font-size: 28rpx;
  color: #333;
  margin-left: 8rpx;
}
.dept-check {
  font-size: 32rpx;
  color: #fb454c;
  font-weight: bold;
  width: 48rpx;
  text-align: center;
  flex-shrink: 0;
}
.dept-uncheck {
  color: #ccc;
  font-weight: normal;
}

</style>
