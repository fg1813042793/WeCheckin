import { computed, reactive } from 'vue'
import {
  changePassword as changeProfilePassword,
  updateProfile as updateProfileInfo,
  uploadAvatar as uploadProfileAvatarFile
} from '../../../api/profile'
import { firstText } from '../../performance/common/helpers'

export function usePerformanceProfile({ state, sameUserId, sanitizeUsers, toast }) {
  const profileDialog = reactive({
    visible: false,
    loading: false,
    uploading: false,
    account: '',
    avatar: '',
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  })

  const profileAvatarPreview = computed(() => firstText(profileDialog.avatar))
  const profileDisplayName = computed(() => firstText(state.user?.name, state.user?.account, state.user?.id, '当前用户'))
  const profileInitial = computed(() => firstText(profileDisplayName.value).slice(0, 1).toUpperCase() || 'U')

  function currentProfileAccount() {
    return firstText(state.user?.account, state.user?.id)
  }

  function currentProfileAvatar() {
    return firstText(state.user?.avatar, state.user?.avatarUrl, state.user?.pic, state.user?.userPic)
  }

  function resetProfileDialog() {
    Object.assign(profileDialog, {
      visible: false,
      loading: false,
      uploading: false,
      account: '',
      avatar: '',
      currentPassword: '',
      newPassword: '',
      confirmPassword: ''
    })
  }

  function openProfileDialog() {
    if (!state.user) return
    Object.assign(profileDialog, {
      visible: true,
      loading: false,
      uploading: false,
      account: currentProfileAccount(),
      avatar: currentProfileAvatar(),
      currentPassword: '',
      newPassword: '',
      confirmPassword: ''
    })
  }

  function closeProfileDialog() {
    if (profileDialog.loading || profileDialog.uploading) return
    resetProfileDialog()
  }

  function chooseProfileAvatar() {
    if (profileDialog.loading || profileDialog.uploading) return
    uni.chooseImage({
      count: 1,
      sizeType: ['compressed'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const filePath = Array.isArray(res.tempFilePaths) ? res.tempFilePaths[0] : ''
        if (!filePath) {
          toast('请选择头像图片')
          return
        }
        uploadProfileAvatar(filePath)
      },
      fail: (error) => {
        if (!String(error?.errMsg || '').toLowerCase().includes('cancel')) {
          toast('选择头像失败')
        }
      }
    })
  }

  async function uploadProfileAvatar(filePath) {
    if (!filePath || profileDialog.uploading) return
    profileDialog.uploading = true
    try {
      const res = await uploadProfileAvatarFile(filePath)
      const avatar = firstText(res.data?.url, res.data?.avatar, res.data?.path)
      if (!avatar) {
        toast('头像上传失败')
        return
      }
      profileDialog.avatar = avatar
      toast('头像已上传，保存后生效')
    } catch (error) {
      toast(error?.msg || error?.message || '头像上传失败')
    } finally {
      profileDialog.uploading = false
    }
  }

  function clearProfileAvatar() {
    if (profileDialog.loading || profileDialog.uploading) return
    profileDialog.avatar = ''
  }

  async function submitProfileDialog() {
    if (profileDialog.loading) return
    if (profileDialog.uploading) {
      toast('头像上传中，请稍后保存')
      return
    }
    const account = String(profileDialog.account || '').trim()
    const avatar = String(profileDialog.avatar || '').trim()
    const currentAccount = currentProfileAccount()
    const accountChanged = account !== currentAccount
    const avatarChanged = avatar !== currentProfileAvatar()
    const newPassword = String(profileDialog.newPassword || '').trim()
    const confirmPassword = String(profileDialog.confirmPassword || '').trim()
    const currentPassword = String(profileDialog.currentPassword || '').trim()
    const passwordChanging = Boolean(newPassword || confirmPassword)
    if (!account) {
      toast('请填写账号')
      return
    }
    if ((accountChanged || passwordChanging) && !currentPassword) {
      toast('修改账号或密码时请填写当前密码')
      return
    }
    if (passwordChanging) {
      if (newPassword.length < 6) {
        toast('新密码至少 6 位')
        return
      }
      if (newPassword !== confirmPassword) {
        toast('两次输入的新密码不一致')
        return
      }
    }
    if (!accountChanged && !avatarChanged && !passwordChanging) {
      toast('没有需要保存的修改')
      return
    }
    profileDialog.loading = true
    try {
      if (accountChanged || avatarChanged) {
        const res = await updateProfileInfo({ account, avatar, currentPassword })
        applyProfileUser(res.data?.user || res.data, currentAccount)
      }
      if (passwordChanging) {
        await changeProfilePassword({
          currentPassword,
          newPassword,
          confirmPassword
        })
      }
      resetProfileDialog()
      toast('个人中心已保存')
    } catch (error) {
      toast(error?.msg || '保存失败')
    } finally {
      profileDialog.loading = false
    }
  }

  function applyProfileUser(user, oldAccount = '') {
    if (!user || !state.user) return
    const previousAccount = firstText(oldAccount, currentProfileAccount())
    const nextAccount = firstText(user.account, user.id, previousAccount)
    const nextUser = {
      ...state.user,
      ...user,
      id: nextAccount,
      account: nextAccount,
      avatar: firstText(user.avatar, user.avatarUrl, user.pic, user.userPic)
    }
    state.user = nextUser
    if (previousAccount && nextAccount && previousAccount !== nextAccount) {
      replaceLocalAccountReferences(previousAccount, nextAccount)
    }
    upsertProfileUser(nextUser, previousAccount)
  }

  function upsertProfileUser(user, oldAccount = '') {
    const next = sanitizeUsers([user])[0]
    if (!next?.id) return
    let replaced = false
    state.users = state.users.map((item) => {
      if (sameUserId(item.id, next.id) || sameUserId(item.id, oldAccount)) {
        replaced = true
        return { ...item, ...next }
      }
      return item
    })
    if (!replaced) {
      state.users.push(next)
    }
  }

  function replaceLocalAccountReferences(oldAccount, nextAccount) {
    for (const review of state.reviews) {
      for (const key of ['employeeId', 'managerId', 'hrbpId', 'hrbpReviewerId']) {
        if (sameUserId(review[key], oldAccount)) {
          review[key] = nextAccount
        }
      }
      for (const history of review.histories || []) {
        if (sameUserId(history.byAccount, oldAccount)) {
          history.byAccount = nextAccount
        }
      }
    }
  }

  return {
    profileDialog,
    profileAvatarPreview,
    profileDisplayName,
    profileInitial,
    chooseProfileAvatar,
    clearProfileAvatar,
    closeProfileDialog,
    openProfileDialog,
    resetProfileDialog,
    submitProfileDialog
  }
}
