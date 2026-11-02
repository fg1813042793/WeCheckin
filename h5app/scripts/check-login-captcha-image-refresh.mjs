import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import process from 'node:process'

const componentPath = resolve(process.cwd(), 'src/components/dingtalk-login-panel/dingtalk-login-panel.vue')
const source = readFileSync(componentPath, 'utf8')

assert.match(
  source,
  /function refreshCaptcha\(\)\s*\{\s*if \(props\.captchaLoading\)\s*return\s*emit\('refresh-captcha'\)\s*\}/,
  'captcha refresh must ignore repeated clicks while a challenge is loading',
)

assert.ok(
  (source.match(/@click="refreshCaptcha"/g) || []).length >= 2,
  'captcha image and empty placeholder must both support click-to-refresh',
)
assert.doesNotMatch(source, /captcha-refresh-button/, 'login captcha must not render a separate refresh button')
assert.doesNotMatch(source, /name="reload"/, 'login captcha must not render a separate reload icon')

console.log('H5 login captcha refresh check passed: image-only refresh, loading guard and retry placeholder')
