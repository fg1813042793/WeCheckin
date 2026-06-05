<template>
  <view class="container">
    <view class="form">
      <view class="form-item">
        <text class="label">打卡标题</text>
        <input class="input" v-model="form.title" placeholder="请输入打卡标题" />
      </view>

      <view class="form-item">
        <text class="label">封面图片</text>
        <view class="upload-area">
          <image v-if="form.cover" :src="form.cover" mode="aspectFill" class="preview-img" @click="changeCover" />
          <view v-else class="upload-placeholder" @click="changeCover">
            <text>+</text>
            <text class="upload-hint">点击上传封面</text>
          </view>
        </view>
      </view>

      <view class="form-item">
        <text class="label">描述</text>
        <textarea class="textarea" v-model="form.desc" placeholder="请输入打卡描述" />
      </view>

      <view class="form-item" v-if="categories.length > 0">
        <text class="label">分类</text>
        <picker :range="categories" range-key="name" @change="onCateChange">
          <view class="picker">{{ form.cateName || '请选择分类' }}</view>
        </picker>
      </view>

      <view class="form-item">
        <text class="label">排序</text>
        <input class="input" v-model="form.sort" type="number" placeholder="数字越小越靠前" />
      </view>

      <view class="form-item" @click="toggleDeptPicker">
        <text class="label">发布部门</text>
        <view class="picker">
          <text v-if="form.publishDeptIds">{{ form.publishDeptNames }}</text>
          <text v-else style="color:#999">全部部门可见（点击选择）</text>
        </view>
      </view>

      <view class="form-item">
        <text class="label">开始时间</text>
        <picker mode="date" :value="form.startTime" @change="onStartTimeChange">
          <view class="picker">{{ form.startTime || '请选择开始时间' }}</view>
        </picker>
      </view>

      <view class="form-item">
        <text class="label">结束时间</text>
        <picker mode="date" :value="form.endTime" @change="onEndTimeChange">
          <view class="picker">{{ form.endTime || '请选择结束时间' }}</view>
        </picker>
      </view>

      <view class="form-item">
        <text class="label">允许重复打卡</text>
        <switch :checked="form.allowRepeat" @change="onRepeatChange" color="#fb454c" />
      </view>

      <view class="form-item">
        <text class="label">每日打卡次数</text>
        <input class="input" v-model="form.dailyLimit" type="number" placeholder="每日最多打卡次数" />
      </view>

      <view class="section-divider"></view>
      <view class="form-item">
        <text class="label section-label">参与报名表单字段</text>
        <text class="section-desc">配置用户报名时填写的字段</text>
      </view>

      <view class="field-list">
        <view class="field-card" v-for="(field, fi) in form.enrollFields" :key="fi">
          <view class="field-header">
            <text class="field-label">{{ field.label || '未命名字段' }}</text>
            <text class="field-type-badge">{{ typeLabel(field.type) }}</text>
            <text v-if="field.required" class="field-required">必填</text>
          </view>
          <view class="field-actions">
            <text class="field-action" @click="editEnrollField(fi)">编辑</text>
            <text class="field-action action-del" @click="removeEnrollField(fi)">删除</text>
          </view>
        </view>
      </view>

      <view class="add-field-btn" @click="addEnrollField">
        <text class="add-field-icon">+</text>
        <text>添加字段</text>
      </view>

      <view class="section-divider"></view>
      <view class="form-item">
        <text class="label section-label">打卡表单字段</text>
        <text class="section-desc">配置用户打卡时需要填写的字段</text>
      </view>

      <view class="field-list">
        <view class="field-card" v-for="(field, fi) in form.fields" :key="fi">
          <view class="field-header">
            <text class="field-label">{{ field.label || '未命名字段' }}</text>
            <text class="field-type-badge">{{ typeLabel(field.type) }}</text>
            <text v-if="field.required" class="field-required">必填</text>
          </view>
          <view class="field-actions">
            <text class="field-action" @click="editField(fi)">编辑</text>
            <text class="field-action action-del" @click="removeField(fi)">删除</text>
          </view>
        </view>
      </view>

      <view class="add-field-btn" @click="addField">
        <text class="add-field-icon">+</text>
        <text>添加字段</text>
      </view>

      <button class="submit-btn" :loading="loading" @click="handleSubmit">保存</button>
    </view>

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

    <view class="modal-mask" v-if="showFieldModal" @click="closeFieldModal">
      <view class="modal-content" @click.stop>
        <text class="modal-title">{{ editingIndex >= 0 ? '编辑字段' : '添加字段' }}</text>

        <view class="modal-item">
          <text class="modal-label">字段名称</text>
          <input class="modal-input" v-model="fieldForm.label" placeholder="例如：运动类型" :disabled="['image','location'].includes(fieldForm.type)" />
        </view>

        <view class="modal-item">
          <text class="modal-label">字段类型</text>
          <picker :range="fieldTypes" range-key="label" @change="onFieldTypeChange">
            <view class="modal-picker">{{ typeLabel(fieldForm.type) || '请选择' }}</view>
          </picker>
        </view>

        <view class="modal-item" v-if="fieldForm.type === 'select'">
          <text class="modal-label">选项（每行一个）</text>
          <textarea class="modal-textarea" v-model="fieldForm.optionsText" placeholder="每行一个选项" />
        </view>

        <view class="modal-item">
          <text class="modal-label">必填</text>
          <switch :checked="fieldForm.required" @change="onFieldRequiredChange" color="#fb454c" />
        </view>

        <view class="modal-actions">
          <view class="modal-btn btn-cancel" @click="closeFieldModal">取消</view>
          <view class="modal-btn btn-confirm" @click="confirmField">确定</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { adminApi } from '../../../api/admin'

export default {
  data() {
    return {
      id: '',
      form: {
        title: '',
        cover: '',
        desc: '',
        cateId: '',
        cateName: '',
        sort: '0',
        startTime: '',
        endTime: '',
        allowRepeat: false,
        dailyLimit: '1',
        enrollFields: [],
        fields: [],
        publishDeptIds: '',
        publishDeptNames: ''
      },
      categories: [
        { id: 1, name: '学习' },
        { id: 2, name: '兴趣' },
        { id: 3, name: '生活' },
        { id: 4, name: '运动' },
        { id: 5, name: '工作' }
      ],
      deptTree: [],
      expandedDeptIds: [],
      showDeptPicker: false,
      selectedDeptIds: [],
      loading: false,
      showFieldModal: false,
      fieldTarget: 'join',
      editingIndex: -1,
      fieldForm: {
        label: '',
        type: 'text',
        required: false,
        optionsText: ''
      },
      fieldTypes: [
        { label: '文本', value: 'text' },
        { label: '数字', value: 'number' },
        { label: '多行文本', value: 'textarea' },
        { label: '选择', value: 'select' },
        { label: '图片', value: 'image' },
        { label: '定位', value: 'location' }
      ]
    }
  },

  computed: {
    visibleDepts() {
      const result = []
      const walk = (nodes, depth) => {
        for (const n of nodes) {
          const hasChildren = n.children && n.children.length > 0
          const expanded = this.expandedDeptIds.includes(n.id)
          result.push({ id: n.id, name: n.name, depth, hasChildren, expanded })
          if (hasChildren && expanded) {
            walk(n.children, depth + 1)
          }
        }
      }
      walk(this.deptTree, 0)
      return result
    }
  },

  onLoad(options) {
    this.id = options.id
    this.loadDetail()
  },

  onPullDownRefresh() {
    this.loadDetail().then(() => {
      uni.stopPullDownRefresh()
    })
  },

  methods: {
    async loadDetail() {
      await this.loadDeptTree()
      try {
        const res = await adminApi.enrollDetail(this.id)
        if (res.data) {
          const data = res.data
          this.form.title = data.title || ''
          this.form.cateId = String(data.cateId || '')
          this.form.cateName = data.cateName || ''
          this.form.sort = String(data.order || data.sort || '0')
          this.form.desc = data.desc || ''
          this.form.cover = data.img || ''

          // Parse start/end time
          if (data.timeStart) {
            const d = new Date(data.timeStart)
            this.form.startTime = d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0')
          }
          if (data.timeEnd) {
            const d = new Date(data.timeEnd)
            this.form.endTime = d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0')
          }

          // Read allowRepeat/dailyLimit from columns
          this.form.allowRepeat = !!data.allowRepeat
          this.form.dailyLimit = String(data.dailyLimit || 1)

          // Parse enrollment form fields from forms (enroll_forms)
          if (data.forms) {
            try {
              const f = typeof data.forms === 'string' ? JSON.parse(data.forms) : data.forms
              this.form.enrollFields = Array.isArray(f) ? f : []
            } catch (e) {
              this.form.enrollFields = []
            }
          }

          // Parse join form fields from joinForms (enroll_join_forms)
          if (data.joinForms) {
            try {
              const f = typeof data.joinForms === 'string' ? JSON.parse(data.joinForms) : data.joinForms
              this.form.fields = Array.isArray(f) ? f : []
            } catch (e) {
              this.form.fields = []
            }
          }

          // Publish department
          this.form.publishDeptIds = data.publishDeptIds || ''
          this.form.publishDeptNames = this.getDeptNames(data.publishDeptIds || '')
        }
      } catch (e) {
        console.error('加载详情失败', e)
      }
    },

    typeLabel(type) {
      const map = { text: '文本', number: '数字', textarea: '多行文本', select: '选择', image: '拍照上传', location: '位置签到' }
      return map[type] || type || '文本'
    },

    changeCover() {
      uni.chooseImage({
        count: 1,
        sizeType: ['compressed'],
        success: (res) => {
          const tempFile = res.tempFilePaths[0]
          uni.showLoading({ title: '上传中...' })
          uni.uploadFile({
            url: 'http://192.168.50.6:8080/upload',
            filePath: tempFile,
            name: 'file',
            success: (uploadRes) => {
              const data = JSON.parse(uploadRes.data)
              if (data.code === 0) {
                this.form.cover = data.data.url
              } else {
                uni.showToast({ title: data.msg || '上传失败', icon: 'none' })
              }
            },
            fail: () => {
              uni.showToast({ title: '上传失败，请重试', icon: 'none' })
            },
            complete: () => {
              uni.hideLoading()
            }
          })
        }
      })
    },

    onCateChange(e) {
      const index = e.detail.value
      const c = this.categories[index]
      if (c) {
        this.form.cateId = String(c.id)
        this.form.cateName = c.name
      }
    },

    onStartTimeChange(e) {
      this.form.startTime = e.detail.value
    },

    onEndTimeChange(e) {
      this.form.endTime = e.detail.value
    },

    onRepeatChange(e) {
      this.form.allowRepeat = e.detail.value
    },

    addField() {
      this.fieldTarget = 'join'
      this.editingIndex = -1
      this.fieldForm = { label: '', type: 'text', required: false, optionsText: '' }
      this.showFieldModal = true
    },

    editField(index) {
      this.fieldTarget = 'join'
      const f = this.form.fields[index]
      this.editingIndex = index
      this.fieldForm = {
        label: f.label || '',
        type: f.type || 'text',
        required: !!f.required,
        optionsText: Array.isArray(f.options) ? f.options.join('\n') : ''
      }
      this.showFieldModal = true
    },

    removeField(index) {
      this.form.fields.splice(index, 1)
    },

    addEnrollField() {
      this.fieldTarget = 'enroll'
      this.editingIndex = -1
      this.fieldForm = { label: '', type: 'text', required: false, optionsText: '' }
      this.showFieldModal = true
    },

    editEnrollField(index) {
      this.fieldTarget = 'enroll'
      const f = this.form.enrollFields[index]
      this.editingIndex = index
      this.fieldForm = {
        label: f.label || '',
        type: f.type || 'text',
        required: !!f.required,
        optionsText: Array.isArray(f.options) ? f.options.join('\n') : ''
      }
      this.showFieldModal = true
    },

    removeEnrollField(index) {
      this.form.enrollFields.splice(index, 1)
    },

    closeFieldModal() {
      this.showFieldModal = false
    },

    onFieldTypeChange(e) {
      const type = this.fieldTypes[e.detail.value].value
      this.fieldForm.type = type
      if (type === 'image') this.fieldForm.label = '拍照上传'
      else if (type === 'location') this.fieldForm.label = '位置签到'
    },

    onFieldRequiredChange(e) {
      this.fieldForm.required = e.detail.value
    },

    confirmField() {
      if (!this.fieldForm.label) {
        uni.showToast({ title: '请输入字段名称', icon: 'none' })
        return
      }
      const field = {
        label: this.fieldForm.label,
        type: this.fieldForm.type,
        required: this.fieldForm.required
      }
      if (this.fieldForm.type === 'select') {
        field.options = this.fieldForm.optionsText
          .split('\n')
          .map(s => s.trim())
          .filter(s => s)
      }
      const target = this.fieldTarget === 'enroll' ? this.form.enrollFields : this.form.fields
      if (this.editingIndex >= 0) {
        target.splice(this.editingIndex, 1, field)
      } else {
        target.push(field)
      }
      this.closeFieldModal()
    },

    async loadDeptTree() {
      try {
        const res = await adminApi.deptTree()
        this.deptTree = res.data || []
        this.expandedDeptIds = this.deptTree.map(n => n.id)
      } catch (e) {
        console.error('加载部门失败', e)
      }
    },

    findDeptInTree(nodes, id) {
      for (const n of nodes) {
        if (n.id === id) return n
        if (n.children && n.children.length) {
          const found = this.findDeptInTree(n.children, id)
          if (found) return found
        }
      }
      return null
    },

    getDeptName(id) {
      const n = this.findDeptInTree(this.deptTree, Number(id))
      return n ? n.name : ''
    },

    getDeptNames(ids) {
      if (!ids) return ''
      return ids.split(',').map(id => this.getDeptName(id)).filter(Boolean).join('、')
    },

    toggleDeptExpand(id) {
      const idx = this.expandedDeptIds.indexOf(id)
      if (idx >= 0) {
        this.expandedDeptIds.splice(idx, 1)
      } else {
        this.expandedDeptIds.push(id)
      }
    },

    toggleDeptSelect(id) {
      const idx = this.selectedDeptIds.indexOf(id)
      if (idx >= 0) {
        this.selectedDeptIds.splice(idx, 1)
      } else {
        this.selectedDeptIds.push(id)
      }
    },

    confirmDeptPicker() {
      this.form.publishDeptIds = this.selectedDeptIds.join(',')
      this.form.publishDeptNames = this.selectedDeptIds.map(id => this.getDeptName(id)).filter(Boolean).join('、')
      this.showDeptPicker = false
    },

    toggleDeptPicker() {
      this.selectedDeptIds = this.form.publishDeptIds ? this.form.publishDeptIds.split(',').map(Number) : []
      this.showDeptPicker = true
    },

    async handleSubmit() {
      if (!this.form.title) {
        uni.showToast({ title: '请输入打卡标题', icon: 'none' })
        return
      }

      this.loading = true
      try {
        await adminApi.enrollEdit({
          id: this.id,
          title: this.form.title,
          cateId: this.form.cateId,
          cateName: this.form.cateName,
          sort: this.form.sort,
          startTime: this.form.startTime,
          endTime: this.form.endTime,
          cover: this.form.cover || '',
          desc: this.form.desc || '',
          allowRepeat: this.form.allowRepeat ? '1' : '0',
          dailyLimit: this.form.dailyLimit || '1',
          enrollForms: JSON.stringify(this.form.enrollFields),
          joinForms: JSON.stringify(this.form.fields),
          publishDeptIds: this.form.publishDeptIds
        })
        uni.showToast({ title: '保存成功', icon: 'success' })
        setTimeout(() => {
          uni.navigateBack()
        }, 1500)
      } catch (e) {
        console.error('保存失败', e)
      } finally {
        this.loading = false
      }
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding: 30rpx;
}

.form {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 30rpx;
}

.form-item {
  padding: 20rpx 0;
  border-bottom: 2rpx solid #f5f5f5;
}

.label {
  font-size: 28rpx;
  color: #333;
  display: block;
  margin-bottom: 16rpx;
}

.input {
  font-size: 28rpx;
  color: #333;
  height: 60rpx;
}

.textarea {
  font-size: 28rpx;
  color: #333;
  width: 100%;
  min-height: 120rpx;
  box-sizing: border-box;
}

.picker {
  font-size: 28rpx;
  color: #333;
  height: 60rpx;
  line-height: 60rpx;
}

.upload-area {
  display: flex;
  align-items: center;
}

.preview-img {
  width: 200rpx;
  height: 200rpx;
  border-radius: 12rpx;
  background: #f0f0f0;
}

.upload-placeholder {
  width: 200rpx;
  height: 200rpx;
  border: 2rpx dashed #ddd;
  border-radius: 12rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #ccc;
  font-size: 60rpx;
}

.upload-hint {
  font-size: 22rpx;
  color: #999;
  margin-top: 8rpx;
}

.submit-btn {
  margin-top: 60rpx;
  background-color: #fb454c;
  color: #fff;
  font-size: 32rpx;
  border-radius: 48rpx;
  height: 96rpx;
  line-height: 96rpx;
}

.submit-btn::after {
  border: none;
}

.section-divider {
  height: 2rpx;
  background-color: #eee;
  margin: 20rpx 0;
}

.section-label {
  font-size: 32rpx;
  font-weight: bold;
}

.section-desc {
  font-size: 24rpx;
  color: #999;
  margin-top: 8rpx;
}

.field-list {
  margin-top: 16rpx;
}

.field-card {
  background-color: #f9f9f9;
  border-radius: 12rpx;
  padding: 20rpx;
  margin-bottom: 12rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.field-header {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
}
.field-header > * {
  margin-right: 12rpx;
}
.field-header > *:last-child {
  margin-right: 0;
}

.field-label {
  font-size: 28rpx;
  color: #333;
  font-weight: 500;
}

.field-type-badge {
  font-size: 20rpx;
  background-color: #e3f2fd;
  color: #1565c0;
  padding: 2rpx 12rpx;
  border-radius: 6rpx;
}

.field-required {
  font-size: 20rpx;
  background-color: #fbe9e7;
  color: #c62828;
  padding: 2rpx 12rpx;
  border-radius: 6rpx;
}

.field-actions {
  display: flex;
  flex-shrink: 0;
}
.field-actions .field-action {
  margin-right: 12rpx;
}
.field-actions .field-action:last-child {
  margin-right: 0;
}

.field-action {
  font-size: 24rpx;
  color: #2499f2;
  padding: 4rpx 16rpx;
}

.action-del {
  color: #fb454c;
}

.add-field-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24rpx;
  border: 2rpx dashed #ddd;
  border-radius: 12rpx;
  color: #999;
  font-size: 26rpx;
  margin-top: 12rpx;
}
.add-field-btn .add-field-icon {
  margin-right: 8rpx;
}

.add-field-icon {
  font-size: 36rpx;
  color: #ccc;
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

.modal-item {
  margin-bottom: 24rpx;
}

.modal-label {
  font-size: 26rpx;
  color: #666;
  display: block;
  margin-bottom: 10rpx;
}

.modal-input {
  height: 72rpx;
  border: 2rpx solid #eee;
  border-radius: 12rpx;
  padding: 0 20rpx;
  font-size: 28rpx;
  color: #333;
}

.modal-textarea {
  width: 100%;
  height: 160rpx;
  border: 2rpx solid #eee;
  border-radius: 12rpx;
  padding: 16rpx;
  font-size: 26rpx;
  color: #333;
  box-sizing: border-box;
}

.modal-picker {
  height: 72rpx;
  line-height: 72rpx;
  border: 2rpx solid #eee;
  border-radius: 12rpx;
  padding: 0 20rpx;
  font-size: 28rpx;
  color: #333;
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

.btn-cancel {
  background-color: #eee;
  color: #666;
}

.btn-confirm {
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
