<template>
  <div
    class="exam pc"
    :style="bgStyle"
  >
    <div v-if="settings.progressBar" class="exam-progress">
      <div class="exam-progress-inner">
        <div class="exam-progress-track">
          <div class="exam-progress-fill" :style="{ width: progressPct + '%' }" />
        </div>
        <span class="exam-progress-label">{{ answeredCount }}/{{ totalQuestions }}</span>
      </div>
    </div>

    <div v-if="loading" class="exam-loading">
      <div class="spinner">
        <div v-for="i in 5" :key="i" :class="'rect' + i" />
      </div>
    </div>

    <div v-else-if="showLogin" class="exam-login">
      <div class="exam-card exam-card--narrow">
        <div class="exam-login-header">
          <h2 class="exam-login-title">登录后参加考试</h2>
          <p class="exam-login-desc" v-if="exam?.title">{{ exam.title }}</p>
        </div>
        <el-form label-position="top" @submit.prevent="doLogin">
          <el-form-item label="账号">
            <el-input v-model="loginForm.name" placeholder="用户名 / 手机号" @keyup.enter="doLogin" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="loginForm.password" type="password" placeholder="请输入密码" show-password @keyup.enter="doLogin" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" class="exam-btn--primary" :loading="loginLoading" @click="doLogin" style="width:100%">登录</el-button>
          </el-form-item>
        </el-form>
        <div class="exam-login-footer">
          <span>还没有账号？</span>
          <a :href="`/login`" target="_blank" class="exam-login-link">去注册</a>
        </div>
      </div>
    </div>

    <div v-else-if="error" class="exam-error">{{ error }}</div>

    <!-- 提交成功页 -->
    <div v-else-if="submitted && !showResultView" class="exam-end">
      <div class="exam-end-content">
        <h3 class="exam-end-title">{{ endContent ? '' : '已交卷，感谢作答' }}</h3>
        <div v-if="endContent" class="exam-end-desc" v-html="endContent" />
        <div v-if="result" class="exam-end-result">
          <el-button v-if="settings.showAnalysis && settings.transcriptVisible !== false" type="primary" size="large" @click="showResultView = true">查看结果</el-button>
          <el-button v-else-if="settings.showAnalysis" type="primary" size="large" @click="showResultView = true">查看解析</el-button>
          <el-button v-else type="primary" size="large" @click="showResultView = true">查看成绩</el-button>
          <div v-if="settings.redirectUrl" class="exam-end-result-copy">
            <el-button type="primary" class="exam-btn--primary" @click="openRedirectUrl">查看结果</el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 成绩单 -->
    <div v-else-if="showResultView && result && (settings.transcriptVisible !== false || settings.showAnalysis)" class="exam-card result-container">
      <div v-if="settings.transcriptVisible !== false" class="exam-header" style="text-align:center">
        <h2 class="exam-title">{{ exam?.title }}</h2>
        <el-tag type="success" size="small">得分 {{ result.score }} / {{ result.fullScore }}</el-tag>
        <el-tag v-if="result.correctCnt !== undefined" type="info" size="small">正确 {{ result.correctCnt }}/{{ totalQuestions }}</el-tag>
      </div>
      <div class="exam-form-scroll">
        <div v-for="(q, i) in questions" :key="q.id" class="q-item" :data-qid="q.id">
          <div class="q-title" v-html="resultTitleHtml(q, i)" />
          <div class="q-answer" v-if="result.results?.[i]">
            <span :style="{ color: result.results[i].correct ? '#67c23a' : '#f56c6c' }">
              {{ result.results[i].correct ? '✓' : '✗' }}
            </span>
            <span style="margin-left:4px;color:#606266">你的答案：{{ formatResultAnswer(q, answers[q.id]) }}</span>
          </div>
          <div v-if="q.examCorrectAnswer && !result.results?.[i]?.correct" style="margin-top:4px;font-size:13px;color:#67c23a">
            ✓ 正确答案：<span v-html="displayAnswer(q, q.examCorrectAnswer)" />
          </div>
          <div v-if="q.examAnalysis" style="margin-top:4px;font-size:13px;color:#909399;background:#f5f7fa;padding:8px;border-radius:4px;white-space:pre-wrap">
            解析：{{ q.examAnalysis }}
          </div>
        </div>
      </div>
    </div>

    <!-- 考试填写 -->
    <div v-else-if="exam" class="exam-card">
      <div v-if="settings.headerImages?.length" class="exam-card-img"><img :src="typeof settings.headerImages[0]==='string'?settings.headerImages[0]:settings.headerImages[0].url" /></div>
      <div class="exam-header">
        <h1 class="exam-title">{{ exam.title }}</h1>
        <p v-if="exam.description" class="exam-desc">{{ exam.description }}</p>
        <div class="exam-meta">
          <el-tag v-if="exam.showScore && totalScore" size="small">满分 {{ totalScore }} 分</el-tag>
          <span class="exam-tag">{{ totalQuestions }} 道题</span>
          <span v-if="remaining > 0" class="exam-tag" :class="remaining < 60000 ? 'exam-tag--danger' : 'exam-tag--warn'">⏱ {{ formatRemaining(remaining) }}</span>
        </div>
      </div>

      <el-form ref="formRef" :model="answers" label-position="top" class="exam-form">
        <template v-if="settings.onePageOneQuestion && currentQuestion">
          <div class="exam-form-scroll" @touchstart="onSwipeStart" @touchend="onSwipeEnd">
            <QuestionField
              :q="currentQuestion"
              :index="currentNavIndex"
              :score="exam.showScore && currentQuestion.examScore ? currentQuestion.examScore : 0"
              :answers="answers"
              :settings="settings"
              :file-lists="fileLists"
              :file-inputs="fileInputs"
              :sig-canvas-map="sigCanvasMap"
              @update:answers="answers = $event"
              @trigger-file="triggerFileInput"
              @remove-file="removeFile"
              @pick-location="pickLocation"
              @sig-start="sigStart"
              @sig-move="sigMove"
              @sig-end="sigEnd"
              @sig-touch-start="sigTouchStart"
              @sig-touch-move="sigTouchMove"
              @clear-signature="clearSignature"
              @add-matrix-row="addMatrixAutoRow"
              @remove-matrix-row="removeMatrixAutoRow"
            />
            <div v-if="settings.answerVisible && !LAYOUT_TYPES.includes(currentQuestion.type)" class="exam-answer-box">
              <div v-if="currentQuestion.examCorrectAnswer" style="margin-bottom:4px">
                <span style="color:#67c23a">✓ 正确答案：</span><span v-html="displayAnswer(currentQuestion, currentQuestion.examCorrectAnswer)" />
              </div>
              <div v-if="currentQuestion.examAnalysis" style="color:#909399;white-space:pre-wrap">
                <span style="font-weight:500">解析：</span>{{ currentQuestion.examAnalysis }}
              </div>
            </div>
          </div>
          <div class="exam-nav">
            <el-button :disabled="currentNavIndex <= 0" @click="goPrev">上一题</el-button>
            <span class="exam-nav-index">{{ currentNavIndex + 1 }} / {{ navQuestions.length }}</span>
            <el-button v-if="!isLast" type="primary" class="exam-btn--primary" @click="goNext">下一题</el-button>
            <el-button v-else type="primary" class="exam-btn--primary" size="large" :loading="submitting" @click="onSubmit">交卷</el-button>
          </div>
        </template>

        <template v-else>
          <div v-for="(q, i) in questions.filter(q => !q.defaultHidden)" :key="q.id" class="q-wrap">
            <QuestionField
              :q="q"
              :index="questions.filter(q => !q.defaultHidden).slice(0, i).filter(x => !LAYOUT_TYPES.includes(x.type)).length"
              :score="exam.showScore && q.examScore ? q.examScore : 0"
              :answers="answers"
              :settings="settings"
              :file-lists="fileLists"
              :file-inputs="fileInputs"
              :sig-canvas-map="sigCanvasMap"
              @update:answers="answers = $event"
              @trigger-file="triggerFileInput"
              @remove-file="removeFile"
              @pick-location="pickLocation"
              @sig-start="sigStart"
              @sig-move="sigMove"
              @sig-end="sigEnd"
              @sig-touch-start="sigTouchStart"
              @sig-touch-move="sigTouchMove"
              @clear-signature="clearSignature"
              @add-matrix-row="addMatrixAutoRow"
              @remove-matrix-row="removeMatrixAutoRow"
            />
          </div>
          <div class="exam-footer">
            <el-button type="primary" class="exam-btn--primary" size="large" :loading="submitting" @click="onSubmit">交卷</el-button>
          </div>
        </template>
      </el-form>
    </div>

    <!-- 答题卡 -->
    <div v-if="settings.answerSheetVisible !== false && !loading && exam && sheetVisible" class="exam-sheet" :style="sheetStyle">
      <div class="exam-sheet-header" @mousedown.prevent="onSheetDragStart">
        <span>≡ 答题卡</span>
        <span class="exam-sheet-close" @click.stop="sheetVisible = false">✕</span>
      </div>
      <div class="exam-sheet-grid">
        <div
          v-for="(q, i) in realQuestions" :key="q.id"
          class="exam-sheet-item"
          :class="{
            'exam-sheet-item--done': isAnswered(q, answers[q.id]),
            'exam-sheet-item--cur': settings.onePageOneQuestion && q === currentQuestion
          }"
          @click="jumpToQuestion(q)"
        >{{ i + 1 }}</div>
      </div>
      <div class="exam-sheet-stat">{{ answeredCount }}/{{ totalQuestions }}</div>
    </div>
    <div v-if="settings.answerSheetVisible !== false && !loading && exam && !sheetVisible" class="exam-sheet-toggle" @click="sheetVisible = true">≡ {{ answeredCount }}/{{ totalQuestions }}</div>

    <el-dialog v-if="showScanner" v-model="showScanner" title="扫码" width="400px" :close-on-click-modal="false" destroy-on-close @opened="onScannerOpen" @close="onScannerClose">
      <div ref="scannerRef" style="width:100%;aspect-ratio:1;overflow:hidden;background:#000;border-radius:8px" />
      <template #footer><el-button @click="showScanner = false">取消</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Html5Qrcode } from 'html5-qrcode'
import { QuillEditor } from '@vueup/vue-quill'
import '@vueup/vue-quill/dist/vue-quill.snow.css'
import QuestionField from '../survey/components/QuestionField.vue'
import { publicFormApi } from '../../api/public'

const route = useRoute()
const exam = ref<any>(null)
const settings = ref<any>({})
const questions = ref<any[]>([])
const answers = ref<any>({})
const loading = ref(true)
const error = ref('')
const submitting = ref(false)
const session = ref('')
const submitted = ref(false)
const endContent = ref('')
const showLogin = ref(false)
const loginLoading = ref(false)
const loginForm = reactive({ name: '', password: '' })
const showScanner = ref(false)
const scannerRef = ref<HTMLDivElement>()
const formRef = ref<any>()
let scanner: Html5Qrcode | null = null

const result = ref<any>(null)
const showResultView = ref(false)
const countdownToStart = ref(0)
let startTimer: ReturnType<typeof setInterval> | null = null

function openRedirectUrl() {
  const url = settings.value?.redirectUrl
  if (url) window.location.href = url
}

function onScannerOpen() {
  // PC1 模板保留扫码弹窗位；当前未启用扫码入口时保持安全 no-op。
}

function onScannerClose() {
  if (scanner) {
    scanner.stop().catch(() => {})
    scanner = null
  }
}

function getDeviceId() {
  const key = '_device_id'
  let id = localStorage.getItem(key)
  if (!id) {
    try { id = crypto.randomUUID() } catch {}
    if (!id) id = 'd_' + Date.now().toString(36) + '_' + Math.random().toString(36).slice(2, 10)
    localStorage.setItem(key, id)
  }
  return id
}

function unwrapOuterP(html: string): string {
  const trimmed = (html || '').trim()
  const match = trimmed.match(/^<p([^>]*)>([\s\S]*)<\/p>$/)
  if (!match) return html || ''
  const attrs = match[1]
  const content = match[2]
  if (/\bql-align-\w+\b/.test(attrs)) {
    return `<span${attrs} style="display:inline-block;width:100%;text-align:inherit">${content}</span>`
  }
  return content
}

function resultTitleHtml(q: any, i: number): string {
  return `<span class="q-num">${i + 1}.</span>${unwrapOuterP(q.title || '')}`
}

const fileLists: Record<string, File[]> = reactive({})
const fileInputs: Record<string, HTMLInputElement> = {}
const sigCanvasMap: Record<string, HTMLCanvasElement> = {}
const sigDrawing = ref(false)
const sigCurId = ref('')
const startAt = ref(0)
const remaining = ref(0)
const currentIndex = ref(0)
const sheetVisible = ref(true)
let countdownTimer: ReturnType<typeof setInterval> | null = null
let draftTimer: ReturnType<typeof setTimeout> | null = null
let userDeptId = 0

const examTimeRange = computed(() => {
  if (!exam.value) return ''
  const st = Number(exam.value.startTime)
  const et = Number(exam.value.endTime)
  if (!st && !et) return ''
  const fmt = (t: number) => {
    const d = new Date(t)
    return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')} ${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`
  }
  if (st && et) return `考试时间：${fmt(st)} ~ ${fmt(et)}`
  if (st) return `开始时间：${fmt(st)}`
  return `结束时间：${fmt(et)}`
})

const bgStyle = computed(() => {
  const s: Record<string, string> = { backgroundColor: 'rgb(28, 144, 153)' }
  const imgs = settings.value.backgroundImages
  if (imgs?.length) {
    const img = typeof imgs[0] === 'string' ? imgs[0] : imgs[0]?.url
    if (img) {
      s.backgroundImage = `url(${img})`
      s.backgroundSize = 'cover'
      s.backgroundPosition = 'center'
    }
  }
  return s
})

const LAYOUT_TYPES = ['description', 'divider', 'pagination']
const realQuestions = computed(() => questions.value.filter((q: any) => !LAYOUT_TYPES.includes(q.type) && !q.defaultHidden))
const totalQuestions = computed(() => realQuestions.value.length)
const totalScore = computed(() => questions.value.reduce((sum: number, q: any) => sum + (Number(q.examScore) || 0), 0))
const answeredCount = computed(() => realQuestions.value.filter((q: any) => isAnswered(q, answers.value[q.id])).length)
const progressPct = computed(() => totalQuestions.value ? Math.round(answeredCount.value / totalQuestions.value * 100) : 0)
const navQuestions = computed(() => realQuestions.value)
const currentNavIndex = computed(() => { const q = questions.value[currentIndex.value]; return q ? navQuestions.value.indexOf(q) : -1 })
const isLast = computed(() => currentIndex.value >= questions.value.length - 1)
const currentQuestion = computed(() => questions.value[currentIndex.value] || null)

const AUTO_SAVE_KEY = 'exam_draft_'
function getDraftKey() { return AUTO_SAVE_KEY + route.params.id }
function saveDraft() {
  if (!settings.value.autoSave) return
  try { localStorage.setItem(getDraftKey(), JSON.stringify({ answers: answers.value, updatedAt: Date.now() })) } catch {}
}
function loadDraft() {
  if (!settings.value.autoSave) return null
  try { const raw = localStorage.getItem(getDraftKey()); return raw ? JSON.parse(raw) : null } catch { return null }
}
function clearDraft() { localStorage.removeItem(getDraftKey()) }
function scheduleDraft() { if (draftTimer) clearTimeout(draftTimer); draftTimer = setTimeout(saveDraft, 2000) }

function goNext() {
  if (currentIndex.value < questions.value.length - 1) {
    currentIndex.value++
    while (currentIndex.value < questions.value.length - 1 && questions.value[currentIndex.value]?.defaultHidden) currentIndex.value++
  }
}
function goPrev() {
  if (currentIndex.value > 0) {
    currentIndex.value--
    while (currentIndex.value > 0 && questions.value[currentIndex.value]?.defaultHidden) currentIndex.value--
  }
}

const swipeStartX = ref(0)
const swipeStartY = ref(0)
function onSwipeStart(e: TouchEvent) { swipeStartX.value = e.touches[0].clientX; swipeStartY.value = e.touches[0].clientY }
function onSwipeEnd(e: TouchEvent) {
  if (sigDrawing.value) return
  const dx = e.changedTouches[0].clientX - swipeStartX.value
  const dy = e.changedTouches[0].clientY - swipeStartY.value
  if (Math.abs(dx) > 30 && Math.abs(dx) > Math.abs(dy) * 1.5) {
    dx > 0 ? goPrev() : goNext()
  }
}

function jumpToQuestion(q: any) {
  const idx = questions.value.indexOf(q)
  if (idx < 0) return
  if (settings.value.onePageOneQuestion) { currentIndex.value = idx }
  else { const el = document.querySelector(`[data-qid="${q.id}"]`); if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' }) }
}

const sheetX = ref(0)
const sheetY = ref(0)
const sheetDragging = ref(false)
const sheetDragStartX = ref(0)
const sheetDragStartY = ref(0)
const sheetOrigX = ref(0)
const sheetOrigY = ref(0)
const sheetStyle = computed(() => {
  const s: Record<string, string> = {}
  if (sheetX.value || sheetY.value) s.transform = `translate(calc(-50% + ${sheetX.value}px), calc(-50% + ${sheetY.value}px))`
  if (sheetDragging.value) s.cursor = 'grabbing'
  return s
})
function onSheetDragStart(e: MouseEvent) {
  sheetDragging.value = true
  sheetDragStartX.value = e.clientX
  sheetDragStartY.value = e.clientY
  sheetOrigX.value = sheetX.value
  sheetOrigY.value = sheetY.value
  document.addEventListener('mousemove', onSheetDragMove)
  document.addEventListener('mouseup', onSheetDragEnd)
}
function onSheetDragMove(e: MouseEvent) {
  if (!sheetDragging.value) return
  sheetX.value = sheetOrigX.value + e.clientX - sheetDragStartX.value
  sheetY.value = sheetOrigY.value + e.clientY - sheetDragStartY.value
}
function onSheetDragEnd() {
  sheetDragging.value = false
  document.removeEventListener('mousemove', onSheetDragMove)
  document.removeEventListener('mouseup', onSheetDragEnd)
}

function isAnswered(q: any, val: any): boolean {
  const type = q.type
  if (val === undefined || val === null) return false
  if (['checkbox', 'file'].includes(type)) return Array.isArray(val) && val.length > 0
  if (['rating', 'nps'].includes(type)) return val > 0
  if (['multiInput', 'hInput'].includes(type)) return Array.isArray(val) && val.some((v: any) => !!v)
  if (type === 'matrixRadio') return Object.keys(val).length > 0
  if (type === 'matrixCheckbox') return Object.values(val).some((v: any) => Array.isArray(v) && v.length > 0)
  if (type === 'matrixFillBlank') return Object.values(val).some((v: any) => !!v)
  if (type === 'matrixAuto') return Array.isArray(val) && val.some((row: any) => row.some((v: any) => !!v))
  if (type === 'dateRange') return Array.isArray(val) && val.some((v: any) => !!v)
  if (type === 'switch') return val === true
  if (type === 'cascade') return Array.isArray(val) && val.length > 0
  if (['user', 'dept'].includes(type)) return q.multiple ? (Array.isArray(val) && val.length > 0) : !!val
  if (type === 'picker') return !!val
  return !!val
}

function getInitVal(q: any): any {
  const type = q.type
  if (type === 'checkbox') return []
  if (type === 'switch') return false
  if (type === 'number') return undefined
  if (type === 'rating') return 0
  if (type === 'nps') return 0
  if (type === 'dateRange') return ['', '']
  if (['matrixRadio', 'matrixCheckbox', 'matrixFillBlank'].includes(type)) return {}
  if (type === 'matrixAuto') return []
  if (['multiInput', 'hInput'].includes(type)) return Array((q.props?.fields || []).length).fill('')
  if (['user', 'dept'].includes(type)) return q.multiple ? [] : ''
  if (['cascade'].includes(type)) return []
  if (type === 'picker') return ''
  if (type === 'file') return []
  return ''
}

function displayAnswer(q: any, val: string): string {
  if (q.type === 'judge') return val === 'true' ? '对' : '错'
  const opts = q.props?.options || []
  const opt = opts.find((o: any) => String(o.value) === String(val))
  return opt?.label || val
}

function formatResultAnswer(q: any, val: any): string {
  if (val === undefined || val === null) return '-'
  if (typeof val === 'string') return val
  if (Array.isArray(val)) return val.join(', ')
  return String(val)
}

function buildUserDeptTree(q: any) {
  const opts = q.props?.options || []
  if (q.type === 'user') {
    const deptMap: Record<string, any> = {}
    opts.forEach((o: any) => {
      const deptId = o.deptId || ''
      if (!deptMap[deptId]) deptMap[deptId] = { value: '__d__' + deptId, label: o.deptName || deptId || '未分组', children: [] }
      deptMap[deptId].children.push({ value: o.value, label: o.label || '成员' })
    })
    return Object.values(deptMap)
  }
  const map: Record<string, any> = {}
  opts.forEach((o: any) => { map[o.value] = { ...o, children: [] } })
  const roots: any[] = []
  opts.forEach((o: any) => {
    if (o.parentId && map[o.parentId]) map[o.parentId].children.push(map[o.value])
    else roots.push(map[o.value])
  })
  return roots
}

function addMatrixAutoRow(qid: string) {
  const q = questions.value.find((x: any) => x.id === qid)
  if (!q) return
  const cols = q.props?.columns?.length || 0
  if (!answers.value[qid]) answers.value[qid] = []
  answers.value[qid].push(Array(cols).fill(''))
}
function removeMatrixAutoRow(qid: string, ri: number) { if (answers.value[qid]) answers.value[qid].splice(ri, 1) }

function triggerFileInput(qid: string) { fileInputs[qid]?.click() }
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
  return { x: (e.clientX - rect.left) * (c.width / rect.width), y: (e.clientY - rect.top) * (c.height / rect.height) }
}
function sigStart(e: MouseEvent | Touch, id: string) {
  const c = sigCanvasMap[id]; if (!c) return
  const p = sigPos(e, c); sigCurId.value = id; sigDrawing.value = true
  const ctx = c.getContext('2d'); if (ctx) { ctx.beginPath(); ctx.moveTo(p.x, p.y) }
}
function sigMove(e: MouseEvent | Touch, id: string) {
  if (!sigDrawing.value || sigCurId.value !== id) return
  const c = sigCanvasMap[id]; if (!c) return
  const p = sigPos(e, c); const ctx = c.getContext('2d'); if (ctx) { ctx.lineTo(p.x, p.y); ctx.stroke() }
}
function sigEnd() {
  sigDrawing.value = false
  const id = sigCurId.value; sigCurId.value = ''
  if (!id) return; const c = sigCanvasMap[id]; if (!c) return
  answers.value[id] = c.toDataURL()
}
function sigTouchStart(e: TouchEvent, id: string) { if (e.touches[0]) sigStart(e.touches[0], id) }
function sigTouchMove(e: TouchEvent, id: string) { if (e.touches[0]) sigMove(e.touches[0], id) }
function clearSignature(id: string) {
  const c = sigCanvasMap[id]; if (!c) return
  const ctx = c.getContext('2d'); if (ctx) ctx.clearRect(0, 0, c.width, c.height)
  answers.value[id] = ''
}

async function doLogin() {
  loginLoading.value = true
  try {
    const res = await publicFormApi.login(loginForm.name, loginForm.password)
    if (res.code === 0) {
      localStorage.setItem('user_token', res.data.token)
      userDeptId = res.data.userInfo?.deptId || 0
      showLogin.value = false; loading.value = true; load()
    } else { ElMessage.error(res.msg || '登录失败') }
  } catch { ElMessage.error('登录失败') }
  finally { loginLoading.value = false }
}

async function load() {
  const id = Array.isArray(route.params.id) ? route.params.id[0] : route.params.id
  if (!id) { error.value = '参数错误'; loading.value = false; return }

  // 检查是否为结果模式
  const resultSession = route.query.session as string
  if (resultSession) {
    try {
      const res = await publicFormApi.examResult(resultSession)
      if (res.code === 0 && res.data) {
        exam.value = res.data.exam
        questions.value = res.data.questions || []
        answers.value = res.data.answers || {}
        settings.value = res.data.settings || {}
        if (res.data.record) {
          const d = res.data
          result.value = {
            score: d.record?.score,
            fullScore: d.record?.totalScore,
            correctCnt: d.results?.filter((r: any) => r.correct).length,
            results: d.results
          }
        }
        submitted.value = true
        showResultView.value = true
        loading.value = false
        return
      }
    } catch {}
  }

  session.value = localStorage.getItem('exam_session_' + id) || ''
  try {
    const res = await publicFormApi.examDetail(id, session.value)
    if (res.code !== 0) {
      if (res.msg === '请先登录') { showLogin.value = true; loading.value = false; return }
      if (res.msg === '考试未开始' && res.data) {
        const startMs = Number(res.data.startTime)
        if (startMs > 0) {
          error.value = '考试未开始'
          const tick = () => {
            const left = startMs - Date.now()
            if (left <= 0) { if (startTimer) clearInterval(startTimer); window.location.reload() }
            countdownToStart.value = Math.max(0, left)
          }
          tick()
          startTimer = setInterval(tick, 1000)
          loading.value = false
          return
        }
      }
      error.value = res.msg || '加载失败'; loading.value = false; return
    }
    exam.value = res.data
    if (exam.value.endTime > 0 && Date.now() > exam.value.endTime) {
      error.value = '考试已结束'; loading.value = false; return
    }
    if (res.data?.settings && typeof res.data.settings === 'string') {
      try { settings.value = JSON.parse(res.data.settings) } catch { settings.value = {} }
    } else { settings.value = res.data?.settings || {} }
    if (settings.value.answerVisible === undefined) settings.value.answerVisible = true
    if (res.data?.session) { session.value = res.data.session; localStorage.setItem('exam_session_' + id, session.value) }
    if (res.data?.startAt) startAt.value = res.data.startAt

    if (settings.value.loginRequired || Number(exam.value?.visibility) === 1 || Number(exam.value?.visibility) === 2) {
      const token = localStorage.getItem('user_token')
      if (!token) { showLogin.value = true; loading.value = false; return }
    }

    const raw = res.data?.schema
    const sch = raw ? (typeof raw === 'string' ? JSON.parse(raw) : raw) : { questions: [] }
    questions.value = sch.questions || []

    // 随机排序
    if (settings.value.randomOrder) {
      const layout = questions.value.filter((q: any) => LAYOUT_TYPES.includes(q.type))
      const nonLayout = questions.value.filter((q: any) => !LAYOUT_TYPES.includes(q.type))
      for (let i = nonLayout.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));
        [nonLayout[i], nonLayout[j]] = [nonLayout[j], nonLayout[i]]
      }
      questions.value = [...layout, ...nonLayout]
    }

    const init: any = {}
    questions.value.forEach((q: any) => { init[q.id] = getInitVal(q) })
    answers.value = init
    const draft = loadDraft()
    if (draft && draft.answers) {
      for (const key of Object.keys(draft.answers)) { if (key in answers.value) answers.value[key] = draft.answers[key] }
    }
    await nextTick()
    questions.value.filter((q: any) => q.type === 'signature').forEach((q: any) => initSigCanvas(q.id))
    while (currentIndex.value < questions.value.length - 1 && questions.value[currentIndex.value]?.defaultHidden) currentIndex.value++
  } catch { error.value = '加载失败' }
  finally { loading.value = false }
  startCountdown()
}

function formatRemaining(ms: number) {
  if (ms <= 0) return '已超时'
  const t = Math.ceil(ms / 1000)
  return `${Math.floor(t / 60)}:${(t % 60).toString().padStart(2, '0')}`
}
function startCountdown() {
  stopCountdown()
  if (!startAt.value) return
  const limit = Number(settings.value.timeLimit)
  const maxSubmit = Number(settings.value.maxSubmitMinutes)
  const endTime = Number(exam.value?.endTime)
  const tick = () => {
    const now = Date.now()
    let deadline = Infinity
    if (limit > 0) deadline = Math.min(deadline, startAt.value + limit * 60 * 1000)
    if (maxSubmit > 0) deadline = Math.min(deadline, startAt.value + maxSubmit * 60 * 1000)
    if (endTime > 0) deadline = Math.min(deadline, endTime)
    if (deadline === Infinity) { remaining.value = 0; return }
    const left = deadline - now
    remaining.value = Math.max(0, left)
    if (left <= 0) { stopCountdown(); ElMessage.warning('作答时间已到，自动交卷'); onSubmit(true) }
  }
  tick(); countdownTimer = setInterval(tick, 1000)
}
function stopCountdown() { if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null } }

function handleSubmitSuccess(resData: any) {
  if (resData?.record) {
    result.value = {
      score: resData.record?.score,
      fullScore: resData.record?.totalScore,
      correctCnt: resData.results?.filter((r: any) => r.correct).length,
      results: resData.results
    }
  }
  submitted.value = true
  if (settings.value.transcriptVisible !== false && !settings.value.showAnalysis) {
    showResultView.value = true
  }
  if (!settings.value.transcriptVisible && !settings.value.showAnalysis) {
    const url = settings.value?.redirectUrl
    const content = settings.value?.endContent
    if (url) { window.location.href = url; return }
    if (content) { endContent.value = content; return }
    ElMessage.success('已交卷')
  }
}

async function onSubmit(skipConfirm = false) {
  if (typeof skipConfirm !== 'boolean') skipConfirm = false
  const id = Number(route.params.id)
  if (!skipConfirm) {
    // 最短交卷时间检查
    const minSubmit = Number(settings.value.minSubmitMinutes)
    if (minSubmit > 0 && startAt.value > 0) {
      const elapsed = (Date.now() - startAt.value) / 60000
      if (elapsed < minSubmit) {
        ElMessage.warning(`距最短交卷时间还有 ${Math.ceil(minSubmit - elapsed)} 分钟`)
        return
      }
    }
    try {
      const vr = await publicFormApi.examValidate(id, { examId: id, answers: answers.value, device: navigator.userAgent, deviceId: getDeviceId() })
      if (vr.data && !vr.data.valid) { ElMessage.warning((vr.data.errors || []).map((e: any) => e.message).join('; ') || '请检查填写内容'); return }
    } catch {}
  }
  if (skipConfirm) {
    submitting.value = true
    try {
      const res = await publicFormApi.examSubmit(id, { examId: id, answers: answers.value, device: navigator.userAgent, session: session.value, autoSubmit: true, deviceId: getDeviceId() })
      if (res.code !== 0) { ElMessage.error(res.msg || '交卷失败') }
      else { stopCountdown(); clearDraft(); localStorage.removeItem('exam_session_' + id); handleSubmitSuccess(res.data) }
    } catch (e: any) { ElMessage.error(e.msg || '交卷失败') }
    finally { submitting.value = false }
    return
  }
  ElMessageBox.confirm('确认交卷？交卷后不可修改', '提示', { type: 'info' }).then(async () => {
    submitting.value = true
    try {
      const res = await publicFormApi.examSubmit(id, { examId: id, answers: answers.value, device: navigator.userAgent, session: session.value, deviceId: getDeviceId() })
      if (res.code !== 0) { ElMessage.error(res.msg || '交卷失败') }
      else { stopCountdown(); clearDraft(); localStorage.removeItem('exam_session_' + id); handleSubmitSuccess(res.data) }
    } catch (e: any) { ElMessage.error(e.msg || '交卷失败') }
    finally { submitting.value = false }
  }).catch(() => {})
}

onMounted(load)
watch(answers, () => { scheduleDraft() }, { deep: true })
watch(() => settings.value.autoSave, (v) => { if (!v) clearDraft() })
watch(() => settings.value.timeLimit, () => { startCountdown() })
onUnmounted(() => { if (draftTimer) clearTimeout(draftTimer); stopCountdown(); if (startTimer) clearInterval(startTimer) })
</script>

<style scoped>
.exam {
  min-height: 100vh;
  background-color: rgb(28, 144, 153);
  padding: 48px 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background-repeat: no-repeat;
  background-position: top;
  background-size: cover;
  background-attachment: fixed;
}
.exam:has(.exam-progress) { padding-top: 72px; }

.exam-progress {
  position: fixed; top: 0; left: 0; right: 0; z-index: 1000;
  background: #fff;
  border-bottom: 1px solid #e8e8e8;
  padding: 8px 24px;
  animation: fadeInDown .3s ease;
}
.exam-progress-inner {
  max-width: 720px; margin: 0 auto;
  display: flex; align-items: center; gap: 12px;
}
.exam-progress-track { flex: 1; height: 6px; background: #e8e8e8; border-radius: 3px; overflow: hidden; }
.exam-progress-fill { height: 100%; background: #3873f6; border-radius: 3px; transition: width .4s ease; }
.exam-progress-label { font-size: 13px; color: #909399; white-space: nowrap; }

.exam-card {
  max-width: 720px;
  margin: 0 auto;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,.08), 0 1px 2px rgba(0,0,0,.04);
  overflow: hidden;
}
.exam-card--narrow { width: 98%; }
.exam-card-img { width: 100%; overflow: hidden; line-height: 0; }
.exam-card-img img { width: 100%; max-height: 200px; object-fit: cover; display: block; }

.exam-header { padding: 32px 36px 20px; border-bottom: 1px solid #f0f0f0; }
.exam-title { font-size: 22px; font-weight: 600; color: #1a1a1a; margin: 0 0 6px; line-height: 1.4; }
.exam-desc { font-size: 14px; color: #666; line-height: 1.6; margin: 0 0 14px; }
.exam-meta { display: flex; gap: 8px; flex-wrap: wrap; align-items: center; }
.exam-tag {
  display: inline-block; padding: 1px 8px; border-radius: 3px;
  font-size: 12px; line-height: 22px; background: #f0f2f5; color: #606266;
}
.exam-tag--warn { background: #fff7e6; color: #fa8c16; }
.exam-tag--danger { background: #fff1f0; color: #f5222d; }

.exam-form { width: 100%; }
.exam-form-scroll { padding: 28px 36px; min-height: 600px; }

.exam-nav {
  display: flex; align-items: center; justify-content: center; gap: 16px;
  padding: 16px 36px; border-top: 1px solid #f0f0f0;
}
.exam-nav-index { font-size: 13px; color: #909399; }

.exam-footer { text-align: center; padding: 24px 0 32px; }

.exam-btn--primary { background: #3873f6; border-color: #3873f6; }
.exam-btn--primary:hover,
.exam-btn--primary:focus { background: #2a5fd9; border-color: #2a5fd9; }

.exam-login { max-width: 720px; margin: 0 auto; }
.exam-login-header { text-align: center; margin-bottom: 28px; }
.exam-login-title { font-size: 20px; color: #333; font-weight: 600; margin: 0 0 8px; }
.exam-login-desc { font-size: 14px; color: #909399; margin: 0; }
.exam-login .el-form { width: 98%; margin: 0 auto; }
.exam-login-footer { text-align: center; margin-top: 8px; font-size: 13px; color: #909399; }
.exam-login-link { color: #3873f6; text-decoration: none; margin-left: 4px; }
.exam-login-link:hover { text-decoration: underline; }

.exam-error { max-width: 720px; margin: 0 auto; padding: 100px 0; text-align: center; color: #909399; font-size: 16px; }

.exam-loading { display: flex; align-items: center; justify-content: center; padding: 120px 0; }
.spinner { width: 50px; height: 40px; font-size: 10px; text-align: center; }
.spinner > div {
  display: inline-block; width: 6px; height: 100%;
  background: #3873f6;
  animation: sk-stretchdelay 1.2s infinite ease-in-out;
}
.spinner .rect2 { animation-delay: -1.1s; }
.spinner .rect3 { animation-delay: -1s; }
.spinner .rect4 { animation-delay: -.9s; }
.spinner .rect5 { animation-delay: -.8s; }
@keyframes sk-stretchdelay {
  0%, 40%, 100% { transform: scaleY(.4); }
  20% { transform: scaleY(1); }
}
@keyframes fadeInDown {
  from { opacity: 0; transform: translateY(-8px); }
  to { opacity: 1; transform: translateY(0); }
}

.exam-end {
  position: absolute; top: 50%; left: 50%; z-index: 300;
  width: 100%; max-width: 650px;
  transform: translate(-50%, -50%);
  text-align: center;
  animation: fadeInUp .4s ease;
}
.exam-end-content {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,.08);
  padding: 60px 40px;
}
.exam-end-title {
  font-size: 22px; font-weight: 400; color: #333;
  margin: 0 0 16px;
}
.exam-end-desc { font-size: 14px; color: #666; white-space: pre-wrap; word-break: break-word; }
.exam-end-result { margin-top: 24px; }
.exam-end-result-copy { margin-top: 8px; }
@keyframes fadeInUp {
  from { opacity: 0; transform: translate(-50%, -30%); }
  to { opacity: 1; transform: translate(-50%, -50%); }
}

.exam-answer-box {
  margin: 12px 0 0 0;
  font-size: 13px;
  border: 1px dashed #ccc;
  border-radius: 4px;
  padding: 10px 12px;
  background: #fafafa;
}

/* 答题卡 */
.exam-sheet {
  position: fixed; right: 24px; top: 50%; z-index: 900;
  width: 100px;
  background: #fff; border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0,0,0,.1);
  padding: 10px 10px 8px;
  display: flex; flex-direction: column;
  user-select: none;
  transform: translateY(-50%);
}
.exam-sheet-header {
  font-size: 13px; font-weight: 600; color: #333;
  text-align: center; margin-bottom: 8px; padding-bottom: 8px;
  border-bottom: 1px solid #f0f0f0;
  cursor: grab;
  display: flex; align-items: center; justify-content: center; gap: 8px;
}
.exam-sheet-close { display: none; cursor: pointer; color: #909399; font-size: 14px; line-height: 1; }
.exam-sheet-close:hover { color: #333; }
.exam-sheet-toggle {
  display: none;
  position: fixed; right: 0; top: 50%; z-index: 900;
  transform: translateY(-50%);
  background: #3873f6; color: #fff; font-size: 12px;
  padding: 8px 6px; border-radius: 6px 0 0 6px;
  cursor: pointer; line-height: 1.4; text-align: center;
  box-shadow: 0 2px 8px rgba(0,0,0,.15);
}
.exam-sheet-grid {
  display: grid; grid-template-columns: repeat(4, 1fr);
  gap: 4px; overflow-y: auto; flex: 1;
}
.exam-sheet-item {
  display: flex; align-items: center; justify-content: center;
  height: 26px; border-radius: 4px; cursor: pointer;
  font-size: 12px; color: #606266; background: #f5f6f8;
  transition: all .15s;
}
.exam-sheet-item:hover { background: #e0e2e6; }
.exam-sheet-item--done { background: #eef2ff; color: #3873f6; }
.exam-sheet-item--cur { border: 2px solid #3873f6; font-weight: 600; }
.exam-sheet-stat {
  text-align: center; font-size: 12px; color: #909399;
  padding-top: 6px; margin-top: 6px; border-top: 1px solid #f0f0f0;
}

/* 成绩单 */
.result-container { padding: 24px 0; }
.result-container .q-item {
  padding: 16px 36px;
  border-bottom: 1px solid #f0f0f0;
}
.result-container .q-title { font-size: 15px; color: #303133; font-weight: 500; margin-bottom: 6px; white-space: pre-wrap; }
.result-container .q-title :deep(.q-num) { color: #3873f6; font-weight: 600; margin-right: 4px; }
.result-container .q-title :deep(img) { max-width: 100%; height: auto; }
.result-container :deep(.ql-align-center) { text-align: center; }
.result-container :deep(.ql-align-right) { text-align: right; }
.result-container :deep(.ql-align-justify) { text-align: justify; }
.result-container :deep(.ql-code-block-container) { background: #f5f5f5; border-radius: 4px; padding: 10px 14px; margin: 6px 0; font-family: monospace; font-size: 13px; line-height: 1.5; overflow-x: auto; }
.result-container :deep(.ql-code-block) { white-space: pre; }
.result-container .q-answer { font-size: 13px; color: #606266; }

/* H5 响应式 */
@media (max-width: 768px) {
  .exam.pc { padding: 0; background: #f5f6f8; }
  .exam-card { border-radius: 0; box-shadow: none; min-height: 100vh; max-width: 100%; }
  .exam-header { padding: 20px 16px 16px; }
  .exam-title { font-size: 18px; }
  .exam-form-scroll { padding: 8px 0; min-height: 280px; }
  .exam-form { max-width: 100%; }
  .exam-form :deep(.q-item) { padding-left: 12px; padding-right: 12px; }
  .exam-form :deep(.q-item--layout) { padding-left: 12px; padding-right: 12px; }
  .exam-form :deep(.el-form-item) { width: 100%; max-width: 100%; }
  .exam-form :deep(.el-form-item__content) { width: 100%; max-width: 100%; }
  .exam-form :deep(.el-input), .exam-form :deep(.el-select),
  .exam-form :deep(.el-textarea), .exam-form :deep(.el-cascader),
  .exam-form :deep(.el-rate), .exam-form :deep(.el-slider) { width: 100% !important; max-width: 100% !important; }
  .exam-nav { padding: 12px 16px; gap: 8px; flex-wrap: wrap; }
  .exam-sheet { right: 8px; width: 80px; }
  .exam-sheet-close { display: block; }
  .exam-sheet-toggle { display: block; }
  .exam-end { max-width: 90%; }
  .exam-end-content { padding: 40px 20px; }
  .exam-login { max-width: 100%; padding: 0; margin-top: 0; display: flex; align-items: flex-start; }
  .exam-login .exam-card--narrow { width: 100%; }
  .exam-login-header { margin-bottom: 12px; padding-top: 8px; }
  .exam-login-title { font-size: 17px; }
  .exam-login .el-form-item { margin-bottom: 12px; }
  .exam-login-footer { margin-top: 0; }
  .result-container .q-item { padding: 12px 16px; }
}
</style>
