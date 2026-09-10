import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const projectPackage = JSON.parse(readFileSync(resolve(root, 'package.json'), 'utf8'))
const uniAppVersion = projectPackage.devDependencies?.['@dcloudio/uni-app']
const uniAppPlusVersion = projectPackage.devDependencies?.['@dcloudio/uni-app-plus']
const uniUtsVersion = projectPackage.devDependencies?.['@dcloudio/uni-uts-v1']
if (!uniAppPlusVersion || uniAppPlusVersion !== uniAppVersion) {
  throw new Error('@dcloudio/uni-app-plus must be installed at the same version as @dcloudio/uni-app')
}
if (!uniUtsVersion || uniUtsVersion !== uniAppVersion) {
  throw new Error('@dcloudio/uni-uts-v1 must be installed at the same version as @dcloudio/uni-app')
}
const files = {
  plugin: 'uni_modules/wecheckin-dingtalk-auth/package.json',
  contract: 'uni_modules/wecheckin-dingtalk-auth/utssdk/interface.uts',
  android: 'uni_modules/wecheckin-dingtalk-auth/utssdk/app-android/index.uts',
  callback: 'uni_modules/wecheckin-dingtalk-auth/utssdk/app-android/DDAuthActivity.kt',
  config: 'uni_modules/wecheckin-dingtalk-auth/utssdk/app-android/config.json',
  androidManifest: 'uni_modules/wecheckin-dingtalk-auth/utssdk/app-android/AndroidManifest.xml',
  ios: 'uni_modules/wecheckin-dingtalk-auth/utssdk/app-ios/index.uts',
}

for (const [name, path] of Object.entries(files)) {
  if (!existsSync(resolve(root, path))) {
    throw new Error(`DingTalk native auth ${name} file missing: ${path}`)
  }
}

const manifest = readFileSync(resolve(root, 'manifest.json'), 'utf8')
const login = readFileSync(resolve(root, 'pages/login/login.vue'), 'utf8')
const api = readFileSync(resolve(root, 'api/index.js'), 'utf8')
const android = readFileSync(resolve(root, files.android), 'utf8')
const callback = readFileSync(resolve(root, files.callback), 'utf8')
const config = readFileSync(resolve(root, files.config), 'utf8')
const androidManifest = readFileSync(resolve(root, files.androidManifest), 'utf8')

for (const snippet of [
  '"packagename" : "com.wecheckin.app"',
]) {
  if (!manifest.includes(snippet)) throw new Error(`Android manifest config missing: ${snippet}`)
}
for (const snippet of [
  'com.alibaba.android:ddopenauth:1.5.0.8',
]) {
  if (!config.includes(snippet)) throw new Error(`DingTalk native dependency missing: ${snippet}`)
}
for (const snippet of [
  'com.alibaba.android.rimet',
  'android:name="${applicationId}.ddauth.DDAuthActivity"',
  'android:targetActivity="com.wecheckin.app.ddauth.DDAuthCallbackActivity"',
  'android:exported="true"',
]) {
  if (!androidManifest.includes(snippet)) throw new Error(`DingTalk native AndroidManifest missing: ${snippet}`)
}
for (const snippet of [
  "import AuthLoginParam from 'com.android.dingtalk.openauth.AuthLoginParam'",
  "import DDAuthApiFactory from 'com.android.dingtalk.openauth.DDAuthApiFactory'",
  'DDAuthApiFactory.createDDAuthApi',
  'authApi.authLogin()',
  'consumeDingTalkAuthResult',
]) {
  if (!android.includes(snippet)) throw new Error(`DingTalk native UTS implementation missing: ${snippet}`)
}
for (const snippet of [
  'package com.wecheckin.app.ddauth',
  'class DDAuthCallbackActivity : Activity()',
  'DDAuthConstant.CALLBACK_EXTRA_AUTH_CODE',
  'DDAuthConstant.CALLBACK_EXTRA_STATE',
  'DDAuthConstant.CALLBACK_EXTRA_ERROR',
]) {
  if (!callback.includes(snippet)) throw new Error(`DingTalk native callback Activity missing: ${snippet}`)
}
for (const snippet of [
  "@/uni_modules/wecheckin-dingtalk-auth",
  '// #ifdef APP-PLUS',
  'startDingTalkNativeAuth',
  'completeDingTalkNativeAuth',
  'consumeDingTalkAuthResult',
  'getDingTalkNativeRedirectURI',
  "const clientId = String(data.clientId || '').trim()",
  "const nativeRedirectUri = String(data.redirectUri || '').trim()",
  "const state = String(data.state || '').trim()",
  "const data = responseData.clientId || responseData.redirectUri || responseData.state",
]) {
  if (!login.includes(snippet)) throw new Error(`login page native DingTalk flow missing: ${snippet}`)
}
if (login.includes('// #ifdef APP-ANDROID')) {
  throw new Error('generic uni-app App builds must retain the native auth proxy via APP-PLUS')
}
for (const forbidden of [
  'dingtalk://dingtalkclient/page/link',
  'plus.runtime.openURL',
  'hbuilder://',
  '/api/v2/auth/dingtalk-app-callback',
]) {
  if (login.includes(forbidden) || readFileSync(resolve(root, 'utils/dingtalkOAuth.js'), 'utf8').includes(forbidden)) {
    throw new Error(`Android native DingTalk login must not retain the App H5 bridge: ${forbidden}`)
  }
}
for (const snippet of ['clientId', 'redirectUri']) {
  if (!api.includes('dingtalkAuthorization(data)') || !login.includes(`data.${snippet}`)) {
    throw new Error(`DingTalk native authorization response field missing: ${snippet}`)
  }
}

console.log('DingTalk Android native auth checks passed.')
