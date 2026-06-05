<template>
  <view class="container">
    <view class="form">
      <view class="form-item avatar-item" @click="chooseAvatar">
        <text class="form-label">头像</text>
        <view class="avatar-wrap">
          <image v-if="formPic" :src="formPic" mode="aspectFill" class="avatar"></image>
          <text v-else class="avatar-text">{{ (formName || '?').charAt(0) }}</text>
          <text class="avatar-arrow">></text>
        </view>
      </view>
      <view class="form-item">
        <text class="form-label">昵称</text>
        <input class="form-input" v-model="formName" placeholder="请输入昵称" placeholder-class="placeholder" />
      </view>
      <view class="form-item">
        <text class="form-label">手机号</text>
        <input class="form-input" v-model="formMobile" placeholder="请输入手机号" type="number" maxlength="11" placeholder-class="placeholder" />
      </view>
    </view>

    <view class="form" v-if="formFields.length > 0">
      <view class="form-group-header">扩展信息</view>
      <view class="form-item" v-for="(f, fi) in formFields" :key="fi">
        <text class="form-label">{{ f.label }}<text v-if="f.required" class="required">*</text></text>
        <input v-if="f.type === '文本' || !f.type" class="form-input" v-model="f.value" :placeholder="'请输入' + f.label" />
        <input v-else-if="f.type === '数字'" class="form-input" type="number" v-model="f.value" :placeholder="'请输入' + f.label" />
        <textarea v-else-if="f.type === '多行文本'" class="form-textarea" v-model="f.value" :placeholder="'请输入' + f.label" />
        <picker v-else-if="f.type === '选择'" :range="f.optionsArr || []" @change="(e) => { f.value = (f.optionsArr || [])[e.detail.value] }">
          <view class="form-input picker-text" :class="{ placeholder: !f.value }">{{ f.value || '请选择' + f.label }}</view>
        </picker>
      </view>
    </view>

    <view class="submit-btn" @click="submit">{{ isEdit ? '保存' : '注册' }}</view>
  </view>
</template>

<script>
import { passportApi, userFormFields } from '../../api/index'
import CONFIG from '../../config/index'

export default {
  data() {
    return {
      formName: '',
      formMobile: '',
      formPic: '',
      formFields: [],
      isEdit: false,
      userId: ''
    }
  },

  onLoad() {
    this.init()
  },

  methods: {
    getUserId() {
      if (this.userId) return this.userId
      const userInfo = uni.getStorageSync('userInfo')
      const token = uni.getStorageSync('token')
      return (userInfo && (userInfo.miniOpenID || userInfo.id)) || token || ''
    },

    async init() {
      this.userId = this.getUserId()
      await Promise.all([this.loadProfile(), this.loadFormConfig()])
    },

    async loadProfile() {
      try {
        const res = await passportApi.getMyDetail({ user_id: this.userId })
        const data = res.data || {}
        if (data && data.id) {
          this.isEdit = true
          this.formName = data.name || ''
          this.formMobile = data.mobile || ''
          this.formPic = data.avatar || ''

          const userForms = this.parseForms(data.forms || data.formList || [])
          // Pre-fill formFields values after loadFormConfig sets them up
          if (this.formFields.length > 0) {
            for (const f of this.formFields) {
              f.value = userForms[f.label] || ''
            }
          } else {
            // Store for later
            this._userForms = userForms
          }
        }
      } catch (e) {
        // 未注册，走注册流程
      }
    },

    async loadFormConfig() {
      try {
        const res = await userFormFields()
        const list = Array.isArray(res.data) ? res.data : []
        this.formFields = list.map(f => ({
          label: f.label,
          type: f.type || '文本',
          required: !!f.required,
          options: f.options || '',
          optionsArr: f.options ? f.options.split(',').map(s => s.trim()) : [],
          value: ''
        }))
        // Pre-fill if we had stored userForms earlier
        if (this._userForms) {
          for (const f of this.formFields) {
            f.value = this._userForms[f.label] || ''
          }
          this._userForms = null
        }
      } catch (e) {
        // ignore
      }
    },

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

    chooseAvatar() {
      uni.chooseImage({
        count: 1,
        sizeType: ['compressed'],
        sourceType: ['album', 'camera'],
        success: (res) => {
          const tempFilePaths = res.tempFilePaths
          if (tempFilePaths.length > 0) {
            uni.showLoading({ title: '上传中...' })
            uni.uploadFile({
              url: CONFIG.BASE_URL + '/upload',
              filePath: tempFilePaths[0],
              name: 'file',
              success: (uploadRes) => {
                uni.hideLoading()
                try {
                  const data = JSON.parse(uploadRes.data)
                  if (data.code === 0 && data.data && data.data.url) {
                    this.formPic = data.data.url
                    uni.showToast({ title: '上传成功', icon: 'success' })
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
        }
      })
    },

    async submit() {
      if (!this.formName) {
        uni.showToast({ title: '请输入昵称', icon: 'none' })
        return
      }

      const forms = this.formFields.filter(f => f.value).map(f => ({ label: f.label, value: f.value }))
      const formsStr = forms.length > 0 ? JSON.stringify(forms) : ''

      try {
        uni.showLoading({ title: this.isEdit ? '保存中...' : '注册中...', mask: true })
        const uid = this.getUserId()

        if (this.isEdit) {
          await passportApi.editBase({
            name: this.formName,
            mobile: this.formMobile,
            pic: this.formPic,
            user_id: uid,
            forms: formsStr
          })
          const userInfo = uni.getStorageSync('userInfo') || {}
          userInfo.name = this.formName
          userInfo.avatar = this.formPic
          uni.setStorageSync('userInfo', userInfo)
          uni.hideLoading()
          uni.showToast({ title: '保存成功', icon: 'success' })
          setTimeout(() => { uni.navigateBack() }, 1500)
        } else {
          const res = await passportApi.register({
            name: this.formName,
            mobile: this.formMobile,
            pic: this.formPic,
            user_id: uid,
            forms: formsStr
          })
          uni.hideLoading()
          if (res.code === 0) {
            if (res.data) {
              uni.setStorageSync('userInfo', res.data)
            }
            uni.showToast({ title: '注册成功', icon: 'success' })
            setTimeout(() => { uni.redirectTo({ url: '/pages/my/my_index' }) }, 1500)
          } else {
            uni.showToast({ title: res.msg || '操作失败', icon: 'none' })
          }
        }
      } catch (e) {
        uni.hideLoading()
        uni.showToast({ title: '操作失败', icon: 'none' })
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

.form {
  margin: 20rpx;
  background-color: #fff;
  border-radius: 16rpx;
  overflow: hidden;
}

.form-group-header {
  padding: 24rpx 30rpx 8rpx;
  font-size: 28rpx;
  font-weight: bold;
  color: #333;
  border-bottom: 1rpx solid #f5f5f5;
}

.form-item {
  display: flex;
  align-items: center;
  padding: 30rpx;
  border-bottom: 1rpx solid #f5f5f5;
}

.form-item:last-child {
  border-bottom: none;
}

.form-label {
  font-size: 30rpx;
  color: #333;
  width: 120rpx;
  flex-shrink: 0;
}

.form-input {
  flex: 1;
  font-size: 30rpx;
  color: #333;
}

.placeholder {
  color: #ccc;
}

.picker-text {
  min-height: 44rpx;
  line-height: 44rpx;
}

.required {
  color: #fb454c;
}

.form-textarea {
  flex: 1;
  min-height: 80rpx;
  font-size: 30rpx;
  color: #333;
}

.avatar-item {
  justify-content: space-between;
}

.avatar-wrap {
  display: flex;
  align-items: center;
}

.avatar {
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  margin-right: 16rpx;
}
.avatar-text {
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  background-color: #fb454c;
  color: #fff;
  font-size: 40rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16rpx;
  flex-shrink: 0;
}

.avatar-arrow {
  font-size: 28rpx;
  color: #ccc;
}

.submit-btn {
  margin: 60rpx 20rpx;
  background-color: #fb454c;
  color: #fff;
  text-align: center;
  padding: 28rpx;
  border-radius: 16rpx;
  font-size: 32rpx;
  font-weight: bold;
}
</style>