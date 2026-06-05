<template>
  <view class="container">
    <scroll-view scroll-y class="form-scroll">
      <view class="form-group">
        <text class="group-title">基本信息</text>
        <view class="form-item">
          <text class="label">标题 <text class="red">*</text></text>
          <input v-model="form.title" placeholder="请输入标题" class="input" />
        </view>
        <view class="form-item">
          <text class="label">类型 <text class="red">*</text></text>
          <picker :range="['活动', '赛事']" @change="(e) => { form.type = e.detail.value === 0 ? 1 : 2 }">
            <view class="picker">{{ form.type === 1 ? '活动' : '赛事' }}</view>
          </picker>
        </view>
        <view class="form-item">
          <text class="label">封面</text>
          <view class="cover-upload" @click="uploadImage">
            <image v-if="form.cover" :src="form.cover" mode="aspectFill" class="cover-preview" />
            <view v-else class="cover-placeholder">+</view>
            <view v-if="form.cover" class="cover-del" @click.stop="form.cover = ''">×</view>
          </view>
        </view>
        <view class="form-item">
          <text class="label">描述</text>
          <textarea v-model="form.desc" placeholder="简短描述" class="textarea" />
        </view>
        <view class="form-item">
          <text class="label">{{ form.type === 2 ? '赛事规则' : '活动规则' }}</text>
          <textarea v-model="form.rules" placeholder="填写活动/赛事的规则说明" class="textarea textarea-lg" />
        </view>
        <view class="form-item">
          <text class="label">分类</text>
          <picker :range="categories" range-key="label" @change="onCateChange">
            <view class="picker" :class="{ placeholder: !form.cateName }">{{ form.cateName || '请选择分类' }}</view>
          </picker>
        </view>
        <view class="form-item">
          <text class="label">排序</text>
          <input v-model="form.order" type="number" placeholder="排序值" class="input" />
        </view>
      </view>

      <view class="form-group">
        <text class="group-title">时间设置</text>
        <view class="form-item">
          <text class="label">报名开始</text>
          <view class="row">
            <picker mode="date" :value="form.regStartDate" @change="(e) => { form.regStartDate = e.detail.value; updateDateTime('regStart') }">
              <view class="picker flex-1" :class="{ placeholder: !form.regStartDate }">{{ form.regStartDate || '日期' }}</view>
            </picker>
            <picker mode="time" :value="form.regStartTime" @change="(e) => { form.regStartTime = e.detail.value; updateDateTime('regStart') }">
              <view class="picker flex-1" :class="{ placeholder: !form.regStartTime }">{{ form.regStartTime || '时间' }}</view>
            </picker>
          </view>
        </view>
        <view class="form-item">
          <text class="label">报名结束</text>
          <view class="row">
            <picker mode="date" :value="form.regEndDate" @change="(e) => { form.regEndDate = e.detail.value; updateDateTime('regEnd') }">
              <view class="picker flex-1" :class="{ placeholder: !form.regEndDate }">{{ form.regEndDate || '日期' }}</view>
            </picker>
            <picker mode="time" :value="form.regEndTime" @change="(e) => { form.regEndTime = e.detail.value; updateDateTime('regEnd') }">
              <view class="picker flex-1" :class="{ placeholder: !form.regEndTime }">{{ form.regEndTime || '时间' }}</view>
            </picker>
          </view>
        </view>
        <view class="form-item">
          <text class="label">活动开始</text>
          <view class="row">
            <picker mode="date" :value="form.eventStartDate" @change="(e) => { form.eventStartDate = e.detail.value; updateDateTime('eventStart') }">
              <view class="picker flex-1" :class="{ placeholder: !form.eventStartDate }">{{ form.eventStartDate || '日期' }}</view>
            </picker>
            <picker mode="time" :value="form.eventStartTime" @change="(e) => { form.eventStartTime = e.detail.value; updateDateTime('eventStart') }">
              <view class="picker flex-1" :class="{ placeholder: !form.eventStartTime }">{{ form.eventStartTime || '时间' }}</view>
            </picker>
          </view>
        </view>
        <view class="form-item">
          <text class="label">活动结束</text>
          <view class="row">
            <picker mode="date" :value="form.eventEndDate" @change="(e) => { form.eventEndDate = e.detail.value; updateDateTime('eventEnd') }">
              <view class="picker flex-1" :class="{ placeholder: !form.eventEndDate }">{{ form.eventEndDate || '日期' }}</view>
            </picker>
            <picker mode="time" :value="form.eventEndTime" @change="(e) => { form.eventEndTime = e.detail.value; updateDateTime('eventEnd') }">
              <view class="picker flex-1" :class="{ placeholder: !form.eventEndTime }">{{ form.eventEndTime || '时间' }}</view>
            </picker>
          </view>
        </view>
      </view>

      <view class="form-group">
        <text class="group-title">发布范围</text>
        <view class="form-item" @click="toggleDeptPicker">
          <text class="label">发布部门</text>
          <view class="picker" :class="{ placeholder: !form.publishDeptIds }">
            {{ form.publishDeptNames || '全部部门可见（点击选择）' }}
          </view>
        </view>
      </view>

      <view class="form-group">
        <text class="group-title">管理角色</text>
        <view class="form-item" @click="toggleUserPicker('organizer')">
          <text class="label">主办人</text>
          <view class="picker" :class="{ placeholder: form.organizers.length === 0 }">
            <text v-if="form.organizers.length > 0">{{ form.organizers.map(u => u.name).join('、') }}</text>
            <text v-else>请选择主办人</text>
          </view>
        </view>
        <view class="form-item" @click="toggleUserPicker('assistant')">
          <text class="label">主办人助理</text>
          <view class="picker" :class="{ placeholder: form.assistants.length === 0 }">
            <text v-if="form.assistants.length > 0">{{ form.assistants.map(u => u.name).join('、') }}</text>
            <text v-else>请选择主办人助理</text>
          </view>
        </view>
        <view class="form-item" v-if="form.type === 2" @click="toggleUserPicker('referee')">
          <text class="label">裁判</text>
          <view class="picker" :class="{ placeholder: form.referees.length === 0 }">
            <text v-if="form.referees.length > 0">{{ form.referees.map(u => u.name).join('、') }}</text>
            <text v-else>请选择裁判</text>
          </view>
        </view>
      </view>

      <view class="form-group">
        <text class="group-title">报名表单</text>
        <view class="form-item" @click="showFormEditor">
          <text class="label">报名表单字段</text>
          <view class="picker" :class="{ placeholder: form.fields.length === 0 }">
            <text v-if="form.fields.length > 0">{{ form.fields.map(f => f.label).join('、') }}</text>
            <text v-else>点击配置报名表单字段</text>
          </view>
        </view>
      </view>

      <view class="form-group" v-if="form.type === 2">
        <text class="group-title">评分项</text>
        <view class="form-item" @click="showScoreEditor">
          <text class="label">评分项</text>
          <view class="picker" :class="{ placeholder: form.scoreFields.length === 0 }">
            <text v-if="form.scoreFields.length > 0">{{ form.scoreFields.map(sf => sf.name).join('、') }}</text>
            <text v-else>点击配置评分项</text>
          </view>
        </view>
      </view>

      <view class="btn-bar">
        <view class="submit-btn" @click="handleSubmit">保存</view>
      </view>
    </scroll-view>

    <!-- 部门选择弹窗 -->
    <view class="modal-mask" v-if="showDeptPicker" @click="showDeptPicker = false">
      <view class="modal-content" @click.stop style="max-height:70vh">
        <text class="modal-title">选择发布部门</text>
        <view v-for="(d, i) in visibleDepts" :key="d.id" class="dept-item" :style="{ paddingLeft: (d.depth * 40 + 20) + 'rpx' }">
          <view class="dept-left" @click="toggleDeptExpand(d.id)">
            <text v-if="d.hasChildren" class="dept-arrow">{{ d.expanded ? '▼' : '▶' }}</text>
            <text v-else class="dept-arrow dept-arrow-placeholder">　</text>
            <text class="dept-name">{{ d.name }}</text>
          </view>
          <text class="dept-check" v-if="selectedDeptIds.includes(d.id)" @click="toggleDeptSelect(d.id)">✓</text>
          <text v-else class="dept-check dept-uncheck" @click="toggleDeptSelect(d.id)">○</text>
        </view>
        <view class="modal-actions">
          <view class="modal-btn btn-cancel" @click="showDeptPicker = false">取消</view>
          <view class="modal-btn btn-confirm" @click="confirmDeptPicker">确定</view>
        </view>
      </view>
    </view>

    <!-- 用户选择弹窗 -->
    <view class="modal-mask" v-if="showUserPickerModal" @click="showUserPickerModal = false">
      <view class="modal-content" @click.stop style="max-height:70vh">
        <text class="modal-title">选择{{ userPickerTitle }}</text>
        <view class="user-picker-depts">
          <view v-for="(d, i) in visibleDepts" :key="d.id" class="dept-item" :style="{ paddingLeft: (d.depth * 40 + 20) + 'rpx' }" @click="handleUserDeptClick(d.id, d.name)">
            <view class="dept-left">
              <text v-if="d.hasChildren" class="dept-arrow" @click.stop="toggleDeptExpand(d.id)">{{ d.expanded ? '▼' : '▶' }}</text>
              <text v-else class="dept-arrow dept-arrow-placeholder">　</text>
              <text class="dept-name">{{ d.name }}</text>
            </view>
            <text class="dept-check" v-if="currentUserDeptId === d.id">◉</text>
            <text v-else class="dept-check dept-uncheck">○</text>
          </view>
        </view>
        <view v-if="userPickerUsers.length > 0" class="user-picker-list">
          <view class="user-picker-title">用户列表</view>
          <view class="user-tags">
            <view class="user-tag" v-for="(u, i) in userPickerUsers" :key="i" :class="{ selected: isUserSelected(u) }" @click="toggleUserSelect(u)">{{ u.name || u.userName || u.nickname }}</view>
          </view>
        </view>
        <view class="modal-actions">
          <view class="modal-btn btn-cancel" @click="showUserPickerModal = false">取消</view>
          <view class="modal-btn btn-confirm" @click="confirmUserPicker">确定</view>
        </view>
      </view>
    </view>

    <!-- 表单字段编辑弹窗 -->
    <view class="modal-mask" v-if="showFieldEditor" @click="closeFieldEditor">
      <view class="modal-content" @click.stop>
        <text class="modal-title">配置报名表单字段</text>
        <view v-for="(f, fi) in editingFields" :key="fi" class="field-editor-row">
          <text class="field-editor-num">{{ fi + 1 }}.</text>
          <input v-model="f.label" placeholder="字段名称" class="field-editor-input" />
          <picker :range="['文本','数字','多行文本','选择','拍照上传','位置签到']" @change="(e) => { f.type = ['text','number','textarea','select','image','location'][e.detail.value] }">
            <view class="field-editor-type">{{ { text:'文本', number:'数字', textarea:'多行文本', select:'选择', image:'拍照上传', location:'位置签到' }[f.type] || '文本' }}</view>
          </picker>
          <input v-if="f.type === 'select'" v-model="f.options" placeholder="选项(逗号分隔)" class="field-editor-options" />
          <view class="field-editor-required" @click="f.required = !f.required">{{ f.required ? '必填' : '选填' }}</view>
          <view class="field-editor-del" @click="editingFields.splice(fi, 1)">×</view>
        </view>
        <view class="field-editor-add" @click="editingFields.push({ label:'', type:'text', options:'', required: false })">+ 添加字段</view>
        <view class="modal-actions">
          <view class="modal-btn btn-confirm" @click="confirmFieldEditor">完成</view>
        </view>
      </view>
    </view>

    <!-- 评分项编辑弹窗 -->
    <view class="modal-mask" v-if="showScoreEditorModal" @click="closeScoreEditor">
      <view class="modal-content" @click.stop>
        <text class="modal-title">配置评分项</text>
        <view v-for="(sf, fi) in editingScores" :key="fi" class="field-editor-row">
          <text class="field-editor-num">{{ fi + 1 }}.</text>
          <input v-model="sf.name" placeholder="评分项名称" class="field-editor-input" />
          <picker :range="['数字','文本','选择']" @change="(e) => { sf.type = ['number','text','select'][e.detail.value] }">
            <view class="field-editor-type">{{ { number:'数字', text:'文本', select:'选择' }[sf.type] || '数字' }}</view>
          </picker>
          <input v-if="sf.type === 'select'" v-model="sf.options" placeholder="选项(逗号分隔)" class="field-editor-options" />
          <view class="field-editor-del" @click="editingScores.splice(fi, 1)">×</view>
        </view>
        <view class="field-editor-add" @click="editingScores.push({ name: '', type: 'number', options: '' })">+ 添加评分项</view>
        <view class="modal-actions">
          <view class="modal-btn btn-confirm" @click="confirmScoreEditor">完成</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { adminApi, dictApi } from '../../../api/admin'
export default {
  data() {
    return {
      id: '',
      form: {
        title: '', type: 1, cover: '', desc: '', rules: '', cateName: '', order: 0,
        regStartDate: '', regStartTime: '', regEndDate: '', regEndTime: '',
        eventStartDate: '', eventStartTime: '', eventEndDate: '', eventEndTime: '',
        publishDeptIds: '', publishDeptNames: '',
        organizers: [], assistants: [], referees: [],
        fields: [],
        scoreFields: []
      },
      categories: [],
      deptTree: [],
      expandedDeptIds: [],
      showDeptPicker: false,
      selectedDeptIds: [],
      showUserPickerModal: false,
      userPickerRole: '',
      userPickerTitle: '',
      currentUserDeptId: null,
      userPickerUsers: [],
      userPickerSelected: [],
      showFieldEditor: false,
      editingFields: [],
      showScoreEditorModal: false,
      editingScores: []
    }
  },

  computed: {
    visibleDepts() {
      const result = []
      const walk = (nodes, depth) => {
        for (const n of nodes) {
          const hasChildren = n.children && n.children.length > 0
          const expanded = this.expandedDeptIds.includes(n.id)
          result.push({ id: n.id, name: n.name, depth, hasChildren, expanded })
          if (hasChildren && expanded) {
            walk(n.children, depth + 1)
          }
        }
      }
      walk(this.deptTree, 0)
      return result
    }
  },

  onLoad(opts) {
    if (opts.id) { this.id = opts.id; this.loadDetail() }
    this.loadDeptTree()
    this.loadCategories()
  },

  methods: {
    toTimestamp(dateStr, timeStr) {
      if (!dateStr) return 0
      return Math.floor(new Date(dateStr + ' ' + (timeStr || '00:00')).getTime() / 1000)
    },

    updateDateTime(prefix) {
      const date = this.form[prefix + 'Date']
      const time = this.form[prefix + 'Time']
      this.form[prefix + 'Str'] = (date && time) ? date + ' ' + time : (date || '')
    },

    async loadDetail() {
      try {
        const [res, deptRes] = await Promise.all([
          adminApi.eventDetail(this.id),
          adminApi.deptTree()
        ])
        this.deptTree = deptRes.data || []
        this.expandedDeptIds = this.deptTree.map(n => n.id)
        const d = res.data || {}
        this.form.title = d.title || ''
        this.form.type = d.type || 1
        this.form.cover = d.img || ''
        this.form.desc = d.desc || ''
        this.form.rules = d.rules || ''
        this.form.cateName = d.cateName || ''
        this.form.order = d.order ?? 0
        this.form.publishDeptIds = d.publishDeptIds || ''
        this.form.publishDeptNames = this.getDeptNames(d.publishDeptIds || '')
        if (d.regStart) {
          const dt = this.tsToParts(d.regStart)
          this.form.regStartDate = dt.date
          this.form.regStartTime = dt.time
        }
        if (d.regEnd) {
          const dt = this.tsToParts(d.regEnd)
          this.form.regEndDate = dt.date
          this.form.regEndTime = dt.time
        }
        if (d.eventStart) {
          const dt = this.tsToParts(d.eventStart)
          this.form.eventStartDate = dt.date
          this.form.eventStartTime = dt.time
        }
        if (d.eventEnd) {
          const dt = this.tsToParts(d.eventEnd)
          this.form.eventEndDate = dt.date
          this.form.eventEndTime = dt.time
        }
        if (d.forms) {
          try { this.form.fields = typeof d.forms === 'string' ? JSON.parse(d.forms) : d.forms } catch {}
        }
        if (d.scoreFields) {
          try { this.form.scoreFields = typeof d.scoreFields === 'string' ? JSON.parse(d.scoreFields) : d.scoreFields } catch {}
        }
        if (d.organizers) {
          this.form.organizers = (Array.isArray(d.organizers) ? d.organizers : []).map(r => ({ userId: r.userId || r.userID, name: r.name }))
        }
        if (d.assistants) {
          this.form.assistants = (Array.isArray(d.assistants) ? d.assistants : []).map(r => ({ userId: r.userId || r.userID, name: r.name }))
        }
        if (d.referees) {
          this.form.referees = (Array.isArray(d.referees) ? d.referees : []).map(r => ({ userId: r.userId || r.userID, name: r.name }))
        }
        if (d.type === 2) {
          await this.loadCategories('competition_type')
        } else {
          await this.loadCategories('activity_type')
        }
      } catch (e) { console.error(e) }
    },

    tsToParts(ts) {
      if (!ts || ts === '0' || ts === 0) return { date: '', time: '' }
      const d = new Date(Number(ts) * 1000)
      if (isNaN(d.getTime())) return { date: '', time: '' }
      const date = d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0')
      const time = String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0')
      return { date, time }
    },

    async loadCategories(dictKey) {
      try {
        const key = dictKey || (this.form.type === 2 ? 'competition_type' : 'activity_type')
        const res = await dictApi.items(key)
        this.categories = Array.isArray(res.data) ? res.data : (res.data.list || [])
      } catch (e) { this.categories = [] }
    },

    onCateChange(e) {
      const c = this.categories[e.detail.value]
      if (c) this.form.cateName = c.label || c.name
    },

    async loadDeptTree() {
      try {
        const res = await adminApi.deptTree()
        this.deptTree = res.data || []
        this.expandedDeptIds = this.deptTree.map(n => n.id)
      } catch (e) { console.error('加载部门失败', e) }
    },

    findDeptInTree(nodes, id) {
      for (const n of nodes) {
        if (n.id === id) return n
        if (n.children && n.children.length) {
          const found = this.findDeptInTree(n.children, id)
          if (found) return found
        }
      }
      return null
    },

    getDeptName(id) {
      const n = this.findDeptInTree(this.deptTree, Number(id))
      return n ? n.name : ''
    },

    getDeptNames(ids) {
      if (!ids) return ''
      return ids.split(',').map(id => this.getDeptName(id)).filter(Boolean).join('、')
    },

    toggleDeptExpand(id) {
      const idx = this.expandedDeptIds.indexOf(id)
      if (idx >= 0) { this.expandedDeptIds.splice(idx, 1) } else { this.expandedDeptIds.push(id) }
    },

    toggleDeptSelect(id) {
      const idx = this.selectedDeptIds.indexOf(id)
      if (idx >= 0) { this.selectedDeptIds.splice(idx, 1) } else { this.selectedDeptIds.push(id) }
    },

    toggleDeptPicker() {
      this.selectedDeptIds = this.form.publishDeptIds ? this.form.publishDeptIds.split(',').map(Number) : []
      this.showDeptPicker = true
    },

    confirmDeptPicker() {
      this.form.publishDeptIds = this.selectedDeptIds.join(',')
      this.form.publishDeptNames = this.selectedDeptIds.map(id => this.getDeptName(id)).filter(Boolean).join('、')
      this.showDeptPicker = false
    },

    toggleUserPicker(role) {
      const titleMap = { organizer: '主办人', assistant: '主办人助理', referee: '裁判' }
      this.userPickerRole = role
      this.userPickerTitle = titleMap[role] || ''
      this.currentUserDeptId = null
      this.userPickerUsers = []
      this.userPickerSelected = (this.form[role + 's'] || []).map(u => u.userId)
      this.showUserPickerModal = true
    },

    async handleUserDeptClick(id) {
      this.currentUserDeptId = id
      this.userPickerSelected = (this.form[this.userPickerRole + 's'] || []).map(u => u.userId)
      try {
        const res = await adminApi.deptUsers({ deptIds: String(id) })
        this.userPickerUsers = Array.isArray(res.data) ? res.data : (res.data.list || [])
      } catch (e) { this.userPickerUsers = [] }
    },

    isUserSelected(u) {
      return this.userPickerSelected.includes(u.id || u.userName || u.openid)
    },

    toggleUserSelect(u) {
      const key = u.id || u.userName || u.openid
      const idx = this.userPickerSelected.indexOf(key)
      if (idx >= 0) { this.userPickerSelected.splice(idx, 1) } else { this.userPickerSelected.push(key) }
    },

    confirmUserPicker() {
      const roleKey = this.userPickerRole + 's'
      const roleMap = { organizers: 'organizers', assistants: 'assistants', referees: 'referees' }
      const key = roleMap[roleKey] || roleKey
      this.form[key] = this.userPickerUsers
        .filter(u => this.userPickerSelected.includes(u.id || u.userName || u.openid))
        .map(u => ({ userId: u.id || u.userName || u.openid, name: u.name || u.userName || u.nickname }))
      this.showUserPickerModal = false
    },

    uploadImage() {
      uni.chooseImage({ count: 1, success: (res) => {
        const tempPath = res.tempFilePaths[0]
        uni.showLoading({ title: '上传中...' })
        uni.uploadFile({
          url: 'http://localhost:8080/upload', filePath: tempPath, name: 'file',
          success: (r) => {
            try {
              const data = JSON.parse(r.data)
              if (data.code === 0) { this.form.cover = data.data.url || data.data }
            } catch (e) { this.form.cover = r.data }
          },
          complete: () => { uni.hideLoading() }
        })
      }})
    },

    showFormEditor() {
      this.editingFields = this.form.fields.map(f => ({ ...f }))
      this.showFieldEditor = true
    },

    closeFieldEditor() {
      this.showFieldEditor = false
    },

    confirmFieldEditor() {
      this.form.fields = this.editingFields.map(f => ({ ...f }))
      this.showFieldEditor = false
    },

    showScoreEditor() {
      this.editingScores = this.form.scoreFields.map(sf => ({ ...sf }))
      this.showScoreEditorModal = true
    },

    closeScoreEditor() {
      this.showScoreEditorModal = false
    },

    confirmScoreEditor() {
      this.form.scoreFields = this.editingScores.map(sf => ({ ...sf }))
      this.showScoreEditorModal = false
    },

    async handleSubmit() {
      if (!this.form.title) { uni.showToast({ title: '请输入标题', icon: 'none' }); return }
      const payload = {
        title: this.form.title,
        type: String(this.form.type),
        status: '1',
        qr: this.form.cover || '',
        cateName: this.form.cateName || '',
        order: String(this.form.order || 0),
        regStart: String(this.toTimestamp(this.form.regStartDate, this.form.regStartTime)),
        regEnd: String(this.toTimestamp(this.form.regEndDate, this.form.regEndTime)),
        eventStart: String(this.toTimestamp(this.form.eventStartDate, this.form.eventStartTime)),
        eventEnd: String(this.toTimestamp(this.form.eventEndDate, this.form.eventEndTime)),
        forms: JSON.stringify(this.form.fields),
        scoreFields: JSON.stringify(this.form.scoreFields),
        obj: JSON.stringify({ desc: this.form.desc, rules: this.form.rules, cover: this.form.cover ? [this.form.cover] : [] }),
        deptId: '0',
        publishDeptIds: this.form.publishDeptIds,
        organizers: JSON.stringify(this.form.organizers.map(u => u.userId)),
        assistants: JSON.stringify(this.form.assistants.map(u => u.userId)),
        referees: JSON.stringify(this.form.referees.map(u => u.userId))
      }
      try {
        if (this.id) {
          payload.id = this.id
          await adminApi.eventEdit(payload)
        } else {
          await adminApi.eventInsert(payload)
        }
        uni.showToast({ title: '保存成功', icon: 'success' })
        setTimeout(() => { uni.navigateBack() }, 1500)
      } catch (e) { uni.showToast({ title: '保存失败', icon: 'none' }) }
    }
  }
}
</script>

<style scoped>
.container { min-height: 100vh; background-color: #f5f5f5; }
.form-scroll { padding: 20rpx; }
.form-group { background-color: #fff; border-radius: 12rpx; padding: 24rpx; margin-bottom: 20rpx; }
.group-title { font-size: 28rpx; font-weight: bold; color: #333; display: block; margin-bottom: 20rpx; padding-bottom: 12rpx; border-bottom: 1rpx solid #f0f0f0; }
.form-item { margin-bottom: 20rpx; }
.form-item:last-child { margin-bottom: 0; }
.label { font-size: 26rpx; color: #333; display: block; margin-bottom: 8rpx; }
.red { color: #fb454c; }
.input { height: 72rpx; background-color: #f5f5f5; border-radius: 8rpx; padding: 0 16rpx; font-size: 26rpx; }
.picker { height: 72rpx; background-color: #f5f5f5; border-radius: 8rpx; padding: 0 16rpx; font-size: 26rpx; line-height: 72rpx; color: #333; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.picker.placeholder { color: #999; }
.textarea { width: 100%; height: 120rpx; background-color: #f5f5f5; border-radius: 8rpx; padding: 12rpx 16rpx; font-size: 26rpx; box-sizing: border-box; }
.textarea-lg { height: 240rpx; }
.cover-upload { width: 200rpx; height: 200rpx; border-radius: 12rpx; overflow: hidden; position: relative; background-color: #f5f5f5; border: 2rpx dashed #ddd; display: flex; align-items: center; justify-content: center; }
.cover-preview { width: 200rpx; height: 200rpx; border-radius: 12rpx; }
.cover-placeholder { font-size: 60rpx; color: #ccc; }
.cover-del { position: absolute; top: 4rpx; right: 4rpx; width: 36rpx; height: 36rpx; background: rgba(0,0,0,0.5); color: #fff; border-radius: 50%; text-align: center; line-height: 36rpx; font-size: 24rpx; }
.btn-bar { padding: 20rpx 0; padding-bottom: calc(40rpx + env(safe-area-inset-bottom)); }
.submit-btn { height: 88rpx; background-color: #fb454c; border-radius: 44rpx; display: flex; align-items: center; justify-content: center; color: #fff; font-size: 32rpx; font-weight: bold; }

.modal-mask { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background-color: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 999; }
.modal-content { background-color: #fff; border-radius: 16rpx; padding: 32rpx; width: 620rpx; max-height: 80vh; overflow-y: auto; }
.modal-title { font-size: 32rpx; font-weight: bold; color: #333; display: block; margin-bottom: 24rpx; }
.modal-actions { display: flex; justify-content: flex-end; margin-top: 32rpx; }
.modal-actions .modal-btn { margin-left: 20rpx; }
.modal-actions .modal-btn:first-child { margin-left: 0; }
.modal-btn { font-size: 28rpx; padding: 16rpx 48rpx; border-radius: 8rpx; text-align: center; }
.btn-cancel { background-color: #eee; color: #666; }
.btn-confirm { background-color: #fb454c; color: #fff; }

.dept-item { display: flex; align-items: center; justify-content: space-between; padding: 20rpx 0; border-bottom: 2rpx solid #f5f5f5; }
.dept-left { display: flex; align-items: center; flex: 1; min-width: 0; }
.dept-arrow { font-size: 20rpx; color: #999; width: 32rpx; text-align: center; flex-shrink: 0; }
.dept-arrow-placeholder { visibility: hidden; }
.dept-name { font-size: 28rpx; color: #333; margin-left: 8rpx; }
.dept-check { font-size: 32rpx; color: #fb454c; font-weight: bold; width: 48rpx; text-align: center; flex-shrink: 0; }
.dept-uncheck { color: #ccc; font-weight: normal; }

.user-picker-depts { max-height: 300rpx; overflow-y: auto; margin-bottom: 16rpx; }
.user-picker-list { border-top: 2rpx solid #eee; padding-top: 16rpx; }
.user-picker-title { font-size: 24rpx; color: #666; margin-bottom: 12rpx; }
.user-tags { display: flex; flex-wrap: wrap; gap: 8rpx; margin-bottom: 16rpx; }
.user-tag { background-color: #f5f5f5; padding: 8rpx 20rpx; border-radius: 20rpx; font-size: 24rpx; color: #333; }
.user-tag.selected { background-color: #fb454c; color: #fff; }

.field-editor-row { display: flex; align-items: center; gap: 8rpx; margin-bottom: 12rpx; }
.field-editor-num { font-size: 24rpx; color: #999; width: 28rpx; flex-shrink: 0; }
.field-editor-input { flex: 1; height: 60rpx; background-color: #f5f5f5; border-radius: 6rpx; padding: 0 12rpx; font-size: 24rpx; min-width: 0; }
.field-editor-type { height: 60rpx; background-color: #f5f5f5; border-radius: 6rpx; padding: 0 12rpx; font-size: 24rpx; line-height: 60rpx; width: 120rpx; text-align: center; }
.field-editor-options { width: 140rpx; height: 60rpx; background-color: #f5f5f5; border-radius: 6rpx; padding: 0 12rpx; font-size: 24rpx; }
.field-editor-required { font-size: 22rpx; color: #fb454c; padding: 0 8rpx; flex-shrink: 0; }
.field-editor-del { font-size: 28rpx; color: #999; padding: 0 8rpx; flex-shrink: 0; }
.field-editor-add { text-align: center; padding: 16rpx; font-size: 26rpx; color: #fb454c; border: 2rpx dashed #ddd; border-radius: 8rpx; margin-top: 8rpx; }
</style>
