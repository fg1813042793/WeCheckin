import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(
  new URL('../src/pages/workflow/components/WorkflowTextarea.vue', import.meta.url),
  'utf8',
)

for (const snippet of [
  'const nativeTextareaTag = \'textarea\'',
  'const composing = ref(false)',
  'function handleCompositionStart()',
  'function handleCompositionEnd(event: unknown)',
  'function handleNativeInput(event: unknown)',
  'if (composing.value || nativeInputComposing(event))',
  '@compositionstart="handleCompositionStart"',
  '@compositionend="handleCompositionEnd"',
  '@input="handleNativeInput"',
  'class="workflow-textarea__native-control"',
  'resize: none;',
]) {
  assert.ok(source.includes(snippet), `WorkflowTextarea.vue missing IME-safe H5 behavior: ${snippet}`)
}

const h5Branch = source.match(/<!-- #ifdef H5 -->[\s\S]*?<!-- #endif -->/)?.[0] || ''
assert.ok(h5Branch.includes(':is="nativeTextareaTag"'), 'H5 workflow textarea must use the browser-native textarea')
assert.ok(!h5Branch.includes('<u-textarea'), 'H5 workflow textarea must not keep the u-textarea value echo loop')

const crossPlatformBranch = source.match(/<!-- #ifndef H5 -->[\s\S]*?<!-- #endif -->/)?.[0] || ''
assert.ok(crossPlatformBranch.includes('<u-textarea'), 'non-H5 platforms must keep using u-textarea')

assert.match(
  source,
  /\.workflow-textarea__native--disabled \.workflow-textarea__native-control\s*\{[^}]*color:\s*var\(--u-main-color\);[^}]*-webkit-text-fill-color:\s*var\(--u-main-color\);[^}]*cursor:\s*not-allowed;/,
  'disabled native workflow textarea must keep readable filled text',
)

assert.match(
  source,
  /\.workflow-textarea__native--disabled \.workflow-textarea__native-control::placeholder\s*\{[^}]*color:\s*var\(--u-light-color\);[^}]*-webkit-text-fill-color:\s*var\(--u-light-color\);/,
  'disabled native workflow textarea placeholder must remain muted',
)

assert.match(
  source,
  /:deep\(\.workflow-textarea__control\.u-textarea--disabled \.u-textarea__field\)[\s\S]*?color:\s*var\(--u-main-color\) !important;[\s\S]*?-webkit-text-fill-color:\s*var\(--u-main-color\) !important;/,
  'disabled uView workflow textarea must keep readable filled text',
)

assert.match(
  source,
  /:deep\(\.workflow-textarea__control\.u-textarea--disabled \.u-textarea__field::placeholder\)[\s\S]*?color:\s*var\(--u-light-color\) !important;[\s\S]*?-webkit-text-fill-color:\s*var\(--u-light-color\) !important;/,
  'disabled uView workflow textarea placeholder must remain muted',
)

console.log('Workflow textarea IME checks passed')
