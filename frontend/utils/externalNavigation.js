export function openExternalURL(url) {
  if (!url) return

  // #ifdef H5
  window.location.href = url
  return
  // #endif

  // #ifdef APP-PLUS
  plus.runtime.openURL(url, () => {
    uni.showToast({ title: '无法打开链接', icon: 'none' })
  })
  return
  // #endif

  // #ifdef MP-WEIXIN
  uni.setClipboardData({
    data: url,
    success: () => uni.showToast({ title: '链接已复制', icon: 'none' }),
    fail: () => uni.showToast({ title: '无法打开链接', icon: 'none' })
  })
  // #endif
}
