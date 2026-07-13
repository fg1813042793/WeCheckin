<template>
  <view>
    <view class="q-item">
    <template v-if="LAYOUT_TYPES.includes(q.type)">
      <view v-if="q.type==='description'" class="q-desc" v-html="q.description" />
      <view v-else-if="q.type==='divider'" class="q-divider" />
    </template>
    <template v-else>
      <view class="q-title">
        <rich-text :nodes="processedFullTitle" />
      </view>
      <view v-if="q.description && q.showDescription !== false" class="q-desc">
        <text class="q-desc-label">说明：</text>
        <rich-text :nodes="processedDesc" />
      </view>
      <view class="q-input-area">
        <!-- input / text -->
        <input v-if="['input','text'].includes(q.type)" class="q-input" :value="localVal" @input="onInput" :placeholder="q.placeholder || '请输入'" />

        <!-- multiInput -->
        <view v-else-if="q.type==='multiInput'" class="q-field-stack">
          <input v-for="(f, fi) in (q.props?.fields||[])" :key="fi" class="q-input" :value="localVal?.[fi]||''" @input="(e) => onArrayInput(fi, e)" :placeholder="f.placeholder||'请输入'" />
        </view>

        <!-- hInput -->
        <view v-else-if="q.type==='hInput'" class="q-field-row">
          <input v-for="(f, fi) in (q.props?.fields||[])" :key="fi" class="q-input" :value="localVal?.[fi]||''" @input="(e) => onArrayInput(fi, e)" :placeholder="f.placeholder||'请输入'" />
        </view>

        <!-- textarea -->
        <textarea v-else-if="q.type==='textarea'" class="q-textarea" :value="localVal" @input="onInput" :placeholder="q.placeholder || '请作答'" />

        <!-- number -->
        <input v-else-if="q.type==='number'" class="q-input" :value="localVal" @input="onInput" type="number" :placeholder="q.placeholder || '请输入'" />

        <!-- radio -->
        <view v-else-if="q.type==='radio'" class="opt-list">
          <view v-for="o in parseOpts(q)" :key="o.value" class="opt" :class="{ active: localVal === o.value }" @click="setVal(o.value)">
            <view class="opt-radio"><view v-if="localVal === o.value" class="opt-radio-dot" /></view>
            <rich-text :nodes="o.label" />
          </view>
        </view>

        <!-- checkbox -->
        <view v-else-if="q.type==='checkbox'" class="opt-list">
          <view v-for="o in parseOpts(q)" :key="o.value" class="opt" :class="{ active: isChk(o.value) }" @click="toggleChk(o.value)">
            <view class="opt-checkbox"><text v-if="isChk(o.value)" class="opt-checkbox-icon">✓</text></view>
            <rich-text :nodes="o.label" />
          </view>
        </view>

        <!-- select / picker -->
        <view v-else-if="['select','picker'].includes(q.type)">
          <picker mode="selector" :range="parseOpts(q).map(o => stripHtml(o.label))" @change="(e) => onSelect(parseOpts(q), e)">
            <view class="q-input">{{ getOptLabel(parseOpts(q), localVal) || '请选择' }}</view>
          </picker>
        </view>

        <!-- cascade -->
        <view v-else-if="q.type==='cascade'">
          <picker mode="multiSelector" :range="cascadeRange(q)" @change="(e) => onCascade(q, e)">
            <view class="q-input">{{ Array.isArray(localVal) ? localVal.join(' / ') : '请选择' }}</view>
          </picker>
        </view>

        <!-- judge -->
        <view v-else-if="q.type==='judge'" class="opt-list">
          <view class="opt" :class="{ active: localVal === 'true' }" @click="setVal('true')">
            <view class="opt-radio"><view v-if="localVal === 'true'" class="opt-radio-dot" /></view>
            <text>对</text>
          </view>
          <view class="opt" :class="{ active: localVal === 'false' }" @click="setVal('false')">
            <view class="opt-radio"><view v-if="localVal === 'false'" class="opt-radio-dot" /></view>
            <text>错</text>
          </view>
        </view>

        <!-- date -->
        <view v-else-if="q.type==='date'">
          <picker mode="date" @change="(e) => setVal(e.detail.value)">
            <view class="q-input">{{ localVal || '请选择日期' }}</view>
          </picker>
        </view>

        <!-- time -->
        <view v-else-if="q.type==='time'">
          <picker mode="time" @change="(e) => setVal(e.detail.value)">
            <view class="q-input">{{ localVal || '请选择时间' }}</view>
          </picker>
        </view>

        <!-- dateRange -->
        <view v-else-if="q.type==='dateRange'" class="q-field-stack">
          <picker mode="date" @change="(e) => { const a = localVal||['','']; a[0]=e.detail.value; setVal(a) }">
            <view class="q-input">{{ (localVal||[''])[0] || '开始日期' }}</view>
          </picker>
          <picker mode="date" @change="(e) => { const a = localVal||['','']; a[1]=e.detail.value; setVal(a) }">
            <view class="q-input">{{ (localVal||[''])[1] || '结束日期' }}</view>
          </picker>
        </view>

        <!-- rating -->
        <view v-else-if="q.type==='rating'" class="q-rating">
          <view v-for="i in (q.props?.maxRating || 5)" :key="i" class="q-star" :class="{ active: i <= (localVal || 0) }" @click="setVal(i)">★</view>
        </view>

        <!-- nps -->
        <view v-else-if="q.type==='nps'" class="q-nps">
          <view class="q-nps-labels"><text>0</text><text>10</text></view>
          <view class="q-nps-stars">
            <view v-for="i in 11" :key="i" class="q-nps-item" :class="{ active: localVal === i - 1 }" @click="setVal(i - 1)">{{ i - 1 }}</view>
          </view>
        </view>

        <!-- switch -->
        <view v-else-if="q.type==='switch'">
          <switch :checked="localVal||false" @change="(e) => setVal(e.detail.value)" color="#3873f6" />
        </view>

        <!-- phone / name / studentId / employeeId / class -->
        <input v-else-if="['phone','name','studentId','employeeId','class'].includes(q.type)" class="q-input" :value="localVal" @input="onInput" placeholder="请输入" />

        <!-- email -->
        <input v-else-if="q.type==='email'" class="q-input" :value="localVal" @input="onInput" placeholder="邮箱地址" />

        <!-- idCard -->
        <input v-else-if="q.type==='idCard'" class="q-input" :value="localVal" @input="onInput" placeholder="身份证号" />

        <!-- password -->
        <input v-else-if="q.type==='password'" class="q-input" :value="localVal" @input="onInput" type="password" placeholder="密码" />

        <!-- scanCode -->
        <view v-else-if="q.type==='scanCode'" class="q-scan-row">
          <input class="q-input q-scan-input" :value="localVal" disabled placeholder="扫码" />
          <text class="q-scan-btn" @click="onScan">扫码</text>
        </view>

        <!-- file -->
        <view v-else-if="q.type==='file'" class="q-file">
          <button class="q-file-btn" @click="pickFile">选择文件</button>
          <view v-if="(fileList||[]).length" class="q-file-list">
            <view v-for="(f, fi) in fileList" :key="fi" class="q-file-item">
              <text class="q-file-name">{{ f.name }}</text>
              <text class="q-file-del" @click="removeFile(fi)">×</text>
            </view>
          </view>
        </view>

        <!-- image -->
        <view v-else-if="q.type==='image'">
          <button class="q-file-btn" @click="pickImage">选择图片</button>
          <image v-if="imageUrl" :src="imageUrl" mode="aspectFit" style="width:200rpx;height:200rpx;margin-top:16rpx;border-radius:8rpx" />
        </view>

        <!-- location -->
        <view v-else-if="q.type==='location'" class="q-location">
          <text v-if="localVal" class="q-location-text">{{ localVal }}</text>
          <button v-else-if="!showLocInput" class="q-file-btn" @click="pickLocation">选择位置</button>
          <input v-else class="q-input" :value="localVal" @input="onInput" placeholder="输入位置描述" />
        </view>

        <!-- richText -->
        <textarea v-else-if="q.type==='richText'" class="q-textarea" :value="localVal" @input="onInput" :placeholder="q.placeholder || '请输入内容'" />

        <!-- signature -->
        <view v-else-if="q.type==='signature'" class="q-sig">
          <view v-if="!localVal" class="q-sig-trigger" @click="openSig">
            <text class="q-sig-placeholder">点击签名</text>
          </view>
          <view v-else class="q-sig-done" @click="openSig">
            <image :src="localVal" mode="aspectFit" class="q-sig-preview" :key="'sig-' + sigKey" />
            <text class="q-sig-rename">重新签名</text>
          </view>
        </view>

        <!-- matrixRadio -->
        <view v-else-if="q.type==='matrixRadio'" class="q-matrix">
          <view v-for="(r, ri) in (q.props?.rows||[])" :key="ri" class="q-matrix-row">
            <text class="q-matrix-label">{{ r?.title||r }}</text>
            <view class="q-matrix-opts">
              <view v-for="(c, ci) in (q.props?.columns||[])" :key="ci" class="q-matrix-opt" :class="{ active: localVal?.[ri] === (c?.value||c?.title||c?.label||c) }" @click="setMatrix(ri, c)">{{ c?.title||c?.label||c }}</view>
            </view>
          </view>
        </view>

        <!-- matrixCheckbox -->
        <view v-else-if="q.type==='matrixCheckbox'" class="q-matrix">
          <view v-for="(r, ri) in (q.props?.rows||[])" :key="ri" class="q-matrix-row">
            <text class="q-matrix-label">{{ r?.title||r }}</text>
            <view class="q-matrix-opts">
              <view v-for="(c, ci) in (q.props?.columns||[])" :key="ci" class="q-matrix-opt" :class="{ active: (localVal?.[ri]||[]).includes(c?.value||c?.title||c?.label||c) }" @click="toggleMatrix(ri, c)">{{ c?.title||c?.label||c }}</view>
            </view>
          </view>
        </view>

        <!-- matrixFillBlank -->
        <view v-else-if="q.type==='matrixFillBlank'" class="q-matrix">
          <view v-for="(r, ri) in (q.props?.rows||[])" :key="ri" class="q-matrix-row">
            <text class="q-matrix-label">{{ r?.title||r }}</text>
            <view v-for="(c, ci) in (q.props?.columns||[])" :key="ci" class="q-matrix-cell">
              <input class="q-input" :value="localVal?.[ri]?.[ci]||''" @input="(e) => setMatrixFill(ri, ci, e.detail.value)" placeholder="填空" />
            </view>
          </view>
        </view>

        <!-- matrixAuto -->
        <view v-else-if="q.type==='matrixAuto'" class="q-matrix">
          <view v-for="(r, ri) in (localVal||[])" :key="ri" class="q-matrix-row">
            <text class="q-matrix-label">{{ ri + 1 }}</text>
            <view v-for="(c, ci) in (q.props?.columns||[])" :key="ci" class="q-matrix-cell">
              <input class="q-input" :value="localVal?.[ri]?.[ci]||''" @input="(e) => { const a = [...(localVal||[])]; if(!a[ri]) a[ri]=[]; a[ri][ci]=e.detail.value; setVal(a) }" :placeholder="c?.label||'值'" />
            </view>
            <text class="q-matrix-del" @click="removeAutoRow(ri)">×</text>
          </view>
          <view class="q-matrix-add"><button class="q-file-btn" @click="addAutoRow">+ 添加行</button></view>
        </view>

        <!-- user / dept -->
        <view v-else-if="['user','dept'].includes(q.type)" class="q-picker-trigger" @click="openUDTree">
          <view class="q-input">{{ udLabel || (q.type==='user'?'选择成员':'选择部门') }}</view>
        </view>

        <!-- fallback -->
        <input v-else class="q-input" :value="localVal" @input="onInput" :placeholder="q.placeholder || '请输入'" />
      </view>
    </template>
  </view>
  <view v-if="showSigPanel" class="sig-overlay" @touchmove.stop.prevent>
    <view class="sig-header">
      <text class="sig-cancel" @click="closeSig">取消</text>
      <text class="sig-title">请签名</text>
      <text class="sig-confirm" @click="confirmSig">确认</text>
    </view>
    <canvas :canvas-id="sigId" :id="sigId" class="sig-canvas" @touchstart="sigStart" @touchmove="sigMove" @touchend="sigEnd" />
    <view class="sig-footer"><text class="sig-clear" @click="clearSig">清除</text></view>
  </view>
  <view v-if="showUDTree" class="ud-overlay" @touchmove.prevent>
    <view class="ud-modal">
      <view class="ud-header">
        <text class="ud-cancel" @click="closeUDTree">{{ isUDMulti ? '取消' : '关闭' }}</text>
        <text class="ud-title">{{ q.type==='user'?'选择成员':'选择部门' }}</text>
        <text v-if="isUDMulti" class="ud-confirm" @click="confirmUDTree">确定</text>
        <text v-else class="ud-confirm" style="visibility:hidden">确定</text>
      </view>
      <scroll-view class="ud-list" scroll-y>
        <view v-for="(node, i) in udTreeFlat" :key="node.value || i" class="ud-tree-node" :style="{ paddingLeft: (node.level * 48 + 32) + 'rpx' }" @click="onUDNodeTap(node)">
          <text v-if="!node.isLeaf" class="ud-arrow">{{ node.expanded ? '▼' : '▶' }}</text>
          <view v-else :class="['ud-chk-box', { 'ud-chk-box--sel': isUDSelected(node.value) }]"><text v-if="isUDSelected(node.value)" class="ud-chk-icon">✓</text></view>
          <text class="ud-label">{{ node.label }}</text>
        </view>
      </scroll-view>
    </view>
  </view>
</view>
</template>

<script>
import CONFIG from '../../config'
import { surveyApi } from '../../api/index'

const LAYOUT_TYPES = ['description', 'divider', 'pagination']

export default {
  name: 'QuestionField',
  props: {
    q: { type: Object, required: true },
    index: { type: Number, default: 0 },
    value: { type: null, default: '' },
    fileList: { type: Array, default: () => [] },
    showNumber: { type: Boolean, default: true },
    qScore: { type: [Number, String], default: '' }
  },
  data() {
    return {
      LAYOUT_TYPES,
      sigId: 'sig_' + (this.q?.id || '') + '_' + Math.random().toString(36).slice(2, 8),
      sigCtx: null,
      sigDrawing: false,
      sigLastX: 0,
      sigLastY: 0,
      showSigPanel: false,
      showUDTree: false,
      showLocInput: false,
      udTreeData: [],
      udTempVals: [],
      isUDMulti: false,
      imageUrl: '',
      sigKey: 0
    }
  },
  computed: {
    localVal: {
      get() { return this.value },
      set(v) { this.$emit('input', v) }
    },
    udLabel() {
      const q = this.q
      if (!q) return ''
      const flat = this.buildUserDeptOpts(q)
      if (q.multiple) {
        const vals = Array.isArray(this.localVal) ? this.localVal : []
        return vals.map(v => { const o = flat.find(x => x.value === v); return o ? o.label : v }).join(', ')
      }
      const o = flat.find(x => x.value === this.localVal)
      return o ? o.label : ''
    },
    processedTitle() {
      if (!this.q?.title) return ''
      return this.q.title.replace(/<img\b/gi, '<img style="max-width:100%;height:auto" ')
    },
    processedDesc() {
      if (!this.q?.description) return ''
      return this.q.description.replace(/<img\b/gi, '<img style="max-width:100%;height:auto" ')
    },
    processedFullTitle() {
      const q = this.q || {}
      let html = ''
      if (this.showNumber) {
        html += `<span style="color:#3873f6;font-weight:600;margin-right:6rpx">${this.index + 1}.</span>`
      }
      if (this.qScore) {
        html += `<span style="font-size:24rpx;color:#fa8c16;font-weight:500;padding-top:4rpx;margin-right:8rpx">(${this.qScore}分)</span>`
      }
      if (q.required) {
        html += `<span style="color:#f56c6c;margin-right:6rpx">*</span>`
      }
      html += this.unwrapOuterP(q.title || '')
      return html.replace(/<img\b/gi, '<img style="max-width:100%;height:auto" ')
    },
    udTreeFlat() {
      const result = []
      const walk = (nodes, level) => {
        nodes.forEach(node => {
          result.push({ ...node, level, isLeaf: !node.children?.length })
          if (node.expanded && node.children?.length) {
            walk(node.children, level + 1)
          }
        })
      }
      walk(this.udTreeData, 0)
      return result
    }
  },
  mounted() {
  },
  methods: {
    stripHtml(html) { return html ? html.replace(/<[^>]+>/g, '').trim() : '' },
    parseOpts(q) {
      const opts = q.props?.options || []
      if (!opts.length) return []
      return opts.map(o => {
        if (typeof o === 'string') return { label: o, value: o }
        return { label: o.label || o.value || '', value: o.value !== undefined ? o.value : (o.label || '') }
      })
    },
    getOptLabel(opts, val) { const o = opts.find(x => x.value === val); return o ? this.stripHtml(o.label) : '' },
    setVal(v) { this.$emit('input', v) },
    onInput(e) { this.setVal(e.detail.value) },
    onScan() {
      uni.scanCode({
        success: (res) => { this.setVal(res.result) },
        fail: () => { uni.showToast({ title: '扫码失败', icon: 'none' }) }
      })
    },
    onArrayInput(fi, e) {
      const a = Array.isArray(this.localVal) ? [...this.localVal] : (this.q?.props?.fields||[]).map(() => '')
      a[fi] = e.detail.value
      this.setVal(a)
    },
    isChk(val) { return Array.isArray(this.localVal) && this.localVal.includes(val) },
    toggleChk(val) {
      let a = this.localVal
      if (!Array.isArray(a)) a = []
      const i = a.indexOf(val)
      if (i >= 0) a.splice(i, 1)
      else a.push(val)
      this.$emit('input', a)
    },
    onSelect(opts, e) {
      const i = Number(e.detail.value)
      this.setVal(opts[i] ? opts[i].value : '')
    },
    cascadeRange(q) {
      const opts = this.parseOpts(q)
      const levels = Math.max(...opts.map(o => (o.label || '').split('/').length), 1)
      const range = []
      for (let i = 0; i < levels; i++) range.push([])
      opts.forEach(o => {
        const parts = (o.label || '').split('/')
        parts.forEach((p, i) => { if (i < range.length && !range[i].includes(p.trim())) range[i].push(p.trim()) })
      })
      return range.map(r => r.length ? r : ['选项'])
    },
    onCascade(q, e) {
      const vals = e.detail.value
      const range = this.cascadeRange(q)
      const parts = vals.map((v, i) => range[i]?.[v] || '')
      this.setVal(parts)
    },
    setMatrix(ri, c) {
      const val = c?.value || c?.title || c?.label || c
      const obj = { ...(this.localVal || {}) }
      obj[ri] = val
      this.setVal(obj)
    },
    toggleMatrix(ri, c) {
      const val = c?.value || c?.title || c?.label || c
      const obj = { ...(this.localVal || {}) }
      if (!Array.isArray(obj[ri])) obj[ri] = []
      const i = obj[ri].indexOf(val)
      if (i >= 0) obj[ri].splice(i, 1)
      else obj[ri].push(val)
      this.setVal(obj)
    },
    setMatrixFill(ri, ci, v) {
      const obj = { ...(this.localVal || {}) }
      if (!obj[ri]) obj[ri] = {}
      obj[ri][ci] = v
      this.setVal(obj)
    },
    addAutoRow() {
      const cols = this.q?.props?.columns?.length || 0
      const arr = [...(this.localVal || [])]
      arr.push(Array(cols).fill(''))
      this.setVal(arr)
    },
    removeAutoRow(ri) {
      const arr = [...(this.localVal || [])]
      arr.splice(ri, 1)
      this.setVal(arr)
    },
    buildUserDeptOpts(q) {
      return (q.props?.options || []).map(o => ({
        label: o.label || o.name || o.value || '',
        value: o.value !== undefined ? o.value : (o.label || '')
      }))
    },
    buildUserDeptTree(q) {
      const opts = q.props?.options || []
      if (!opts.length) return []
      if (q.type === 'user') {
        const deptMap = {}
        opts.forEach(o => {
          const deptId = o.deptId || ''
          if (!deptMap[deptId]) {
            deptMap[deptId] = { value: '__d__' + deptId, label: o.deptName || deptId || '未分组', children: [] }
          }
          deptMap[deptId].children.push({ value: o.value, label: o.label || '成员' })
        })
        return Object.values(deptMap)
      }
      const map = {}
      opts.forEach(o => { map[o.value] = { ...o, children: [] } })
      const roots = []
      opts.forEach(o => {
        if (o.parentId && map[o.parentId]) {
          map[o.parentId].children.push(map[o.value])
        } else {
          roots.push(map[o.value])
        }
      })
      return roots
    },
    getUDLabel(q, val) {
      const o = this.buildUserDeptOpts(q).find(x => x.value === val)
      return o ? o.label : val
    },
    openUDTree() {
      this.isUDMulti = !!this.q?.multiple
      const expandAll = (nodes) => nodes.forEach(n => { n.expanded = true; if (n.children) expandAll(n.children) })
      const tree = this.buildUserDeptTree(this.q)
      expandAll(tree)
      this.udTreeData = tree
      if (this.isUDMulti) {
        this.udTempVals = Array.isArray(this.localVal) ? [...this.localVal] : []
      }
      this.showUDTree = true
    },
    closeUDTree() { this.showUDTree = false },
    onUDNodeTap(node) {
      if (!node.isLeaf) {
        node.expanded = !node.expanded
        this.$forceUpdate()
        return
      }
      if (this.isUDMulti) {
        const i = this.udTempVals.indexOf(node.value)
        if (i >= 0) this.udTempVals.splice(i, 1)
        else this.udTempVals.push(node.value)
      } else {
        this.setVal(node.value)
        this.showUDTree = false
      }
    },
    isUDSelected(val) {
      if (this.isUDMulti) return this.udTempVals.includes(val)
      return this.localVal === val
    },
    confirmUDTree() {
      this.setVal([...this.udTempVals])
      this.showUDTree = false
    },
    pickFile() {
      const accept = this.q?.props?.accept || ''
      const fileTypes = this.q?.fileTypes || []
      const maxCount = this.q?.maxFileCount || 9
      if (accept.includes('image') || fileTypes.includes('image')) {
        uni.chooseImage({
          count: maxCount,
          success: (res) => { this.doUpload(res.tempFilePaths, res.tempFiles) },
          fail: () => { uni.showToast({ title: '选择图片失败', icon: 'none' }) }
        })
      } else if (typeof uni.chooseFile === 'function') {
        uni.chooseFile({
          count: maxCount,
          success: (res) => {
            const paths = res.tempFiles?.map(f => f.path || f.tempFilePath) || []
            this.doUpload(paths, res.tempFiles)
          },
          fail: () => { uni.showToast({ title: '选择文件失败', icon: 'none' }) }
        })
      } else {
        uni.showToast({ title: '暂不支持此文件类型', icon: 'none' })
      }
    },
    doUpload(paths, tempFiles) {
      if (!paths.length) return
      uni.showLoading({ title: '上传中...' })
      let uploaded = 0
      const total = paths.length
      for (const p of paths) {
        uni.uploadFile({
          url: CONFIG.BASE_URL + '/upload',
          filePath: p,
          name: 'file',
          success: (res) => {
            if (res.statusCode !== 200) {
              const msg = res.statusCode === 413 ? '上传文件过大' : ('上传失败(状态' + res.statusCode + ')')
              uni.showToast({ title: msg, icon: 'none' })
            } else {
              try {
                const data = JSON.parse(res.data)
                if (data.code === 0 && data.data.url) {
                  const fullUrl = (data.data.domain || '') + data.data.url
                  this.fileList.push({ name: data.data.url.split('/').pop() || 'file', url: fullUrl })
                } else {
                  uni.showToast({ title: data.msg || '上传失败', icon: 'none' })
                }
              } catch (e) {
                uni.showToast({ title: '上传失败，文件可能过大或不支持', icon: 'none' })
              }
            }
          },
          fail: () => { uni.showToast({ title: '网络异常，上传失败', icon: 'none' }) },
          complete: () => {
            uploaded++
            if (uploaded >= total) {
              uni.hideLoading()
              this.$emit('update:fileList', this.fileList)
              this.setVal(this.fileList.map(f => f.url).join(','))
            }
          }
        })
      }
    },

    removeFile(fi) {
      this.fileList.splice(fi, 1)
      this.$emit('update:fileList', this.fileList)
      this.setVal(this.fileList.length ? this.fileList.map(f => f.url).join(',') : '')
    },
    pickImage() {
      uni.chooseImage({
        count: 1,
        success: (res) => {
          const p = res.tempFilePaths[0]
          uni.showLoading({ title: '上传中...' })
          uni.uploadFile({
            url: CONFIG.BASE_URL + '/upload',
            filePath: p,
            name: 'file',
            success: (r) => {
              if (r.statusCode !== 200) {
                uni.showToast({ title: r.statusCode === 413 ? '图片过大' : '上传失败', icon: 'none' })
              } else {
                try {
                  const data = JSON.parse(r.data)
                  if (data.code === 0 && data.data.url) {
                    const url = (data.data.domain || '') + data.data.url
                    this.imageUrl = url
                    this.setVal(url)
                  } else {
                    uni.showToast({ title: data.msg || '上传失败', icon: 'none' })
                  }
                } catch (e) {
                  uni.showToast({ title: '上传失败', icon: 'none' })
                }
              }
            },
            fail: () => { uni.showToast({ title: '图片上传失败', icon: 'none' }) },
            complete: () => { uni.hideLoading() }
          })
        },
        fail: () => { uni.showToast({ title: '选择图片失败', icon: 'none' }) }
      })
    },
    pickLocation() {
      try {
        uni.chooseLocation({
          success: (res) => { this.setVal(`${res.address} (${res.latitude},${res.longitude})`) },
          fail: (err) => {
            if (err?.errMsg?.includes('cancel')) return
            uni.showToast({ title: '定位失败，已切换手动输入', icon: 'none' })
            this.showLocInput = true
          }
        })
      } catch (e) {
        uni.showToast({ title: '定位功能不可用', icon: 'none' })
        this.showLocInput = true
      }
    },
    openSig() {
      this.showSigPanel = true
      this.sigDrawing = false
      this.$emit('sig-open')
      this.$nextTick(() => {
        this.sigCtx = uni.createCanvasContext(this.sigId, this)
        this.sigCtx.setStrokeStyle('#333')
        this.sigCtx.setLineWidth(4)
        this.sigCtx.setLineCap('round')
      })
    },
    closeSig() {
      this.showSigPanel = false
      this.sigCtx = null
      this.$emit('sig-close')
    },
    confirmSig() {
      if (!this.sigCtx) { this.closeSig(); return }
      // 等待帧渲染完成再截图
      uni.canvasToTempFilePath({
        canvasId: this.sigId,
        success: (res) => {
          this.sigKey++
          const tmpPath = res.tempFilePath
          uni.saveFile({
            tempFilePath: tmpPath,
            success: (s) => { this.setVal(s.savedFilePath); this.closeSig() },
            fail: () => { this.setVal(tmpPath); this.closeSig() }
          })
        },
        fail: () => { uni.showToast({ title: '签名保存失败', icon: 'none' }); this.closeSig() }
      }, this)
    },
    sigStart(e) {
      if (!this.sigCtx) return
      this.sigDrawing = true
      const t = e.touches[0]
      this.sigLastX = t.x
      this.sigLastY = t.y
      this.sigCtx.beginPath()
      this.sigCtx.moveTo(t.x, t.y)
      this.sigCtx.stroke()
      this.sigCtx.draw(true)
    },
    sigMove(e) {
      if (!this.sigDrawing || !this.sigCtx) return
      const t = e.touches[0]
      this.sigCtx.beginPath()
      this.sigCtx.moveTo(this.sigLastX, this.sigLastY)
      this.sigCtx.lineTo(t.x, t.y)
      this.sigCtx.stroke()
      this.sigCtx.draw(true)
      this.sigLastX = t.x
      this.sigLastY = t.y
    },
    sigEnd() {
      this.sigDrawing = false
    },
    clearSig() {
      if (!this.sigCtx) return
      uni.createSelectorQuery().in(this).select('.sig-canvas').boundingClientRect((rect) => {
        if (rect) {
          this.sigCtx.clearRect(0, 0, rect.width, rect.height)
          this.sigCtx.draw()
        }
      }).exec()
    },
    unwrapOuterP(html) {
      if (!html) return ''
      const trimmed = html.trim()
      const match = trimmed.match(/^<p([^>]*)>([\s\S]*)<\/p>$/)
      if (!match) return html
      const attrs = match[1]
      const content = match[2]
      if (/\bql-align-\w+\b/.test(attrs)) {
        return `<span${attrs} style="display:inline-block;width:100%;text-align:inherit">${content}</span>`
      }
      return content
    }
  }
}
</script>

<style scoped>
.q-item { margin-bottom: 28rpx; }
.q-title { font-size: 30rpx; color: #333; font-weight: 500; margin-bottom: 20rpx; word-break: break-word; line-height: 1.5; }
.q-title :deep(img) { max-width: 100%; height: auto; }
:deep(.ql-align-center) { text-align: center; }
:deep(.ql-align-right) { text-align: right; }
:deep(.ql-align-justify) { text-align: justify; }
:deep(.ql-code-block-container) { background: #f5f5f5; border-radius: 6px; padding: 12px 16px; margin: 8px 0; font-family: monospace; font-size: 26rpx; line-height: 1.6; overflow-x: auto; }
:deep(.ql-code-block) { white-space: pre; }
.q-desc { background: #f5f5f5; border-radius: 8rpx; padding: 16rpx 20rpx; margin-bottom: 16rpx; font-size: 26rpx; color: #333; line-height: 1.6; word-break: break-word; display: flex; align-items: flex-start; white-space: pre-wrap; }
.q-desc-label { font-weight: 500; color: #666; flex-shrink: 0; }
.q-input-area { padding-left: 10rpx; }
.q-input { background: #f7f8fa; border-radius: 12rpx; padding: 20rpx 24rpx; font-size: 28rpx; min-height: 88rpx; width: 100%; box-sizing: border-box; }
.q-textarea { background: #f7f8fa; border-radius: 12rpx; padding: 20rpx 24rpx; font-size: 28rpx; width: 100%; box-sizing: border-box; min-height: 200rpx; }
.q-field-stack { display: flex; flex-direction: column; gap: 12rpx; }
.q-field-row { display: flex; flex-wrap: wrap; gap: 12rpx; }
.q-field-row .q-input { flex: 1; min-width: 200rpx; }
.q-desc { font-size: 28rpx; color: #666; padding: 16rpx 0; line-height: 1.6; }
.q-divider { height: 2rpx; background: #e8e8e8; margin: 16rpx 0; }

.opt-list { display: flex; flex-direction: column; gap: 16rpx; }
.opt { display: flex; align-items: center; gap: 16rpx; padding: 20rpx 24rpx; background: #f7f8fa; border-radius: 12rpx; border: 2rpx solid transparent; min-height: 48rpx; }
.opt.active { background: #eef2ff; border-color: #3873f6; }
.opt-radio { width: 40rpx; height: 40rpx; border-radius: 50%; border: 2rpx solid #d0d0d0; flex-shrink: 0; display: flex; align-items: center; justify-content: center; }
.opt.active .opt-radio { border-color: #3873f6; }
.opt-radio-dot { width: 22rpx; height: 22rpx; border-radius: 50%; background: #3873f6; }
.opt-checkbox { width: 40rpx; height: 40rpx; border-radius: 8rpx; border: 2rpx solid #d0d0d0; flex-shrink: 0; display: flex; align-items: center; justify-content: center; }
.opt.active .opt-checkbox { border-color: #3873f6; background: #3873f6; }
.opt-checkbox-icon { color: #fff; font-size: 24rpx; }

.q-rating { display: flex; gap: 12rpx; }
.q-star { font-size: 60rpx; color: #ddd; }
.q-star.active { color: #f5a623; }
.q-nps { padding: 8rpx 0; }
.q-nps-labels { display: flex; justify-content: space-between; font-size: 24rpx; color: #909399; margin-bottom: 12rpx; }
.q-nps-stars { display: flex; gap: 8rpx; flex-wrap: wrap; }
.q-nps-item { width: 64rpx; height: 64rpx; border-radius: 8rpx; background: #f7f8fa; color: #606266; font-size: 26rpx; display: flex; align-items: center; justify-content: center; }
.q-nps-item.active { background: #3873f6; color: #fff; }
.q-file-btn { background: #f7f8fa; border: 1px dashed #d0d0d0; border-radius: 12rpx; padding: 20rpx 24rpx; font-size: 28rpx; color: #606266; text-align: center; width: 100%; }
.q-file-list { margin-top: 12rpx; display: flex; flex-direction: column; gap: 8rpx; }
.q-file-item { display: flex; align-items: center; gap: 12rpx; font-size: 26rpx; }
.q-file-name { flex: 1; color: #333; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.q-file-del { color: #f56c6c; padding: 4rpx 12rpx; font-size: 32rpx; }
.q-location-text { font-size: 26rpx; color: #3873f6; }
.q-scan-row { display: flex; align-items: center; gap: 16rpx; }
.q-scan-input { flex: 1; }
.q-scan-btn { flex-shrink: 0; padding: 16rpx 28rpx; background: #3873f6; color: #fff; border-radius: 12rpx; font-size: 28rpx; text-align: center; }

.q-matrix { border: 1rpx solid #e8e8e8; border-radius: 12rpx; overflow: hidden; }
.q-matrix-row { display: flex; align-items: center; border-bottom: 1rpx solid #f0f0f0; }
.q-matrix-row:last-child { border-bottom: none; }
.q-matrix-label { width: 120rpx; padding: 16rpx 12rpx; font-size: 26rpx; color: #666; flex-shrink: 0; text-align: center; background: #fafafa; }
.q-matrix-opts { display: flex; gap: 0; flex: 1; }
.q-matrix-opt { flex: 1; padding: 16rpx 8rpx; text-align: center; font-size: 24rpx; color: #606266; border-left: 1rpx solid #f0f0f0; }
.q-matrix-opt.active { background: #eef2ff; color: #3873f6; font-weight: 500; }
.q-matrix-cell { flex: 1; padding: 8rpx; border-left: 1rpx solid #f0f0f0; }
.q-matrix-cell .q-input { min-height: 60rpx; padding: 8rpx 12rpx; font-size: 24rpx; }
.q-matrix-del { padding: 16rpx 12rpx; color: #f56c6c; font-size: 28rpx; flex-shrink: 0; }
.q-matrix-add { padding: 12rpx; }
.q-rich-text { font-size: 28rpx; color: #666; line-height: 1.6; }

.q-sig { width: 100%; }
.q-sig-trigger { display: flex; align-items: center; justify-content: center; height: 160rpx; border: 2rpx dashed #d0d0d0; border-radius: 12rpx; background: #f7f8fa; }
.q-sig-placeholder { font-size: 28rpx; color: #909399; }
.q-sig-done { text-align: center; }
.q-sig-preview { width: 100%; height: 200rpx; border-radius: 12rpx; margin-top: 12rpx; }
.q-sig-rename { font-size: 24rpx; color: #3873f6; margin-top: 8rpx; display: inline-block; }

.sig-overlay { position: fixed; inset: 0; z-index: 9999; background: #fff; display: flex; flex-direction: column; touch-action: none; overscroll-behavior: none; }
.sig-header { display: flex; align-items: center; justify-content: space-between; padding: 24rpx 32rpx; border-bottom: 1rpx solid #f0f0f0; }
.sig-cancel { font-size: 28rpx; color: #909399; padding: 8rpx; }
.sig-title { font-size: 32rpx; font-weight: 500; color: #333; }
.sig-confirm { font-size: 28rpx; color: #3873f6; font-weight: 500; padding: 8rpx; }
.sig-canvas { flex: 1; width: 100%; }
.sig-footer { display: flex; justify-content: center; padding: 24rpx 32rpx; border-top: 1rpx solid #f0f0f0; }
.sig-clear { font-size: 28rpx; color: #f56c6c; padding: 8rpx 32rpx; }

.q-picker-trigger { width: 100%; }
.ud-overlay { position: fixed; inset: 0; z-index: 9999; background: rgba(0,0,0,.5); display: flex; align-items: flex-end; }
.ud-modal { width: 100%; max-height: 70vh; background: #fff; border-radius: 24rpx 24rpx 0 0; display: flex; flex-direction: column; }
.ud-header { display: flex; align-items: center; justify-content: space-between; padding: 28rpx 32rpx; border-bottom: 1rpx solid #f0f0f0; }
.ud-cancel { font-size: 28rpx; color: #909399; padding: 8rpx; }
.ud-title { font-size: 32rpx; font-weight: 500; color: #333; }
.ud-confirm { font-size: 28rpx; color: #3873f6; font-weight: 500; padding: 8rpx; }
.ud-list { max-height: 60vh; overflow-y: auto; padding: 16rpx 0; }
.ud-tree-node { display: flex; align-items: center; gap: 12rpx; padding: 22rpx 32rpx; min-height: 72rpx; }
.ud-arrow { font-size: 20rpx; color: #909399; width: 28rpx; text-align: center; flex-shrink: 0; }
.ud-chk-box { width: 40rpx; height: 40rpx; border-radius: 8rpx; border: 2rpx solid #d0d0d0; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.ud-chk-box--sel { background: #3873f6; border-color: #3873f6; }
.ud-chk-icon { color: #fff; font-size: 24rpx; }
</style>
