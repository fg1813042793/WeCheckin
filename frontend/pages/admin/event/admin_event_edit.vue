<template>
  <view />
</template>

<script>
export default {
  onLoad(opts) {
    if (opts.id) {
      uni.redirectTo({ url: '/pages/admin/event/admin_event_add?id=' + opts.id })
    } else {
      uni.redirectTo({ url: '/pages/admin/event/admin_event_add' })
    }
  }
}
</script>
