<template>
  <view class="brand-line">
    <view class="brand-logo">
      <image class="brand-logo-img" v-if="appLogoUrl" :src="appLogoUrl" mode="aspectFill" />
      <text v-else>{{ appLogoText }}</text>
    </view>
    <view>
      <text class="brand-title">{{ appDisplayTitle }}</text>
      <text class="brand-subtitle">{{ subtitle }}</text>
    </view>
  </view>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  appConfig: {
    type: Object,
    default: () => ({})
  },
  subtitle: {
    type: String,
    default: '钉钉 OA 微应用'
  }
})

const appDisplayTitle = computed(() => firstText(props.appConfig.appTitle, props.appConfig.appName, '钉钉H5应用'))
const appLogoText = computed(() => firstText(props.appConfig.logoText, 'H5').slice(0, 4))
const appLogoUrl = computed(() => firstText(props.appConfig.logoUrl, props.appConfig.logoURL))

function firstText(...values) {
  for (const value of values) {
    const text = String(value || '').trim()
    if (text) return text
  }
  return ''
}
</script>
