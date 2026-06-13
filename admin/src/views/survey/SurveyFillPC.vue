<template>
  <div class="fill-page">
    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="survey" class="fill-container">
      <div class="header">
        <h1>{{ survey.title }}</h1>
        <p v-if="survey.description" class="desc">{{ survey.description }}</p>
        <div class="meta">
          <el-tag v-if="survey.anonymous === 1" size="small">匿名收集</el-tag>
          <el-tag v-if="survey.showResult === 1" size="small">提交后查看结果</el-tag>
          <el-tag size="small">{{ questions.length }} 道题</el-tag>
        </div>
      </div>
      <div class="q-list">
        <div v-for="(q, i) in questions" :key="q.id" class="q-item">
          <div class="q-title">
            <span class="q-num">{{ i + 1 }}.</span>
            <span class="q-text" v-html="q.title" />
            <span v-if="q.required" class="q-req">*</span>
          </div>
          <div class="preview-body">
            <el-input v-if="['input','text'].includes(q.type)" v-model="answers[q.id]" :placeholder="q.placeholder || '请输入'" />
            <div v-else-if="q.type==='multiInput'" class="field-vertical">
              <el-input v-for="(f, fi) in (q.props?.fields||[])" :key="fi" v-model="answers[q.id][fi]" :placeholder="f.placeholder||'请输入'" />
            </div>
            <div v-else-if="q.type==='hInput'" class="field-horizontal">
              <el-input v-for="(f, fi) in (q.props?.fields||[])" :key="fi" v-model="answers[q.id][fi]" :placeholder="f.placeholder||'请输入'" />
            </div>
            <el-input v-else-if="q.type==='textarea'" v-model="answers[q.id]" type="textarea" :rows="3" :placeholder="q.placeholder || '请输入'" />
            <el-input-number v-else-if="q.type==='number'" v-model="answers[q.id]" style="width:100%;--el-input-width:100%" />
            <el-radio-group v-else-if="q.type==='radio'" v-model="answers[q.id]" class="preview-options preview-radio-group" :style="optionGrid(q)">
              <el-radio v-for="o in (q.props?.options||[])" :key="o.value" :value="o.value"><span v-html="o.label" /></el-radio>
            </el-radio-group>
            <el-checkbox-group v-else-if="q.type==='checkbox'" v-model="answers[q.id]" class="preview-options preview-checkbox-group" :style="optionGrid(q)">
              <el-checkbox v-for="o in (q.props?.options||[])" :key="o.value" :value="o.value"><span v-html="o.label" /></el-checkbox>
            </el-checkbox-group>
            <el-select v-else-if="q.type==='select'" v-model="answers[q.id]" placeholder="请选择" style="width:100%" clearable :teleported="false">
              <el-option v-for="o in (q.props?.options||[])" :key="o.value" :value="o.value"><span v-html="o.label" /></el-option>
            </el-select>
            <el-select v-else-if="q.type==='picker'" v-model="answers[q.id]" placeholder="请选择" style="width:100%" clearable :teleported="false">
              <el-option v-for="o in (q.props?.options||[])" :key="o.value" :value="o.value"><span v-html="o.label" /></el-option>
            </el-select>
            <el-cascader v-else-if="q.type==='cascade'" v-model="answers[q.id]" placeholder="请选择" style="width:100%" :options="q.props?.options||[]" clearable />
            <el-radio-group v-else-if="q.type==='judge'" v-model="answers[q.id]" class="preview-options preview-radio-group">
              <el-radio value="true">对</el-radio>
              <el-radio value="false">错</el-radio>
            </el-radio-group>
            <div v-else-if="q.type==='rating'" style="padding:4px 0">
              <el-rate v-model="answers[q.id]" :max="q.props?.maxRating || 5" />
            </div>
            <div v-else-if="q.type==='nps'" class="preview-nps">
              <div class="nps-labels"><span>0</span><span>10</span></div>
              <el-rate v-model="answers[q.id]" :max="10" show-score score-template="{value}" />
            </div>
            <el-date-picker v-else-if="q.type==='date'" v-model="answers[q.id]" type="date" placeholder="选择日期" style="width:100%" />
            <el-time-picker v-else-if="q.type==='time'" v-model="answers[q.id]" placeholder="选择时间" style="width:100%" />
            <el-switch v-else-if="q.type==='switch'" v-model="answers[q.id]" />
            <el-divider v-else-if="q.type==='divider'" style="margin:4px 0" />
            <div v-else-if="q.type==='description'" class="preview-plain" v-html="q.description" />
            <div v-else-if="q.type==='file'" class="file-upload-wrap">
              <input :ref="el => { if(el) fileInputs[q.id]=el as HTMLInputElement }" type="file" :multiple="q.props?.multiple !== false" style="display:none" @change="(e: any) => onFileInput(q.id, e)" />
              <el-button text @click="triggerFileInput(q.id)"><svg viewBox="0 0 1024 1024" width="16" height="16" fill="currentColor" style="vertical-align:middle;margin-right:4px"><path d="M854.6 288.6L639.4 73.4c-6-6-14.1-9.4-22.6-9.4H192c-17.7 0-32 14.3-32 32v832c0 17.7 14.3 32 32 32h640c17.7 0 32-14.3 32-32V311.3c0-8.5-3.4-16.7-9.4-22.7z"/></svg>选择文件</el-button>
              <div v-if="(fileLists[q.id]||[]).length" class="file-list">
                <div v-for="(f, fi) in fileLists[q.id]" :key="fi" class="file-item"><span class="file-name">{{ f.name }}</span><el-button text size="small" type="danger" @click="removeFile(q.id, fi)" style="padding:0 4px">×</el-button></div>
              </div>
            </div>
            <div v-else-if="q.type==='location'" class="preview-location">
              <div v-if="answers[q.id]" class="location-result">{{ answers[q.id] }}</div>
              <el-button v-else text @click="pickLocation(q.id)">
                <svg viewBox="0 0 1024 1024" width="16" height="16" fill="currentColor" style="vertical-align:middle;margin-right:4px"><path d="M512 64C367.2 64 248 183.2 248 328c0 163.2 233.6 524.8 252 551.2 3.2 4.8 8 7.2 12 7.2s8.8-2.4 12-7.2C542.4 852.8 776 491.2 776 328 776 183.2 656.8 64 512 64z m0 400c-39.2 0-72-32.8-72-72s32.8-72 72-72 72 32.8 72 72-32.8 72-72 72z"/></svg>选择位置
              </el-button>
            </div>
            <el-input v-else-if="q.type==='phone'" placeholder="手机号" v-model="answers[q.id]" />
            <el-input v-else-if="q.type==='email'" placeholder="邮箱地址" v-model="answers[q.id]" />
            <el-input v-else-if="q.type==='idCard'" placeholder="身份证号" v-model="answers[q.id]" />
            <el-input v-else-if="q.type==='password'" type="password" placeholder="密码" v-model="answers[q.id]" />
            <el-cascader v-else-if="q.type==='user'||q.type==='dept'" v-model="answers[q.id]" :placeholder="q.type==='user'?'选择成员':'选择部门'" style="width:100%" :options="buildUserDeptTree(q)" :props="{ multiple: !!q.multiple, emitPath: false }" clearable />
            <div v-else-if="q.type==='dateRange'" class="field-vertical">
              <el-date-picker v-model="answers[q.id][0]" type="date" placeholder="开始日期" style="width:100%" />
              <el-date-picker v-model="answers[q.id][1]" type="date" placeholder="结束日期" style="width:100%" />
            </div>
            <div v-else-if="q.type==='matrixRadio'" class="preview-matrix">
              <table><thead><tr><th class="corner">行\\列</th><th v-for="c in (q.props?.columns||[{title:'列A'},{title:'列B'}])" :key="typeof c==='string'?c:(c.title||c.label)">{{ typeof c==='string'?c:(c.title||c.label) }}</th></tr></thead><tbody><tr v-for="(r, ri) in (q.props?.rows||[{title:'行1'},{title:'行2'}])" :key="typeof r==='string'?r:r.title"><td class="matrix-label">{{ typeof r==='string'?r:r.title }}</td><td v-for="c in (q.props?.columns||[{title:'列A'},{title:'列B'}])" :key="typeof c==='string'?c:(c.title||c.label)"><el-radio-group :model-value="answers[q.id][ri]" @update:model-value="(v: any) => answers[q.id][ri] = v"><el-radio :value="typeof c==='string'?c:(c.title||c.label)" /></el-radio-group></td></tr></tbody></table>
            </div>
            <div v-else-if="q.type==='matrixCheckbox'" class="preview-matrix">
              <table><thead><tr><th class="corner">行\\列</th><th v-for="c in (q.props?.columns||[{title:'列A'},{title:'列B'}])" :key="typeof c==='string'?c:(c.title||c.label)">{{ typeof c==='string'?c:(c.title||c.label) }}</th></tr></thead><tbody><tr v-for="(r, ri) in (q.props?.rows||[{title:'行1'},{title:'行2'}])" :key="typeof r==='string'?r:r.title"><td class="matrix-label">{{ typeof r==='string'?r:r.title }}</td><td v-for="c in (q.props?.columns||[{title:'列A'},{title:'列B'}])" :key="typeof c==='string'?c:(c.title||c.label)"><el-checkbox-group :model-value="answers[q.id][ri]||[]" @update:model-value="(v: any) => answers[q.id][ri] = v"><el-checkbox :value="typeof c==='string'?c:(c.title||c.label)" /></el-checkbox-group></td></tr></tbody></table>
            </div>
            <div v-else-if="q.type==='matrixFillBlank'" class="preview-matrix">
              <table><thead><tr><th class="corner">行\\列</th><th v-for="c in (q.props?.columns||[{title:'列A'},{title:'列B'}])" :key="typeof c==='string'?c:(c.title||c.label)">{{ typeof c==='string'?c:(c.title||c.label) }}</th></tr></thead><tbody><tr v-for="(r, ri) in (q.props?.rows||[{title:'行1'},{title:'行2'}])" :key="typeof r==='string'?r:r.title"><td class="matrix-label">{{ typeof r==='string'?r:r.title }}</td><td v-for="(c, ci) in (q.props?.columns||[{title:'列A'},{title:'列B'}])" :key="typeof c==='string'?c:(c.title||c.label)"><el-input :model-value="answers[q.id][ri]?.[ci]" @update:model-value="(v: any) => { if(!answers[q.id][ri]) answers[q.id][ri]={}; answers[q.id][ri][ci]=v }" placeholder="填空" size="small" style="width:100%" /></td></tr></tbody></table>
            </div>
            <div v-else-if="q.type==='matrixAuto'" class="preview-matrix">
              <table><thead><tr><th class="corner">#</th><th v-for="c in (q.props?.columns||[])" :key="c.label||c.id||c">{{ c.label||c }}</th><th style="width:40px"></th></tr></thead><tbody><tr v-for="(r, ri) in (answers[q.id]||[])" :key="ri"><td class="matrix-label">{{ ri+1 }}</td><td v-for="(c, ci) in (q.props?.columns||[])" :key="c.label||c.id||c"><el-input v-model="answers[q.id][ri][ci]" size="small" :placeholder="c.label||'值'" style="width:100%" /></td><td><el-button text size="small" type="danger" @click="removeMatrixAutoRow(q.id, ri)" style="padding:2px"><svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="9"/><line x1="8" y1="12" x2="16" y2="12"/></svg></el-button></td></tr></tbody></table>
              <div style="display:flex;align-items:center;gap:8px;padding:6px 8px"><el-button size="small" text @click="addMatrixAutoRow(q.id)">+ 添加行</el-button></div>
            </div>
            <div v-else-if="q.type==='questionSet'" class="preview-plain">问题组（内部题）</div>
            <div v-else-if="q.type==='pagination'" class="preview-plain">—— 分页 ——</div>
            <div v-else-if="q.type==='richText'" style="border:1px solid #dcdfe6;border-radius:4px;overflow:hidden">
              <QuillEditor v-model:content="answers[q.id]" content-type="html" :options="{ theme: 'snow', placeholder: q.placeholder || '输入富文本内容...' }" style="min-height:150px" />
            </div>
            <el-input v-else-if="q.type==='scanCode'" v-model="answers[q.id]" placeholder="扫码" class="scan-code-input">
              <template #prefix><svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg></template>
              <template #suffix>
                <el-button text type="primary" size="small" @click="openScanner(q.id)" style="margin-right:-8px;height:28px">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 7V4a2 2 0 0 1 2-2h3M1 17v3a2 2 0 0 0 2 2h3M23 7V4a2 2 0 0 0-2-2h-3M23 17v3a2 2 0 0 1-2 2h-3"/><rect x="8" y="8" width="8" height="8" rx="1"/></svg>
                </el-button>
              </template>
            </el-input>
            <div v-else-if="q.type==='signature'" class="signature-pad-wrap">
              <canvas :ref="el => sigCanvasMap[q.id] = el as HTMLCanvasElement" class="sig-canvas" @mousedown="sigStart($event, q.id)" @mousemove="sigMove($event, q.id)" @mouseup="sigEnd" @mouseleave="sigEnd" @touchstart.prevent="e => sigTouchStart(e, q.id)" @touchmove.prevent="e => sigTouchMove(e, q.id)" @touchend="sigEnd"></canvas>
              <div style="display:flex;gap:8px;margin-top:4px"><el-button size="small" text @click="clearSignature(q.id)">清除</el-button></div>
            </div>
            <el-input v-else v-model="answers[q.id]" placeholder="请输入" />
          </div>
        </div>
      </div>
      <div class="footer">
        <el-button type="primary" size="large" :loading="submitting" @click="onSubmit">提交</el-button>
      </div>
    </div>
  </div>
  <el-dialog v-if="showScanner" v-model="showScanner" title="扫码" width="400px" :close-on-click-modal="false" destroy-on-close @opened="onScannerOpen" @close="onScannerClose">
    <div ref="scannerRef" style="width:100%;aspect-ratio:1;overflow:hidden;background:#000;border-radius:8px"></div>
    <template #footer>
      <el-button @click="showScanner = false">取消</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, nextTick, onBeforeUnmount } from 'vue'
import { Html5Qrcode } from 'html5-qrcode'
import { QuillEditor } from '@vueup/vue-quill'
import '@vueup/vue-quill/dist/vue-quill.snow.css'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

const route = useRoute()
const survey = ref<any>(null)
const questions = ref<any[]>([])
const answers = ref<any>({})
const loading = ref(true)
const error = ref('')
const submitting = ref(false)
const session = ref('')
const showScanner = ref(false)
const scanQid = ref('')
const scannerRef = ref<HTMLDivElement>()
let scanner: Html5Qrcode | null = null
const fileLists: Record<string, File[]> = reactive({})
const fileInputs: Record<string, HTMLInputElement> = {}
const sigCanvasMap: Record<string, HTMLCanvasElement> = {}
const sigDrawing = ref(false)
const sigCurId = ref('')

const API_BASE = import.meta.env.VITE_API_BASE || ''

async function apiGet(path: string) {
  const res = await fetch(`${API_BASE}${path}`)
  return res.json()
}

async function apiPost(path: string, data: any) {
  const res = await fetch(`${API_BASE}${path}`, { method: 'POST', body: JSON.stringify(data), headers: { 'Content-Type': 'application/json' } })
  return res.json()
}

function getInitVal(q: any): any {
  const type = q.type
  if (type === 'checkbox') return []
  if (type === 'switch') return false
  if (type === 'number') return undefined
  if (type === 'rating') return 0
  if (type === 'nps') return 0
  if (type === 'dateRange') return ['', '']
  if (['matrixRadio','matrixCheckbox','matrixFillBlank'].includes(type)) return {}
  if (type === 'matrixAuto') return (q.props?.rows||[]).map(() => (q.props?.columns||[]).map(() => ''))
  if (['multiInput','hInput'].includes(type)) return (q.props?.fields||[]).map(() => '')
  if (['user','dept'].includes(type)) return q.multiple ? [] : ''
  if (['cascade','picker'].includes(type)) return []
  if (type === 'file') return []
  return ''
}

function optionGrid(q: any) {
  const cols = q.optionLayout
  if (!cols || cols <= 1) return {}
  return { display: 'grid', gridTemplateColumns: `repeat(${cols}, 1fr)`, gap: '4px' }
}

function buildUserDeptTree(q: any) {
  const opts = q.props?.options || []
  if (q.type === 'user') {
    const deptMap: Record<string, any> = {}
    opts.forEach((o: any) => {
      const deptId = o.deptId || ''
      if (!deptMap[deptId]) {
        deptMap[deptId] = { value: '__d__' + deptId, label: o.deptName || deptId || '未分组', children: [] }
      }
      deptMap[deptId].children.push({ value: o.value, label: o.label || '成员' })
    })
    return Object.values(deptMap)
  }
  const map: Record<string, any> = {}
  opts.forEach((o: any) => { map[o.value] = { ...o, children: [] } })
  const roots: any[] = []
  opts.forEach((o: any) => {
    if (o.parentId && map[o.parentId]) {
      map[o.parentId].children.push(map[o.value])
    } else {
      roots.push(map[o.value])
    }
  })
  return roots
}

function addMatrixAutoRow(qid: string) {
  const q = questions.value.find(x => x.id === qid)
  if (!q) return
  const cols = q.props?.columns?.length || 0
  if (!answers.value[qid]) answers.value[qid] = []
  answers.value[qid].push(Array(cols).fill(''))
}
function removeMatrixAutoRow(qid: string, ri: number) {
  if (!answers.value[qid]) return
  answers.value[qid].splice(ri, 1)
}
function triggerFileInput(qid: string) {
  fileInputs[qid]?.click()
}
function onFileInput(qid: string, e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files?.length) return
  if (!fileLists[qid]) fileLists[qid] = []
  for (const f of input.files) fileLists[qid].push(f)
  input.value = ''
  answers.value[qid] = fileLists[qid].map(f => f.name)
}
function removeFile(qid: string, idx: number) {
  if (fileLists[qid]) fileLists[qid].splice(idx, 1)
  answers.value[qid] = fileLists[qid]?.length ? fileLists[qid].map(f => f.name) : ''
}
function pickLocation(qid: string) {
  if (!navigator.geolocation) { ElMessage.warning('浏览器不支持定位'); return }
  navigator.geolocation.getCurrentPosition(
    (pos) => { answers.value[qid] = `${pos.coords.latitude},${pos.coords.longitude}` },
    () => { ElMessage.warning('定位失败，请检查权限设置') },
    { enableHighAccuracy: true, timeout: 10000 }
  )
}

function initSigCanvas(id: string) {
  const c = sigCanvasMap[id]
  if (!c) return
  if (c.offsetWidth && c.offsetWidth !== c.width) c.setAttribute('width', String(c.offsetWidth))
  if (c.offsetHeight && c.offsetHeight !== c.height) c.setAttribute('height', String(c.offsetHeight))
  const ctx = c.getContext('2d')
  if (ctx) { ctx.strokeStyle = '#333'; ctx.lineWidth = 3; ctx.lineCap = 'round' }
}
function sigPos(e: MouseEvent | Touch, c: HTMLCanvasElement) {
  const rect = c.getBoundingClientRect()
  return {
    x: (e.clientX - rect.left) * (c.width / rect.width),
    y: (e.clientY - rect.top) * (c.height / rect.height)
  }
}
function sigStart(e: MouseEvent | Touch, id: string) {
  const c = sigCanvasMap[id]
  if (!c) return
  const p = sigPos(e, c)
  sigCurId.value = id
  sigDrawing.value = true
  const ctx = c.getContext('2d')
  if (ctx) { ctx.beginPath(); ctx.moveTo(p.x, p.y) }
}
function sigMove(e: MouseEvent | Touch, id: string) {
  if (!sigDrawing.value || sigCurId.value !== id) return
  const c = sigCanvasMap[id]
  if (!c) return
  const p = sigPos(e, c)
  const ctx = c.getContext('2d')
  if (ctx) { ctx.lineTo(p.x, p.y); ctx.stroke() }
}
function sigEnd() { sigDrawing.value = false; sigCurId.value = '' }
function sigTouchStart(e: TouchEvent, id: string) {
  if (e.touches[0]) sigStart(e.touches[0], id)
}
function sigTouchMove(e: TouchEvent, id: string) {
  if (e.touches[0]) sigMove(e.touches[0], id)
}
function clearSignature(id: string) {
  const c = sigCanvasMap[id]
  if (!c) return
  const ctx = c.getContext('2d')
  if (ctx) ctx.clearRect(0, 0, c.width, c.height)
}

function openScanner(qid: string) {
  scanQid.value = qid
  showScanner.value = true
}
function onScannerOpen() {
  if (!scannerRef.value) return
  scanner = new Html5Qrcode('scanner-ref-fallback')
  scannerRef.value.id = 'scanner-ref-fallback'
  scanner.start(
    { facingMode: 'environment' },
    { fps: 10, qrbox: { width: 250, height: 250 } },
    (decodedText) => {
      if (scanQid.value) answers.value[scanQid.value] = decodedText
      showScanner.value = false
    },
    () => {}
  ).catch(() => {})
}
function onScannerClose() {
  if (scanner) {
    scanner.stop().catch(() => {})
    scanner = null
  }
}

async function load() {
  const id = route.params.id
  if (!id) { error.value = '参数错误'; loading.value = false; return }
  // 从 localStorage 恢复 session，刷新时不丢失
  session.value = localStorage.getItem('survey_session_' + id) || ''
  try {
    const res = await apiGet(`/survey/view?id=${id}&session=${session.value}`)
    if (res.code !== 0) { error.value = res.msg || '加载失败'; loading.value = false; return }
    survey.value = res.data
    // 更新 session（首次获取或后端重用时）
    if (res.data?.session) {
      session.value = res.data.session
      localStorage.setItem('survey_session_' + id, session.value)
    }
    const raw = res.data?.schema
    const sch = raw ? (typeof raw === 'string' ? JSON.parse(raw) : raw) : { questions: [] }
    questions.value = sch.questions || []
    const init: any = {}
    questions.value.forEach((q: any) => {
      init[q.id] = getInitVal(q)
    })
    answers.value = init
    await nextTick()
    questions.value.filter((q: any) => q.type === 'signature').forEach((q: any) => initSigCanvas(q.id))
  } catch { error.value = '加载失败' }
  finally { loading.value = false }
}

async function onSubmit() {
  const id = Number(route.params.id)
  try {
    const vr = await apiPost('/survey/validate', { surveyId: id, answers: answers.value })
    if (vr.data && !vr.data.ok) {
      const msgs = (vr.data.errors || []).map((e: any) => e.message).join('; ')
      ElMessage.warning(msgs || '请检查填写内容')
      return
    }
  } catch {}

  ElMessageBox.confirm('确认提交？提交后不可修改', '提示', { type: 'info' }).then(async () => {
    submitting.value = true
    try {
      const res = await apiPost('/survey/submit', { surveyId: id, answers: answers.value, device: navigator.userAgent, session: session.value })
      if (res.code !== 0) {
        ElMessage.error(res.msg || '提交失败')
      } else {
        ElMessage.success('已提交')
        localStorage.removeItem('survey_session_' + id)
        survey.value = null
        questions.value = []
      }
    } catch (e: any) { ElMessage.error(e.msg || '提交失败') }
    finally { submitting.value = false }
  }).catch(() => {})
}

onMounted(load)
</script>

<style scoped>
.fill-page { min-height:100vh; background:#f5f6f8; padding:40px 0; }
.fill-container { max-width:800px; margin:0 auto; }
.loading, .error { text-align:center; padding:100px 0; color:#999; font-size:16px; }
.header { background:#fff; border-radius:12px; padding:32px 36px; margin-bottom:20px; box-shadow:0 2px 12px rgba(0,0,0,0.04); }
.header h1 { font-size:24px; font-weight:600; color:#1a1a1a; margin:0 0 8px; }
.header .desc { font-size:14px; color:#666; line-height:1.6; margin:0 0 12px; }
.header .meta { display:flex; gap:8px; }
.q-list { display:flex; flex-direction:column; gap:12px; }
.q-item { background:#fff; border-radius:12px; padding:24px 28px; box-shadow:0 2px 12px rgba(0,0,0,0.04); }

.q-title { font-size:15px; color:#333; margin-bottom:16px; font-weight:500; word-break:break-word; }
.q-num { color:#409eff; font-weight:600; }
.q-text { word-break:break-word; }
.q-text :deep(p), .q-text :deep(div), .q-text :deep(h1), .q-text :deep(h2), .q-text :deep(h3), .q-text :deep(h4), .q-text :deep(h5), .q-text :deep(h6), .q-text :deep(blockquote), .q-text :deep(ul), .q-text :deep(ol), .q-text :deep(li) { display:inline; }
.q-req { color:#f56c6c; margin-left:2px; }

.preview-body { width:100%; min-height:28px; }
.preview-body .el-input,
.preview-body .el-select,
.preview-body .el-date-editor,
.preview-body .el-cascader,
.preview-body .el-input-number { width:100%; }
.field-vertical { display:flex; flex-direction:column; gap:4px; width:100%; }
.field-horizontal { display:flex; flex-wrap:wrap; gap:4px; width:100%; }
.field-horizontal .el-input { flex:1; min-width:120px; }

.preview-options { display:flex; flex-direction:column; gap:4px; width:100%; }
.preview-radio-group,
.preview-checkbox-group { gap:4px; align-items:flex-start; text-align:left; }
.preview-plain { font-size:13px; color:#606266; white-space:pre-wrap; }

.preview-nps { padding:4px 0; }
.preview-nps .nps-labels { display:flex; justify-content:space-between; font-size:12px; color:#909399; margin-bottom:2px; }

.preview-matrix { border:1px solid #d0d0d0; border-radius:6px; overflow-x:auto; }
.preview-matrix table { width:100%; border-collapse:collapse; font-size:13px; }
.preview-matrix th,
.preview-matrix td { border:1px solid #e0e0e0; padding:6px 10px; text-align:center; }
.preview-matrix th { background:#f0f2f5; font-weight:500; color:#303133; }
.preview-matrix .corner { background:#f0f2f5; color:#909399; font-size:12px; }
.preview-matrix .matrix-label { background:#fafafa; font-weight:500; text-align:center; min-width:60px; white-space:nowrap; }

.file-upload-placeholder { border:1px dashed #d9d9d9; border-radius:6px; padding:12px; text-align:center; color:#999; margin:4px 0; }
.file-upload-placeholder svg { display:block; margin:0 auto 4px; }
.file-upload-placeholder span { font-size:13px; }
.file-upload-wrap { border:1px dashed #d9d9d9; border-radius:6px; padding:12px; }
.file-list { margin-top:8px; display:flex; flex-direction:column; gap:4px; }
.file-item { display:flex; align-items:center; gap:6px; font-size:13px; padding:2px 0; }
.file-name { color:#303133; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; flex:1; }
.location-result { font-size:13px; color:#409eff; }

.signature-pad-wrap { width:100%; }
.sig-canvas { border:1px dashed #d9d9d9; border-radius:6px; width:100%; height:120px; cursor:crosshair; }
.scan-code-input :deep(.el-input__wrapper) { width:100%; }
.preview-body :deep(.el-select-dropdown__item) { height:auto; min-height:28px; padding:4px 10px; white-space:normal; line-height:1.3; }

.footer { text-align:center; padding:24px 0; }
</style>
