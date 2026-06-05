<template>
  <view class="container">
    <view class="form-card" v-if="form">
      <view class="avatar-section">
        <image v-if="form.avatar" :src="form.avatar" mode="aspectFill" class="avatar" @click="changeAvatar"></image>
        <text v-else class="avatar-text" @click="changeAvatar">{{ (form.name || '?').charAt(0) }}</text>
        <text class="avatar-tip">点击更换头像</text>
      </view>

      <view class="form-group">
        <text class="label">用户名</text>
        <input class="input" v-model="form.name" placeholder="请输入用户名" />
      </view>

      <view class="form-group">
        <text class="label">手机号</text>
        <input class="input" v-model="form.mobile" placeholder="请输入手机号" type="text" />
      </view>

      <view class="form-group" @click="toggleDeptPicker">
        <text class="label">所属部门</text>
        <view class="picker">
          <text v-if="form.deptIds">{{ form.deptNames }}</text>
          <text v-else style="color:#999">选择所属部门（可多选）</text>
        </view>
      </view>

      <view class="form-group" v-if="form.avatar">
        <text class="label">头像链接</text>
        <input class="input" v-model="form.avatar" placeholder="头像URL" />
      </view>

      <view class="form-group" v-if="formFields.length > 0">
        <text class="label">扩展表单数据</text>
        <view class="forms-list">
          <view class="form-field-item" v-for="(f, fi) in formFields" :key="fi">
            <text class="field-label">{{ f.label }}<text v-if="f.required" class="required">*</text></text>
            <input v-if="f.type === '文本' || !f.type" class="input" v-model="f.value" :placeholder="'请输入' + f.label" />
            <input v-else-if="f.type === '数字'" class="input" type="number" v-model="f.value" :placeholder="'请输入' + f.label" />
            <textarea v-else-if="f.type === '多行文本'" class="field-textarea" v-model="f.value" :placeholder="'请输入' + f.label" />
            <picker v-else-if="f.type === '选择'" :range="f.optionsArr || []" @change="(e) => { f.value = (f.optionsArr || [])[e.detail.value] }">
              <view class="input" :class="{ 'input-placeholder': !f.value }">{{ f.value || '请选择' + f.label }}</view>
            </picker>
          </view>
        </view>
      </view>

      <view class="btn-row">
        <view class="save-btn" @click="save">保存</view>
      </view>
    </view>

    <view class="loading" v-else>
      <text>加载中...</text>
    </view>

    <!-- 部门选择弹窗 -->
    <view class="modal-mask" v-if="showDeptPicker" @click="showDeptPicker = false">
      <view class="modal-content" @click.stop style="max-height:70vh">
        <text class="modal-title">选择所属部门</text>
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

<script>
import CONFIG from '../../../config/index'
import { adminApi } from '../../../api/admin'

export default {
  data() {
    return {
      form: null,
      userId: '',
      isCreate: false,
      formFields: [],
      deptTree: [],
      expandedDeptIds: [],
      showDeptPicker: false,
      selectedDeptIds: []
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
    if (options.id) {
      this.userId = options.id
      this.loadUser(options.id)
    } else {
      this.isCreate = true
      uni.setNavigationBarTitle({ title: '增加用户' })
      this.initForm()
    }
  },

  onUnload() {
    if (this.isCreate) {
      uni.setNavigationBarTitle({ title: '编辑用户' })
    }
  },

  methods: {
    parseForms(raw) {
      if (!raw) return {}
      try {
        const arr = typeof raw === 'string' ? JSON.parse(raw) : raw
        if (Array.isArray(arr)) {
          const map = {}
          for (const item of arr) {
            map[item.label] = item.value
          }
          return map
        }
        return {}
      } catch (e) {
        return {}
      }
    },

    async loadFormConfig() {
      try {
        const res = await adminApi.userFormFields()
        const list = Array.isArray(res.data) ? res.data : []
        this.formFields = list.map(f => ({
          id: f.id,
          label: f.label,
          type: f.type || '文本',
          required: !!f.required,
          options: f.options || '',
          optionsArr: f.options ? f.options.split(',').map(s => s.trim()) : [],
          value: ''
        }))
      } catch (e) {
        this.formFields = []
      }
    },

    async loadUser(id) {
      try {
        const [userRes, configRes, deptRes] = await Promise.all([
          adminApi.userDetailById(id),
          adminApi.userFormFields(),
          adminApi.deptTree()
        ])
        this.deptTree = deptRes.data || []
        this.expandedDeptIds = this.deptTree.map(n => n.id)

        const list = Array.isArray(configRes.data) ? configRes.data : []
        const fieldDefs = list.map(f => ({
          id: f.id,
          label: f.label,
          type: f.type || '文本',
          required: !!f.required,
          options: f.options || '',
          optionsArr: f.options ? f.options.split(',').map(s => s.trim()) : [],
          value: ''
        }))

        const userForms = this.parseForms(userRes.data.forms || userRes.data.formList || [])
        for (const def of fieldDefs) {
          def.value = userForms[def.label] || ''
        }
        this.formFields = fieldDefs

        const deptIds = Array.isArray(userRes.data.deptIds) ? userRes.data.deptIds : []
        this.form = {
          name: userRes.data.name || '',
          mobile: userRes.data.mobile || '',
          avatar: userRes.data.avatar || '',
          deptIds: deptIds.join(','),
          deptNames: deptIds.map(id => this.getDeptName(id)).filter(Boolean).join('、')
        }
      } catch (e) {
        console.error('加载用户信息失败', e)
        uni.showToast({ title: '加载失败', icon: 'none' })
      }
    },

    initForm() {
      this.form = { name: '', mobile: '', avatar: '', deptIds: '', deptNames: '' }
      this.loadFormConfig()
      this.loadDeptTree()
    },

    changeAvatar() {
      uni.chooseImage({
        count: 1,
        success: (res) => {
          const tempFile = res.tempFilePaths[0]
          uni.showLoading({ title: '上传中...' })
          const token = uni.getStorageSync('admin_token')
          uni.uploadFile({
            url: CONFIG.BASE_URL + '/upload',
            filePath: tempFile,
            name: 'file',
            header: {
              'Authorization': token || ''
            },
            success: (uploadRes) => {
              uni.hideLoading()
              try {
                const data = JSON.parse(uploadRes.data)
                if (data.code === 0 && data.data && data.data.url) {
                  this.form.avatar = data.data.url
                } else {
                  uni.showToast({ title: data.msg || '上传失败', icon: 'none' })
                }
              } catch (e) {
                uni.showToast({ title: '上传失败', icon: 'none' })
              }
            },
            fail: () => {
              uni.hideLoading()
              uni.showToast({ title: '上传失败', icon: 'none' })
            }
          })
        }
      })
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

    toggleDeptPicker() {
      this.selectedDeptIds = this.form.deptIds ? this.form.deptIds.split(',').map(Number) : []
      this.showDeptPicker = true
    },

    confirmDeptPicker() {
      this.form.deptIds = this.selectedDeptIds.join(',')
      this.form.deptNames = this.selectedDeptIds.map(id => this.getDeptName(id)).filter(Boolean).join('、')
      this.showDeptPicker = false
    },

    async save() {
      if (!this.form.name) {
        uni.showToast({ title: '请输入用户名', icon: 'none' })
        return
      }

      const forms = this.formFields.filter(f => f.value).map(f => ({ label: f.label, value: f.value }))
      const formsStr = forms.length > 0 ? JSON.stringify(forms) : ''

      try {
        if (this.isCreate) {
          await adminApi.userAdd({
            name: this.form.name,
            mobile: this.form.mobile,
            pic: this.form.avatar,
            forms: formsStr,
            deptIds: this.form.deptIds
          })
        } else {
          await adminApi.userEdit({
            id: this.userId,
            name: this.form.name,
            mobile: this.form.mobile,
            pic: this.form.avatar,
            forms: formsStr,
            deptIds: this.form.deptIds
          })
        }
        uni.showToast({ title: '保存成功', icon: 'success' })
        setTimeout(() => {
          uni.navigateBack()
        }, 1500)
      } catch (e) {
        console.error('保存失败', e)
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

.form-card {
  background-color: #fff;
  border-radius: 20rpx;
  padding: 40rpx;
}

.avatar-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 40rpx;
}

.avatar {
  width: 140rpx;
  height: 140rpx;
  border-radius: 50%;
  background-color: #f0f0f0;
}
.avatar-text {
  width: 140rpx;
  height: 140rpx;
  border-radius: 50%;
  background-color: #fb454c;
  color: #fff;
  font-size: 56rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-tip {
  font-size: 24rpx;
  color: #999;
  margin-top: 12rpx;
}

.form-group {
  margin-bottom: 30rpx;
}

.label {
  display: block;
  font-size: 28rpx;
  color: #333;
  margin-bottom: 12rpx;
}

.input {
  width: 100%;
  height: 80rpx;
  border: 2rpx solid #eee;
  border-radius: 12rpx;
  padding: 0 20rpx;
  font-size: 28rpx;
  color: #333;
  box-sizing: border-box;
}

.btn-row {
  margin-top: 60rpx;
}

.save-btn {
  width: 100%;
  height: 88rpx;
  line-height: 88rpx;
  text-align: center;
  background-color: #2499f2;
  color: #fff;
  font-size: 32rpx;
  border-radius: 44rpx;
}

.forms-list {
  margin-bottom: 12rpx;
}
.form-field-item {
  margin-bottom: 20rpx;
}
.field-label {
  font-size: 26rpx;
  color: #666;
  display: block;
  margin-bottom: 10rpx;
}
.required {
  color: #fb454c;
}
.field-textarea {
  width: 100%;
  min-height: 120rpx;
  background-color: #f5f5f5;
  border-radius: 12rpx;
  padding: 16rpx 20rpx;
  font-size: 28rpx;
  color: #333;
  box-sizing: border-box;
}
.input-placeholder {
  color: #bbb;
}

.loading {
  text-align: center;
  padding-top: 300rpx;
  font-size: 28rpx;
  color: #999;
}

.picker {
  font-size: 28rpx;
  color: #333;
  height: 60rpx;
  line-height: 60rpx;
  border: 2rpx solid #eee;
  border-radius: 12rpx;
  padding: 0 20rpx;
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
